package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// GetChecklistItemByID charge un item de checklist avec son contexte de workflow en vérifiant l'isolation multi-tenant
func (r *WorkflowRepository) GetChecklistItemByID(ctx context.Context, orgID, itemID uuid.UUID) (*domain.ChecklistItem, error) {
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
		JOIN dossiers d ON d.id = i.dossier_id
		LEFT JOIN users u_asg ON i.assigned_user_id = u_asg.id
		LEFT JOIN users u_lck ON i.locked_by = u_lck.id
		LEFT JOIN users u_mkr ON i.maker_user_id = u_mkr.id
		LEFT JOIN users u_chk ON i.checker_user_id = u_chk.id
		LEFT JOIN users u_app ON i.approver_user_id = u_app.id
		WHERE i.id = $1 AND d.organization_id = $2
	`
	var it domain.ChecklistItem
	var reqStr, stStr, secClassStr, wfStr string
	var titleAR, descFR string
	var asgName, lckName, mkrName, chkName, appName string

	err := r.pool.QueryRow(ctx, query, itemID, orgID).Scan(
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
func (r *WorkflowRepository) CheckoutItem(ctx context.Context, orgID, itemID, userID uuid.UUID, userRole, reason string) (*domain.ChecklistItem, error) {
	item, err := r.GetChecklistItemByID(ctx, orgID, itemID)
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

	return r.GetChecklistItemByID(ctx, orgID, itemID)
}

// CheckinItem restitue la section avec sauvegarde de la révision incrémentale v0.x
func (r *WorkflowRepository) CheckinItem(ctx context.Context, orgID, itemID, userID uuid.UUID, userRole string, content *string, gedDocID *uuid.UUID, comment string) (*domain.ChecklistItem, error) {
	item, err := r.GetChecklistItemByID(ctx, orgID, itemID)
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

	return r.GetChecklistItemByID(ctx, orgID, itemID)
}

// AssignItem attribue la responsabilité de la section à un collaborateur ou pôle métier
func (r *WorkflowRepository) AssignItem(ctx context.Context, orgID, itemID uuid.UUID, assignedUserID *uuid.UUID, dept string, dueDate *time.Time) (*domain.ChecklistItem, error) {
	_, err := r.GetChecklistItemByID(ctx, orgID, itemID)
	if err != nil {
		return nil, err
	}

	if assignedUserID != nil {
		var userExists bool
		err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND organization_id = $2)`, assignedUserID, orgID).Scan(&userExists)
		if err != nil {
			return nil, fmt.Errorf("erreur vérification utilisateur : %w", err)
		}
		if !userExists {
			return nil, fmt.Errorf("utilisateur introuvable dans votre organisation")
		}
	}

	now := time.Now().UTC()
	query := `
		UPDATE dossier_checklist_items
		SET assigned_user_id = $1,
		    assigned_department = $2,
		    due_date = $3,
		    updated_at = $4
		WHERE id = $5
	`
	_, err = r.pool.Exec(ctx, query, assignedUserID, dept, dueDate, now, itemID)
	if err != nil {
		return nil, fmt.Errorf("échec assignation de la section : %w", err)
	}
	return r.GetChecklistItemByID(ctx, orgID, itemID)
}

// SubmitForReview soumet la section complétée au relecteur qualité (Checker / AQ)
func (r *WorkflowRepository) SubmitForReview(ctx context.Context, orgID, itemID, makerID uuid.UUID, makerRole, notes string) (*domain.ChecklistItem, error) {
	item, err := r.GetChecklistItemByID(ctx, orgID, itemID)
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

	return r.GetChecklistItemByID(ctx, orgID, itemID)
}

// ReviewItem valide ou rejette le document lors de l'étape de contrôle qualité (Checker)
func (r *WorkflowRepository) ReviewItem(ctx context.Context, orgID, itemID, checkerID uuid.UUID, checkerRole, decision, notes string) (*domain.ChecklistItem, error) {
	item, err := r.GetChecklistItemByID(ctx, orgID, itemID)
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

	return r.GetChecklistItemByID(ctx, orgID, itemID)
}

