package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/audit"
	"homolog/internal/repository"
)

// WorkflowHandler gère les endpoints REST du cycle de vie documentaire Maker-Checker-Approver BPF
type WorkflowHandler struct {
	workflowRepo *repository.WorkflowRepository
	dossierRepo  *repository.DossierRepository
	auditService *audit.Service
	pool         *pgxpool.Pool
}

// NewWorkflowHandler instancie le gestionnaire d'API du workflow collaboratif
func NewWorkflowHandler(
	workflowRepo *repository.WorkflowRepository,
	dossierRepo *repository.DossierRepository,
	auditService *audit.Service,
	pool *pgxpool.Pool,
) *WorkflowHandler {
	return &WorkflowHandler{
		workflowRepo: workflowRepo,
		dossierRepo:  dossierRepo,
		auditService: auditService,
		pool:         pool,
	}
}

// CheckoutSection verrouille une section CTD en édition exclusive pour un rédacteur (Maker)
func (h *WorkflowHandler) CheckoutSection(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'élément invalide")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if strings.TrimSpace(req.Reason) == "" {
		req.Reason = "Édition technique de la section"
	}

	item, err := h.workflowRepo.CheckoutItem(r.Context(), itemID, user.UserID, user.Role, req.Reason)
	if err != nil {
		if errors.Is(err, repository.ErrSectionLocked) {
			writeJSONError(w, http.StatusConflict, "Cette section est déjà verrouillée en édition par un autre collaborateur")
			return
		}
		log.Printf("[ERROR] CheckoutItem %s: %v", itemID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de verrouiller la section")
		return
	}

	h.logAudit(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Email,
		user.Role,
		"WORKFLOW_CHECKOUT",
		"dossier_checklist_items",
		itemID.String(),
		map[string]any{
			"action":           "checkout",
			"item_code":        item.ItemCode,
			"current_revision": item.CurrentRevision,
			"reason":           req.Reason,
		},
	)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Section %s verrouillée en édition (Check-out)", item.ItemCode),
		"item":    item,
	})
}

// CheckinSection enregistre une nouvelle révision de travail (v0.x) et lève le verrou
func (h *WorkflowHandler) CheckinSection(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'élément invalide")
		return
	}

	var req struct {
		Content         *string    `json:"content"`
		GEDDocumentID   *uuid.UUID `json:"ged_document_id"`
		Comment         string     `json:"comment"`
		RevisionSummary string     `json:"revision_summary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide")
		return
	}
	if strings.TrimSpace(req.Comment) == "" && strings.TrimSpace(req.RevisionSummary) != "" {
		req.Comment = strings.TrimSpace(req.RevisionSummary)
	}
	if strings.TrimSpace(req.Comment) == "" {
		req.Comment = "Mise à jour de la documentation"
	}

	item, err := h.workflowRepo.CheckinItem(r.Context(), itemID, user.UserID, user.Role, req.Content, req.GEDDocumentID, req.Comment)
	if err != nil {
		if errors.Is(err, repository.ErrSectionNotLocked) {
			writeJSONError(w, http.StatusBadRequest, "La section n'est pas verrouillée")
			return
		}
		if errors.Is(err, repository.ErrUnauthorizedCheckin) {
			writeJSONError(w, http.StatusForbidden, "Seul l'utilisateur ayant posé le verrou est habilité à effectuer le Check-in")
			return
		}
		log.Printf("[ERROR] CheckinItem %s: %v", itemID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible d'effectuer le Check-in")
		return
	}

	h.logAudit(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Email,
		user.Role,
		"WORKFLOW_CHECKIN",
		"dossier_checklist_items",
		itemID.String(),
		map[string]any{
			"action":       "checkin",
			"item_code":    item.ItemCode,
			"new_revision": item.CurrentRevision,
			"comment":      req.Comment,
		},
	)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Révision %s enregistrée avec succès (Check-in)", item.CurrentRevision),
		"item":    item,
	})
}

// AssignSection affecte la section à un rédacteur et un pôle métier avec date butoir
func (h *WorkflowHandler) AssignSection(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'élément invalide")
		return
	}

	var req struct {
		AssignedUserID     *uuid.UUID `json:"assigned_user_id"`
		AssignedDepartment string     `json:"assigned_department"`
		DueDate            *string    `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide")
		return
	}

	if req.AssignedDepartment == "" {
		req.AssignedDepartment = "RA"
	}

	var dueTime *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		t, err := time.Parse("2006-01-02", *req.DueDate)
		if err == nil {
			dueTime = &t
		}
	}

	item, err := h.workflowRepo.AssignItem(r.Context(), itemID, req.AssignedUserID, req.AssignedDepartment, dueTime)
	if err != nil {
		log.Printf("[ERROR] AssignItem %s: %v", itemID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible d'assigner la section")
		return
	}

	h.logAudit(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Email,
		user.Role,
		"WORKFLOW_ASSIGN",
		"dossier_checklist_items",
		itemID.String(),
		map[string]any{
			"action":     "assign",
			"item_code":  item.ItemCode,
			"department": req.AssignedDepartment,
			"user_id":    req.AssignedUserID,
		},
	)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Section %s affectée au pôle %s", item.ItemCode, req.AssignedDepartment),
		"item":    item,
	})
}

