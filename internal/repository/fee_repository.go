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
	ErrFeeScheduleNotFound = errors.New("barème de redevance ANPP introuvable pour cette procédure")
	ErrPaymentSlipNotFound = errors.New("bordereau de versement introuvable")
)

// FeeRepository gère la persistance des barèmes de redevances et des bordereaux de versement
type FeeRepository struct {
	pool *pgxpool.Pool
}

// NewFeeRepository instancie le repository des redevances
func NewFeeRepository(pool *pgxpool.Pool) *FeeRepository {
	return &FeeRepository{pool: pool}
}

// ListFeeSchedules retourne l'ensemble des grilles tarifaires
func (r *FeeRepository) ListFeeSchedules(ctx context.Context, onlyActive bool) ([]domain.ANPPFeeSchedule, error) {
	query := `
		SELECT 
			id, category, procedure_code, label_fr, COALESCE(label_ar, ''),
			legal_reference, amount_dzd, valid_from, valid_to, is_active,
			COALESCE(notes, ''), created_at, updated_at
		FROM anpp_fee_schedules
	`
	if onlyActive {
		query += " WHERE is_active = TRUE AND (valid_to IS NULL OR valid_to >= CURRENT_DATE)"
	}
	query += " ORDER BY category ASC, procedure_code ASC, valid_from DESC"

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("requête barèmes échouée: %w", err)
	}
	defer rows.Close()

	schedules := make([]domain.ANPPFeeSchedule, 0)
	for rows.Next() {
		var s domain.ANPPFeeSchedule
		err := rows.Scan(
			&s.ID, &s.Category, &s.ProcedureCode, &s.LabelFR, &s.LabelAR,
			&s.LegalReference, &s.AmountDZD, &s.ValidFrom, &s.ValidTo, &s.IsActive,
			&s.Notes, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan barème échoué: %w", err)
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}

// GetFeeScheduleForProcedure recherche le barème légal applicable pour une procédure à une date donnée
func (r *FeeRepository) GetFeeScheduleForProcedure(ctx context.Context, procedureCode string, targetDate time.Time) (*domain.ANPPFeeSchedule, error) {
	query := `
		SELECT 
			id, category, procedure_code, label_fr, COALESCE(label_ar, ''),
			legal_reference, amount_dzd, valid_from, valid_to, is_active,
			COALESCE(notes, ''), created_at, updated_at
		FROM anpp_fee_schedules
		WHERE procedure_code = $1 
		  AND valid_from <= $2
		  AND (valid_to IS NULL OR valid_to >= $2)
		ORDER BY valid_from DESC
		LIMIT 1
	`
	var s domain.ANPPFeeSchedule
	err := r.pool.QueryRow(ctx, query, procedureCode, targetDate).Scan(
		&s.ID, &s.Category, &s.ProcedureCode, &s.LabelFR, &s.LabelAR,
		&s.LegalReference, &s.AmountDZD, &s.ValidFrom, &s.ValidTo, &s.IsActive,
		&s.Notes, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFeeScheduleNotFound
		}
		return nil, fmt.Errorf("recherche barème pour %s échouée: %w", procedureCode, err)
	}
	return &s, nil
}

