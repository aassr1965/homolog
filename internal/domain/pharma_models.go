package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SiteRole définit le rôle d'un site de fabrication selon l'ANPP (1.H.8 & 1.M)
type SiteRole string

const (
	SiteRoleSyntheseSA              SiteRole = "synthese_sa"
	SiteRoleIntermediaire           SiteRole = "intermediaire"
	SiteRoleFormulateurPF           SiteRole = "formulateur_pf"
	SiteRoleConditionneurPrimaire   SiteRole = "conditionneur_primaire"
	SiteRoleConditionneurSecondaire SiteRole = "conditionneur_secondaire"
	SiteRoleControleAnalytique      SiteRole = "controle_analytique"
	SiteRoleStockage                SiteRole = "stockage"
)

// MedicineOriginType catégorise le produit selon la matrice ANPP Juillet 2025
type MedicineOriginType string

const (
	OriginPrinceps         MedicineOriginType = "princeps"
	OriginGeneriqueImporte MedicineOriginType = "generique_importe"
	OriginGeneriqueLocal   MedicineOriginType = "generique_local"
)

// EconomicFileType type de dossier économique (1.H.4 & 1.H.5)
type EconomicFileType string

const (
	EconomicFilePCSU         EconomicFileType = "pcsu"          // Fabrication locale (Prix de Revient / Sortie Usine)
	EconomicFileFOBBenchmark EconomicFileType = "fob_benchmark" // Importation (Prix FOB + Benchmark international)
)

// Establishment représente un établissement pharmaceutique agréé par le MIPH
type Establishment struct {
	ID                            uuid.UUID  `json:"id"`
	OrganizationID                uuid.UUID  `json:"organization_id"`
	NameFR                        string     `json:"name_fr"`
	NameAR                        string     `json:"name_ar,omitempty"`
	LegalStatus                   string     `json:"legal_status,omitempty"` // SARL, SPA, EURL...
	MIPHLicenseNumber             string     `json:"miph_license_number,omitempty"`
	MIPHLicenseGrantDate          *time.Time `json:"miph_license_grant_date,omitempty"`
	MIPHLicenseExpiryDate         *time.Time `json:"miph_license_expiry_date,omitempty"`
	TechnicalDirectorName         string     `json:"technical_director_name"` // PDT (1.A.7)
	TechnicalDirectorDecisionRef  string     `json:"technical_director_decision_ref,omitempty"`
	TechnicalDirectorDecisionDate *time.Time `json:"technical_director_decision_date,omitempty"`
	AddressFR                     string     `json:"address_fr,omitempty"`
	AddressAR                     string     `json:"address_ar,omitempty"`
	Wilaya                        string     `json:"wilaya,omitempty"`
	Phone                         string     `json:"phone,omitempty"`
	Email                         string     `json:"email,omitempty"`
	IsActive                      bool       `json:"is_active"`
	CreatedAt                     time.Time  `json:"created_at"`
	UpdatedAt                     time.Time  `json:"updated_at"`
}

// Validate valide la conformité réglementaire de l'établissement
func (e *Establishment) Validate() error {
	if strings.TrimSpace(e.NameFR) == "" {
		return errors.New("la raison sociale (français) est obligatoire")
	}
	if strings.TrimSpace(e.TechnicalDirectorName) == "" {
		return errors.New("le nom du Pharmacien Directeur Technique (PDT - 1.A.7) est obligatoire")
	}
	return nil
}

// ManufacturingSite représente un site intervenant dans la chaîne de fabrication ou de contrôle (1.H.8)
type ManufacturingSite struct {
	ID                         uuid.UUID  `json:"id"`
	OrganizationID             uuid.UUID  `json:"organization_id"`
	EstablishmentID            *uuid.UUID `json:"establishment_id,omitempty"`
	SiteNameFR                 string     `json:"site_name_fr"`
	SiteNameAR                 string     `json:"site_name_ar,omitempty"`
	CountryCode                string     `json:"country_code"` // ISO-3166-1 alpha-2 (DZ, FR, IN, etc.)
	Address                    string     `json:"address,omitempty"`
	SiteRole                   SiteRole   `json:"site_role"`
	ManufacturingLicenseNumber string     `json:"manufacturing_license_number,omitempty"` // 1.M.S.1, 1.M.P.1
	GMPCertificateNumber       string     `json:"gmp_certificate_number,omitempty"`       // 1.M.S.2, 1.M.P.2
	GMPGrantDate               *time.Time `json:"gmp_grant_date,omitempty"`
	GMPExpiryDate              *time.Time `json:"gmp_expiry_date,omitempty"`
	GMPIssuingAuthority        string     `json:"gmp_issuing_authority,omitempty"`
	IsActive                   bool       `json:"is_active"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`
}

