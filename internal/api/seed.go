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

	// 2. Utilisateurs et Rôles Réglementaires BPF (Jalon 1.2 / CDC Section 7.1)
	initialUsers := []struct {
		Email     string
		Password  string
		Role      domain.UserRole
		FirstName string
		LastName  string
	}{
		// Comptes Démonstration & Matrice des 10 Rôles BPF
		{
			Email:     "admin.lab@pharma.dz",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleAdminLaboratoire,
			FirstName: "Karim",
			LastName:  "Mokrani (Admin Lab)",
		},
		{
			Email:     "pdt@pharma.dz",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleDirecteurTechnique,
			FirstName: "Dr. Yacine",
			LastName:  "Benkhalid (PDT 1.A.7)",
		},
		{
			Email:     "ra.lead@pharma.dz",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleResponsableAffairesReg,
			FirstName: "Sarah",
			LastName:  "Benali (RA Lead)",
		},
		{
			Email:     "ra.assoc@pharma.dz",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleResponsableAffairesReg,
			FirstName: "Sofiane",
			LastName:  "Hamidi (RA Associate)",
		},
		{
			Email:     "qa.lead@pharma.dz",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleContributeurQualiteBPF,
			FirstName: "Nadir",
			LastName:  "Touati (QA Manager)",
		},
		{
			Email:     "qc.lead@pharma.dz",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleContributeurRDProd,
			FirstName: "Leila",
			LastName:  "Mansouri (QC & Stabilités)",
		},
		{
			Email:     "prod.lead@pharma.dz",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleContributeurRDProd,
			FirstName: "Omar",
			LastName:  "Belkacem (Production & Validation)",
		},
		{
			Email:     "consultant.ext@cro.com",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleConsultantExterne,
			FirstName: "David",
			LastName:  "Moreau (Expert CRO Bioéquivalence)",
		},
		{
			Email:     "auditeur.insp@anpp.dz",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleAuditeurInspecteur,
			FirstName: "Inspecteur",
			LastName:  "ANPP (Auditeur Qualité)",
		},
		{
			Email:     "api.holder@zhejiang-chem.com",
			Password:  "Homolog2026!Dev",
			Role:      domain.RolePartenaireFournisseurAPI,
			FirstName: "Wei",
			LastName:  "Zhang (API Holder Zhejiang Chem)",
		},
		{
			Email:     "admin.sys@cloud.dz",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleAdminSystemeLocal,
			FirstName: "Ahmed",
			LastName:  "Ziani (Admin Système Zero-Trust)",
		},
		{
			Email:     "super.admin@homolog.dz",
			Password:  "Homolog2026!Dev",
			Role:      domain.RoleSuperAdminEditeur,
			FirstName: "SuperAdmin",
			LastName:  "Éditeur SaaS",
		},

		// Comptes de compatibilité historique Sanad Pharma
		{
			Email:     "dt@sanadpharma.dz",
			Password:  "P@ssw0rdPharma2026!",
			Role:      domain.RoleDirecteurTechnique,
			FirstName: "Amine",
			LastName:  "K. (PDT)",
		},
		{
			Email:     "admin@sanadpharma.dz",
			Password:  "Admin@Pharma2026!",
			Role:      domain.RoleAdminLaboratoire,
			FirstName: "Karim",
			LastName:  "M.",
		},
		{
			Email:     "reg@sanadpharma.dz",
			Password:  "Reg@Pharma2026!",
			Role:      domain.RoleResponsableAffairesReg,
			FirstName: "Sarah",
			LastName:  "B.",
		},
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
		} else {
			// S'assurer que le rôle correspond au seed
			_, _ = pool.Exec(ctx, "UPDATE users SET role = $1 WHERE organization_id = $2 AND email = $3 AND role != $1", string(u.Role), orgID, u.Email)
		}
	}

	// 3. Référentiel Démo Jalon 1.3 : Établissement, Sites, Médicaments & Dispositifs Médicaux
	seedPharmaDemoData(ctx, pool, orgID)

	return nil
}

