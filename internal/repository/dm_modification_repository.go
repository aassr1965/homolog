package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/domain"
)

// DMModificationRepository gère la persistance des modifications post-homologation
type DMModificationRepository struct {
	pool *pgxpool.Pool
}

func NewDMModificationRepository(pool *pgxpool.Pool) *DMModificationRepository {
	return &DMModificationRepository{pool: pool}
}

// CreateDMModification insère une nouvelle demande de modification
func (r *DMModificationRepository) CreateDMModification(ctx context.Context, m *domain.DMModification) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now
	if m.Status == "" {
		m.Status = domain.DMModStatusBrouillon
	}
	if m.FeeAmount <= 0 {
		m.FeeAmount = m.CalculateFee()
	}

	query := `
		INSERT INTO dm_modifications (
			id, organization_id, medical_device_id, dossier_id,
			modification_code, modification_type, category_code, category_label,
			title, description, justification, current_state, proposed_state,
			impact_risk_iso14971, impact_biocompatibility_iso10993, impact_sterilization_iso11135,
			impact_packaging_iso11607, impact_clinical_iso14155, impact_quality_iso13485,
			fee_amount, fee_receipt_ref, fee_payment_date,
			status, avenant_number, avenant_date, decision_notes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11, $12, $13,
			$14, $15, $16,
			$17, $18, $19,
			$20, $21, $22,
			$23, $24, $25, $26,
			$27, $28
		)
	`

	_, err := r.pool.Exec(ctx, query,
		m.ID, m.OrganizationID, m.MedicalDeviceID, m.DossierID,
		m.ModificationCode, m.ModificationType, m.CategoryCode, m.CategoryLabel,
		m.Title, m.Description, m.Justification, m.CurrentState, m.ProposedState,
		m.ImpactRiskISO14971, m.ImpactBiocompatibilityISO10993, m.ImpactSterilizationISO11135,
		m.ImpactPackagingISO11607, m.ImpactClinicalISO14155, m.ImpactQualityISO13485,
		m.FeeAmount, m.FeeReceiptRef, m.FeePaymentDate,
		m.Status, m.AvenantNumber, m.AvenantDate, m.DecisionNotes,
		m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("échec création modification DM: %w", err)
	}
	return nil
}

// GetDMModificationByID récupère une modification avec son dispositif médical rattaché
func (r *DMModificationRepository) GetDMModificationByID(ctx context.Context, orgID, modID uuid.UUID) (*domain.DMModification, error) {
	query := `
		SELECT
			m.id, m.organization_id, m.medical_device_id, m.dossier_id,
			m.modification_code, m.modification_type, m.category_code, m.category_label,
			m.title, m.description, COALESCE(m.justification, ''),
			COALESCE(m.current_state, ''), COALESCE(m.proposed_state, ''),
			COALESCE(m.impact_risk_iso14971, false), COALESCE(m.impact_biocompatibility_iso10993, false),
			COALESCE(m.impact_sterilization_iso11135, false), COALESCE(m.impact_packaging_iso11607, false),
			COALESCE(m.impact_clinical_iso14155, false), COALESCE(m.impact_quality_iso13485, false),
			m.fee_amount, COALESCE(m.fee_receipt_ref, ''), m.fee_payment_date,
			m.status, m.anpp_submission_date, m.anpp_decision_date,
			COALESCE(m.avenant_number, ''), m.avenant_date, COALESCE(m.decision_notes, ''),
			m.pdt_signed_at, m.pdt_signed_by, COALESCE(m.pdt_signature_seal, ''), COALESCE(m.pdt_signature_notes, ''),
			m.created_at, m.updated_at,
			d.commercial_name_fr, COALESCE(d.commercial_name_ar, ''), d.manufacturer_reference, d.dm_class,
			COALESCE(d.homologation_decision_number, ''), d.homologation_date
		FROM dm_modifications m
		JOIN medical_devices d ON m.medical_device_id = d.id
		WHERE m.id = $1 AND m.organization_id = $2
	`

	var mod domain.DMModification
	var dev domain.MedicalDevice

	err := r.pool.QueryRow(ctx, query, modID, orgID).Scan(
		&mod.ID, &mod.OrganizationID, &mod.MedicalDeviceID, &mod.DossierID,
		&mod.ModificationCode, &mod.ModificationType, &mod.CategoryCode, &mod.CategoryLabel,
		&mod.Title, &mod.Description, &mod.Justification,
		&mod.CurrentState, &mod.ProposedState,
		&mod.ImpactRiskISO14971, &mod.ImpactBiocompatibilityISO10993,
		&mod.ImpactSterilizationISO11135, &mod.ImpactPackagingISO11607,
		&mod.ImpactClinicalISO14155, &mod.ImpactQualityISO13485,
		&mod.FeeAmount, &mod.FeeReceiptRef, &mod.FeePaymentDate,
		&mod.Status, &mod.ANPPSubmissionDate, &mod.ANPPDecisionDate,
		&mod.AvenantNumber, &mod.AvenantDate, &mod.DecisionNotes,
		&mod.PDTSignedAt, &mod.PDTSignedBy, &mod.PDTSignatureSeal, &mod.PDTSignatureNotes,
		&mod.CreatedAt, &mod.UpdatedAt,
		&dev.CommercialNameFR, &dev.CommercialNameAR, &dev.ManufacturerReference, &dev.DMClass,
		&dev.HomologationDecisionNumber, &dev.HomologationDate,
	)
	if err != nil {
		return nil, fmt.Errorf("modification DM introuvable: %w", err)
	}

	dev.ID = mod.MedicalDeviceID
	dev.OrganizationID = mod.OrganizationID
	mod.MedicalDevice = &dev

	return &mod, nil
}