// Validate valide les attributs du site de fabrication
func (s *ManufacturingSite) Validate() error {
	if strings.TrimSpace(s.SiteNameFR) == "" {
		return errors.New("le nom du site de fabrication (français) est obligatoire")
	}
	if strings.TrimSpace(s.CountryCode) == "" || len(s.CountryCode) != 2 {
		return errors.New("le code pays du site doit comporter 2 lettres (norme ISO-3166-1 alpha-2)")
	}
	switch s.SiteRole {
	case SiteRoleSyntheseSA, SiteRoleIntermediaire, SiteRoleFormulateurPF,
		SiteRoleConditionneurPrimaire, SiteRoleConditionneurSecondaire,
		SiteRoleControleAnalytique, SiteRoleStockage:
		// valide
	default:
		return errors.New("rôle de site non reconnu selon la nomenclature ANPP")
	}
	return nil
}

// Medicine représente la fiche exhaustive d'un médicament (Guide CTD Juillet 2025)
type Medicine struct {
	ID                    uuid.UUID          `json:"id"`
	OrganizationID        uuid.UUID          `json:"organization_id"`
	ProductID             uuid.UUID          `json:"product_id"`
	OriginType            MedicineOriginType `json:"origin_type"` // princeps, generique_importe, generique_local
	BrandNameFR           string             `json:"brand_name_fr"`
	BrandNameAR           string             `json:"brand_name_ar,omitempty"`
	DCIFR                 string             `json:"dci_fr"`
	DCIAR                 string             `json:"dci_ar,omitempty"`
	DosageFR              string             `json:"dosage_fr"`
	DosageAR              string             `json:"dosage_ar,omitempty"`
	PharmaceuticalFormFR  string             `json:"pharmaceutical_form_fr"`
	PharmaceuticalFormAR  string             `json:"pharmaceutical_form_ar,omitempty"`
	AdministrationRouteFR string             `json:"administration_route_fr"`
	AdministrationRouteAR string             `json:"administration_route_ar,omitempty"`
	ATCCode               string             `json:"atc_code,omitempty"` // Code OMS

	// Substance Active (3.2.S)
	HasPharmacopoeiaMonograph bool   `json:"has_pharmacopoeia_monograph"`
	PharmacopoeiaName         string `json:"pharmacopoeia_name,omitempty"` // Ph. Eur., USP...
	HasCEP                    bool   `json:"has_cep"`
	CEPNumber                 string `json:"cep_number,omitempty"` // 1.M.S.3 / 3.2.S.3
	CEPAccessLetterHeld       bool   `json:"cep_access_letter_held"`
	APIMFNumber               string `json:"apimf_dmf_number,omitempty"` // DPSA
	PolymorphicForm           string `json:"polymorphic_form,omitempty"` // 1.P.4 / 3.2.S.1.3
	ParticleSizeCriticality   bool   `json:"particle_size_criticality"`  // 1.P.4 / 3.2.S.4

	// Produit Fini (3.2.P)
	StorageConditionsFR      string `json:"storage_conditions_fr,omitempty"`
	StorageConditionsAR      string `json:"storage_conditions_ar,omitempty"`
	ShelfLifeMonths          int    `json:"shelf_life_months"`
	InUseShelfLife           string `json:"in_use_shelf_life,omitempty"`
	ValidationBatchFilesHeld bool   `json:"validation_batch_files_held"` // 3.2.P.3.5 / 1.P.8

	// Données Économiques (1.H.4 / 1.H.5)
	EconomicFileType         EconomicFileType `json:"economic_file_type"`
	PCSUAmount               *float64         `json:"pcsu_amount,omitempty"`
	FOBPrice                 *float64         `json:"fob_price,omitempty"`
	FOBCurrency              string           `json:"fob_currency,omitempty"`
	HasChamberOfCommerceVisa bool             `json:"has_chamber_of_commerce_visa"`
	HasBenchmarkDossier      bool             `json:"has_benchmark_dossier"` // 1.H.5

	// Producteur Étranger & Traçabilité Importation (ANPP 1.H.1 & 1.H.8)
	ForeignMAHName      string             `json:"foreign_mah_name,omitempty"`      // Laboratoire Titulaire étranger (MAH)
	ForeignMAHCountry   string             `json:"foreign_mah_country,omitempty"`   // Pays titulaire ISO-2 (FR, DE, etc.)
	CPPNumber           string             `json:"cpp_number,omitempty"`            // N° Certificat CPP Modèle OMS (1.H.1)
	CPPIssuingAuthority string             `json:"cpp_issuing_authority,omitempty"`// Autorité émettrice étrangère (ANSM, FDA...)
	CPPIssueDate        *time.Time         `json:"cpp_issue_date,omitempty"`       // Date délivrance CPP
	ManufacturerSiteID  *uuid.UUID         `json:"manufacturer_site_id,omitempty"` // Site Producteur / Fabricant PF (1.M.P)
	ManufacturerSite    *ManufacturingSite `json:"manufacturer_site,omitempty"`    // Données jointes du site producteur

	// Références Réglementaires
	RegistrationDecisionNumber string     `json:"registration_decision_number,omitempty"`
	RegistrationDate           *time.Time `json:"registration_date,omitempty"`
	ExpiryDate                 *time.Time `json:"expiry_date,omitempty"` // Échéance quinquennale
	Status                     string     `json:"status"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`

	// Sites associés (Cartographie 1.H.8)
	AssociatedSites []ManufacturingSite `json:"associated_sites,omitempty"`
}

