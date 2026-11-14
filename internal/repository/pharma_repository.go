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
	"homolog/internal/auth"
	"homolog/internal/domain"
)

var (
	ErrNotFound = errors.New("ressource non trouvée")
)

// PharmaRepository gère la persistance des établissements, sites, médicaments et DM
type PharmaRepository struct {
	pool *pgxpool.Pool
}

// NewPharmaRepository instancie le repository
func NewPharmaRepository(pool *pgxpool.Pool) *PharmaRepository {
	return &PharmaRepository{pool: pool}
}

// ============================================================================
// ÉTABLISSEMENTS
// ============================================================================

func (r *PharmaRepository) CreateEstablishment(ctx context.Context, e *domain.Establishment) error {
	if err := e.Validate(); err != nil {
		return err
	}
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	now := time.Now().UTC()
	e.CreatedAt = now
	e.UpdatedAt = now

	query := `
		INSERT INTO establishments (
			id, organization_id, name_fr, name_ar, legal_status,
			miph_license_number, miph_license_grant_date, miph_license_expiry_date,
			technical_director_name, technical_director_decision_ref, technical_director_decision_date,
			address_fr, address_ar, wilaya, phone, email, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18, $19
		)
	`
	_, err := r.pool.Exec(ctx, query,
		e.ID, e.OrganizationID, e.NameFR, e.NameAR, e.LegalStatus,
		e.MIPHLicenseNumber, e.MIPHLicenseGrantDate, e.MIPHLicenseExpiryDate,
		e.TechnicalDirectorName, e.TechnicalDirectorDecisionRef, e.TechnicalDirectorDecisionDate,
		e.AddressFR, e.AddressAR, e.Wilaya, e.Phone, e.Email, e.IsActive, e.CreatedAt, e.UpdatedAt,
	)
	return err
}

