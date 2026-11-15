package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/domain"
)

var (
	ErrDossierLocked = errors.New("le dossier est verrouillé pour dépôt et ne peut plus être modifié")
)

// DossierRepository gère la persistance des dossiers d'enregistrement ANPP et leurs checklists
type DossierRepository struct {
	pool *pgxpool.Pool
}

// NewDossierRepository instancie le repository des dossiers ANPP
func NewDossierRepository(pool *pgxpool.Pool) *DossierRepository {
	return &DossierRepository{pool: pool}
}

// CreateDossier insère un nouveau dossier et génère tous ses éléments de checklist de manière transactionnelle
func (r *DossierRepository) CreateDossier(ctx context.Context, d *domain.Dossier, items []domain.ChecklistItem) (*domain.Dossier, error) {
	if err := d.ValidateDossier(); err != nil {
		return nil, fmt.Errorf("validation dossier échouée : %w", err)
	}

	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	now := time.Now().UTC()
	d.CreatedAt = now
	d.UpdatedAt = now
	if d.Status == "" {
		d.Status = domain.DossierStatusInPreparation
	}

	// Calcul initial de la complétude
	d.CompletenessPercent = domain.CalculateCompleteness(items)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("échec début transaction : %w", err)
	}
	defer tx.Rollback(ctx)

	// Insertion du dossier maître
	insertDossierQuery := `
		INSERT INTO dossiers (
			id, organization_id, medicine_id, medical_device_id, product_type,
			dossier_code, procedure_code, procedure_type, internal_reference,
			anpp_official_reference, target_deposit_date, tracking_number,
			deposit_receipt_ref, deposit_receipt_date, submission_notes,
			status, responsible_id, completeness_percent, is_locked_for_deposit,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19,
			$20, $21
		)
	`
	procTypeStr := string(d.ProcedureCode)
	_, err = tx.Exec(ctx, insertDossierQuery,
		d.ID, d.OrganizationID, d.MedicineID, d.MedicalDeviceID, string(d.ProductType),
		d.DossierCode, string(d.ProcedureCode), procTypeStr, d.InternalReference,
		d.ANPPOfficialReference, d.TargetDepositDate, d.TrackingNumber,
		d.DepositReceiptRef, d.DepositReceiptDate, d.SubmissionNotes,
		string(d.Status), d.ResponsibleID, d.CompletenessPercent, d.IsLockedForDeposit,
		d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("échec insertion dossier : %w", err)
	}

	// Insertion en lot des éléments de checklist
	if len(items) > 0 {
		insertItemQuery := `
			INSERT INTO dossier_checklist_items (
				id, dossier_id, module_number, sub_module, item_code,
				title_fr, title_ar, description_fr, requirement_level,
				is_mandatory, ged_document_id, status, validation_notes,
				sort_order, security_classification, assigned_department, workflow_status, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9,
				$10, $11, $12, $13,
				$14, $15, $16, $17, $18, $19
			)
		`
		for i := range items {
			item := &items[i]
			if item.ID == uuid.Nil {
				item.ID = uuid.New()
			}
			item.DossierID = d.ID
			if item.Status == "" {
				item.Status = domain.ItemStatusMissing
			}
			if item.WorkflowStatus == "" {
				item.WorkflowStatus = domain.WorkflowStatusDraftPending
			}
			if item.AssignedDepartment == "" {
				item.AssignedDepartment = determineDefaultDepartment(item.ItemCode, item.SubModule, item.ModuleNumber)
			}
			if item.SecurityClassification == "" {
				if item.ItemCode == "1.A.4" || item.ItemCode == "1.H.4" || item.ItemCode == "1.H.5" {
					item.SecurityClassification = domain.ClassificationSecretCommercial
				} else if strings.HasPrefix(item.ItemCode, "3.2.S") || strings.HasPrefix(item.ItemCode, "3.2.P") || item.SubModule == "3.2.S" || item.SubModule == "3.2.P" {
					item.SecurityClassification = domain.ClassificationSecretIndustriel
				} else {
					item.SecurityClassification = domain.ClassificationStandard
				}
			}
			item.CreatedAt = now
			item.UpdatedAt = now

			_, err = tx.Exec(ctx, insertItemQuery,
				item.ID, item.DossierID, item.ModuleNumber, item.SubModule, item.ItemCode,
				item.TitleFR, item.TitleAR, item.DescriptionFR, string(item.RequirementLevel),
				item.IsMandatory, item.GEDDocumentID, string(item.Status), item.ValidationNotes,
				item.SortOrder, string(item.SecurityClassification), item.AssignedDepartment, string(item.WorkflowStatus),
				item.CreatedAt, item.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("échec insertion item de checklist [%s] : %w", item.ItemCode, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("échec commit transaction dossier : %w", err)
	}

	d.Items = items
	d.ChecklistData = items
	return d, nil
}

// GetDossierByID récupère un dossier complet avec ses items de checklist et documents rattachés
func (r *DossierRepository) GetDossierByID(ctx context.Context, orgID, dossierID uuid.UUID) (*domain.Dossier, error) {
	queryDossier := `
		SELECT
			d.id, d.organization_id, d.medicine_id, d.medical_device_id, d.product_type,
			d.dossier_code, d.procedure_code, d.internal_reference, d.anpp_official_reference,
			d.anpp_deposit_date, d.anpp_decision_date, d.anpp_validity_end_date,
			d.target_deposit_date, d.tracking_number, d.deposit_receipt_ref, d.deposit_receipt_date,
			d.submission_notes, d.status, d.responsible_id, d.completeness_percent,
			d.is_locked_for_deposit, d.locked_by_dt_id, d.locked_at,
			d.pdt_signed_at, d.pdt_signed_by, d.pdt_signature_seal, d.pdt_signature_role,
			d.created_at, d.updated_at
		FROM dossiers d
		WHERE d.id = $1 AND d.organization_id = $2
	`

	var d domain.Dossier
	var pTypeStr, pCodeStr, statusStr string

	err := r.pool.QueryRow(ctx, queryDossier, dossierID, orgID).Scan(
		&d.ID, &d.OrganizationID, &d.MedicineID, &d.MedicalDeviceID, &pTypeStr,
		&d.DossierCode, &pCodeStr, &d.InternalReference, &d.ANPPOfficialReference,
		&d.ANPPDepositDate, &d.ANPPDecisionDate, &d.ANPPValidityEndDate,
		&d.TargetDepositDate, &d.TrackingNumber, &d.DepositReceiptRef, &d.DepositReceiptDate,
		&d.SubmissionNotes, &statusStr, &d.ResponsibleID, &d.CompletenessPercent,
		&d.IsLockedForDeposit, &d.LockedByDTID, &d.LockedAt,
		&d.PDTSignedAt, &d.PDTSignedBy, &d.PDTSignatureSeal, &d.PDTSignatureRole,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("échec lecture dossier : %w", err)
	}

	d.ProductType = domain.ProductType(pTypeStr)
	d.ProcedureCode = domain.ProcedureCode(pCodeStr)
	d.Status = domain.DossierStatus(statusStr)

	// Chargement des métadonnées produit associées
	if d.MedicineID != nil {
		medQuery := `
			SELECT id, organization_id, brand_name_fr, COALESCE(brand_name_ar, ''), dci_fr, COALESCE(dci_ar, ''),
			       dosage_fr, pharmaceutical_form_fr, origin_type, economic_file_type, status, created_at, updated_at
			FROM medicines
			WHERE id = $1
		`
		var m domain.Medicine
		var originStr, econStr string
		err := r.pool.QueryRow(ctx, medQuery, *d.MedicineID).Scan(
			&m.ID, &m.OrganizationID, &m.BrandNameFR, &m.BrandNameAR, &m.DCIFR, &m.DCIAR,
			&m.DosageFR, &m.PharmaceuticalFormFR, &originStr, &econStr, &m.Status, &m.CreatedAt, &m.UpdatedAt,
		)
		if err == nil {
			m.OriginType = domain.MedicineOriginType(originStr)
			m.EconomicFileType = domain.EconomicFileType(econStr)
			d.Medicine = &m
		}
	} else if d.MedicalDeviceID != nil {
		dmQuery := `
			SELECT id, organization_id, commercial_name_fr, COALESCE(commercial_name_ar, ''), dm_class,
			       is_sterile, manufacturer_reference, status, created_at, updated_at
			FROM medical_devices
			WHERE id = $1
		`
		var dm domain.MedicalDevice
		err := r.pool.QueryRow(ctx, dmQuery, *d.MedicalDeviceID).Scan(
			&dm.ID, &dm.OrganizationID, &dm.CommercialNameFR, &dm.CommercialNameAR, &dm.DMClass,
			&dm.IsSterile, &dm.ManufacturerReference, &dm.Status, &dm.CreatedAt, &dm.UpdatedAt,
		)
		if err == nil {
			d.MedicalDevice = &dm
		}
	}

	// Chargement des items de checklist avec jointure GED si document rattaché et informations du Workflow
	queryItems := `
		SELECT
			i.id, i.dossier_id, i.module_number, i.sub_module, i.item_code,
			i.title_fr, COALESCE(i.title_ar, ''), COALESCE(i.description_fr, ''),
			i.requirement_level, i.is_mandatory, i.ged_document_id, i.status,
			i.validation_notes, i.sort_order, COALESCE(i.security_classification, 'STANDARD'),
			COALESCE(i.workflow_status, 'draft_pending'), COALESCE(i.assigned_department, 'RA'),
			i.assigned_user_id, i.due_date, i.is_locked, i.locked_by, i.locked_at, i.lock_reason,
			COALESCE(i.current_revision, 'v0.1'), i.maker_user_id, i.maker_submitted_at,
			i.checker_user_id, i.checker_reviewed_at, i.checker_decision, i.checker_notes,
			i.approver_user_id, i.approver_signed_at, i.approver_signature_seal,
			i.authored_content, i.template_code, i.created_at, i.updated_at,
			g.id, g.title_fr, g.document_code, g.current_version, g.status, g.taxonomy_path,
			COALESCE(u_asg.first_name || ' ' || u_asg.last_name, ''),
			COALESCE(u_lck.first_name || ' ' || u_lck.last_name, ''),
			COALESCE(u_mkr.first_name || ' ' || u_mkr.last_name, ''),
			COALESCE(u_chk.first_name || ' ' || u_chk.last_name, ''),
			COALESCE(u_app.first_name || ' ' || u_app.last_name, '')
		FROM dossier_checklist_items i
		LEFT JOIN ged_documents g ON i.ged_document_id = g.id
		LEFT JOIN users u_asg ON i.assigned_user_id = u_asg.id
		LEFT JOIN users u_lck ON i.locked_by = u_lck.id
		LEFT JOIN users u_mkr ON i.maker_user_id = u_mkr.id
		LEFT JOIN users u_chk ON i.checker_user_id = u_chk.id
		LEFT JOIN users u_app ON i.approver_user_id = u_app.id
		WHERE i.dossier_id = $1
		ORDER BY i.module_number ASC, i.sort_order ASC, i.item_code ASC
	`
	rows, err := r.pool.Query(ctx, queryItems, dossierID)
	if err != nil {
		return nil, fmt.Errorf("échec lecture items checklist : %w", err)
	}
	defer rows.Close()

	var items []domain.ChecklistItem
	for rows.Next() {
		var it domain.ChecklistItem
		var reqLvlStr, itStatusStr, secClassStr, wfStatusStr string
		var titleAR, descFR string
		var gedID *uuid.UUID
		var gTitle, gCode, gVer, gStatus, gTaxo *string
		var asgName, lckName, mkrName, chkName, appName string

		err := rows.Scan(
			&it.ID, &it.DossierID, &it.ModuleNumber, &it.SubModule, &it.ItemCode,
			&it.TitleFR, &titleAR, &descFR,
			&reqLvlStr, &it.IsMandatory, &it.GEDDocumentID, &itStatusStr,
			&it.ValidationNotes, &it.SortOrder, &secClassStr,
			&wfStatusStr, &it.AssignedDepartment,
			&it.AssignedUserID, &it.DueDate, &it.IsLocked, &it.LockedBy, &it.LockedAt, &it.LockReason,
			&it.CurrentRevision, &it.MakerUserID, &it.MakerSubmittedAt,
			&it.CheckerUserID, &it.CheckerReviewedAt, &it.CheckerDecision, &it.CheckerNotes,
			&it.ApproverUserID, &it.ApproverSignedAt, &it.ApproverSignatureSeal,
			&it.AuthoredContent, &it.TemplateCode, &it.CreatedAt, &it.UpdatedAt,
			&gedID, &gTitle, &gCode, &gVer, &gStatus, &gTaxo,
			&asgName, &lckName, &mkrName, &chkName, &appName,
		)
		if err != nil {
			return nil, fmt.Errorf("échec scan item checklist : %w", err)
		}

		it.TitleAR = titleAR
		it.DescriptionFR = descFR
		it.RequirementLevel = domain.RequirementLevel(reqLvlStr)
		it.Status = domain.ChecklistItemStatus(itStatusStr)
		it.SecurityClassification = domain.SecurityClassification(secClassStr)
		it.WorkflowStatus = domain.WorkflowStatus(wfStatusStr)
		it.AssignedUserName = asgName
		it.LockedByName = lckName
		it.MakerUserName = mkrName
		it.CheckerUserName = chkName
		it.ApproverUserName = appName

		if gedID != nil && gTitle != nil {
			docCode := ""
			if gCode != nil {
				docCode = *gCode
			}
			currVer := "v1.0"
			if gVer != nil {
				currVer = *gVer
			}
			docStatus := domain.GEDStatusEnVigueur
			if gStatus != nil {
				docStatus = domain.GEDDocumentStatus(*gStatus)
			}
			taxo := ""
			if gTaxo != nil {
				taxo = *gTaxo
			}

			it.GEDDocument = &domain.GEDDocument{
				ID:             *gedID,
				TitleFR:        *gTitle,
				DocumentCode:   docCode,
				CurrentVersion: currVer,
				Status:         docStatus,
				TaxonomyPath:   taxo,
			}
		}

		items = append(items, it)
	}

	d.Items = items
	d.ChecklistData = items
	d.CompletenessPercent = domain.CalculateCompleteness(items)

	return &d, nil
}

// ListDossiers liste tous les dossiers d'une organisation avec filtres optionnels et cloisonnement strict (Need-to-Know)
func (r *DossierRepository) ListDossiers(ctx context.Context, orgID uuid.UUID, restrictedUserID *uuid.UUID, procFilter string, statusFilter string) ([]domain.Dossier, error) {
	baseQuery := `
		SELECT
			d.id, d.organization_id, d.medicine_id, d.medical_device_id, d.product_type,
			d.dossier_code, d.procedure_code, d.internal_reference, d.anpp_official_reference,
			d.anpp_deposit_date, d.anpp_decision_date, d.anpp_validity_end_date,
			d.target_deposit_date, d.tracking_number, d.deposit_receipt_ref, d.deposit_receipt_date,
			d.submission_notes, d.status, d.responsible_id, d.completeness_percent,
			d.is_locked_for_deposit, d.locked_by_dt_id, d.locked_at,
			d.pdt_signed_at, d.pdt_signed_by, d.pdt_signature_seal, d.pdt_signature_role,
			d.created_at, d.updated_at,
			m.brand_name_fr, m.dci_fr, m.dosage_fr,
			dm.commercial_name_fr, dm.dm_class
		FROM dossiers d
		LEFT JOIN medicines m ON d.medicine_id = m.id
		LEFT JOIN medical_devices dm ON d.medical_device_id = dm.id
		WHERE d.organization_id = $1
	`

	args := []interface{}{orgID}
	argIdx := 2

	if restrictedUserID != nil {
		baseQuery += fmt.Sprintf(` AND d.id IN (
			SELECT dossier_id FROM dossier_assignments WHERE user_id = $%d AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
			UNION
			SELECT dossier_id FROM dossier_consultants WHERE user_id = $%d AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		)`, argIdx, argIdx)
		args = append(args, *restrictedUserID)
		argIdx++
	}

	if strings.TrimSpace(procFilter) != "" {
		baseQuery += fmt.Sprintf(" AND d.procedure_code = $%d", argIdx)
		args = append(args, strings.TrimSpace(procFilter))
		argIdx++
	}
	if strings.TrimSpace(statusFilter) != "" {
		baseQuery += fmt.Sprintf(" AND d.status = $%d", argIdx)
		args = append(args, strings.TrimSpace(statusFilter))
		argIdx++
	}

	baseQuery += " ORDER BY d.created_at DESC"

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("échec lecture liste dossiers : %w", err)
	}
	defer rows.Close()

	var dossiers []domain.Dossier
	for rows.Next() {
		var d domain.Dossier
		var pTypeStr, pCodeStr, statusStr string
		var medBrand, medDCI, medDosage *string
		var dmComm, dmClass *string

		err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.MedicineID, &d.MedicalDeviceID, &pTypeStr,
			&d.DossierCode, &pCodeStr, &d.InternalReference, &d.ANPPOfficialReference,
			&d.ANPPDepositDate, &d.ANPPDecisionDate, &d.ANPPValidityEndDate,
			&d.TargetDepositDate, &d.TrackingNumber, &d.DepositReceiptRef, &d.DepositReceiptDate,
			&d.SubmissionNotes, &statusStr, &d.ResponsibleID, &d.CompletenessPercent,
			&d.IsLockedForDeposit, &d.LockedByDTID, &d.LockedAt,
			&d.PDTSignedAt, &d.PDTSignedBy, &d.PDTSignatureSeal, &d.PDTSignatureRole,
			&d.CreatedAt, &d.UpdatedAt,
			&medBrand, &medDCI, &medDosage,
			&dmComm, &dmClass,
		)
		if err != nil {
			return nil, fmt.Errorf("échec scan ligne dossier : %w", err)
		}

		d.ProductType = domain.ProductType(pTypeStr)
		d.ProcedureCode = domain.ProcedureCode(pCodeStr)
		d.Status = domain.DossierStatus(statusStr)

		if medBrand != nil {
			d.Medicine = &domain.Medicine{
				BrandNameFR: *medBrand,
			}
			if medDCI != nil {
				d.Medicine.DCIFR = *medDCI
			}
			if medDosage != nil {
				d.Medicine.DosageFR = *medDosage
			}
		}
		if dmComm != nil {
			d.MedicalDevice = &domain.MedicalDevice{
				CommercialNameFR: *dmComm,
			}
			if dmClass != nil {
				d.MedicalDevice.DMClass = *dmClass
			}
		}

		dossiers = append(dossiers, d)
	}

	return dossiers, nil
}

