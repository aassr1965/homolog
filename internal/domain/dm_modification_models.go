package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Types de modifications post-homologation (Décret 20-324 Art. 22 à 25)
const (
	DMModTypeSubstantielle    = "SUBSTANTIELLE"
	DMModTypeNonSubstantielle = "NON_SUBSTANTIELLE"
)

// Statuts d'instruction d'une modification DM
const (
	DMModStatusBrouillon      = "brouillon"
	DMModStatusSoumis         = "soumis_ddm"
	DMModStatusEnInstruction  = "en_instruction"
	DMModStatusApprouve       = "approuve_avenant"
	DMModStatusRejete         = "rejete"
	DMModStatusNouveauDossier = "nouveau_dossier_requis"
)

// Tarifs réglementaires officiels (Ordonnance n° 21-07 du 8 juin 2021 Art. 31)
const (
	DMFeeSubstantielle    = 45000.00 // DZD (Redevance avenant substantiel BNA Chéraga)
	DMFeeNonSubstantielle = 15000.00 // DZD (Redevance notification administrative)
)

// DMCategoryDef représente une catégorie standard de modification post-homologation
type DMCategoryDef struct {
	Code        string  `json:"code"`
	Type        string  `json:"type"`
	Label       string  `json:"label"`
	LegalRef    string  `json:"legal_ref"`
	DefaultFee  float64 `json:"default_fee"`
	ISOStandard string  `json:"iso_standard"`
}

// Catalogue officiel des modifications DM
var StandardDMCategories = []DMCategoryDef{
	// 1. Modifications Substantielles (Art. 23 Décret 20-324)
	{
		Code:        "DM_SUB_01",
		Type:        DMModTypeSubstantielle,
		Label:       "Modification de la destination ou des indications cliniques prévues",
		LegalRef:    "Décret 20-324 Art. 23 & ISO 14155",
		DefaultFee:  DMFeeSubstantielle,
		ISOStandard: "ISO 14155",
	},
	{
		Code:        "DM_SUB_02",
		Type:        DMModTypeSubstantielle,
		Label:       "Modification de la composition des matières premières en contact direct patient",
		LegalRef:    "Décret 20-324 Art. 23 & ISO 10993",
		DefaultFee:  DMFeeSubstantielle,
		ISOStandard: "ISO 10993",
	},
	{
		Code:        "DM_SUB_03",
		Type:        DMModTypeSubstantielle,
		Label:       "Changement du procédé ou cycle de stérilisation ou site sous-traitant",
		LegalRef:    "Décret 20-324 Art. 23 & ISO 11135 / ISO 11137",
		DefaultFee:  DMFeeSubstantielle,
		ISOStandard: "ISO 11135",
	},
	{
		Code:        "DM_SUB_04",
		Type:        DMModTypeSubstantielle,
		Label:       "Modification du conditionnement primaire et de la barrière stérile",
		LegalRef:    "Décret 20-324 Art. 23 & ISO 11607",
		DefaultFee:  DMFeeSubstantielle,
		ISOStandard: "ISO 11607",
	},
	{
		Code:        "DM_SUB_05",
		Type:        DMModTypeSubstantielle,
		Label:       "Modification majeure de conception ou spécifications de performance",
		LegalRef:    "Décret 20-324 Art. 23 & ISO 14971",
		DefaultFee:  DMFeeSubstantielle,
		ISOStandard: "ISO 14971",
	},
	{
		Code:        "DM_SUB_06",
		Type:        DMModTypeSubstantielle,
		Label:       "Transfert ou changement du site de fabrication ou sous-traitant critique",
		LegalRef:    "Décret 20-324 Art. 23 & ISO 13485",
		DefaultFee:  DMFeeSubstantielle,
		ISOStandard: "ISO 13485",
	},
	{
		Code:        "DM_SUB_07",
		Type:        DMModTypeSubstantielle,
		Label:       "Reclassification ou changement de règle de classification",
		LegalRef:    "Décret 20-324 Annexe I",
		DefaultFee:  DMFeeSubstantielle,
		ISOStandard: "Décret 20-324",
	},

	// 2. Modifications Non-Substantielles (Art. 24 Décret 20-324)
	{
		Code:        "DM_MIN_01",
		Type:        DMModTypeNonSubstantielle,
		Label:       "Changement administratif (raison sociale ou adresse du siège sans transfert d'usine)",
		LegalRef:    "Décret 20-324 Art. 24",
		DefaultFee:  DMFeeNonSubstantielle,
		ISOStandard: "Administratif",
	},
	{
		Code:        "DM_MIN_02",
		Type:        DMModTypeNonSubstantielle,
		Label:       "Changement de dénomination commerciale ou référence catalogue secondaire",
		LegalRef:    "Décret 20-324 Art. 24",
		DefaultFee:  DMFeeNonSubstantielle,
		ISOStandard: "Commercial",
	},
	{
		Code:        "DM_MIN_03",
		Type:        DMModTypeNonSubstantielle,
		Label:       "Mise à jour mineure de l'étiquetage ou notice d'utilisation bilingue (Arabe / Français)",
		LegalRef:    "Décret 20-324 Art. 24",
		DefaultFee:  DMFeeNonSubstantielle,
		ISOStandard: "Étiquetage",
	},
	{
		Code:        "DM_MIN_04",
		Type:        DMModTypeNonSubstantielle,
		Label:       "Conditionnement tertiaire ou logistique sans impact barrière stérile",
		LegalRef:    "Décret 20-324 Art. 24",
		DefaultFee:  DMFeeNonSubstantielle,
		ISOStandard: "Logistique",
	},
}

