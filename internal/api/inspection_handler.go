package api

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
	"homolog/internal/storage"
)

// InspectionHandler gère les observations d'inspection BPF (LD-DVIV04/V01), les plans CAPA
// ainsi que le statut réglementaire post-décision (actif, suspendu, retiré, refusé) des produits.
type InspectionHandler struct {
	pool            *pgxpool.Pool
	auditSvc        *audit.Service
	gedRepo         *repository.GEDRepository
	storageProvider storage.StorageProvider
}

// NewInspectionHandler initialise le handler Inspections/CAPA & Statuts DE/DH.
func NewInspectionHandler(pool *pgxpool.Pool, auditSvc *audit.Service, gedRepo *repository.GEDRepository, storageProvider storage.StorageProvider) *InspectionHandler {
	return &InspectionHandler{
		pool:            pool,
		auditSvc:        auditSvc,
		gedRepo:         gedRepo,
		storageProvider: storageProvider,
	}
}

// ListFindings retourne les observations d'inspection et CAPA de l'organisation.
func (h *InspectionHandler) ListFindings(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT id, organization_id, inspection_reference, inspection_date, inspector_authority,
		       severity, description, COALESCE(capa_plan, ''), COALESCE(responsible_name, ''),
		       due_date, status, ged_report_document_id, created_by, closed_by, closed_at, created_at, updated_at
		FROM inspection_findings
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, user.OrganizationID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur chargement observations d'inspection")
		return
	}
	defer rows.Close()

	findings := make([]domain.InspectionFinding, 0)
	for rows.Next() {
		var it domain.InspectionFinding
		if rows.Scan(
			&it.ID, &it.OrganizationID, &it.InspectionReference, &it.InspectionDate, &it.InspectorAuthority,
			&it.Severity, &it.Description, &it.CAPAPlan, &it.ResponsibleName,
			&it.DueDate, &it.Status, &it.GEDReportDocumentID, &it.CreatedBy, &it.ClosedBy, &it.ClosedAt, &it.CreatedAt, &it.UpdatedAt,
		) == nil {
			findings = append(findings, it)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"findings": findings,
		"count":    len(findings),
	})
}

// CreateFinding enregistre une observation d'inspection BPF avec classification officielle (CRITIQUE, MAJEUR, AUTRE).
func (h *InspectionHandler) CreateFinding(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var req struct {
		InspectionReference string `json:"inspection_reference"`
		InspectionDate      string `json:"inspection_date"`
		InspectorAuthority  string `json:"inspector_authority"`
		Severity            string `json:"severity"`
		Description         string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}
	if strings.TrimSpace(req.Description) == "" {
		writeJSONError(w, http.StatusBadRequest, "La description de l'écart est obligatoire")
		return
	}

	sev := strings.ToUpper(strings.TrimSpace(req.Severity))
	if sev != "CRITIQUE" && sev != "MAJEUR" && sev != "AUTRE" {
		sev = "MAJEUR"
	}
	ref := strings.TrimSpace(req.InspectionReference)
	if ref == "" {
		ref = fmt.Sprintf("INSP-%s-%s", time.Now().Format("20060102"), strings.ToUpper(uuid.NewString()[:4]))
	}
	authName := strings.TrimSpace(req.InspectorAuthority)
	if authName == "" {
		authName = "ANPP (DVTIV)"
	}
	inspDate := time.Now().UTC()
	if strings.TrimSpace(req.InspectionDate) != "" {
		if parsed, err := time.Parse("2006-01-02", strings.TrimSpace(req.InspectionDate)); err == nil {
			inspDate = parsed
		}
	}

	id := uuid.New()
	_, err := h.pool.Exec(r.Context(), `
		INSERT INTO inspection_findings (
			id, organization_id, inspection_reference, inspection_date, inspector_authority,
			severity, description, status, created_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,'ouvert',$8)
	`, id, user.OrganizationID, ref, inspDate, authName, sev, strings.TrimSpace(req.Description), user.UserID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur enregistrement observation d'inspection")
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "INSPECTION_FINDING_CREATED",
			EntityType:     "inspection_finding",
			EntityID:       ref,
			Details: map[string]any{
				"finding_id":          id.String(),
				"severity":            sev,
				"inspector_authority": authName,
			},
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":                   id,
		"inspection_reference": ref,
		"severity":             sev,
		"status":               "ouvert",
	})
}

