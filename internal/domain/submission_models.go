package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CheckSeverity niveau de gravité d'un contrôle pré-dépôt
type CheckSeverity string

const (
	SeverityCritical CheckSeverity = "CRITICAL"
	SeverityWarning  CheckSeverity = "WARNING"
	SeverityInfo     CheckSeverity = "INFO"
)

// CheckItemStatus résultat unitaire d'un contrôle
type CheckItemStatus string

const (
	CheckStatusConforme      CheckItemStatus = "CONFORME"
	CheckStatusBloquant      CheckItemStatus = "BLOQUANT"
	CheckStatusAvertissement CheckItemStatus = "AVERTISSEMENT"
)

// PreSubmissionCheckItem détail d'une règle réglementaire auditée
type PreSubmissionCheckItem struct {
	Code        string          `json:"code" xml:"code"`
	Category    string          `json:"category" xml:"category"` // ADMINISTRATIF, FINANCIER_BV1, TECHNIQUE_CTD, VISA_PDT
	Label       string          `json:"label" xml:"label"`
	Status      CheckItemStatus `json:"status" xml:"status"`
	Severity    CheckSeverity   `json:"severity" xml:"severity"`
	Details     string          `json:"details" xml:"details"`
	Remediation string          `json:"remediation,omitempty" xml:"remediation,omitempty"`
}

// PreSubmissionValidationReport synthèse d'audit de conformité pré-dépôt ANPP
type PreSubmissionValidationReport struct {
	DossierID             uuid.UUID                `json:"dossier_id"`
	DossierCode           string                   `json:"dossier_code"`
	ProcedureCode         ProcedureCode            `json:"procedure_code"`
	ProductType           ProductType              `json:"product_type"`
	ProductName           string                   `json:"product_name"`
	IsEligible            bool                     `json:"is_eligible"`
	CompletenessScore     int                      `json:"completeness_score"` // 0-100%
	CriticalBlockersCount int                      `json:"critical_blockers_count"`
	WarningsCount         int                      `json:"warnings_count"`
	Checks                []PreSubmissionCheckItem `json:"checks"`
	EvaluatedAt           time.Time                `json:"evaluated_at"`
	EvaluatedBy           string                   `json:"evaluated_by"`
}

// ManifestDocumentItem pièce documentaire scellée dans le manifeste E-Services
type ManifestDocumentItem struct {
	ModuleNumber           int                    `json:"module_number" xml:"module_number"`
	ItemCode               string                 `json:"item_code" xml:"item_code"`
	TitleFR                string                 `json:"title_fr" xml:"title_fr"`
	RequirementLevel       RequirementLevel       `json:"requirement_level" xml:"requirement_level"`
	IsMandatory            bool                   `json:"is_mandatory" xml:"is_mandatory"`
	Status                 ChecklistItemStatus    `json:"status" xml:"status"`
	SecurityClassification SecurityClassification `json:"security_classification" xml:"security_classification"`
	DocumentCode           string                 `json:"document_code,omitempty" xml:"document_code,omitempty"`
	DocumentVersion        string                 `json:"document_version,omitempty" xml:"document_version,omitempty"`
	TaxonomyPath           string                 `json:"taxonomy_path,omitempty" xml:"taxonomy_path,omitempty"`
	SHA256Checksum         string                 `json:"sha256_checksum,omitempty" xml:"sha256_checksum,omitempty"`
}