// SubmitReview soumet le document achevé pour examen par le Checker Qualité
func (h *WorkflowHandler) SubmitReview(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'élément invalide")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if strings.TrimSpace(req.Notes) == "" {
		req.Notes = "Soumis pour relecture et contrôle qualité"
	}

	item, err := h.workflowRepo.SubmitForReview(r.Context(), itemID, user.UserID, user.Role, req.Notes)
	if err != nil {
		log.Printf("[ERROR] SubmitForReview %s: %v", itemID, err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAudit(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Email,
		user.Role,
		"WORKFLOW_SUBMIT_REVIEW",
		"dossier_checklist_items",
		itemID.String(),
		map[string]any{
			"action":    "submit_review",
			"item_code": item.ItemCode,
			"revision":  item.CurrentRevision,
		},
	)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Section %s soumise pour revue qualité (Checker/AQ)", item.ItemCode),
		"item":    item,
	})
}

// ReviewDecision enregistre la décision d'approbation qualité ou de rejet motivé (Checker)
func (h *WorkflowHandler) ReviewDecision(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'élément invalide")
		return
	}

	var req struct {
		Decision string `json:"decision"` // "approved" ou "rejected"
		Notes    string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide")
		return
	}

	if req.Decision != "approved" && req.Decision != "rejected" {
		writeJSONError(w, http.StatusBadRequest, "Décision invalide (valeurs attendues : approved ou rejected)")
		return
	}
	if strings.TrimSpace(req.Notes) == "" {
		writeJSONError(w, http.StatusBadRequest, "Un motif de revue qualité est obligatoire (ALCOA+)")
		return
	}

	item, err := h.workflowRepo.ReviewItem(r.Context(), itemID, user.UserID, user.Role, req.Decision, req.Notes)
	if err != nil {
		if errors.Is(err, repository.ErrSelfReviewNotAllowed) {
			writeJSONError(w, http.StatusForbidden, "Ségrégation des fonctions BPF : Le rédacteur ne peut pas valider ou rejeter son propre document")
			return
		}
		log.Printf("[ERROR] ReviewDecision %s: %v", itemID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible d'enregistrer la décision de revue")
		return
	}

	h.logAudit(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Email,
		user.Role,
		"WORKFLOW_REVIEW_DECISION",
		"dossier_checklist_items",
		itemID.String(),
		map[string]any{
			"action":    "review_decision",
			"item_code": item.ItemCode,
			"decision":  req.Decision,
			"notes":     req.Notes,
		},
	)

	statusLabel := "validée (Prêt pour visa DT)"
	if req.Decision == "rejected" {
		statusLabel = "renvoyée en rédaction pour corrections"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Revue qualité enregistrée : Section %s %s", item.ItemCode, statusLabel),
		"item":    item,
	})
}

