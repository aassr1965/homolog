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

var (
	ErrWorkflowItemNotFound = errors.New("section ou élément de checklist introuvable")
	ErrSectionLocked        = errors.New("la section est verrouillée par un autre collaborateur")
	ErrSectionNotLocked     = errors.New("la section n'est pas verrouillée")
	ErrUnauthorizedCheckin  = errors.New("seul le détenteur du verrou peut effectuer le Check-in")
	ErrSelfReviewNotAllowed = errors.New("ségrégation des fonctions BPF : le rédacteur ne peut pas relire ou approuver son propre document")
)

// GatingBlockedError indique qu'une transition est bloquée par une dépendance réglementaire CTD/DM
type GatingBlockedError struct {
	Result domain.GatingEvaluationResult
}

func (e *GatingBlockedError) Error() string {
	return e.Result.SummaryMessage
}

// WorkflowRepository gère la persistance de l'atelier documentaire et des revues BPF
type WorkflowRepository struct {
	pool *pgxpool.Pool
}

// NewWorkflowRepository instancie le repository du workflow Maker-Checker
func NewWorkflowRepository(pool *pgxpool.Pool) *WorkflowRepository {
	return &WorkflowRepository{pool: pool}
}

// LockExpiryResult représente le résultat de l'auto-libération d'un verrou expiré
type LockExpiryResult struct {
	ItemID   uuid.UUID
	ItemCode string
	LockedBy uuid.UUID
	LockedAt time.Time
}

// ReleaseExpiredLocks libère automatiquement les verrous de Check-out dont le bail a expiré.
// Conforme ALCOA+ "Timely" : un verrou non restitué dans le délai BPF (bailMinutes) est
// libéré automatiquement avec traçabilité complète dans l'audit trail de revue.
// Retourne la liste des items libérés et le nombre total traité.
func (r *WorkflowRepository) ReleaseExpiredLocks(ctx context.Context, bailMinutes int) ([]LockExpiryResult, error) {
	if bailMinutes <= 0 {
		bailMinutes = 30
	}
	cutoff := time.Now().UTC().Add(-time.Duration(bailMinutes) * time.Minute)

	// Identifier les items dont le bail est expiré
	selectQuery := `
		SELECT i.id, i.item_code, i.locked_by, i.locked_at
		FROM dossier_checklist_items i
		WHERE i.is_locked = TRUE
		  AND i.locked_at IS NOT NULL
		  AND i.locked_at < $1
	`
	rows, err := r.pool.Query(ctx, selectQuery, cutoff)
	if err != nil {
		return nil, fmt.Errorf("échec lecture des verrous expirés : %w", err)
	}
	defer rows.Close()

	var expired []LockExpiryResult
	for rows.Next() {
		var res LockExpiryResult
		if err := rows.Scan(&res.ItemID, &res.ItemCode, &res.LockedBy, &res.LockedAt); err != nil {
			continue
		}
		expired = append(expired, res)
	}
	rows.Close()

	if len(expired) == 0 {
		return nil, nil
	}

	now := time.Now().UTC()

	for _, item := range expired {
		// Libération du verrou avec retour à l'état précédent
		releaseQuery := `
			UPDATE dossier_checklist_items
			SET is_locked      = FALSE,
			    locked_by      = NULL,
			    locked_at      = NULL,
			    lock_reason    = NULL,
			    workflow_status = CASE
			        WHEN workflow_status = 'drafting' THEN 'draft_pending'
			        ELSE workflow_status
			    END,
			    updated_at     = $1
			WHERE id = $2
			  AND is_locked = TRUE
			  AND locked_at < $3
		`
		_, err := r.pool.Exec(ctx, releaseQuery, now, item.ItemID, cutoff)
		if err != nil {
			continue // Meilleur effort : on continue sur les autres items
		}

		// Entrée d'audit trail ALCOA+ — traçabilité immuable de l'expiration
		auditMsg := fmt.Sprintf(
			"Verrou de Check-out expiré automatiquement après %d minutes d'inactivité (ALCOA+ Timely). "+
				"Section restituée à l'état draft_pending. Rédacteur précédent : %s.",
			bailMinutes, item.LockedBy.String(),
		)
		_ = r.AddReviewEntry(ctx, item.ItemID, item.LockedBy, "SYSTEM", domain.ReviewTypeAnnotation, auditMsg, "")
	}

	return expired, nil
}