// ANPPEservicesManifest package structuré de télé-soumission E-Services ANPP
type ANPPEservicesManifest struct {
	XMLName xml.Name `json:"-" xml:"ANPPEservicesManifest"`

	PortalStandard string    `json:"portal_standard" xml:"portal_standard"`
	DossierCode    string    `json:"dossier_code" xml:"dossier_code"`
	ProcedureCode  string    `json:"procedure_code" xml:"procedure_code"`
	ProductType    string    `json:"product_type" xml:"product_type"`
	GeneratedAt    time.Time `json:"generated_at" xml:"generated_at"`
	GeneratorEmail string    `json:"generator_email" xml:"generator_email"`

	// Établissement pharmaceutique demandeur (Décret 20-325)
	Applicant struct {
		TradeName                 string `json:"trade_name" xml:"trade_name"`
		AgreementNumber           string `json:"agreement_number" xml:"agreement_number"`
		AgreementDate             string `json:"agreement_date,omitempty" xml:"agreement_date,omitempty"`
		Wilaya                    string `json:"wilaya,omitempty" xml:"wilaya,omitempty"`
		TechnicalDirectorName     string `json:"technical_director_name" xml:"technical_director_name"`
		TechnicalDirectorDecision string `json:"technical_director_decision,omitempty" xml:"technical_director_decision,omitempty"`
	} `json:"applicant" xml:"applicant"`

	// Produit de santé (Médicament ou Dispositif Médical)
	Product struct {
		BrandName          string `json:"brand_name" xml:"brand_name"`
		GenericNameDCI     string `json:"generic_name_dci,omitempty" xml:"generic_name_dci,omitempty"`
		Dosage             string `json:"dosage,omitempty" xml:"dosage,omitempty"`
		PharmaceuticalForm string `json:"pharmaceutical_form,omitempty" xml:"pharmaceutical_form,omitempty"`
		DMClass            string `json:"dm_class,omitempty" xml:"dm_class,omitempty"`
		OriginType         string `json:"origin_type,omitempty" xml:"origin_type,omitempty"`
	} `json:"product" xml:"product"`

	// Quittance financière Trésorerie ANPP (BV1)
	FinancialClearance struct {
		SlipNumber          string  `json:"slip_number" xml:"slip_number"`
		SlipType            string  `json:"slip_type" xml:"slip_type"`
		AmountDZD           float64 `json:"amount_dzd" xml:"amount_dzd"`
		AmountWordsFR       string  `json:"amount_words_fr" xml:"amount_words_fr"`
		AmountWordsAR       string  `json:"amount_words_ar" xml:"amount_words_ar"`
		PaymentReference    string  `json:"payment_reference" xml:"payment_reference"`
		PaidAt              string  `json:"paid_at" xml:"paid_at"`
		BeneficiaryTreasury string  `json:"beneficiary_treasury" xml:"beneficiary_treasury"`
		IsCleared           bool    `json:"is_cleared" xml:"is_cleared"`
	} `json:"financial_clearance" xml:"financial_clearance"`

	// Inventaire des pièces jointes et conformité CTD
	Inventory struct {
		TotalItemsCount     int                    `json:"total_items_count" xml:"total_items_count"`
		MandatoryItemsCount int                    `json:"mandatory_items_count" xml:"mandatory_items_count"`
		AttachedItemsCount  int                    `json:"attached_items_count" xml:"attached_items_count"`
		CompletenessPercent int                    `json:"completeness_percent" xml:"completeness_percent"`
		Documents           []ManifestDocumentItem `json:"documents" xml:"documents>item"`
	} `json:"inventory" xml:"inventory"`

	// Sceau d'intégrité réglementaire et Visa PDT (21 CFR Part 11)
	IntegritySeal struct {
		IsPDTSigned       bool       `json:"is_pdt_signed" xml:"is_pdt_signed"`
		PDTSignedBy       string     `json:"pdt_signed_by,omitempty" xml:"pdt_signed_by,omitempty"`
		PDTSignedAt       *time.Time `json:"pdt_signed_at,omitempty" xml:"pdt_signed_at,omitempty"`
		PDTSignatureRole  string     `json:"pdt_signature_role,omitempty" xml:"pdt_signature_role,omitempty"`
		PDTSignatureSeal  string     `json:"pdt_signature_seal,omitempty" xml:"pdt_signature_seal,omitempty"`
		ManifestDigestHex string     `json:"manifest_digest_hex" xml:"manifest_digest_hex"`
	} `json:"integrity_seal" xml:"integrity_seal"`
}

