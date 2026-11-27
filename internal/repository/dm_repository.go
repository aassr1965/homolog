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

// DMRepository gère l'accès aux données pour l'homologation 6 sections, la recevabilité 13 points et les régularisations
type DMRepository struct {
	pool *pgxpool.Pool
}

// NewDMRepository crée une nouvelle instance de DMRepository
func NewDMRepository(pool *pgxpool.Pool) *DMRepository {
	return &DMRepository{pool: pool}
}

// GetOrganizationByID récupère les informations de l'établissement / organisation pour le pré-remplissage ANPP
func (r *DMRepository) GetOrganizationByID(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error) {
	var o domain.Organization
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, legal_name, COALESCE(miph_license_number, ''), COALESCE(technical_director_name, ''),
		       COALESCE(address, ''), COALESCE(phone, ''), COALESCE(email, '')
		FROM organizations WHERE id = $1
	`, orgID).Scan(&o.ID, &o.Slug, &o.LegalName, &o.MIPHLicenseNumber, &o.TechnicalDirectorName, &o.Address, &o.Phone, &o.Email)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// UpdateMedicalDeviceHomologation met à jour les 6 sections techniques d'un DM
func (r *DMRepository) UpdateMedicalDeviceHomologation(ctx context.Context, dm *domain.MedicalDevice) error {
	query := `
		UPDATE medical_devices SET
			composition_details = $1,
			has_latex = $2,
			has_phthalates = $3,
			has_active_pharma = $4,
			has_biological_origin = $5,
			has_cmr_substances = $6,
			risk_management_iso14971_ref = $7,
			risk_management_status = $8,
			biocompatibility_iso10993_ref = $9,
			biocompatibility_status = $10,
			clinical_eval_iso14155_ref = $11,
			clinical_eval_status = $12,
			sterilization_validation_ref = $13,
			packaging_stability_iso11607_ref = $14,
			sample_batch_pf = $15,
			sample_batch_mp = $16,
			sample_units_count = $17,
			intended_use_fr = $18,
			sterilization_method = $19,
			is_sterile = $20,
			is_single_use = $21,
			ce_certificate_number = $22,
			notified_body_name = $23,
			notified_body_number = $24,
			iso_13485_certificate_number = $25,
			clv_fsc_number = $26,
			clv_issuing_authority = $27,
			gmdn_code = $28,
			updated_at = NOW()
		WHERE id = $29 AND organization_id = $30
	`
	tag, err := r.pool.Exec(ctx, query,
		dm.CompositionDetails,
		dm.HasLatex,
		dm.HasPhthalates,
		dm.HasActivePharma,
		dm.HasBiologicalOrigin,
		dm.HasCMRSubstances,
		dm.RiskManagementISO14971Ref,
		dm.RiskManagementStatus,
		dm.BiocompatibilityISO10993Ref,
		dm.BiocompatibilityStatus,
		dm.ClinicalEvalISO14155Ref,
		dm.ClinicalEvalStatus,
		dm.SterilizationValidationRef,
		dm.PackagingStabilityISO11607Ref,
		dm.SampleBatchPF,
		dm.SampleBatchMP,
		dm.SampleUnitsCount,
		dm.IntendedUseFR,
		dm.SterilizationMethod,
		dm.IsSterile,
		dm.IsSingleUse,
		dm.CECertificateNumber,
		dm.NotifiedBodyName,
		dm.NotifiedBodyNumber,
		dm.ISO13485CertificateNumber,
		dm.CLVFSCNumber,
		dm.CLVIssuingAuthority,
		dm.GMDNCode,
		dm.ID,
		dm.OrganizationID,
	)
	if err != nil {
		return fmt.Errorf("erreur mise à jour homologation DM: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("dispositif médical introuvable")
	}
	return nil
}

// GetMedicalDeviceHomologation récupère les informations techniques détaillées d'un DM
func (r *DMRepository) GetMedicalDeviceHomologation(ctx context.Context, orgID, deviceID uuid.UUID) (*domain.MedicalDevice, error) {
	query := `
		SELECT 
			id, organization_id, product_id, commercial_name_fr, COALESCE(commercial_name_ar, ''),
			manufacturer_reference, dm_class, COALESCE(classification_rule, ''), is_sterile, COALESCE(sterilization_method, ''),
			is_single_use, COALESCE(intended_use_fr, ''), COALESCE(intended_use_ar, ''), COALESCE(gmdn_code, ''), COALESCE(origin_type, 'fabrication_locale'),
			COALESCE(legal_manufacturer_name, ''), COALESCE(legal_manufacturer_country, 'DZ'), manufacturer_site_id, establishment_id,
			COALESCE(ce_certificate_number, ''), COALESCE(notified_body_name, ''), COALESCE(notified_body_number, ''), ce_expiry_date,
			COALESCE(iso_13485_certificate_number, ''), COALESCE(clv_fsc_number, ''), COALESCE(clv_issuing_authority, ''), clv_issue_date,
			COALESCE(economic_file_type, 'pcsu'), pcsu_amount, cif_price, COALESCE(cif_currency, 'EUR'),
			COALESCE(homologation_decision_number, ''), homologation_date, homologation_expiry_date, status,
			COALESCE(composition_details, ''), COALESCE(has_latex, false), COALESCE(has_phthalates, false), COALESCE(has_active_pharma, false), COALESCE(has_biological_origin, false), COALESCE(has_cmr_substances, false),
			COALESCE(risk_management_iso14971_ref, ''), COALESCE(risk_management_status, 'conforme'),
			COALESCE(biocompatibility_iso10993_ref, ''), COALESCE(biocompatibility_status, 'conforme'),
			COALESCE(clinical_eval_iso14155_ref, ''), COALESCE(clinical_eval_status, 'conforme'),
			COALESCE(sterilization_validation_ref, ''), COALESCE(packaging_stability_iso11607_ref, ''),
			COALESCE(sample_batch_pf, ''), COALESCE(sample_batch_mp, ''), COALESCE(sample_units_count, 3),
			created_at, updated_at
		FROM medical_devices
		WHERE id = $1 AND organization_id = $2
	`
	row := r.pool.QueryRow(ctx, query, deviceID, orgID)
	var dm domain.MedicalDevice

	err := row.Scan(
		&dm.ID, &dm.OrganizationID, &dm.ProductID, &dm.CommercialNameFR, &dm.CommercialNameAR,
		&dm.ManufacturerReference, &dm.DMClass, &dm.ClassificationRule, &dm.IsSterile, &dm.SterilizationMethod,
		&dm.IsSingleUse, &dm.IntendedUseFR, &dm.IntendedUseAR, &dm.GMDNCode, &dm.OriginType,
		&dm.LegalManufacturerName, &dm.LegalManufacturerCountry, &dm.ManufacturerSiteID, &dm.EstablishmentID,
		&dm.CECertificateNumber, &dm.NotifiedBodyName, &dm.NotifiedBodyNumber, &dm.CEExpiryDate,
		&dm.ISO13485CertificateNumber, &dm.CLVFSCNumber, &dm.CLVIssuingAuthority, &dm.CLVIssueDate,
		&dm.EconomicFileType, &dm.PCSUAmount, &dm.CIFPrice, &dm.CIFCurrency,
		&dm.HomologationDecisionNumber, &dm.HomologationDate, &dm.HomologationExpiryDate, &dm.Status,
		&dm.CompositionDetails, &dm.HasLatex, &dm.HasPhthalates, &dm.HasActivePharma, &dm.HasBiologicalOrigin, &dm.HasCMRSubstances,
		&dm.RiskManagementISO14971Ref, &dm.RiskManagementStatus,
		&dm.BiocompatibilityISO10993Ref, &dm.BiocompatibilityStatus,
		&dm.ClinicalEvalISO14155Ref, &dm.ClinicalEvalStatus,
		&dm.SterilizationValidationRef, &dm.PackagingStabilityISO11607Ref,
		&dm.SampleBatchPF, &dm.SampleBatchMP, &dm.SampleUnitsCount,
		&dm.CreatedAt, &dm.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("dispositif médical introuvable")
		}
		return nil, fmt.Errorf("erreur lecture DM homologation: %w", err)
	}

	return &dm, nil
}

// GetRecevabiliteEvaluation récupère ou génère l'évaluation de recevabilité en 13 points pour un dossier DM
func (r *DMRepository) GetRecevabiliteEvaluation(ctx context.Context, orgID, dossierID uuid.UUID) (*domain.DMRecevabiliteEvaluation, error) {
	query := `
		SELECT 
			e.id, e.organization_id, e.dossier_id, e.medical_device_id,
			e.deposit_order_number, e.deposit_date, e.electronic_copies_count, e.electronic_media_type,
			e.item_1_status, e.item_2_status, e.item_3_status, e.item_4_status,
			e.item_5_status, e.item_6_status, e.item_7_status, e.item_8_status,
			e.item_9_status, e.item_10_status, e.item_11_status, e.item_12_status,
			e.item_13_status, e.conformity_score, e.decision, COALESCE(e.reserves_notes, ''),
			e.anpp_agent_name, e.pdt_user_id, e.pdt_signed_at, COALESCE(e.pdt_signature_seal, ''),
			COALESCE(e.sha256_hash, ''), e.created_at, e.updated_at,
			d.dossier_code,
			COALESCE(u.first_name || ' ' || u.last_name, '')
		FROM dm_recevabilite_evaluations e
		JOIN dossiers d ON d.id = e.dossier_id
		LEFT JOIN users u ON u.id = e.pdt_user_id
		WHERE e.dossier_id = $1 AND e.organization_id = $2
	`
	row := r.pool.QueryRow(ctx, query, dossierID, orgID)
	var eval domain.DMRecevabiliteEvaluation

	err := row.Scan(
		&eval.ID, &eval.OrganizationID, &eval.DossierID, &eval.MedicalDeviceID,
		&eval.DepositOrderNumber, &eval.DepositDate, &eval.ElectronicCopiesCount, &eval.ElectronicMediaType,
		&eval.Item1Status, &eval.Item2Status, &eval.Item3Status, &eval.Item4Status,
		&eval.Item5Status, &eval.Item6Status, &eval.Item7Status, &eval.Item8Status,
		&eval.Item9Status, &eval.Item10Status, &eval.Item11Status, &eval.Item12Status,
		&eval.Item13Status, &eval.ConformityScore, &eval.Decision, &eval.ReservesNotes,
		&eval.ANPPAgentName, &eval.PDTUserID, &eval.PDTSignedAt, &eval.PDTSignatureSeal,
		&eval.SHA256Hash, &eval.CreatedAt, &eval.UpdatedAt,
		&eval.DossierCode,
		&eval.PDTFullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// S'il n'existe pas encore, on recherche le dossier pour l'initialiser
			return r.initDefaultRecevabiliteEvaluation(ctx, orgID, dossierID)
		}
		return nil, fmt.Errorf("erreur récupération évaluation recevabilité DM: %w", err)
	}

	// Récupérer le DM associé
	dm, err := r.GetMedicalDeviceHomologation(ctx, orgID, eval.MedicalDeviceID)
	if err == nil {
		eval.MedicalDevice = dm
	}

	return &eval, nil
}

// initDefaultRecevabiliteEvaluation initialise une évaluation de recevabilité par défaut pour un dossier DM
func (r *DMRepository) initDefaultRecevabiliteEvaluation(ctx context.Context, orgID, dossierID uuid.UUID) (*domain.DMRecevabiliteEvaluation, error) {
	// Récupérer le DM associé au dossier
	var deviceID uuid.UUID
	var dossierCode string
	err := r.pool.QueryRow(ctx, `
		SELECT medical_device_id, dossier_code 
		FROM dossiers 
		WHERE id = $1 AND organization_id = $2
	`, dossierID, orgID).Scan(&deviceID, &dossierCode)
	if err != nil {
		return nil, fmt.Errorf("dossier DM introuvable pour initialisation: %w", err)
	}

	orderNum := fmt.Sprintf("ANPP/DDM/REC/2026/%04d", time.Now().UnixNano()%10000)
	eval := &domain.DMRecevabiliteEvaluation{
		ID:                    uuid.New(),
		OrganizationID:        orgID,
		DossierID:             dossierID,
		MedicalDeviceID:       deviceID,
		DepositOrderNumber:    orderNum,
		DepositDate:           time.Now(),
		ElectronicCopiesCount: 4,
		ElectronicMediaType:   "usb_sealed_sha256",
		Item1Status:           "conforme",
		Item2Status:           "conforme",
		Item3Status:           "conforme",
		Item4Status:           "conforme",
		Item5Status:           "conforme",
		Item6Status:           "conforme",
		Item7Status:           "conforme",
		Item8Status:           "conforme",
		Item9Status:           "conforme",
		Item10Status:          "conforme",
		Item11Status:          "conforme",
		Item12Status:          "conforme",
		Item13Status:          "conforme",
		ConformityScore:       13,
		Decision:              domain.RecevabiliteDecisionComplet,
		ANPPAgentName:         "Dr. Amel Bouzid (Sous-Direction de l'Homologation DDM)",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
		DossierCode:           dossierCode,
	}

	insertQuery := `
		INSERT INTO dm_recevabilite_evaluations (
			id, organization_id, dossier_id, medical_device_id,
			deposit_order_number, deposit_date, electronic_copies_count, electronic_media_type,
			item_1_status, item_2_status, item_3_status, item_4_status,
			item_5_status, item_6_status, item_7_status, item_8_status,
			item_9_status, item_10_status, item_11_status, item_12_status,
			item_13_status, conformity_score, decision, anpp_agent_name,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24,
			$25, $26
		)
	`
	_, err = r.pool.Exec(ctx, insertQuery,
		eval.ID, eval.OrganizationID, eval.DossierID, eval.MedicalDeviceID,
		eval.DepositOrderNumber, eval.DepositDate, eval.ElectronicCopiesCount, eval.ElectronicMediaType,
		eval.Item1Status, eval.Item2Status, eval.Item3Status, eval.Item4Status,
		eval.Item5Status, eval.Item6Status, eval.Item7Status, eval.Item8Status,
		eval.Item9Status, eval.Item10Status, eval.Item11Status, eval.Item12Status,
		eval.Item13Status, eval.ConformityScore, eval.Decision, eval.ANPPAgentName,
		eval.CreatedAt, eval.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("erreur enregistrement évaluation par défaut: %w", err)
	}

	dm, _ := r.GetMedicalDeviceHomologation(ctx, orgID, deviceID)
	eval.MedicalDevice = dm
	return eval, nil
}

// SaveRecevabiliteEvaluation enregistre les modifications apportées à l'évaluation des 13 points
func (r *DMRepository) SaveRecevabiliteEvaluation(ctx context.Context, eval *domain.DMRecevabiliteEvaluation) error {
	eval.CalculateConformity()
	query := `
		UPDATE dm_recevabilite_evaluations SET
			deposit_order_number = $1,
			deposit_date = $2,
			electronic_copies_count = $3,
			electronic_media_type = $4,
			item_1_status = $5,
			item_2_status = $6,
			item_3_status = $7,
			item_4_status = $8,
			item_5_status = $9,
			item_6_status = $10,
			item_7_status = $11,
			item_8_status = $12,
			item_9_status = $13,
			item_10_status = $14,
			item_11_status = $15,
			item_12_status = $16,
			item_13_status = $17,
			conformity_score = $18,
			decision = $19,
			reserves_notes = $20,
			anpp_agent_name = $21,
			updated_at = NOW()
		WHERE id = $22 AND organization_id = $23
	`
	tag, err := r.pool.Exec(ctx, query,
		eval.DepositOrderNumber, eval.DepositDate, eval.ElectronicCopiesCount, eval.ElectronicMediaType,
		eval.Item1Status, eval.Item2Status, eval.Item3Status, eval.Item4Status,
		eval.Item5Status, eval.Item6Status, eval.Item7Status, eval.Item8Status,
		eval.Item9Status, eval.Item10Status, eval.Item11Status, eval.Item12Status,
		eval.Item13Status, eval.ConformityScore, eval.Decision, eval.ReservesNotes,
		eval.ANPPAgentName, eval.ID, eval.OrganizationID,
	)
	if err != nil {
		return fmt.Errorf("erreur sauvegarde recevabilité DM: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("évaluation de recevabilité introuvable")
	}
	return nil
}

// SealRecevabiliteEvaluation appose la signature et le scellement cryptographique 21 CFR Part 11 du Pharmacien DT
func (r *DMRepository) SealRecevabiliteEvaluation(ctx context.Context, orgID, evalID, userID uuid.UUID, seal, hash string) error {
	query := `
		UPDATE dm_recevabilite_evaluations SET
			pdt_user_id = $1,
			pdt_signed_at = NOW(),
			pdt_signature_seal = $2,
			sha256_hash = $3,
			updated_at = NOW()
		WHERE id = $4 AND organization_id = $5
	`
	tag, err := r.pool.Exec(ctx, query, userID, seal, hash, evalID, orgID)
	if err != nil {
		return fmt.Errorf("erreur scellement recevabilité DM: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("évaluation introuvable pour scellement")
	}
	return nil
}

// ListDMRegularizations liste les DM sous mesures transitoires
func (r *DMRepository) ListDMRegularizations(ctx context.Context, orgID uuid.UUID, statusFilter string) ([]domain.DMRegularization, error) {
	query := `
		SELECT 
			r.id, r.organization_id, r.medical_device_id, r.regularization_code,
			COALESCE(r.derogation_number, ''), r.initial_market_date, r.status,
			COALESCE(r.ddm_acknowledgement_number, ''), r.ddm_deposit_date, r.legal_deadline,
			r.quarterly_volume_imported, r.quarterly_volume_distributed,
			r.hospital_distribution_share, r.retail_distribution_share,
			r.last_declaration_date, COALESCE(r.observations, ''), r.created_at, r.updated_at,
			d.commercial_name_fr, COALESCE(d.commercial_name_ar, ''), d.manufacturer_reference,
			d.dm_class, d.is_sterile, COALESCE(d.origin_type, 'fabrication_locale'), COALESCE(d.legal_manufacturer_name, ''),
			COALESCE(d.legal_manufacturer_country, 'DZ')
		FROM dm_regularizations r
		JOIN medical_devices d ON d.id = r.medical_device_id
		WHERE r.organization_id = $1
	`
	args := []interface{}{orgID}
	if statusFilter != "" && statusFilter != "all" {
		query += " AND r.status = $2"
		args = append(args, statusFilter)
	}
	query += " ORDER BY r.legal_deadline ASC, r.created_at DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("erreur requête liste régularisations DM: %w", err)
	}
	defer rows.Close()

	var list []domain.DMRegularization
	for rows.Next() {
		var reg domain.DMRegularization
		var dm domain.MedicalDevice
		err := rows.Scan(
			&reg.ID, &reg.OrganizationID, &reg.MedicalDeviceID, &reg.RegularizationCode,
			&reg.DerogationNumber, &reg.InitialMarketDate, &reg.Status,
			&reg.DDMAcknowledgementNumber, &reg.DDMDepositDate, &reg.LegalDeadline,
			&reg.QuarterlyVolumeImported, &reg.QuarterlyVolumeDistributed,
			&reg.HospitalDistributionShare, &reg.RetailDistributionShare,
			&reg.LastDeclarationDate, &reg.Observations, &reg.CreatedAt, &reg.UpdatedAt,
			&dm.CommercialNameFR, &dm.CommercialNameAR, &dm.ManufacturerReference,
			&dm.DMClass, &dm.IsSterile, &dm.OriginType, &dm.LegalManufacturerName,
			&dm.LegalManufacturerCountry,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan régularisation DM: %w", err)
		}
		dm.ID = reg.MedicalDeviceID
		dm.OrganizationID = reg.OrganizationID
		reg.MedicalDevice = &dm
		reg.CalculateUrgency()
		list = append(list, reg)
	}

	return list, nil
}

// GetDMRegularizationByID récupère une régularisation par son identifiant
func (r *DMRepository) GetDMRegularizationByID(ctx context.Context, orgID, id uuid.UUID) (*domain.DMRegularization, error) {
	query := `
		SELECT 
			r.id, r.organization_id, r.medical_device_id, r.regularization_code,
			COALESCE(r.derogation_number, ''), r.initial_market_date, r.status,
			COALESCE(r.ddm_acknowledgement_number, ''), r.ddm_deposit_date, r.legal_deadline,
			r.quarterly_volume_imported, r.quarterly_volume_distributed,
			r.hospital_distribution_share, r.retail_distribution_share,
			r.last_declaration_date, COALESCE(r.observations, ''), r.created_at, r.updated_at,
			d.commercial_name_fr, COALESCE(d.commercial_name_ar, ''), d.manufacturer_reference,
			d.dm_class, d.is_sterile, COALESCE(d.origin_type, 'fabrication_locale'), COALESCE(d.legal_manufacturer_name, ''),
			COALESCE(d.legal_manufacturer_country, 'DZ')
		FROM dm_regularizations r
		JOIN medical_devices d ON d.id = r.medical_device_id
		WHERE r.id = $1 AND r.organization_id = $2
	`
	row := r.pool.QueryRow(ctx, query, id, orgID)
	var reg domain.DMRegularization
	var dm domain.MedicalDevice
	err := row.Scan(
		&reg.ID, &reg.OrganizationID, &reg.MedicalDeviceID, &reg.RegularizationCode,
		&reg.DerogationNumber, &reg.InitialMarketDate, &reg.Status,
		&reg.DDMAcknowledgementNumber, &reg.DDMDepositDate, &reg.LegalDeadline,
		&reg.QuarterlyVolumeImported, &reg.QuarterlyVolumeDistributed,
		&reg.HospitalDistributionShare, &reg.RetailDistributionShare,
		&reg.LastDeclarationDate, &reg.Observations, &reg.CreatedAt, &reg.UpdatedAt,
		&dm.CommercialNameFR, &dm.CommercialNameAR, &dm.ManufacturerReference,
		&dm.DMClass, &dm.IsSterile, &dm.OriginType, &dm.LegalManufacturerName,
		&dm.LegalManufacturerCountry,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("dossier de régularisation DM introuvable")
		}
		return nil, fmt.Errorf("erreur lecture régularisation DM: %w", err)
	}
	dm.ID = reg.MedicalDeviceID
	dm.OrganizationID = reg.OrganizationID
	reg.MedicalDevice = &dm
	reg.CalculateUrgency()
	return &reg, nil
}

// CreateDMRegularization insère un nouveau suivi de régularisation DM
func (r *DMRepository) CreateDMRegularization(ctx context.Context, reg *domain.DMRegularization) error {
	if reg.ID == uuid.Nil {
		reg.ID = uuid.New()
	}
	reg.CreatedAt = time.Now()
	reg.UpdatedAt = time.Now()
	if reg.InitialMarketDate.IsZero() {
		reg.InitialMarketDate = time.Now().AddDate(-1, 0, 0)
	}
	if reg.LegalDeadline.IsZero() {
		reg.LegalDeadline = time.Now().AddDate(0, 3, 0) // J+90 par défaut
	}
	if reg.LastDeclarationDate.IsZero() {
		reg.LastDeclarationDate = time.Now()
	}

	query := `
		INSERT INTO dm_regularizations (
			id, organization_id, medical_device_id, regularization_code,
			derogation_number, initial_market_date, status,
			ddm_acknowledgement_number, ddm_deposit_date, legal_deadline,
			quarterly_volume_imported, quarterly_volume_distributed,
			hospital_distribution_share, retail_distribution_share,
			last_declaration_date, observations, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18
		)
	`
	_, err := r.pool.Exec(ctx, query,
		reg.ID, reg.OrganizationID, reg.MedicalDeviceID, reg.RegularizationCode,
		reg.DerogationNumber, reg.InitialMarketDate, reg.Status,
		reg.DDMAcknowledgementNumber, reg.DDMDepositDate, reg.LegalDeadline,
		reg.QuarterlyVolumeImported, reg.QuarterlyVolumeDistributed,
		reg.HospitalDistributionShare, reg.RetailDistributionShare,
		reg.LastDeclarationDate, reg.Observations, reg.CreatedAt, reg.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("erreur création dossier régularisation DM: %w", err)
	}
	return nil
}

// UpdateDMRegularization met à jour les données d'un dossier de régularisation
func (r *DMRepository) UpdateDMRegularization(ctx context.Context, reg *domain.DMRegularization) error {
	query := `
		UPDATE dm_regularizations SET
			derogation_number = $1,
			status = $2,
			ddm_acknowledgement_number = $3,
			ddm_deposit_date = $4,
			legal_deadline = $5,
			quarterly_volume_imported = $6,
			quarterly_volume_distributed = $7,
			hospital_distribution_share = $8,
			retail_distribution_share = $9,
			last_declaration_date = $10,
			observations = $11,
			updated_at = NOW()
		WHERE id = $12 AND organization_id = $13
	`
	tag, err := r.pool.Exec(ctx, query,
		reg.DerogationNumber, reg.Status,
		reg.DDMAcknowledgementNumber, reg.DDMDepositDate, reg.LegalDeadline,
		reg.QuarterlyVolumeImported, reg.QuarterlyVolumeDistributed,
		reg.HospitalDistributionShare, reg.RetailDistributionShare,
		reg.LastDeclarationDate, reg.Observations,
		reg.ID, reg.OrganizationID,
	)
	if err != nil {
		return fmt.Errorf("erreur mise à jour régularisation DM: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("dossier de régularisation introuvable")
	}
	return nil
}

// GetDMRegularizationStats compile les statistiques consolidées pour le tableau de bord
func (r *DMRepository) GetDMRegularizationStats(ctx context.Context, orgID uuid.UUID) (*domain.DMRegularizationStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'commercialise_derogatoire' THEN 1 END) as derog,
			COUNT(CASE WHEN status = 'dossier_depose_ddm' THEN 1 END) as depose,
			COUNT(CASE WHEN status = 'en_instruction' THEN 1 END) as instruct,
			COUNT(CASE WHEN status = 'homologue_regularise' THEN 1 END) as regular,
			COUNT(CASE WHEN status = 'retire' THEN 1 END) as retire,
			COUNT(CASE WHEN status != 'homologue_regularise' AND legal_deadline <= CURRENT_DATE + INTERVAL '30 days' THEN 1 END) as crit,
			COUNT(CASE WHEN status != 'homologue_regularise' AND legal_deadline > CURRENT_DATE + INTERVAL '30 days' AND legal_deadline <= CURRENT_DATE + INTERVAL '60 days' THEN 1 END) as urg,
			COUNT(CASE WHEN status != 'homologue_regularise' AND legal_deadline > CURRENT_DATE + INTERVAL '60 days' AND legal_deadline <= CURRENT_DATE + INTERVAL '90 days' THEN 1 END) as alert,
			COALESCE(SUM(quarterly_volume_distributed), 0) as tot_vol
		FROM dm_regularizations
		WHERE organization_id = $1
	`
	row := r.pool.QueryRow(ctx, query, orgID)
	var stats domain.DMRegularizationStats
	err := row.Scan(
		&stats.TotalDevicesCount,
		&stats.DerogatoireCount,
		&stats.DeposeDDMCount,
		&stats.EnInstructionCount,
		&stats.RegulariseCount,
		&stats.RetireCount,
		&stats.CriticalDeadlineCount,
		&stats.UrgentDeadlineCount,
		&stats.AlertDeadlineCount,
		&stats.TotalVolumeDistributed,
	)
	if err != nil {
		return nil, fmt.Errorf("erreur calcul statistiques régularisation DM: %w", err)
	}
	return &stats, nil
}

