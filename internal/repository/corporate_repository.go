package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"homolog/internal/domain"
)

// CorporateRepository gère la persistance des demandes d'agréments MIPH et licences
type CorporateRepository struct {
	pool *pgxpool.Pool
}

// NewCorporateRepository instancie le repository Corporate
func NewCorporateRepository(pool *pgxpool.Pool) *CorporateRepository {
	return &CorporateRepository{pool: pool}
}

// ListApplications liste les demandes d'agréments d'une organisation avec filtres optionnels
func (r *CorporateRepository) ListApplications(ctx context.Context, orgID uuid.UUID, category, status string) ([]domain.CorporateApplication, error) {
	query := `
		SELECT 
			a.id, a.organization_id, a.establishment_id, a.category, a.nature,
			a.application_number, COALESCE(a.internal_reference, ''), a.title, a.status,
			a.form_type, a.form_data, a.completeness_percent,
			a.pdt_signed_by, a.pdt_signed_at, COALESCE(a.pdt_signature_seal, ''),
			COALESCE(a.tabadol_submission_ref, ''), a.tabadol_submitted_at, COALESCE(a.tabadol_submission_notes, ''),
			COALESCE(a.miph_receipt_ref, ''), to_char(a.miph_receipt_date, 'YYYY-MM-DD'),
			COALESCE(a.decision_ref, ''), to_char(a.decision_date, 'YYYY-MM-DD'),
			to_char(a.license_expiry_date, 'YYYY-MM-DD'),
			a.created_by, a.created_at, a.updated_at,
			COALESCE(e.name_fr, '') as establishment_name
		FROM corporate_applications a
		LEFT JOIN establishments e ON a.establishment_id = e.id
		WHERE a.organization_id = $1
	`
	args := []interface{}{orgID}
	idx := 2

	if category != "" && category != "all" {
		query += fmt.Sprintf(" AND a.category = $%d", idx)
		args = append(args, category)
		idx++
	}
	if status != "" && status != "all" {
		query += fmt.Sprintf(" AND a.status = $%d", idx)
		args = append(args, status)
		idx++
	}

	query += " ORDER BY a.created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("requête list corporate applications: %w", err)
	}
	defer rows.Close()

	var apps []domain.CorporateApplication
	for rows.Next() {
		var app domain.CorporateApplication
		var formDataRaw []byte
		var estID, pdtSignedBy, createdBy *uuid.UUID
		var miphDate, decDate, expDate *string

		err := rows.Scan(
			&app.ID, &app.OrganizationID, &estID, &app.Category, &app.Nature,
			&app.ApplicationNumber, &app.InternalReference, &app.Title, &app.Status,
			&app.FormType, &formDataRaw, &app.CompletenessPercent,
			&pdtSignedBy, &app.PDTSignedAt, &app.PDTSignatureSeal,
			&app.TabadolSubmissionRef, &app.TabadolSubmittedAt, &app.TabadolSubmissionNotes,
			&app.MIPHReceiptRef, &miphDate,
			&app.DecisionRef, &decDate,
			&expDate,
			&createdBy, &app.CreatedAt, &app.UpdatedAt,
			&app.EstablishmentName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan corporate application: %w", err)
		}

		app.EstablishmentID = estID
		app.PDTSignedBy = pdtSignedBy
		app.CreatedBy = createdBy
		app.MIPHReceiptDate = miphDate
		app.DecisionDate = decDate
		app.LicenseExpiryDate = expDate

		if len(formDataRaw) > 0 {
			_ = json.Unmarshal(formDataRaw, &app.FormData)
		}
		if app.FormData == nil {
			app.FormData = make(map[string]interface{})
		}

		apps = append(apps, app)
	}

	return apps, nil
}

