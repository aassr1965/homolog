package api

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// DMHandler gère les endpoints REST pour l'homologation 6 sections, la recevabilité et la régularisation des DM
type DMHandler struct {
	dmRepo      *repository.DMRepository
	dossierRepo *repository.DossierRepository
	auditSvc    *audit.Service
}

// NewDMHandler initialise une nouvelle instance de DMHandler
func NewDMHandler(dmRepo *repository.DMRepository, dossierRepo *repository.DossierRepository, auditSvc *audit.Service) *DMHandler {
	return &DMHandler{
		dmRepo:      dmRepo,
		dossierRepo: dossierRepo,
		auditSvc:    auditSvc,
	}
}

// GetMedicalDeviceHomologation renvoie les données complètes des 6 sections techniques d'un DM
func (h *DMHandler) GetMedicalDeviceHomologation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant DM invalide")
		return
	}

	dm, err := h.dmRepo.GetMedicalDeviceHomologation(r.Context(), user.OrganizationID, deviceID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"medical_device": dm,
	})
}

// UpdateMedicalDeviceHomologation met à jour les paramètres des 6 sections techniques
func (h *DMHandler) UpdateMedicalDeviceHomologation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	deviceID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant DM invalide")
		return
	}

	var req domain.MedicalDevice
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide: "+err.Error())
		return
	}

	req.ID = deviceID
	req.OrganizationID = user.OrganizationID

	if err := h.dmRepo.UpdateMedicalDeviceHomologation(r.Context(), &req); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Audit trail
	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "DM_HOMOLOGATION_UPDATE",
			EntityType:     "medical_devices",
			EntityID:       deviceID.String(),
			Details: map[string]any{
				"commercial_name": req.CommercialNameFR,
				"dm_class":        req.DMClass,
				"is_sterile":      req.IsSterile,
			},
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Fiche technique d'homologation DM mise à jour avec succès",
	})
}

// GetRecevabiliteEvaluation renvoie l'évaluation de recevabilité en 13 points pour un dossier DM
func (h *DMHandler) GetRecevabiliteEvaluation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	eval, err := h.dmRepo.GetRecevabiliteEvaluation(r.Context(), user.OrganizationID, dossierID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"evaluation": eval,
	})
}

// UpdateRecevabiliteEvaluation enregistre la grille d'évaluation de recevabilité
func (h *DMHandler) UpdateRecevabiliteEvaluation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	var req domain.DMRecevabiliteEvaluation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide: "+err.Error())
		return
	}

	// Récupérer l'enregistrement existant
	eval, err := h.dmRepo.GetRecevabiliteEvaluation(r.Context(), user.OrganizationID, dossierID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Évaluation de recevabilité introuvable: "+err.Error())
		return
	}

	if eval.PDTSignatureSeal != "" {
		writeJSONError(w, http.StatusForbidden, "Cette évaluation a déjà été scellée par le Pharmacien DT et ne peut plus être modifiée (21 CFR Part 11)")
		return
	}

	// Mettre à jour les champs
	if req.DepositOrderNumber != "" {
		eval.DepositOrderNumber = req.DepositOrderNumber
	}
	if !req.DepositDate.IsZero() {
		eval.DepositDate = req.DepositDate
	}
	if req.ElectronicCopiesCount > 0 {
		eval.ElectronicCopiesCount = req.ElectronicCopiesCount
	}
	if req.ElectronicMediaType != "" {
		eval.ElectronicMediaType = req.ElectronicMediaType
	}
	if req.Item1Status != "" {
		eval.Item1Status = req.Item1Status
	}
	if req.Item2Status != "" {
		eval.Item2Status = req.Item2Status
	}
	if req.Item3Status != "" {
		eval.Item3Status = req.Item3Status
	}
	if req.Item4Status != "" {
		eval.Item4Status = req.Item4Status
	}
	if req.Item5Status != "" {
		eval.Item5Status = req.Item5Status
	}
	if req.Item6Status != "" {
		eval.Item6Status = req.Item6Status
	}
	if req.Item7Status != "" {
		eval.Item7Status = req.Item7Status
	}
	if req.Item8Status != "" {
		eval.Item8Status = req.Item8Status
	}
	if req.Item9Status != "" {
		eval.Item9Status = req.Item9Status
	}
	if req.Item10Status != "" {
		eval.Item10Status = req.Item10Status
	}
	if req.Item11Status != "" {
		eval.Item11Status = req.Item11Status
	}
	if req.Item12Status != "" {
		eval.Item12Status = req.Item12Status
	}
	if req.Item13Status != "" {
		eval.Item13Status = req.Item13Status
	}
	eval.ReservesNotes = req.ReservesNotes
	if req.ANPPAgentName != "" {
		eval.ANPPAgentName = req.ANPPAgentName
	}

	if err := h.dmRepo.SaveRecevabiliteEvaluation(r.Context(), eval); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "DM_RECEVABILITE_UPDATE",
			EntityType:     "dm_recevabilite_evaluations",
			EntityID:       eval.ID.String(),
			Details: map[string]any{
				"deposit_order": eval.DepositOrderNumber,
				"score":         eval.ConformityScore,
				"decision":      eval.Decision,
			},
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"message":    "Évaluation de recevabilité enregistrée",
		"evaluation": eval,
	})
}

