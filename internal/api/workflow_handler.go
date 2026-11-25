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
	"homolog/internal/auth"
	"homolog/internal/domain"
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

	item, err := h.workflowRepo.CheckoutItem(r.Context(), user.OrganizationID, itemID, user.UserID, user.Role, req.Reason)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowItemNotFound) {
			writeJSONError(w, http.StatusNotFound, "Section introuvable")
			return
		}
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

	item, err := h.workflowRepo.CheckinItem(r.Context(), user.OrganizationID, itemID, user.UserID, user.Role, req.Content, req.GEDDocumentID, req.Comment)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowItemNotFound) {
			writeJSONError(w, http.StatusNotFound, "Section introuvable")
			return
		}
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

	item, err := h.workflowRepo.AssignItem(r.Context(), user.OrganizationID, itemID, req.AssignedUserID, req.AssignedDepartment, dueTime)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowItemNotFound) {
			writeJSONError(w, http.StatusNotFound, "Section introuvable")
			return
		}
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

	item, err := h.workflowRepo.SubmitForReview(r.Context(), user.OrganizationID, itemID, user.UserID, user.Role, req.Notes)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowItemNotFound) {
			writeJSONError(w, http.StatusNotFound, "Section introuvable")
			return
		}
		var gatingErr *repository.GatingBlockedError
		if errors.As(err, &gatingErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success":             false,
				"code":                "GATING_BLOCKED",
				"error":               gatingErr.Error(),
				"evaluation":          gatingErr.Result,
				"unmet_prerequisites": gatingErr.Result.UnmetPrerequisites,
			})
			return
		}
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
		Comments string `json:"comments"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide")
		return
	}

	notesFinal := strings.TrimSpace(req.Notes)
	if notesFinal == "" {
		notesFinal = strings.TrimSpace(req.Comments)
	}

	if req.Decision != "approved" && req.Decision != "rejected" {
		writeJSONError(w, http.StatusBadRequest, "Décision invalide (valeurs attendues : approved ou rejected)")
		return
	}
	if notesFinal == "" {
		writeJSONError(w, http.StatusBadRequest, "Un motif de revue qualité est obligatoire (ALCOA+)")
		return
	}

	item, err := h.workflowRepo.ReviewItem(r.Context(), user.OrganizationID, itemID, user.UserID, user.Role, req.Decision, notesFinal)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowItemNotFound) {
			writeJSONError(w, http.StatusNotFound, "Section introuvable")
			return
		}
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
		Password                  string `json:"password"`
		TOTPCode                  string `json:"totp_code"`
		SolemnDeclaration         bool   `json:"solemn_declaration"`
		SolemnDeclarationAccepted bool   `json:"solemn_declaration_accepted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide")
		return
	}

	isAccepted := req.SolemnDeclaration || req.SolemnDeclarationAccepted
	if !isAccepted {
		writeJSONError(w, http.StatusBadRequest, "La déclaration solennelle sur l'honneur du Pharmacien DT est obligatoire pour sceller le document")
		return
	}

	// Génération de l'empreinte de signature cryptographique SHA-256
	sealInput := fmt.Sprintf("HOMOLOG-PDT-SEAL:%s:%s:%s:%d", itemID, user.UserID, user.Email, time.Now().UnixNano())
	sealHashBytes := sha256.Sum256([]byte(sealInput))
	sealHash := hex.EncodeToString(sealHashBytes[:])

	item, err := h.workflowRepo.ApproveItem(r.Context(), user.OrganizationID, itemID, user.UserID, user.Role, sealHash)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowItemNotFound) {
			writeJSONError(w, http.StatusNotFound, "Section introuvable")
			return
		}
		var gatingErr *repository.GatingBlockedError
		if errors.As(err, &gatingErr) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success":             false,
				"code":                "GATING_BLOCKED",
				"error":               gatingErr.Error(),
				"evaluation":          gatingErr.Result,
				"unmet_prerequisites": gatingErr.Result.UnmetPrerequisites,
			})
			return
		}
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

	reviews, err := h.workflowRepo.GetItemReviews(r.Context(), user.OrganizationID, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowItemNotFound) {
			writeJSONError(w, http.StatusNotFound, "Section introuvable")
			return
		}
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