// AttachGEDDocument associe un document approuvé du coffre-fort GED à une exigence de checklist
func (r *DossierRepository) AttachGEDDocument(ctx context.Context, orgID, dossierID, itemID, gedDocID uuid.UUID, notes *string) (*domain.ChecklistItem, error) {
	// Vérifier que le dossier existe, appartient à l'org et n'est pas verrouillé
	var isLocked bool
	checkQuery := `SELECT is_locked_for_deposit FROM dossiers WHERE id = $1 AND organization_id = $2`
	err := r.pool.QueryRow(ctx, checkQuery, dossierID, orgID).Scan(&isLocked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("échec vérification dossier : %w", err)
	}
	if isLocked {
		return nil, ErrDossierLocked
	}

	// Vérifier que le document GED existe dans l'organisation
	var gedExists bool
	gedCheck := `SELECT EXISTS(SELECT 1 FROM ged_documents WHERE id = $1 AND organization_id = $2)`
	if err := r.pool.QueryRow(ctx, gedCheck, gedDocID, orgID).Scan(&gedExists); err != nil || !gedExists {
		return nil, fmt.Errorf("document GED introuvable ou non autorisé")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("échec début transaction : %w", err)
	}
	defer tx.Rollback(ctx)

	// Mise à jour de l'item de checklist
	now := time.Now().UTC()
	updateItemQuery := `
		UPDATE dossier_checklist_items
		SET ged_document_id = $1, status = $2, validation_notes = $3, updated_at = $4
		WHERE id = $5 AND dossier_id = $6
		RETURNING id, dossier_id, module_number, sub_module, item_code, title_fr, requirement_level, is_mandatory, ged_document_id, status, validation_notes, updated_at
	`
	var it domain.ChecklistItem
	var reqStr, stStr string
	err = tx.QueryRow(ctx, updateItemQuery, gedDocID, string(domain.ItemStatusAttached), notes, now, itemID, dossierID).Scan(
		&it.ID, &it.DossierID, &it.ModuleNumber, &it.SubModule, &it.ItemCode, &it.TitleFR,
		&reqStr, &it.IsMandatory, &it.GEDDocumentID, &stStr, &it.ValidationNotes, &it.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("échec mise à jour item checklist : %w", err)
	}
	it.RequirementLevel = domain.RequirementLevel(reqStr)
	it.Status = domain.ChecklistItemStatus(stStr)

	// Recalcul de la complétude
	newPct, err := r.recalculateCompletenessTx(ctx, tx, dossierID)
	if err != nil {
		return nil, fmt.Errorf("échec recalcul complétude : %w", err)
	}

	// Si complétude > 0 et dossier était brouillon, passer en préparation
	_, err = tx.Exec(ctx, `
		UPDATE dossiers
		SET completeness_percent = $1,
		    status = CASE WHEN status = 'brouillon' THEN 'en_preparation' ELSE status END,
		    updated_at = $2
		WHERE id = $3
	`, newPct, now, dossierID)
	if err != nil {
		return nil, fmt.Errorf("échec mise à jour complétude dossier : %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("échec commit rattachement : %w", err)
	}

	return &it, nil
}

// DetachGEDDocument détache le document GED d'un item de checklist
func (r *DossierRepository) DetachGEDDocument(ctx context.Context, orgID, dossierID, itemID uuid.UUID) (*domain.ChecklistItem, error) {
	var isLocked bool
	checkQuery := `SELECT is_locked_for_deposit FROM dossiers WHERE id = $1 AND organization_id = $2`
	err := r.pool.QueryRow(ctx, checkQuery, dossierID, orgID).Scan(&isLocked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("échec vérification dossier : %w", err)
	}
	if isLocked {
		return nil, ErrDossierLocked
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("échec début transaction : %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	updateItemQuery := `
		UPDATE dossier_checklist_items
		SET ged_document_id = NULL, status = $1, validation_notes = NULL, updated_at = $2
		WHERE id = $3 AND dossier_id = $4
		RETURNING id, dossier_id, module_number, sub_module, item_code, title_fr, requirement_level, is_mandatory, ged_document_id, status, updated_at
	`
	var it domain.ChecklistItem
	var reqStr, stStr string
	err = tx.QueryRow(ctx, updateItemQuery, string(domain.ItemStatusMissing), now, itemID, dossierID).Scan(
		&it.ID, &it.DossierID, &it.ModuleNumber, &it.SubModule, &it.ItemCode, &it.TitleFR,
		&reqStr, &it.IsMandatory, &it.GEDDocumentID, &stStr, &it.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("échec détachement document : %w", err)
	}
	it.RequirementLevel = domain.RequirementLevel(reqStr)
	it.Status = domain.ChecklistItemStatus(stStr)

	// Recalcul complétude
	newPct, err := r.recalculateCompletenessTx(ctx, tx, dossierID)
	if err != nil {
		return nil, fmt.Errorf("échec recalcul complétude : %w", err)
	}

	_, err = tx.Exec(ctx, `UPDATE dossiers SET completeness_percent = $1, updated_at = $2 WHERE id = $3`, newPct, now, dossierID)
	if err != nil {
		return nil, fmt.Errorf("échec mise à jour complétude dossier : %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("échec commit détachement : %w", err)
	}

	return &it, nil
}

// UpdateChecklistItemStatus met à jour le statut d'un élément de checklist (ex: approuvé DT ou validé qualité)
func (r *DossierRepository) UpdateChecklistItemStatus(ctx context.Context, orgID, dossierID, itemID uuid.UUID, newStatus domain.ChecklistItemStatus, notes *string) (*domain.ChecklistItem, error) {
	var isLocked bool
	checkQuery := `SELECT is_locked_for_deposit FROM dossiers WHERE id = $1 AND organization_id = $2`
	err := r.pool.QueryRow(ctx, checkQuery, dossierID, orgID).Scan(&isLocked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("échec vérification dossier : %w", err)
	}
	if isLocked {
		return nil, ErrDossierLocked
	}

	now := time.Now().UTC()
	updateQuery := `
		UPDATE dossier_checklist_items
		SET status = $1, validation_notes = COALESCE($2, validation_notes), updated_at = $3
		WHERE id = $4 AND dossier_id = $5
		RETURNING id, dossier_id, module_number, sub_module, item_code, title_fr, requirement_level, is_mandatory, ged_document_id, status, validation_notes, updated_at
	`
	var it domain.ChecklistItem
	var reqStr, stStr string
	err = r.pool.QueryRow(ctx, updateQuery, string(newStatus), notes, now, itemID, dossierID).Scan(
		&it.ID, &it.DossierID, &it.ModuleNumber, &it.SubModule, &it.ItemCode, &it.TitleFR,
		&reqStr, &it.IsMandatory, &it.GEDDocumentID, &stStr, &it.ValidationNotes, &it.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("échec mise à jour statut item : %w", err)
	}
	it.RequirementLevel = domain.RequirementLevel(reqStr)
	it.Status = domain.ChecklistItemStatus(stStr)

	return &it, nil
}

// SignDossierPDT appose le visa électronique 21 CFR Part 11 du Pharmacien Directeur Technique et verrouille le dossier
func (r *DossierRepository) SignDossierPDT(ctx context.Context, orgID, dossierID, userID uuid.UUID, userRole string, signatureSeal string) (*domain.Dossier, error) {
	// Vérifier l'éligibilité du rôle (doit être directeur technique / pharmacien responsable / admin)
	normRole := strings.ToLower(userRole)
	if normRole != "pharmacien_responsable" && normRole != "admin" && normRole != "directeur_technique" {
		return nil, errors.New("seul le Pharmacien Directeur Technique (ou Pharmacien Responsable) est habilité à apposer le visa de conformité préalable au dépôt ANPP")
	}

	d, err := r.GetDossierByID(ctx, orgID, dossierID)
	if err != nil {
		return nil, err
	}

	if d.IsLockedForDeposit && d.PDTSignedAt != nil {
		return nil, errors.New("le dossier porte déjà le visa officiel du Pharmacien Directeur Technique")
	}

	// Contrôle de complétude stricte à 100%
	if d.CompletenessPercent < 100 {
		return nil, fmt.Errorf("impossibilité d'apposer le visa : la complétude documentaire est de %d%% (100%% exigés). Toutes les pièces [E] doivent être fournies", d.CompletenessPercent)
	}

	// Contrôle d'absence de pièces obligatoires manquantes
	for _, it := range d.Items {
		if it.IsMandatory && it.RequirementLevel == domain.RequirementExiged {
			if it.GEDDocumentID == nil || it.Status == domain.ItemStatusMissing || it.Status == domain.ItemStatusRejected {
				return nil, fmt.Errorf("pièce obligatoire manquante ou rejetée : [%s] %s", it.ItemCode, it.TitleFR)
			}
		}
	}

	now := time.Now().UTC()
	signQuery := `
		UPDATE dossiers
		SET pdt_signed_at = $1,
		    pdt_signed_by = $2,
		    pdt_signature_seal = $3,
		    pdt_signature_role = $4,
		    status = $5,
		    is_locked_for_deposit = TRUE,
		    locked_by_dt_id = $6,
		    locked_at = $7,
		    updated_at = $8
		WHERE id = $9 AND organization_id = $10
	`
	roleTitle := "Pharmacien Directeur Technique (PDT)"
	newStatus := string(domain.DossierStatusValidatedInternal)

	_, err = r.pool.Exec(ctx, signQuery,
		now, userID, signatureSeal, roleTitle,
		newStatus, userID, now, now,
		dossierID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("échec enregistrement visa PDT : %w", err)
	}

	return r.GetDossierByID(ctx, orgID, dossierID)
}

// UpdateDossierStatus met à jour l'état réglementaire du dossier dans le cycle de vie ANPP
func (r *DossierRepository) UpdateDossierStatus(
	ctx context.Context, orgID, dossierID uuid.UUID,
	newStatus domain.DossierStatus, notes *string,
	receiptRef *string, receiptDate *time.Time,
	anppOfficialRef *string, decisionDate *time.Time, validityEndDate *time.Time,
) (*domain.Dossier, error) {
	d, err := r.GetDossierByID(ctx, orgID, dossierID)
	if err != nil {
		return nil, err
	}

	// Règles de transition
	if newStatus == domain.DossierStatusSubmittedANPP {
		// Le dossier doit être signé PDT
		if d.PDTSignedAt == nil || d.CompletenessPercent < 100 {
			return nil, errors.New("le dossier doit avoir reçu le visa de conformité du Pharmacien Directeur Technique et atteindre 100% de complétude avant enregistrement du dépôt ANPP")
		}
	}

	now := time.Now().UTC()
	updateQuery := `
		UPDATE dossiers
		SET status = $1,
		    submission_notes = COALESCE($2, submission_notes),
		    deposit_receipt_ref = COALESCE($3, deposit_receipt_ref),
		    deposit_receipt_date = COALESCE($4, deposit_receipt_date),
		    anpp_official_reference = COALESCE($5, anpp_official_reference),
		    anpp_decision_date = COALESCE($6, anpp_decision_date),
		    anpp_validity_end_date = COALESCE($7, anpp_validity_end_date),
		    updated_at = $8
		WHERE id = $9 AND organization_id = $10
	`

	_, err = r.pool.Exec(ctx, updateQuery,
		string(newStatus), notes, receiptRef, receiptDate,
		anppOfficialRef, decisionDate, validityEndDate,
		now, dossierID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("échec mise à jour statut dossier : %w", err)
	}

	return r.GetDossierByID(ctx, orgID, dossierID)
}

// recalculateCompletenessTx recalcule le pourcentage de complétude pour un dossier donné dans une transaction
func (r *DossierRepository) recalculateCompletenessTx(ctx context.Context, tx pgx.Tx, dossierID uuid.UUID) (int, error) {
	query := `
		SELECT requirement_level, is_mandatory, status
		FROM dossier_checklist_items
		WHERE dossier_id = $1
	`
	rows, err := tx.Query(ctx, query, dossierID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	totalMandatory := 0
	fulfilledMandatory := 0

	for rows.Next() {
		var reqLvlStr, stStr string
		var isMandatory bool
		if err := rows.Scan(&reqLvlStr, &isMandatory, &stStr); err != nil {
			return 0, err
		}

		if isMandatory && domain.RequirementLevel(reqLvlStr) == domain.RequirementExiged {
			totalMandatory++
			st := domain.ChecklistItemStatus(stStr)
			if st == domain.ItemStatusAttached || st == domain.ItemStatusValidated || st == domain.ItemStatusApprovedDT {
				fulfilledMandatory++
			}
		}
	}

	if totalMandatory == 0 {
		return 100, nil
	}
	return (fulfilledMandatory * 100) / totalMandatory, nil
}

// GetDossierAssignment retourne l'affectation nominative et les prérogatives de secret d'un utilisateur sur un dossier
func (r *DossierRepository) GetDossierAssignment(ctx context.Context, dossierID, userID uuid.UUID) (*domain.DossierAssignment, error) {
	query := `
		SELECT 
			a.id, a.dossier_id, a.user_id, a.role_in_dossier, 
			a.can_view_financials, a.can_view_industrial_secret, 
			a.assigned_at, a.expires_at, a.notes, a.created_at,
			u.email, u.first_name, u.last_name, u.role
		FROM dossier_assignments a
		JOIN users u ON a.user_id = u.id
		WHERE a.dossier_id = $1 AND a.user_id = $2 AND (a.expires_at IS NULL OR a.expires_at > CURRENT_TIMESTAMP)
	`
	var a domain.DossierAssignment
	var email, fn, ln, role string
	err := r.pool.QueryRow(ctx, query, dossierID, userID).Scan(
		&a.ID, &a.DossierID, &a.UserID, &a.RoleInDossier,
		&a.CanViewFinancials, &a.CanViewIndustrialSecret,
		&a.AssignedAt, &a.ExpiresAt, &a.Notes, &a.CreatedAt,
		&email, &fn, &ln, &role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("erreur lecture affectation dossier: %w", err)
	}
	a.UserEmail = email
	a.UserFirstName = fn
	a.UserLastName = ln
	a.UserRole = role
	return &a, nil
}

// ListDossierMembers liste tous les membres affectés à un dossier avec leurs détails d'identité
func (r *DossierRepository) ListDossierMembers(ctx context.Context, dossierID uuid.UUID) ([]domain.DossierAssignment, error) {
	query := `
		SELECT 
			a.id, a.dossier_id, a.user_id, a.role_in_dossier, 
			a.can_view_financials, a.can_view_industrial_secret, 
			a.assigned_at, a.expires_at, a.notes, a.created_at,
			u.email, u.first_name, u.last_name, u.role
		FROM dossier_assignments a
		JOIN users u ON a.user_id = u.id
		WHERE a.dossier_id = $1
		ORDER BY a.created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, dossierID)
	if err != nil {
		return nil, fmt.Errorf("erreur listage membres dossier: %w", err)
	}
	defer rows.Close()

	var members []domain.DossierAssignment
	for rows.Next() {
		var m domain.DossierAssignment
		var email, fn, ln, role string
		err := rows.Scan(
			&m.ID, &m.DossierID, &m.UserID, &m.RoleInDossier,
			&m.CanViewFinancials, &m.CanViewIndustrialSecret,
			&m.AssignedAt, &m.ExpiresAt, &m.Notes, &m.CreatedAt,
			&email, &fn, &ln, &role,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan membre dossier: %w", err)
		}
		m.UserEmail = email
		m.UserFirstName = fn
		m.UserLastName = ln
		m.UserRole = role
		members = append(members, m)
	}
	return members, nil
}

// AssignDossierMember ajoute ou met à jour une affectation nominative à un dossier
func (r *DossierRepository) AssignDossierMember(ctx context.Context, a *domain.DossierAssignment) error {
	query := `
		INSERT INTO dossier_assignments (
			dossier_id, user_id, role_in_dossier, can_view_financials, can_view_industrial_secret, expires_at, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (dossier_id, user_id) DO UPDATE SET
			role_in_dossier = EXCLUDED.role_in_dossier,
			can_view_financials = EXCLUDED.can_view_financials,
			can_view_industrial_secret = EXCLUDED.can_view_industrial_secret,
			expires_at = EXCLUDED.expires_at,
			notes = EXCLUDED.notes;
	`
	_, err := r.pool.Exec(ctx, query,
		a.DossierID, a.UserID, a.RoleInDossier,
		a.CanViewFinancials, a.CanViewIndustrialSecret,
		a.ExpiresAt, a.Notes,
	)
	if err != nil {
		return fmt.Errorf("erreur assignation membre dossier: %w", err)
	}
	return nil
}

// RemoveDossierMember révoque l'affectation d'un membre à un dossier
func (r *DossierRepository) RemoveDossierMember(ctx context.Context, dossierID, userID uuid.UUID) error {
	query := `DELETE FROM dossier_assignments WHERE dossier_id = $1 AND user_id = $2`
	_, err := r.pool.Exec(ctx, query, dossierID, userID)
	if err != nil {
		return fmt.Errorf("erreur révocation membre dossier: %w", err)
	}
	return nil
}

// determineDefaultDepartment affecte le pôle métier selon la matrice de responsabilités BPF (Section 7.4 CDC)
func determineDefaultDepartment(itemCode, subModule string, moduleNum int) string {
	if moduleNum == 1 {
		return "RA"
	}
	if strings.HasPrefix(itemCode, "3.2.P.3") || strings.HasPrefix(itemCode, "1.P.8") {
		return "PROD"
	}
	if strings.HasPrefix(itemCode, "3.2.P.2") || strings.HasPrefix(itemCode, "1.M.P.3") || moduleNum == 2 {
		return "RD"
	}
	if strings.HasPrefix(itemCode, "3.2.S.4") || strings.HasPrefix(itemCode, "3.2.P.5") || strings.HasPrefix(itemCode, "3.2.P.8") || strings.HasPrefix(itemCode, "1.P.4") || strings.HasPrefix(itemCode, "1.P.5") {
		return "QC"
	}
	if strings.HasPrefix(itemCode, "5.3.1") || strings.HasPrefix(itemCode, "3.2.S") || strings.HasPrefix(itemCode, "1.M.S") {
		return "EXTERNAL"
	}
	if moduleNum == 3 {
		return "QC"
	}
	return "RA"
}