// GetChecklistItemByID charge un item de checklist avec son contexte de workflow
func (r *WorkflowRepository) GetChecklistItemByID(ctx context.Context, itemID uuid.UUID) (*domain.ChecklistItem, error) {
	query := `
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
			COALESCE(u_asg.first_name || ' ' || u_asg.last_name, ''),
			COALESCE(u_lck.first_name || ' ' || u_lck.last_name, ''),
			COALESCE(u_mkr.first_name || ' ' || u_mkr.last_name, ''),
			COALESCE(u_chk.first_name || ' ' || u_chk.last_name, ''),
			COALESCE(u_app.first_name || ' ' || u_app.last_name, '')
		FROM dossier_checklist_items i
		LEFT JOIN users u_asg ON i.assigned_user_id = u_asg.id
		LEFT JOIN users u_lck ON i.locked_by = u_lck.id
		LEFT JOIN users u_mkr ON i.maker_user_id = u_mkr.id
		LEFT JOIN users u_chk ON i.checker_user_id = u_chk.id
		LEFT JOIN users u_app ON i.approver_user_id = u_app.id
		WHERE i.id = $1
	`
	var it domain.ChecklistItem
	var reqStr, stStr, secClassStr, wfStr string
	var titleAR, descFR string
	var asgName, lckName, mkrName, chkName, appName string

	err := r.pool.QueryRow(ctx, query, itemID).Scan(
		&it.ID, &it.DossierID, &it.ModuleNumber, &it.SubModule, &it.ItemCode,
		&it.TitleFR, &titleAR, &descFR,
		&reqStr, &it.IsMandatory, &it.GEDDocumentID, &stStr,
		&it.ValidationNotes, &it.SortOrder, &secClassStr,
		&wfStr, &it.AssignedDepartment,
		&it.AssignedUserID, &it.DueDate, &it.IsLocked, &it.LockedBy, &it.LockedAt, &it.LockReason,
		&it.CurrentRevision, &it.MakerUserID, &it.MakerSubmittedAt,
		&it.CheckerUserID, &it.CheckerReviewedAt, &it.CheckerDecision, &it.CheckerNotes,
		&it.ApproverUserID, &it.ApproverSignedAt, &it.ApproverSignatureSeal,
		&it.AuthoredContent, &it.TemplateCode, &it.CreatedAt, &it.UpdatedAt,
		&asgName, &lckName, &mkrName, &chkName, &appName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkflowItemNotFound
		}
		return nil, fmt.Errorf("lecture item checklist échouée : %w", err)
	}

	it.TitleAR = titleAR
	it.DescriptionFR = descFR
	it.RequirementLevel = domain.RequirementLevel(reqStr)
	it.Status = domain.ChecklistItemStatus(stStr)
	it.SecurityClassification = domain.SecurityClassification(secClassStr)
	it.WorkflowStatus = domain.WorkflowStatus(wfStr)
	it.AssignedUserName = asgName
	it.LockedByName = lckName
	it.MakerUserName = mkrName
	it.CheckerUserName = chkName
	it.ApproverUserName = appName

	return &it, nil
}

// GetDossierItems charge tous les items d'un dossier pour évaluer le DAG de dépendances et le gating
func (r *WorkflowRepository) GetDossierItems(ctx context.Context, dossierID uuid.UUID) ([]domain.ChecklistItem, error) {
	query := `
		SELECT 
			id, dossier_id, module_number, COALESCE(sub_module, ''), item_code,
			title_fr, COALESCE(title_ar, ''), COALESCE(description_fr, ''),
			requirement_level, is_mandatory, ged_document_id, status,
			COALESCE(workflow_status, 'draft_pending'), COALESCE(current_revision, 'v0.1')
		FROM dossier_checklist_items
		WHERE dossier_id = $1
		ORDER BY module_number ASC, sort_order ASC, item_code ASC
	`
	rows, err := r.pool.Query(ctx, query, dossierID)
	if err != nil {
		return nil, fmt.Errorf("lecture items du dossier échouée : %w", err)
	}
	defer rows.Close()

	var items []domain.ChecklistItem
	for rows.Next() {
		var it domain.ChecklistItem
		var reqStr, stStr, wfStr string
		err := rows.Scan(
			&it.ID, &it.DossierID, &it.ModuleNumber, &it.SubModule, &it.ItemCode,
			&it.TitleFR, &it.TitleAR, &it.DescriptionFR,
			&reqStr, &it.IsMandatory, &it.GEDDocumentID, &stStr,
			&wfStr, &it.CurrentRevision,
		)
		if err != nil {
			return nil, err
		}
		it.RequirementLevel = domain.RequirementLevel(reqStr)
		it.Status = domain.ChecklistItemStatus(stStr)
		it.WorkflowStatus = domain.WorkflowStatus(wfStr)
		items = append(items, it)
	}
	return items, nil
}