// GetItemGating retourne l'évaluation complète des verrous réglementaires et prérequis d'une section
func (h *WorkflowHandler) GetItemGating(w http.ResponseWriter, r *http.Request) {
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

	eval, err := h.workflowRepo.EvaluateItemGating(r.Context(), user.OrganizationID, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkflowItemNotFound) {
			writeJSONError(w, http.StatusNotFound, "Section introuvable")
			return
		}
		log.Printf("[ERROR] GetItemGating %s: %v", itemID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible d'évaluer les verrous réglementaires")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"evaluation": eval,
	})
}

// GetDossierGatingSummary retourne la phase active du DAG et la liste des sections bloquées
func (h *WorkflowHandler) GetDossierGatingSummary(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de dossier invalide")
		return
	}

	summary, err := h.workflowRepo.GetDossierGatingSummary(r.Context(), user.OrganizationID, dossierID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Dossier introuvable")
			return
		}
		log.Printf("[ERROR] GetDossierGatingSummary %s: %v", dossierID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de calculer la synthèse du chemin critique")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"summary": summary,
	})
}

// GetDossierSmartPlanning calcule et retourne le rétroplanning dynamique inversé et l'analyse de chemin critique
func (h *WorkflowHandler) GetDossierSmartPlanning(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de dossier invalide")
		return
	}

	analysis, err := h.workflowRepo.GetDossierSmartPlanning(r.Context(), user.OrganizationID, dossierID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Dossier introuvable")
			return
		}
		log.Printf("[ERROR] GetDossierSmartPlanning %s: %v", dossierID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de calculer le rétroplanning inversé")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"planning": analysis,
	})
}

type SetDossierTargetDatesRequest struct {
	TargetDepositDate   *string `json:"target_deposit_date"`    // "YYYY-MM-DD"
	ANPPValidityEndDate *string `json:"anpp_validity_end_date"` // "YYYY-MM-DD"
	WorkingDaysPerWeek  *int    `json:"working_days_per_week"`  // 5 ou 6 jours ouvrables par semaine
}

// SetDossierTargetDates configure ou met à jour la date cible de dépôt ou l'échéance légale de renouvellement
func (h *WorkflowHandler) SetDossierTargetDates(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de dossier invalide")
		return
	}

	var req SetDossierTargetDatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	var targetDate *time.Time
	if req.TargetDepositDate != nil && strings.TrimSpace(*req.TargetDepositDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TargetDepositDate))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Format de date cible invalide (attendu YYYY-MM-DD)")
			return
		}
		targetDate = &t
	}

	var validityEnd *time.Time
	if req.ANPPValidityEndDate != nil && strings.TrimSpace(*req.ANPPValidityEndDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.ANPPValidityEndDate))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Format de date de validité invalide (attendu YYYY-MM-DD)")
			return
		}
		validityEnd = &t
	}

	var workingDays *int
	if req.WorkingDaysPerWeek != nil {
		wd := *req.WorkingDaysPerWeek
		if wd == 5 || wd == 6 {
			workingDays = &wd
		} else {
			defWd := 5
			workingDays = &defWd
		}
	}

	err = h.workflowRepo.UpdateDossierTargetDates(r.Context(), user.OrganizationID, dossierID, targetDate, validityEnd, workingDays)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Dossier introuvable")
			return
		}
		log.Printf("[ERROR] SetDossierTargetDates %s: %v", dossierID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de mettre à jour les dates du dossier")
		return
	}

	// Traçabilité ALCOA+ Audit Trail
	h.logAudit(r.Context(), user.OrganizationID, user.UserID, user.Email, user.Role,
		"SET_TARGET_DATES", "dossier", dossierID.String(), map[string]any{
			"target_deposit_date":    req.TargetDepositDate,
			"anpp_validity_end_date": req.ANPPValidityEndDate,
			"working_days_per_week":  req.WorkingDaysPerWeek,
		})

	// Recalcul immédiat du planning après modification
	analysis, _ := h.workflowRepo.GetDossierSmartPlanning(r.Context(), user.OrganizationID, dossierID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"message":  "Dates cibles et rétroplanning mis à jour avec succès",
		"planning": analysis,
	})
}