// EvaluatePreSubmission effectue l'audit de complétude et d'éligibilité pré-dépôt ANPP
func EvaluatePreSubmission(dossier *Dossier, slips []ANPPPaymentSlip, est *Establishment, evaluatorEmail string) *PreSubmissionValidationReport {
	report := &PreSubmissionValidationReport{
		DossierID:     dossier.ID,
		DossierCode:   dossier.DossierCode,
		ProcedureCode: dossier.ProcedureCode,
		ProductType:   dossier.ProductType,
		EvaluatedAt:   time.Now().UTC(),
		EvaluatedBy:   evaluatorEmail,
		Checks:        make([]PreSubmissionCheckItem, 0),
	}

	if dossier.Medicine != nil {
		report.ProductName = fmt.Sprintf("%s (%s %s)", dossier.Medicine.BrandNameFR, dossier.Medicine.DCIFR, dossier.Medicine.DosageFR)
	} else if dossier.MedicalDevice != nil {
		report.ProductName = fmt.Sprintf("%s (Classe %s)", dossier.MedicalDevice.CommercialNameFR, dossier.MedicalDevice.DMClass)
	} else {
		report.ProductName = dossier.DossierCode
	}

	// 1. Contrôles ADMINISTRATIFS (Établissement demandeur & Pharmacien DT)
	if est != nil && est.MIPHLicenseNumber != "" {
		report.Checks = append(report.Checks, PreSubmissionCheckItem{
			Code:     "ADM_MIPH_AGREEMENT",
			Category: "ADMINISTRATIF",
			Label:    "Agrément Établissement Pharmaceutique (MIPH)",
			Status:   CheckStatusConforme,
			Severity: SeverityCritical,
			Details:  fmt.Sprintf("Agrément N° %s valide délivré à %s", est.MIPHLicenseNumber, est.NameFR),
		})
	} else {
		report.Checks = append(report.Checks, PreSubmissionCheckItem{
			Code:        "ADM_MIPH_AGREEMENT",
			Category:    "ADMINISTRATIF",
			Label:       "Agrément Établissement Pharmaceutique (MIPH)",
			Status:      CheckStatusBloquant,
			Severity:    SeverityCritical,
			Details:     "Aucun agrément MIPH enregistré ou valide pour l'établissement demandeur.",
			Remediation: "Compléter la fiche établissement dans le Référentiel Pharmaceutique.",
		})
	}

	if est != nil && est.TechnicalDirectorName != "" && est.TechnicalDirectorDecisionRef != "" {
		report.Checks = append(report.Checks, PreSubmissionCheckItem{
			Code:     "ADM_PDT_NOMINATION",
			Category: "ADMINISTRATIF",
			Label:    "Décision Nominative du Pharmacien DT (1.A.7)",
			Status:   CheckStatusConforme,
			Severity: SeverityCritical,
			Details:  fmt.Sprintf("Pharmacien DT : %s (Décision : %s)", est.TechnicalDirectorName, est.TechnicalDirectorDecisionRef),
		})
	} else {
		report.Checks = append(report.Checks, PreSubmissionCheckItem{
			Code:        "ADM_PDT_NOMINATION",
			Category:    "ADMINISTRATIF",
			Label:       "Décision Nominative du Pharmacien DT (1.A.7)",
			Status:      CheckStatusBloquant,
			Severity:    SeverityCritical,
			Details:     "Identité ou décision ministérielle nominative du Pharmacien DT manquante.",
			Remediation: "Renseigner la décision ministérielle nominative dans la fiche établissement.",
		})
	}

	// 2. Contrôles FINANCIERS (Redevance d'enregistrement ANPP & Bordereau BV1)
	var validBV1 *ANPPPaymentSlip
	for i := range slips {
		s := slips[i]
		if s.SlipType == "BV1" && s.Status == "paye" {
			validBV1 = &s
			break
		}
	}

	if validBV1 != nil && validBV1.PaymentReference != "" {
		report.Checks = append(report.Checks, PreSubmissionCheckItem{
			Code:     "FIN_BV1_RECEIPT",
			Category: "FINANCIER_BV1",
			Label:    "Quittance Trésorerie ANPP (Bordereau BV1)",
			Status:   CheckStatusConforme,
			Severity: SeverityCritical,
			Details:  fmt.Sprintf("Bordereau %s réglé (Quittance : %s, Montant : %s DZD)", validBV1.SlipNumber, validBV1.PaymentReference, formatDZDAmount(validBV1.AmountDZD)),
		})
	} else {
		// Vérifier si un BV1 a été émis mais pas encore payé
		hasPendingBV1 := false
		for _, s := range slips {
			if s.SlipType == "BV1" && s.Status == "emis" {
				hasPendingBV1 = true
				break
			}
		}

		if hasPendingBV1 {
			report.Checks = append(report.Checks, PreSubmissionCheckItem{
				Code:        "FIN_BV1_RECEIPT",
				Category:    "FINANCIER_BV1",
				Label:       "Quittance Trésorerie ANPP (Bordereau BV1)",
				Status:      CheckStatusBloquant,
				Severity:    SeverityCritical,
				Details:     "Bordereau BV1 généré mais en attente d'enregistrement de quittance du Trésor.",
				Remediation: "Enregistrer le numéro de quittance du Trésor public ou l'avis d'opéré de virement.",
			})
		} else {
			report.Checks = append(report.Checks, PreSubmissionCheckItem{
				Code:        "FIN_BV1_RECEIPT",
				Category:    "FINANCIER_BV1",
				Label:       "Quittance Trésorerie ANPP (Bordereau BV1)",
				Status:      CheckStatusBloquant,
				Severity:    SeverityCritical,
				Details:     "Aucun bordereau de versement BV1 émis pour ce dossier d'enregistrement.",
				Remediation: "Générer le bordereau BV1 conforme au barème réglementaire en vigueur.",
			})
		}
	}

	// 3. Contrôles TECHNIQUES CTD (Pièces obligatoires [E] et complétude)
	completeness := CalculateCompleteness(dossier.Items)
	report.CompletenessScore = completeness

	missingMandatory := make([]string, 0)
	for _, it := range dossier.Items {
		if it.IsMandatory && it.RequirementLevel == RequirementExiged {
			if it.GEDDocumentID == nil || it.Status == ItemStatusMissing || it.Status == ItemStatusRejected {
				missingMandatory = append(missingMandatory, fmt.Sprintf("[%s] %s", it.ItemCode, it.TitleFR))
			}
		}
	}

	if len(missingMandatory) == 0 && completeness == 100 {
		report.Checks = append(report.Checks, PreSubmissionCheckItem{
			Code:     "CTD_MANDATORY_COMPLETION",
			Category: "TECHNIQUE_CTD",
			Label:    "Complétude des Pièces Exigées [E]",
			Status:   CheckStatusConforme,
			Severity: SeverityCritical,
			Details:  fmt.Sprintf("Complétude à 100%%. Les %d pièces obligatoires sont toutes attachées et conformes.", len(dossier.Items)),
		})
	} else {
		report.Checks = append(report.Checks, PreSubmissionCheckItem{
			Code:        "CTD_MANDATORY_COMPLETION",
			Category:    "TECHNIQUE_CTD",
			Label:       "Complétude des Pièces Exigées [E]",
			Status:      CheckStatusBloquant,
			Severity:    SeverityCritical,
			Details:     fmt.Sprintf("Complétude à %d%%. %d pièce(s) obligatoire(s) manquante(s) ou rejetée(s).", completeness, len(missingMandatory)),
			Remediation: "Rattacher les documents GED requis pour chaque module incomplet.",
		})
	}

	// 4. Contrôles de VISA DU PHARMACIEN DIRECTEUR TECHNIQUE (1.A.7)
	if dossier.PDTSignedAt != nil && dossier.PDTSignatureSeal != nil && *dossier.PDTSignatureSeal != "" {
		signerStr := "Pharmacien DT"
		if dossier.PDTSignedBy != nil {
			signerStr = dossier.PDTSignedBy.String()
		}
		sealSnippet := *dossier.PDTSignatureSeal
		if len(sealSnippet) > 12 {
			sealSnippet = sealSnippet[:12]
		}
		report.Checks = append(report.Checks, PreSubmissionCheckItem{
			Code:     "VISA_PDT_LEGAL",
			Category: "VISA_PDT",
			Label:    "Visa Réglementaire Pharmacien DT (21 CFR Part 11)",
			Status:   CheckStatusConforme,
			Severity: SeverityCritical,
			Details:  fmt.Sprintf("Visa apposé par %s le %s (Sceau : %s...)", signerStr, dossier.PDTSignedAt.Format("02/01/2006 15:04"), sealSnippet),
		})
	} else {
		report.Checks = append(report.Checks, PreSubmissionCheckItem{
			Code:        "VISA_PDT_LEGAL",
			Category:    "VISA_PDT",
			Label:       "Visa Réglementaire Pharmacien DT (21 CFR Part 11)",
			Status:      CheckStatusBloquant,
			Severity:    SeverityCritical,
			Details:     "Le dossier n'a pas encore reçu le visa formel du Pharmacien Directeur Technique.",
			Remediation: "Le Pharmacien DT titulaire doit apposer sa signature électronique nominative.",
		})
	}

	// 5. Avertissements non bloquants (Ex: Pièces facultatives [F] non rattachées)
	optionalMissing := 0
	for _, it := range dossier.Items {
		if it.RequirementLevel == RequirementFacultative && it.GEDDocumentID == nil {
			optionalMissing++
		}
	}
	if optionalMissing > 0 {
		report.Checks = append(report.Checks, PreSubmissionCheckItem{
			Code:     "CTD_OPTIONAL_DOCS",
			Category: "TECHNIQUE_CTD",
			Label:    "Pièces Complémentaires Facultatives [F]",
			Status:   CheckStatusAvertissement,
			Severity: SeverityWarning,
			Details:  fmt.Sprintf("%d pièce(s) facultative(s) non fournie(s) (non bloquant pour le dépôt).", optionalMissing),
		})
	}

	// Calcul global d'éligibilité
	blockers := 0
	warnings := 0
	for _, c := range report.Checks {
		if c.Status == CheckStatusBloquant {
			blockers++
		} else if c.Status == CheckStatusAvertissement {
			warnings++
		}
	}

	report.CriticalBlockersCount = blockers
	report.WarningsCount = warnings
	report.IsEligible = (blockers == 0)

	return report
}