// CheckoutItem verrouille une section CTD pour rédaction exclusive
func (r *WorkflowRepository) CheckoutItem(ctx context.Context, itemID, userID uuid.UUID, userRole, reason string) (*domain.ChecklistItem, error) {
	item, err := r.GetChecklistItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	if item.IsLocked && item.LockedBy != nil && *item.LockedBy != userID {
		return nil, ErrSectionLocked
	}

	now := time.Now().UTC()
	query := `
		UPDATE dossier_checklist_items
		SET is_locked = TRUE,
		    locked_by = $1,
		    locked_at = $2,
		    lock_reason = $3,
		    workflow_status = CASE WHEN workflow_status = 'draft_pending' THEN 'drafting' ELSE workflow_status END,
		    maker_user_id = $1,
		    updated_at = $2
		WHERE id = $4
	`
	_, err = r.pool.Exec(ctx, query, userID, now, reason, itemID)
	if err != nil {
		return nil, fmt.Errorf("échec verrouillage checkout : %w", err)
	}

	// Ajouter l'entrée dans l'audit trail de revue
	_ = r.AddReviewEntry(ctx, itemID, userID, userRole, domain.ReviewTypeAnnotation, "Verrouillé pour édition (Check-out) : "+reason, item.CurrentRevision)

	return r.GetChecklistItemByID(ctx, itemID)
}

// CheckinItem restitue la section avec sauvegarde de la révision incrémentale v0.x
func (r *WorkflowRepository) CheckinItem(ctx context.Context, itemID, userID uuid.UUID, userRole string, content *string, gedDocID *uuid.UUID, comment string) (*domain.ChecklistItem, error) {
	item, err := r.GetChecklistItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	if !item.IsLocked {
		return nil, ErrSectionNotLocked
	}
	if item.LockedBy != nil && *item.LockedBy != userID {
		return nil, ErrUnauthorizedCheckin
	}

	nextRev := domain.NextMinorRevision(item.CurrentRevision)
	now := time.Now().UTC()

	query := `
		UPDATE dossier_checklist_items
		SET is_locked = FALSE,
		    locked_by = NULL,
		    locked_at = NULL,
		    lock_reason = NULL,
		    current_revision = $1,
		    authored_content = COALESCE($2, authored_content),
		    ged_document_id = COALESCE($3, ged_document_id),
		    status = CASE WHEN $3 IS NOT NULL THEN 'rattache' ELSE status END,
		    updated_at = $4
		WHERE id = $5
	`
	_, err = r.pool.Exec(ctx, query, nextRev, content, gedDocID, now, itemID)
	if err != nil {
		return nil, fmt.Errorf("échec enregistrement checkin : %w", err)
	}

	summary := fmt.Sprintf("Révision %s enregistrée (Check-in) : %s", nextRev, comment)
	_ = r.AddReviewEntry(ctx, itemID, userID, userRole, domain.ReviewTypeAnnotation, summary, nextRev)

	return r.GetChecklistItemByID(ctx, itemID)
}

// AssignItem attribue la responsabilité de la section à un collaborateur ou pôle métier
func (r *WorkflowRepository) AssignItem(ctx context.Context, itemID uuid.UUID, assignedUserID *uuid.UUID, dept string, dueDate *time.Time) (*domain.ChecklistItem, error) {
	now := time.Now().UTC()
	query := `
		UPDATE dossier_checklist_items
		SET assigned_user_id = $1,
		    assigned_department = $2,
		    due_date = $3,
		    updated_at = $4
		WHERE id = $5
	`
	_, err := r.pool.Exec(ctx, query, assignedUserID, dept, dueDate, now, itemID)
	if err != nil {
		return nil, fmt.Errorf("échec assignation de la section : %w", err)
	}
	return r.GetChecklistItemByID(ctx, itemID)
}