// ListDMModifications liste les modifications avec filtres optionnels
func (r *DMModificationRepository) ListDMModifications(ctx context.Context, orgID uuid.UUID, deviceID *uuid.UUID, typeFilter, statusFilter string) ([]domain.DMModification, error) {
	baseQuery := `
		SELECT
			m.id, m.organization_id, m.medical_device_id, m.dossier_id,
			m.modification_code, m.modification_type, m.category_code, m.category_label,
			m.title, m.description, COALESCE(m.justification, ''),
			COALESCE(m.current_state, ''), COALESCE(m.proposed_state, ''),
			COALESCE(m.impact_risk_iso14971, false), COALESCE(m.impact_biocompatibility_iso10993, false),
			COALESCE(m.impact_sterilization_iso11135, false), COALESCE(m.impact_packaging_iso11607, false),
			COALESCE(m.impact_clinical_iso14155, false), COALESCE(m.impact_quality_iso13485, false),
			m.fee_amount, COALESCE(m.fee_receipt_ref, ''), m.fee_payment_date,
			m.status, m.anpp_submission_date, m.anpp_decision_date,
			COALESCE(m.avenant_number, ''), m.avenant_date, COALESCE(m.decision_notes, ''),
			m.pdt_signed_at, m.pdt_signed_by, COALESCE(m.pdt_signature_seal, ''), COALESCE(m.pdt_signature_notes, ''),
			m.created_at, m.updated_at,
			d.commercial_name_fr, d.manufacturer_reference, d.dm_class
		FROM dm_modifications m
		JOIN medical_devices d ON m.medical_device_id = d.id
		WHERE m.organization_id = $1
	`

	args := []interface{}{orgID}
	argIdx := 2

	if deviceID != nil && *deviceID != uuid.Nil {
		baseQuery += fmt.Sprintf(" AND m.medical_device_id = $%d", argIdx)
		args = append(args, *deviceID)
		argIdx++
	}
	if strings.TrimSpace(typeFilter) != "" && strings.ToUpper(typeFilter) != "ALL" {
		baseQuery += fmt.Sprintf(" AND m.modification_type = $%d", argIdx)
		args = append(args, strings.ToUpper(typeFilter))
		argIdx++
	}
	if strings.TrimSpace(statusFilter) != "" && strings.ToUpper(statusFilter) != "ALL" {
		baseQuery += fmt.Sprintf(" AND m.status = $%d", argIdx)
		args = append(args, statusFilter)
		argIdx++
	}

	baseQuery += " ORDER BY m.created_at DESC"

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("échec lecture liste modifications DM: %w", err)
	}
	defer rows.Close()

	var list []domain.DMModification
	for rows.Next() {
		var mod domain.DMModification
		var dev domain.MedicalDevice

		err := rows.Scan(
			&mod.ID, &mod.OrganizationID, &mod.MedicalDeviceID, &mod.DossierID,
			&mod.ModificationCode, &mod.ModificationType, &mod.CategoryCode, &mod.CategoryLabel,
			&mod.Title, &mod.Description, &mod.Justification,
			&mod.CurrentState, &mod.ProposedState,
			&mod.ImpactRiskISO14971, &mod.ImpactBiocompatibilityISO10993,
			&mod.ImpactSterilizationISO11135, &mod.ImpactPackagingISO11607,
			&mod.ImpactClinicalISO14155, &mod.ImpactQualityISO13485,
			&mod.FeeAmount, &mod.FeeReceiptRef, &mod.FeePaymentDate,
			&mod.Status, &mod.ANPPSubmissionDate, &mod.ANPPDecisionDate,
			&mod.AvenantNumber, &mod.AvenantDate, &mod.DecisionNotes,
			&mod.PDTSignedAt, &mod.PDTSignedBy, &mod.PDTSignatureSeal, &mod.PDTSignatureNotes,
			&mod.CreatedAt, &mod.UpdatedAt,
			&dev.CommercialNameFR, &dev.ManufacturerReference, &dev.DMClass,
		)
		if err != nil {
			return nil, fmt.Errorf("échec scan ligne modification DM: %w", err)
		}

		dev.ID = mod.MedicalDeviceID
		dev.OrganizationID = mod.OrganizationID
		mod.MedicalDevice = &dev
		list = append(list, mod)
	}

	return list, nil
}