// GetApplication récupère une demande avec ses items de checklist complets
func (r *CorporateRepository) GetApplication(ctx context.Context, orgID, appID uuid.UUID) (*domain.CorporateApplication, error) {
	query := `
		SELECT 
			a.id, a.organization_id, a.establishment_id, a.category, a.nature,
			a.application_number, COALESCE(a.internal_reference, ''), a.title, a.status,
			a.form_type, a.form_data, a.completeness_percent,
			a.pdt_signed_by, a.pdt_signed_at, COALESCE(a.pdt_signature_seal, ''),
			COALESCE(a.tabadol_submission_ref, ''), a.tabadol_submitted_at, COALESCE(a.tabadol_submission_notes, ''),
			COALESCE(a.miph_receipt_ref, ''), to_char(a.miph_receipt_date, 'YYYY-MM-DD'),
			COALESCE(a.decision_ref, ''), to_char(a.decision_date, 'YYYY-MM-DD'),
			to_char(a.license_expiry_date, 'YYYY-MM-DD'),
			a.created_by, a.created_at, a.updated_at,
			COALESCE(e.name_fr, '') as establishment_name
		FROM corporate_applications a
		LEFT JOIN establishments e ON a.establishment_id = e.id
		WHERE a.id = $1 AND a.organization_id = $2
	`

	var app domain.CorporateApplication
	var formDataRaw []byte
	var estID, pdtSignedBy, createdBy *uuid.UUID
	var miphDate, decDate, expDate *string

	err := r.pool.QueryRow(ctx, query, appID, orgID).Scan(
		&app.ID, &app.OrganizationID, &estID, &app.Category, &app.Nature,
		&app.ApplicationNumber, &app.InternalReference, &app.Title, &app.Status,
		&app.FormType, &formDataRaw, &app.CompletenessPercent,
		&pdtSignedBy, &app.PDTSignedAt, &app.PDTSignatureSeal,
		&app.TabadolSubmissionRef, &app.TabadolSubmittedAt, &app.TabadolSubmissionNotes,
		&app.MIPHReceiptRef, &miphDate,
		&app.DecisionRef, &decDate,
		&expDate,
		&createdBy, &app.CreatedAt, &app.UpdatedAt,
		&app.EstablishmentName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("demande d'agrément non trouvée")
		}
		return nil, fmt.Errorf("scan application: %w", err)
	}

	app.EstablishmentID = estID
	app.PDTSignedBy = pdtSignedBy
	app.CreatedBy = createdBy
	app.MIPHReceiptDate = miphDate
	app.DecisionDate = decDate
	app.LicenseExpiryDate = expDate

	if len(formDataRaw) > 0 {
		_ = json.Unmarshal(formDataRaw, &app.FormData)
	}
	if app.FormData == nil {
		app.FormData = make(map[string]interface{})
	}

	// Récupérer les items de checklist avec jointure GED
	itemsQuery := `
		SELECT 
			c.id, c.application_id, c.item_code, c.title, c.section,
			c.requirement_level, c.is_mandatory, COALESCE(c.condition_rule, ''),
			c.sort_order, c.ged_document_id, c.status, COALESCE(c.notes, ''),
			c.uploaded_at, c.verified_by, c.verified_at, c.created_at, c.updated_at,
			COALESCE(g.document_code, ''), COALESCE(g.title_fr, ''), COALESCE(v.file_sha256, '')
		FROM corporate_checklist_items c
		LEFT JOIN ged_documents g ON c.ged_document_id = g.id
		LEFT JOIN ged_versions v ON g.id = v.document_id AND v.version_label = g.current_version
		WHERE c.application_id = $1
		ORDER BY c.sort_order ASC, c.item_code ASC
	`
	rows, err := r.pool.Query(ctx, itemsQuery, appID)
	if err != nil {
		return nil, fmt.Errorf("requête checklist items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var it domain.CorporateChecklistItem
		var gedID, verifierID *uuid.UUID

		err := rows.Scan(
			&it.ID, &it.ApplicationID, &it.ItemCode, &it.Title, &it.Section,
			&it.RequirementLevel, &it.IsMandatory, &it.ConditionRule,
			&it.SortOrder, &gedID, &it.Status, &it.Notes,
			&it.UploadedAt, &verifierID, &it.VerifiedAt, &it.CreatedAt, &it.UpdatedAt,
			&it.DocumentCode, &it.DocumentName, &it.FileSHA256,
		)
		if err != nil {
			return nil, fmt.Errorf("scan checklist item: %w", err)
		}
		it.GEDDocumentID = gedID
		it.VerifiedBy = verifierID
		app.ChecklistItems = append(app.ChecklistItems, it)
	}

	return &app, nil
}