// CreateOrUpdateFeeSchedule enregistre une nouvelle grille tarifaire et clôture automatiquement l'ancienne
func (r *FeeRepository) CreateOrUpdateFeeSchedule(ctx context.Context, s *domain.ANPPFeeSchedule) (*domain.ANPPFeeSchedule, error) {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now
	s.IsActive = true

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("début transaction barème échoué: %w", err)
	}
	defer tx.Rollback(ctx)

	// Clôturer l'ancien tarif en vigueur pour cette procédure à la veille de la nouvelle date d'effet
	dayBefore := s.ValidFrom.AddDate(0, 0, -1)
	closeOldQuery := `
		UPDATE anpp_fee_schedules
		SET valid_to = $1, is_active = FALSE, updated_at = $2
		WHERE procedure_code = $3 AND (valid_to IS NULL OR valid_to > $1) AND id <> $4
	`
	_, err = tx.Exec(ctx, closeOldQuery, dayBefore, now, s.ProcedureCode, s.ID)
	if err != nil {
		return nil, fmt.Errorf("clôture ancien barème échouée: %w", err)
	}

	// Insérer la nouvelle grille
	insertQuery := `
		INSERT INTO anpp_fee_schedules (
			id, category, procedure_code, label_fr, label_ar,
			legal_reference, amount_dzd, valid_from, valid_to,
			is_active, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.Exec(ctx, insertQuery,
		s.ID, s.Category, s.ProcedureCode, s.LabelFR, s.LabelAR,
		s.LegalReference, s.AmountDZD, s.ValidFrom, s.ValidTo,
		s.IsActive, s.Notes, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insertion nouveau barème échouée: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit barème échoué: %w", err)
	}

	return s, nil
}

// CreatePaymentSlip insère un nouveau bordereau de versement BV1 ou BV2
func (r *FeeRepository) CreatePaymentSlip(ctx context.Context, slip *domain.ANPPPaymentSlip) (*domain.ANPPPaymentSlip, error) {
	if slip.ID == "" {
		slip.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	slip.CreatedAt = now
	slip.UpdatedAt = now
	if slip.GeneratedAt.IsZero() {
		slip.GeneratedAt = now
	}
	if slip.Status == "" {
		slip.Status = domain.PaymentStatusEmis
	}
	if slip.BankAccountRIB == "" {
		slip.BankAccountRIB = "00100624030000001234 (Compte Trésorerie Centrale ANPP Alger)"
	}

	query := `
		INSERT INTO anpp_payment_slips (
			id, organization_id, dossier_id, fee_schedule_id,
			slip_type, slip_number, payment_purpose, amount_dzd,
			amount_words_fr, amount_words_ar, bank_account_rib,
			status, generated_by, generated_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.pool.Exec(ctx, query,
		slip.ID, slip.OrganizationID, slip.DossierID, slip.FeeScheduleID,
		slip.SlipType, slip.SlipNumber, slip.PaymentPurpose, slip.AmountDZD,
		slip.AmountWordsFR, slip.AmountWordsAR, slip.BankAccountRIB,
		slip.Status, slip.GeneratedBy, slip.GeneratedAt, slip.CreatedAt, slip.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("création bordereau versement échouée: %w", err)
	}

	return slip, nil
}

// GetPaymentSlip retourne un bordereau de versement enrichi avec les informations du dossier et produit
func (r *FeeRepository) GetPaymentSlip(ctx context.Context, orgID, slipID string) (*domain.ANPPPaymentSlip, error) {
	query := `
		SELECT 
			s.id, s.organization_id, org.legal_name AS organization_name,
			s.dossier_id, d.dossier_code,
			COALESCE(p_med.brand_name_fr, p_dm.commercial_name_fr, '') AS product_brand_name,
			COALESCE(est.name_fr, '') AS establishment_name,
			s.fee_schedule_id, COALESCE(sch.legal_reference, '') AS fee_schedule_reference,
			s.slip_type, s.slip_number, s.payment_purpose, s.amount_dzd,
			s.amount_words_fr, s.amount_words_ar, s.bank_account_rib,
			s.status, COALESCE(s.payment_reference, ''), s.paid_at,
			s.receipt_document_id, COALESCE(doc.title_fr, '') AS receipt_document_title,
			s.generated_by, COALESCE(TRIM(CONCAT(u.first_name, ' ', u.last_name)), '') AS generated_by_name,
			s.generated_at, s.created_at, s.updated_at
		FROM anpp_payment_slips s
		JOIN organizations org ON org.id = s.organization_id
		JOIN dossiers d ON d.id = s.dossier_id
		LEFT JOIN medicines p_med ON p_med.id = d.medicine_id
		LEFT JOIN medical_devices p_dm ON p_dm.id = d.medical_device_id
		LEFT JOIN establishments est ON est.id = org.id
		LEFT JOIN anpp_fee_schedules sch ON sch.id = s.fee_schedule_id
		LEFT JOIN ged_documents doc ON doc.id = s.receipt_document_id
		LEFT JOIN users u ON u.id = s.generated_by
		WHERE s.id = $1 AND s.organization_id = $2
	`
	var slip domain.ANPPPaymentSlip
	err := r.pool.QueryRow(ctx, query, slipID, orgID).Scan(
		&slip.ID, &slip.OrganizationID, &slip.OrganizationName,
		&slip.DossierID, &slip.DossierCode,
		&slip.ProductBrandName, &slip.EstablishmentName,
		&slip.FeeScheduleID, &slip.FeeScheduleReference,
		&slip.SlipType, &slip.SlipNumber, &slip.PaymentPurpose, &slip.AmountDZD,
		&slip.AmountWordsFR, &slip.AmountWordsAR, &slip.BankAccountRIB,
		&slip.Status, &slip.PaymentReference, &slip.PaidAt,
		&slip.ReceiptDocumentID, &slip.ReceiptDocumentTitle,
		&slip.GeneratedBy, &slip.GeneratedByName,
		&slip.GeneratedAt, &slip.CreatedAt, &slip.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentSlipNotFound
		}
		return nil, fmt.Errorf("requête bordereau versement échouée: %w", err)
	}
	return &slip, nil
}