// SubmitForReview soumet la section complétée au relecteur qualité (Checker / AQ)
func (r *WorkflowRepository) SubmitForReview(ctx context.Context, itemID, makerID uuid.UUID, makerRole, notes string) (*domain.ChecklistItem, error) {
	item, err := r.GetChecklistItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	if item.IsLocked {
		return nil, errors.New("veuillez effectuer le Check-in de la section avant de la soumettre en revue")
	}

	// Contrôle des verrous réglementaires chronologiques (Gating CTD)
	allItems, err := r.GetDossierItems(ctx, item.DossierID)
	if err != nil {
		return nil, fmt.Errorf("contrôle des verrous réglementaires impossible : %w", err)
	}

	gatingResult := domain.EvaluateItemGating(allItems, item, domain.GatingActionSubmitReview)
	if gatingResult.IsBlocked {
		return nil, &GatingBlockedError{Result: gatingResult}
	}

	now := time.Now().UTC()
	query := `
		UPDATE dossier_checklist_items
		SET workflow_status = 'under_review',
		    maker_user_id = $1,
		    maker_submitted_at = $2,
		    updated_at = $2
		WHERE id = $3
	`
	_, err = r.pool.Exec(ctx, query, makerID, now, itemID)
	if err != nil {
		return nil, fmt.Errorf("échec soumission en revue : %w", err)
	}

	_ = r.AddReviewEntry(ctx, itemID, makerID, makerRole, domain.ReviewTypeAnnotation, "Projet soumis pour relecture qualité (Checker/AQ) : "+notes, item.CurrentRevision)

	return r.GetChecklistItemByID(ctx, itemID)
}

// ReviewItem valide ou rejette le document lors de l'étape de contrôle qualité (Checker)
func (r *WorkflowRepository) ReviewItem(ctx context.Context, itemID, checkerID uuid.UUID, checkerRole, decision, notes string) (*domain.ChecklistItem, error) {
	item, err := r.GetChecklistItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	// Garde-fou BPF SoD : non-auto-approbation (le Maker ne peut pas être le Checker)
	if item.MakerUserID != nil && *item.MakerUserID == checkerID {
		return nil, ErrSelfReviewNotAllowed
	}

	now := time.Now().UTC()
	var newWfStatus string
	var revType domain.ReviewType

	if decision == "approved" {
		newWfStatus = "pending_approval" // Prêt pour visa DT
		revType = domain.ReviewTypeQualityPassed
	} else {
		newWfStatus = "drafting" // Renvoyé en rédaction
		revType = domain.ReviewTypeRejected
	}

	query := `
		UPDATE dossier_checklist_items
		SET workflow_status = $1,
		    checker_user_id = $2,
		    checker_reviewed_at = $3,
		    checker_decision = $4,
		    checker_notes = $5,
		    updated_at = $3
		WHERE id = $6
	`
	_, err = r.pool.Exec(ctx, query, newWfStatus, checkerID, now, decision, notes, itemID)
	if err != nil {
		return nil, fmt.Errorf("échec enregistrement revue qualité : %w", err)
	}

	_ = r.AddReviewEntry(ctx, itemID, checkerID, checkerRole, revType, notes, item.CurrentRevision)

	return r.GetChecklistItemByID(ctx, itemID)
}

