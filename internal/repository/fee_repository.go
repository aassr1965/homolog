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

const paymentSlipSelectFields = `
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
	s.generated_at, s.created_at, s.updated_at,
	COALESCE(s.currency, 'DZD'), s.currency_amount, COALESCE(s.bank_name, ''),
	COALESCE(s.payment_mode, 'virement'), COALESCE(s.bank_receipt_ref, ''), s.bank_receipt_date,
	COALESCE(s.requires_fx_attestation, FALSE), s.fx_rate, s.fx_rate_date,
	COALESCE(s.fx_attestation_ref, ''), s.fx_attestation_doc_id, COALESCE(fx_doc.title_fr, ''),
	COALESCE(s.verification_status, 'pending'), s.verified_by,
	COALESCE(TRIM(CONCAT(v_u.first_name, ' ', v_u.last_name)), '') AS verified_by_name,
	s.verified_at, COALESCE(s.verification_notes, ''),
	COALESCE(s.amount_concordance_ok, FALSE), COALESCE(s.sod_check_passed, FALSE)
`

const paymentSlipFromJoin = `
	FROM anpp_payment_slips s
	JOIN organizations org ON org.id = s.organization_id
	JOIN dossiers d ON d.id = s.dossier_id
	LEFT JOIN medicines p_med ON p_med.id = d.medicine_id
	LEFT JOIN medical_devices p_dm ON p_dm.id = d.medical_device_id
	LEFT JOIN establishments est ON est.id = org.id
	LEFT JOIN anpp_fee_schedules sch ON sch.id = s.fee_schedule_id
	LEFT JOIN ged_documents doc ON doc.id = s.receipt_document_id
	LEFT JOIN ged_documents fx_doc ON fx_doc.id = s.fx_attestation_doc_id
	LEFT JOIN users u ON u.id = s.generated_by
	LEFT JOIN users v_u ON v_u.id = s.verified_by
`

func scanSlipRow(row pgx.Row, slip *domain.ANPPPaymentSlip) error {
	return row.Scan(
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
		&slip.Currency, &slip.CurrencyAmount, &slip.BankName,
		&slip.PaymentMode, &slip.BankReceiptRef, &slip.BankReceiptDate,
		&slip.RequiresFXAttestation, &slip.FXRate, &slip.FXRateDate,
		&slip.FXAttestationRef, &slip.FXAttestationDocID, &slip.FXAttestationDocTitle,
		&slip.VerificationStatus, &slip.VerifiedBy,
		&slip.VerifiedByName,
		&slip.VerifiedAt, &slip.VerificationNotes,
		&slip.AmountConcordanceOK, &slip.SoDCheckPassed,
	)
}