// SealRecevabiliteEvaluation permet au Pharmacien DT de viser et sceller l'acte officiel ES-REG-03-DM
func (h *DMHandler) SealRecevabiliteEvaluation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	var req struct {
		SolemnDeclaration bool   `json:"solemn_declaration"`
		DecisionMIPH      string `json:"decision_miph"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide: "+err.Error())
		return
	}

	if !req.SolemnDeclaration {
		writeJSONError(w, http.StatusUnprocessableEntity, "L'engagement solennel du Pharmacien Directeur Technique est obligatoire selon l'Arrêté du 10 mai 2021")
		return
	}

	eval, err := h.dmRepo.GetRecevabiliteEvaluation(r.Context(), user.OrganizationID, dossierID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Évaluation de recevabilité introuvable: "+err.Error())
		return
	}

	if eval.Decision == domain.RecevabiliteDecisionIrrecevable {
		writeJSONError(w, http.StatusUnprocessableEntity, "Impossible de sceller un dossier jugé irrecevable (score insuffisant ou pièces bloquantes non conformes)")
		return
	}

	seal := eval.GenerateSeal(user.Email, req.DecisionMIPH)
	hash := fmt.Sprintf("%x", time.Now().UnixNano())

	if err := h.dmRepo.SealRecevabiliteEvaluation(r.Context(), user.OrganizationID, eval.ID, user.UserID, seal, hash); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "DM_RECEVABILITE_SEAL_PDT",
			EntityType:     "dm_recevabilite_evaluations",
			EntityID:       eval.ID.String(),
			Details: map[string]any{
				"seal":          seal,
				"decision":      eval.Decision,
				"deposit_order": eval.DepositOrderNumber,
			},
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Récépissé officiel de recevabilité DM visé et scellé avec succès (ES-REG-03-DM)",
		"seal":    seal,
	})
}

// ListDMRegularizations renvoie la liste des DM sous mesures transitoires
func (h *DMHandler) ListDMRegularizations(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	statusFilter := r.URL.Query().Get("status")
	list, err := h.dmRepo.ListDMRegularizations(r.Context(), user.OrganizationID, statusFilter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"regularizations": list,
		"count":           len(list),
	})
}

// GetDMRegularizationStats compile les indicateurs KPIs pour le cockpit
func (h *DMHandler) GetDMRegularizationStats(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	stats, err := h.dmRepo.GetDMRegularizationStats(r.Context(), user.OrganizationID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"stats":   stats,
	})
}

// CreateDMRegularization enregistre un DM sous mesures transitoires
func (h *DMHandler) CreateDMRegularization(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var reg domain.DMRegularization
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide: "+err.Error())
		return
	}

	reg.OrganizationID = user.OrganizationID
	if err := reg.Validate(); err != nil {
		writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if err := h.dmRepo.CreateDMRegularization(r.Context(), &reg); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "DM_REGULARIZATION_CREATE",
			EntityType:     "dm_regularizations",
			EntityID:       reg.ID.String(),
			Details: map[string]any{
				"code":   reg.RegularizationCode,
				"status": reg.Status,
			},
		})
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success":        true,
		"message":        "Dossier de régularisation créé avec succès",
		"regularization": reg,
	})
}

// UpdateDMRegularization met à jour les informations d'un dossier transitoire
func (h *DMHandler) UpdateDMRegularization(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	regID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant régularisation invalide")
		return
	}

	var req domain.DMRegularization
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps de requête invalide: "+err.Error())
		return
	}

	req.ID = regID
	req.OrganizationID = user.OrganizationID

	if err := h.dmRepo.UpdateDMRegularization(r.Context(), &req); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "DM_REGULARIZATION_UPDATE",
			EntityType:     "dm_regularizations",
			EntityID:       regID.String(),
			Details: map[string]any{
				"status": req.Status,
			},
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Dossier de régularisation mis à jour avec succès",
	})
}

// ExportDMCanevas génère le classeur officiel CANEVAS-DECLARATION-DM-NON-HOMOLOGUES.xlsx (format CSV/XLSX conforme)
func (h *DMHandler) ExportDMCanevas(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	rows, err := h.dmRepo.GetDMCanevasExportRows(r.Context(), user.OrganizationID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur génération canevas: "+err.Error())
		return
	}

	filename := fmt.Sprintf("CANEVAS-DECLARATION-DM-NON-HOMOLOGUES_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	// BOM UTF-8 indispensable pour qu'Excel s'ouvre proprement
	_, _ = w.Write([]byte("\xef\xbb\xbf"))
	writer := csv.NewWriter(w)
	writer.Comma = ';'
	for _, row := range rows {
		_ = writer.Write(row)
	}
	writer.Flush()
}