// ApproveItem appose le visa légal officiel 21 CFR Part 11 du Pharmacien DT (Approver) et scelle le document en v1.0
func (r *WorkflowRepository) ApproveItem(ctx context.Context, itemID, approverID uuid.UUID, approverRole, seal string) (*domain.ChecklistItem, error) {
	item, err := r.GetChecklistItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	// Garde-fou SoD : le Maker ne peut pas être l'Approver unique
	if item.MakerUserID != nil && *item.MakerUserID == approverID {
		return nil, ErrSelfReviewNotAllowed
	}

	// Contrôle des verrous réglementaires chronologiques (Gating CTD)
	allItems, err := r.GetDossierItems(ctx, item.DossierID)
	if err != nil {
		return nil, fmt.Errorf("contrôle des verrous réglementaires impossible : %w", err)
	}

	gatingResult := domain.EvaluateItemGating(allItems, item, domain.GatingActionApprovePDT)
	if gatingResult.IsBlocked {
		return nil, &GatingBlockedError{Result: gatingResult}
	}

	now := time.Now().UTC()
	query := `
		UPDATE dossier_checklist_items
		SET workflow_status = 'approved_sealed',
		    current_revision = 'v1.0',
		    status = 'approuve_dt',
		    approver_user_id = $1,
		    approver_signed_at = $2,
		    approver_signature_seal = $3,
		    updated_at = $2
		WHERE id = $4
	`
	_, err = r.pool.Exec(ctx, query, approverID, now, seal, itemID)
	if err != nil {
		return nil, fmt.Errorf("échec scellement visa DT : %w", err)
	}

	_ = r.AddReviewEntry(ctx, itemID, approverID, approverRole, domain.ReviewTypeQualityPassed, "Visa légal apposé sous 21 CFR Part 11 par le Pharmacien DT. Document officiellement scellé v1.0.", "v1.0")

	return r.GetChecklistItemByID(ctx, itemID)
}

// AddReviewEntry ajoute une ligne immuable dans l'audit trail des annotations
func (r *WorkflowRepository) AddReviewEntry(ctx context.Context, itemID, authorID uuid.UUID, role string, revType domain.ReviewType, comment, version string) error {
	query := `
		INSERT INTO dossier_item_reviews (dossier_item_id, author_id, author_role, review_type, comment, revision_version)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, itemID, authorID, role, string(revType), comment, version)
	return err
}

// GetItemReviews liste toutes les annotations et revues d'un item
func (r *WorkflowRepository) GetItemReviews(ctx context.Context, itemID uuid.UUID) ([]domain.DossierItemReview, error) {
	query := `
		SELECT r.id, r.dossier_item_id, r.author_id, r.author_role, r.review_type, r.comment, r.revision_version, r.created_at,
		       COALESCE(u.first_name || ' ' || u.last_name, '')
		FROM dossier_item_reviews r
		LEFT JOIN users u ON r.author_id = u.id
		WHERE r.dossier_item_id = $1
		ORDER BY r.created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, itemID)
	if err != nil {
		return nil, fmt.Errorf("échec lecture revues item : %w", err)
	}
	defer rows.Close()

	var list []domain.DossierItemReview
	for rows.Next() {
		var rev domain.DossierItemReview
		var tStr, authorName string
		err := rows.Scan(
			&rev.ID, &rev.DossierItemID, &rev.AuthorID, &rev.AuthorRole, &tStr, &rev.Comment, &rev.RevisionVersion, &rev.CreatedAt, &authorName,
		)
		if err != nil {
			return nil, err
		}
		rev.ReviewType = domain.ReviewType(tStr)
		rev.AuthorName = authorName
		list = append(list, rev)
	}
	return list, nil
}

