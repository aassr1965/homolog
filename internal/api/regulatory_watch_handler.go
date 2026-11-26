package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// RegulatoryWatchHandler gère le corpus juridique pharmaceutique algérien et la veille réglementaire
type RegulatoryWatchHandler struct {
	repo         *repository.RegulatoryWatchRepository
	auditService *audit.Service
}

// NewRegulatoryWatchHandler instancie le gestionnaire de veille réglementaire
func NewRegulatoryWatchHandler(
	repo *repository.RegulatoryWatchRepository,
	auditService *audit.Service,
) *RegulatoryWatchHandler {
	return &RegulatoryWatchHandler{
		repo:         repo,
		auditService: auditService,
	}
}

// ListRegulatoryWatch retourne les textes législatifs et réglementaires
func (h *RegulatoryWatchHandler) ListRegulatoryWatch(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	category := r.URL.Query().Get("category")
	impact := r.URL.Query().Get("impact")
	status := r.URL.Query().Get("status")

	items, err := h.repo.ListRegulatoryWatchItems(r.Context(), user.OrganizationID, category, impact, status)
	if err != nil {
		log.Printf("[ERROR] ListRegulatoryWatch: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Erreur lecture veille réglementaire")
		return
	}

	// Décompte par niveau d'impact
	critiqueCount := 0
	majeurCount := 0
	actionRequiredCount := 0
	for _, it := range items {
		if it.ImpactLevel == domain.ImpactLevelCritique {
			critiqueCount++
		} else if it.ImpactLevel == domain.ImpactLevelMajeur {
			majeurCount++
		}
		if it.RequiresActionPlan() {
			actionRequiredCount++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":                 items,
		"total":                 len(items),
		"critique_count":        critiqueCount,
		"majeur_count":          majeurCount,
		"action_required_count": actionRequiredCount,
	})
}

// CreateRegulatoryWatchItemRequest payload de création d'un texte juridique
type CreateRegulatoryWatchItemRequest struct {
	TitleFR          string                            `json:"title_fr"`
	TitleAR          string                            `json:"title_ar,omitempty"`
	LegalReference   string                            `json:"legal_reference"`
	JORADPNumber     string                            `json:"joradp_number,omitempty"`
	PublishedDate    string                            `json:"published_date"` // YYYY-MM-DD
	EffectiveDate    string                            `json:"effective_date"` // YYYY-MM-DD
	Category         domain.RegulatoryCategory         `json:"category"`
	ImpactLevel      domain.ImpactLevel                `json:"impact_level"`
	SummaryFR        string                            `json:"summary_fr"`
	ImpactAnalysis   string                            `json:"impact_analysis,omitempty"`
	ComplianceStatus domain.RegulatoryComplianceStatus `json:"compliance_status"`
	SourceURL        string                            `json:"source_url,omitempty"`
}

// CreateRegulatoryWatchItem enregistre un nouveau texte
func (h *RegulatoryWatchHandler) CreateRegulatoryWatchItem(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var req CreateRegulatoryWatchItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	pubDate, err := time.Parse("2006-01-02", req.PublishedDate)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Date de publication invalide (attendu: YYYY-MM-DD)")
		return
	}
	effDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		effDate = pubDate
	}

	if req.Category == "" {
		req.Category = domain.RegCatEnregistrementAMM
	}
	if req.ImpactLevel == "" {
		req.ImpactLevel = domain.ImpactLevelMajeur
	}
	if req.ComplianceStatus == "" {
		req.ComplianceStatus = domain.ComplianceStatusEnVigueur
	}

	item := &domain.RegulatoryWatchItem{
		OrganizationID:   user.OrganizationID,
		TitleFR:          req.TitleFR,
		TitleAR:          req.TitleAR,
		LegalReference:   req.LegalReference,
		JORADPNumber:     req.JORADPNumber,
		PublishedDate:    pubDate,
		EffectiveDate:    effDate,
		Category:         req.Category,
		ImpactLevel:      req.ImpactLevel,
		SummaryFR:        req.SummaryFR,
		ImpactAnalysis:   req.ImpactAnalysis,
		ComplianceStatus: req.ComplianceStatus,
		SourceURL:        req.SourceURL,
		CreatedBy:        &user.UserID,
	}

	created, err := h.repo.CreateRegulatoryWatchItem(r.Context(), item)
	if err != nil {
		log.Printf("[ERROR] CreateRegulatoryWatchItem: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible d'enregistrer le texte réglementaire")
		return
	}

	// Audit Trail BPF
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionRegulatoryWatchCreate,
		EntityType:     "regulatory_watch",
		EntityID:       created.ID.String(),
		Details: map[string]any{
			"legal_ref":    created.LegalReference,
			"title_fr":     created.TitleFR,
			"category":     created.Category,
			"impact_level": created.ImpactLevel,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, created)
}

// GetRegulatoryWatchItem retourne le détail d'un texte légal
func (h *RegulatoryWatchHandler) GetRegulatoryWatchItem(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	item, err := h.repo.GetRegulatoryWatchItem(r.Context(), user.OrganizationID, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Texte juridique introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur lecture texte juridique")
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// UpdateRegulatoryWatchAnalysisRequest mise à jour de l'analyse d'impact
type UpdateRegulatoryWatchAnalysisRequest struct {
	ImpactAnalysis   string                            `json:"impact_analysis"`
	ComplianceStatus domain.RegulatoryComplianceStatus `json:"compliance_status"`
}

// UpdateRegulatoryWatchAnalysis met à jour l'analyse d'impact d'un texte
func (h *RegulatoryWatchHandler) UpdateRegulatoryWatchAnalysis(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant invalide")
		return
	}

	var req UpdateRegulatoryWatchAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	updated, err := h.repo.UpdateRegulatoryWatchAnalysis(r.Context(), user.OrganizationID, itemID, req.ImpactAnalysis, req.ComplianceStatus)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Texte juridique introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur mise à jour analyse")
		return
	}

	// Audit Trail BPF
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionRegulatoryWatchUpdate,
		EntityType:     "regulatory_watch",
		EntityID:       itemID.String(),
		Details: map[string]any{
			"legal_ref":         updated.LegalReference,
			"compliance_status": updated.ComplianceStatus,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, updated)
}

// DownloadRegulatoryWatchPDF sert le fichier PDF complet stocké en base de données (BYTEA) pour une guideline ANPP.
func (h *RegulatoryWatchHandler) DownloadRegulatoryWatchPDF(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de texte invalide")
		return
	}

	pdfBytes, filename, err := h.repo.GetRegulatoryWatchItemPDF(r.Context(), user.OrganizationID, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Document PDF non disponible pour ce texte")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur lecture PDF réglementaire")
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	_, _ = w.Write(pdfBytes)
}