// UpdateDMModification met à jour les champs descriptifs et d'impact d'une modification
func (r *DMModificationRepository) UpdateDMModification(ctx context.Context, m *domain.DMModification) error {
	m.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE dm_modifications SET
			title = $1,
			description = $2,
			justification = $3,
			current_state = $4,
			proposed_state = $5,
			impact_risk_iso14971 = $6,
			impact_biocompatibility_iso10993 = $7,
			impact_sterilization_iso11135 = $8,
			impact_packaging_iso11607 = $9,
			impact_clinical_iso14155 = $10,
			impact_quality_iso13485 = $11,
			fee_receipt_ref = $12,
			updated_at = $13
		WHERE id = $14 AND organization_id = $15
	`
	_, err := r.pool.Exec(ctx, query,
		m.Title, m.Description, m.Justification, m.CurrentState, m.ProposedState,
		m.ImpactRiskISO14971, m.ImpactBiocompatibilityISO10993, m.ImpactSterilizationISO11135,
		m.ImpactPackagingISO11607, m.ImpactClinicalISO14155, m.ImpactQualityISO13485,
		m.FeeReceiptRef, m.UpdatedAt,
		m.ID, m.OrganizationID,
	)
	if err != nil {
		return fmt.Errorf("échec mise à jour modification DM: %w", err)
	}
	return nil
}

// SealDMModification appose le visa et le sceau cryptographique du Pharmacien DT
func (r *DMModificationRepository) SealDMModification(ctx context.Context, orgID, modID, userID uuid.UUID, seal, notes string) error {
	now := time.Now().UTC()
	query := `
		UPDATE dm_modifications SET
			pdt_signed_at = $1,
			pdt_signed_by = $2,
			pdt_signature_seal = $3,
			pdt_signature_notes = $4,
			status = CASE WHEN status = 'brouillon' THEN 'soumis_ddm' ELSE status END,
			anpp_submission_date = CASE WHEN anpp_submission_date IS NULL THEN CURRENT_DATE ELSE anpp_submission_date END,
			updated_at = $1
		WHERE id = $5 AND organization_id = $6
	`
	res, err := r.pool.Exec(ctx, query, now, userID, seal, notes, modID, orgID)
	if err != nil {
		return fmt.Errorf("échec scellement modification DM: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("modification introuvable ou non autorisée")
	}
	return nil
}

// IssueAvenant enregistre la décision officielle d'homologation de la DDM (Avenant)
func (r *DMModificationRepository) IssueAvenant(ctx context.Context, orgID, modID uuid.UUID, avenantNumber, decisionNotes string) error {
	now := time.Now().UTC()
	query := `
		UPDATE dm_modifications SET
			status = 'approuve_avenant',
			avenant_number = $1,
			avenant_date = CURRENT_DATE,
			anpp_decision_date = CURRENT_DATE,
			decision_notes = $2,
			updated_at = $3
		WHERE id = $4 AND organization_id = $5
	`
	res, err := r.pool.Exec(ctx, query, avenantNumber, decisionNotes, now, modID, orgID)
	if err != nil {
		return fmt.Errorf("échec octroi avenant DDM: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("modification introuvable")
	}
	return nil
}

// GetDMModificationStats calcule les indicateurs consolidés pour le cockpit
func (r *DMModificationRepository) GetDMModificationStats(ctx context.Context, orgID uuid.UUID) (*domain.DMModificationStats, error) {
	query := `
		SELECT
			COUNT(*) AS total_count,
			COALESCE(COUNT(*) FILTER (WHERE modification_type = 'SUBSTANTIELLE'), 0) AS sub_count,
			COALESCE(COUNT(*) FILTER (WHERE modification_type = 'NON_SUBSTANTIELLE'), 0) AS non_sub_count,
			COALESCE(COUNT(*) FILTER (WHERE status IN ('soumis_ddm', 'en_instruction')), 0) AS inst_count,
			COALESCE(COUNT(*) FILTER (WHERE status = 'approuve_avenant'), 0) AS app_count,
			COALESCE(SUM(fee_amount), 0) AS total_fees
		FROM dm_modifications
		WHERE organization_id = $1
	`
	var stats domain.DMModificationStats
	err := r.pool.QueryRow(ctx, query, orgID).Scan(
		&stats.TotalCount,
		&stats.SubstantielleCount,
		&stats.NonSubstantielleCount,
		&stats.InInstructionCount,
		&stats.ApprouveAvenantCount,
		&stats.TotalFeesDZD,
	)
	if err != nil {
		return nil, fmt.Errorf("échec calcul statistiques modifications DM: %w", err)
	}
	return &stats, nil
}