// GetMasterTemplates liste les modèles maîtres pharmaceutiques
func (r *WorkflowRepository) GetMasterTemplates(ctx context.Context, category string) ([]domain.MasterDocumentTemplate, error) {
	query := `
		SELECT id, code, title_fr, COALESCE(title_ar, ''), category, module_number, sub_module, target_procedure, content_template, is_active, created_at, updated_at
		FROM master_document_templates
		WHERE is_active = TRUE AND ($1 = '' OR category = $1)
		ORDER BY module_number ASC, code ASC
	`
	rows, err := r.pool.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("échec lecture templates : %w", err)
	}
	defer rows.Close()

	var list []domain.MasterDocumentTemplate
	for rows.Next() {
		var t domain.MasterDocumentTemplate
		err := rows.Scan(&t.ID, &t.Code, &t.TitleFR, &t.TitleAR, &t.Category, &t.ModuleNumber, &t.SubModule, &t.TargetProcedure, &t.ContentTemplate, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

// GetMasterTemplateByCode charge un template par son code
func (r *WorkflowRepository) GetMasterTemplateByCode(ctx context.Context, code string) (*domain.MasterDocumentTemplate, error) {
	query := `
		SELECT id, code, title_fr, COALESCE(title_ar, ''), category, module_number, sub_module, target_procedure, content_template, is_active, created_at, updated_at
		FROM master_document_templates
		WHERE code = $1 AND is_active = TRUE
	`
	var t domain.MasterDocumentTemplate
	err := r.pool.QueryRow(ctx, query, code).Scan(&t.ID, &t.Code, &t.TitleFR, &t.TitleAR, &t.Category, &t.ModuleNumber, &t.SubModule, &t.TargetProcedure, &t.ContentTemplate, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkflowItemNotFound
		}
		return nil, err
	}
	return &t, nil
}

// SaveGeneratedReport sauvegarde un état de sortie officiel avec son empreinte SHA-256
func (r *WorkflowRepository) SaveGeneratedReport(ctx context.Context, rep *domain.GeneratedReport) error {
	metaJSON, _ := json.Marshal(rep.ReportMetadata)
	query := `
		INSERT INTO generated_reports (
			id, organization_id, dossier_id, report_code, report_title, format, watermark_status,
			sha256_hash, file_size_bytes, signed_by_pdt_id, signed_at, qr_verification_url, report_metadata, created_by, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (id) DO UPDATE SET
			watermark_status = EXCLUDED.watermark_status,
			sha256_hash = EXCLUDED.sha256_hash,
			signed_by_pdt_id = EXCLUDED.signed_by_pdt_id,
			signed_at = EXCLUDED.signed_at,
			qr_verification_url = EXCLUDED.qr_verification_url,
			report_metadata = EXCLUDED.report_metadata
	`
	if rep.ID == uuid.Nil {
		rep.ID = uuid.New()
	}
	if rep.CreatedAt.IsZero() {
		rep.CreatedAt = time.Now().UTC()
	}

	_, err := r.pool.Exec(ctx, query,
		rep.ID, rep.OrganizationID, rep.DossierID, rep.ReportCode, rep.ReportTitle, rep.Format,
		string(rep.WatermarkStatus), rep.SHA256Hash, rep.FileSizeBytes, rep.SignedByPDTID,
		rep.SignedAt, rep.QRVerificationURL, metaJSON, rep.CreatedBy, rep.CreatedAt,
	)
	return err
}

// GetGeneratedReport charge un rapport par code pour un dossier
func (r *WorkflowRepository) GetGeneratedReport(ctx context.Context, orgID, dossierID uuid.UUID, reportCode string) (*domain.GeneratedReport, error) {
	query := `
		SELECT g.id, g.organization_id, g.dossier_id, g.report_code, g.report_title, g.format,
		       g.watermark_status, g.sha256_hash, g.file_size_bytes, g.signed_by_pdt_id, g.signed_at,
		       g.qr_verification_url, g.report_metadata, g.created_by, g.created_at,
		       COALESCE(u.first_name || ' ' || u.last_name, '')
		FROM generated_reports g
		LEFT JOIN users u ON g.signed_by_pdt_id = u.id
		WHERE g.organization_id = $1 AND g.dossier_id = $2 AND g.report_code = $3
		ORDER BY g.created_at DESC
		LIMIT 1
	`
	var rep domain.GeneratedReport
	var wmStr, pdtName string
	var metaBytes []byte

	err := r.pool.QueryRow(ctx, query, orgID, dossierID, reportCode).Scan(
		&rep.ID, &rep.OrganizationID, &rep.DossierID, &rep.ReportCode, &rep.ReportTitle, &rep.Format,
		&wmStr, &rep.SHA256Hash, &rep.FileSizeBytes, &rep.SignedByPDTID, &rep.SignedAt,
		&rep.QRVerificationURL, &metaBytes, &rep.CreatedBy, &rep.CreatedAt, &pdtName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkflowItemNotFound
		}
		return nil, err
	}
	rep.WatermarkStatus = domain.WatermarkStatus(wmStr)
	rep.SignedByPDTName = pdtName
	_ = json.Unmarshal(metaBytes, &rep.ReportMetadata)
	return &rep, nil
}