// SubmitCAPA enregistre le plan d'actions correctives et préventives (CAPA) pour un écart d'inspection.
func (h *InspectionHandler) SubmitCAPA(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	findingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'observation invalide")
		return
	}

	var req struct {
		CAPAPlan        string `json:"capa_plan"`
		ResponsibleName string `json:"responsible_name"`
		DueDate         string `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}
	if strings.TrimSpace(req.CAPAPlan) == "" {
		writeJSONError(w, http.StatusBadRequest, "Le plan CAPA est obligatoire")
		return
	}

	var dueDate *time.Time
	if strings.TrimSpace(req.DueDate) != "" {
		if parsed, pErr := time.Parse("2006-01-02", strings.TrimSpace(req.DueDate)); pErr == nil {
			dueDate = &parsed
		}
	}

	var ref string
	err = h.pool.QueryRow(r.Context(), `
		UPDATE inspection_findings
		SET capa_plan = $1,
		    responsible_name = $2,
		    due_date = COALESCE($3, due_date),
		    status = CASE WHEN status = 'ouvert' THEN 'capa_soumis' ELSE status END,
		    updated_at = NOW()
		WHERE id = $4 AND organization_id = $5
		RETURNING inspection_reference
	`, strings.TrimSpace(req.CAPAPlan), strings.TrimSpace(req.ResponsibleName), dueDate, findingID, user.OrganizationID).Scan(&ref)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Observation d'inspection introuvable")
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "INSPECTION_CAPA_SUBMITTED",
			EntityType:     "inspection_finding",
			EntityID:       ref,
			Details: map[string]any{
				"finding_id":       findingID.String(),
				"responsible_name": req.ResponsibleName,
			},
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "capa_soumis"})
}

// CloseFinding clôture l'écart d'inspection et archive le rapport CAPA dans la GED (02_04_CAPA_Deviations).
func (h *InspectionHandler) CloseFinding(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	findingID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'observation invalide")
		return
	}

	var it domain.InspectionFinding
	err = h.pool.QueryRow(r.Context(), `
		SELECT id, organization_id, inspection_reference, inspection_date, inspector_authority,
		       severity, description, COALESCE(capa_plan, ''), COALESCE(responsible_name, ''), status
		FROM inspection_findings
		WHERE id = $1 AND organization_id = $2
	`, findingID, user.OrganizationID).Scan(
		&it.ID, &it.OrganizationID, &it.InspectionReference, &it.InspectionDate, &it.InspectorAuthority,
		&it.Severity, &it.Description, &it.CAPAPlan, &it.ResponsibleName, &it.Status,
	)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Observation d'inspection introuvable")
		return
	}

	reportHTML := fmt.Sprintf(`<!DOCTYPE html><html lang="fr"><head><meta charset="UTF-8"><title>Clôture CAPA %s</title></head><body>
<h1>Fiche de Clôture Observation Inspection &amp; CAPA (%s)</h1>
<p><strong>Référence Inspection :</strong> %s | <strong>Autorité :</strong> %s</p>
<p><strong>Classification ANPP :</strong> %s</p>
<p><strong>Description de l'écart :</strong> %s</p>
<p><strong>Plan CAPA exécuté :</strong> %s</p>
<p><strong>Responsable :</strong> %s</p>
<p><strong>Clôturé par :</strong> %s (%s) le %s UTC</p>
</body></html>`,
		html.EscapeString(it.InspectionReference),
		html.EscapeString(it.InspectionReference),
		html.EscapeString(it.InspectionReference),
		html.EscapeString(it.InspectorAuthority),
		html.EscapeString(it.Severity),
		html.EscapeString(it.Description),
		html.EscapeString(it.CAPAPlan),
		html.EscapeString(it.ResponsibleName),
		html.EscapeString(user.Email),
		html.EscapeString(user.Role),
		time.Now().UTC().Format("2006-01-02 15:04:05"),
	)

	var gedDocUUID *uuid.UUID
	var gedDocCode string
	if gedIDStr, code := archiveGeneratedDocInGED(r, user, h.gedRepo, h.storageProvider, h.auditSvc, gedArchiveSpec{
		CodePrefix:      "CAPA-INSP",
		TitleFR:         fmt.Sprintf("Clôture CAPA Inspection %s (%s)", it.InspectionReference, it.Severity),
		TaxonomyPath:    "02_Assurance_Qualite_BPF/02_04_CAPA_Deviations",
		DocumentType:    "CAPA_INSPECTION",
		Department:      "assurance_qualite",
		Tags:            []string{"ANPP", "LD-DVIV04", "CAPA", it.Severity, it.InspectionReference},
		RevisionSummary: "Clôture d'écart d'inspection BPF et vérification CAPA",
		MimeType:        "text/html",
		Filename:        fmt.Sprintf("CAPA-%s.html", it.InspectionReference),
		SourceTag:       "inspection_capa_auto_archive",
	}, []byte(reportHTML)); gedIDStr != "" {
		if parsed, pErr := uuid.Parse(gedIDStr); pErr == nil {
			gedDocUUID = &parsed
			gedDocCode = code
		}
	}

	_, err = h.pool.Exec(r.Context(), `
		UPDATE inspection_findings
		SET status = 'cloture',
		    ged_report_document_id = COALESCE($1, ged_report_document_id),
		    closed_by = $2,
		    closed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $3 AND organization_id = $4
	`, gedDocUUID, user.UserID, findingID, user.OrganizationID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur clôture observation")
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "INSPECTION_FINDING_CLOSED",
			EntityType:     "inspection_finding",
			EntityID:       it.InspectionReference,
			Details: map[string]any{
				"finding_id":      findingID.String(),
				"ged_report_code": gedDocCode,
			},
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":                 "cloture",
		"ged_report_document_id": gedDocUUID,
		"ged_report_code":        gedDocCode,
	})
}

// UpdateProductRegulatoryStatus met à jour le statut post-décision DE/DH (actif, suspendu, retire_temporaire, retire_definitif, refuse).
func (h *InspectionHandler) UpdateProductRegulatoryStatus(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	kind := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "kind")))
	if kind != "pp" && kind != "dm" {
		writeJSONError(w, http.StatusBadRequest, "Type de produit invalide (pp ou dm)")
		return
	}
	prodID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant produit invalide")
		return
	}

	var req struct {
		RegulatoryStatus string `json:"regulatory_status"`
		DecisionRef      string `json:"status_decision_ref"`
		EffectiveDate    string `json:"status_effective_date"`
		Reason           string `json:"status_reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	st := strings.ToLower(strings.TrimSpace(req.RegulatoryStatus))
	switch st {
	case "actif", "suspendu", "retire_temporaire", "retire_definitif", "refuse":
	default:
		writeJSONError(w, http.StatusBadRequest, "Statut réglementaire invalide")
		return
	}

	var effDate *time.Time
	if strings.TrimSpace(req.EffectiveDate) != "" {
		if parsed, pErr := time.Parse("2006-01-02", strings.TrimSpace(req.EffectiveDate)); pErr == nil {
			effDate = &parsed
		}
	}

	table := "medicines"
	if kind == "dm" {
		table = "medical_devices"
	}

	tag, err := h.pool.Exec(r.Context(), fmt.Sprintf(`
		UPDATE %s
		SET regulatory_status = $1,
		    status_decision_ref = $2,
		    status_effective_date = $3,
		    status_reason = $4,
		    updated_at = NOW()
		WHERE id = $5 AND organization_id = $6
	`, table), st, strings.TrimSpace(req.DecisionRef), effDate, strings.TrimSpace(req.Reason), prodID, user.OrganizationID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur mise à jour statut réglementaire")
		return
	}
	if tag.RowsAffected() == 0 {
		writeJSONError(w, http.StatusNotFound, "Produit introuvable")
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "PRODUCT_REGULATORY_STATUS_UPDATED",
			EntityType:     table,
			EntityID:       prodID.String(),
			Details: map[string]any{
				"regulatory_status":   st,
				"status_decision_ref": req.DecisionRef,
				"status_reason":       req.Reason,
			},
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":                prodID,
		"product_kind":      kind,
		"regulatory_status": st,
	})
}
