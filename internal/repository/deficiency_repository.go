package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/domain"
)

var (
	ErrLetterNotFound = errors.New("lettre de compléments ANPP introuvable")
	ErrTaskNotFound   = errors.New("tâche de complément introuvable")
)

// DeficiencyRepository gère le suivi des lettres et questions de compléments ANPP
type DeficiencyRepository struct {
	pool *pgxpool.Pool
}

// NewDeficiencyRepository instancie le repository des compléments ANPP
func NewDeficiencyRepository(pool *pgxpool.Pool) *DeficiencyRepository {
	return &DeficiencyRepository{pool: pool}
}

// CreateLetterWithTasks insère une nouvelle notification officielle ANPP et génère les tâches unitaires décomposées
func (r *DeficiencyRepository) CreateLetterWithTasks(ctx context.Context, letter *domain.ANPPDeficiencyLetter, tasks []domain.ANPPDeficiencyTask) (*domain.ANPPDeficiencyLetter, error) {
	if letter.ID == "" {
		letter.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	letter.CreatedAt = now
	letter.UpdatedAt = now
	if letter.Status == "" {
		letter.Status = domain.DeficiencyLetterStatusOpen
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("échec début transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insertion de la lettre officielle
	insertLetterQuery := `
		INSERT INTO anpp_deficiency_letters (
			id, organization_id, dossier_id, reference_number,
			received_date, legal_deadline_days, due_date, status,
			document_id, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = tx.Exec(ctx, insertLetterQuery,
		letter.ID,
		letter.OrganizationID,
		letter.DossierID,
		letter.ReferenceNumber,
		letter.ReceivedDate,
		letter.LegalDeadlineDays,
		letter.DueDate,
		letter.Status,
		letter.DocumentID,
		letter.Notes,
		letter.CreatedAt,
		letter.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insertion lettre de compléments impossible: %w", err)
	}

	// Insertion des tâches unitaires associées
	insertTaskQuery := `
		INSERT INTO anpp_deficiency_tasks (
			id, letter_id, dossier_id, organization_id, question_number,
			question_text, module_target, priority, assigned_to,
			due_date, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	createdTasks := make([]domain.ANPPDeficiencyTask, 0, len(tasks))
	for i, t := range tasks {
		taskID := uuid.New().String()
		t.ID = taskID
		t.LetterID = &letter.ID
		t.DossierID = letter.DossierID
		t.OrganizationID = letter.OrganizationID
		if t.QuestionNumber <= 0 {
			t.QuestionNumber = i + 1
		}
		if t.Status == "" {
			t.Status = domain.DeficiencyTaskStatusPending
		}
		if t.Priority == "" {
			t.Priority = "normale"
		}
		if t.ModuleTarget == "" {
			t.ModuleTarget = "General"
		}
		if t.DueDate.IsZero() {
			t.DueDate = letter.DueDate
		}
		t.CreatedAt = now
		t.UpdatedAt = now

		_, err = tx.Exec(ctx, insertTaskQuery,
			t.ID,
			t.LetterID,
			t.DossierID,
			t.OrganizationID,
			t.QuestionNumber,
			t.QuestionText,
			t.ModuleTarget,
			t.Priority,
			t.AssignedTo,
			t.DueDate,
			t.Status,
			t.CreatedAt,
			t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("insertion tâche #%d impossible: %w", t.QuestionNumber, err)
		}
		createdTasks = append(createdTasks, t)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("échec commit transaction lettre compléments: %w", err)
	}

	letter.Tasks = createdTasks
	letter.ComputeMetrics(now)
	return letter, nil
}

// GetLetter récupère une lettre de compléments par son ID avec toutes ses tâches
func (r *DeficiencyRepository) GetLetter(ctx context.Context, orgID, letterID string) (*domain.ANPPDeficiencyLetter, error) {
	query := `
		SELECT 
			l.id, l.organization_id, l.dossier_id, l.reference_number,
			l.received_date, l.legal_deadline_days, l.due_date, l.status,
			l.document_id, COALESCE(doc.title_fr, '') AS document_title,
			COALESCE(l.notes, '') AS notes, l.created_at, l.updated_at,
			d.dossier_code, COALESCE(p_med.brand_name_fr, p_dm.commercial_name_fr, '') AS product_name
		FROM anpp_deficiency_letters l
		JOIN dossiers d ON d.id = l.dossier_id
		LEFT JOIN medicines p_med ON p_med.id = d.medicine_id
		LEFT JOIN medical_devices p_dm ON p_dm.id = d.medical_device_id
		LEFT JOIN ged_documents doc ON doc.id = l.document_id
		WHERE l.id = $1 AND l.organization_id = $2
	`
	var l domain.ANPPDeficiencyLetter
	err := r.pool.QueryRow(ctx, query, letterID, orgID).Scan(
		&l.ID, &l.OrganizationID, &l.DossierID, &l.ReferenceNumber,
		&l.ReceivedDate, &l.LegalDeadlineDays, &l.DueDate, &l.Status,
		&l.DocumentID, &l.DocumentTitle, &l.Notes, &l.CreatedAt, &l.UpdatedAt,
		&l.DossierReference, &l.ProductBrandName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLetterNotFound
		}
		return nil, fmt.Errorf("requête lettre compléments échouée: %w", err)
	}

	// Récupération des tâches
	tasks, err := r.ListTasksByLetter(ctx, orgID, letterID)
	if err != nil {
		return nil, err
	}
	l.Tasks = tasks
	l.ComputeMetrics(time.Now().UTC())

	return &l, nil
}

// ListTasksByLetter retourne la liste des tâches associées à une lettre
func (r *DeficiencyRepository) ListTasksByLetter(ctx context.Context, orgID, letterID string) ([]domain.ANPPDeficiencyTask, error) {
	query := `
		SELECT 
			t.id, t.letter_id, t.dossier_id, t.organization_id,
			t.question_number, t.question_text, COALESCE(t.module_target, 'General'),
			COALESCE(t.priority, 'normale'), t.assigned_to,
			COALESCE(TRIM(CONCAT(u_ass.first_name, ' ', u_ass.last_name)), '') AS assigned_user_name,
			COALESCE(u_ass.role, '') AS assigned_user_role,
			t.due_date, t.status,
			t.response_document_id, COALESCE(doc_resp.title_fr, '') AS response_document_title,
			COALESCE(t.response_summary, ''),
			t.validated_by_dt_id, COALESCE(TRIM(CONCAT(u_dt.first_name, ' ', u_dt.last_name)), '') AS validated_by_name,
			t.validated_at, t.created_at, t.updated_at
		FROM anpp_deficiency_tasks t
		LEFT JOIN users u_ass ON u_ass.id = t.assigned_to
		LEFT JOIN users u_dt ON u_dt.id = t.validated_by_dt_id
		LEFT JOIN ged_documents doc_resp ON doc_resp.id = t.response_document_id
		WHERE t.letter_id = $1 AND t.organization_id = $2
		ORDER BY t.question_number ASC
	`
	rows, err := r.pool.Query(ctx, query, letterID, orgID)
	if err != nil {
		return nil, fmt.Errorf("requête tâches compléments échouée: %w", err)
	}
	defer rows.Close()

	tasks := make([]domain.ANPPDeficiencyTask, 0)
	for rows.Next() {
		var t domain.ANPPDeficiencyTask
		err := rows.Scan(
			&t.ID, &t.LetterID, &t.DossierID, &t.OrganizationID,
			&t.QuestionNumber, &t.QuestionText, &t.ModuleTarget,
			&t.Priority, &t.AssignedTo,
			&t.AssignedUserName, &t.AssignedUserRole,
			&t.DueDate, &t.Status,
			&t.ResponseDocumentID, &t.ResponseDocumentTitle,
			&t.ResponseSummary,
			&t.ValidatedByDTID, &t.ValidatedByName,
			&t.ValidatedAt, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tâche complément échoué: %w", err)
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// ListLettersByDossier retourne toutes les lettres de compléments pour un dossier
func (r *DeficiencyRepository) ListLettersByDossier(ctx context.Context, orgID, dossierID string) ([]domain.ANPPDeficiencyLetter, error) {
	query := `
		SELECT 
			l.id, l.organization_id, l.dossier_id, l.reference_number,
			l.received_date, l.legal_deadline_days, l.due_date, l.status,
			l.document_id, COALESCE(doc.title_fr, '') AS document_title,
			COALESCE(l.notes, '') AS notes, l.created_at, l.updated_at,
			d.dossier_code, COALESCE(p_med.brand_name_fr, p_dm.commercial_name_fr, '') AS product_name
		FROM anpp_deficiency_letters l
		JOIN dossiers d ON d.id = l.dossier_id
		LEFT JOIN medicines p_med ON p_med.id = d.medicine_id
		LEFT JOIN medical_devices p_dm ON p_dm.id = d.medical_device_id
		LEFT JOIN ged_documents doc ON doc.id = l.document_id
		WHERE l.dossier_id = $1 AND l.organization_id = $2
		ORDER BY l.received_date DESC
	`
	rows, err := r.pool.Query(ctx, query, dossierID, orgID)
	if err != nil {
		return nil, fmt.Errorf("requête liste lettres dossier échouée: %w", err)
	}
	defer rows.Close()

	now := time.Now().UTC()
	letters := make([]domain.ANPPDeficiencyLetter, 0)
	for rows.Next() {
		var l domain.ANPPDeficiencyLetter
		err := rows.Scan(
			&l.ID, &l.OrganizationID, &l.DossierID, &l.ReferenceNumber,
			&l.ReceivedDate, &l.LegalDeadlineDays, &l.DueDate, &l.Status,
			&l.DocumentID, &l.DocumentTitle, &l.Notes, &l.CreatedAt, &l.UpdatedAt,
			&l.DossierReference, &l.ProductBrandName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan lettre compléments échoué: %w", err)
		}
		letters = append(letters, l)
	}

	// Charger les tâches pour chaque lettre
	for i := range letters {
		tasks, err := r.ListTasksByLetter(ctx, orgID, letters[i].ID)
		if err == nil {
			letters[i].Tasks = tasks
		}
		letters[i].ComputeMetrics(now)
	}

	return letters, nil
}

// ListAllActiveDeficiencies retourne la synthèse globale et les alertes pour l'organisation
func (r *DeficiencyRepository) ListAllActiveDeficiencies(ctx context.Context, orgID string) (*domain.DeficiencyDashboardSummary, error) {
	query := `
		SELECT 
			l.id, l.organization_id, l.dossier_id, l.reference_number,
			l.received_date, l.legal_deadline_days, l.due_date, l.status,
			l.document_id, COALESCE(doc.title_fr, '') AS document_title,
			COALESCE(l.notes, '') AS notes, l.created_at, l.updated_at,
			d.dossier_code, COALESCE(p_med.brand_name_fr, p_dm.commercial_name_fr, '') AS product_name
		FROM anpp_deficiency_letters l
		JOIN dossiers d ON d.id = l.dossier_id
		LEFT JOIN medicines p_med ON p_med.id = d.medicine_id
		LEFT JOIN medical_devices p_dm ON p_dm.id = d.medical_device_id
		LEFT JOIN ged_documents doc ON doc.id = l.document_id
		WHERE l.organization_id = $1 AND l.status != $2
		ORDER BY l.due_date ASC
	`
	rows, err := r.pool.Query(ctx, query, orgID, domain.DeficiencyLetterStatusSubmitted)
	if err != nil {
		return nil, fmt.Errorf("requête dashboard compléments échouée: %w", err)
	}
	defer rows.Close()

	now := time.Now().UTC()
	summary := &domain.DeficiencyDashboardSummary{
		RecentLetters: make([]domain.ANPPDeficiencyLetter, 0),
	}

	for rows.Next() {
		var l domain.ANPPDeficiencyLetter
		err := rows.Scan(
			&l.ID, &l.OrganizationID, &l.DossierID, &l.ReferenceNumber,
			&l.ReceivedDate, &l.LegalDeadlineDays, &l.DueDate, &l.Status,
			&l.DocumentID, &l.DocumentTitle, &l.Notes, &l.CreatedAt, &l.UpdatedAt,
			&l.DossierReference, &l.ProductBrandName,
		)
		if err != nil {
			return nil, fmt.Errorf("scan dashboard compléments échoué: %w", err)
		}

		tasks, err := r.ListTasksByLetter(ctx, orgID, l.ID)
		if err == nil {
			l.Tasks = tasks
		}
		l.ComputeMetrics(now)

		summary.TotalActiveLetters++
		if l.UrgencyBadge == "critical" || l.UrgencyBadge == "expired" {
			summary.CriticalLetters++
		} else if l.UrgencyBadge == "urgent" {
			summary.UrgentLetters++
		}

		for _, t := range l.Tasks {
			if t.Status == domain.DeficiencyTaskStatusPending || t.Status == domain.DeficiencyTaskStatusInProgress {
				summary.PendingTasksCount++
			} else if t.Status == domain.DeficiencyTaskStatusValidatedDT {
				summary.ValidatedTasksCount++
			}
		}

		summary.RecentLetters = append(summary.RecentLetters, l)
	}

	return summary, nil
}

// GetTask récupère une tâche spécifique
func (r *DeficiencyRepository) GetTask(ctx context.Context, orgID, taskID string) (*domain.ANPPDeficiencyTask, error) {
	query := `
		SELECT 
			t.id, t.letter_id, t.dossier_id, t.organization_id,
			t.question_number, t.question_text, COALESCE(t.module_target, 'General'),
			COALESCE(t.priority, 'normale'), t.assigned_to,
			COALESCE(TRIM(CONCAT(u_ass.first_name, ' ', u_ass.last_name)), '') AS assigned_user_name,
			COALESCE(u_ass.role, '') AS assigned_user_role,
			t.due_date, t.status,
			t.response_document_id, COALESCE(doc_resp.title_fr, '') AS response_document_title,
			COALESCE(t.response_summary, ''),
			t.validated_by_dt_id, COALESCE(TRIM(CONCAT(u_dt.first_name, ' ', u_dt.last_name)), '') AS validated_by_name,
			t.validated_at, t.created_at, t.updated_at
		FROM anpp_deficiency_tasks t
		LEFT JOIN users u_ass ON u_ass.id = t.assigned_to
		LEFT JOIN users u_dt ON u_dt.id = t.validated_by_dt_id
		LEFT JOIN ged_documents doc_resp ON doc_resp.id = t.response_document_id
		WHERE t.id = $1 AND t.organization_id = $2
	`
	var t domain.ANPPDeficiencyTask
	err := r.pool.QueryRow(ctx, query, taskID, orgID).Scan(
		&t.ID, &t.LetterID, &t.DossierID, &t.OrganizationID,
		&t.QuestionNumber, &t.QuestionText, &t.ModuleTarget,
		&t.Priority, &t.AssignedTo,
		&t.AssignedUserName, &t.AssignedUserRole,
		&t.DueDate, &t.Status,
		&t.ResponseDocumentID, &t.ResponseDocumentTitle,
		&t.ResponseSummary,
		&t.ValidatedByDTID, &t.ValidatedByName,
		&t.ValidatedAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("requête tâche complément échouée: %w", err)
	}

	t.DaysRemaining, _ = domain.CalculateDeadlineAndUrgency(t.DueDate, time.Now().UTC())
	return &t, nil
}

// UpdateTask met à jour la réponse, le document joint ou l'assignation d'une tâche
func (r *DeficiencyRepository) UpdateTask(ctx context.Context, task *domain.ANPPDeficiencyTask) error {
	task.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE anpp_deficiency_tasks
		SET 
			status = $1,
			response_document_id = $2,
			response_summary = $3,
			assigned_to = $4,
			priority = $5,
			due_date = $6,
			updated_at = $7
		WHERE id = $8 AND organization_id = $9
	`
	res, err := r.pool.Exec(ctx, query,
		task.Status,
		task.ResponseDocumentID,
		task.ResponseSummary,
		task.AssignedTo,
		task.Priority,
		task.DueDate,
		task.UpdatedAt,
		task.ID,
		task.OrganizationID,
	)
	if err != nil {
		return fmt.Errorf("mise à jour tâche complément échouée: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// ValidateTaskDT appose le visa officiel du Pharmacien Directeur Technique sur une réponse à un complément
func (r *DeficiencyRepository) ValidateTaskDT(ctx context.Context, orgID, taskID, dtUserID, comment string) error {
	now := time.Now().UTC()
	query := `
		UPDATE anpp_deficiency_tasks
		SET 
			status = $1,
			validated_by_dt_id = $2,
			validated_at = $3,
			response_summary = CASE 
				WHEN $4 <> '' THEN COALESCE(response_summary, '') || E'\n[Visa PDT]: ' || $4
				ELSE response_summary 
			END,
			updated_at = $3
		WHERE id = $5 AND organization_id = $6
	`
	res, err := r.pool.Exec(ctx, query,
		domain.DeficiencyTaskStatusValidatedDT,
		dtUserID,
		now,
		comment,
		taskID,
		orgID,
	)
	if err != nil {
		return fmt.Errorf("visa Pharmacien DT sur tâche échoué: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// AttachDocumentToLetter rattache ou met à jour le document officiel GED (scan/décharge) sur une lettre
func (r *DeficiencyRepository) AttachDocumentToLetter(ctx context.Context, orgID, letterID, documentID string) error {
	now := time.Now().UTC()
	query := `
		UPDATE anpp_deficiency_letters
		SET document_id = $1, updated_at = $2
		WHERE id = $3 AND organization_id = $4
	`
	res, err := r.pool.Exec(ctx, query, documentID, now, letterID, orgID)
	if err != nil {
		return fmt.Errorf("rattachement pièce GED sur lettre échoué: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrLetterNotFound
	}
	return nil
}