// GetDMCanevasExportRows génère les lignes tabulaires conformes à CANEVAS-DECLARATION-DM-NON-HOMOLOGUES.xlsx
func (r *DMRepository) GetDMCanevasExportRows(ctx context.Context, orgID uuid.UUID) ([][]string, error) {
	var estName, estAgreement string
	_ = r.pool.QueryRow(ctx, `SELECT name_fr, miph_license_number FROM establishments WHERE organization_id = $1 LIMIT 1`, orgID).Scan(&estName, &estAgreement)
	if estName == "" {
		estName = "ÉTABLISSEMENT PHARMACEUTIQUE SOUMIS"
	}

	regs, err := r.ListDMRegularizations(ctx, orgID, "all")
	if err != nil {
		return nil, err
	}

	headers := []string{
		"N° Ordre",
		"Code Suivi Régularisation",
		"Établissement Déclarant",
		"N° Agrément MIPH",
		"Dénomination Commerciale DM",
		"Référence Fabricant",
		"Classe de Risque",
		"Statut Stérilité",
		"Origine du Dispositif",
		"Fabricant Légal & Pays",
		"N° Dérogation / Notification",
		"Date Mise sur Marché Algérie",
		"Statut Régularisation DDM",
		"N° Accusé Dépôt DDM",
		"Date Dépôt DDM",
		"Échéance Légale Butoir",
		"Jours Restants",
		"Volume Trimestriel Importé",
		"Volume Trimestriel Distribué",
		"Part Hôpitaux (%)",
		"Part Officines (%)",
		"Date Dernière Déclaration",
		"Observations & Traçabilité",
	}

	rows := [][]string{headers}
	for i, reg := range regs {
		dmName := "DM"
		dmRef := ""
		dmClass := ""
		dmSterile := "Non stérile"
		dmOrigin := "Importé"
		legalMfr := ""
		if reg.MedicalDevice != nil {
			dmName = reg.MedicalDevice.CommercialNameFR
			dmRef = reg.MedicalDevice.ManufacturerReference
			dmClass = reg.MedicalDevice.DMClass
			if reg.MedicalDevice.IsSterile {
				dmSterile = "Stérile"
			}
			if reg.MedicalDevice.OriginType == "fabrication_locale" {
				dmOrigin = "Fabrication Locale"
			}
			legalMfr = fmt.Sprintf("%s (%s)", reg.MedicalDevice.LegalManufacturerName, reg.MedicalDevice.LegalManufacturerCountry)
		}

		ddmDateStr := "En attente"
		if reg.DDMDepositDate != nil {
			ddmDateStr = reg.DDMDepositDate.Format("02/01/2006")
		}

		statusLabel := reg.Status
		switch reg.Status {
		case "commercialise_derogatoire":
			statusLabel = "Commercialisé (Mesure Transitoire)"
		case "dossier_depose_ddm":
			statusLabel = "Dossier Déposé au Guichet DDM"
		case "en_instruction":
			statusLabel = "En cours d'Instruction Technique"
		case "homologue_regularise":
			statusLabel = "Homologué & Régularisé (Décision Octroyée)"
		case "retire":
			statusLabel = "Retiré du Marché"
		}

		row := []string{
			fmt.Sprintf("%d", i+1),
			reg.RegularizationCode,
			estName,
			estAgreement,
			dmName,
			dmRef,
			fmt.Sprintf("Classe %s", dmClass),
			dmSterile,
			dmOrigin,
			legalMfr,
			reg.DerogationNumber,
			reg.InitialMarketDate.Format("02/01/2006"),
			statusLabel,
			reg.DDMAcknowledgementNumber,
			ddmDateStr,
			reg.LegalDeadline.Format("02/01/2006"),
			fmt.Sprintf("%d", reg.DaysRemaining),
			fmt.Sprintf("%.0f", reg.QuarterlyVolumeImported),
			fmt.Sprintf("%.0f", reg.QuarterlyVolumeDistributed),
			fmt.Sprintf("%d%%", reg.HospitalDistributionShare),
			fmt.Sprintf("%d%%", reg.RetailDistributionShare),
			reg.LastDeclarationDate.Format("02/01/2006"),
			reg.Observations,
		}
		rows = append(rows, row)
	}

	return rows, nil
}
