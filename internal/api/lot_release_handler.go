package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/reporting"
	"homolog/internal/repository"
	"homolog/internal/storage"
)

// LotReleaseHandler gère les bordereaux de libération de lots importés (DCE) et la fiche DVTIV PDT/QPPV.
type LotReleaseHandler struct {
	pool            *pgxpool.Pool
	auditSvc        *audit.Service
	gedRepo         *repository.GEDRepository
	corpRepo        *repository.CorporateRepository
	storageProvider storage.StorageProvider
}

// NewLotReleaseHandler initialise le handler DCE / DVTIV.
func NewLotReleaseHandler(pool *pgxpool.Pool, auditSvc *audit.Service, gedRepo *repository.GEDRepository, corpRepo *repository.CorporateRepository, storageProvider storage.StorageProvider) *LotReleaseHandler {
	return &LotReleaseHandler{
		pool:            pool,
		auditSvc:        auditSvc,
		gedRepo:         gedRepo,
		corpRepo:        corpRepo,
		storageProvider: storageProvider,
	}
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
	if gedDocID, gedDocCode := archiveGeneratedDocInGED(r, user, h.gedRepo, h.storageProvider, h.auditSvc, gedArchiveSpec{
		CodePrefix:      "DCE-LOT",
		TitleFR:         fmt.Sprintf("%s — Lot %s (%s)", prefix, it.LotNumber, it.CommercialName),
		TaxonomyPath:    "02_Assurance_Qualite_BPF/02_02_Certificats_BPF_ISO",
		DocumentType:    "FORMULAIRE_ANPP",
		Department:      "assurance_qualite",
		Tags:            []string{"ANPP", "DCE", "LiberationLot", it.LotNumber},
		RevisionSummary: "Bordereau / certificat officiel DCE généré — En attente de visa PDT",
		MimeType:        "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		Filename:        filename,
		SourceTag:       "dce_lot_auto_archive",
		MedicalDeviceID: it.MedicalDeviceID,
	}, docBytes); gedDocID != "" {
		w.Header().Set("X-GED-Document-Id", gedDocID)
		w.Header().Set("X-GED-Document-Code", gedDocCode)
	}
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

	var pdt, assistant, qppv *domain.User
	rows, err := h.pool.Query(r.Context(), `
		SELECT first_name, last_name, email, role
		FROM users
		WHERE organization_id = $1 AND is_active = TRUE
		  AND (role IN ('directeur_technique', 'pharmacien_assistant', 'qppv')
		       OR LOWER(email) LIKE '%qppv%' OR LOWER(last_name) LIKE '%qppv%')
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
				} else if (u.Role == "qppv" || strings.Contains(strings.ToLower(u.Email+u.LastName), "qppv")) && qppv == nil {
					cp := u
					qppv = &cp
				}
			}
		}
	}

	// Enrichir depuis establishments (adresse, wilaya, téléphone, email, agrément, PDT)
	var primaryEstID *uuid.UUID
	var eid uuid.UUID
	var eAddr, eWilaya, ePhone, eEmail, eLicense, ePDT string
	if h.pool.QueryRow(r.Context(), `
		SELECT id, COALESCE(address_fr, ''), COALESCE(wilaya, ''), COALESCE(phone, ''),
		       COALESCE(email, ''), COALESCE(miph_license_number, ''), COALESCE(technical_director_name, '')
		FROM establishments
		WHERE organization_id = $1 AND is_active = TRUE
		ORDER BY created_at ASC
		LIMIT 1
	`, user.OrganizationID).Scan(&eid, &eAddr, &eWilaya, &ePhone, &eEmail, &eLicense, &ePDT) == nil {
		primaryEstID = &eid
		if org.Address == "" && eAddr != "" {
			if eWilaya != "" && !strings.Contains(strings.ToLower(eAddr), strings.ToLower(eWilaya)) {
				org.Address = eAddr + ", " + eWilaya
			} else {
				org.Address = eAddr
			}
		}
		if org.Phone == "" && ePhone != "" {
			org.Phone = ePhone
		}
		if org.Email == "" && eEmail != "" {
			org.Email = eEmail
		}
		if org.MIPHLicenseNumber == "" && eLicense != "" {
			org.MIPHLicenseNumber = eLicense
		}
		if org.TechnicalDirectorName == "" && ePDT != "" {
			org.TechnicalDirectorName = ePDT
		}
	}

	// Déduire les statuts (Fabricant, Importateur, Exploitant de DE) depuis les agréments Corporate ou dossiers
	var estTypes []string
	catRows, catErr := h.pool.Query(r.Context(), `
		SELECT DISTINCT category FROM corporate_applications WHERE organization_id = $1
	`, user.OrganizationID)
	if catErr == nil {
		defer catRows.Close()
		for catRows.Next() {
			var cat string
			if catRows.Scan(&cat) == nil && cat != "" {
				estTypes = append(estTypes, cat)
			}
		}
	}
	if len(estTypes) == 0 {
		procRows, procErr := h.pool.Query(r.Context(), `
			SELECT DISTINCT procedure_code FROM dossiers WHERE organization_id = $1
		`, user.OrganizationID)
		if procErr == nil {
			defer procRows.Close()
			for procRows.Next() {
				var pc string
				if procRows.Scan(&pc) == nil {
					p := strings.ToUpper(pc)
					if strings.HasSuffix(p, "F") {
						estTypes = append(estTypes, "FABRICATION")
					} else if strings.HasSuffix(p, "I") {
						estTypes = append(estTypes, "IMPORTATION")
					}
				}
			}
		}
		var dmCount int
		_ = h.pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM medical_devices WHERE organization_id = $1`, user.OrganizationID).Scan(&dmCount)
		if dmCount > 0 {
			estTypes = append(estTypes, "EXPLOITANT")
		}
	}

	var marketedCount int
	_ = h.pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM medicines WHERE organization_id = $1`, user.OrganizationID).Scan(&marketedCount)

	var hasPSMF, hasPSMFNational bool
	_ = h.pool.QueryRow(r.Context(), `
		SELECT
			EXISTS(SELECT 1 FROM ged_documents WHERE organization_id = $1 AND (LOWER(title_fr) LIKE '%psmf%' OR LOWER(document_code) LIKE '%psmf%' OR 'PSMF' = ANY(tags))),
			EXISTS(SELECT 1 FROM ged_documents WHERE organization_id = $1 AND (LOWER(title_fr) LIKE '%psmf national%' OR LOWER(title_fr) LIKE '%psmf_national%' OR LOWER(document_code) LIKE '%psmf-nat%' OR 'PSMF_NATIONAL' = ANY(tags)))
	`, user.OrganizationID).Scan(&hasPSMF, &hasPSMFNational)

	docBytes, err := reporting.GenerateDVTIVBasePDTQPPVDOCX(&org, pdt, assistant, qppv, marketedCount, hasPSMF, hasPSMFNational, estTypes...)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur génération formulaire DVTIV: "+err.Error())
		return
	}

	filename := fmt.Sprintf("ANPP-DVTIV-Base-PDT-QPPV_%s.docx", time.Now().Format("2006-01-02"))
	if gedDocID, gedDocCode, linkedCount := h.archiveDVTIVInGEDAndLinkChecklists(r, user, &org, primaryEstID, filename, docBytes); gedDocID != "" {
		w.Header().Set("X-GED-Document-Id", gedDocID)
		w.Header().Set("X-GED-Document-Code", gedDocCode)
		w.Header().Set("X-Checklists-Linked", strconv.Itoa(linkedCount))
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	_, _ = w.Write(docBytes)
}

type gedArchiveSpec struct {
	CodePrefix      string
	TitleFR         string
	TaxonomyPath    string
	DocumentType    string
	Department      string
	Tags            []string
	RevisionSummary string
	MimeType        string
	Filename        string
	SourceTag       string
	EstablishmentID *uuid.UUID
	MedicalDeviceID *uuid.UUID
}

// archiveGeneratedDocInGED sauvegarde un document officiel généré dans le stockage CAS + GED (statut BROUILLON) et trace dans l'Audit Trail.
func archiveGeneratedDocInGED(
	r *http.Request,
	user *auth.SessionUser,
	gedRepo *repository.GEDRepository,
	store storage.StorageProvider,
	auditSvc *audit.Service,
	spec gedArchiveSpec,
	fileBytes []byte,
) (string, string) {
	if gedRepo == nil || store == nil || user == nil || len(fileBytes) == 0 {
		return "", ""
	}

	ctx := r.Context()
	fileSHA256, sizeBytes, storagePath, err := store.Save(ctx, bytes.NewReader(fileBytes))
	if err != nil {
		return "", ""
	}

	docCode := fmt.Sprintf("%s-%s-%s", spec.CodePrefix, time.Now().Format("20060102-150405"), uuid.NewString()[:4])
	doc := &domain.GEDDocumentDetailed{
		ID:              uuid.New(),
		OrganizationID:  user.OrganizationID,
		TaxonomyPath:    spec.TaxonomyPath,
		TitleFR:         spec.TitleFR,
		DocumentCode:    docCode,
		DocumentType:    spec.DocumentType,
		Department:      spec.Department,
		Tags:            spec.Tags,
		Status:          domain.GEDStatusBrouillon,
		CurrentVersion:  "v0.1",
		EstablishmentID: spec.EstablishmentID,
		MedicalDeviceID: spec.MedicalDeviceID,
		CreatedBy:       &user.UserID,
	}
	initialVersion := &domain.GEDVersionDetail{
		ID:               uuid.New(),
		VersionLabel:     "v0.1",
		IsMajorVersion:   false,
		RevisionSummary:  spec.RevisionSummary,
		FileSHA256:       fileSHA256,
		FileSizeBytes:    sizeBytes,
		MimeType:         spec.MimeType,
		OriginalFilename: spec.Filename,
		StoragePath:      storagePath,
		AuthorID:         &user.UserID,
	}

	createdDoc, err := gedRepo.CreateDocument(ctx, doc, initialVersion)
	if err != nil {
		return "", ""
	}

	if auditSvc != nil {
		_, _ = auditSvc.Log(ctx, &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         audit.ActionDocUpload,
			EntityType:     "ged_document",
			EntityID:       createdDoc.DocumentCode,
			Details: map[string]any{
				"document_id":   createdDoc.ID.String(),
				"title_fr":      createdDoc.TitleFR,
				"taxonomy_path": createdDoc.TaxonomyPath,
				"version_label": initialVersion.VersionLabel,
				"file_sha256":   initialVersion.FileSHA256,
				"file_size":     initialVersion.FileSizeBytes,
				"original_file": initialVersion.OriginalFilename,
				"source":        spec.SourceTag,
			},
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}
	return createdDoc.ID.String(), createdDoc.DocumentCode
}

// archiveDVTIVInGEDAndLinkChecklists archive la fiche DVTIV générée en GED (statut BROUILLON)
// et la pré-rattache aux items de checklist concernés (ex: EXP.PV.01) sans les valider automatiquement.
func (h *LotReleaseHandler) archiveDVTIVInGEDAndLinkChecklists(
	r *http.Request,
	user *auth.SessionUser,
	org *domain.Organization,
	estID *uuid.UUID,
	filename string,
	docBytes []byte,
) (string, string, int) {
	titleFR := "Fiche Officielle ANPP DVTIV — Base PDT & QPPV"
	if org != nil && strings.TrimSpace(org.LegalName) != "" {
		titleFR = fmt.Sprintf("%s (%s)", titleFR, strings.TrimSpace(org.LegalName))
	}
	docIDStr, docCode := archiveGeneratedDocInGED(r, user, h.gedRepo, h.storageProvider, h.auditSvc, gedArchiveSpec{
		CodePrefix:      "DVTIV",
		TitleFR:         titleFR,
		TaxonomyPath:    "04_Vigilance_Post_Market/04_01_Pharmacovigilance",
		DocumentType:    "FORMULAIRE_ANPP",
		Department:      "pharmacovigilance",
		Tags:            []string{"ANPP", "DVTIV", "PDT", "QPPV", "Pharmacovigilance"},
		RevisionSummary: "Génération automatique depuis le Parapheur DT — En attente de signature (PDT/QPPV) et validation qualité",
		MimeType:        "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		Filename:        filename,
		SourceTag:       "dvtiv_auto_archive",
		EstablishmentID: estID,
	}, docBytes)
	if docIDStr == "" {
		return "", "", 0
	}
	docUUID, _ := uuid.Parse(docIDStr)
	linkedCount := h.linkDVTIVToChecklists(r.Context(), user.OrganizationID, docUUID)
	return docIDStr, docCode, linkedCount
}

func (h *LotReleaseHandler) linkDVTIVToChecklists(ctx context.Context, orgID, gedDocID uuid.UUID) int {
	if h.corpRepo == nil || h.pool == nil {
		return 0
	}

	rows, err := h.pool.Query(ctx, `
		SELECT ci.application_id, ci.id
		FROM corporate_checklist_items ci
		JOIN corporate_applications ca ON ca.id = ci.application_id
		WHERE ca.organization_id = $1
		  AND ca.status NOT IN ('rejete', 'agrement_octroye')
		  AND ci.item_code IN ('EXP.PV.01')
		  AND ci.status IN ('manquant', 'rattache', 'rejete')
	`, orgID)
	if err != nil {
		return 0
	}
	defer rows.Close()

	type target struct {
		appID  uuid.UUID
		itemID uuid.UUID
	}
	var targets []target
	for rows.Next() {
		var t target
		if rows.Scan(&t.appID, &t.itemID) == nil {
			targets = append(targets, t)
		}
	}

	linkedCount := 0
	for _, t := range targets {
		if err := h.corpRepo.AttachDocumentToItem(ctx, orgID, t.appID, t.itemID, gedDocID); err == nil {
			_, _ = h.pool.Exec(ctx, `
				UPDATE corporate_checklist_items
				SET notes = $1, updated_at = NOW()
				WHERE id = $2 AND application_id = $3
			`, "Fiche DVTIV PDT/QPPV générée automatiquement — En attente de signature PDT/QPPV et validation qualité.", t.itemID, t.appID)
			linkedCount++
		}
	}
	return linkedCount
}

// DownloadDCEForm génère et télécharge les formulaires et canevas officiels DCE de l'ANPP pré-remplis.
func (h *LotReleaseHandler) DownloadDCEForm(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	formType := strings.ToLower(strings.TrimSpace(chi.URLParam(r, "formType")))
	org := h.loadOrganizationEnriched(r.Context(), user.OrganizationID)
	lots := h.loadOrganizationLotReleases(r.Context(), user.OrganizationID, 20)

	var (
		outBytes    []byte
		err         error
		filename    string
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	)

	switch formType {
	case "composition-dossier":
		outBytes, err = reporting.GenerateDCECompositionDossierDOCX()
		filename = "ANPP-DCE-Composition-Dossier-Liberation.docx"
	case "canevas-certificat":
		var selected *domain.ImportLotRelease
		if lotIDStr := strings.TrimSpace(r.URL.Query().Get("lot_id")); lotIDStr != "" {
			if lid, parseErr := uuid.Parse(lotIDStr); parseErr == nil {
				for _, l := range lots {
					if l.ID == lid {
						selected = l
						break
					}
				}
			}
		} else if len(lots) > 0 {
			selected = lots[0]
		}
		outBytes, err = reporting.GenerateDCECanevasCertificatLiberationDOCX(org, selected)
		filename = "ANPP-DCE-Canevas-Certificat-Liberation.docx"
	case "declaration-quotidienne-pp":
		outBytes, err = reporting.GenerateDCEDeclarationQuotidienneDOCX("pp", org, lots)
		filename = fmt.Sprintf("ANPP-DCE-Declaration-Quotidienne-PP_%s.docx", time.Now().Format("2006-01-02"))
	case "declaration-quotidienne-dm":
		outBytes, err = reporting.GenerateDCEDeclarationQuotidienneDOCX("dm", org, lots)
		filename = fmt.Sprintf("ANPP-DCE-Declaration-Quotidienne-DM_%s.docx", time.Now().Format("2006-01-02"))
	case "declaration-journaliere-pp":
		outBytes, err = reporting.GenerateDCEDeclarationJournaliereXLSX("pp", org, lots)
		filename = fmt.Sprintf("ANPP-DCE-Declaration-Journaliere-PP_%s.xlsx", time.Now().Format("2006-01-02"))
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "declaration-journaliere-dm":
		outBytes, err = reporting.GenerateDCEDeclarationJournaliereXLSX("dm", org, lots)
		filename = fmt.Sprintf("ANPP-DDM-Declaration-Journaliere-DM_%s.xlsx", time.Now().Format("2006-01-02"))
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "declaration-journaliere-pp-dm":
		outBytes, err = reporting.GenerateDCEDeclarationJournaliereXLSX("pp_dm", org, lots)
		filename = fmt.Sprintf("ANPP-DCE-Declaration-Journaliere-PP-DM_%s.xlsx", time.Now().Format("2006-01-02"))
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "ca-sr-secondaire":
		outBytes, err = reporting.GenerateDCETemplateCASRSecondaireDOCX(org)
		filename = "ANPP-DCE-Prototype-CA-SR-Secondaire.docx"
	default:
		writeJSONError(w, http.StatusBadRequest, "Type de formulaire DCE inconnu")
		return
	}

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur génération formulaire DCE: "+err.Error())
		return
	}

	// Archiver en GED les certificats et déclarations réglementaires (pas les simples modèles/aides-mémoire)
	if formType != "composition-dossier" && formType != "ca-sr-secondaire" {
		if gedDocID, gedDocCode := archiveGeneratedDocInGED(r, user, h.gedRepo, h.storageProvider, h.auditSvc, gedArchiveSpec{
			CodePrefix:      "DCE-DECL",
			TitleFR:         fmt.Sprintf("Déclaration / Certificat DCE (%s)", formType),
			TaxonomyPath:    "01_Affaires_Reglementaires/01_06_Correspondances_ANPP",
			DocumentType:    "FORMULAIRE_ANPP",
			Department:      "affaires_reglementaires",
			Tags:            []string{"ANPP", "DCE", formType},
			RevisionSummary: "Document DCE généré — En attente de signature / dépôt ANPP",
			MimeType:        contentType,
			Filename:        filename,
			SourceTag:       "dce_form_auto_archive",
		}, outBytes); gedDocID != "" {
			w.Header().Set("X-GED-Document-Id", gedDocID)
			w.Header().Set("X-GED-Document-Code", gedDocCode)
		}
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	_, _ = w.Write(outBytes)
}

func (h *LotReleaseHandler) loadOrganizationEnriched(ctx context.Context, orgID uuid.UUID) *domain.Organization {
	var org domain.Organization
	_ = h.pool.QueryRow(ctx, `
		SELECT id, slug, legal_name, COALESCE(miph_license_number, ''), COALESCE(technical_director_name, ''),
		       COALESCE(address, ''), COALESCE(phone, ''), COALESCE(email, '')
		FROM organizations WHERE id = $1
	`, orgID).Scan(&org.ID, &org.Slug, &org.LegalName, &org.MIPHLicenseNumber, &org.TechnicalDirectorName, &org.Address, &org.Phone, &org.Email)

	var eAddr, eWilaya, ePhone, eEmail, eLicense, ePDT string
	if h.pool.QueryRow(ctx, `
		SELECT COALESCE(address_fr, ''), COALESCE(wilaya, ''), COALESCE(phone, ''),
		       COALESCE(email, ''), COALESCE(miph_license_number, ''), COALESCE(technical_director_name, '')
		FROM establishments
		WHERE organization_id = $1 AND is_active = TRUE
		ORDER BY created_at ASC
		LIMIT 1
	`, orgID).Scan(&eAddr, &eWilaya, &ePhone, &eEmail, &eLicense, &ePDT) == nil {
		if org.Address == "" && eAddr != "" {
			if eWilaya != "" && !strings.Contains(strings.ToLower(eAddr), strings.ToLower(eWilaya)) {
				org.Address = eAddr + ", " + eWilaya
			} else {
				org.Address = eAddr
			}
		}
		if org.Phone == "" && ePhone != "" {
			org.Phone = ePhone
		}
		if org.Email == "" && eEmail != "" {
			org.Email = eEmail
		}
		if org.MIPHLicenseNumber == "" && eLicense != "" {
			org.MIPHLicenseNumber = eLicense
		}
		if org.TechnicalDirectorName == "" && ePDT != "" {
			org.TechnicalDirectorName = ePDT
		}
	}
	return &org
}

func (h *LotReleaseHandler) loadOrganizationLotReleases(ctx context.Context, orgID uuid.UUID, limit int) []*domain.ImportLotRelease {
	rows, err := h.pool.Query(ctx, `
		SELECT id, organization_id, product_kind, commercial_name, designation, form_aspect,
		       lot_number, ddf, ddp, sales_units, homologation_or_de_ref, dm_class,
		       production_site, qc_site, release_site, created_at
		FROM import_lot_releases
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, orgID, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var list []*domain.ImportLotRelease
	for rows.Next() {
		var it domain.ImportLotRelease
		if rows.Scan(
			&it.ID, &it.OrganizationID, &it.ProductKind, &it.CommercialName, &it.Designation, &it.FormAspect,
			&it.LotNumber, &it.DDF, &it.DDP, &it.SalesUnits, &it.HomologationOrDERef, &it.DMClass,
			&it.ProductionSite, &it.QCSite, &it.ReleaseSite, &it.CreatedAt,
		) == nil {
			list = append(list, &it)
		}
	}
	return list
}
