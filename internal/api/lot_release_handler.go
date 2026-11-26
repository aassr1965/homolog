package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/reporting"
)

// LotReleaseHandler gère les bordereaux de libération de lots importés (DCE) et la fiche DVTIV PDT/QPPV.
type LotReleaseHandler struct {
	pool     *pgxpool.Pool
	auditSvc *audit.Service
}

// NewLotReleaseHandler initialise le handler DCE / DVTIV.
func NewLotReleaseHandler(pool *pgxpool.Pool, auditSvc *audit.Service) *LotReleaseHandler {
	return &LotReleaseHandler{pool: pool, auditSvc: auditSvc}
}

// ListLotReleases liste les dossiers de libération de lots importés de l'organisation.
func (h *LotReleaseHandler) ListLotReleases(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	kind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind")))
	query := `
		SELECT id, organization_id, product_kind, medical_device_id, medicine_id,
		       commercial_name, designation, form_aspect, lot_number, serial_number,
		       ddf, ddp, sales_units, homologation_or_de_ref, dm_class, holder_or_exploiter,
		       import_program_attached, production_site, qc_site, release_site,
		       customs_decl_number, invoice_number, origin_cert_attached, tech_form_attached,
		       sampling_plan_attached, status, created_by, signed_by_pdt_id, signed_at, created_at, updated_at
		FROM import_lot_releases
		WHERE organization_id = $1
	`
	args := []any{user.OrganizationID}
	if kind == "dm" || kind == "pp" {
		query += ` AND product_kind = $2`
		args = append(args, kind)
	}
	query += ` ORDER BY created_at DESC LIMIT 100`

	rows, err := h.pool.Query(r.Context(), query, args...)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur chargement libérations de lot: "+err.Error())
		return
	}
	defer rows.Close()

	items := make([]domain.ImportLotRelease, 0)
	for rows.Next() {
		var it domain.ImportLotRelease
		if err := rows.Scan(
			&it.ID, &it.OrganizationID, &it.ProductKind, &it.MedicalDeviceID, &it.MedicineID,
			&it.CommercialName, &it.Designation, &it.FormAspect, &it.LotNumber, &it.SerialNumber,
			&it.DDF, &it.DDP, &it.SalesUnits, &it.HomologationOrDERef, &it.DMClass, &it.HolderOrExploiter,
			&it.ImportProgramAttached, &it.ProductionSite, &it.QCSite, &it.ReleaseSite,
			&it.CustomsDeclNumber, &it.InvoiceNumber, &it.OriginCertAttached, &it.TechFormAttached,
			&it.SamplingPlanAttached, &it.Status, &it.CreatedBy, &it.SignedByPDTID, &it.SignedAt, &it.CreatedAt, &it.UpdatedAt,
		); err == nil {
			items = append(items, it)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"lot_releases": items,
		"count":        len(items),
	})
}

// CreateLotRelease crée un bordereau de réception / libération de lot importé (DM ou PP).
func (h *LotReleaseHandler) CreateLotRelease(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var req domain.CreateImportLotReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}
	kind := strings.ToLower(strings.TrimSpace(req.ProductKind))
	if kind != "dm" && kind != "pp" {
		kind = "dm"
	}
	if strings.TrimSpace(req.CommercialName) == "" || strings.TrimSpace(req.LotNumber) == "" {
		writeJSONError(w, http.StatusBadRequest, "La dénomination commerciale et le numéro de lot sont obligatoires")
		return
	}

	var ddf, ddp *time.Time
	if strings.TrimSpace(req.DDF) != "" {
		if t, err := time.Parse("2006-01-02", strings.TrimSpace(req.DDF)); err == nil {
			ddf = &t
		}
	}
	if strings.TrimSpace(req.DDP) != "" {
		if t, err := time.Parse("2006-01-02", strings.TrimSpace(req.DDP)); err == nil {
			ddp = &t
		}
	}

	id := uuid.New()
	_, err := h.pool.Exec(r.Context(), `
		INSERT INTO import_lot_releases (
			id, organization_id, product_kind, medical_device_id, medicine_id,
			commercial_name, designation, form_aspect, lot_number, serial_number,
			ddf, ddp, sales_units, homologation_or_de_ref, dm_class, holder_or_exploiter,
			import_program_attached, production_site, qc_site, release_site,
			customs_decl_number, invoice_number, origin_cert_attached, tech_form_attached,
			sampling_plan_attached, status, created_by
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
			$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			$21,$22,$23,$24,$25,'brouillon',$26
		)
	`, id, user.OrganizationID, kind, req.MedicalDeviceID, req.MedicineID,
		strings.TrimSpace(req.CommercialName), strings.TrimSpace(req.Designation), strings.TrimSpace(req.FormAspect),
		strings.TrimSpace(req.LotNumber), strings.TrimSpace(req.SerialNumber),
		ddf, ddp, strings.TrimSpace(req.SalesUnits), strings.TrimSpace(req.HomologationOrDERef),
		strings.TrimSpace(req.DMClass), strings.TrimSpace(req.HolderOrExploiter),
		req.ImportProgramAttached, strings.TrimSpace(req.ProductionSite), strings.TrimSpace(req.QCSite),
		strings.TrimSpace(req.ReleaseSite), strings.TrimSpace(req.CustomsDeclNumber), strings.TrimSpace(req.InvoiceNumber),
		req.OriginCertAttached, req.TechFormAttached, req.SamplingPlanAttached, user.UserID,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur enregistrement libération de lot: "+err.Error())
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "LOT_RELEASE_CREATED",
			EntityType:     "import_lot_release",
			EntityID:       id.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"product_kind":    kind,
				"commercial_name": req.CommercialName,
				"lot_number":      req.LotNumber,
			},
		})
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":      id.String(),
		"message": "Bordereau de libération de lot enregistré",
	})
}