// GetReportBySHA256 charge un rapport pour le portail public de vérification QR-Code
func (r *WorkflowRepository) GetReportBySHA256(ctx context.Context, sha256 string) (*domain.GeneratedReport, error) {
	query := `
		SELECT g.id, g.organization_id, g.dossier_id, g.report_code, g.report_title, g.format,
		       g.watermark_status, g.sha256_hash, g.file_size_bytes, g.signed_by_pdt_id, g.signed_at,
		       g.qr_verification_url, g.report_metadata, g.created_by, g.created_at,
		       COALESCE(u.first_name || ' ' || u.last_name, '')
		FROM generated_reports g
		LEFT JOIN users u ON g.signed_by_pdt_id = u.id
		WHERE g.sha256_hash = $1
		ORDER BY g.created_at DESC
		LIMIT 1
	`
	var rep domain.GeneratedReport
	var wmStr, pdtName string
	var metaBytes []byte

	err := r.pool.QueryRow(ctx, query, sha256).Scan(
		&rep.ID, &rep.OrganizationID, &rep.DossierID, &rep.ReportCode, &rep.ReportTitle, &rep.Format,
		&wmStr, &rep.SHA256Hash, &rep.FileSizeBytes, &rep.SignedByPDTID, &rep.SignedAt,
		&rep.QRVerificationURL, &metaBytes, &rep.CreatedBy, &rep.CreatedAt, &pdtName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkflowItemNotFound
		}
		return nil, err
	}
	rep.WatermarkStatus = domain.WatermarkStatus(wmStr)
	rep.SignedByPDTName = pdtName
	_ = json.Unmarshal(metaBytes, &rep.ReportMetadata)
	return &rep, nil
}

// EvaluateItemGating évalue l'état des verrous pour une section donnée (utilisé pour l'inspection UI)
func (r *WorkflowRepository) EvaluateItemGating(ctx context.Context, itemID uuid.UUID) (*domain.GatingEvaluationResult, error) {
	item, err := r.GetChecklistItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	allItems, err := r.GetDossierItems(ctx, item.DossierID)
	if err != nil {
		return nil, err
	}

	res := domain.EvaluateItemGating(allItems, item, domain.GatingActionInspection)
	return &res, nil
}

// BlockingPrerequisiteSummary détaille une tâche bloquante prioritaire (clé de voûte)
type BlockingPrerequisiteSummary struct {
	ItemCode     string   `json:"item_code"`
	TitleFR      string   `json:"title_fr"`
	BlocksCount  int      `json:"blocks_count"`
	BlockedCodes []string `json:"blocked_codes"`
}

// DossierGatingSummary représente la synthèse globale du DAG de dépendances du dossier
type DossierGatingSummary struct {
	DossierID       uuid.UUID                       `json:"dossier_id"`
	ActivePhase     int                             `json:"active_phase"`
	ActivePhaseName string                          `json:"active_phase_name"`
	TotalItems      int                             `json:"total_items"`
	BlockedCount    int                             `json:"blocked_count"`
	BlockedItems    []domain.GatingEvaluationResult `json:"blocked_items"`
	BlockingPrereqs []BlockingPrerequisiteSummary   `json:"blocking_prereqs"`
}

// GetDossierGatingSummary calcule la phase DAG active et recense toutes les sections bloquées
func (r *WorkflowRepository) GetDossierGatingSummary(ctx context.Context, dossierID uuid.UUID) (*DossierGatingSummary, error) {
	allItems, err := r.GetDossierItems(ctx, dossierID)
	if err != nil {
		return nil, err
	}

	activePhase := domain.CalculateDossierDAGPhase(allItems)
	phaseNames := map[int]string{
		1: "Phase 1 : Méthodes Analytiques & Spécifications Matières",
		2: "Phase 2 : Validation Procédé & Lots Pilotes",
		3: "Phase 3 : Stabilités Zone IVb & Bioéquivalence",
		4: "Phase 4 : QOS (Module 2.3), Notice & RCP",
		5: "Phase 5 : Engagements Réglementaires & Dépôt Légal PDT",
	}

	summary := &DossierGatingSummary{
		DossierID:       dossierID,
		ActivePhase:     activePhase,
		ActivePhaseName: phaseNames[activePhase],
		TotalItems:      len(allItems),
		BlockedItems:    make([]domain.GatingEvaluationResult, 0),
		BlockingPrereqs: make([]BlockingPrerequisiteSummary, 0),
	}

	for i := range allItems {
		it := &allItems[i]
		res := domain.EvaluateItemGating(allItems, it, domain.GatingActionInspection)
		if res.IsBlocked {
			summary.BlockedItems = append(summary.BlockedItems, res)
		}
	}
	summary.BlockedCount = len(summary.BlockedItems)

	// Recenser les pièces bloquantes (débloqueurs critiques)
	blockingMap := make(map[string]*BlockingPrerequisiteSummary)
	titlesByCode := make(map[string]string)
	for _, it := range allItems {
		titlesByCode[it.ItemCode] = it.TitleFR
	}

	for _, b := range summary.BlockedItems {
		for _, u := range b.UnmetPrerequisites {
			reqCode := u.RequiredItemCode
			if reqCode == "" {
				continue
			}
			bp, exists := blockingMap[reqCode]
			if !exists {
				title := u.RequiredTitle
				if t, ok := titlesByCode[reqCode]; ok && t != "" {
					title = t
				}
				bp = &BlockingPrerequisiteSummary{
					ItemCode:     reqCode,
					TitleFR:      title,
					BlocksCount:  0,
					BlockedCodes: make([]string, 0),
				}
				blockingMap[reqCode] = bp
			}
			bp.BlocksCount++
			bp.BlockedCodes = append(bp.BlockedCodes, b.TargetItemCode)
		}
	}

	for _, bp := range blockingMap {
		summary.BlockingPrereqs = append(summary.BlockingPrereqs, *bp)
	}

	return summary, nil
}