func seedPharmaDemoData(ctx context.Context, pool *pgxpool.Pool, orgID string) {
	// ========================================================================
	// 3.1 ÉTABLISSEMENTS PHARMACEUTIQUES AGRÉÉS MIPH (Décret 20-325 & 1.A.7)
	// ========================================================================
	establishments := []struct {
		NameFR         string
		NameAR         string
		LegalStatus    string
		MIPHLicense    string
		GrantDate      string
		ExpiryDate     string
		DTName         string
		DTDecisionRef  string
		DTDecisionDate string
		AddressFR      string
		Wilaya         string
		Phone          string
		Email          string
	}{
		{
			NameFR:         "Laboratoires Pharmaceutiques Algériens (LPA)",
			NameAR:         "مخابر الأدوية الجزائرية",
			LegalStatus:    "SPA",
			MIPHLicense:    "MIPH-DZ-2024-0089",
			GrantDate:      "2024-01-15",
			ExpiryDate:     "2029-01-14",
			DTName:         "Dr. Amine K.",
			DTDecisionRef:  "DEC-MIPH-DT-2024/0412",
			DTDecisionDate: "2024-02-01",
			AddressFR:      "Zone Industrielle Oued Smar, Lot N° 14",
			Wilaya:         "Alger",
			Phone:          "+213 23 92 10 00",
			Email:          "contact@lpa-pharma.dz",
		},
		{
			NameFR:         "Groupe Industriel Saidal (SPA)",
			NameAR:         "مجمع صيدال الصناعي",
			LegalStatus:    "SPA",
			MIPHLicense:    "MIPH-DZ-2022-0015",
			GrantDate:      "2022-03-01",
			ExpiryDate:     "2027-02-28",
			DTName:         "Dr. Nassim B.",
			DTDecisionRef:  "DEC-MIPH-DT-2022/0118",
			DTDecisionDate: "2022-03-15",
			AddressFR:      "Route Nationale N° 5, El Harrach",
			Wilaya:         "Alger",
			Phone:          "+213 21 52 14 00",
			Email:          "contact@saidalgroup.dz",
		},
		{
			NameFR:         "Maghreb Medical Devices (MMD)",
			NameAR:         "مغرب للأجهزة الطبية",
			LegalStatus:    "SARL",
			MIPHLicense:    "MIPH-DZ-2024-0115",
			GrantDate:      "2024-02-10",
			ExpiryDate:     "2029-02-09",
			DTName:         "Dr. Reda B.",
			DTDecisionRef:  "DEC-MIPH-DT-2024/0074",
			DTDecisionDate: "2024-02-25",
			AddressFR:      "Zone Industrielle Hassi Ameur, Lot 84",
			Wilaya:         "Oran",
			Phone:          "+213 41 62 88 12",
			Email:          "contact@mmd-dz.com",
		},
		{
			NameFR:         "Constantine Pharma CDMO",
			NameAR:         "قسنطينة فارما للتصنيع التعاقدي",
			LegalStatus:    "SPA",
			MIPHLicense:    "MIPH-DZ-2023-0042",
			GrantDate:      "2023-06-20",
			ExpiryDate:     "2028-06-19",
			DTName:         "Dr. Karima T.",
			DTDecisionRef:  "DEC-MIPH-DT-2023/0188",
			DTDecisionDate: "2023-07-01",
			AddressFR:      "Zone d'Activité Nouvelle Ville Ali Mendjeli",
			Wilaya:         "Constantine",
			Phone:          "+213 31 77 40 50",
			Email:          "regulatory@constantine-cdmo.dz",
		},
	}

	estMap := make(map[string]string) // name_fr -> id
	for _, e := range establishments {
		var id string
		err := pool.QueryRow(ctx, "SELECT id FROM establishments WHERE organization_id = $1 AND name_fr = $2", orgID, e.NameFR).Scan(&id)
		if err != nil {
			insertQuery := `
				INSERT INTO establishments (
					organization_id, name_fr, name_ar, legal_status,
					miph_license_number, miph_license_grant_date, miph_license_expiry_date,
					technical_director_name, technical_director_decision_ref, technical_director_decision_date,
					address_fr, wilaya, phone, email, is_active
				) VALUES (
					$1, $2, $3, $4,
					$5, $6, $7,
					$8, $9, $10,
					$11, $12, $13, $14, true
				) RETURNING id
			`
			if err := pool.QueryRow(ctx, insertQuery,
				orgID, e.NameFR, e.NameAR, e.LegalStatus,
				e.MIPHLicense, e.GrantDate, e.ExpiryDate,
				e.DTName, e.DTDecisionRef, e.DTDecisionDate,
				e.AddressFR, e.Wilaya, e.Phone, e.Email,
			).Scan(&id); err == nil {
				log.Printf("[Seed Pharma] Établissement inséré : %s", e.NameFR)
			}
		}
		if id != "" {
			estMap[e.NameFR] = id
		}
	}

	lpaEstID := estMap["Laboratoires Pharmaceutiques Algériens (LPA)"]
	mmdEstID := estMap["Maghreb Medical Devices (MMD)"]
	cdmoEstID := estMap["Constantine Pharma CDMO"]

	// ========================================================================
	// 3.2 SITES DE FABRICATION ET DE CONTRÔLE (Module 1.H.8 & 1.M)
	// ========================================================================
	sites := []struct {
		EstablishmentID *string
		SiteNameFR      string
		SiteNameAR      string
		CountryCode     string
		Address         string
		Role            string
		LicenceNumber   string
		GMPCertNumber   string
		GMPGrant        string
		GMPExpiry       string
		GMPAuthority    string
	}{
		{
			EstablishmentID: &lpaEstID,
			SiteNameFR:      "Usine LPA Oued Smar (Formes Sèches)",
			SiteNameAR:      "مصنع واد السمار للأشكال الجافة",
			CountryCode:     "DZ",
			Address:         "Zone Industrielle Oued Smar, Alger",
			Role:            "formulateur_pf",
			LicenceNumber:   "ML-DZ-2023-0145",
			GMPCertNumber:   "BPF-ANPP-2024-012",
			GMPGrant:        "2024-03-10",
			GMPExpiry:       "2027-03-09",
			GMPAuthority:    "Agence Nationale des Produits Pharmaceutiques (ANPP)",
		},
		{
			EstablishmentID: &lpaEstID,
			SiteNameFR:      "Site Synthèse Substance Active (API Europe)",
			SiteNameAR:      "موقع تركيب المادة الفعالة",
			CountryCode:     "FR",
			Address:         "Zone Chimique Rhône-Alpes, France",
			Role:            "synthese_sa",
			LicenceNumber:   "ML-FR-ANSM-2022-88",
			GMPCertNumber:   "GMP-EU-ANSM-2023-441",
			GMPGrant:        "2023-05-18",
			GMPExpiry:       "2026-05-17",
			GMPAuthority:    "Agence Nationale de Sécurité du Médicament (ANSM)",
		},
		{
			EstablishmentID: &cdmoEstID,
			SiteNameFR:      "Centre Analytique & Libération de Lots Constantine",
			SiteNameAR:      "مركز التحاليل المخبرية وتحرير الحصص قسنطينة",
			CountryCode:     "DZ",
			Address:         "Zone d'Activité Ali Mendjeli, Constantine",
			Role:            "controle_analytique",
			LicenceNumber:   "ML-DZ-2022-0089",
			GMPCertNumber:   "BPF-ANPP-2023-045",
			GMPGrant:        "2023-04-12",
			GMPExpiry:       "2026-04-11",
			GMPAuthority:    "Agence Nationale des Produits Pharmaceutiques (ANPP)",
		},
		{
			EstablishmentID: &lpaEstID,
			SiteNameFR:      "Unité de Conditionnement & Sérialisation Blida",
			SiteNameAR:      "وحدة التغليف الثانوي والتشفير بالبليدة",
			CountryCode:     "DZ",
			Address:         "Zone Industrielle Ben Boulaid, Blida",
			Role:            "conditionneur_secondaire",
			LicenceNumber:   "ML-DZ-2024-0201",
			GMPCertNumber:   "BPF-ANPP-2024-033",
			GMPGrant:        "2024-05-15",
			GMPExpiry:       "2027-05-14",
			GMPAuthority:    "Agence Nationale des Produits Pharmaceutiques (ANPP)",
		},
		{
			EstablishmentID: nil,
			SiteNameFR:      "Site Synthèse Substance Active API Hyderabad",
			SiteNameAR:      "موقع تركيب المادة الفعالة حيدر آباد",
			CountryCode:     "IN",
			Address:         "Plot 42 Genome Valley, Shameerpet, Hyderabad, Telangana, Inde",
			Role:            "synthese_sa",
			LicenceNumber:   "ML-IN-FDA-2021-992",
			GMPCertNumber:   "GMP-WHO-CDSCO-2023-112",
			GMPGrant:        "2023-09-01",
			GMPExpiry:       "2026-08-31",
			GMPAuthority:    "Central Drugs Standard Control Organization (CDSCO) / WHO",
		},
		{
			EstablishmentID: &mmdEstID,
			SiteNameFR:      "Usine MMD Salle Blanche & Assemblage DM Hassi Ameur",
			SiteNameAR:      "مصنع الغرف النظيفة لإنتاج المستلزمات الطبية حاسي عامر",
			CountryCode:     "DZ",
			Address:         "Zone Industrielle Hassi Ameur, Oran",
			Role:            "formulateur_pf",
			LicenceNumber:   "ML-DZ-DM-2024-0012",
			GMPCertNumber:   "BPF-ANPP-DM-2024-005",
			GMPGrant:        "2024-02-20",
			GMPExpiry:       "2027-02-19",
			GMPAuthority:    "Agence Nationale des Produits Pharmaceutiques (ANPP) - ISO 13485",
		},
	}

	siteMap := make(map[string]string)
	for _, s := range sites {
		var id string
		err := pool.QueryRow(ctx, "SELECT id FROM manufacturing_sites WHERE organization_id = $1 AND site_name_fr = $2", orgID, s.SiteNameFR).Scan(&id)
		if err != nil {
			var estRef any
			if s.EstablishmentID != nil && *s.EstablishmentID != "" {
				estRef = *s.EstablishmentID
			}
			insertQuery := `
				INSERT INTO manufacturing_sites (
					organization_id, establishment_id, site_name_fr, site_name_ar, country_code,
					address, site_role, manufacturing_license_number, gmp_certificate_number,
					gmp_grant_date, gmp_expiry_date, gmp_issuing_authority, is_active
				) VALUES (
					$1, $2, $3, $4, $5,
					$6, $7, $8, $9,
					$10, $11, $12, true
				) RETURNING id
			`
			if err := pool.QueryRow(ctx, insertQuery,
				orgID, estRef, s.SiteNameFR, s.SiteNameAR, s.CountryCode,
				s.Address, s.Role, s.LicenceNumber, s.GMPCertNumber,
				s.GMPGrant, s.GMPExpiry, s.GMPAuthority,
			).Scan(&id); err == nil {
				log.Printf("[Seed Pharma] Site inséré : %s", s.SiteNameFR)
			}
		}
		if id != "" {
			siteMap[s.SiteNameFR] = id
		}
	}

	// ========================================================================
	// 3.3 MÉDICAMENTS (Guide CTD ANPP Juillet 2025 v01)
	// ========================================================================
	medicines := []struct {
		OriginType       string
		BrandNameFR      string
		BrandNameAR      string
		DCIFR            string
		DCIAR            string
		DosageFR         string
		DosageAR         string
		FormFR           string
		FormAR           string
		RouteFR          string
		RouteAR          string
		ATCCode          string
		HasPhEur         bool
		PhEurName        string
		HasCEP           bool
		CEPNumber        string
		CEPAccess        bool
		APIMFNumber      string
		StorageFR        string
		ShelfLife        int
		EconomicType     string
		PCSU             *float64
		FOBPrice         *float64
		FOBCurrency      string
		HasVisa          bool
		HasBenchmark     bool
		RegDecision      string
		Status           string
		AssociatedSiteFR string
	}{
		{
			OriginType:       "generique_local",
			BrandNameFR:      "Paracétamol LPA 500 mg",
			BrandNameAR:      "باراسيتامول أل بي أ 500 ملغ",
			DCIFR:            "Paracétamol",
			DCIAR:            "باراسيتامول",
			DosageFR:         "500 mg",
			DosageAR:         "500 ملغ",
			FormFR:           "Comprimé sécable",
			FormAR:           "أقراص قابلة للقسمة",
			RouteFR:          "Voie orale",
			RouteAR:          "عن طريق الفم",
			ATCCode:          "N02BE01",
			HasPhEur:         true,
			PhEurName:        "Ph. Eur. 11.0",
			HasCEP:           true,
			CEPNumber:        "R1-CEP 2021-089-Rev 02",
			CEPAccess:        true,
			StorageFR:        "A conserver à une température ne dépassant pas 25°C à l'abri de l'humidité",
			ShelfLife:        36,
			EconomicType:     "pcsu",
			PCSU:             floatPtr(145.50),
			RegDecision:      "DE-2024-0982-REG",
			Status:           "valide",
			AssociatedSiteFR: "Usine LPA Oued Smar (Formes Sèches)",
		},
		{
			OriginType:       "generique_local",
			BrandNameFR:      "Amoxicilline LPA 1 g",
			BrandNameAR:      "أموكسيسيلين أل بي أ 1 غ",
			DCIFR:            "Amoxicilline trihydratée",
			DCIAR:            "أموكسيسيلين ثلاثي الهيدرات",
			DosageFR:         "1000 mg",
			DosageAR:         "1000 ملغ",
			FormFR:           "Comprimé dispersible",
			FormAR:           "أقراص قابلة للتشتت",
			RouteFR:          "Voie orale",
			RouteAR:          "عن طريق الفم",
			ATCCode:          "J01CA04",
			HasPhEur:         true,
			PhEurName:        "Ph. Eur. 11.0",
			HasCEP:           true,
			CEPNumber:        "R0-CEP 2020-142",
			CEPAccess:        true,
			StorageFR:        "A conserver à une température ne dépassant pas 25°C",
			ShelfLife:        24,
			EconomicType:     "pcsu",
			PCSU:             floatPtr(320.00),
			Status:           "valide",
			AssociatedSiteFR: "Usine LPA Oued Smar (Formes Sèches)",
		},
		{
			OriginType:       "princeps",
			BrandNameFR:      "Herceptin 440 mg",
			BrandNameAR:      "هيرسيبتين 440 ملغ",
			DCIFR:            "Trastuzumab",
			DCIAR:            "تراستوزوماب",
			DosageFR:         "440 mg",
			DosageAR:         "440 ملغ",
			FormFR:           "Poudre pour solution à diluer pour perfusion",
			FormAR:           "مسحوق لمحلول الحقن الوريدي",
			RouteFR:          "Voie intraveineuse",
			RouteAR:          "عن طريق الوريد",
			ATCCode:          "L01FD01",
			HasPhEur:         true,
			PhEurName:        "Ph. Eur.",
			HasCEP:           false,
			APIMFNumber:      "DPSA-EDQM-BIO-2021-004",
			StorageFR:        "Conserver au réfrigérateur entre 2°C et 8°C. Ne pas congeler.",
			ShelfLife:        48,
			EconomicType:     "fob_benchmark",
			FOBPrice:         floatPtr(450.00),
			FOBCurrency:      "EUR",
			HasVisa:          true,
			HasBenchmark:     true,
			Status:           "valide",
			AssociatedSiteFR: "Site Synthèse Substance Active (API Europe)",
		},
		{
			OriginType:       "generique_local",
			BrandNameFR:      "Metformine LPA 850 mg",
			BrandNameAR:      "ميتفورمين أل بي أ 850 ملغ",
			DCIFR:            "Metformine chlorhydrate",
			DCIAR:            "كلورهيدرات الميتفورمين",
			DosageFR:         "850 mg",
			DosageAR:         "850 ملغ",
			FormFR:           "Comprimé pelliculé",
			FormAR:           "أقراص مغلفة",
			RouteFR:          "Voie orale",
			RouteAR:          "عن طريق الفم",
			ATCCode:          "A10BA02",
			HasPhEur:         true,
			PhEurName:        "Ph. Eur. 11.0",
			HasCEP:           true,
			CEPNumber:        "R1-CEP 2019-210-Rev 01",
			CEPAccess:        true,
			StorageFR:        "A conserver à une température ne dépassant pas 30°C",
			ShelfLife:        36,
			EconomicType:     "pcsu",
			PCSU:             floatPtr(180.00),
			RegDecision:      "DE-2023-0441-REG",
			Status:           "valide",
			AssociatedSiteFR: "Usine LPA Oued Smar (Formes Sèches)",
		},
		{
			OriginType:       "generique_importe",
			BrandNameFR:      "Lovenox 4000 UI anti-Xa / 0,4 ml",
			BrandNameAR:      "لوفينوكس 4000 وحدة / 0.4 مل",
			DCIFR:            "Énoxaparine sodique",
			DCIAR:            "إينوكسابارين صوديوم",
			DosageFR:         "4000 UI / 0.4 ml",
			DosageAR:         "4000 وحدة / 0.4 مل",
			FormFR:           "Solution injectable en seringue préremplie",
			FormAR:           "محلول للحقن في حقنة معبأة مسبقاً",
			RouteFR:          "Voie sous-cutanée",
			RouteAR:          "تحت الجلد",
			ATCCode:          "B01AB05",
			HasPhEur:         true,
			PhEurName:        "Ph. Eur. 11.0",
			HasCEP:           false,
			CEPAccess:        true,
			APIMFNumber:      "APIMF-EDQM-2022-774",
			StorageFR:        "A conserver à une température ne dépassant pas 25°C. Ne pas congeler.",
			ShelfLife:        36,
			EconomicType:     "fob_benchmark",
			FOBPrice:         floatPtr(18.50),
			FOBCurrency:      "EUR",
			HasVisa:          true,
			HasBenchmark:     true,
			RegDecision:      "DE-2022-1102-REG",
			Status:           "valide",
			AssociatedSiteFR: "Site Synthèse Substance Active (API Europe)",
		},
		{
			OriginType:       "generique_local",
			BrandNameFR:      "Insuline Glargine LPA 100 UI/ml",
			BrandNameAR:      "أنسولين غلارجين أل بي أ 100 وحدة/مل",
			DCIFR:            "Insuline glargine",
			DCIAR:            "أنسولين غلارجين",
			DosageFR:         "100 UI/ml (cartouche 3 ml)",
			DosageAR:         "100 وحدة/مل (خرطوشة 3 مل)",
			FormFR:           "Solution injectable en cartouche pour stylo",
			FormAR:           "محلول للحقن في خرطوشة للقلم",
			RouteFR:          "Voie sous-cutanée",
			RouteAR:          "تحت الجلد",
			ATCCode:          "A10AE04",
			HasPhEur:         true,
			PhEurName:        "Ph. Eur.",
			HasCEP:           false,
			CEPAccess:        true,
			APIMFNumber:      "DPSA-ANPP-BIO-2023/018",
			StorageFR:        "Conserver au réfrigérateur entre 2°C et 8°C. Ne pas congeler.",
			ShelfLife:        30,
			EconomicType:     "pcsu",
			PCSU:             floatPtr(2450.00),
			Status:           "en_preparation",
			AssociatedSiteFR: "Usine LPA Oued Smar (Formes Sèches)",
		},
	}

	for _, m := range medicines {
		var medID string
		err := pool.QueryRow(ctx, "SELECT id FROM medicines WHERE organization_id = $1 AND brand_name_fr = $2", orgID, m.BrandNameFR).Scan(&medID)
		if err != nil {
			insertMed := `
				INSERT INTO medicines (
					organization_id, origin_type, brand_name_fr, brand_name_ar,
					dci_fr, dci_ar, dosage_fr, dosage_ar, pharmaceutical_form_fr, pharmaceutical_form_ar,
					administration_route_fr, administration_route_ar, atc_code,
					has_pharmacopoeia_monograph, pharmacopoeia_name, has_cep, cep_number, cep_access_letter_held,
					apimf_dmf_number, storage_conditions_fr, shelf_life_months,
					economic_file_type, pcsu_amount, fob_price, fob_currency,
					has_chamber_of_commerce_visa, has_benchmark_dossier,
					registration_decision_number, status
				) VALUES (
					$1, $2, $3, $4,
					$5, $6, $7, $8, $9, $10,
					$11, $12, $13,
					$14, $15, $16, $17, $18,
					$19, $20, $21,
					$22, $23, $24, $25,
					$26, $27,
					$28, $29
				) RETURNING id
			`
			if err := pool.QueryRow(ctx, insertMed,
				orgID, m.OriginType, m.BrandNameFR, m.BrandNameAR,
				m.DCIFR, m.DCIAR, m.DosageFR, m.DosageAR, m.FormFR, m.FormAR,
				m.RouteFR, m.RouteAR, m.ATCCode,
				m.HasPhEur, m.PhEurName, m.HasCEP, m.CEPNumber, m.CEPAccess,
				m.APIMFNumber, m.StorageFR, m.ShelfLife,
				m.EconomicType, m.PCSU, m.FOBPrice, m.FOBCurrency,
				m.HasVisa, m.HasBenchmark,
				m.RegDecision, m.Status,
			).Scan(&medID); err == nil {
				log.Printf("[Seed Pharma] Médicament inséré : %s", m.BrandNameFR)
			}
		}

		// Association au site intervenant
		if medID != "" && m.AssociatedSiteFR != "" {
			if sID, ok := siteMap[m.AssociatedSiteFR]; ok && sID != "" {
				_, _ = pool.Exec(ctx, `
					INSERT INTO medicine_manufacturing_sites (medicine_id, site_id, role_in_product)
					VALUES ($1, $2, 'fabricant_pf')
					ON CONFLICT DO NOTHING
				`, medID, sID)
			}
		}
	}

	// ========================================================================
	// 3.4 DISPOSITIFS MÉDICAUX (Décret 20-324 & Arrêté 10/05/2021)
	// ========================================================================
	devices := []struct {
		NameFR        string
		NameAR        string
		Ref           string
		Class         string
		Rule          string
		IsSterile     bool
		IsSingleUse   bool
		IntendedUseFR string
		IntendedUseAR string
		DecisionNum   string
		Status        string
	}{
		{
			NameFR:        "Seringue Stérile 5 ml 3 Pièces avec Aiguille",
			NameAR:        "حقنة معقمة 5 مل 3 قطع مع إبرة",
			Ref:           "LPA-DM-5ML-01",
			Class:         "IIa",
			Rule:          "Règle 2 (Dispositif médical invasif à usage chirurgical temporaire)",
			IsSterile:     true,
			IsSingleUse:   true,
			IntendedUseFR: "Injection parentérale et prélèvement de fluides médicaux chez l'adulte et l'enfant.",
			IntendedUseAR: "حقن المحاليل وسحب السوائل الطبية",
			DecisionNum:   "DH-2023-118-DM",
			Status:        "valide",
		},
		{
			NameFR:        "Abaisse-langue en Bois Stérile Emballage Individuel",
			NameAR:        "خافض لسان خشبي معقم بتغليف فردي",
			Ref:           "MMD-AL-BOIS-01",
			Class:         "I",
			Rule:          "Règle 1 (Dispositif médical non invasif en contact avec muqueuse saine)",
			IsSterile:     true,
			IsSingleUse:   true,
			IntendedUseFR: "Abaissement de la langue pour faciliter l'examen clinique de la cavité buccale et du pharynx.",
			IntendedUseAR: "خفض اللسان لتسهيل الفحص السريري لتجويف الفم والبلعوم",
			DecisionNum:   "DH-2022-045-DM",
			Status:        "valide",
		},
		{
			NameFR:        "Sonde d'Alimentation Entérale Polyuréthane 12 Fr",
			NameAR:        "أنبوب تغذية معوية من البولي يوريثان 12 فرنش مع سلك توجيه",
			Ref:           "MMD-ENT-PUR-12FR",
			Class:         "IIb",
			Rule:          "Règle 5 (Dispositif médical invasif par orifice naturel à moyen et long terme)",
			IsSterile:     true,
			IsSingleUse:   true,
			IntendedUseFR: "Administration entérale de solutions nutritives et médicaments par voie naso-gastrique ou naso-jéjunale.",
			IntendedUseAR: "التغذية المعوية وإعطاء الأدوية عبر الأنف للمعدة أو الصائم",
			DecisionNum:   "DH-2024-219-DM",
			Status:        "en_cours",
		},
		{
			NameFR:        "Stent Coronarien Actif au Sirolimus Cobalt-Chrome (DES)",
			NameAR:        "دعامة تاجية نشطة محررة لدواء سيروليموس",
			Ref:           "MMD-COR-STENT-DES30",
			Class:         "III",
			Rule:          "Règle 8 (Dispositif médical implantable à haut risque cardiologique)",
			IsSterile:     true,
			IsSingleUse:   true,
			IntendedUseFR: "Amélioration du diamètre luminal coronarien en cas de cardiopathie ischémique avec sténose athéromateuse.",
			IntendedUseAR: "تحسين تدفق الدم في الشرايين التاجية في حالات التصلب العصيدي وانسداد الشرايين",
			DecisionNum:   "DH-2024-003-DM",
			Status:        "en_cours",
		},
		{
			NameFR:        "Bandelettes Réactives Glucose Capillaire (Boîte de 50)",
			NameAR:        "شرائط اختبار نسبة السكر في الدم الشعيري (علبة 50)",
			Ref:           "MMD-IVD-GLUCO-50",
			Class:         "DMDIV",
			Rule:          "Annexe Arrêté du 10 mai 2021 (Dispositif de Diagnostic In Vitro d'Autodiagnostic)",
			IsSterile:     false,
			IsSingleUse:   true,
			IntendedUseFR: "Mesure quantitative in vitro de la glycémie dans le sang capillaire frais total chez le diabétique.",
			IntendedUseAR: "القياس الذاتي الكمي لمستوى الجلوكوز في الدم الشعيري الطازج لدى مرضى السكري",
			DecisionNum:   "DH-2023-308-DM",
			Status:        "valide",
		},
	}

	for _, d := range devices {
		var id string
		err := pool.QueryRow(ctx, "SELECT id FROM medical_devices WHERE organization_id = $1 AND manufacturer_reference = $2", orgID, d.Ref).Scan(&id)
		if err != nil {
			insertDM := `
				INSERT INTO medical_devices (
					organization_id, commercial_name_fr, commercial_name_ar, manufacturer_reference,
					dm_class, classification_rule, is_sterile, is_single_use,
					intended_use_fr, intended_use_ar, homologation_decision_number, status
				) VALUES (
					$1, $2, $3, $4,
					$5, $6, $7, $8,
					$9, $10, $11, $12
				) RETURNING id
			`
			if err := pool.QueryRow(ctx, insertDM,
				orgID, d.NameFR, d.NameAR, d.Ref,
				d.Class, d.Rule, d.IsSterile, d.IsSingleUse,
				d.IntendedUseFR, d.IntendedUseAR, d.DecisionNum, d.Status,
			).Scan(&id); err == nil {
				log.Printf("[Seed Pharma] Dispositif Médical inséré : %s (%s)", d.NameFR, d.Class)
			}
		}
	}

	log.Println("[Seed Pharma] Données de référence enrichies injectées avec succès.")
}

func floatPtr(v float64) *float64 {
	return &v
}
