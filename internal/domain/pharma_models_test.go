package domain_test

import (
	"testing"

	"homolog/internal/domain"
)

func TestEstablishmentValidation(t *testing.T) {
	tests := []struct {
		name    string
		est     domain.Establishment
		wantErr bool
	}{
		{
			name: "Établissement valide",
			est: domain.Establishment{
				NameFR:                "Laboratoires Saidal SPA",
				TechnicalDirectorName: "Dr. Amine Mansouri",
			},
			wantErr: false,
		},
		{
			name: "Raison sociale manquante",
			est: domain.Establishment{
				NameFR:                "",
				TechnicalDirectorName: "Dr. Mansouri",
			},
			wantErr: true,
		},
		{
			name: "Directeur Technique (PDT 1.A.7) manquant",
			est: domain.Establishment{
				NameFR:                "Laboratoire Test",
				TechnicalDirectorName: "   ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.est.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Establishment.Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestManufacturingSiteValidation(t *testing.T) {
	tests := []struct {
		name    string
		site    domain.ManufacturingSite
		wantErr bool
	}{
		{
			name: "Site valide en Algérie (Formulateur PF)",
			site: domain.ManufacturingSite{
				SiteNameFR:  "Usine de Constantine",
				CountryCode: "DZ",
				SiteRole:    domain.SiteRoleFormulateurPF,
			},
			wantErr: false,
		},
		{
			name: "Site étranger valide (Synthèse API Inde)",
			site: domain.ManufacturingSite{
				SiteNameFR:  "Aurobindo Pharma Unit VII",
				CountryCode: "IN",
				SiteRole:    domain.SiteRoleSyntheseSA,
			},
			wantErr: false,
		},
		{
			name: "Nom du site manquant",
			site: domain.ManufacturingSite{
				SiteNameFR:  "",
				CountryCode: "DZ",
				SiteRole:    domain.SiteRoleSyntheseSA,
			},
			wantErr: true,
		},
		{
			name: "Code pays non ISO 2 lettres",
			site: domain.ManufacturingSite{
				SiteNameFR:  "Site Test",
				CountryCode: "DZA", // 3 lettres au lieu de 2
				SiteRole:    domain.SiteRoleSyntheseSA,
			},
			wantErr: true,
		},
		{
			name: "Rôle de site invalide",
			site: domain.ManufacturingSite{
				SiteNameFR:  "Site Test",
				CountryCode: "DZ",
				SiteRole:    domain.SiteRole("role_inconnu"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.site.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ManufacturingSite.Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestMedicineValidation(t *testing.T) {
	fobPrice := 12.50
	tests := []struct {
		name    string
		med     domain.Medicine
		wantErr bool
	}{
		{
			name: "Princeps valide importé",
			med: domain.Medicine{
				OriginType:            domain.OriginPrinceps,
				BrandNameFR:           "Augmentin",
				DCIFR:                 "Amoxicilline + Acide clavulanique",
				DosageFR:              "1g / 125mg",
				PharmaceuticalFormFR:  "Poudre pour suspension buvable",
				AdministrationRouteFR: "Voie orale",
				EconomicFileType:      domain.EconomicFileFOBBenchmark,
				FOBPrice:              &fobPrice,
			},
			wantErr: false,
		},
		{
			name: "Générique local valide (Dossier PCSU 1.H.4)",
			med: domain.Medicine{
				OriginType:            domain.OriginGeneriqueLocal,
				BrandNameFR:           "Amoxi-Saidal",
				DCIFR:                 "Amoxicilline",
				DosageFR:              "500 mg",
				PharmaceuticalFormFR:  "Gélule",
				AdministrationRouteFR: "Voie orale",
				EconomicFileType:      domain.EconomicFilePCSU,
			},
			wantErr: false,
		},
		{
			name: "Générique local invalide (Refus dossier FOB car fabrication locale)",
			med: domain.Medicine{
				OriginType:            domain.OriginGeneriqueLocal,
				BrandNameFR:           "Amoxi-Saidal",
				DCIFR:                 "Amoxicilline",
				DosageFR:              "500 mg",
				PharmaceuticalFormFR:  "Gélule",
				AdministrationRouteFR: "Voie orale",
				EconomicFileType:      domain.EconomicFileFOBBenchmark, // Invalide pour le local !
			},
			wantErr: true,
		},
		{
			name: "DCI manquante",
			med: domain.Medicine{
				OriginType:            domain.OriginPrinceps,
				BrandNameFR:           "Médicament Test",
				DCIFR:                 "",
				DosageFR:              "500 mg",
				PharmaceuticalFormFR:  "Comprimé",
				AdministrationRouteFR: "Voie orale",
			},
			wantErr: true,
		},
		{
			name: "Type d'origine invalide",
			med: domain.Medicine{
				OriginType:            domain.MedicineOriginType("inconnu"),
				BrandNameFR:           "Médicament Test",
				DCIFR:                 "DCI Test",
				DosageFR:              "500 mg",
				PharmaceuticalFormFR:  "Comprimé",
				AdministrationRouteFR: "Voie orale",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.med.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Medicine.Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestMedicalDeviceValidation(t *testing.T) {
	tests := []struct {
		name    string
		dm      domain.MedicalDevice
		wantErr bool
	}{
		{
			name: "Dispositif Médical valide (Classe IIa stérile)",
			dm: domain.MedicalDevice{
				CommercialNameFR:      "Cathéter Intraveineux Stérile 20G",
				ManufacturerReference: "CAT-IV-20G-DZ",
				DMClass:               "IIa",
				IsSterile:             true,
				IsSingleUse:           true,
			},
			wantErr: false,
		},
		{
			name: "Dispositif Médical Diagnostic In Vitro (DMDIV)",
			dm: domain.MedicalDevice{
				CommercialNameFR:      "Kit de Test Rapide Glycémie",
				ManufacturerReference: "GLUC-TEST-50",
				DMClass:               "DMDIV",
			},
			wantErr: false,
		},
		{
			name: "Classe de DM inconnue",
			dm: domain.MedicalDevice{
				CommercialNameFR:      "Instrument chirurgical",
				ManufacturerReference: "INST-001",
				DMClass:               "ClasseIV", // Invalide
			},
			wantErr: true,
		},
		{
			name: "Référence fabricant manquante",
			dm: domain.MedicalDevice{
				CommercialNameFR:      "Seringue 5ml",
				ManufacturerReference: "   ",
				DMClass:               "I",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dm.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MedicalDevice.Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