// SyncSmartPlanningDueDates propage les dates du chemin critique inversé aux sections CTD du dossier
func (h *WorkflowHandler) SyncSmartPlanningDueDates(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de dossier invalide")
		return
	}

	analysis, err := h.workflowRepo.GetDossierSmartPlanning(r.Context(), user.OrganizationID, dossierID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Dossier introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Impossible d'analyser le rétroplanning")
		return
	}

	items, err := h.workflowRepo.GetDossierItems(r.Context(), dossierID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Impossible de lire les items du dossier")
		return
	}

	dueDates := domain.CalculateSmartDueDates(analysis, items)
	updatedCount, err := h.workflowRepo.SyncSmartPlanningDueDates(r.Context(), user.OrganizationID, dossierID, dueDates)
	if err != nil {
		log.Printf("[ERROR] SyncSmartPlanningDueDates %s: %v", dossierID, err)
		writeJSONError(w, http.StatusInternalServerError, "Échec de la propagation des dates cibles aux tâches")
		return
	}

	// Traçabilité ALCOA+ Audit Trail
	h.logAudit(r.Context(), user.OrganizationID, user.UserID, user.Email, user.Role,
		"SYNC_SMART_PLANNING_DUE_DATES", "dossier", dossierID.String(), map[string]any{
			"updated_items_count": updatedCount,
			"target_deposit_date": analysis.TargetDepositDate.Format("2006-01-02"),
			"overall_health":      analysis.OverallHealth,
		})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":             true,
		"message":             fmt.Sprintf("Échéances synchronisées avec succès pour %d sections CTD", updatedCount),
		"updated_items_count": updatedCount,
		"planning":            analysis,
	})
}

// AlgerianHolidayItem représente un jour férié officiel algérien
type AlgerianHolidayItem struct {
	HolidayDate string `json:"holiday_date"` // YYYY-MM-DD
	NameFR      string `json:"name_fr"`
	HolidayType string `json:"holiday_type"` // NATIONAL, RELIGIEUX
	IsObserved  bool   `json:"is_observed"`
	Year        int    `json:"year"`
	DayOfWeek   string `json:"day_of_week"`
}

