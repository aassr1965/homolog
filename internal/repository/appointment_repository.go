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

// AppointmentRepository assure la persistance des rendez-vous et séances de dépôt ANPP
type AppointmentRepository struct {
	pool *pgxpool.Pool
}

// NewAppointmentRepository instancie le repository des rendez-vous ANPP
func NewAppointmentRepository(pool *pgxpool.Pool) *AppointmentRepository {
	return &AppointmentRepository{pool: pool}
}

// CreateAppointment enregistre un nouveau créneau officiel de séance ou dépôt
func (r *AppointmentRepository) CreateAppointment(ctx context.Context, a *domain.ANPPAppointment) (*domain.ANPPAppointment, error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	now := time.Now().UTC()
	a.CreatedAt = now
	a.UpdatedAt = now

	query := `
		INSERT INTO anpp_deposit_appointments (
			id, organization_id, dossier_id, appointment_type, scheduled_at,
			location, slot_reference, contact_officer, assigned_delegate_id,
			status, notes, confirmation_document_id, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15
		)
	`
	_, err := r.pool.Exec(ctx, query,
		a.ID, a.OrganizationID, a.DossierID, string(a.AppointmentType), a.ScheduledAt,
		a.Location, a.SlotReference, a.ContactOfficer, a.AssignedDelegateID,
		string(a.Status), a.Notes, a.ConfirmationDocumentID, a.CreatedBy, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("échec création rendez-vous ANPP: %w", err)
	}

	return r.GetAppointment(ctx, a.OrganizationID, a.ID)
}

// GetAppointment retourne un rendez-vous avec jointure sur le dossier et l'utilisateur délégué
func (r *AppointmentRepository) GetAppointment(ctx context.Context, orgID, appID uuid.UUID) (*domain.ANPPAppointment, error) {
	query := `
		SELECT 
			a.id, a.organization_id, a.dossier_id,
			COALESCE(d.dossier_code, '') AS dossier_code,
			COALESCE(p_med.brand_name_fr, p_dm.commercial_name_fr, '') AS product_brand_name,
			a.appointment_type, a.scheduled_at, a.location, a.slot_reference,
			COALESCE(a.contact_officer, ''), a.assigned_delegate_id,
			COALESCE(TRIM(CONCAT(u.first_name, ' ', u.last_name)), '') AS assigned_delegate_name,
			a.status, COALESCE(a.notes, ''), a.confirmation_document_id, a.created_by,
			a.created_at, a.updated_at
		FROM anpp_deposit_appointments a
		LEFT JOIN dossiers d ON a.dossier_id = d.id
		LEFT JOIN medicines p_med ON d.medicine_id = p_med.id
		LEFT JOIN medical_devices p_dm ON d.medical_device_id = p_dm.id
		LEFT JOIN users u ON a.assigned_delegate_id = u.id
		WHERE a.organization_id = $1 AND a.id = $2
	`

	var a domain.ANPPAppointment
	var appTypeStr, statusStr string

	err := r.pool.QueryRow(ctx, query, orgID, appID).Scan(
		&a.ID, &a.OrganizationID, &a.DossierID,
		&a.DossierCode, &a.ProductBrandName,
		&appTypeStr, &a.ScheduledAt, &a.Location, &a.SlotReference,
		&a.ContactOfficer, &a.AssignedDelegateID, &a.AssignedDelegateName,
		&statusStr, &a.Notes, &a.ConfirmationDocumentID, &a.CreatedBy,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("échec lecture rendez-vous ANPP: %w", err)
	}

	a.AppointmentType = domain.AppointmentType(appTypeStr)
	a.Status = domain.AppointmentStatus(statusStr)
	return &a, nil
}

// ListAppointments liste les créneaux avec filtres optionnels
func (r *AppointmentRepository) ListAppointments(ctx context.Context, orgID uuid.UUID, dossierID *uuid.UUID, statusFilter string, upcomingOnly bool) ([]domain.ANPPAppointment, error) {
	query := `
		SELECT 
			a.id, a.organization_id, a.dossier_id,
			COALESCE(d.dossier_code, '') AS dossier_code,
			COALESCE(p_med.brand_name_fr, p_dm.commercial_name_fr, '') AS product_brand_name,
			a.appointment_type, a.scheduled_at, a.location, a.slot_reference,
			COALESCE(a.contact_officer, ''), a.assigned_delegate_id,
			COALESCE(TRIM(CONCAT(u.first_name, ' ', u.last_name)), '') AS assigned_delegate_name,
			a.status, COALESCE(a.notes, ''), a.confirmation_document_id, a.created_by,
			a.created_at, a.updated_at
		FROM anpp_deposit_appointments a
		LEFT JOIN dossiers d ON a.dossier_id = d.id
		LEFT JOIN medicines p_med ON d.medicine_id = p_med.id
		LEFT JOIN medical_devices p_dm ON d.medical_device_id = p_dm.id
		LEFT JOIN users u ON a.assigned_delegate_id = u.id
		WHERE a.organization_id = $1
	`
	args := []any{orgID}
	argIdx := 2

	if dossierID != nil {
		query += fmt.Sprintf(" AND a.dossier_id = $%d", argIdx)
		args = append(args, *dossierID)
		argIdx++
	}

	if statusFilter != "" {
		query += fmt.Sprintf(" AND a.status = $%d", argIdx)
		args = append(args, statusFilter)
		argIdx++
	}

	if upcomingOnly {
		query += " AND a.scheduled_at >= NOW() AND a.status IN ('planifie', 'confirme_anpp')"
	}

	query += " ORDER BY a.scheduled_at ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("échec liste rendez-vous ANPP: %w", err)
	}
	defer rows.Close()

	var list []domain.ANPPAppointment
	for rows.Next() {
		var a domain.ANPPAppointment
		var appTypeStr, statusStr string

		err := rows.Scan(
			&a.ID, &a.OrganizationID, &a.DossierID,
			&a.DossierCode, &a.ProductBrandName,
			&appTypeStr, &a.ScheduledAt, &a.Location, &a.SlotReference,
			&a.ContactOfficer, &a.AssignedDelegateID, &a.AssignedDelegateName,
			&statusStr, &a.Notes, &a.ConfirmationDocumentID, &a.CreatedBy,
			&a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("échec scan rendez-vous ANPP: %w", err)
		}

		a.AppointmentType = domain.AppointmentType(appTypeStr)
		a.Status = domain.AppointmentStatus(statusStr)
		list = append(list, a)
	}

	if list == nil {
		list = []domain.ANPPAppointment{}
	}
	return list, nil
}

// UpdateAppointmentStatus met à jour l'état d'un rendez-vous (ex: confirme_anpp, effectue, reporte)
func (r *AppointmentRepository) UpdateAppointmentStatus(ctx context.Context, orgID, appID uuid.UUID, newStatus domain.AppointmentStatus, notes *string) (*domain.ANPPAppointment, error) {
	now := time.Now().UTC()
	query := `
		UPDATE anpp_deposit_appointments
		SET status = $1,
		    notes = CASE WHEN $2::text IS NOT NULL AND $2::text != '' THEN $2::text ELSE notes END,
		    updated_at = $3
		WHERE organization_id = $4 AND id = $5
	`
	cmd, err := r.pool.Exec(ctx, query, string(newStatus), notes, now, orgID, appID)
	if err != nil {
		return nil, fmt.Errorf("échec mise à jour statut rendez-vous: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return r.GetAppointment(ctx, orgID, appID)
}
