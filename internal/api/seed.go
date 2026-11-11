package api

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/auth"
	"homolog/internal/domain"
)

// SeedInitialData vérifie et insère l'organisation et les comptes BPF essentiels au 1er lancement
func SeedInitialData(ctx context.Context, pool *pgxpool.Pool, orgSlug, orgName string) error {
	// 1. Vérifier si l'organisation existe
	var orgID string
	err := pool.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = $1", orgSlug).Scan(&orgID)
	if err != nil {
		// Création de l'organisation
		query := `
			INSERT INTO organizations (slug, legal_name, miph_license_number, technical_director_name)
			VALUES ($1, $2, 'MIPH-DZ-2024-0089', 'Dr. Amine K.')
			RETURNING id
		`
		if err := pool.QueryRow(ctx, query, orgSlug, orgName).Scan(&orgID); err != nil {
			return err
		}
		log.Printf("[Init] Organisation créée : %s (%s)", orgName, orgSlug)
	}

	// 2. Utilisateurs par défaut
	initialUsers := []struct {
		Email     string
		Password  string
		Role      domain.UserRole
		FirstName string
		LastName  string
	}{
		{
			Email:     "dt@homolog.dz",
			Password:  "P@ssw0rdPharma2026!",
			Role:      domain.RoleDirecteurTechnique,
			FirstName: "Amine",
			LastName:  "K.",
		},
		{
			Email:     "admin@homolog.dz",
			Password:  "Admin@Pharma2026!",
			Role:      domain.RoleAdminLaboratoire,
			FirstName: "Karim",
			LastName:  "M.",
		},
		{
			Email:     "reg@homolog.dz",
			Password:  "Reg@Pharma2026!",
			Role:      domain.RoleResponsableAffairesReg,
			FirstName: "Sarah",
			LastName:  "B.",
		},
	}

	for _, u := range initialUsers {
		var exists bool
		_ = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE organization_id = $1 AND email = $2)", orgID, u.Email).Scan(&exists)
		if !exists {
			pwdHash, err := auth.HashPassword(u.Password, nil)
			if err != nil {
				return err
			}

			insertUserQuery := `
				INSERT INTO users (organization_id, email, password_hash, first_name, last_name, role)
				VALUES ($1, $2, $3, $4, $5, $6)
			`
			_, err = pool.Exec(ctx, insertUserQuery, orgID, u.Email, pwdHash, u.FirstName, u.LastName, string(u.Role))
			if err != nil {
				return err
			}
			log.Printf("[Init] Compte créé : %s (%s) - Rôle: %s", u.Email, u.FirstName+" "+u.LastName, u.Role)
		}
	}

	// 3. Référentiel Démo Jalon 1.3 : Établissement, Sites, Médicaments & Dispositifs Médicaux
	seedPharmaDemoData(ctx, pool, orgID)

	return nil
}