// Validate valide la cohérence pharmaceutique selon le guide CTD ANPP 2025
func (m *Medicine) Validate() error {
	if strings.TrimSpace(m.BrandNameFR) == "" {
		return errors.New("la dénomination commerciale (français) est obligatoire")
	}
	if strings.TrimSpace(m.DCIFR) == "" {
		return errors.New("la Dénomination Commune Internationale (DCI) est obligatoire")
	}
	if strings.TrimSpace(m.DosageFR) == "" {
		return errors.New("le dosage du médicament est obligatoire")
	}
	if strings.TrimSpace(m.PharmaceuticalFormFR) == "" {
		return errors.New("la forme pharmaceutique est obligatoire")
	}
	if strings.TrimSpace(m.AdministrationRouteFR) == "" {
		return errors.New("la voie d'administration est obligatoire")
	}

	switch m.OriginType {
	case OriginPrinceps, OriginGeneriqueImporte, OriginGeneriqueLocal:
		// valide
	default:
		return errors.New("type d'origine invalide : doit être 'princeps', 'generique_importe' ou 'generique_local'")
	}

	// Règles spécifiques Fabrication Locale (NF/RF/MF)
	if m.OriginType == OriginGeneriqueLocal {
		if m.EconomicFileType != EconomicFilePCSU {
			return errors.New("pour un produit fabriqué localement, le type de dossier économique doit être 'pcsu'")
		}
	}

	// Règles spécifiques Importation (NI/RI/MI - Producteur & CPP)
	if m.OriginType == OriginGeneriqueImporte || m.OriginType == OriginPrinceps {
		if m.EconomicFileType == EconomicFileFOBBenchmark && m.FOBPrice != nil && *m.FOBPrice > 0 {
			if strings.TrimSpace(m.FOBCurrency) == "" {
				m.FOBCurrency = "EUR"
			}
		}
		if m.ForeignMAHCountry != "" && len(strings.TrimSpace(m.ForeignMAHCountry)) != 2 {
			return errors.New("le code pays du titulaire étranger doit comporter 2 lettres (ISO-3166-1 alpha-2)")
		}
	}

	if m.ShelfLifeMonths <= 0 {
		m.ShelfLifeMonths = 36 // Valeur standard 36 mois par défaut
	}

	return nil
}

// Constantes d'origine DM (Décret 20-324)
const (
	DMOriginLocal   = "fabrication_locale"
	DMOriginImporte = "importe"
)