// GetDossierSmartPlanning calcule le rétroplanning dynamique inversé et l'analyse de santé du chemin critique
func (r *WorkflowRepository) GetDossierSmartPlanning(ctx context.Context, dossierID uuid.UUID) (*domain.SmartPlanningAnalysis, error) {
	dossierQuery := `
		SELECT id, organization_id, dossier_code, procedure_code, internal_reference,
		       anpp_deposit_date, anpp_validity_end_date, target_deposit_date,
		       tracking_number, deposit_receipt_ref, status, is_locked_for_deposit,
		       pdt_signed_at, created_at
		FROM dossiers
		WHERE id = $1
	`
	var d domain.Dossier
	var procCodeStr, statusStr string

	err := r.pool.QueryRow(ctx, dossierQuery, dossierID).Scan(
		&d.ID, &d.OrganizationID, &d.DossierCode, &procCodeStr, &d.InternalReference,
		&d.ANPPDepositDate, &d.ANPPValidityEndDate, &d.TargetDepositDate,
		&d.TrackingNumber, &d.DepositReceiptRef, &statusStr, &d.IsLockedForDeposit,
		&d.PDTSignedAt, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("échec lecture dossier pour smart planning : %w", err)
	}
	d.ProcedureCode = domain.ProcedureCode(procCodeStr)
	d.Status = domain.DossierStatus(statusStr)

	items, err := r.GetDossierItems(ctx, dossierID)
	if err != nil {
		return nil, fmt.Errorf("échec lecture items pour smart planning : %w", err)
	}

	analysis := domain.EvaluateSmartPlanning(&d, items, time.Now())
	return analysis, nil
}

// UpdateDossierTargetDates permet de fixer ou ajuster la date cible de dépôt ou l'échéance de validité AMM
func (r *WorkflowRepository) UpdateDossierTargetDates(ctx context.Context, dossierID uuid.UUID, targetDate, validityEndDate *time.Time) error {
	query := `
		UPDATE dossiers
		SET target_deposit_date = $2,
		    anpp_validity_end_date = COALESCE($3, anpp_validity_end_date),
		    updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, dossierID, targetDate, validityEndDate)
	if err != nil {
		return fmt.Errorf("échec mise à jour dates cibles dossier : %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SyncSmartPlanningDueDates propage les dates du chemin critique inversé dans les due_date des items du dossier
func (r *WorkflowRepository) SyncSmartPlanningDueDates(ctx context.Context, dossierID uuid.UUID, dueDates map[string]time.Time) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("échec début transaction sync dates : %w", err)
	}
	defer tx.Rollback(ctx)

	updatedCount := 0
	for itemCode, dueDate := range dueDates {
		dStr := dueDate.Format("2006-01-02")
		query := `
			UPDATE dossier_checklist_items
			SET due_date = $2,
			    updated_at = NOW()
			WHERE dossier_id = $1 AND (item_code = $3 OR item_code LIKE $4)
		`
		tag, err := tx.Exec(ctx, query, dossierID, dStr, itemCode, itemCode+".%")
		if err != nil {
			return 0, fmt.Errorf("échec sync due_date item %s : %w", itemCode, err)
		}
		updatedCount += int(tag.RowsAffected())
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("échec commit sync dates : %w", err)
	}

	return updatedCount, nil
}