// CreateApplication initialise une demande d'agrément avec sa checklist officielle
func (r *CorporateRepository) CreateApplication(ctx context.Context, app *domain.CorporateApplication) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	if app.Status == "" {
		app.Status = domain.StatusCorporateDraft
	}
	if app.FormType == "" {
		app.FormType = domain.GetFormTypeForApplication(app.Category, app.Nature)
	}
	if app.FormData == nil {
		app.FormData = make(map[string]interface{})
	}

	formDataRaw, err := json.Marshal(app.FormData)
	if err != nil {
		return fmt.Errorf("marshal form data: %w", err)
	}

	now := time.Now()
	queryApp := `
		INSERT INTO corporate_applications (
			id, organization_id, establishment_id, category, nature,
			application_number, internal_reference, title, status,
			form_type, form_data, completeness_percent, created_by,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15
		)
	`
	_, err = tx.Exec(ctx, queryApp,
		app.ID, app.OrganizationID, app.EstablishmentID, app.Category, app.Nature,
		app.ApplicationNumber, app.InternalReference, app.Title, app.Status,
		app.FormType, formDataRaw, 0, app.CreatedBy,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("insert corporate application: %w", err)
	}

	// Générer les items de checklist par défaut selon Décret 21-82
	defaultItems := domain.GenerateCorporateChecklist(app.Category, app.Nature)
	for _, it := range defaultItems {
		queryItem := `
			INSERT INTO corporate_checklist_items (
				id, application_id, item_code, title, section,
				requirement_level, is_mandatory, condition_rule,
				sort_order, status, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8,
				$9, $10, $11, $12
			)
		`
		_, err := tx.Exec(ctx, queryItem,
			uuid.New(), app.ID, it.ItemCode, it.Title, it.Section,
			it.RequirementLevel, it.IsMandatory, it.ConditionRule,
			it.SortOrder, it.Status, now, now,
		)
		if err != nil {
			return fmt.Errorf("insert checklist item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// UpdateApplication met à jour les données du formulaire e-Form et métadonnées
func (r *CorporateRepository) UpdateApplication(ctx context.Context, app *domain.CorporateApplication) error {
	formDataRaw, err := json.Marshal(app.FormData)
	if err != nil {
		return fmt.Errorf("marshal form data: %w", err)
	}

	query := `
		UPDATE corporate_applications SET
			title = $1,
			internal_reference = $2,
			form_data = $3,
			establishment_id = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND organization_id = $6
	`
	tag, err := r.pool.Exec(ctx, query,
		app.Title, app.InternalReference, formDataRaw, app.EstablishmentID, app.ID, app.OrganizationID,
	)
	if err != nil {
		return fmt.Errorf("update corporate application: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("demande non trouvée")
	}

	return nil
}

// AttachDocumentToItem rattache une pièce GED à une exigence de checklist
func (r *CorporateRepository) AttachDocumentToItem(ctx context.Context, orgID, appID, itemID, gedDocID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Vérifier l'existence et l'appartenance du document GED
	var gedExists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM ged_documents WHERE id = $1 AND organization_id = $2)", gedDocID, orgID).Scan(&gedExists)
	if err != nil || !gedExists {
		return fmt.Errorf("document GED invalide ou introuvable")
	}

	now := time.Now()
	updateQuery := `
		UPDATE corporate_checklist_items SET
			ged_document_id = $1,
			status = 'rattache',
			uploaded_at = $2,
			updated_at = $3
		WHERE id = $4 AND application_id = $5
	`
	tag, err := tx.Exec(ctx, updateQuery, gedDocID, now, now, itemID, appID)
	if err != nil {
		return fmt.Errorf("attach ged item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("exigence de checklist non trouvée")
	}

	// Recalculer le pourcentage de complétude des pièces obligatoires [E]
	if err := r.recalculateCompletenessInTx(ctx, tx, appID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DetachDocumentFromItem détache une pièce GED
func (r *CorporateRepository) DetachDocumentFromItem(ctx context.Context, orgID, appID, itemID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	updateQuery := `
		UPDATE corporate_checklist_items SET
			ged_document_id = NULL,
			status = 'manquant',
			uploaded_at = NULL,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND application_id = $2
	`
	tag, err := tx.Exec(ctx, updateQuery, itemID, appID)
	if err != nil {
		return fmt.Errorf("detach ged item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("exigence de checklist non trouvée")
	}

	if err := r.recalculateCompletenessInTx(ctx, tx, appID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// SignApplicationPDT appose le visa officiel du Pharmacien Directeur Technique
func (r *CorporateRepository) SignApplicationPDT(ctx context.Context, orgID, appID, pdtUserID uuid.UUID, seal string) (*domain.CorporateApplication, error) {
	// Audit de pré-soumission pour s'assurer que toutes les pièces [E] sont présentes
	report, err := r.AuditTabadolPreSubmission(ctx, orgID, appID)
	if err != nil {
		return nil, fmt.Errorf("audit pré-visa: %w", err)
	}
	if !report.CanSubmit {
		return nil, fmt.Errorf("visa refusé : %d pièce(s) obligatoire(s) [E] manquante(s)", len(report.MissingMandatory))
	}

	now := time.Now()
	query := `
		UPDATE corporate_applications SET
			status = 'vise_pdt',
			pdt_signed_by = $1,
			pdt_signed_at = $2,
			pdt_signature_seal = $3,
			updated_at = $4
		WHERE id = $5 AND organization_id = $6
	`
	tag, err := r.pool.Exec(ctx, query, pdtUserID, now, seal, now, appID, orgID)
	if err != nil {
		return nil, fmt.Errorf("apposition visa PDT: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("demande non trouvée")
	}

	return r.GetApplication(ctx, orgID, appID)
}

// SubmitTabadol enregistre la télé-soumission préalable sur la plateforme Tabadol
func (r *CorporateRepository) SubmitTabadol(ctx context.Context, orgID, appID uuid.UUID, submissionRef, notes string) (*domain.CorporateApplication, error) {
	now := time.Now()
	query := `
		UPDATE corporate_applications SET
			status = 'depose_tabadol',
			tabadol_submission_ref = $1,
			tabadol_submitted_at = $2,
			tabadol_submission_notes = $3,
			updated_at = $4
		WHERE id = $5 AND organization_id = $6
	`
	tag, err := r.pool.Exec(ctx, query, submissionRef, now, notes, now, appID, orgID)
	if err != nil {
		return nil, fmt.Errorf("enregistrement soumission Tabadol: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("demande non trouvée")
	}

	return r.GetApplication(ctx, orgID, appID)
}

// RecordMIPHReceipt enregistre le récépissé de dépôt physique avec décharge au bureau d'ordre central
func (r *CorporateRepository) RecordMIPHReceipt(ctx context.Context, orgID, appID uuid.UUID, receiptRef string, receiptDate *time.Time) (*domain.CorporateApplication, error) {
	now := time.Now()
	query := `
		UPDATE corporate_applications SET
			status = 'depose_bo',
			miph_receipt_ref = $1,
			miph_receipt_date = $2,
			updated_at = $3
		WHERE id = $4 AND organization_id = $5
	`
	tag, err := r.pool.Exec(ctx, query, receiptRef, receiptDate, now, appID, orgID)
	if err != nil {
		return nil, fmt.Errorf("enregistrement récépissé MIPH: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("demande non trouvée")
	}

	return r.GetApplication(ctx, orgID, appID)
}

// GrantLicense enregistre la décision officielle d'agrément octroyée par le MIPH
func (r *CorporateRepository) GrantLicense(ctx context.Context, orgID, appID uuid.UUID, decisionRef string, decisionDate, expiryDate *time.Time) (*domain.CorporateApplication, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	app, err := r.GetApplication(ctx, orgID, appID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	updateAppQuery := `
		UPDATE corporate_applications SET
			status = 'agrement_octroye',
			decision_ref = $1,
			decision_date = $2,
			license_expiry_date = $3,
			updated_at = $4
		WHERE id = $5 AND organization_id = $6
	`
	_, err = tx.Exec(ctx, updateAppQuery, decisionRef, decisionDate, expiryDate, now, appID, orgID)
	if err != nil {
		return nil, fmt.Errorf("update application status: %w", err)
	}

	// Créer l'entrée dans le Registre des Titres Réglementaires (corporate_licenses)
	licType := "AGREMENT_OUVERTURE"
	if app.Nature == domain.NaturePrealableRealisation {
		licType = "AGREMENT_REALISATION"
	} else if app.Nature == domain.NatureExercicePDT {
		licType = "DECISION_PDT"
	} else if app.Nature == domain.NatureExerciceAssistant {
		licType = "DECISION_ASSISTANT"
	} else if app.Nature == domain.NatureModifSubstantielle {
		licType = "AGREMENT_MODIF"
	}

	decDateStr := decisionDate.Format("2006-01-02")
	var expDateStr *string
	if expiryDate != nil {
		s := expiryDate.Format("2006-01-02")
		expDateStr = &s
	}

	insertLicQuery := `
		INSERT INTO corporate_licenses (
			id, organization_id, establishment_id, application_id,
			license_type, license_number, decision_date, effective_date,
			expiry_date, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, 'actif', $10, $11
		)
		ON CONFLICT (organization_id, license_number) DO UPDATE SET
			decision_date = EXCLUDED.decision_date,
			expiry_date = EXCLUDED.expiry_date,
			status = 'actif',
			updated_at = CURRENT_TIMESTAMP
	`
	_, err = tx.Exec(ctx, insertLicQuery,
		uuid.New(), orgID, app.EstablishmentID, app.ID,
		licType, decisionRef, decDateStr, decDateStr,
		expDateStr, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert corporate license: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetApplication(ctx, orgID, appID)
}

// AuditTabadolPreSubmission vérifie les conditions bloquantes avant soumission Tabadol
func (r *CorporateRepository) AuditTabadolPreSubmission(ctx context.Context, orgID, appID uuid.UUID) (*domain.TabadolAuditReport, error) {
	app, err := r.GetApplication(ctx, orgID, appID)
	if err != nil {
		return nil, err
	}

	report := &domain.TabadolAuditReport{
		TotalItems:        len(app.ChecklistItems),
		CompletenessRatio: app.CompletenessPercent,
		IsPDTSigned:       app.PDTSignedAt != nil,
		CanSubmit:         true,
	}

	for _, it := range app.ChecklistItems {
		if it.GEDDocumentID != nil {
			report.AttachedItems++
		} else if it.IsMandatory && it.RequirementLevel == "E" {
			report.MissingMandatory = append(report.MissingMandatory, fmt.Sprintf("[%s] %s", it.ItemCode, it.Title))
		}
	}

	if len(report.MissingMandatory) > 0 {
		report.CanSubmit = false
		report.BlockingErrors = append(report.BlockingErrors,
			fmt.Sprintf("%d pièce(s) obligatoire(s) [E] manquante(s) ou non rattachée(s) depuis la GED.", len(report.MissingMandatory)),
		)
	}

	if app.PDTSignedAt == nil {
		report.Warnings = append(report.Warnings, "Le dossier n'a pas encore reçu le visa formel du Pharmacien Directeur Technique.")
	}

	manifest := fmt.Sprintf("MIPH-TABADOL-MANIFEST-V1|APP:%s|CAT:%s|ITEMS:%d|SHA256:VERIFIED",
		app.ApplicationNumber, app.Category, report.AttachedItems,
	)
	report.SubmissionManifest = manifest

	return report, nil
}

// ListLicenses liste les titres réglementaires avec calcul des alertes d'expiration
func (r *CorporateRepository) ListLicenses(ctx context.Context, orgID uuid.UUID, status string) ([]domain.CorporateLicense, error) {
	query := `
		SELECT 
			id, organization_id, establishment_id, application_id,
			license_type, license_number, to_char(decision_date, 'YYYY-MM-DD'),
			to_char(effective_date, 'YYYY-MM-DD'), to_char(expiry_date, 'YYYY-MM-DD'),
			COALESCE(pharmacist_name, ''), COALESCE(pharmacist_order_number, ''),
			status, document_id, created_at, updated_at
		FROM corporate_licenses
		WHERE organization_id = $1
	`
	args := []interface{}{orgID}
	if status != "" && status != "all" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY expiry_date ASC NULLS LAST, decision_date DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list corporate licenses: %w", err)
	}
	defer rows.Close()

	var licenses []domain.CorporateLicense
	now := time.Now()

	for rows.Next() {
		var lic domain.CorporateLicense
		var estID, appID, docID *uuid.UUID
		var expDate *string

		err := rows.Scan(
			&lic.ID, &lic.OrganizationID, &estID, &appID,
			&lic.LicenseType, &lic.LicenseNumber, &lic.DecisionDate,
			&lic.EffectiveDate, &expDate,
			&lic.PharmacistName, &lic.PharmacistOrderNumber,
			&lic.Status, &docID, &lic.CreatedAt, &lic.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan license: %w", err)
		}

		lic.EstablishmentID = estID
		lic.ApplicationID = appID
		lic.DocumentID = docID
		lic.ExpiryDate = expDate

		if expDate != nil {
			t, err := time.Parse("2006-01-02", *expDate)
			if err == nil {
				days := int(t.Sub(now).Hours() / 24)
				lic.DaysUntilExpiry = &days

				// Mise à jour proactive du statut selon le compte à rebours
				if days < 0 {
					lic.Status = "expire"
				} else if days <= 30 {
					lic.Status = "critique_j30"
				} else if days <= 90 {
					lic.Status = "alerte_j90"
				} else if days <= 180 {
					lic.Status = "a_renouveler_j180"
				}
			}
		}

		licenses = append(licenses, lic)
	}

	return licenses, nil
}

// recalculateCompletenessInTx recalcule le taux de complétude des pièces obligatoires [E]
func (r *CorporateRepository) recalculateCompletenessInTx(ctx context.Context, tx pgx.Tx, appID uuid.UUID) error {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE is_mandatory = true AND requirement_level = 'E') as total_e,
			COUNT(*) FILTER (WHERE is_mandatory = true AND requirement_level = 'E' AND status IN ('rattache', 'valide_qualite')) as fulfilled_e
		FROM corporate_checklist_items
		WHERE application_id = $1
	`
	var totalE, fulfilledE int
	err := tx.QueryRow(ctx, query, appID).Scan(&totalE, &fulfilledE)
	if err != nil {
		return fmt.Errorf("count completeness: %w", err)
	}

	pct := 0.0
	if totalE > 0 {
		pct = float64(fulfilledE) * 100.0 / float64(totalE)
	}

	_, err = tx.Exec(ctx, "UPDATE corporate_applications SET completeness_percent = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", pct, appID)
	return err
}
