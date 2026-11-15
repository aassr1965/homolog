package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// SubmissionHandler gère l'assistant de validation pré-dépôt et le générateur de manifeste E-Services ANPP
type SubmissionHandler struct {
	dossierRepo  *repository.DossierRepository
	feeRepo      *repository.FeeRepository
	pharmaRepo   *repository.PharmaRepository
	auditService *audit.Service
}

// NewSubmissionHandler initialise le gestionnaire de pré-soumission E-Services ANPP
func NewSubmissionHandler(
	dossierRepo *repository.DossierRepository,
	feeRepo *repository.FeeRepository,
	pharmaRepo *repository.PharmaRepository,
	auditService *audit.Service,
) *SubmissionHandler {
	return &SubmissionHandler{
		dossierRepo:  dossierRepo,
		feeRepo:      feeRepo,
		pharmaRepo:   pharmaRepo,
		auditService: auditService,
	}
}

// RegisterRoutes enregistre les endpoints de l'assistant de télé-soumission E-Services ANPP
func (h *SubmissionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/dossiers/{id}", func(r chi.Router) {
		r.Get("/pre-submission-audit", h.GetPreSubmissionAudit)
		r.Get("/eservices-manifest", h.GetEservicesManifest)
		r.Post("/lock-pre-submission", h.LockPreSubmission)
	})
}

// GetPreSubmissionAudit audite la complétude et l'éligibilité d'un dossier avant soumission officielle
func (h *SubmissionHandler) GetPreSubmissionAudit(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	// 1. Récupérer le dossier complet avec ses items de checklist et documents GED
	dossier, err := h.dossierRepo.GetDossierByID(r.Context(), user.OrganizationID, dossierID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Dossier réglementaire introuvable")
			return
		}
		log.Printf("[ERROR] GetDossierByID %s: %v", dossierID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger le dossier")
		return
	}

	// 2. Récupérer les bordereaux financiers associés (BV1/BV2)
	slips, err := h.feeRepo.ListPaymentSlipsByDossier(r.Context(), user.OrganizationID.String(), dossierID.String())
	if err != nil {
		log.Printf("[WARN] ListPaymentSlipsByDossier: %v", err)
		slips = []domain.ANPPPaymentSlip{}
	}

	// 3. Récupérer l'établissement demandeur
	var est *domain.Establishment
	estList, _, err := h.pharmaRepo.ListEstablishments(r.Context(), user.OrganizationID, 1, 0)
	if err == nil && len(estList) > 0 {
		est = &estList[0]
	}

	// 4. Évaluer l'audit de complétude réglementaire
	report := domain.EvaluatePreSubmission(dossier, slips, est, user.Email)

	// 5. Consigner dans l'Audit Trail BPF
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionPreSubmissionAudit,
		EntityType:     "dossier",
		EntityID:       dossierID.String(),
		Details: map[string]any{
			"dossier_code":       dossier.DossierCode,
			"is_eligible":        report.IsEligible,
			"completeness_score": report.CompletenessScore,
			"blockers_count":     report.CriticalBlockersCount,
			"warnings_count":     report.WarningsCount,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, report)
}

// GetEservicesManifest génère le manifeste réglementaire numérique E-Services ANPP (JSON ou XML)
func (h *SubmissionHandler) GetEservicesManifest(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	dossier, err := h.dossierRepo.GetDossierByID(r.Context(), user.OrganizationID, dossierID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Dossier réglementaire introuvable")
		return
	}

	slips, _ := h.feeRepo.ListPaymentSlipsByDossier(r.Context(), user.OrganizationID.String(), dossierID.String())

	var est *domain.Establishment
	estList, _, err := h.pharmaRepo.ListEstablishments(r.Context(), user.OrganizationID, 1, 0)
	if err == nil && len(estList) > 0 {
		est = &estList[0]
	}

	manifest, err := domain.BuildANPPEservicesManifest(dossier, slips, est, user.Email)
	if err != nil {
		log.Printf("[ERROR] BuildANPPEservicesManifest: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Échec de génération du manifeste numérique")
		return
	}

	// Consigner l'exportation du manifeste
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionManifestExport,
		EntityType:     "dossier",
		EntityID:       dossierID.String(),
		Details: map[string]any{
			"dossier_code":    dossier.DossierCode,
			"manifest_digest": manifest.IntegritySeal.ManifestDigestHex,
			"format":          r.URL.Query().Get("format"),
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "xml" {
		xmlBytes, err := manifest.ToXML()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Erreur génération XML")
			return
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="ANPP_Manifest_%s.xml"`, dossier.DossierCode))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(xmlBytes)
		return
	}

	// Défaut : JSON
	jsonBytes, err := manifest.ToJSON()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur génération JSON")
		return
	}

	if r.URL.Query().Get("download") == "true" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="ANPP_Manifest_%s.json"`, dossier.DossierCode))
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonBytes)
}