// DMModification représente une demande de modification post-homologation
type DMModification struct {
	ID               uuid.UUID  `json:"id"`
	OrganizationID   uuid.UUID  `json:"organization_id"`
	MedicalDeviceID  uuid.UUID  `json:"medical_device_id"`
	DossierID        *uuid.UUID `json:"dossier_id,omitempty"`
	ModificationCode string     `json:"modification_code"`
	ModificationType string     `json:"modification_type"` // SUBSTANTIELLE ou NON_SUBSTANTIELLE
	CategoryCode     string     `json:"category_code"`
	CategoryLabel    string     `json:"category_label"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Justification    string     `json:"justification"`
	CurrentState     string     `json:"current_state"`
	ProposedState    string     `json:"proposed_state"`

	// Matrice d'Impact Normatif ISO
	ImpactRiskISO14971             bool `json:"impact_risk_iso14971"`
	ImpactBiocompatibilityISO10993 bool `json:"impact_biocompatibility_iso10993"`
	ImpactSterilizationISO11135    bool `json:"impact_sterilization_iso11135"`
	ImpactPackagingISO11607        bool `json:"impact_packaging_iso11607"`
	ImpactClinicalISO14155         bool `json:"impact_clinical_iso14155"`
	ImpactQualityISO13485          bool `json:"impact_quality_iso13485"`

	// Financier & Redevance (Ord. 21-07 Art. 31)
	FeeAmount      float64    `json:"fee_amount"`
	FeeReceiptRef  string     `json:"fee_receipt_ref"`
	FeePaymentDate *time.Time `json:"fee_payment_date,omitempty"`

	// Workflow Réglementaire
	Status             string     `json:"status"`
	ANPPSubmissionDate *time.Time `json:"anpp_submission_date,omitempty"`
	ANPPDecisionDate   *time.Time `json:"anpp_decision_date,omitempty"`
	AvenantNumber      string     `json:"avenant_number"`
	AvenantDate        *time.Time `json:"avenant_date,omitempty"`
	DecisionNotes      string     `json:"decision_notes"`

	// Scellement Électronique Pharmacien DT (21 CFR Part 11)
	PDTSignedAt       *time.Time `json:"pdt_signed_at,omitempty"`
	PDTSignedBy       *uuid.UUID `json:"pdt_signed_by,omitempty"`
	PDTSignatureSeal  string     `json:"pdt_signature_seal"`
	PDTSignatureNotes string     `json:"pdt_signature_notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Entités rattachées (chargées optionnellement)
	MedicalDevice *MedicalDevice `json:"medical_device,omitempty"`
}

// CalculateFee détermine le montant de la redevance légale
func (m *DMModification) CalculateFee() float64 {
	if m.ModificationType == DMModTypeSubstantielle {
		return DMFeeSubstantielle
	}
	return DMFeeNonSubstantielle
}

// GenerateSeal génère une empreinte cryptographique SHA-256 infalsifiable pour le Pharmacien DT
func (m *DMModification) GenerateSeal(pdtName, decisionMIPH string) string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%.2f|%s|%s|%s",
		m.ID.String(),
		m.OrganizationID.String(),
		m.MedicalDeviceID.String(),
		m.ModificationCode,
		m.ModificationType,
		m.FeeAmount,
		pdtName,
		decisionMIPH,
		time.Now().UTC().Format(time.RFC3339),
	)
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("DM-AVN-SEAL-%s", hex.EncodeToString(hash[:16]))
}

// Validate vérifie l'intégrité et la conformité des champs obligatoires
func (m *DMModification) Validate() error {
	if m.OrganizationID == uuid.Nil {
		return errors.New("l'identifiant de l'organisation est obligatoire")
	}
	if m.MedicalDeviceID == uuid.Nil {
		return errors.New("l'identifiant du dispositif médical est obligatoire")
	}
	if strings.TrimSpace(m.ModificationCode) == "" {
		return errors.New("le code de modification est obligatoire")
	}
	if m.ModificationType != DMModTypeSubstantielle && m.ModificationType != DMModTypeNonSubstantielle {
		return errors.New("type de modification invalide (SUBSTANTIELLE ou NON_SUBSTANTIELLE requis)")
	}
	if strings.TrimSpace(m.Title) == "" {
		return errors.New("le titre de la modification est obligatoire")
	}
	if strings.TrimSpace(m.Description) == "" {
		return errors.New("la description détaillée est obligatoire")
	}
	if m.FeeAmount <= 0 {
		m.FeeAmount = m.CalculateFee()
	}
	return nil
}

// DMModificationStats récapitule les indicateurs du cockpit de modifications
type DMModificationStats struct {
	TotalCount            int     `json:"total_count"`
	SubstantielleCount    int     `json:"substantielle_count"`
	NonSubstantielleCount int     `json:"non_substantielle_count"`
	InInstructionCount    int     `json:"in_instruction_count"`
	ApprouveAvenantCount  int     `json:"approuve_avenant_count"`
	TotalFeesDZD          float64 `json:"total_fees_dzd"`
}
