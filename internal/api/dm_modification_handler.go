package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// DMModificationHandler expose les routes REST de gestion des modifications DM (Décret 20-324)
type DMModificationHandler struct {
	modRepo  *repository.DMModificationRepository
	dmRepo   *repository.DMRepository
	auditSvc *audit.Service
}

func NewDMModificationHandler(
	modRepo *repository.DMModificationRepository,
	dmRepo *repository.DMRepository,
	auditSvc *audit.Service,
) *DMModificationHandler {
	return &DMModificationHandler{
		modRepo:  modRepo,
		dmRepo:   dmRepo,
		auditSvc: auditSvc,
	}
}

// ListModifications renvoie la liste des modifications post-homologation
func (h *DMModificationHandler) ListModifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetSessionUser(ctx)
	if user == nil {
		http.Error(w, `{"error":"Non authentifié"}`, http.StatusUnauthorized)
		return
	}

	deviceIDStr := r.URL.Query().Get("device_id")
	var deviceID *uuid.UUID
	if deviceIDStr != "" {
		if id, err := uuid.Parse(deviceIDStr); err == nil {
			deviceID = &id
		}
	}
	typeFilter := r.URL.Query().Get("type")
	statusFilter := r.URL.Query().Get("status")

	mods, err := h.modRepo.ListDMModifications(ctx, user.OrganizationID, deviceID, typeFilter, statusFilter)
	if err != nil {
		http.Error(w, `{"error":"Impossible de charger les modifications DM"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"modifications": mods,
		"total":         len(mods),
	})
}

// GetModification renvoie le détail d'une modification
func (h *DMModificationHandler) GetModification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetSessionUser(ctx)
	if user == nil {
		http.Error(w, `{"error":"Non authentifié"}`, http.StatusUnauthorized)
		return
	}

	modID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"ID de modification invalide"}`, http.StatusBadRequest)
		return
	}

	mod, err := h.modRepo.GetDMModificationByID(ctx, user.OrganizationID, modID)
	if err != nil {
		http.Error(w, `{"error":"Modification DM introuvable"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"modification": mod,
	})
}

// CreateModification crée une nouvelle demande de modification
func (h *DMModificationHandler) CreateModification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetSessionUser(ctx)
	if user == nil {
		http.Error(w, `{"error":"Non authentifié"}`, http.StatusUnauthorized)
		return
	}

	var req domain.DMModification
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Payload JSON invalide"}`, http.StatusBadRequest)
		return
	}

	req.OrganizationID = user.OrganizationID
	if err := req.Validate(); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if err := h.modRepo.CreateDMModification(ctx, &req); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(ctx, &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "CREATE",
			EntityType:     "DM_MODIFICATION",
			EntityID:       req.ID.String(),
			Details: map[string]any{
				"code":  req.ModificationCode,
				"type":  req.ModificationType,
				"title": req.Title,
				"fee":   req.FeeAmount,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"modification": req,
		"message":      "Demande de modification post-homologation créée avec succès",
	})
}

// UpdateModification met à jour les informations d'une modification
func (h *DMModificationHandler) UpdateModification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetSessionUser(ctx)
	if user == nil {
		http.Error(w, `{"error":"Non authentifié"}`, http.StatusUnauthorized)
		return
	}

	modID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"ID invalide"}`, http.StatusBadRequest)
		return
	}

	var req domain.DMModification
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Payload JSON invalide"}`, http.StatusBadRequest)
		return
	}

	req.ID = modID
	req.OrganizationID = user.OrganizationID

	if err := h.modRepo.UpdateDMModification(ctx, &req); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Modification DM mise à jour avec succès",
	})
}

// SealModification appose le visa Pharmacien DT et le sceau 21 CFR Part 11
func (h *DMModificationHandler) SealModification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetSessionUser(ctx)
	if user == nil {
		http.Error(w, `{"error":"Non authentifié"}`, http.StatusUnauthorized)
		return
	}

	modID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"ID invalide"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		SolemnDeclaration bool   `json:"solemn_declaration"`
		DecisionMIPH      string `json:"decision_miph"`
		Notes             string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if !body.SolemnDeclaration {
		http.Error(w, `{"error":"L'engagement solennel du Pharmacien DT est obligatoire (21 CFR Part 11)"}`, http.StatusBadRequest)
		return
	}

	mod, err := h.modRepo.GetDMModificationByID(ctx, user.OrganizationID, modID)
	if err != nil {
		http.Error(w, `{"error":"Modification DM introuvable"}`, http.StatusNotFound)
		return
	}

	pdtName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	decisionMIPH := body.DecisionMIPH
	if decisionMIPH == "" {
		decisionMIPH = "DEC-MIPH-PDT-2022-7714"
	}

	seal := mod.GenerateSeal(pdtName, decisionMIPH)

	if err := h.modRepo.SealDMModification(ctx, user.OrganizationID, modID, user.UserID, seal, body.Notes); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(ctx, &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "SEAL_DT",
			EntityType:     "DM_MODIFICATION",
			EntityID:       modID.String(),
			Details: map[string]any{
				"code": mod.ModificationCode,
				"seal": seal,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message": "Acte de modification scellé par le Pharmacien DT avec succès",
		"seal":    seal,
		"status":  domain.DMModStatusSoumis,
	})
}

// IssueAvenant enregistre la décision officielle d'homologation de la DDM
func (h *DMModificationHandler) IssueAvenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetSessionUser(ctx)
	if user == nil {
		http.Error(w, `{"error":"Non authentifié"}`, http.StatusUnauthorized)
		return
	}

	modID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"ID invalide"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		AvenantNumber string `json:"avenant_number"`
		DecisionNotes string `json:"decision_notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	if strings.TrimSpace(body.AvenantNumber) == "" {
		body.AvenantNumber = "AVN-DDM-2026-" + modID.String()[:8]
	}

	if err := h.modRepo.IssueAvenant(ctx, user.OrganizationID, modID, body.AvenantNumber, body.DecisionNotes); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":        "Avenant DDM délivré et enregistré avec succès",
		"avenant_number": body.AvenantNumber,
		"status":         domain.DMModStatusApprouve,
	})
}

// GetStats renvoie les indicateurs du cockpit de modifications DM
func (h *DMModificationHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetSessionUser(ctx)
	if user == nil {
		http.Error(w, `{"error":"Non authentifié"}`, http.StatusUnauthorized)
		return
	}

	stats, err := h.modRepo.GetDMModificationStats(ctx, user.OrganizationID)
	if err != nil {
		http.Error(w, `{"error":"Impossible de charger les statistiques"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"stats": stats,
	})
}

// GetCategories renvoie le catalogue standard des catégories de modifications
func (h *DMModificationHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"categories": domain.StandardDMCategories,
	})
}