// ApproveSection appose le visa officiel 21 CFR Part 11 du Pharmacien Directeur Technique et scelle le document en v1.0
func (h *WorkflowHandler) ApproveSection(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	// Contrôle d'éligibilité du rôle (PDT légal ou Admin labo habilité)
	normRole := strings.ToLower(user.Role)
	if normRole != "pharmacien_responsable" && normRole != "directeur_technique" && normRole != "admin" {
		writeJSONError(w, http.StatusForbidden, "Garde-fou BPF : Seul le Pharmacien Directeur Technique titulaire (1.A.7) est habilité à apposer le visa légal d'approbation")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'élément invalide")
		return
	}

	var req struct {
		Password          string `json:"password"`
		TOTPCode          string `json:"totp_code"`
		SolemnDeclaration bool   `json:"solemn_declaration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide")
		return
	}

	if !req.SolemnDeclaration {
		writeJSONError(w, http.StatusBadRequest, "La déclaration solennelle sur l'honneur du Pharmacien DT est obligatoire pour sceller le document")
		return
	}

	// Génération de l'empreinte de signature cryptographique SHA-256
	sealInput := fmt.Sprintf("HOMOLOG-PDT-SEAL:%s:%s:%s:%d", itemID, user.UserID, user.Email, time.Now().UnixNano())
	sealHashBytes := sha256.Sum256([]byte(sealInput))
	sealHash := hex.EncodeToString(sealHashBytes[:])

	item, err := h.workflowRepo.ApproveItem(r.Context(), itemID, user.UserID, user.Role, sealHash)
	if err != nil {
		if errors.Is(err, repository.ErrSelfReviewNotAllowed) {
			writeJSONError(w, http.StatusForbidden, "Ségrégation des fonctions BPF : Le rédacteur ne peut pas s'auto-approuver")
			return
		}
		log.Printf("[ERROR] ApproveSection %s: %v", itemID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de sceller le document")
		return
	}

	h.logAudit(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Email,
		user.Role,
		"WORKFLOW_PDT_APPROVAL",
		"dossier_checklist_items",
		itemID.String(),
		map[string]any{
			"action":         "approve_sealed",
			"item_code":      item.ItemCode,
			"signature_seal": sealHash,
			"revision":       "v1.0",
		},
	)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"message":        fmt.Sprintf("Section %s officiellement visée et scellée en v1.0 par le Pharmacien DT", item.ItemCode),
		"signature_seal": sealHash,
		"item":           item,
	})
}

// GetItemReviews retourne l'historique complet des annotations et décisions qualité d'une section
func (h *WorkflowHandler) GetItemReviews(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'élément invalide")
		return
	}

	reviews, err := h.workflowRepo.GetItemReviews(r.Context(), itemID)
	if err != nil {
		log.Printf("[ERROR] GetItemReviews %s: %v", itemID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger les revues")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"item_id": itemID,
		"reviews": reviews,
		"total":   len(reviews),
	})
}

// GetMasterTemplates retourne la liste des modèles maîtres filtrés par catégorie
func (h *WorkflowHandler) GetMasterTemplates(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	templates, err := h.workflowRepo.GetMasterTemplates(r.Context(), category)
	if err != nil {
		log.Printf("[ERROR] GetMasterTemplates: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger les modèles maîtres")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"total":     len(templates),
	})
}

// GetMasterTemplateByCode retourne un modèle maître par son code unique
func (h *WorkflowHandler) GetMasterTemplateByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	template, err := h.workflowRepo.GetMasterTemplateByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowItemNotFound) {
			writeJSONError(w, http.StatusNotFound, "Modèle maître introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur chargement modèle")
		return
	}

	writeJSON(w, http.StatusOK, template)
}

func (h *WorkflowHandler) logAudit(ctx context.Context, orgID, userID uuid.UUID, email, role, action, entityType, entityID string, details map[string]any) {
	if h.auditService == nil {
		return
	}
	_, _ = h.auditService.Log(ctx, &audit.Entry{
		OrganizationID: orgID,
		UserID:         &userID,
		UserEmail:      email,
		UserRole:       role,
		Action:         action,
		EntityType:     entityType,
		EntityID:       entityID,
		Details:        details,
	})
}