// LockPreSubmissionPayload payload optionnel pour verrouiller et certifier le dossier
type LockPreSubmissionPayload struct {
	SubmissionNotes *string `json:"submission_notes,omitempty"`
}

// LockPreSubmission verrouille le dossier et scelle la certification pré-dépôt E-Services
func (h *SubmissionHandler) LockPreSubmission(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	// Réservé exclusivement au Pharmacien Directeur Technique ou rôle autorisé
	if user.Role != string(domain.RoleDirecteurTechnique) && user.Role != string(domain.RoleResponsableAffairesReg) {
		writeJSONError(w, http.StatusForbidden, "Seul le Pharmacien Directeur Technique (1.A.7) ou le Responsable Réglementaire peut certifier et verrouiller la pré-soumission")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	dossier, err := h.dossierRepo.GetDossierByID(r.Context(), user.OrganizationID, dossierID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Dossier réglementaire introuvable")
		return
	}

	slips, _ := h.feeRepo.ListPaymentSlipsByDossier(r.Context(), user.OrganizationID.String(), dossierID.String())

	var est *domain.Establishment
	estList, _, err := h.pharmaRepo.ListEstablishments(r.Context(), user.OrganizationID, 1, 0)
	if err == nil && len(estList) > 0 {
		est = &estList[0]
	}

	// Vérification stricte d'éligibilité
	report := domain.EvaluatePreSubmission(dossier, slips, est, user.Email)
	if !report.IsEligible {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
			"error":          "Le dossier n'est pas éligible au télé-dépôt E-Services ANPP en raison de bloqueurs critiques non résolus.",
			"blockers_count": report.CriticalBlockersCount,
			"checks":         report.Checks,
		})
		return
	}

	var payload LockPreSubmissionPayload
	_ = json.NewDecoder(r.Body).Decode(&payload)

	// Génération du manifeste scellé
	manifest, err := domain.BuildANPPEservicesManifest(dossier, slips, est, user.Email)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Impossible de générer le sceau numérique du manifeste")
		return
	}

	// Mise à jour du statut du dossier vers valide_interne
	notes := "Certifié conforme et prêt pour télé-dépôt E-Services ANPP"
	if payload.SubmissionNotes != nil && *payload.SubmissionNotes != "" {
		notes = fmt.Sprintf("%s — %s", notes, *payload.SubmissionNotes)
	}

	updatedDossier, err := h.dossierRepo.UpdateDossierStatus(
		r.Context(),
		user.OrganizationID,
		dossierID,
		domain.DossierStatusValidatedInternal,
		&notes,
		nil, nil, nil, nil, nil,
	)
	if err != nil {
		log.Printf("[ERROR] UpdateDossierStatus: %v", err)
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("Erreur lors de la transition d'état : %v", err))
		return
	}

	// Enregistrement de l'action dans l'Audit Trail
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionPreSubmissionLock,
		EntityType:     "dossier",
		EntityID:       dossierID.String(),
		Details: map[string]any{
			"dossier_code":       dossier.DossierCode,
			"manifest_digest":    manifest.IntegritySeal.ManifestDigestHex,
			"previous_status":    dossier.Status,
			"new_status":         domain.DossierStatusValidatedInternal,
			"completeness_score": report.CompletenessScore,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"success":         true,
		"message":         "Dossier officiellement certifié et verrouillé pour le télé-dépôt sur le portail E-Services de l'ANPP.",
		"manifest_digest": manifest.IntegritySeal.ManifestDigestHex,
		"dossier":         updatedDossier,
		"manifest":        manifest,
	})
}