func (r *PharmaRepository) GetEstablishmentByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Establishment, error) {
	query := `
		SELECT 
			id, organization_id, name_fr,
			COALESCE(name_ar, ''),
			COALESCE(legal_status, ''),
			COALESCE(miph_license_number, ''),
			miph_license_grant_date,
			miph_license_expiry_date,
			technical_director_name,
			COALESCE(technical_director_decision_ref, ''),
			technical_director_decision_date,
			COALESCE(address_fr, ''),
			COALESCE(address_ar, ''),
			COALESCE(wilaya, ''),
			COALESCE(phone, ''),
			COALESCE(email, ''),
			is_active, created_at, updated_at
		FROM establishments
		WHERE organization_id = $1 AND id = $2
	`
	var e domain.Establishment
	err := r.pool.QueryRow(ctx, query, orgID, id).Scan(
		&e.ID, &e.OrganizationID, &e.NameFR, &e.NameAR, &e.LegalStatus,
		&e.MIPHLicenseNumber, &e.MIPHLicenseGrantDate, &e.MIPHLicenseExpiryDate,
		&e.TechnicalDirectorName, &e.TechnicalDirectorDecisionRef, &e.TechnicalDirectorDecisionDate,
		&e.AddressFR, &e.AddressAR, &e.Wilaya, &e.Phone, &e.Email, &e.IsActive, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *PharmaRepository) ListEstablishments(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.Establishment, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM establishments WHERE organization_id = $1`, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			id, organization_id, name_fr,
			COALESCE(name_ar, ''),
			COALESCE(legal_status, ''),
			COALESCE(miph_license_number, ''),
			miph_license_grant_date,
			miph_license_expiry_date,
			technical_director_name,
			COALESCE(technical_director_decision_ref, ''),
			technical_director_decision_date,
			COALESCE(address_fr, ''),
			COALESCE(address_ar, ''),
			COALESCE(wilaya, ''),
			COALESCE(phone, ''),
			COALESCE(email, ''),
			is_active, created_at, updated_at
		FROM establishments
		WHERE organization_id = $1
		ORDER BY name_fr ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.Establishment
	for rows.Next() {
		var e domain.Establishment
		if err := rows.Scan(
			&e.ID, &e.OrganizationID, &e.NameFR, &e.NameAR, &e.LegalStatus,
			&e.MIPHLicenseNumber, &e.MIPHLicenseGrantDate, &e.MIPHLicenseExpiryDate,
			&e.TechnicalDirectorName, &e.TechnicalDirectorDecisionRef, &e.TechnicalDirectorDecisionDate,
			&e.AddressFR, &e.AddressAR, &e.Wilaya, &e.Phone, &e.Email, &e.IsActive, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}
	return list, total, nil
}

func (r *PharmaRepository) UpdateEstablishment(ctx context.Context, e *domain.Establishment) error {
	if err := e.Validate(); err != nil {
		return err
	}
	e.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE establishments SET
			name_fr = $3, name_ar = $4, legal_status = $5,
			miph_license_number = $6, miph_license_grant_date = $7, miph_license_expiry_date = $8,
			technical_director_name = $9, technical_director_decision_ref = $10, technical_director_decision_date = $11,
			address_fr = $12, address_ar = $13, wilaya = $14, phone = $15, email = $16, is_active = $17, updated_at = $18
		WHERE organization_id = $1 AND id = $2
	`
	ct, err := r.pool.Exec(ctx, query,
		e.OrganizationID, e.ID, e.NameFR, e.NameAR, e.LegalStatus,
		e.MIPHLicenseNumber, e.MIPHLicenseGrantDate, e.MIPHLicenseExpiryDate,
		e.TechnicalDirectorName, e.TechnicalDirectorDecisionRef, e.TechnicalDirectorDecisionDate,
		e.AddressFR, e.AddressAR, e.Wilaya, e.Phone, e.Email, e.IsActive, e.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ============================================================================
// SITES DE FABRICATION ET DE CONTRÔLE (1.H.8 & 1.M)
// ============================================================================

func (r *PharmaRepository) CreateManufacturingSite(ctx context.Context, s *domain.ManufacturingSite) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now().UTC()
	s.CreatedAt = now
	s.UpdatedAt = now

	query := `
		INSERT INTO manufacturing_sites (
			id, organization_id, establishment_id, site_name_fr, site_name_ar,
			country_code, address, site_role,
			manufacturing_license_number, gmp_certificate_number, gmp_grant_date, gmp_expiry_date, gmp_issuing_authority,
			is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11, $12, $13,
			$14, $15, $16
		)
	`
	_, err := r.pool.Exec(ctx, query,
		s.ID, s.OrganizationID, s.EstablishmentID, s.SiteNameFR, s.SiteNameAR,
		s.CountryCode, s.Address, s.SiteRole,
		s.ManufacturingLicenseNumber, s.GMPCertificateNumber, s.GMPGrantDate, s.GMPExpiryDate, s.GMPIssuingAuthority,
		s.IsActive, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *PharmaRepository) GetManufacturingSiteByID(ctx context.Context, orgID, id uuid.UUID) (*domain.ManufacturingSite, error) {
	query := `
		SELECT 
			id, organization_id, establishment_id, site_name_fr,
			COALESCE(site_name_ar, ''),
			country_code,
			COALESCE(address, ''),
			site_role,
			COALESCE(manufacturing_license_number, ''),
			COALESCE(gmp_certificate_number, ''),
			gmp_grant_date,
			gmp_expiry_date,
			COALESCE(gmp_issuing_authority, ''),
			is_active, created_at, updated_at
		FROM manufacturing_sites
		WHERE organization_id = $1 AND id = $2
	`
	var s domain.ManufacturingSite
	err := r.pool.QueryRow(ctx, query, orgID, id).Scan(
		&s.ID, &s.OrganizationID, &s.EstablishmentID, &s.SiteNameFR, &s.SiteNameAR,
		&s.CountryCode, &s.Address, &s.SiteRole,
		&s.ManufacturingLicenseNumber, &s.GMPCertificateNumber, &s.GMPGrantDate, &s.GMPExpiryDate, &s.GMPIssuingAuthority,
		&s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *PharmaRepository) UpdateManufacturingSite(ctx context.Context, s *domain.ManufacturingSite) error {
	if err := s.Validate(); err != nil {
		return err
	}
	s.UpdatedAt = time.Now().UTC()
	query := `
		UPDATE manufacturing_sites SET
			establishment_id = $3, site_name_fr = $4, site_name_ar = $5,
			country_code = $6, address = $7, site_role = $8,
			manufacturing_license_number = $9, gmp_certificate_number = $10,
			gmp_grant_date = $11, gmp_expiry_date = $12, gmp_issuing_authority = $13,
			is_active = $14, updated_at = $15
		WHERE organization_id = $1 AND id = $2
	`
	ct, err := r.pool.Exec(ctx, query,
		s.OrganizationID, s.ID, s.EstablishmentID, s.SiteNameFR, s.SiteNameAR,
		s.CountryCode, s.Address, s.SiteRole,
		s.ManufacturingLicenseNumber, s.GMPCertificateNumber,
		s.GMPGrantDate, s.GMPExpiryDate, s.GMPIssuingAuthority,
		s.IsActive, s.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PharmaRepository) ListManufacturingSites(ctx context.Context, orgID uuid.UUID, roleFilter string, limit, offset int) ([]domain.ManufacturingSite, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var countQuery strings.Builder
	var listQuery strings.Builder
	args := []any{orgID}

	countQuery.WriteString(`SELECT COUNT(*) FROM manufacturing_sites WHERE organization_id = $1`)
	listQuery.WriteString(`
		SELECT 
			id, organization_id, establishment_id, site_name_fr,
			COALESCE(site_name_ar, ''),
			country_code,
			COALESCE(address, ''),
			site_role,
			COALESCE(manufacturing_license_number, ''),
			COALESCE(gmp_certificate_number, ''),
			gmp_grant_date,
			gmp_expiry_date,
			COALESCE(gmp_issuing_authority, ''),
			is_active, created_at, updated_at
		FROM manufacturing_sites
		WHERE organization_id = $1
	`)

	if roleFilter != "" {
		args = append(args, roleFilter)
		cond := fmt.Sprintf(` AND site_role = $%d`, len(args))
		countQuery.WriteString(cond)
		listQuery.WriteString(cond)
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery.String(), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	listQuery.WriteString(fmt.Sprintf(` ORDER BY site_name_fr ASC LIMIT $%d OFFSET $%d`, len(args)-1, len(args)))

	rows, err := r.pool.Query(ctx, listQuery.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.ManufacturingSite
	for rows.Next() {
		var s domain.ManufacturingSite
		if err := rows.Scan(
			&s.ID, &s.OrganizationID, &s.EstablishmentID, &s.SiteNameFR, &s.SiteNameAR,
			&s.CountryCode, &s.Address, &s.SiteRole,
			&s.ManufacturingLicenseNumber, &s.GMPCertificateNumber, &s.GMPGrantDate, &s.GMPExpiryDate, &s.GMPIssuingAuthority,
			&s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}
	return list, total, nil
}

// ============================================================================
// MÉDICAMENTS (Guide CTD Juillet 2025)
// ============================================================================

func (r *PharmaRepository) CreateMedicine(ctx context.Context, m *domain.Medicine, siteIDs []uuid.UUID) error {
	if err := m.Validate(); err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	m.CreatedAt = now
	m.UpdatedAt = now

	// 1. Insertion dans la table générique products
	if m.ProductID == uuid.Nil {
		m.ProductID = uuid.New()
	}
	prodQuery := `
		INSERT INTO products (
			id, organization_id, product_type, brand_name_fr, brand_name_ar,
			dci, dosage, pharmaceutical_form, administration_route,
			manufacturer_name, marketing_status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.Exec(ctx, prodQuery,
		m.ProductID, m.OrganizationID, domain.ProductTypeMedicament, m.BrandNameFR, m.BrandNameAR,
		m.DCIFR, m.DosageFR, m.PharmaceuticalFormFR, m.AdministrationRouteFR,
		m.BrandNameFR, m.Status, now, now,
	)
	if err != nil {
		return fmt.Errorf("création produit parent échouée: %w", err)
	}

	// 2. Insertion dans la table détaillée medicines
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	medQuery := `
		INSERT INTO medicines (
			id, organization_id, product_id, origin_type,
			brand_name_fr, brand_name_ar, dci_fr, dci_ar,
			dosage_fr, dosage_ar, pharmaceutical_form_fr, pharmaceutical_form_ar,
			administration_route_fr, administration_route_ar, atc_code,
			has_pharmacopoeia_monograph, pharmacopoeia_name, has_cep, cep_number, cep_access_letter_held,
			apimf_dmf_number, polymorphic_form, particle_size_criticality,
			storage_conditions_fr, storage_conditions_ar, shelf_life_months, in_use_shelf_life, validation_batch_files_held,
			economic_file_type, pcsu_amount, fob_price, fob_currency, has_chamber_of_commerce_visa, has_benchmark_dossier,
			registration_decision_number, registration_date, expiry_date, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23,
			$24, $25, $26, $27, $28,
			$29, $30, $31, $32, $33, $34,
			$35, $36, $37, $38, $39, $40
		)
	`
	_, err = tx.Exec(ctx, medQuery,
		m.ID, m.OrganizationID, m.ProductID, m.OriginType,
		m.BrandNameFR, m.BrandNameAR, m.DCIFR, m.DCIAR,
		m.DosageFR, m.DosageAR, m.PharmaceuticalFormFR, m.PharmaceuticalFormAR,
		m.AdministrationRouteFR, m.AdministrationRouteAR, m.ATCCode,
		m.HasPharmacopoeiaMonograph, m.PharmacopoeiaName, m.HasCEP, m.CEPNumber, m.CEPAccessLetterHeld,
		m.APIMFNumber, m.PolymorphicForm, m.ParticleSizeCriticality,
		m.StorageConditionsFR, m.StorageConditionsAR, m.ShelfLifeMonths, m.InUseShelfLife, m.ValidationBatchFilesHeld,
		m.EconomicFileType, m.PCSUAmount, m.FOBPrice, m.FOBCurrency, m.HasChamberOfCommerceVisa, m.HasBenchmarkDossier,
		m.RegistrationDecisionNumber, m.RegistrationDate, m.ExpiryDate, m.Status, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("création médicament échouée: %w", err)
	}

	// 3. Liaison aux sites intervenants (1.H.8)
	for _, siteID := range siteIDs {
		linkQuery := `
			INSERT INTO medicine_manufacturing_sites (medicine_id, site_id, role_in_product)
			VALUES ($1, $2, 'intervenant')
			ON CONFLICT (medicine_id, site_id, role_in_product) DO NOTHING
		`
		if _, err := tx.Exec(ctx, linkQuery, m.ID, siteID); err != nil {
			return fmt.Errorf("liaison site échouée: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PharmaRepository) GetMedicineByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Medicine, error) {
	query := `
		SELECT 
			id, organization_id, product_id, origin_type,
			brand_name_fr,
			COALESCE(brand_name_ar, ''),
			dci_fr,
			COALESCE(dci_ar, ''),
			dosage_fr,
			COALESCE(dosage_ar, ''),
			pharmaceutical_form_fr,
			COALESCE(pharmaceutical_form_ar, ''),
			administration_route_fr,
			COALESCE(administration_route_ar, ''),
			COALESCE(atc_code, ''),
			has_pharmacopoeia_monograph,
			COALESCE(pharmacopoeia_name, ''),
			has_cep,
			COALESCE(cep_number, ''),
			cep_access_letter_held,
			COALESCE(apimf_dmf_number, ''),
			COALESCE(polymorphic_form, ''),
			particle_size_criticality,
			COALESCE(storage_conditions_fr, ''),
			COALESCE(storage_conditions_ar, ''),
			shelf_life_months,
			COALESCE(in_use_shelf_life, ''),
			validation_batch_files_held,
			economic_file_type,
			pcsu_amount,
			fob_price,
			COALESCE(fob_currency, ''),
			has_chamber_of_commerce_visa,
			has_benchmark_dossier,
			COALESCE(registration_decision_number, ''),
			registration_date,
			expiry_date,
			status, created_at, updated_at
		FROM medicines
		WHERE organization_id = $1 AND id = $2
	`
	var m domain.Medicine
	err := r.pool.QueryRow(ctx, query, orgID, id).Scan(
		&m.ID, &m.OrganizationID, &m.ProductID, &m.OriginType,
		&m.BrandNameFR, &m.BrandNameAR, &m.DCIFR, &m.DCIAR,
		&m.DosageFR, &m.DosageAR, &m.PharmaceuticalFormFR, &m.PharmaceuticalFormAR,
		&m.AdministrationRouteFR, &m.AdministrationRouteAR, &m.ATCCode,
		&m.HasPharmacopoeiaMonograph, &m.PharmacopoeiaName, &m.HasCEP, &m.CEPNumber, &m.CEPAccessLetterHeld,
		&m.APIMFNumber, &m.PolymorphicForm, &m.ParticleSizeCriticality,
		&m.StorageConditionsFR, &m.StorageConditionsAR, &m.ShelfLifeMonths, &m.InUseShelfLife, &m.ValidationBatchFilesHeld,
		&m.EconomicFileType, &m.PCSUAmount, &m.FOBPrice, &m.FOBCurrency, &m.HasChamberOfCommerceVisa, &m.HasBenchmarkDossier,
		&m.RegistrationDecisionNumber, &m.RegistrationDate, &m.ExpiryDate, &m.Status, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Récupération des sites intervenants associés (1.H.8)
	sitesQuery := `
		SELECT 
			s.id, s.organization_id, s.establishment_id, s.site_name_fr,
			COALESCE(s.site_name_ar, ''),
			s.country_code,
			COALESCE(s.address, ''),
			s.site_role,
			COALESCE(s.manufacturing_license_number, ''),
			COALESCE(s.gmp_certificate_number, ''),
			s.gmp_grant_date,
			s.gmp_expiry_date,
			COALESCE(s.gmp_issuing_authority, ''),
			s.is_active, s.created_at, s.updated_at
		FROM manufacturing_sites s
		JOIN medicine_manufacturing_sites mms ON s.id = mms.site_id
		WHERE mms.medicine_id = $1
		ORDER BY s.site_name_fr ASC
	`
	rows, err := r.pool.Query(ctx, sitesQuery, m.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var s domain.ManufacturingSite
			if err := rows.Scan(
				&s.ID, &s.OrganizationID, &s.EstablishmentID, &s.SiteNameFR, &s.SiteNameAR,
				&s.CountryCode, &s.Address, &s.SiteRole,
				&s.ManufacturingLicenseNumber, &s.GMPCertificateNumber, &s.GMPGrantDate, &s.GMPExpiryDate, &s.GMPIssuingAuthority,
				&s.IsActive, &s.CreatedAt, &s.UpdatedAt,
			); err == nil {
				m.AssociatedSites = append(m.AssociatedSites, s)
			}
		}
	}

	return &m, nil
}

// UpdateMedicine met à jour les informations d'un médicament et synchronise le produit parent
func (r *PharmaRepository) UpdateMedicine(ctx context.Context, m *domain.Medicine, siteIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	m.UpdatedAt = now

	// 1. Mise à jour de la table medicines
	query := `
		UPDATE medicines SET
			origin_type = $3,
			brand_name_fr = $4,
			brand_name_ar = $5,
			dci_fr = $6,
			dci_ar = $7,
			dosage_fr = $8,
			dosage_ar = $9,
			pharmaceutical_form_fr = $10,
			pharmaceutical_form_ar = $11,
			administration_route_fr = $12,
			administration_route_ar = $13,
			atc_code = $14,
			has_pharmacopoeia_monograph = $15,
			pharmacopoeia_name = $16,
			has_cep = $17,
			cep_number = $18,
			cep_access_letter_held = $19,
			apimf_dmf_number = $20,
			polymorphic_form = $21,
			particle_size_criticality = $22,
			storage_conditions_fr = $23,
			storage_conditions_ar = $24,
			shelf_life_months = $25,
			in_use_shelf_life = $26,
			validation_batch_files_held = $27,
			economic_file_type = $28,
			pcsu_amount = $29,
			fob_price = $30,
			fob_currency = $31,
			has_chamber_of_commerce_visa = $32,
			has_benchmark_dossier = $33,
			status = $34,
			updated_at = $35
		WHERE organization_id = $1 AND id = $2
	`
	tag, err := tx.Exec(ctx, query,
		m.OrganizationID, m.ID, m.OriginType,
		m.BrandNameFR, m.BrandNameAR, m.DCIFR, m.DCIAR,
		m.DosageFR, m.DosageAR, m.PharmaceuticalFormFR, m.PharmaceuticalFormAR,
		m.AdministrationRouteFR, m.AdministrationRouteAR, m.ATCCode,
		m.HasPharmacopoeiaMonograph, m.PharmacopoeiaName, m.HasCEP, m.CEPNumber, m.CEPAccessLetterHeld,
		m.APIMFNumber, m.PolymorphicForm, m.ParticleSizeCriticality,
		m.StorageConditionsFR, m.StorageConditionsAR, m.ShelfLifeMonths, m.InUseShelfLife, m.ValidationBatchFilesHeld,
		m.EconomicFileType, m.PCSUAmount, m.FOBPrice, m.FOBCurrency, m.HasChamberOfCommerceVisa, m.HasBenchmarkDossier,
		m.Status, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("mise à jour médicament échouée: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("médicament introuvable ou non autorisé")
	}

	// 2. Mise à jour du produit parent
	if m.ProductID != uuid.Nil {
		_, err = tx.Exec(ctx, `
			UPDATE products SET
				commercial_name = $1,
				updated_at = $2
			WHERE id = $3 AND organization_id = $4
		`, m.BrandNameFR, now, m.ProductID, m.OrganizationID)
		if err != nil {
			return fmt.Errorf("mise à jour produit parent échouée: %w", err)
		}
	}

	// 3. Mise à jour des sites associés si une liste est explicitement fournie
	if len(siteIDs) > 0 {
		_, _ = tx.Exec(ctx, "DELETE FROM medicine_manufacturing_sites WHERE medicine_id = $1", m.ID)
		for _, siteID := range siteIDs {
			linkQuery := `
				INSERT INTO medicine_manufacturing_sites (medicine_id, site_id, role_in_product)
				VALUES ($1, $2, 'intervenant')
				ON CONFLICT (medicine_id, site_id, role_in_product) DO NOTHING
			`
			if _, err := tx.Exec(ctx, linkQuery, m.ID, siteID); err != nil {
				return fmt.Errorf("liaison site échouée: %w", err)
			}
		}
	}

	return tx.Commit(ctx)
}

// ArchiveMedicine applique un soft delete réglementaire avec traçabilité 21 CFR Part 11 / BPF
func (r *PharmaRepository) ArchiveMedicine(ctx context.Context, orgID, id uuid.UUID, reason, archivedBy string) error {
	now := time.Now()
	query := `
		UPDATE medicines SET
			status = 'archive',
			updated_at = $3
		WHERE organization_id = $1 AND id = $2
	`
	tag, err := r.pool.Exec(ctx, query, orgID, id, now)
	if err != nil {
		return fmt.Errorf("archivage médicament échoué: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("médicament introuvable ou non autorisé")
	}
	return nil
}

func (r *PharmaRepository) ListMedicines(ctx context.Context, orgID uuid.UUID, originFilter, search string, limit, offset int) ([]domain.Medicine, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var countQuery strings.Builder
	var listQuery strings.Builder
	args := []any{orgID}

	countQuery.WriteString(`SELECT COUNT(*) FROM medicines WHERE organization_id = $1`)
	listQuery.WriteString(`
		SELECT 
			id, organization_id, product_id, origin_type,
			brand_name_fr,
			COALESCE(brand_name_ar, ''),
			dci_fr,
			COALESCE(dci_ar, ''),
			dosage_fr,
			COALESCE(dosage_ar, ''),
			pharmaceutical_form_fr,
			COALESCE(pharmaceutical_form_ar, ''),
			administration_route_fr,
			COALESCE(administration_route_ar, ''),
			COALESCE(atc_code, ''),
			has_pharmacopoeia_monograph,
			COALESCE(pharmacopoeia_name, ''),
			has_cep,
			COALESCE(cep_number, ''),
			cep_access_letter_held,
			COALESCE(apimf_dmf_number, ''),
			COALESCE(polymorphic_form, ''),
			particle_size_criticality,
			COALESCE(storage_conditions_fr, ''),
			COALESCE(storage_conditions_ar, ''),
			shelf_life_months,
			COALESCE(in_use_shelf_life, ''),
			validation_batch_files_held,
			economic_file_type,
			pcsu_amount,
			fob_price,
			COALESCE(fob_currency, ''),
			has_chamber_of_commerce_visa,
			has_benchmark_dossier,
			COALESCE(registration_decision_number, ''),
			registration_date,
			expiry_date,
			status, created_at, updated_at
		FROM medicines
		WHERE organization_id = $1
	`)

	if originFilter != "" {
		args = append(args, originFilter)
		cond := fmt.Sprintf(` AND origin_type = $%d`, len(args))
		countQuery.WriteString(cond)
		listQuery.WriteString(cond)
	}

	if search != "" {
		args = append(args, "%"+search+"%")
		cond := fmt.Sprintf(` AND (brand_name_fr ILIKE $%d OR dci_fr ILIKE $%d)`, len(args), len(args))
		countQuery.WriteString(cond)
		listQuery.WriteString(cond)
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery.String(), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	listQuery.WriteString(fmt.Sprintf(` ORDER BY brand_name_fr ASC LIMIT $%d OFFSET $%d`, len(args)-1, len(args)))

	rows, err := r.pool.Query(ctx, listQuery.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.Medicine
	for rows.Next() {
		var m domain.Medicine
		if err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.ProductID, &m.OriginType,
			&m.BrandNameFR, &m.BrandNameAR, &m.DCIFR, &m.DCIAR,
			&m.DosageFR, &m.DosageAR, &m.PharmaceuticalFormFR, &m.PharmaceuticalFormAR,
			&m.AdministrationRouteFR, &m.AdministrationRouteAR, &m.ATCCode,
			&m.HasPharmacopoeiaMonograph, &m.PharmacopoeiaName, &m.HasCEP, &m.CEPNumber, &m.CEPAccessLetterHeld,
			&m.APIMFNumber, &m.PolymorphicForm, &m.ParticleSizeCriticality,
			&m.StorageConditionsFR, &m.StorageConditionsAR, &m.ShelfLifeMonths, &m.InUseShelfLife, &m.ValidationBatchFilesHeld,
			&m.EconomicFileType, &m.PCSUAmount, &m.FOBPrice, &m.FOBCurrency, &m.HasChamberOfCommerceVisa, &m.HasBenchmarkDossier,
			&m.RegistrationDecisionNumber, &m.RegistrationDate, &m.ExpiryDate, &m.Status, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, m)
	}
	return list, total, nil
}

// ============================================================================
// DISPOSITIFS MÉDICAUX (Décret 20-324 & Arrêté 10/05/2021)
// ============================================================================

func (r *PharmaRepository) CreateMedicalDevice(ctx context.Context, dm *domain.MedicalDevice) error {
	if err := dm.Validate(); err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()
	dm.CreatedAt = now
	dm.UpdatedAt = now

	// 1. Insertion dans products
	if dm.ProductID == uuid.Nil {
		dm.ProductID = uuid.New()
	}
	prodQuery := `
		INSERT INTO products (
			id, organization_id, product_type, brand_name_fr, brand_name_ar,
			dm_class, is_sterile, dm_classification_rule,
			manufacturer_name, marketing_status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = tx.Exec(ctx, prodQuery,
		dm.ProductID, dm.OrganizationID, domain.ProductTypeDispositifMedical, dm.CommercialNameFR, dm.CommercialNameAR,
		dm.DMClass, dm.IsSterile, dm.ClassificationRule,
		dm.ManufacturerReference, dm.Status, now, now,
	)
	if err != nil {
		return fmt.Errorf("création produit DM parent échouée: %w", err)
	}

	// 2. Insertion dans medical_devices
	if dm.ID == uuid.Nil {
		dm.ID = uuid.New()
	}
	dmQuery := `
		INSERT INTO medical_devices (
			id, organization_id, product_id, commercial_name_fr, commercial_name_ar,
			manufacturer_reference, dm_class, classification_rule, is_sterile, is_single_use,
			intended_use_fr, intended_use_ar, homologation_decision_number, homologation_date, homologation_expiry_date,
			status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18
		)
	`
	_, err = tx.Exec(ctx, dmQuery,
		dm.ID, dm.OrganizationID, dm.ProductID, dm.CommercialNameFR, dm.CommercialNameAR,
		dm.ManufacturerReference, dm.DMClass, dm.ClassificationRule, dm.IsSterile, dm.IsSingleUse,
		dm.IntendedUseFR, dm.IntendedUseAR, dm.HomologationDecisionNumber, dm.HomologationDate, dm.HomologationExpiryDate,
		dm.Status, dm.CreatedAt, dm.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("création DM échouée: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PharmaRepository) GetMedicalDeviceByID(ctx context.Context, orgID, id uuid.UUID) (*domain.MedicalDevice, error) {
	query := `
		SELECT 
			id, organization_id, product_id, commercial_name_fr,
			COALESCE(commercial_name_ar, ''),
			manufacturer_reference, dm_class,
			COALESCE(classification_rule, ''),
			is_sterile, is_single_use,
			COALESCE(intended_use_fr, ''),
			COALESCE(intended_use_ar, ''),
			COALESCE(homologation_decision_number, ''),
			homologation_date, homologation_expiry_date,
			status, created_at, updated_at
		FROM medical_devices
		WHERE organization_id = $1 AND id = $2
	`
	var dm domain.MedicalDevice
	err := r.pool.QueryRow(ctx, query, orgID, id).Scan(
		&dm.ID, &dm.OrganizationID, &dm.ProductID, &dm.CommercialNameFR, &dm.CommercialNameAR,
		&dm.ManufacturerReference, &dm.DMClass, &dm.ClassificationRule, &dm.IsSterile, &dm.IsSingleUse,
		&dm.IntendedUseFR, &dm.IntendedUseAR, &dm.HomologationDecisionNumber, &dm.HomologationDate, &dm.HomologationExpiryDate,
		&dm.Status, &dm.CreatedAt, &dm.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &dm, nil
}

func (r *PharmaRepository) ListMedicalDevices(ctx context.Context, orgID uuid.UUID, classFilter, search string, limit, offset int) ([]domain.MedicalDevice, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var countQuery strings.Builder
	var listQuery strings.Builder
	args := []any{orgID}

	countQuery.WriteString(`SELECT COUNT(*) FROM medical_devices WHERE organization_id = $1`)
	listQuery.WriteString(`
		SELECT 
			id, organization_id, product_id, commercial_name_fr,
			COALESCE(commercial_name_ar, ''),
			manufacturer_reference, dm_class,
			COALESCE(classification_rule, ''),
			is_sterile, is_single_use,
			COALESCE(intended_use_fr, ''),
			COALESCE(intended_use_ar, ''),
			COALESCE(homologation_decision_number, ''),
			homologation_date, homologation_expiry_date,
			status, created_at, updated_at
		FROM medical_devices
		WHERE organization_id = $1
	`)

	if classFilter != "" {
		args = append(args, classFilter)
		cond := fmt.Sprintf(` AND dm_class = $%d`, len(args))
		countQuery.WriteString(cond)
		listQuery.WriteString(cond)
	}

	if search != "" {
		args = append(args, "%"+search+"%")
		cond := fmt.Sprintf(` AND (commercial_name_fr ILIKE $%d OR manufacturer_reference ILIKE $%d)`, len(args), len(args))
		countQuery.WriteString(cond)
		listQuery.WriteString(cond)
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery.String(), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	listQuery.WriteString(fmt.Sprintf(` ORDER BY commercial_name_fr ASC LIMIT $%d OFFSET $%d`, len(args)-1, len(args)))

	rows, err := r.pool.Query(ctx, listQuery.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.MedicalDevice
	for rows.Next() {
		var dm domain.MedicalDevice
		if err := rows.Scan(
			&dm.ID, &dm.OrganizationID, &dm.ProductID, &dm.CommercialNameFR, &dm.CommercialNameAR,
			&dm.ManufacturerReference, &dm.DMClass, &dm.ClassificationRule, &dm.IsSterile, &dm.IsSingleUse,
			&dm.IntendedUseFR, &dm.IntendedUseAR, &dm.HomologationDecisionNumber, &dm.HomologationDate, &dm.HomologationExpiryDate,
			&dm.Status, &dm.CreatedAt, &dm.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, dm)
	}
	return list, total, nil
}

// VerifyUserPassword vérifie le mot de passe de l'utilisateur pour la signature 21 CFR Part 11
func (r *PharmaRepository) VerifyUserPassword(ctx context.Context, userID uuid.UUID, password string) (bool, error) {
	var pwdHash string
	err := r.pool.QueryRow(ctx, "SELECT password_hash FROM users WHERE id = $1", userID).Scan(&pwdHash)
	if err != nil {
		return false, err
	}
	return auth.VerifyPassword(password, pwdHash)
}