// ApproveItem appose le visa légal officiel 21 CFR Part 11 du Pharmacien DT (Approver) et scelle le document en v1.0
func (r *WorkflowRepository) ApproveItem(ctx context.Context, orgID, itemID, approverID uuid.UUID, approverRole, seal string) (*domain.ChecklistItem, error) {
	item, err := r.GetChecklistItemByID(ctx, orgID, itemID)
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

	return r.GetChecklistItemByID(ctx, orgID, itemID)
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

// GetItemReviews liste toutes les annotations et revues d'un item après vérification multi-tenant
func (r *WorkflowRepository) GetItemReviews(ctx context.Context, orgID, itemID uuid.UUID) ([]domain.DossierItemReview, error) {
	_, err := r.GetChecklistItemByID(ctx, orgID, itemID)
	if err != nil {
		return nil, err
	}

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

// EvaluateItemGating évalue l'état des verrous pour une section donnée (utilisé pour l'inspection UI) en vérifiant l'isolation multi-tenant
func (r *WorkflowRepository) EvaluateItemGating(ctx context.Context, orgID, itemID uuid.UUID) (*domain.GatingEvaluationResult, error) {
	item, err := r.GetChecklistItemByID(ctx, orgID, itemID)
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
func (r *WorkflowRepository) GetDossierGatingSummary(ctx context.Context, orgID, dossierID uuid.UUID) (*DossierGatingSummary, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM dossiers WHERE id = $1 AND organization_id = $2)`, dossierID, orgID).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("erreur vérification appartenance dossier : %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

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
func (r *WorkflowRepository) GetDossierSmartPlanning(ctx context.Context, orgID, dossierID uuid.UUID) (*domain.SmartPlanningAnalysis, error) {
	dossierQuery := `
		SELECT d.id, d.organization_id, d.dossier_code, d.procedure_code, d.internal_reference,
		       d.anpp_deposit_date, d.anpp_validity_end_date, d.target_deposit_date,
		       COALESCE(d.working_days_per_week, o.working_days_per_week, 5),
		       d.tracking_number, d.deposit_receipt_ref, d.status, d.is_locked_for_deposit,
		       d.pdt_signed_at, d.created_at
		FROM dossiers d
		LEFT JOIN organizations o ON o.id = d.organization_id
		WHERE d.id = $1 AND d.organization_id = $2
	`
	var d domain.Dossier
	var procCodeStr, statusStr string

	err := r.pool.QueryRow(ctx, dossierQuery, dossierID, orgID).Scan(
		&d.ID, &d.OrganizationID, &d.DossierCode, &procCodeStr, &d.InternalReference,
		&d.ANPPDepositDate, &d.ANPPValidityEndDate, &d.TargetDepositDate,
		&d.WorkingDaysPerWeek,
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

// UpdateDossierTargetDates permet de fixer ou ajuster la date cible de dépôt, l'échéance de validité AMM et la base de jours ouvrables
func (r *WorkflowRepository) UpdateDossierTargetDates(ctx context.Context, orgID, dossierID uuid.UUID, targetDate, validityEndDate *time.Time, workingDaysPerWeek *int) error {
	query := `
		UPDATE dossiers
		SET target_deposit_date = $3,
		    anpp_validity_end_date = COALESCE($4, anpp_validity_end_date),
		    working_days_per_week = COALESCE($5, working_days_per_week),
		    updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
	`
	tag, err := r.pool.Exec(ctx, query, dossierID, orgID, targetDate, validityEndDate, workingDaysPerWeek)
	if err != nil {
		return fmt.Errorf("échec mise à jour dates cibles dossier : %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SyncSmartPlanningDueDates propage les dates du chemin critique inversé dans les due_date des items du dossier
func (r *WorkflowRepository) SyncSmartPlanningDueDates(ctx context.Context, orgID, dossierID uuid.UUID, dueDates map[string]time.Time) (int, error) {
	// Vérification stricte de l'isolation tenant sur le dossier
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM dossiers WHERE id = $1 AND organization_id = $2)`, dossierID, orgID).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("erreur vérification appartenance dossier : %w", err)
	}
	if !exists {
		return 0, ErrNotFound
	}

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

// GetParapheurPendingItems récupère toutes les pièces au statut 'pending_approval'
// au sein de l'organisation pour le parapheur unifié du Pharmacien Directeur Technique
func (r *WorkflowRepository) GetParapheurPendingItems(ctx context.Context, orgID uuid.UUID) ([]*domain.ParapheurPendingItem, error) {
	query := `
		SELECT 
			i.id,
			i.dossier_id,
			d.dossier_code,
			d.procedure_code,
			d.internal_reference,
			d.target_deposit_date,
			COALESCE(
				NULLIF(TRIM(m.brand_name_fr || ' (' || COALESCE(m.dci_fr, '') || ')'), '()'),
				NULLIF(TRIM(md.commercial_name_fr || ' (' || COALESCE(md.manufacturer_reference, '') || ')'), '()'),
				NULLIF(TRIM(p.brand_name_fr || ' (' || COALESCE(p.dci, '') || ')'), '()'),
				d.internal_reference,
				d.dossier_code
			) as product_name,
			i.module_number,
			COALESCE(i.sub_module, ''),
			i.item_code,
			i.title_fr,
			COALESCE(i.title_ar, ''),
			i.requirement_level,
			i.workflow_status,
			COALESCE(i.current_revision, 'v0.1'),
			COALESCE(i.assigned_department, 'RA'),
			i.due_date,
			i.maker_user_id,
			COALESCE(um.first_name || ' ' || um.last_name, ''),
			i.maker_submitted_at,
			i.checker_user_id,
			COALESCE(uc.first_name || ' ' || uc.last_name, ''),
			i.checker_reviewed_at,
			i.checker_decision,
			i.checker_notes,
			i.ged_document_id,
			gd.title_fr,
			i.created_at,
			i.updated_at
		FROM dossier_checklist_items i
		JOIN dossiers d ON d.id = i.dossier_id
		LEFT JOIN medicines m ON m.id = d.medicine_id
		LEFT JOIN medical_devices md ON md.id = d.medical_device_id
		LEFT JOIN products p ON p.id = d.product_id
		LEFT JOIN users um ON um.id = i.maker_user_id
		LEFT JOIN users uc ON uc.id = i.checker_user_id
		LEFT JOIN ged_documents gd ON gd.id = i.ged_document_id
		WHERE d.organization_id = $1
		  AND i.workflow_status = 'pending_approval'
		  AND d.status != 'archive'
		ORDER BY 
			CASE WHEN d.target_deposit_date IS NOT NULL THEN d.target_deposit_date ELSE '9999-12-31' END ASC,
			i.module_number ASC,
			i.sort_order ASC,
			i.item_code ASC
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("requête parapheur échouée : %w", err)
	}
	defer rows.Close()

	var items []*domain.ParapheurPendingItem
	dossierIDsMap := make(map[uuid.UUID]bool)

	for rows.Next() {
		var it domain.ParapheurPendingItem
		var procCodeStr, reqStr, wfStr string
		var gedTitle *string

		err := rows.Scan(
			&it.ItemID,
			&it.DossierID,
			&it.DossierCode,
			&procCodeStr,
			&it.InternalReference,
			&it.TargetDepositDate,
			&it.ProductName,
			&it.ModuleNumber,
			&it.SubModule,
			&it.ItemCode,
			&it.TitleFR,
			&it.TitleAR,
			&reqStr,
			&wfStr,
			&it.CurrentRevision,
			&it.AssignedDepartment,
			&it.DueDate,
			&it.MakerUserID,
			&it.MakerUserName,
			&it.MakerSubmittedAt,
			&it.CheckerUserID,
			&it.CheckerUserName,
			&it.CheckerReviewedAt,
			&it.CheckerDecision,
			&it.CheckerNotes,
			&it.GEDDocumentID,
			&gedTitle,
			&it.CreatedAt,
			&it.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan parapheur item échoué : %w", err)
		}

		it.ProcedureCode = procCodeStr
		it.RequirementLevel = reqStr
		it.WorkflowStatus = wfStr
		it.GEDDocumentTitle = gedTitle

		if it.TargetDepositDate != nil {
			days := int(time.Until(*it.TargetDepositDate).Hours() / 24)
			it.DaysToTargetDeposit = &days
		}

		items = append(items, &it)
		dossierIDsMap[it.DossierID] = true
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("itération lignes parapheur : %w", err)
	}

	// Évaluation du gating CTD pour chaque item du parapheur
	dossierItemsCache := make(map[uuid.UUID][]domain.ChecklistItem)
	for dID := range dossierIDsMap {
		dItems, err := r.GetDossierItems(ctx, dID)
		if err == nil {
			dossierItemsCache[dID] = dItems
		}
	}

	for _, it := range items {
		dItems, ok := dossierItemsCache[it.DossierID]
		if ok {
			dummyItem := domain.ChecklistItem{
				ID:           it.ItemID,
				DossierID:    it.DossierID,
				ItemCode:     it.ItemCode,
				ModuleNumber: it.ModuleNumber,
				SubModule:    it.SubModule,
			}
			gatingRes := domain.EvaluateItemGating(dItems, &dummyItem, domain.GatingActionApprovePDT)
			if gatingRes.IsBlocked {
				it.IsGatingBlocked = true
				it.GatingBlockReason = gatingRes.SummaryMessage
			}
		}
	}

	return items, nil
}

// BatchApproveSections applique le visa officiel 21 CFR Part 11 du Pharmacien DT sur un lot de pièces documentaires
func (r *WorkflowRepository) BatchApproveSections(
	ctx context.Context,
	orgID, approverID uuid.UUID,
	approverRole, approverEmail, approverName string,
	itemIDs []uuid.UUID,
) (*domain.BatchApproveResponse, error) {
	now := time.Now().UTC()
	response := &domain.BatchApproveResponse{
		Success:        true,
		TotalRequested: len(itemIDs),
		TotalApproved:  0,
		TotalFailed:    0,
		ExecutedAt:     now,
		ApproverID:     approverID,
		ApproverName:   approverName,
		Items:          make([]domain.BatchApproveItemResult, 0, len(itemIDs)),
	}

	dossierCache := make(map[uuid.UUID][]domain.ChecklistItem)

	for _, itemID := range itemIDs {
		itemRes := domain.BatchApproveItemResult{
			ItemID: itemID,
		}

		item, err := r.GetChecklistItemByID(ctx, orgID, itemID)
		if err != nil {
			itemRes.Success = false
			itemRes.ErrorMessage = "Section introuvable ou n'appartenant pas à votre établissement"
			response.TotalFailed++
			response.Items = append(response.Items, itemRes)
			continue
		}

		itemRes.ItemCode = item.ItemCode

		// Récupérer le code dossier
		var dCode string
		_ = r.pool.QueryRow(ctx, `SELECT dossier_code FROM dossiers WHERE id = $1`, item.DossierID).Scan(&dCode)
		itemRes.DossierCode = dCode

		if item.WorkflowStatus != domain.WorkflowStatusPendingApproval {
			itemRes.Success = false
			itemRes.ErrorMessage = fmt.Sprintf("Statut invalide : attendu 'pending_approval', actuel '%s'", item.WorkflowStatus)
			response.TotalFailed++
			response.Items = append(response.Items, itemRes)
			continue
		}

		// Garde-fou SoD : Maker != Approver
		if item.MakerUserID != nil && *item.MakerUserID == approverID {
			itemRes.Success = false
			itemRes.ErrorMessage = "Ségrégation des fonctions BPF : Le rédacteur ne peut pas s'auto-approuver"
			response.TotalFailed++
			response.Items = append(response.Items, itemRes)
			continue
		}

		// Contrôle de Gating
		dItems, ok := dossierCache[item.DossierID]
		if !ok {
			dItems, err = r.GetDossierItems(ctx, item.DossierID)
			if err != nil {
				itemRes.Success = false
				itemRes.ErrorMessage = fmt.Sprintf("Erreur lecture dépendances dossier : %v", err)
				response.TotalFailed++
				response.Items = append(response.Items, itemRes)
				continue
			}
			dossierCache[item.DossierID] = dItems
		}

		gatingRes := domain.EvaluateItemGating(dItems, item, domain.GatingActionApprovePDT)
		if gatingRes.IsBlocked {
			itemRes.Success = false
			itemRes.ErrorMessage = fmt.Sprintf("Verrou Gating CTD actif : %s", gatingRes.SummaryMessage)
			response.TotalFailed++
			response.Items = append(response.Items, itemRes)
			continue
		}

		// Génération du sceau cryptographique SHA-256 individuel 21 CFR Part 11
		sealInput := fmt.Sprintf("HOMOLOG-PDT-BATCH-SEAL:%s:%s:%s:%d", itemID, approverID, approverEmail, now.UnixNano())
		sealBytes := sha256.Sum256([]byte(sealInput))
		sealHash := hex.EncodeToString(sealBytes[:])

		updateQuery := `
			UPDATE dossier_checklist_items
			SET workflow_status = 'approved_sealed',
			    current_revision = 'v1.0',
			    status = 'approuve_dt',
			    approver_user_id = $1,
			    approver_signed_at = $2,
			    approver_signature_seal = $3,
			    is_locked = false,
			    updated_at = $2
			WHERE id = $4
		`
		_, err = r.pool.Exec(ctx, updateQuery, approverID, now, sealHash, itemID)
		if err != nil {
			itemRes.Success = false
			itemRes.ErrorMessage = fmt.Sprintf("Erreur mise à jour base de données : %v", err)
			response.TotalFailed++
			response.Items = append(response.Items, itemRes)
			continue
		}

		// Entrée d'audit trail dans dossier_item_reviews
		reviewComment := "Visa officiel Pharmacien DT en signature par lot (Batch 21 CFR Part 11). Document scellé v1.0."
		if approverRole == "pharmacien_assistant" {
			delegation, _ := r.GetActiveDTDelegation(ctx, orgID, approverID)
			if delegation != nil {
				reviewComment = fmt.Sprintf("Visa officiel Pharmacien DT par intérim (Pharmacien Assistant - Mandat Décret 21-82 Art. 24, Réf MIPH: %s). Document scellé v1.0.", delegation.MIPHNotificationRef)
			}
		}
		_ = r.AddReviewEntry(ctx, itemID, approverID, approverRole, domain.ReviewTypeQualityPassed, reviewComment, "v1.0")

		// Mettre à jour le cache local pour les éventuels déblocages de gating sur les pièces suivantes du même dossier
		for idx := range dItems {
			if dItems[idx].ID == itemID {
				dItems[idx].WorkflowStatus = domain.WorkflowStatusApprovedSealed
				dItems[idx].CurrentRevision = "v1.0"
			}
		}

		itemRes.Success = true
		itemRes.SealHash = sealHash
		response.TotalApproved++
		response.Items = append(response.Items, itemRes)
	}

	if response.TotalFailed > 0 && response.TotalApproved == 0 {
		response.Success = false
	}

	return response, nil
}

// GetActiveDTDelegation retourne le mandat d'intérim actif d'un pharmacien assistant s'il existe (Décret 21-82 Art. 24)
func (r *WorkflowRepository) GetActiveDTDelegation(ctx context.Context, orgID, delegateID uuid.UUID) (*domain.DTDelegation, error) {
	query := `
		SELECT d.id, d.organization_id, d.delegator_id, u1.first_name || ' ' || u1.last_name AS delegator_name,
		       d.delegate_id, u2.first_name || ' ' || u2.last_name AS delegate_name, u2.email AS delegate_email,
		       d.miph_notification_ref, d.reason, d.start_date, d.end_date, d.is_active,
		       d.revoked_at, d.revoked_by, d.revoke_reason, d.created_at, d.updated_at
		FROM dt_delegations d
		JOIN users u1 ON u1.id = d.delegator_id
		JOIN users u2 ON u2.id = d.delegate_id
		WHERE d.organization_id = $1
		  AND d.delegate_id = $2
		  AND d.is_active = TRUE
		  AND d.revoked_at IS NULL
		  AND NOW() >= d.start_date
		  AND NOW() <= d.end_date
		ORDER BY d.created_at DESC
		LIMIT 1
	`
	var d domain.DTDelegation
	err := r.pool.QueryRow(ctx, query, orgID, delegateID).Scan(
		&d.ID, &d.OrganizationID, &d.DelegatorID, &d.DelegatorName,
		&d.DelegateID, &d.DelegateName, &d.DelegateEmail,
		&d.MIPHNotificationRef, &d.Reason, &d.StartDate, &d.EndDate, &d.IsActive,
		&d.RevokedAt, &d.RevokedBy, &d.RevokeReason, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur recherche délégation DT : %w", err)
	}
	return &d, nil
}

// ListDTDelegations liste tous les mandats d'intérim / délégations de l'établissement
func (r *WorkflowRepository) ListDTDelegations(ctx context.Context, orgID uuid.UUID) ([]domain.DTDelegation, error) {
	query := `
		SELECT d.id, d.organization_id, d.delegator_id, u1.first_name || ' ' || u1.last_name AS delegator_name,
		       d.delegate_id, u2.first_name || ' ' || u2.last_name AS delegate_name, u2.email AS delegate_email,
		       d.miph_notification_ref, d.reason, d.start_date, d.end_date, d.is_active,
		       d.revoked_at, d.revoked_by, d.revoke_reason, d.created_at, d.updated_at
		FROM dt_delegations d
		JOIN users u1 ON u1.id = d.delegator_id
		JOIN users u2 ON u2.id = d.delegate_id
		WHERE d.organization_id = $1
		ORDER BY d.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("erreur liste délégations DT : %w", err)
	}
	defer rows.Close()

	delegations := make([]domain.DTDelegation, 0)
	for rows.Next() {
		var d domain.DTDelegation
		if err := rows.Scan(
			&d.ID, &d.OrganizationID, &d.DelegatorID, &d.DelegatorName,
			&d.DelegateID, &d.DelegateName, &d.DelegateEmail,
			&d.MIPHNotificationRef, &d.Reason, &d.StartDate, &d.EndDate, &d.IsActive,
			&d.RevokedAt, &d.RevokedBy, &d.RevokeReason, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erreur scan délégation DT : %w", err)
		}
		delegations = append(delegations, d)
	}
	return delegations, nil
}