// GetPaymentSlip retourne un bordereau de versement enrichi avec les informations du dossier et produit
func (r *FeeRepository) GetPaymentSlip(ctx context.Context, orgID, slipID string) (*domain.ANPPPaymentSlip, error) {
	query := "SELECT " + paymentSlipSelectFields + " " + paymentSlipFromJoin + " WHERE s.id = $1 AND s.organization_id = $2"
	var slip domain.ANPPPaymentSlip
	err := scanSlipRow(r.pool.QueryRow(ctx, query, slipID, orgID), &slip)
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
	query := "SELECT " + paymentSlipSelectFields + " " + paymentSlipFromJoin + " WHERE s.dossier_id = $1 AND s.organization_id = $2 ORDER BY s.generated_at DESC"
	rows, err := r.pool.Query(ctx, query, dossierID, orgID)
	if err != nil {
		return nil, fmt.Errorf("requête liste bordereaux échouée: %w", err)
	}
	defer rows.Close()

	slips := make([]domain.ANPPPaymentSlip, 0)
	for rows.Next() {
		var slip domain.ANPPPaymentSlip
		err := scanSlipRow(rows, &slip)
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

// VerifyPaymentSlip1A4 applique le contrôle de concordance financière et de l'attestation Banque d'Algérie
func (r *FeeRepository) VerifyPaymentSlip1A4(
	ctx context.Context,
	orgID, slipID, verifierID, verifierName string,
	req domain.VerifyPaymentSlip1A4Request,
	expectedAmount float64,
	isImported bool,
) (*domain.ANPPPaymentSlip, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("début transaction vérification 1.A.4 échouée: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Récupération du bordereau
	var slip domain.ANPPPaymentSlip
	queryGet := "SELECT " + paymentSlipSelectFields + " " + paymentSlipFromJoin + " WHERE s.id = $1 AND s.organization_id = $2 FOR UPDATE OF s"
	if err := scanSlipRow(tx.QueryRow(ctx, queryGet, slipID, orgID), &slip); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentSlipNotFound
		}
		return nil, fmt.Errorf("lecture bordereau pour vérification échouée: %w", err)
	}

	// 2. Contrôle de concordance financière
	amountOK := slip.AmountDZD >= expectedAmount

	// 3. Contrôle de l'attestation de taux de change Banque d'Algérie
	requiresFX := req.RequiresFXAttestation
	if isImported && req.Currency != "" && req.Currency != "DZD" {
		requiresFX = true
	}

	fxOK := true
	if requiresFX {
		if req.FXRate == nil || *req.FXRate <= 0 || req.FXAttestationDocID == nil || req.FXAttestationRef == "" {
			fxOK = false
		}
	}

	// 4. Détermination du statut de conformité
	verifStatus := "conforme"
	payStatus := "quittance_recue"
	if !amountOK || !fxOK {
		verifStatus = "non_conforme"
		payStatus = "paye" // Payé mais incomplet/non conforme
	}

	// Parsing dates
	var paidDate *time.Time
	if req.BankReceiptDate != "" {
		if t, err := time.Parse("2006-01-02", req.BankReceiptDate); err == nil {
			paidDate = &t
		}
	}
	var fxDate *time.Time
	if req.FXRateDate != nil && *req.FXRateDate != "" {
		if t, err := time.Parse("2006-01-02", *req.FXRateDate); err == nil {
			fxDate = &t
		}
	}

	currency := req.Currency
	if currency == "" {
		currency = "DZD"
	}
	paymentMode := req.PaymentMode
	if paymentMode == "" {
		paymentMode = "virement"
	}

	now := time.Now().UTC()

	// 5. Mise à jour de anpp_payment_slips
	updateSlipQuery := `
		UPDATE anpp_payment_slips
		SET
			currency = $1,
			currency_amount = $2,
			bank_name = $3,
			payment_mode = $4,
			bank_receipt_ref = $5,
			bank_receipt_date = $6,
			receipt_document_id = $7,
			requires_fx_attestation = $8,
			fx_rate = $9,
			fx_rate_date = $10,
			fx_attestation_ref = $11,
			fx_attestation_doc_id = $12,
			verification_status = $13,
			verified_by = $14,
			verified_at = $15,
			verification_notes = $16,
			amount_concordance_ok = $17,
			sod_check_passed = TRUE,
			status = $18,
			payment_reference = $19,
			paid_at = $20,
			updated_at = $21
		WHERE id = $22 AND organization_id = $23
	`
	_, err = tx.Exec(ctx, updateSlipQuery,
		currency,
		req.CurrencyAmount,
		req.BankName,
		paymentMode,
		req.BankReceiptRef,
		paidDate,
		req.ReceiptDocumentID,
		requiresFX,
		req.FXRate,
		fxDate,
		req.FXAttestationRef,
		req.FXAttestationDocID,
		verifStatus,
		verifierID,
		now,
		req.VerificationNotes,
		amountOK,
		payStatus,
		req.BankReceiptRef,
		paidDate,
		now,
		slipID,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("mise à jour vérification bordereau échouée: %w", err)
	}

	// 6. Si conforme, mise à jour de la checklist CTD (item 1.A.4)
	if verifStatus == "conforme" && req.ReceiptDocumentID != nil {
		notes := fmt.Sprintf("Conformité 1.A.4 scellée le %s : Bordereau %s acquitté (%.2f DZD), Quittance %s (%s).",
			now.Format("02/01/2006"), slip.SlipNumber, slip.AmountDZD, req.BankReceiptRef, req.BankName)
		if requiresFX && req.FXRate != nil {
			notes += fmt.Sprintf(" Attestation Change BA n° %s (1 %s = %.4f DZD).", req.FXAttestationRef, currency, *req.FXRate)
		}
		notes += fmt.Sprintf(" Visa vérificateur : %s.", verifierName)

		updateChecklistQuery := `
			UPDATE dossier_checklist_items
			SET 
				ged_document_id = $1,
				status = 'valide_qualite',
				validation_notes = $2,
				updated_at = $3
			WHERE dossier_id = $4 AND item_code = '1.A.4'
		`
		_, err = tx.Exec(ctx, updateChecklistQuery, req.ReceiptDocumentID, notes, now, slip.DossierID)
		if err != nil {
			return nil, fmt.Errorf("mise à jour checklist 1.A.4 échouée: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit vérification 1.A.4 échoué: %w", err)
	}

	return r.GetPaymentSlip(ctx, orgID, slipID)
}

// GetSection1A4Summary compile un état des lieux exhaustif de la section 1.A.4 pour un dossier
func (r *FeeRepository) GetSection1A4Summary(ctx context.Context, orgID, dossierID string) (*domain.Section1A4Summary, error) {
	// 1. Lire le dossier
	var dossierCode, procCode, productType string
	err := r.pool.QueryRow(ctx, `
		SELECT dossier_code, COALESCE(procedure_code, 'NI'), COALESCE(product_type, 'medicament')
		FROM dossiers
		WHERE id = $1 AND organization_id = $2
	`, dossierID, orgID).Scan(&dossierCode, &procCode, &productType)
	if err != nil {
		return nil, fmt.Errorf("dossier introuvable: %w", err)
	}

	isImported := (procCode == "NI" || procCode == "RI" || procCode == "MI" || procCode == "dm_importe")

	// 2. Barème attendu
	feeSched, err := r.GetFeeScheduleForProcedure(ctx, procCode, time.Now())
	var expectedAmt float64
	var feeLabel, legalRef string
	if err == nil && feeSched != nil {
		expectedAmt = feeSched.AmountDZD
		feeLabel = feeSched.LabelFR
		legalRef = feeSched.LegalReference
	} else {
		// Valeurs par défaut si barème spécifique non matché
		if isImported {
			expectedAmt = 250000.00
			feeLabel = "Enregistrement Médicament (Importation)"
			legalRef = "Décret Exécutif 20-325 & Arrêté des Redevances ANPP"
		} else {
			expectedAmt = 150000.00
			feeLabel = "Enregistrement Médicament (Fabrication Locale)"
			legalRef = "Décret Exécutif 20-325 & Arrêté des Redevances ANPP"
		}
	}

	// 3. Lire le statut de la section 1.A.4 dans dossier_checklist_items
	var itemStatus string
	_ = r.pool.QueryRow(ctx, `
		SELECT status FROM dossier_checklist_items
		WHERE dossier_id = $1 AND item_code = '1.A.4'
	`, dossierID).Scan(&itemStatus)
	if itemStatus == "" {
		itemStatus = "manquant"
	}

	// 4. Lire les bordereaux existants (prendre BV1 en priorité)
	slips, _ := r.ListPaymentSlipsByDossier(ctx, orgID, dossierID)
	var activeSlip *domain.ANPPPaymentSlip
	for i := range slips {
		if slips[i].SlipType == domain.SlipTypeBV1 {
			activeSlip = &slips[i]
			break
		}
	}
	if activeSlip == nil && len(slips) > 0 {
		activeSlip = &slips[0]
	}

	summary := &domain.Section1A4Summary{
		DossierID:            dossierID,
		DossierCode:          dossierCode,
		ProcedureCode:        procCode,
		IsImported:           isImported,
		Section1A4Status:     itemStatus,
		ActiveSlip:           activeSlip,
		ExpectedFeeAmountDZD: expectedAmt,
		FeeScheduleLabel:     feeLabel,
		LegalReference:       legalRef,
		ComplianceChecklist:  make([]string, 0),
		BlockingIssues:       make([]string, 0),
	}

	if activeSlip == nil {
		summary.BlockingIssues = append(summary.BlockingIssues, "Bordereau de versement BV1 non émis pour ce dossier.")
		return summary, nil
	}

	summary.ComplianceChecklist = append(summary.ComplianceChecklist, fmt.Sprintf("Bordereau %s émis (%s)", activeSlip.SlipNumber, activeSlip.SlipType))

	// Concordance financière
	if activeSlip.AmountDZD >= expectedAmt {
		summary.IsFeeSatisfied = true
		summary.ComplianceChecklist = append(summary.ComplianceChecklist, fmt.Sprintf("Concordance montant vérifiée (%.2f DZD >= %.2f DZD)", activeSlip.AmountDZD, expectedAmt))
	} else {
		summary.BlockingIssues = append(summary.BlockingIssues, fmt.Sprintf("Montant du bordereau (%.2f DZD) inférieur au barème officiel exigible (%.2f DZD)", activeSlip.AmountDZD, expectedAmt))
	}

	// Quittance bancaire
	if activeSlip.BankReceiptRef != "" && activeSlip.ReceiptDocumentID != nil {
		summary.IsReceiptAttached = true
		summary.ComplianceChecklist = append(summary.ComplianceChecklist, fmt.Sprintf("Quittance bancaire rattachée : réf %s (%s)", activeSlip.BankReceiptRef, activeSlip.BankName))
	} else if activeSlip.Status == "paye" || activeSlip.Status == "quittance_recue" {
		summary.IsReceiptAttached = true
		summary.ComplianceChecklist = append(summary.ComplianceChecklist, fmt.Sprintf("Paiement enregistré avec référence %s", activeSlip.PaymentReference))
	} else {
		summary.BlockingIssues = append(summary.BlockingIssues, "Quittance ou avis d'opéré bancaire non rattaché.")
	}

	// Change Banque d'Algérie
	summary.IsFXRequired = isImported && activeSlip.Currency != "DZD" && activeSlip.Currency != ""
	if summary.IsFXRequired {
		if activeSlip.FXRate != nil && *activeSlip.FXRate > 0 && activeSlip.FXAttestationDocID != nil {
			summary.IsFXSatisfied = true
			summary.ComplianceChecklist = append(summary.ComplianceChecklist, fmt.Sprintf("Attestation de change Banque d'Algérie conforme (1 %s = %.4f DZD)", activeSlip.Currency, *activeSlip.FXRate))
		} else {
			summary.BlockingIssues = append(summary.BlockingIssues, fmt.Sprintf("Attestation de taux de change officielle Banque d'Algérie requise pour ce dossier en devises (%s).", activeSlip.Currency))
		}
	} else {
		summary.IsFXSatisfied = true
		if !isImported {
			summary.ComplianceChecklist = append(summary.ComplianceChecklist, "Production locale en DZD : Attestation de change non exigée.")
		}
	}

	summary.OverallCompliant = summary.IsFeeSatisfied && summary.IsReceiptAttached && summary.IsFXSatisfied && (activeSlip.VerificationStatus == "conforme")

	return summary, nil
}

// GetPreSubmissionByDossier retourne l'enregistrement de pré-soumission pour un dossier
func (r *FeeRepository) GetPreSubmissionByDossier(ctx context.Context, orgID, dossierID string) (*domain.ANPPPreSubmission, error) {
	query := `
		SELECT 
			p.id, p.organization_id, p.dossier_id, p.anpp_reference, p.anpp_notification_date,
			p.decision_type, p.fee_amount_dzd, p.slip_reference, p.payment_date,
			p.response_document_id, COALESCE(doc_resp.title_fr, ''),
			p.payment_document_id, COALESCE(doc_pay.title_fr, ''),
			p.verification_status, p.verified_by,
			COALESCE(TRIM(CONCAT(u.first_name, ' ', u.last_name)), ''),
			p.verified_at, COALESCE(p.verification_notes, ''),
			p.created_at, p.updated_at
		FROM anpp_pre_submissions p
		LEFT JOIN ged_documents doc_resp ON doc_resp.id = p.response_document_id
		LEFT JOIN ged_documents doc_pay ON doc_pay.id = p.payment_document_id
		LEFT JOIN users u ON u.id = p.verified_by
		WHERE p.dossier_id = $1 AND p.organization_id = $2
	`
	var p domain.ANPPPreSubmission
	err := r.pool.QueryRow(ctx, query, dossierID, orgID).Scan(
		&p.ID, &p.OrganizationID, &p.DossierID, &p.ANPPReference, &p.ANPPNotificationDate,
		&p.DecisionType, &p.FeeAmountDZD, &p.SlipReference, &p.PaymentDate,
		&p.ResponseDocumentID, &p.ResponseDocumentTitle,
		&p.PaymentDocumentID, &p.PaymentDocumentTitle,
		&p.VerificationStatus, &p.VerifiedBy,
		&p.VerifiedByName,
		&p.VerifiedAt, &p.VerificationNotes,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("lecture pré-soumission échouée: %w", err)
	}
	return &p, nil
}

// VerifyPreSubmission1A5 enregistre et valide la conformité de l'accord et des frais de pré-soumission
func (r *FeeRepository) VerifyPreSubmission1A5(
	ctx context.Context,
	orgID, dossierID, verifierID, verifierName string,
	req domain.VerifyPreSubmission1A5Request,
) (*domain.ANPPPreSubmission, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("début transaction 1.A.5 échouée: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Lire le dossier pour vérifier la procédure
	var procCode string
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(procedure_code, 'NI')
		FROM dossiers
		WHERE id = $1 AND organization_id = $2
	`, dossierID, orgID).Scan(&procCode)
	if err != nil {
		return nil, fmt.Errorf("dossier introuvable: %w", err)
	}

	isPreSubRequired := (procCode == "NI" || procCode == "NF" || procCode == "MI" || procCode == "MF")

	// Parse dates
	notifDate, err := time.Parse("2006-01-02", req.ANPPNotificationDate)
	if err != nil {
		notifDate = time.Now()
	}
	payDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		payDate = time.Now()
	}

	feeAmount := req.FeeAmountDZD
	if feeAmount <= 0 {
		feeAmount = 50000.00
	}

	decType := req.DecisionType
	if decType == "" {
		decType = "favorable"
	}

	// Déterminer la conformité
	verifStatus := "conforme"
	if isPreSubRequired {
		if req.ANPPReference == "" || req.SlipReference == "" || req.ResponseDocumentID == nil || req.PaymentDocumentID == nil || decType != "favorable" {
			verifStatus = "non_conforme"
		}
	}

	now := time.Now().UTC()

	// Insérer ou mettre à jour dans anpp_pre_submissions
	upsertQuery := `
		INSERT INTO anpp_pre_submissions (
			organization_id, dossier_id, anpp_reference, anpp_notification_date,
			decision_type, fee_amount_dzd, slip_reference, payment_date,
			response_document_id, payment_document_id, verification_status,
			verified_by, verified_at, verification_notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $15)
		ON CONFLICT (dossier_id) DO UPDATE SET
			anpp_reference = EXCLUDED.anpp_reference,
			anpp_notification_date = EXCLUDED.anpp_notification_date,
			decision_type = EXCLUDED.decision_type,
			fee_amount_dzd = EXCLUDED.fee_amount_dzd,
			slip_reference = EXCLUDED.slip_reference,
			payment_date = EXCLUDED.payment_date,
			response_document_id = EXCLUDED.response_document_id,
			payment_document_id = EXCLUDED.payment_document_id,
			verification_status = EXCLUDED.verification_status,
			verified_by = EXCLUDED.verified_by,
			verified_at = EXCLUDED.verified_at,
			verification_notes = EXCLUDED.verification_notes,
			updated_at = EXCLUDED.updated_at
	`
	_, err = tx.Exec(ctx, upsertQuery,
		orgID, dossierID, req.ANPPReference, notifDate,
		decType, feeAmount, req.SlipReference, payDate,
		req.ResponseDocumentID, req.PaymentDocumentID, verifStatus,
		verifierID, now, req.VerificationNotes, now,
	)
	if err != nil {
		return nil, fmt.Errorf("enregistrement pré-soumission échoué: %w", err)
	}

	// Si conforme, mettre à jour la checklist 1.A.5
	if verifStatus == "conforme" && req.ResponseDocumentID != nil {
		notes := fmt.Sprintf("Accord de pré-soumission ANPP validé le %s : Réf %s (décision %s). Redevance acquittée (%.2f DZD, réf %s). Visa : %s",
			now.Format("02/01/2006"), req.ANPPReference, decType, feeAmount, req.SlipReference, verifierName)
		updateChecklistQuery := `
			UPDATE dossier_checklist_items
			SET 
				ged_document_id = $1,
				status = 'valide_qualite',
				validation_notes = $2,
				updated_at = $3
			WHERE dossier_id = $4 AND item_code = '1.A.5'
		`
		_, err = tx.Exec(ctx, updateChecklistQuery, req.ResponseDocumentID, notes, now, dossierID)
		if err != nil {
			return nil, fmt.Errorf("mise à jour checklist 1.A.5 échouée: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit 1.A.5 échoué: %w", err)
	}

	return r.GetPreSubmissionByDossier(ctx, orgID, dossierID)
}

// GetSection1A5Summary compile l'état de conformité de l'accord et des frais de pré-soumission
func (r *FeeRepository) GetSection1A5Summary(ctx context.Context, orgID, dossierID string) (*domain.Section1A5Summary, error) {
	// 1. Lire le dossier
	var dossierCode, procCode string
	err := r.pool.QueryRow(ctx, `
		SELECT dossier_code, COALESCE(procedure_code, 'NI')
		FROM dossiers
		WHERE id = $1 AND organization_id = $2
	`, dossierID, orgID).Scan(&dossierCode, &procCode)
	if err != nil {
		return nil, fmt.Errorf("dossier introuvable: %w", err)
	}

	isPreSubRequired := (procCode == "NI" || procCode == "NF" || procCode == "MI" || procCode == "MF")

	// 2. Lire le statut de la section 1.A.5 dans dossier_checklist_items
	var itemStatus string
	_ = r.pool.QueryRow(ctx, `
		SELECT status FROM dossier_checklist_items
		WHERE dossier_id = $1 AND item_code = '1.A.5'
	`, dossierID).Scan(&itemStatus)
	if itemStatus == "" {
		if !isPreSubRequired {
			itemStatus = "non_applicable"
		} else {
			itemStatus = "manquant"
		}
	}

	preSub, _ := r.GetPreSubmissionByDossier(ctx, orgID, dossierID)

	summary := &domain.Section1A5Summary{
		DossierID:               dossierID,
		DossierCode:             dossierCode,
		ProcedureCode:           procCode,
		IsPreSubmissionRequired: isPreSubRequired,
		Section1A5Status:        itemStatus,
		PreSubmission:           preSub,
		ComplianceChecklist:     make([]string, 0),
		BlockingIssues:          make([]string, 0),
	}

	if !isPreSubRequired {
		summary.OverallCompliant = true
		summary.ComplianceChecklist = append(summary.ComplianceChecklist, fmt.Sprintf("Procédure %s : Accord de pré-soumission non applicable (Décret 20-325)", procCode))
		return summary, nil
	}

	if preSub == nil {
		summary.BlockingIssues = append(summary.BlockingIssues, "Accord formel et frais de pré-soumission non consignés.")
		return summary, nil
	}

	if preSub.ANPPReference != "" && preSub.DecisionType == "favorable" {
		summary.ComplianceChecklist = append(summary.ComplianceChecklist, fmt.Sprintf("Accord préalable favorable de l'ANPP : réf %s du %s", preSub.ANPPReference, preSub.ANPPNotificationDate.Format("02/01/2006")))
	} else {
		summary.BlockingIssues = append(summary.BlockingIssues, "Accord formel préalable de l'ANPP non favorable ou référence manquante.")
	}

	if preSub.SlipReference != "" && preSub.PaymentDocumentID != nil {
		summary.ComplianceChecklist = append(summary.ComplianceChecklist, fmt.Sprintf("Quittance redevance de pré-soumission rattachée : réf %s (%.2f DZD)", preSub.SlipReference, preSub.FeeAmountDZD))
	} else {
		summary.BlockingIssues = append(summary.BlockingIssues, "Justificatif de versement de la redevance de pré-soumission non rattaché.")
	}

	if preSub.ResponseDocumentID != nil {
		summary.ComplianceChecklist = append(summary.ComplianceChecklist, "Copie officielle de la réponse de l'ANPP archivée en GED.")
	} else {
		summary.BlockingIssues = append(summary.BlockingIssues, "Scan de la réponse formelle de l'ANPP manquant dans la GED.")
	}

	summary.OverallCompliant = (len(summary.BlockingIssues) == 0 && preSub.VerificationStatus == "conforme")
	return summary, nil
}

// GetCostConsolidationSummary agrège l'ensemble des dépenses réglementaires par produit et par année
func (r *FeeRepository) GetCostConsolidationSummary(ctx context.Context, orgID string) (*domain.CostConsolidationSummary, error) {
	// 1. Lire le nom de l'organisation
	var orgName string
	_ = r.pool.QueryRow(ctx, "SELECT legal_name FROM organizations WHERE id = $1", orgID).Scan(&orgName)

	// 2. Lire l'ensemble des dossiers avec leurs produits
	queryDossiers := `
		SELECT 
			d.id, d.dossier_code, COALESCE(d.procedure_code, 'NI'),
			COALESCE(d.product_type, 'medicament'),
			COALESCE(p_med.id, p_dm.id, d.id),
			COALESCE(p_med.brand_name_fr, p_dm.commercial_name_fr, d.dossier_code),
			EXTRACT(YEAR FROM d.created_at)::int
		FROM dossiers d
		LEFT JOIN medicines p_med ON p_med.id = d.medicine_id
		LEFT JOIN medical_devices p_dm ON p_dm.id = d.medical_device_id
		WHERE d.organization_id = $1
		ORDER BY d.created_at DESC
	`
	rowsD, err := r.pool.Query(ctx, queryDossiers, orgID)
	if err != nil {
		return nil, fmt.Errorf("requête consolidation dossiers échouée: %w", err)
	}
	defer rowsD.Close()

	type dossierMeta struct {
		ID            string
		Code          string
		ProcedureCode string
		ProductType   string
		ProductID     string
		ProductName   string
		Year          int
	}
	dossiers := make([]dossierMeta, 0)
	for rowsD.Next() {
		var dm dossierMeta
		if err := rowsD.Scan(&dm.ID, &dm.Code, &dm.ProcedureCode, &dm.ProductType, &dm.ProductID, &dm.ProductName, &dm.Year); err == nil {
			dossiers = append(dossiers, dm)
		}
	}
	rowsD.Close()

	summary := &domain.CostConsolidationSummary{
		OrganizationID:    orgID,
		OrganizationName:  orgName,
		TotalDossiers:     len(dossiers),
		ProductBreakdown:  make([]domain.ProductCostSummary, 0),
		YearlyBreakdown:   make([]domain.YearlyCostSummary, 0),
		CategoryBreakdown: make([]domain.CategoryCostSummary, 0),
	}

	yearlyMap := make(map[int]*domain.YearlyCostSummary)
	catTotals := map[string]float64{
		"pre_submission":     0,
		"initial_deposit":    0,
		"complementary_lncm": 0,
		"variations":         0,
		"dm":                 0,
	}

	for _, d := range dossiers {
		var preSubAmt float64
		_ = r.pool.QueryRow(ctx, `
			SELECT COALESCE(fee_amount_dzd, 0)
			FROM anpp_pre_submissions
			WHERE dossier_id = $1 AND organization_id = $2
		`, d.ID, orgID).Scan(&preSubAmt)

		slips, _ := r.ListPaymentSlipsByDossier(ctx, orgID, d.ID)

		var bv1Amt, bv2Amt, varAmt, paidAmt, commAmt float64
		for _, s := range slips {
			commAmt += s.AmountDZD
			if s.Status == "paye" || s.Status == "quittance_recue" {
				paidAmt += s.AmountDZD
			}
			if s.SlipType == "BV1" {
				bv1Amt += s.AmountDZD
			} else if s.SlipType == "BV2" {
				bv2Amt += s.AmountDZD
			} else {
				varAmt += s.AmountDZD
			}
		}

		if preSubAmt > 0 {
			commAmt += preSubAmt
			paidAmt += preSubAmt // Pré-soumission considérée acquittée lors de sa validation
		}

		statusFin := "en_attente"
		if paidAmt >= commAmt && commAmt > 0 {
			statusFin = "acquitte"
			summary.FullySettledCount++
		} else if paidAmt > 0 {
			statusFin = "partiel"
			summary.PendingSettledCount++
		} else {
			summary.PendingSettledCount++
		}

		prodSum := domain.ProductCostSummary{
			ProductID:         d.ProductID,
			ProductName:       d.ProductName,
			ProductType:       d.ProductType,
			DossierID:         d.ID,
			DossierCode:       d.Code,
			ProcedureCode:     d.ProcedureCode,
			PreSubmissionDZD:  preSubAmt,
			InitialDepositDZD: bv1Amt,
			ComplementaryDZD:  bv2Amt,
			VariationsDZD:     varAmt,
			TotalCommittedDZD: commAmt,
			TotalPaidDZD:      paidAmt,
			FinancialStatus:   statusFin,
		}
		summary.ProductBreakdown = append(summary.ProductBreakdown, prodSum)

		summary.TotalFeesEmittedDZD += commAmt
		summary.TotalFeesPaidDZD += paidAmt

		// Consolidation annuelle
		yr := d.Year
		if yr == 0 {
			yr = time.Now().Year()
		}
		if _, ok := yearlyMap[yr]; !ok {
			yearlyMap[yr] = &domain.YearlyCostSummary{Year: yr}
		}
		ys := yearlyMap[yr]
		ys.DossierCount++
		ys.TotalEmittedDZD += commAmt
		ys.TotalPaidDZD += paidAmt
		ys.TotalPendingDZD += (commAmt - paidAmt)
		ys.PreSubmissionDZD += preSubAmt
		ys.InitialDepositDZD += bv1Amt
		ys.ComplementaryDZD += bv2Amt

		// Catégories
		catTotals["pre_submission"] += preSubAmt
		catTotals["initial_deposit"] += bv1Amt
		catTotals["complementary_lncm"] += bv2Amt
		catTotals["variations"] += varAmt
		if d.ProductType == "dispositif_medical" {
			catTotals["dm"] += (bv1Amt + bv2Amt)
		}
	}

	summary.TotalFeesPendingDZD = summary.TotalFeesEmittedDZD - summary.TotalFeesPaidDZD
	if summary.TotalFeesPendingDZD < 0 {
		summary.TotalFeesPendingDZD = 0
	}

	// Construire YearlyBreakdown
	for _, ys := range yearlyMap {
		summary.YearlyBreakdown = append(summary.YearlyBreakdown, *ys)
	}

	// Construire CategoryBreakdown
	catLabels := map[string]string{
		"pre_submission":     "Frais de Pré-soumission (Section 1.A.5)",
		"initial_deposit":    "Redevances Dépôt Initial BV1 (Section 1.A.4)",
		"complementary_lncm": "Analyses LNCM & Expertises BV2",
		"variations":         "Redevances de Modifications & Variations",
		"dm":                 "Homologations Dispositifs Médicaux",
	}
	for catKey, label := range catLabels {
		amt := catTotals[catKey]
		pct := 0.0
		if summary.TotalFeesEmittedDZD > 0 {
			pct = (amt / summary.TotalFeesEmittedDZD) * 100.0
		}
		summary.CategoryBreakdown = append(summary.CategoryBreakdown, domain.CategoryCostSummary{
			Category:   catKey,
			LabelFR:    label,
			AmountDZD:  amt,
			Percentage: pct,
		})
	}

	return summary, nil
}