// GetAlgerianHolidays retourne le calendrier complet des jours fériés légaux et chômés en Algérie
func (h *WorkflowHandler) GetAlgerianHolidays(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT holiday_date, name_fr, holiday_type, is_observed
		FROM algerian_holidays
		ORDER BY holiday_date ASC
	`)
	if err != nil {
		log.Printf("[ERROR] GetAlgerianHolidays: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger le calendrier des jours fériés")
		return
	}
	defer rows.Close()

	holidays := make([]AlgerianHolidayItem, 0)
	for rows.Next() {
		var t time.Time
		var it AlgerianHolidayItem
		if err := rows.Scan(&t, &it.NameFR, &it.HolidayType, &it.IsObserved); err != nil {
			continue
		}
		it.HolidayDate = t.Format("2006-01-02")
		it.Year = t.Year()
		it.DayOfWeek = getDayOfWeekFR(t.Weekday())
		holidays = append(holidays, it)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"holidays":        holidays,
		"count":           len(holidays),
		"legal_reference": "Loi n° 63-278 du 26 juillet 1963 modifiée et complétée par la Loi n° 23-07 du 3 mai 2023 (extension à 3 jours chômés payés pour l'Aïd El-Fitr et l'Aïd El-Adha)",
	})
}

func getDayOfWeekFR(w time.Weekday) string {
	days := []string{"Dimanche", "Lundi", "Mardi", "Mercredi", "Jeudi", "Vendredi", "Samedi"}
	if int(w) >= 0 && int(w) < len(days) {
		return days[w]
	}
	return ""
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

// GetParapheurPendingItems liste les pièces en attente de visa réglementaire DT pour l'organisation
func (h *WorkflowHandler) GetParapheurPendingItems(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	normRole := strings.ToLower(user.Role)
	if normRole == "pharmacien_assistant" {
		delegation, _ := h.workflowRepo.GetActiveDTDelegation(r.Context(), user.OrganizationID, user.UserID)
		if delegation == nil {
			writeJSONError(w, http.StatusForbidden, "Accès restreint au Pharmacien DT titulaire ou à un Pharmacien Assistant titulaire d'un mandat d'intérim actif (Décret 21-82 Art. 24)")
			return
		}
	} else if normRole != "pharmacien_responsable" && normRole != "directeur_technique" && !user.HasPermission("DOC_APPROVE_PDT") {
		writeJSONError(w, http.StatusForbidden, "Accès restreint au Pharmacien Directeur Technique (1.A.7)")
		return
	}

	items, err := h.workflowRepo.GetParapheurPendingItems(r.Context(), user.OrganizationID)
	if err != nil {
		log.Printf("[ERROR] GetParapheurPendingItems: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger les pièces du parapheur")
		return
	}

	if items == nil {
		items = []*domain.ParapheurPendingItem{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"count": len(items),
	})
}

// BatchApproveSections applique le visa officiel 21 CFR Part 11 sur une liasse de pièces documentaires
func (h *WorkflowHandler) BatchApproveSections(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	normRole := strings.ToLower(user.Role)
	var activeDelegation *domain.DTDelegation
	if normRole == "pharmacien_assistant" {
		var err error
		activeDelegation, err = h.workflowRepo.GetActiveDTDelegation(r.Context(), user.OrganizationID, user.UserID)
		if err != nil || activeDelegation == nil {
			writeJSONError(w, http.StatusForbidden, "Garde-fou BPF & Décret 21-82 : Aucun mandat d'intérim actif ou notifié au MIPH pour ce Pharmacien Assistant")
			return
		}
	} else if normRole != "pharmacien_responsable" && normRole != "directeur_technique" && !user.HasPermission("DOC_APPROVE_PDT") {
		writeJSONError(w, http.StatusForbidden, "Garde-fou BPF : Seul le Pharmacien Directeur Technique titulaire (1.A.7) ou son Pharmacien Assistant sous mandat d'intérim (Décret 21-82) est habilité à apposer ce visa légal")
		return
	}

	var req domain.BatchApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	if len(req.ItemIDs) == 0 {
		writeJSONError(w, http.StatusBadRequest, "Aucune section sélectionnée pour la signature par lot")
		return
	}

	if !req.SolemnDeclarationAccepted {
		writeJSONError(w, http.StatusBadRequest, "La déclaration solennelle sur l'honneur du Pharmacien DT est obligatoire pour valider la signature par lot")
		return
	}

	// Re-authentification de sécurité (mot de passe ou TOTP 21 CFR Part 11) si mot de passe ou code fourni
	if strings.TrimSpace(req.Password) != "" || strings.TrimSpace(req.TOTPCode) != "" {
		var passwordHash string
		var totpEnabled bool
		var totpSecret *string
		userQuery := `SELECT password_hash, totp_enabled, totp_secret FROM users WHERE id = $1`
		err := h.pool.QueryRow(r.Context(), userQuery, user.UserID).Scan(&passwordHash, &totpEnabled, &totpSecret)
		if err == nil {
			if strings.TrimSpace(req.Password) != "" {
				validPassword, err := auth.VerifyPassword(req.Password, passwordHash)
				if err != nil || !validPassword {
					writeJSONError(w, http.StatusUnauthorized, "Mot de passe incorrect : signature par lot refusée")
					return
				}
			}
			if totpEnabled && totpSecret != nil && strings.TrimSpace(req.TOTPCode) != "" {
				if !auth.ValidateTOTPCode(req.TOTPCode, *totpSecret) {
					writeJSONError(w, http.StatusUnauthorized, "Code 2FA TOTP invalide : signature par lot refusée")
					return
				}
			}
		}
	}

	approverName := user.Email
	var uFirst, uLast string
	_ = h.pool.QueryRow(r.Context(), `SELECT first_name, last_name FROM users WHERE id = $1`, user.UserID).Scan(&uFirst, &uLast)
	if uFirst != "" || uLast != "" {
		approverName = strings.TrimSpace(uFirst + " " + uLast)
	}

	res, err := h.workflowRepo.BatchApproveSections(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Role,
		user.Email,
		approverName,
		req.ItemIDs,
	)
	if err != nil {
		log.Printf("[ERROR] BatchApproveSections: %v", err)
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur lors de la signature par lot : %v", err))
		return
	}

	// Audit trail consolidé
	for _, it := range res.Items {
		if it.Success {
			auditDetails := map[string]any{
				"item_code":    it.ItemCode,
				"dossier_code": it.DossierCode,
				"seal_hash":    it.SealHash,
				"batch_mode":   true,
			}
			if activeDelegation != nil {
				auditDetails["delegation_id"] = activeDelegation.ID
				auditDetails["miph_notification_ref"] = activeDelegation.MIPHNotificationRef
				auditDetails["interim_mandate"] = true
			}
			h.logAudit(
				r.Context(),
				user.OrganizationID,
				user.UserID,
				user.Email,
				user.Role,
				"batch_approve_sealed",
				"workflow_item",
				it.ItemID.String(),
				auditDetails,
			)
		}
	}

	writeJSON(w, http.StatusOK, res)
}

// ListDTDelegations liste tous les mandats d'intérim DT pour l'organisation
func (h *WorkflowHandler) ListDTDelegations(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	delegations, err := h.workflowRepo.ListDTDelegations(r.Context(), user.OrganizationID)
	if err != nil {
		log.Printf("[ERROR] ListDTDelegations: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger les délégations")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"delegations": delegations,
		"count":       len(delegations),
	})
}

// GetActiveDTDelegation renvoie le mandat d'intérim actif pour l'utilisateur courant s'il est assistant DT
func (h *WorkflowHandler) GetActiveDTDelegation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	delegation, err := h.workflowRepo.GetActiveDTDelegation(r.Context(), user.OrganizationID, user.UserID)
	if err != nil {
		log.Printf("[ERROR] GetActiveDTDelegation: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Erreur recherche mandat d'intérim")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"has_active_delegation": delegation != nil,
		"delegation":            delegation,
	})
}

// CreateDTDelegation enregistre un mandat d'intérim / suppléance (Décret 21-82 Art. 24)
func (h *WorkflowHandler) CreateDTDelegation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	normRole := strings.ToLower(user.Role)
	if normRole != "directeur_technique" && normRole != "pharmacien_responsable" {
		writeJSONError(w, http.StatusForbidden, "Seul le Pharmacien Directeur Technique titulaire peut déléguer ses fonctions")
		return
	}

	var req domain.CreateDTDelegationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	if req.DelegateID == uuid.Nil {
		writeJSONError(w, http.StatusBadRequest, "Le pharmacien assistant délégataire est requis")
		return
	}
	if strings.TrimSpace(req.MIPHNotificationRef) == "" {
		writeJSONError(w, http.StatusBadRequest, "La référence de la notification transmise au MIPH est obligatoire (Décret 21-82 Art. 24)")
		return
	}
	if strings.TrimSpace(req.Reason) == "" {
		writeJSONError(w, http.StatusBadRequest, "Le motif de la délégation / intérim est obligatoire")
		return
	}

	delegation, err := h.workflowRepo.CreateDTDelegation(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		req.DelegateID,
		strings.TrimSpace(req.MIPHNotificationRef),
		strings.TrimSpace(req.Reason),
		req.StartDate,
		req.EndDate,
	)
	if err != nil {
		log.Printf("[ERROR] CreateDTDelegation: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAudit(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Email,
		user.Role,
		"create_dt_delegation",
		"dt_delegation",
		delegation.ID.String(),
		map[string]any{
			"delegate_id":           req.DelegateID.String(),
			"miph_notification_ref": req.MIPHNotificationRef,
			"start_date":            req.StartDate,
			"end_date":              req.EndDate,
			"reason":                req.Reason,
		},
	)

	writeJSON(w, http.StatusCreated, delegation)
}

// RevokeDTDelegation met fin anticipée à un mandat d'intérim
func (h *WorkflowHandler) RevokeDTDelegation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	normRole := strings.ToLower(user.Role)
	if normRole != "directeur_technique" && normRole != "pharmacien_responsable" {
		writeJSONError(w, http.StatusForbidden, "Seul le Pharmacien Directeur Technique titulaire peut révoquer un mandat")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/parapheur/delegations/")
	idStr = strings.TrimSuffix(idStr, "/revoke")
	delegationID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID de délégation invalide")
		return
	}

	var req domain.RevokeDTDelegationRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "Reprise de fonctions du Pharmacien Directeur Technique titulaire"
	}

	if err := h.workflowRepo.RevokeDTDelegation(r.Context(), user.OrganizationID, delegationID, user.UserID, reason); err != nil {
		log.Printf("[ERROR] RevokeDTDelegation: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAudit(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Email,
		user.Role,
		"revoke_dt_delegation",
		"dt_delegation",
		delegationID.String(),
		map[string]any{
			"revoke_reason": reason,
		},
	)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Mandat d'intérim révoqué avec succès",
	})
}

// ListDTPermanentDelegations liste les délégations permanentes ciblées (hors intérim)
func (h *WorkflowHandler) ListDTPermanentDelegations(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	list, err := h.workflowRepo.ListDTPermanentDelegations(r.Context(), user.OrganizationID)
	if err != nil {
		log.Printf("[ERROR] ListDTPermanentDelegations: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger les délégations permanentes")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"permanent_delegations": list,
		"count":                 len(list),
	})
}

// GetActiveDTPermanentDelegation renvoie la délégation permanente active pour l'utilisateur courant
func (h *WorkflowHandler) GetActiveDTPermanentDelegation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	del, err := h.workflowRepo.GetActiveDTPermanentDelegation(r.Context(), user.OrganizationID, user.UserID)
	if err != nil {
		log.Printf("[ERROR] GetActiveDTPermanentDelegation: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Erreur recherche délégation permanente")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"has_active_permanent_delegation": del != nil,
		"permanent_delegation":            del,
	})
}

// CreateDTPermanentDelegation crée une délégation permanente ciblée Qualité/Production (hors intérim)
func (h *WorkflowHandler) CreateDTPermanentDelegation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	normRole := strings.ToLower(user.Role)
	if normRole != "directeur_technique" && normRole != "pharmacien_responsable" {
		writeJSONError(w, http.StatusForbidden, "Seul le Pharmacien Directeur Technique titulaire peut accorder une délégation permanente")
		return
	}

	var req domain.CreateDTPermanentDelegationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	if req.DelegateID == uuid.Nil {
		writeJSONError(w, http.StatusBadRequest, "Le pharmacien assistant délégataire est requis")
		return
	}
	if strings.TrimSpace(req.HabilitationRef) == "" {
		writeJSONError(w, http.StatusBadRequest, "La référence de la note/matrice d'habilitation interne est obligatoire")
		return
	}

	del, err := h.workflowRepo.CreateDTPermanentDelegation(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		req.DelegateID,
		strings.TrimSpace(req.HabilitationRef),
		req.AllowedTaxonomyPrefixes,
		strings.TrimSpace(req.Notes),
	)
	if err != nil {
		log.Printf("[ERROR] CreateDTPermanentDelegation: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAudit(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Email,
		user.Role,
		"create_dt_permanent_delegation",
		"dt_permanent_delegation",
		del.ID.String(),
		map[string]any{
			"delegate_id":               req.DelegateID.String(),
			"habilitation_ref":          del.HabilitationRef,
			"allowed_taxonomy_prefixes": del.AllowedTaxonomyPrefixes,
			"notes":                     del.Notes,
		},
	)

	writeJSON(w, http.StatusCreated, del)
}

// RevokeDTPermanentDelegation révoque une délégation permanente ciblée
func (h *WorkflowHandler) RevokeDTPermanentDelegation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	normRole := strings.ToLower(user.Role)
	if normRole != "directeur_technique" && normRole != "pharmacien_responsable" {
		writeJSONError(w, http.StatusForbidden, "Seul le Pharmacien Directeur Technique titulaire peut révoquer une délégation permanente")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/parapheur/permanent-delegations/")
	idStr = strings.TrimSuffix(idStr, "/revoke")
	delegationID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID de délégation invalide")
		return
	}

	var req domain.RevokeDTDelegationRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "Révocation de l'habilitation permanente par le Directeur Technique"
	}

	if err := h.workflowRepo.RevokeDTPermanentDelegation(r.Context(), user.OrganizationID, delegationID, user.UserID, reason); err != nil {
		log.Printf("[ERROR] RevokeDTPermanentDelegation: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAudit(
		r.Context(),
		user.OrganizationID,
		user.UserID,
		user.Email,
		user.Role,
		"revoke_dt_permanent_delegation",
		"dt_permanent_delegation",
		delegationID.String(),
		map[string]any{
			"revoke_reason": reason,
		},
	)

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Délégation permanente révoquée avec succès",
	})
}

// GetAvailableAssistants liste STRICTEMENT les pharmaciens assistants éligibles pour un mandat de suppléance DT
func (h *WorkflowHandler) GetAvailableAssistants(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	query := `
		SELECT id, first_name, last_name, email, role
		FROM users
		WHERE organization_id = $1 
		  AND is_active = TRUE
		  AND role = 'pharmacien_assistant'
		  AND id != $2
		ORDER BY first_name, last_name
	`
	rows, err := h.pool.Query(r.Context(), query, user.OrganizationID, user.UserID)
	if err != nil {
		log.Printf("[ERROR] GetAvailableAssistants: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Erreur chargement assistants")
		return
	}
	defer rows.Close()

	type AssistantItem struct {
		ID        uuid.UUID `json:"id"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		Email     string    `json:"email"`
		Role      string    `json:"role"`
	}

	assistants := make([]AssistantItem, 0)
	for rows.Next() {
		var a AssistantItem
		if err := rows.Scan(&a.ID, &a.FirstName, &a.LastName, &a.Email, &a.Role); err == nil {
			assistants = append(assistants, a)
		}
	}

	writeJSON(w, http.StatusOK, assistants)
}