// DownloadOfficialLotReleaseDOCX génère le formulaire officiel .docx ANPP DCE pré-rempli pour un lot importé.
func (h *LotReleaseHandler) DownloadOfficialLotReleaseDOCX(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	lotID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de lot invalide")
		return
	}

	var it domain.ImportLotRelease
	err = h.pool.QueryRow(r.Context(), `
		SELECT id, organization_id, product_kind, medical_device_id, medicine_id,
		       commercial_name, designation, form_aspect, lot_number, serial_number,
		       ddf, ddp, sales_units, homologation_or_de_ref, dm_class, holder_or_exploiter,
		       import_program_attached, production_site, qc_site, release_site,
		       customs_decl_number, invoice_number, origin_cert_attached, tech_form_attached,
		       sampling_plan_attached, status
		FROM import_lot_releases
		WHERE id = $1 AND organization_id = $2
	`, lotID, user.OrganizationID).Scan(
		&it.ID, &it.OrganizationID, &it.ProductKind, &it.MedicalDeviceID, &it.MedicineID,
		&it.CommercialName, &it.Designation, &it.FormAspect, &it.LotNumber, &it.SerialNumber,
		&it.DDF, &it.DDP, &it.SalesUnits, &it.HomologationOrDERef, &it.DMClass, &it.HolderOrExploiter,
		&it.ImportProgramAttached, &it.ProductionSite, &it.QCSite, &it.ReleaseSite,
		&it.CustomsDeclNumber, &it.InvoiceNumber, &it.OriginCertAttached, &it.TechFormAttached,
		&it.SamplingPlanAttached, &it.Status,
	)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Dossier de libération de lot introuvable")
		return
	}

	var org domain.Organization
	_ = h.pool.QueryRow(r.Context(), `
		SELECT id, slug, legal_name, COALESCE(miph_license_number, ''), COALESCE(technical_director_name, ''),
		       COALESCE(address, ''), COALESCE(phone, ''), COALESCE(email, '')
		FROM organizations WHERE id = $1
	`, user.OrganizationID).Scan(&org.ID, &org.Slug, &org.LegalName, &org.MIPHLicenseNumber, &org.TechnicalDirectorName, &org.Address, &org.Phone, &org.Email)

	docType := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))
	var docBytes []byte
	var prefix string
	if docType == "certificat" && it.ProductKind == "pp" {
		docBytes, err = reporting.GenerateDCECertificatLiberationPPDOCX(&it)
		prefix = "ANPP-DCE-Certificat-Liberation-PP"
	} else if it.ProductKind == "pp" {
		docBytes, err = reporting.GenerateDCEPPLotReceptionDOCX(&org, &it)
		prefix = "ANPP-DCE-Reception-PP-Importes"
	} else {
		docBytes, err = reporting.GenerateDCEDMLotReceptionDOCX(&org, &it)
		prefix = "ANPP-DCE-Reception-DM-Importes"
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur génération formulaire DCE: "+err.Error())
		return
	}

	filename := fmt.Sprintf("%s_%s.docx", prefix, it.LotNumber)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	_, _ = w.Write(docBytes)
}

// DownloadDVTIVPDTQPPVForm génère le formulaire officiel DVTIV « Base de données centralisées des PDT et QPPV.docx ».
func (h *LotReleaseHandler) DownloadDVTIVPDTQPPVForm(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var org domain.Organization
	_ = h.pool.QueryRow(r.Context(), `
		SELECT id, slug, legal_name, COALESCE(miph_license_number, ''), COALESCE(technical_director_name, ''),
		       COALESCE(address, ''), COALESCE(phone, ''), COALESCE(email, '')
		FROM organizations WHERE id = $1
	`, user.OrganizationID).Scan(&org.ID, &org.Slug, &org.LegalName, &org.MIPHLicenseNumber, &org.TechnicalDirectorName, &org.Address, &org.Phone, &org.Email)

	var pdt, assistant *domain.User
	rows, err := h.pool.Query(r.Context(), `
		SELECT first_name, last_name, email, role
		FROM users
		WHERE organization_id = $1 AND is_active = TRUE AND role IN ('directeur_technique', 'pharmacien_assistant')
		ORDER BY created_at ASC
	`, user.OrganizationID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var u domain.User
			if rows.Scan(&u.FirstName, &u.LastName, &u.Email, &u.Role) == nil {
				if u.Role == "directeur_technique" && pdt == nil {
					cp := u
					pdt = &cp
				} else if u.Role == "pharmacien_assistant" && assistant == nil {
					cp := u
					assistant = &cp
				}
			}
		}
	}

	docBytes, err := reporting.GenerateDVTIVBasePDTQPPVDOCX(&org, pdt, assistant, nil)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur génération formulaire DVTIV: "+err.Error())
		return
	}

	filename := fmt.Sprintf("ANPP-DVTIV-Base-PDT-QPPV_%s.docx", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	_, _ = w.Write(docBytes)
}
