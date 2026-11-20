package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// SettingsHandler gère les paramètres réglementaires et barèmes découplés par tenant
type SettingsHandler struct {
	repo         *repository.SettingsRepository
	auditService *audit.Service
}

// NewSettingsHandler instancie le handler des paramètres
func NewSettingsHandler(repo *repository.SettingsRepository, auditService *audit.Service) *SettingsHandler {
	return &SettingsHandler{
		repo:         repo,
		auditService: auditService,
	}
}

// ListTenantRegulatorySettings liste l'ensemble des paramètres réglementaires en vigueur pour l'organisation
func (h *SettingsHandler) ListTenantRegulatorySettings(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	settings, err := h.repo.GetActiveSettingsList(r.Context(), user.OrganizationID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur chargement des paramètres réglementaires")
		return
	}

	// Si aucun paramètre n'est initialisé, on initialise automatiquement les valeurs par défaut
	if len(settings) == 0 {
		_ = h.repo.EnsureTenantDefaults(r.Context(), user.OrganizationID)
		settings, _ = h.repo.GetActiveSettingsList(r.Context(), user.OrganizationID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success":  true,
		"count":    len(settings),
		"settings": settings,
	})
}

// GetTenantSettingHistory retourne l'historique temporel d'un paramètre
func (h *SettingsHandler) GetTenantSettingHistory(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		writeJSONError(w, http.StatusBadRequest, "Clé de paramètre requise")
		return
	}

	history, err := h.repo.GetSettingHistory(r.Context(), user.OrganizationID, key)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur chargement de l'historique")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"key":     key,
		"count":   len(history),
		"history": history,
	})
}

// UpdateTenantRegulatorySetting met à jour un paramètre réglementaire (soit en direct soit par proposition Maker-Checker)
func (h *SettingsHandler) UpdateTenantRegulatorySetting(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		writeJSONError(w, http.StatusBadRequest, "Clé de paramètre requise")
		return
	}

	var req domain.UpdateTenantSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Payload JSON invalide")
		return
	}

	req.NewValue = strings.TrimSpace(req.NewValue)
	req.RegulatoryReason = strings.TrimSpace(req.RegulatoryReason)
	if req.NewValue == "" {
		writeJSONError(w, http.StatusBadRequest, "La nouvelle valeur ne peut être vide")
		return
	}
	if req.RegulatoryReason == "" {
		writeJSONError(w, http.StatusBadRequest, "Le motif réglementaire est obligatoire (ex: référence décret ou notification ANPP)")
		return
	}

	normRole := strings.ToLower(user.Role)
	isDT := normRole == "directeur_technique" || normRole == "pharmacien_responsable" || normRole == "pharmacien_assistant"

	// Si l'utilisateur est le DT ou si direct approval est possible :
	if isDT && !req.RequireDTApproval {
		err := h.repo.DirectApplySetting(r.Context(), user.OrganizationID, key, req.NewValue, req.RegulatoryReason, user.UserID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Erreur application du paramètre: "+err.Error())
			return
		}

		if h.auditService != nil {
			_, _ = h.auditService.Log(r.Context(), &audit.Entry{
				OrganizationID: user.OrganizationID,
				UserID:         &user.UserID,
				UserEmail:      user.Email,
				UserRole:       user.Role,
				Action:         "TENANT_SETTING_DIRECTLY_APPLIED",
				EntityType:     "TENANT_REGULATORY_SETTING",
				EntityID:       key,
				Details: map[string]any{
					"key":               key,
					"new_value":         req.NewValue,
					"regulatory_reason": req.RegulatoryReason,
					"applied_by":        user.Email,
				},
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"message": "Paramètre réglementaire mis à jour et scellé avec succès",
			"status":  "ACTIVE",
		})
		return
	}

	// Sinon, enregistrement comme proposition Maker en attente du visa DT
	err := h.repo.ProposeSettingUpdate(r.Context(), user.OrganizationID, key, req.NewValue, req.RegulatoryReason, user.UserID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur enregistrement de la proposition: "+err.Error())
		return
	}

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "TENANT_SETTING_UPDATE_PROPOSED",
			EntityType:     "TENANT_REGULATORY_SETTING",
			EntityID:       key,
			Details: map[string]any{
				"key":               key,
				"proposed_value":    req.NewValue,
				"regulatory_reason": req.RegulatoryReason,
				"proposed_by":       user.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Proposition enregistrée. En attente du visa réglementaire du Pharmacien Directeur Technique",
		"status":  "PENDING_APPROVAL_DT",
	})
}

// ApproveTenantRegulatorySetting valide une proposition de modification en attente (réservé au Pharmacien DT)
func (h *SettingsHandler) ApproveTenantRegulatorySetting(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	normRole := strings.ToLower(user.Role)
	if normRole != "directeur_technique" && normRole != "pharmacien_responsable" && normRole != "pharmacien_assistant" && normRole != "admin" {
		writeJSONError(w, http.StatusForbidden, "Seul le Pharmacien Directeur Technique ou son suppléant est habilité à approuver les modifications institutionnelles")
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		writeJSONError(w, http.StatusBadRequest, "Clé de paramètre requise")
		return
	}

	var req domain.ApproveTenantSettingRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if !req.SolemnCheck {
		writeJSONError(w, http.StatusUnprocessableEntity, "L'engagement solennel de conformité réglementaire est obligatoire (21 CFR Part 11)")
		return
	}

	err := h.repo.ApproveAndApplySetting(r.Context(), user.OrganizationID, key, user.UserID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur approbation du paramètre: "+err.Error())
		return
	}

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "TENANT_SETTING_APPROVED_BY_DT",
			EntityType:     "TENANT_REGULATORY_SETTING",
			EntityID:       key,
			Details: map[string]any{
				"key":         key,
				"approved_by": user.Email,
				"solemn_visa": true,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Paramètre réglementaire approuvé et mis en vigueur avec succès",
		"status":  "ACTIVE",
	})
}