func seedPharmaDemoData(ctx context.Context, pool *pgxpool.Pool, orgID string) {
	// Vérifier si un établissement existe déjà
	var count int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM establishments WHERE organization_id = $1", orgID).Scan(&count)
	if count > 0 {
		return
	}

	// 3.1 Création de l'Établissement Principal
	var estID string
	insertEst := `
		INSERT INTO establishments (
			organization_id, name_fr, name_ar, legal_status,
			miph_license_number, miph_license_grant_date, miph_license_expiry_date,
			technical_director_name, technical_director_decision_ref, technical_director_decision_date,
			address_fr, wilaya, phone, email, is_active
		) VALUES (
			$1, 'Laboratoires Pharmaceutiques Algériens (LPA)', 'مخابر الأدوية الجزائرية', 'SPA',
			'MIPH-DZ-2024-0089', '2024-01-15', '2029-01-14',
			'Dr. Amine K.', 'DEC-MIPH-DT-2024/0412', '2024-02-01',
			'Zone Industrielle Oued Smar, Lot N° 14', 'Alger', '+213 23 92 10 00', 'contact@lpa-pharma.dz', true
		) RETURNING id
	`
	err := pool.QueryRow(ctx, insertEst, orgID).Scan(&estID)
	if err != nil {
		log.Printf("[Seed Pharma] Note insertion établissement : %v", err)
		return
	}
	log.Printf("[Seed Pharma] Établissement créé : LPA SPA (%s)", estID)

	// 3.2 Création des Sites de Fabrication Intervenants (1.H.8)
	var siteFormulateurID string
	insertSite1 := `
		INSERT INTO manufacturing_sites (
			organization_id, establishment_id, site_name_fr, site_name_ar, country_code,
			address, site_role, manufacturing_license_number, gmp_certificate_number,
			gmp_grant_date, gmp_expiry_date, gmp_issuing_authority, is_active
		) VALUES (
			$1, $2, 'Usine LPA Oued Smar (Formes Sèches)', 'مصنع واد السمار للأشكال الجافة', 'DZ',
			'Zone Industrielle Oued Smar, Alger', 'formulateur_pf', 'ML-DZ-2023-0145', 'BPF-ANPP-2024-012',
			'2024-03-10', '2027-03-09', 'Agence Nationale des Produits Pharmaceutiques (ANPP)', true
		) RETURNING id
	`
	_ = pool.QueryRow(ctx, insertSite1, orgID, estID).Scan(&siteFormulateurID)

	insertSite2 := `
		INSERT INTO manufacturing_sites (
			organization_id, establishment_id, site_name_fr, site_name_ar, country_code,
			address, site_role, manufacturing_license_number, gmp_certificate_number,
			gmp_grant_date, gmp_expiry_date, gmp_issuing_authority, is_active
		) VALUES (
			$1, $2, 'Site Synthèse Substance Active (API Europe)', 'موقع تركيب المادة الفعالة', 'FR',
			'Zone Chimique Rhône-Alpes, France', 'synthese_sa', 'ML-FR-ANSM-2022-88', 'GMP-EU-ANSM-2023-441',
			'2023-05-18', '2026-05-17', 'Agence Nationale de Sécurité du Médicament (ANSM)', true
		)
	`
	_, _ = pool.Exec(ctx, insertSite2, orgID, estID)

	// 3.3 Création des Médicaments de Démonstration (Princeps, Générique local PCSU, Générique importé)
	insertMed1 := `
		INSERT INTO medicines (
			organization_id, origin_type, brand_name_fr, brand_name_ar,
			dci_fr, dci_ar, dosage_fr, dosage_ar, pharmaceutical_form_fr, pharmaceutical_form_ar,
			administration_route_fr, administration_route_ar, atc_code,
			has_pharmacopoeia_monograph, pharmacopoeia_name, has_cep, cep_number,
			storage_conditions_fr, shelf_life_months, economic_file_type, pcsu_amount,
			registration_decision_number, status
		) VALUES (
			$1, 'generique_local', 'Paracétamol LPA 500 mg', 'باراسيتامول أل بي أ 500 ملغ',
			'Paracétamol', 'باراسيتامول', '500 mg', '500 ملغ', 'Comprimé sécable', 'أقراص قابلة للقسمة',
			'Voie orale', 'عن طريق الفم', 'N02BE01',
			true, 'Ph. Eur. 11.0', true, 'R1-CEP 2021-089-Rev 02',
			'A conserver à une température ne dépassant pas 25°C à l''abri de l''humidité', 36, 'pcsu', 145.50,
			'DE-2024-0982-REG', 'en_preparation'
		) RETURNING id
	`
	var med1ID string
	_ = pool.QueryRow(ctx, insertMed1, orgID).Scan(&med1ID)

	// Lier le site de fabrication intervenant au Paracétamol
	if med1ID != "" && siteFormulateurID != "" {
		_, _ = pool.Exec(ctx, `
			INSERT INTO medicine_manufacturing_sites (medicine_id, site_id, role_in_product)
			VALUES ($1, $2, 'fabricant_pf')
			ON CONFLICT DO NOTHING
		`, med1ID, siteFormulateurID)
	}

	insertMed2 := `
		INSERT INTO medicines (
			organization_id, origin_type, brand_name_fr, brand_name_ar,
			dci_fr, dci_ar, dosage_fr, dosage_ar, pharmaceutical_form_fr, pharmaceutical_form_ar,
			administration_route_fr, administration_route_ar, atc_code,
			has_pharmacopoeia_monograph, pharmacopoeia_name, has_cep, cep_number,
			storage_conditions_fr, shelf_life_months, economic_file_type, pcsu_amount,
			status
		) VALUES (
			$1, 'generique_local', 'Amoxicilline LPA 1 g', 'أموكسيسيلين أل بي أ 1 غ',
			'Amoxicilline trihydratée', 'أموكسيسيلين ثلاثي الهيدرات', '1000 mg', '1000 ملغ', 'Comprimé dispersible', 'أقراص قابلة للتشتت',
			'Voie orale', 'عن طريق الفم', 'J01CA04',
			true, 'Ph. Eur. 11.0', true, 'R0-CEP 2020-142',
			'A conserver à une température ne dépassant pas 25°C', 24, 'pcsu', 320.00,
			'en_preparation'
		)
	`
	_, _ = pool.Exec(ctx, insertMed2, orgID)

	insertMed3 := `
		INSERT INTO medicines (
			organization_id, origin_type, brand_name_fr, brand_name_ar,
			dci_fr, dci_ar, dosage_fr, dosage_ar, pharmaceutical_form_fr, pharmaceutical_form_ar,
			administration_route_fr, administration_route_ar, atc_code,
			has_pharmacopoeia_monograph, pharmacopoeia_name, has_cep,
			storage_conditions_fr, shelf_life_months, economic_file_type, fob_price, fob_currency,
			has_chamber_of_commerce_visa, has_benchmark_dossier, status
		) VALUES (
			$1, 'princeps', 'Herceptin 440 mg', 'هيرسيبتين 440 ملغ',
			'Trastuzumab', 'تراستوزوماب', '440 mg', '440 ملغ', 'Poudre pour solution à diluer pour perfusion', 'مسحوق لمحلول الحقن الوريدي',
			'Voie intraveineuse', 'عن طريق الوريد', 'L01FD01',
			true, 'Ph. Eur.', false,
			'Conserver au réfrigérateur entre 2°C et 8°C. Ne pas congeler.', 48, 'fob_benchmark', 450.00, 'EUR',
			true, true, 'en_preparation'
		)
	`
	_, _ = pool.Exec(ctx, insertMed3, orgID)

	// 3.4 Création des Dispositifs Médicaux (Décret 20-324 & Arrêté 10/05/2021)
	insertDM := `
		INSERT INTO medical_devices (
			organization_id, commercial_name_fr, commercial_name_ar, manufacturer_reference,
			dm_class, classification_rule, is_sterile, is_single_use,
			intended_use_fr, intended_use_ar, homologation_decision_number, status
		) VALUES (
			$1, 'Seringue Stérile 5 ml 3 Pièces avec Aiguille', 'حقنة معقمة 5 مل 3 قطع مع إبرة', 'LPA-DM-5ML-01',
			'IIa', 'Règle 2 (Dispositif médical invasif à usage chirurgical temporaire)', true, true,
			'Injection parentérale et prélèvement de fluides médicaux chez l''adulte et l''enfant.', 'حقن المحاليل وسحب السوائل الطبية',
			'DH-2023-118-DM', 'en_cours'
		)
	`
	_, _ = pool.Exec(ctx, insertDM, orgID)
	log.Println("[Seed Pharma] Données initiales Jalon 1.3 injectées avec succès.")
}