// CreateDTDelegation crée un mandat d'intérim officiel conforme Décret 21-82 Art. 24 (durée max 3 mois)
func (r *WorkflowRepository) CreateDTDelegation(ctx context.Context, orgID, delegatorID, delegateID uuid.UUID, miphRef, reason string, startDate, endDate time.Time) (*domain.DTDelegation, error) {
	if !endDate.After(startDate) {
		return nil, fmt.Errorf("la date de fin doit être postérieure à la date de début")
	}

	// Garde-fou réglementaire Décret 21-82 Art. 24 : Durée maximale d'un intérim = 3 mois (92 jours)
	maxDuration := 92 * 24 * time.Hour
	if endDate.Sub(startDate) > maxDuration {
		return nil, fmt.Errorf("dépassement de la durée légale (Décret 21-82 Art. 24) : Un mandat d'intérim DT ne peut excéder trois (3) mois")
	}

	// Vérifier que le délégataire appartient bien à la même organisation
	var delegateRole string
	err := r.pool.QueryRow(ctx, "SELECT role FROM users WHERE id = $1 AND organization_id = $2", delegateID, orgID).Scan(&delegateRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("le pharmacien assistant désigné n'appartient pas à votre établissement")
		}
		return nil, fmt.Errorf("erreur vérification délégataire : %w", err)
	}

	query := `
		INSERT INTO dt_delegations (organization_id, delegator_id, delegate_id, miph_notification_ref, reason, start_date, end_date, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
		RETURNING id, created_at, updated_at
	`
	var d domain.DTDelegation
	d.OrganizationID = orgID
	d.DelegatorID = delegatorID
	d.DelegateID = delegateID
	d.MIPHNotificationRef = miphRef
	d.Reason = reason
	d.StartDate = startDate
	d.EndDate = endDate
	d.IsActive = true

	err = r.pool.QueryRow(ctx, query, orgID, delegatorID, delegateID, miphRef, reason, startDate, endDate).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("erreur insertion délégation DT : %w", err)
	}

	// Renseigner les noms
	_ = r.pool.QueryRow(ctx, "SELECT first_name || ' ' || last_name FROM users WHERE id = $1", delegatorID).Scan(&d.DelegatorName)
	_ = r.pool.QueryRow(ctx, "SELECT first_name || ' ' || last_name, email FROM users WHERE id = $1", delegateID).Scan(&d.DelegateName, &d.DelegateEmail)

	return &d, nil
}

// RevokeDTDelegation révoque un mandat d'intérim
func (r *WorkflowRepository) RevokeDTDelegation(ctx context.Context, orgID, delegationID, revokedByID uuid.UUID, reason string) error {
	query := `
		UPDATE dt_delegations
		SET is_active = FALSE,
		    revoked_at = NOW(),
		    revoked_by = $1,
		    revoke_reason = $2,
		    updated_at = NOW()
		WHERE id = $3 AND organization_id = $4 AND is_active = TRUE
	`
	tag, err := r.pool.Exec(ctx, query, revokedByID, reason, delegationID, orgID)
	if err != nil {
		return fmt.Errorf("erreur révocation délégation : %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("mandat d'intérim introuvable ou déjà révoqué")
	}
	return nil
}