// BuildANPPEservicesManifest assemble le paquet officiel prêt pour soumission E-Services ANPP
func BuildANPPEservicesManifest(dossier *Dossier, slips []ANPPPaymentSlip, est *Establishment, generatorEmail string) (*ANPPEservicesManifest, error) {
	manifest := &ANPPEservicesManifest{
		PortalStandard: "ANPP_ESERVICES_REGULATORY_PACKAGE_V2.4",
		DossierCode:    dossier.DossierCode,
		ProcedureCode:  string(dossier.ProcedureCode),
		ProductType:    string(dossier.ProductType),
		GeneratedAt:    time.Now().UTC(),
		GeneratorEmail: generatorEmail,
	}

	// Établissement
	if est != nil {
		manifest.Applicant.TradeName = est.NameFR
		manifest.Applicant.AgreementNumber = est.MIPHLicenseNumber
		if est.MIPHLicenseGrantDate != nil {
			manifest.Applicant.AgreementDate = est.MIPHLicenseGrantDate.Format("2006-01-02")
		}
		manifest.Applicant.Wilaya = est.Wilaya
		manifest.Applicant.TechnicalDirectorName = est.TechnicalDirectorName
		manifest.Applicant.TechnicalDirectorDecision = est.TechnicalDirectorDecisionRef
	}

	// Produit
	if dossier.Medicine != nil {
		manifest.Product.BrandName = dossier.Medicine.BrandNameFR
		manifest.Product.GenericNameDCI = dossier.Medicine.DCIFR
		manifest.Product.Dosage = dossier.Medicine.DosageFR
		manifest.Product.PharmaceuticalForm = dossier.Medicine.PharmaceuticalFormFR
		manifest.Product.OriginType = string(dossier.Medicine.OriginType)
	} else if dossier.MedicalDevice != nil {
		manifest.Product.BrandName = dossier.MedicalDevice.CommercialNameFR
		manifest.Product.DMClass = string(dossier.MedicalDevice.DMClass)
	}

	// Redevance BV1
	var paidBV1 *ANPPPaymentSlip
	for i := range slips {
		if slips[i].SlipType == "BV1" && slips[i].Status == "paye" {
			paidBV1 = &slips[i]
			break
		}
	}
	if paidBV1 != nil {
		manifest.FinancialClearance.SlipNumber = paidBV1.SlipNumber
		manifest.FinancialClearance.SlipType = paidBV1.SlipType
		manifest.FinancialClearance.AmountDZD = paidBV1.AmountDZD
		manifest.FinancialClearance.AmountWordsFR = paidBV1.AmountWordsFR
		manifest.FinancialClearance.AmountWordsAR = paidBV1.AmountWordsAR
		manifest.FinancialClearance.PaymentReference = paidBV1.PaymentReference
		if paidBV1.PaidAt != nil {
			manifest.FinancialClearance.PaidAt = paidBV1.PaidAt.Format("2006-01-02")
		}
		manifest.FinancialClearance.BeneficiaryTreasury = "Agent Comptable de l'ANPP - Compte Trésor 300.005.12345/Alger"
		manifest.FinancialClearance.IsCleared = true
	}

	// Inventaire
	manifest.Inventory.TotalItemsCount = len(dossier.Items)
	mandatoryCount := 0
	attachedCount := 0
	docItems := make([]ManifestDocumentItem, 0, len(dossier.Items))

	for _, it := range dossier.Items {
		if it.IsMandatory {
			mandatoryCount++
		}
		if it.GEDDocumentID != nil {
			attachedCount++
		}

		mDoc := ManifestDocumentItem{
			ModuleNumber:           it.ModuleNumber,
			ItemCode:               it.ItemCode,
			TitleFR:                it.TitleFR,
			RequirementLevel:       it.RequirementLevel,
			IsMandatory:            it.IsMandatory,
			Status:                 it.Status,
			SecurityClassification: it.SecurityClassification,
		}

		if it.GEDDocument != nil {
			mDoc.DocumentCode = it.GEDDocument.DocumentCode
			mDoc.DocumentVersion = it.GEDDocument.CurrentVersion
			mDoc.TaxonomyPath = it.GEDDocument.TaxonomyPath

			// Calcul d'empreinte d'intégrité standardisée si non fournie
			hasher := sha256.New()
			hasher.Write([]byte(fmt.Sprintf("%s:%s:%s", it.GEDDocument.ID, it.GEDDocument.DocumentCode, it.GEDDocument.CurrentVersion)))
			mDoc.SHA256Checksum = hex.EncodeToString(hasher.Sum(nil))
		}

		docItems = append(docItems, mDoc)
	}

	manifest.Inventory.MandatoryItemsCount = mandatoryCount
	manifest.Inventory.AttachedItemsCount = attachedCount
	manifest.Inventory.CompletenessPercent = CalculateCompleteness(dossier.Items)
	manifest.Inventory.Documents = docItems

	// Visa PDT
	if dossier.PDTSignedAt != nil {
		manifest.IntegritySeal.IsPDTSigned = true
		if dossier.PDTSignedBy != nil {
			manifest.IntegritySeal.PDTSignedBy = dossier.PDTSignedBy.String()
		}
		manifest.IntegritySeal.PDTSignedAt = dossier.PDTSignedAt
		if dossier.PDTSignatureRole != nil {
			manifest.IntegritySeal.PDTSignatureRole = *dossier.PDTSignatureRole
		}
		if dossier.PDTSignatureSeal != nil {
			manifest.IntegritySeal.PDTSignatureSeal = *dossier.PDTSignatureSeal
		}
	}

	// Calcul du Digest SHA-256 global du paquet
	rawBytes, err := json.Marshal(struct {
		DossierCode string `json:"dossier_code"`
		Applicant   any    `json:"applicant"`
		Product     any    `json:"product"`
		Financial   any    `json:"financial"`
		Inventory   any    `json:"inventory"`
	}{
		DossierCode: manifest.DossierCode,
		Applicant:   manifest.Applicant,
		Product:     manifest.Product,
		Financial:   manifest.FinancialClearance,
		Inventory:   manifest.Inventory,
	})
	if err != nil {
		return nil, fmt.Errorf("échec sérialisation empreinte manifeste : %w", err)
	}

	hasher := sha256.New()
	hasher.Write(rawBytes)
	manifest.IntegritySeal.ManifestDigestHex = hex.EncodeToString(hasher.Sum(nil))

	return manifest, nil
}

// ToJSON sérialise le manifeste en JSON formaté
func (m *ANPPEservicesManifest) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// ToXML sérialise le manifeste en XML réglementaire
func (m *ANPPEservicesManifest) ToXML() ([]byte, error) {
	xmlData, err := xml.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	header := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	return append(header, xmlData...), nil
}

func formatDZDAmount(val float64) string {
	return fmt.Sprintf("%.2f", val)
}