// ListPaymentSlipsByDossier retourne tous les bordereaux émis pour un dossier
func (r *FeeRepository) ListPaymentSlipsByDossier(ctx context.Context, orgID, dossierID string) ([]domain.ANPPPaymentSlip, error) {
	query := `
		SELECT 
			s.id, s.organization_id, org.legal_name AS organization_name,
			s.dossier_id, d.dossier_code,
			COALESCE(p_med.brand_name_fr, p_dm.commercial_name_fr, '') AS product_brand_name,
			COALESCE(est.name_fr, '') AS establishment_name,
			s.fee_schedule_id, COALESCE(sch.legal_reference, '') AS fee_schedule_reference,
			s.slip_type, s.slip_number, s.payment_purpose, s.amount_dzd,
			s.amount_words_fr, s.amount_words_ar, s.bank_account_rib,
			s.status, COALESCE(s.payment_reference, ''), s.paid_at,
			s.receipt_document_id, COALESCE(doc.title_fr, '') AS receipt_document_title,
			s.generated_by, COALESCE(TRIM(CONCAT(u.first_name, ' ', u.last_name)), '') AS generated_by_name,
			s.generated_at, s.created_at, s.updated_at
		FROM anpp_payment_slips s
		JOIN organizations org ON org.id = s.organization_id
		JOIN dossiers d ON d.id = s.dossier_id
		LEFT JOIN medicines p_med ON p_med.id = d.medicine_id
		LEFT JOIN medical_devices p_dm ON p_dm.id = d.medical_device_id
		LEFT JOIN establishments est ON est.id = org.id
		LEFT JOIN anpp_fee_schedules sch ON sch.id = s.fee_schedule_id
		LEFT JOIN ged_documents doc ON doc.id = s.receipt_document_id
		LEFT JOIN users u ON u.id = s.generated_by
		WHERE s.dossier_id = $1 AND s.organization_id = $2
		ORDER BY s.generated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, dossierID, orgID)
	if err != nil {
		return nil, fmt.Errorf("requête liste bordereaux échouée: %w", err)
	}
	defer rows.Close()

	slips := make([]domain.ANPPPaymentSlip, 0)
	for rows.Next() {
		var slip domain.ANPPPaymentSlip
		err := rows.Scan(
			&slip.ID, &slip.OrganizationID, &slip.OrganizationName,
			&slip.DossierID, &slip.DossierCode,
			&slip.ProductBrandName, &slip.EstablishmentName,
			&slip.FeeScheduleID, &slip.FeeScheduleReference,
			&slip.SlipType, &slip.SlipNumber, &slip.PaymentPurpose, &slip.AmountDZD,
			&slip.AmountWordsFR, &slip.AmountWordsAR, &slip.BankAccountRIB,
			&slip.Status, &slip.PaymentReference, &slip.PaidAt,
			&slip.ReceiptDocumentID, &slip.ReceiptDocumentTitle,
			&slip.GeneratedBy, &slip.GeneratedByName,
			&slip.GeneratedAt, &slip.CreatedAt, &slip.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan bordereau échoué: %w", err)
		}
		slips = append(slips, slip)
	}
	return slips, nil
}

// UpdatePaymentSlipStatus met à jour le statut du bordereau (marquage payé, quittance Trésor)
func (r *FeeRepository) UpdatePaymentSlipStatus(ctx context.Context, orgID, slipID, status, paymentRef string, paidAt *time.Time, receiptDocID *string) error {
	now := time.Now().UTC()
	query := `
		UPDATE anpp_payment_slips
		SET 
			status = $1,
			payment_reference = $2,
			paid_at = $3,
			receipt_document_id = $4,
			updated_at = $5
		WHERE id = $6 AND organization_id = $7
	`
	res, err := r.pool.Exec(ctx, query,
		status,
		paymentRef,
		paidAt,
		receiptDocID,
		now,
		slipID,
		orgID,
	)
	if err != nil {
		return fmt.Errorf("mise à jour bordereau échouée: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrPaymentSlipNotFound
	}
	return nil
}