// MedicalDevice représente un dispositif médical selon le Décret 20-324 et l'Arrêté du 10/05/2021
type MedicalDevice struct {
	ID                         uuid.UUID          `json:"id"`
	OrganizationID             uuid.UUID          `json:"organization_id"`
	ProductID                  uuid.UUID          `json:"product_id"`
	OriginType                 string             `json:"origin_type"` // 'fabrication_locale' ou 'importe'
	CommercialNameFR           string             `json:"commercial_name_fr"`
	CommercialNameAR           string             `json:"commercial_name_ar,omitempty"`
	ManufacturerReference      string             `json:"manufacturer_reference"`
	DMClass                    string             `json:"dm_class"` // "I", "IIa", "IIb", "III", "DMDIV"
	ClassificationRule         string             `json:"classification_rule,omitempty"`
	IsSterile                  bool               `json:"is_sterile"`
	SterilizationMethod        string             `json:"sterilization_method,omitempty"`
	IsSingleUse                bool               `json:"is_single_use"`
	IntendedUseFR              string             `json:"intended_use_fr,omitempty"`
	IntendedUseAR              string             `json:"intended_use_ar,omitempty"`
	GMDNCode                   string             `json:"gmdn_code,omitempty"`

	// Fabricant Légal & Traçabilité Importation (Décret 20-324)
	LegalManufacturerName      string             `json:"legal_manufacturer_name,omitempty"`
	LegalManufacturerCountry   string             `json:"legal_manufacturer_country,omitempty"`
	ManufacturerSiteID         *uuid.UUID         `json:"manufacturer_site_id,omitempty"`
	ManufacturerSite           *ManufacturingSite `json:"manufacturer_site,omitempty"`
	EstablishmentID            *uuid.UUID         `json:"establishment_id,omitempty"`
	Establishment              *Establishment     `json:"establishment,omitempty"`

	// Conformité Réglementaire Internationale
	CECertificateNumber        string             `json:"ce_certificate_number,omitempty"`
	NotifiedBodyName           string             `json:"notified_body_name,omitempty"`
	NotifiedBodyNumber         string             `json:"notified_body_number,omitempty"`
	CEExpiryDate               *time.Time         `json:"ce_expiry_date,omitempty"`
	ISO13485CertificateNumber  string             `json:"iso_13485_certificate_number,omitempty"`
	CLVFSCNumber               string             `json:"clv_fsc_number,omitempty"`
	CLVIssuingAuthority        string             `json:"clv_issuing_authority,omitempty"`
	CLVIssueDate               *time.Time         `json:"clv_issue_date,omitempty"`

	// Volet Économique DM
	EconomicFileType           string             `json:"economic_file_type,omitempty"` // 'pcsu' ou 'cif_benchmark'
	PCSUAmount                 *float64           `json:"pcsu_amount,omitempty"`
	CIFPrice                   *float64           `json:"cif_price,omitempty"`
	CIFCurrency                string             `json:"cif_currency,omitempty"`

	// Références Réglementaires ANPP
	HomologationDecisionNumber string             `json:"homologation_decision_number,omitempty"`
	HomologationDate           *time.Time         `json:"homologation_date,omitempty"`
	HomologationExpiryDate     *time.Time         `json:"homologation_expiry_date,omitempty"`
	Status                     string             `json:"status"`
	CreatedAt                  time.Time          `json:"created_at"`
	UpdatedAt                  time.Time          `json:"updated_at"`
}

// Validate valide les attributs du dispositif médical
func (dm *MedicalDevice) Validate() error {
	if strings.TrimSpace(dm.CommercialNameFR) == "" {
		return errors.New("la dénomination commerciale du DM est obligatoire")
	}
	if strings.TrimSpace(dm.ManufacturerReference) == "" {
		return errors.New("la référence fabricant du DM est obligatoire")
	}
	switch dm.DMClass {
	case "I", "IIa", "IIb", "III", "DMDIV":
		// valide
	default:
		return errors.New("classe de DM non reconnue : doit être 'I', 'IIa', 'IIb', 'III' ou 'DMDIV'")
	}
	if dm.OriginType == "" {
		dm.OriginType = DMOriginLocal
	}
	if dm.OriginType != DMOriginLocal && dm.OriginType != DMOriginImporte {
		return errors.New("l'origine du DM doit être 'fabrication_locale' ou 'importe'")
	}
	if dm.OriginType == DMOriginImporte {
		if dm.LegalManufacturerCountry != "" && len(strings.TrimSpace(dm.LegalManufacturerCountry)) != 2 {
			return errors.New("le code pays du fabricant légal étranger doit comporter 2 lettres (ISO-3166-1 alpha-2)")
		}
	}
	return nil
}

// MedicineSiteLink association entre un médicament et un site intervenant (1.H.8)
type MedicineSiteLink struct {
	ID            uuid.UUID `json:"id"`
	MedicineID    uuid.UUID `json:"medicine_id"`
	SiteID        uuid.UUID `json:"site_id"`
	RoleInProduct string    `json:"role_in_product"` // fabricant_sa, fabricant_pf, conditionneur_primaire, conditionneur_secondaire, controle_liberation
	CreatedAt     time.Time `json:"created_at"`
}
