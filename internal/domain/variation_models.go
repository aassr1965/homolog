package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// VariationClassification qualifie la sévérité officielle d'une modification selon la Ligne Directrice ANPP LD-DEPP07/V01
type VariationClassification string

const (
	ClassMmin VariationClassification = "Mmin" // Modification Mineure (impact minime ou nul)
	ClassMmod VariationClassification = "Mmod" // Modification Modérée (impact potentiel)
	ClassMmaj VariationClassification = "Mmaj" // Modification Majeure (répercussions significatives)
)

// SeverityRank attribue un rang numérique pour le calcul de criticité globale
func (c VariationClassification) SeverityRank() int {
	switch c {
	case ClassMmaj:
		return 3
	case ClassMmod:
		return 2
	case ClassMmin:
		return 1
	default:
		return 0
	}
}

// IsHigherThan compare deux classifications de variation
func (c VariationClassification) IsHigherThan(other VariationClassification) bool {
	return c.SeverityRank() > other.SeverityRank()
}

// Catégories des 34 Rubriques officielles (LD-DEPP07/V01)
const (
	CategoryAdministrative     = "ADMINISTRATIVE"       // Rubriques 1 à 5
	CategoryCEPAPI             = "CEP_API"              // Rubriques 6.a & 6.b
	CategorySubstanceActive32S = "SUBSTANCE_ACTIVE_32S" // Rubriques 7 à 16
	CategoryProduitFini32P     = "PRODUIT_FINI_32P"     // Rubriques 17 à 34
)

// ChangeImpactLevel qualifie le niveau d'impact CTD dans le Change Impact Tree
type ChangeImpactLevel string

const (
	ImpactDirectRed      ChangeImpactLevel = "DIRECT_RED"      // Section CTD au cœur de la modification (remplacement obligatoire)
	ImpactRippleOrange   ChangeImpactLevel = "RIPPLE_ORANGE"   // Section corollaire affectée (effet d'entraînement, vérification/attestation requise)
	ImpactPreservedGreen ChangeImpactLevel = "PRESERVED_GREEN" // Section préservée (candidate attestation de non-modification)
)

// CTDSectionImpact représente l'impact sur une section CTD
type CTDSectionImpact struct {
	SectionCode string            `json:"section_code"` // Ex: "1.A.3", "3.2.S.2", "1.P.2"
	Title       string            `json:"title"`
	Module      int               `json:"module"` // 1, 2, 3, 4, 5
	Level       ChangeImpactLevel `json:"level"`
	Reason      string            `json:"reason"`
}

// VariationCondition représente une condition préalable officielle C1..Cn
type VariationCondition struct {
	ID          string `json:"id"`          // Ex: "C1", "C2"
	Description string `json:"description"` // Ex: "Le profil d'impuretés de la substance active demeure inchangé"
	LegalRef    string `json:"legal_ref"`   // Ex: "LD-DEPP07 Section 5.4.2.1"
}

// RubriqueDefinition encapsule une rubrique officielle parmi les 34
type RubriqueDefinition struct {
	Code                       string                  `json:"code"` // Ex: "1", "7.d.1", "19.a", "30.b"
	Number                     int                     `json:"number"`
	SubCode                    string                  `json:"sub_code"`
	Title                      string                  `json:"title"`
	Category                   string                  `json:"category"`
	CategoryLabel              string                  `json:"category_label"`
	DefaultClassification      VariationClassification `json:"default_classification"`
	EscalatedClassification    VariationClassification `json:"escalated_classification"`
	EscalationConditionSummary string                  `json:"escalation_condition_summary"`
	Conditions                 []VariationCondition    `json:"conditions"`
	RequiredDocuments          []string                `json:"required_documents"`
	ImpactedSections           []CTDSectionImpact      `json:"impacted_sections"`
}

// VariationRequest représente une demande formelle de modification post-AMM
type VariationRequest struct {
	ID                       uuid.UUID               `json:"id"`
	OrganizationID           uuid.UUID               `json:"organization_id"`
	MedicineID               uuid.UUID               `json:"medicine_id"`
	DossierID                *uuid.UUID              `json:"dossier_id,omitempty"`
	TrackingNumber           string                  `json:"tracking_number"`
	Title                    string                  `json:"title"`
	Description              string                  `json:"description,omitempty"`
	IsGrouping               bool                    `json:"is_grouping"`
	GroupingJustification    string                  `json:"grouping_justification,omitempty"`
	OverallClassification    VariationClassification `json:"overall_classification"`
	NominalClassification    VariationClassification `json:"nominal_classification"`
	HasEscalation            bool                    `json:"has_escalation"`
	EscalationSummary        string                  `json:"escalation_summary,omitempty"`
	CurrentApprovedSituation string                  `json:"current_approved_situation"`
	ProposedSituation        string                  `json:"proposed_situation"`
	TechnicalJustification   string                  `json:"technical_justification"`
	NonModificationAttested  bool                    `json:"non_modification_attested"`
	NonModificationDetails   string                  `json:"non_modification_details,omitempty"`
	FeeAmountDZD             float64                 `json:"fee_amount_dzd"`
	FeeSlipID                *uuid.UUID              `json:"fee_slip_id,omitempty"`
	Status                   string                  `json:"status"` // brouillon, en_evaluation, reclassifie, pret_visa_dt, soumis_anpp, approuve, rejete
	DTVerifiedBy             *uuid.UUID              `json:"dt_verified_by,omitempty"`
	DTVerifiedAt             *time.Time              `json:"dt_verified_at,omitempty"`
	DTSignatureSeal          string                  `json:"dt_signature_seal,omitempty"`
	DTVerificationNotes      string                  `json:"dt_verification_notes,omitempty"`
	CreatedBy                *uuid.UUID              `json:"created_by,omitempty"`
	CreatedAt                time.Time               `json:"created_at"`
	UpdatedAt                time.Time               `json:"updated_at"`

	// Relations enrichies
	MedicineName       string              `json:"medicine_name,omitempty"`
	MedicineDCI        string              `json:"medicine_dci,omitempty"`
	MedicineDosage     string              `json:"medicine_dosage,omitempty"`
	MedicineForm       string              `json:"medicine_form,omitempty"`
	RegistrationNumber string              `json:"registration_number,omitempty"`
	Items              []VariationItem     `json:"items,omitempty"`
	Documents          []VariationDocument `json:"documents,omitempty"`
	ChangeImpactTree   *ChangeImpactTree   `json:"change_impact_tree,omitempty"`
}

// VariationItem représente une rubrique sélectionnée dans la demande
type VariationItem struct {
	ID                      uuid.UUID               `json:"id"`
	VariationRequestID      uuid.UUID               `json:"variation_request_id"`
	RubriqueCode            string                  `json:"rubrique_code"`
	RubriqueTitle           string                  `json:"rubrique_title"`
	Category                string                  `json:"category"`
	NominalClassification   VariationClassification `json:"nominal_classification"`
	EffectiveClassification VariationClassification `json:"effective_classification"`
	IsEscalated             bool                    `json:"is_escalated"`
	EscalationReason        string                  `json:"escalation_reason,omitempty"`
	ConditionsEvaluated     map[string]bool         `json:"conditions_evaluated"`
	AllConditionsMet        bool                    `json:"all_conditions_met"`
	CurrentState            string                  `json:"current_state,omitempty"`
	ProposedState           string                  `json:"proposed_state,omitempty"`
	Notes                   string                  `json:"notes,omitempty"`
	CreatedAt               time.Time               `json:"created_at"`
	UpdatedAt               time.Time               `json:"updated_at"`
}

// VariationDocument représente l'une des 12 pièces obligatoires exigées par la LD-DEPP07
type VariationDocument struct {
	ID                 uuid.UUID  `json:"id"`
	VariationRequestID uuid.UUID  `json:"variation_request_id"`
	DocIndex           int        `json:"doc_index"` // 1 à 12
	DocCode            string     `json:"doc_code"`
	Title              string     `json:"title"`
	Description        string     `json:"description,omitempty"`
	IsMandatory        bool       `json:"is_mandatory"`
	IsAttached         bool       `json:"is_attached"`
	GEDDocumentID      *uuid.UUID `json:"ged_document_id,omitempty"`
	GEDFileName        string     `json:"ged_file_name,omitempty"`
	GEDHashSHA256      string     `json:"ged_hash_sha256,omitempty"`
	VerifiedStatus     string     `json:"verified_status"` // en_attente, conforme, non_conforme
	VerifiedBy         *uuid.UUID `json:"verified_by,omitempty"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	Notes              string     `json:"notes,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ChangeImpactTree consolide l'arborescence des impacts tri-couleur pour l'UI et le CoDir
type ChangeImpactTree struct {
	DirectRed      []CTDSectionImpact `json:"direct_red"`
	RippleOrange   []CTDSectionImpact `json:"ripple_orange"`
	PreservedGreen []CTDSectionImpact `json:"preserved_green"`
	Stats          map[string]int     `json:"stats"`
}

// MasterRubriquesRegistry définit les 34 Rubriques officielles de l'ANPP (Ligne Directrice LD-DEPP07/V01)
var MasterRubriquesRegistry = []RubriqueDefinition{
	// =========================================================================
	// I. MODIFICATIONS ADMINISTRATIVES (Rubriques 1 à 5)
	// =========================================================================
	{
		Code:                       "1",
		Number:                     1,
		SubCode:                    "1",
		Title:                      "Changement dans la dénomination commerciale du produit fini",
		Category:                   CategoryAdministrative,
		CategoryLabel:              "I. Modifications Administratives",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si risque de confusion phonétique ou visuelle avec un autre produit enregistré",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Absence de risque de confusion phonétique ou visuelle avec une spécialité existante", LegalRef: "LD-DEPP07 Section 5.4.1.1"},
			{ID: "C2", Description: "La DCI et la composition qualitative et quantitative demeurent strictement inchangées", LegalRef: "LD-DEPP07 Section 5.4.1.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-06", "DOC-12"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.A.3", Title: "Formulaire officiel de demande", Module: 1, Level: ImpactDirectRed, Reason: "Mise à jour du nom de marque"},
			{SectionCode: "1.P.1", Title: "Résumé des Caractéristiques du Produit (RCP)", Module: 1, Level: ImpactDirectRed, Reason: "Modification du titre du RCP"},
			{SectionCode: "1.P.2", Title: "Notice bilingue FR/AR et maquettes d'étui", Module: 1, Level: ImpactDirectRed, Reason: "Remplacement complet des maquettes"},
			{SectionCode: "3.2.P.1", Title: "Description et composition", Module: 3, Level: ImpactRippleOrange, Reason: "Vérification cohérence dénomination"},
		},
	},
	{
		Code:                       "2",
		Number:                     2,
		SubCode:                    "2",
		Title:                      "Changement dans la dénomination de la substance active, d'un excipient ou du conditionnement",
		Category:                   CategoryAdministrative,
		CategoryLabel:              "I. Modifications Administratives",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si altération de la spécification chimique ou de la monographie",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "L'entité chimique ou la spécification demeure strictement identique", LegalRef: "LD-DEPP07 Section 5.4.1.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-07", "DOC-12"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.A.3", Title: "Formulaire de demande", Module: 1, Level: ImpactDirectRed, Reason: "Actualisation de la dénomination"},
			{SectionCode: "3.2.S.1", Title: "Informations générales SA", Module: 3, Level: ImpactRippleOrange, Reason: "Harmonisation nomenclature"},
		},
	},
	{
		Code:                    "3",
		Number:                  3,
		SubCode:                 "3",
		Title:                   "Changement ou attribution du code ATC par l'OMS",
		Category:                CategoryAdministrative,
		CategoryLabel:           "I. Modifications Administratives",
		DefaultClassification:   ClassMmin,
		EscalatedClassification: ClassMmin,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Attribution officielle par le centre collaborateur de l'OMS pour la méthodologie des statistiques sur les médicaments", LegalRef: "LD-DEPP07 Section 5.4.1.3"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-07"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.A.3", Title: "Formulaire de demande", Module: 1, Level: ImpactDirectRed, Reason: "Mise à jour du code ATC"},
			{SectionCode: "1.P.1", Title: "RCP", Module: 1, Level: ImpactDirectRed, Reason: "Section 5.1 du RCP"},
		},
	},
	{
		Code:                       "4",
		Number:                     4,
		SubCode:                    "4",
		Title:                      "Changement du nom et/ou adresse du titulaire AMM, fabricant API, PF ou sous-traitant",
		Category:                   CategoryAdministrative,
		CategoryLabel:              "I. Modifications Administratives",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si le site physique ou l'entité juridique sous-jacente change",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "L'entité juridique et le site physique de fabrication demeurent identiques (simple changement administratif d'adresse ou raison sociale)", LegalRef: "LD-DEPP07 Section 5.4.1.4"},
			{ID: "C2", Description: "Autorisation administrative d'établissement ou agrément MIPH à jour fourni", LegalRef: "LD-DEPP07 Section 5.4.1.4"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-06", "DOC-08", "DOC-11"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.A.3", Title: "Formulaire de demande", Module: 1, Level: ImpactDirectRed, Reason: "Mise à jour adresse/raison sociale"},
			{SectionCode: "1.H.8", Title: "Liste des sites intervenants visée", Module: 1, Level: ImpactDirectRed, Reason: "Remplacement de la liste des sites"},
			{SectionCode: "1.P.2", Title: "Articles de conditionnement", Module: 1, Level: ImpactRippleOrange, Reason: "Mentions légales du fabricant sur étui"},
		},
	},
	{
		Code:                       "5",
		Number:                     5,
		SubCode:                    "5",
		Title:                      "Suppression d'un site de fabrication, contrôle, conditionnement ou stockage",
		Category:                   CategoryAdministrative,
		CategoryLabel:              "I. Modifications Administratives",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si aucun site alternatif n'est préalablement approuvé dans le dossier",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Maintien d'au moins un site alternatif approuvé pour chaque opération concernée", LegalRef: "LD-DEPP07 Section 5.4.1.5"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-06", "DOC-07"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.H.8", Title: "Liste des sites intervenants", Module: 1, Level: ImpactDirectRed, Reason: "Suppression du site de la liste"},
			{SectionCode: "3.2.P.3.1", Title: "Fabricants du produit fini", Module: 3, Level: ImpactRippleOrange, Reason: "Actualisation du tableau des sites"},
		},
	},

	// =========================================================================
	// II. MODIFICATIONS RELATIVES AU CEP OU PRÉQUALIFICATION API (Rubriques 6.a & 6.b)
	// =========================================================================
	{
		Code:                       "6.a.1",
		Number:                     6,
		SubCode:                    "6.a.1",
		Title:                      "Soumission d'un CEP nouveau ou mis à jour d'un fabricant déjà approuvé (Mise à jour sans impact)",
		Category:                   CategoryCEPAPI,
		CategoryLabel:              "II. CEP & Préqualification API",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si révision de la voie de synthèse, profil d'impuretés ou réessai",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Aucun changement dans la voie de synthèse ni le profil d'impuretés de la SA", LegalRef: "LD-DEPP07 Section 5.4.1.2"},
			{ID: "C2", Description: "La période de recontrôle et les conditions de stockage demeurent inchangées", LegalRef: "LD-DEPP07 Section 5.4.1.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05", "DOC-07"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.M.S.3", Title: "CEP EDQM et lettre d'accès", Module: 1, Level: ImpactDirectRed, Reason: "Nouveau CEP visé"},
			{SectionCode: "3.2.S.4", Title: "Contrôle de la SA", Module: 3, Level: ImpactRippleOrange, Reason: "Vérification spécifications"},
		},
	},
	{
		Code:                       "6.b.1",
		Number:                     6,
		SubCode:                    "6.b.1",
		Title:                      "Soumission d'un CEP provenant d'un nouveau fabricant d'API",
		Category:                   CategoryCEPAPI,
		CategoryLabel:              "II. CEP & Préqualification API",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si le nouveau fabricant n'a pas de CEP valide ou si des impuretés nouvelles sont détectées",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le CEP présenté est valide et délivré par l'EDQM sans réserve", LegalRef: "LD-DEPP07 Section 5.4.1.2"},
			{ID: "C2", Description: "Les spécifications du produit fini ne sont pas altérées par le nouveau grade d'API", LegalRef: "LD-DEPP07 Section 5.4.1.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05", "DOC-06", "DOC-11"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.M.S.1", Title: "Agrément et BPF du nouveau fabricant", Module: 1, Level: ImpactDirectRed, Reason: "Preuve GMP du fabricant"},
			{SectionCode: "1.M.S.3", Title: "CEP EDQM et lettre d'accès nominative", Module: 1, Level: ImpactDirectRed, Reason: "CEP et mandat d'accès"},
			{SectionCode: "3.2.S.2", Title: "Procédé et contrôles SA", Module: 3, Level: ImpactDirectRed, Reason: "Remplacement dossier SA"},
			{SectionCode: "3.2.P.5", Title: "Contrôle du produit fini", Module: 3, Level: ImpactRippleOrange, Reason: "Analyse comparative 3 lots"},
		},
	},

	// =========================================================================
	// III. MODIFICATIONS QUALITATIVES SUBSTANCE ACTIVE 3.2.S (Rubriques 7 à 16)
	// =========================================================================
	{
		Code:                       "7.a.1",
		Number:                     7,
		SubCode:                    "7.a.1",
		Title:                      "Remplacement ou ajout d'un site de contrôle analytique de la SA (méthodes pharmacopée)",
		Category:                   CategorySubstanceActive32S,
		CategoryLabel:              "III. Qualité Substance Active 3.2.S",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si les méthodes d'essai ne sont pas validées ou sont non-pharmacopée",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le site dispose d'un certificat BPF valide pour le contrôle analytique", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
			{ID: "C2", Description: "La méthode d'essai est officiellement reconnue à la pharmacopée ou transfert validé", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-06", "DOC-11"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.H.8", Title: "Liste des sites", Module: 1, Level: ImpactDirectRed, Reason: "Ajout du laboratoire de contrôle"},
			{SectionCode: "3.2.S.4.3", Title: "Validation des méthodes SA", Module: 3, Level: ImpactRippleOrange, Reason: "Rapport de transfert de méthode"},
		},
	},
	{
		Code:                       "7.b.1",
		Number:                     7,
		SubCode:                    "7.b.1",
		Title:                      "Nouveau site pour la production du matériau de départ de la SA (Starting Material)",
		Category:                   CategorySubstanceActive32S,
		CategoryLabel:              "III. Qualité Substance Active 3.2.S",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si changement des spécifications ou du profil d'impuretés du starting material",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Les spécifications d'acceptation du matériau de départ sont identiques ou resserrées", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
			{ID: "C2", Description: "Aucune impureté nouvelle critique n'est introduite dans la chaîne de synthèse", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05", "DOC-06"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.2.2", Title: "Description du procédé de synthèse", Module: 3, Level: ImpactDirectRed, Reason: "Spécifications starting material"},
			{SectionCode: "3.2.S.3.2", Title: "Impuretés SA", Module: 3, Level: ImpactRippleOrange, Reason: "Vérification profil impuretés"},
		},
	},
	{
		Code:                       "7.d.1",
		Number:                     7,
		SubCode:                    "7.d.1",
		Title:                      "Nouveau fabricant pour la production complète de la Substance Active (avec CEP ou ASMF)",
		Category:                   CategorySubstanceActive32S,
		CategoryLabel:              "III. Qualité Substance Active 3.2.S",
		DefaultClassification:      ClassMmod,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si absence de CEP ou changement substantiel de profil de pureté / impuretés génotoxiques",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "La voie de synthèse et les spécifications de libération sont équivalentes ou supérieures", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
			{ID: "C2", Description: "Données de stabilité comparative Zone IVb engagées sur 3 lots de produit fini", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
			{ID: "C3", Description: "Certificat BPF valide délivré par une autorité reconnue ou l'ANPP", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05", "DOC-06", "DOC-11"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.M.S.1", Title: "Agrément et BPF Fabricant SA", Module: 1, Level: ImpactDirectRed, Reason: "Certificat BPF nouveau fabricant"},
			{SectionCode: "3.2.S.2", Title: "Procédé de fabrication SA", Module: 3, Level: ImpactDirectRed, Reason: "Dossier de synthèse complet"},
			{SectionCode: "3.2.P.5", Title: "Contrôle Produit Fini", Module: 3, Level: ImpactRippleOrange, Reason: "CoA comparatifs 3 lots"},
			{SectionCode: "3.2.P.8", Title: "Stabilité Produit Fini Zone IVb", Module: 3, Level: ImpactRippleOrange, Reason: "Engagement stabilité temps réel"},
		},
	},
	{
		Code:                       "8.a",
		Number:                     8,
		SubCode:                    "8.a",
		Title:                      "Ajout ou changement d'un bloc de fabrication ou unité de production sur un site API approuvé",
		Category:                   CategorySubstanceActive32S,
		CategoryLabel:              "III. Qualité Substance Active 3.2.S",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si le nouveau bloc modifie la chaîne de confinement ou les contrôles d'ambiance",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le bloc est situé sur le site déjà approuvé et couvert par le certificat BPF", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-06", "DOC-11"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.2.1", Title: "Fabricants SA", Module: 3, Level: ImpactDirectRed, Reason: "Mention du nouveau bloc"},
		},
	},
	{
		Code:                       "9.a.1",
		Number:                     9,
		SubCode:                    "9.a.1",
		Title:                      "Changement mineur dans le procédé de synthèse de la SA (sans altération du profil d'impuretés)",
		Category:                   CategorySubstanceActive32S,
		CategoryLabel:              "III. Qualité Substance Active 3.2.S",
		DefaultClassification:      ClassMmod,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si altération du profil d'impuretés ou introduction d'un nouveau solvant toxique",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le profil de pureté énantiomérique et les impuretés restent dans les seuils approuvés", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
			{ID: "C2", Description: "Aucun solvant de classe 1 ou nouveau réactif génotoxique utilisé", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.2.2", Title: "Procédé de fabrication SA", Module: 3, Level: ImpactDirectRed, Reason: "Description du procédé optimisé"},
			{SectionCode: "3.2.S.2.5", Title: "Validation de procédé", Module: 3, Level: ImpactDirectRed, Reason: "Rapport de validation"},
		},
	},
	{
		Code:                    "9.a.2",
		Number:                  9,
		SubCode:                 "9.a.2",
		Title:                   "Changement important dans le procédé de synthèse de la SA (altération de voie de synthèse)",
		Category:                CategorySubstanceActive32S,
		CategoryLabel:           "III. Qualité Substance Active 3.2.S",
		DefaultClassification:   ClassMmaj,
		EscalatedClassification: ClassMmaj,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Dossier complet de requalification analytique et stabilité comparative", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05", "DOC-07"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.2", Title: "Procédé de fabrication", Module: 3, Level: ImpactDirectRed, Reason: "Réécriture complète Module 3.2.S"},
			{SectionCode: "3.2.P.2", Title: "Développement pharmaceutique", Module: 3, Level: ImpactRippleOrange, Reason: "Impact sur formulation"},
		},
	},
	{
		Code:                       "10.a.1",
		Number:                     10,
		SubCode:                    "10.a.1",
		Title:                      "Augmentation de la taille de lot de la SA (jusqu'à 10 fois la taille approuvée)",
		Category:                   CategorySubstanceActive32S,
		CategoryLabel:              "III. Qualité Substance Active 3.2.S",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si changement d'équipement majeur ou plus de 10 fois",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "L'augmentation ne dépasse pas un facteur 10 par rapport au lot pilote approuvé", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
			{ID: "C2", Description: "Les paramètres critiques de procédé demeurent inchangés", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.2.2", Title: "Taille de lot SA", Module: 3, Level: ImpactDirectRed, Reason: "Actualisation échelle de synthèse"},
		},
	},
	{
		Code:                       "10.c.1",
		Number:                     10,
		SubCode:                    "10.c.1",
		Title:                      "Augmentation supérieure à 10 fois la taille de lot approuvée de la SA",
		Category:                   CategorySubstanceActive32S,
		CategoryLabel:              "III. Qualité Substance Active 3.2.S",
		DefaultClassification:      ClassMmod,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si déviation sur la cinétique ou validation de procédé non conforme",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Validation industrielle complète sur au moins 3 lots d'échelle commerciale", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.2.2", Title: "Taille de lot", Module: 3, Level: ImpactDirectRed, Reason: "Validation échelle industrielle"},
			{SectionCode: "3.2.S.2.5", Title: "Validation de procédé", Module: 3, Level: ImpactDirectRed, Reason: "Rapport 3 lots industriels"},
		},
	},
	{
		Code:                    "11.a",
		Number:                  11,
		SubCode:                 "11.a",
		Title:                   "Resserrement des limites des contrôles en cours de fabrication durant la synthèse de l'API",
		Category:                CategorySubstanceActive32S,
		CategoryLabel:           "III. Qualité Substance Active 3.2.S",
		DefaultClassification:   ClassMmin,
		EscalatedClassification: ClassMmin,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le changement vise à garantir une plus grande constance de qualité de l'API", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.2.4", Title: "Contrôles en cours SA", Module: 3, Level: ImpactDirectRed, Reason: "Mise à jour des critères resserrés"},
		},
	},
	{
		Code:                    "12.e",
		Number:                  12,
		SubCode:                 "12.e",
		Title:                   "Resserrement d'un critère d'acceptation dans les spécifications de la SA",
		Category:                CategorySubstanceActive32S,
		CategoryLabel:           "III. Qualité Substance Active 3.2.S",
		DefaultClassification:   ClassMmin,
		EscalatedClassification: ClassMmin,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le resserrement est conforme aux données historiques de libération", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.4.1", Title: "Spécifications SA", Module: 3, Level: ImpactDirectRed, Reason: "Mise à jour fiche de spécification"},
		},
	},
	{
		Code:                    "12.f.1",
		Number:                  12,
		SubCode:                 "12.f.1",
		Title:                   "Élargissement d'un critère d'acceptation des spécifications de la SA",
		Category:                CategorySubstanceActive32S,
		CategoryLabel:           "III. Qualité Substance Active 3.2.S",
		DefaultClassification:   ClassMmaj,
		EscalatedClassification: ClassMmaj,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Justification toxicologique et clinique de l'innocuité de l'élargissement", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.4.1", Title: "Spécifications SA", Module: 3, Level: ImpactDirectRed, Reason: "Modification majeure spécifications"},
			{SectionCode: "3.2.S.4.5", Title: "Justification spécifications", Module: 3, Level: ImpactDirectRed, Reason: "Dossier toxicologique M4"},
		},
	},
	{
		Code:                       "13.b",
		Number:                     13,
		SubCode:                    "13.b",
		Title:                      "Passage d'une méthode analytique interne à une méthode de Pharmacopée officielle reconnue pour la SA",
		Category:                   CategorySubstanceActive32S,
		CategoryLabel:              "III. Qualité Substance Active 3.2.S",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si la méthode pharmacopée nécessite une adaptation pour la matrice considérée",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "La monographie est applicable sans modification à la substance active considérée", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.4.2", Title: "Procédures analytiques SA", Module: 3, Level: ImpactDirectRed, Reason: "Référence pharmacopée"},
		},
	},
	{
		Code:                       "14.a",
		Number:                     14,
		SubCode:                    "14.a",
		Title:                      "Changement de conditionnement primaire de la SA (matériau équivalent ou plus protecteur)",
		Category:                   CategorySubstanceActive32S,
		CategoryLabel:              "III. Qualité Substance Active 3.2.S",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si substance active liquide, stérile ou instable",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le matériau nouveau offre une protection barrière équivalente ou supérieure contre l'humidité et la lumière", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.6", Title: "Système de fermeture SA", Module: 3, Level: ImpactDirectRed, Reason: "Description nouveau conditionnement"},
		},
	},
	{
		Code:                       "15.a",
		Number:                     15,
		SubCode:                    "15.a",
		Title:                      "Réduction de la période de recontrôle ou durée de conservation de la SA",
		Category:                   CategorySubstanceActive32S,
		CategoryLabel:              "III. Qualité Substance Active 3.2.S",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si causée par un profil d'instabilité inattendu en cours de stockage",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Les spécifications de conformité sont respectées sur la période raccourcie", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.7", Title: "Stabilité SA", Module: 3, Level: ImpactDirectRed, Reason: "Raccourcissement durée de recontrôle"},
		},
	},
	{
		Code:                    "16.a",
		Number:                  16,
		SubCode:                 "16.a",
		Title:                   "Conditions de stockage plus restrictives pour la SA sur l'étiquette",
		Category:                CategorySubstanceActive32S,
		CategoryLabel:           "III. Qualité Substance Active 3.2.S",
		DefaultClassification:   ClassMmin,
		EscalatedClassification: ClassMmod,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Mesure de précaution additionnelle pour préserver la stabilité de la SA", LegalRef: "LD-DEPP07 Section 5.4.2.1"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.S.7", Title: "Stabilité SA", Module: 3, Level: ImpactDirectRed, Reason: "Conditions de conservation SA"},
		},
	},

	// =========================================================================
	// IV. MODIFICATIONS QUALITATIVES PRODUIT FINI 3.2.P (Rubriques 17 à 34)
	// =========================================================================
	{
		Code:                       "17.a",
		Number:                     17,
		SubCode:                    "17.a",
		Title:                      "Système d'aromatisation ou coloration du PF (ajustement de teneur sans impact thérapeutique)",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si dissolution modifiée ou allergène potentiel ajouté",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le profil de dissolution in vitro est superposable (f2 >= 50)", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Colorant conforme à la réglementation des additifs pharmaceutiques", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05", "DOC-12"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.1", Title: "Composition produit fini", Module: 3, Level: ImpactDirectRed, Reason: "Formule unitaire excipient"},
			{SectionCode: "1.P.2", Title: "Notice et étiquetage", Module: 1, Level: ImpactRippleOrange, Reason: "Mention des colorants/arômes"},
		},
	},
	{
		Code:                       "18.a",
		Number:                     18,
		SubCode:                    "18.a",
		Title:                      "Masse d'enrobage ou capsules vides pour formes orales solides à libération immédiate",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si altération de la cinétique de dissolution",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "La formule de base du comprimé nu ou contenu reste inchangée", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Profils de dissolution multipoints conformes (f2 >= 50)", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.1", Title: "Composition PF", Module: 3, Level: ImpactDirectRed, Reason: "Masse unitaire capsule/enrobage"},
		},
	},
	{
		Code:                       "19.a",
		Number:                     19,
		SubCode:                    "19.a",
		Title:                      "Remplacement d'un excipient par un excipient comparable à concentration similaire (formes orales solides)",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmod,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si l'excipient a un impact sur la biodisponibilité (ex: agent tensio-actif, liant critique) ou si f2 < 50",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "L'excipient remplaçant remplit la même fonction technologique avec spécifications équivalentes", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Profils de dissolution comparatifs multipoints à pH 1.2, 4.5 et 6.8 démontrant la similarité (f2 >= 50)", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C3", Description: "Données de stabilité en Zone IVb engagées sur 3 lots industriels", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05", "DOC-12"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.1", Title: "Composition qualitative et quantitative", Module: 3, Level: ImpactDirectRed, Reason: "Formule cadre révisée"},
			{SectionCode: "3.2.P.2", Title: "Développement pharmaceutique", Module: 3, Level: ImpactDirectRed, Reason: "Profils de dissolution comparatifs"},
			{SectionCode: "3.2.P.8", Title: "Stabilité Zone IVb", Module: 3, Level: ImpactDirectRed, Reason: "Rapport de stabilité 3 lots"},
			{SectionCode: "1.P.2", Title: "Notice patient FR/AR", Module: 1, Level: ImpactRippleOrange, Reason: "Mise à jour liste des excipients"},
		},
	},
	{
		Code:                       "20.a",
		Number:                     20,
		SubCode:                    "20.a",
		Title:                      "Nouveau site ou remplacement pour le conditionnement secondaire du produit fini",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si absence de certificat BPF valide pour le conditionnement secondaire",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le site dispose d'un certificat BPF en vigueur pour le conditionnement secondaire", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-06", "DOC-11"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.H.8", Title: "Liste des sites intervenants", Module: 1, Level: ImpactDirectRed, Reason: "Ajout du conditionneur secondaire"},
			{SectionCode: "3.2.P.3.1", Title: "Fabricants PF", Module: 3, Level: ImpactDirectRed, Reason: "Tableau des sites"},
		},
	},
	{
		Code:                       "20.b",
		Number:                     20,
		SubCode:                    "20.b",
		Title:                      "Nouveau site pour le conditionnement primaire du PF (formes solides orales non stériles)",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si conditionnement stérile ou si matériau d'emballage primaire modifié",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Forme pharmaceutique non stérile", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Les spécifications de l'emballage primaire sont rigoureusement identiques", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-06", "DOC-11"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.H.8", Title: "Liste des sites intervenants", Module: 1, Level: ImpactDirectRed, Reason: "Ajout du site de blister / flaconnage"},
			{SectionCode: "3.2.P.3.1", Title: "Fabricants PF", Module: 3, Level: ImpactDirectRed, Reason: "Mise à jour sites"},
		},
	},
	{
		Code:                       "20.c",
		Number:                     20,
		SubCode:                    "20.c",
		Title:                      "Nouveau site de fabrication du Produit Fini (procédé conventionnel non complexe)",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmod,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si procédé complexe (libération modifiée, stérile, biologique) ou absence de validation 3 lots",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le procédé de fabrication est conventionnel (mélange, compression, remplissage)", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Rapport de validation industrielle concluant sur 3 lots consécutifs", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C3", Description: "Profils de dissolution comparatifs superposables avec le site d'origine", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05", "DOC-06", "DOC-11"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.H.8", Title: "Liste des sites intervenants", Module: 1, Level: ImpactDirectRed, Reason: "Nouveau site formulateur"},
			{SectionCode: "1.P.8", Title: "Engagement de validation 3 lots", Module: 1, Level: ImpactDirectRed, Reason: "Lettre d'engagement obligatoire"},
			{SectionCode: "3.2.P.3.5", Title: "Validation de procédé PF", Module: 3, Level: ImpactDirectRed, Reason: "Dossiers de lots industriels"},
			{SectionCode: "3.2.P.8", Title: "Stabilité Zone IVb", Module: 3, Level: ImpactRippleOrange, Reason: "Suivi stabilité 3 lots"},
		},
	},
	{
		Code:                       "21.a",
		Number:                     21,
		SubCode:                    "21.a",
		Title:                      "Remplacement ou ajout d'un site de contrôle analytique du Produit Fini",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si transfert de méthode non validé ou méthode non pharmacopée",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le laboratoire d'analyse détient une autorisation d'établissement BPF valide", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Le rapport de transfert analytique atteste de la comparabilité des résultats", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-06", "DOC-11"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "1.H.8", Title: "Liste des sites intervenants", Module: 1, Level: ImpactDirectRed, Reason: "Ajout du laboratoire de contrôle"},
			{SectionCode: "3.2.P.5.3", Title: "Validation des méthodes d'analyse", Module: 3, Level: ImpactRippleOrange, Reason: "Rapport de transfert de méthode"},
		},
	},
	{
		Code:                       "22.a",
		Number:                     22,
		SubCode:                    "22.a",
		Title:                      "Augmentation de la taille de lot du Produit Fini (jusqu'à 10 fois la taille approuvée)",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si changement d'équipement critique ou taille supérieure à 10 fois",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "L'augmentation est au maximum d'un facteur 10 par rapport à l'échelle enregistrée", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Maintien des principes de fonctionnement des équipements de production", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.3.2", Title: "Formule par lot", Module: 3, Level: ImpactDirectRed, Reason: "Calcul nouvelle formule par lot"},
		},
	},
	{
		Code:                       "22.c",
		Number:                     22,
		SubCode:                    "22.c",
		Title:                      "Augmentation supérieure à 10 fois la taille de lot du Produit Fini",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmod,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si procédé complexe ou cinétique de séchage/granulation altérée",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Validation sur au moins 3 lots d'échelle industrielle complète", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.3.2", Title: "Formule par lot", Module: 3, Level: ImpactDirectRed, Reason: "Formule industrielle"},
			{SectionCode: "3.2.P.3.5", Title: "Validation de procédé", Module: 3, Level: ImpactDirectRed, Reason: "Rapport 3 lots"},
		},
	},
	{
		Code:                       "23",
		Number:                     23,
		SubCode:                    "23",
		Title:                      "Changement dans le procédé de fabrication du Produit Fini (sans changement de principe)",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmod,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si changement fondamental de principe de procédé (ex: voie humide vers compression directe)",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le principe fondamental de fabrication reste identique", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Les spécifications de libération et la cinétique de dissolution sont maintenues", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.3.3", Title: "Description du procédé", Module: 3, Level: ImpactDirectRed, Reason: "Description du nouveau procédé"},
			{SectionCode: "3.2.P.3.5", Title: "Validation de procédé", Module: 3, Level: ImpactDirectRed, Reason: "Rapport de requalification"},
		},
	},
	{
		Code:                    "24.a",
		Number:                  24,
		SubCode:                 "24.a",
		Title:                   "Resserrement des contrôles en cours de fabrication du Produit Fini (IPC)",
		Category:                CategoryProduitFini32P,
		CategoryLabel:           "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:   ClassMmin,
		EscalatedClassification: ClassMmin,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Amélioration de la maîtrise statistique du procédé de fabrication", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.3.4", Title: "Contrôles en cours PF", Module: 3, Level: ImpactDirectRed, Reason: "Mise à jour limites IPC"},
		},
	},
	{
		Code:                       "25",
		Number:                     25,
		SubCode:                    "25",
		Title:                      "Remplacement d'un excipient à risque EST par un excipient d'origine végétale ou synthétique",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si altération des profils de libération ou dissolution",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "L'excipient remplaçant élimine tout risque d'encéphalopathie spongiforme transmissible", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Profils de dissolution superposables (f2 >= 50)", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.1", Title: "Composition", Module: 3, Level: ImpactDirectRed, Reason: "Remplacement excipient EST"},
			{SectionCode: "3.2.P.4.5", Title: "Excipients d'origine humaine ou animale", Module: 3, Level: ImpactDirectRed, Reason: "Attestation sans EST"},
		},
	},
	{
		Code:                    "26.c",
		Number:                  26,
		SubCode:                 "26.c",
		Title:                   "Resserrement des spécifications d'un excipient",
		Category:                CategoryProduitFini32P,
		CategoryLabel:           "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:   ClassMmin,
		EscalatedClassification: ClassMmin,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le resserrement améliore la pureté et la sécurité de l'excipient", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.4.1", Title: "Spécifications des excipients", Module: 3, Level: ImpactDirectRed, Reason: "Fiche technique révisée"},
		},
	},
	{
		Code:                       "27",
		Number:                     27,
		SubCode:                    "27",
		Title:                      "Mise à jour des spécifications du Produit Fini suite à révision d'une monographie de Pharmacopée",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si la méthode adaptée diverge de la monographie officielle",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Alignement strict sur la dernière édition en vigueur de la pharmacopée reconnue", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.5.1", Title: "Spécifications PF", Module: 3, Level: ImpactDirectRed, Reason: "Mise à jour monographie"},
		},
	},
	{
		Code:                    "28.c",
		Number:                  28,
		SubCode:                 "28.c",
		Title:                   "Resserrement d'un critère d'acceptation dans les spécifications du Produit Fini",
		Category:                CategoryProduitFini32P,
		CategoryLabel:           "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:   ClassMmin,
		EscalatedClassification: ClassMmin,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le resserrement est basé sur l'historique de libération industrielle", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.5.1", Title: "Spécifications PF", Module: 3, Level: ImpactDirectRed, Reason: "Critères resserrés"},
		},
	},
	{
		Code:                    "29.b",
		Number:                  29,
		SubCode:                 "29.b",
		Title:                   "Ajout d'une procédure analytique complémentaire pour le contrôle du Produit Fini",
		Category:                CategoryProduitFini32P,
		CategoryLabel:           "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:   ClassMmin,
		EscalatedClassification: ClassMmod,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Rapport complet de validation analytique selon les exigences ICH Q2(R1)", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.5.2", Title: "Procédures analytiques PF", Module: 3, Level: ImpactDirectRed, Reason: "Ajout protocole analytique"},
			{SectionCode: "3.2.P.5.3", Title: "Validation des méthodes", Module: 3, Level: ImpactDirectRed, Reason: "Rapport ICH Q2"},
		},
	},
	{
		Code:                       "30.a",
		Number:                     30,
		SubCode:                    "30.a",
		Title:                      "Changement ou ajout de type de contenant primaire (formes pharmaceutiques non stériles)",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmod,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si contenant stérile ou si interaction contenant-contenu observée",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Forme non stérile", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Étude d'interaction contenant-contenu démontrant la compatibilité", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C3", Description: "Données de stabilité Zone IVb sur 3 lots conditionnés dans le nouveau contenant", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05", "DOC-12"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.7", Title: "Système de fermeture PF", Module: 3, Level: ImpactDirectRed, Reason: "Description du nouveau contenant"},
			{SectionCode: "3.2.P.8", Title: "Stabilité Zone IVb", Module: 3, Level: ImpactDirectRed, Reason: "Données stabilité nouveau packaging"},
			{SectionCode: "1.P.2", Title: "Notice et maquettes", Module: 1, Level: ImpactRippleOrange, Reason: "Mise à jour présentation emballage"},
		},
	},
	{
		Code:                    "30.b",
		Number:                  30,
		SubCode:                 "30.b",
		Title:                   "Changement de conditionnement primaire pour médicaments injectables ou stériles",
		Category:                CategoryProduitFini32P,
		CategoryLabel:           "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:   ClassMmaj,
		EscalatedClassification: ClassMmaj,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Validation complète de l'intégrité de la fermeture hermétique (CCI)", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "Validation du procédé de stérilisation ou de remplissage aseptique", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05", "DOC-11", "DOC-12"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.7", Title: "Système de fermeture", Module: 3, Level: ImpactDirectRed, Reason: "Validation packaging stérile"},
			{SectionCode: "3.2.P.3.5", Title: "Validation procédé aseptique", Module: 3, Level: ImpactDirectRed, Reason: "Media fill et qualification"},
			{SectionCode: "3.2.P.8", Title: "Stabilités Zone IVb", Module: 3, Level: ImpactDirectRed, Reason: "Données temps réel 3 lots"},
		},
	},
	{
		Code:                       "31.a",
		Number:                     31,
		SubCode:                    "31.a",
		Title:                      "Composition du conditionnement primaire (formes solides : blister vers autre blister équivalent)",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmin,
		EscalatedClassification:    ClassMmod,
		EscalationConditionSummary: "Si la perméabilité à l'humidité ou oxygène est accrue",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Le nouveau film blister présente une perméabilité à la vapeur d'eau équivalente ou inférieure (ex: PVC/PVDC)", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.7", Title: "Système de fermeture", Module: 3, Level: ImpactDirectRed, Reason: "Spécifications nouveau complexe blister"},
		},
	},
	{
		Code:                       "32.b",
		Number:                     32,
		SubCode:                    "32.b",
		Title:                      "Extension de la durée de conservation du Produit Fini pour la vente",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmod,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si absence de données en temps réel complètes sous conditions Zone IVb (30°C / 75% HR)",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Données de stabilité en temps réel complètes couvrant la durée sollicitée sous conditions climatiques Zone IVb", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
			{ID: "C2", Description: "L'étude porte sur 3 lots industriels complets dans leur conditionnement commercial définitif", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05", "DOC-12"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.8.1", Title: "Résumé stabilité", Module: 3, Level: ImpactDirectRed, Reason: "Nouvelle durée de conservation"},
			{SectionCode: "3.2.P.8.3", Title: "Données de stabilité", Module: 3, Level: ImpactDirectRed, Reason: "Rapport complet Zone IVb"},
			{SectionCode: "1.P.1", Title: "RCP", Module: 1, Level: ImpactDirectRed, Reason: "Rubrique 6.3 durée de conservation"},
			{SectionCode: "1.P.2", Title: "Notice et étiquetage", Module: 1, Level: ImpactDirectRed, Reason: "Date de péremption étuis"},
		},
	},
	{
		Code:                    "33.a",
		Number:                  33,
		SubCode:                 "33.a",
		Title:                   "Réduction de la durée de conservation après première ouverture ou reconstitution",
		Category:                CategoryProduitFini32P,
		CategoryLabel:           "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:   ClassMmod,
		EscalatedClassification: ClassMmod,
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Mesure de précaution pour prévenir tout risque d'instabilité en cours d'utilisation", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-05", "DOC-12"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.8", Title: "Stabilité en cours d'utilisation", Module: 3, Level: ImpactDirectRed, Reason: "Rapport stabilité in-use"},
			{SectionCode: "1.P.1", Title: "RCP", Module: 1, Level: ImpactDirectRed, Reason: "Section 6.3 du RCP"},
			{SectionCode: "1.P.2", Title: "Notice patient", Module: 1, Level: ImpactDirectRed, Reason: "Instructions de conservation"},
		},
	},
	{
		Code:                       "34",
		Number:                     34,
		SubCode:                    "34",
		Title:                      "Changement des conditions de stockage du Produit Fini ou en cours d'utilisation",
		Category:                   CategoryProduitFini32P,
		CategoryLabel:              "IV. Qualité Produit Fini 3.2.P",
		DefaultClassification:      ClassMmod,
		EscalatedClassification:    ClassMmaj,
		EscalationConditionSummary: "Si passage à une condition plus contraignante (ex: température ambiante vers chaîne du froid)",
		Conditions: []VariationCondition{
			{ID: "C1", Description: "Rapport de stabilité Zone IVb démontrant la dégradation sous les anciennes conditions", LegalRef: "LD-DEPP07 Section 5.4.2.2"},
		},
		RequiredDocuments: []string{"DOC-01", "DOC-02", "DOC-03", "DOC-05", "DOC-12"},
		ImpactedSections: []CTDSectionImpact{
			{SectionCode: "3.2.P.8", Title: "Conditions de conservation", Module: 3, Level: ImpactDirectRed, Reason: "Nouvelles conditions de stockage"},
			{SectionCode: "1.P.1", Title: "RCP", Module: 1, Level: ImpactDirectRed, Reason: "Section 6.4 du RCP"},
			{SectionCode: "1.P.2", Title: "Notice et maquettes", Module: 1, Level: ImpactDirectRed, Reason: "Pictogrammes et mentions de conservation"},
		},
	},
}

// MasterVariationDocumentsList définit les 12 pièces obligatoires de la LD-DEPP07 Section 5.3
var MasterVariationDocumentsList = []struct {
	Index       int
	Code        string
	Title       string
	Description string
	IsMandatory bool
}{
	{1, "DOC-01-LETTRE-DEMANDE", "Lettre de demande officielle de modification", "Courrier officiel signé et cacheté par le Pharmacien Directeur Technique", true},
	{2, "DOC-02-TABLEAU-COMPARATIF", "Tableau comparatif standardisé (ES-REG-04)", "État actuel approuvé vs Situation proposée avec justifications", true},
	{3, "DOC-03-JUSTIF-TECHNIQUE", "Justificatif technico-réglementaire", "Mémoire technique motivant le changement et analyse de risque", true},
	{4, "DOC-04-BORDEREAU-BV2", "Bordereau de versement BV2 & Quittance", "Quittance bancaire du Trésor Public conforme Décret exécutif 21-82", true},
	{5, "DOC-05-SECTIONS-CTD-REVISEES", "Sections révisées du dossier CTD", "Remplacement page à page des sections concernées", true},
	{6, "DOC-06-LISTE-SITES-VISEE", "Liste actualisée des intervenants (1.H.8)", "Cartographie de tous les sites (SA, PF, conditionneurs, contrôleurs) visée PDT", true},
	{7, "DOC-07-ATTESTATION-NON-MODIF", "Attestation formelle de non-modification", "Déclaration sur l'honneur du PDT de non-altération des autres sections", true},
	{8, "DOC-08-COPIE-DE-VIGUEUR", "Copie de la Décision d'Enregistrement (DE)", "Décision d'enregistrement ministérielle originale en cours de validité", true},
	{9, "DOC-09-ACCUSES-DEPOTS-EN-COURS", "Accusés de dépôt des procédures en cours", "Bordereaux des autres variations ou renouvellement en cours d'instruction", true},
	{10, "DOC-10-APPROBATION-ORIGINE", "Approbation pays d'origine (Import)", "Preuve d'approbation par les autorités sanitaires du pays d'origine", true},
	{11, "DOC-11-CERTIFICATS-BPF-SITES", "Autorisations des sites et BPF/GMP", "Certificats BPF/BPL en vigueur pour les sites et opérations modifiées", true},
	{12, "DOC-12-INFOS-PRODUIT-REVISEES", "Informations produit révisées", "RCP (1.P.1), Notice bilingue FR/AR (1.P.2) et maquettes couleurs visées PDT", true},
}

// FindRubriqueDefinition recherche la définition d'une rubrique par son code
func FindRubriqueDefinition(code string) *RubriqueDefinition {
	codeNorm := strings.TrimSpace(code)
	for i := range MasterRubriquesRegistry {
		if MasterRubriquesRegistry[i].Code == codeNorm {
			return &MasterRubriquesRegistry[i]
		}
	}
	return nil
}

// EvaluateRubriqueItem évalue si les conditions sont satisfaites et applique la règle d'escalade automatique
func EvaluateRubriqueItem(rubriqueCode string, conditionsChecked map[string]bool) (nominal VariationClassification, effective VariationClassification, isEscalated bool, reason string) {
	rub := FindRubriqueDefinition(rubriqueCode)
	if rub == nil {
		return ClassMmin, ClassMmin, false, "Rubrique inconnue"
	}

	nominal = rub.DefaultClassification
	effective = nominal
	isEscalated = false

	if len(rub.Conditions) == 0 {
		return nominal, effective, false, ""
	}

	unmetConditions := make([]string, 0)
	for _, c := range rub.Conditions {
		if checked, ok := conditionsChecked[c.ID]; !ok || !checked {
			unmetConditions = append(unmetConditions, c.ID+" ("+c.Description+")")
		}
	}

	if len(unmetConditions) > 0 {
		isEscalated = true
		effective = rub.EscalatedClassification
		reason = "Escalade automatique de " + string(nominal) + " vers " + string(effective) + " : condition(s) non remplie(s) : " + strings.Join(unmetConditions, " ; ")
	}

	return nominal, effective, isEscalated, reason
}

// CalculateOverallClassification calcule le niveau global pour une ou plusieurs rubriques (Grouping)
func CalculateOverallClassification(items []VariationItem) (overall VariationClassification, hasEscalation bool, highestItem string) {
	if len(items) == 0 {
		return ClassMmin, false, ""
	}

	overall = ClassMmin
	hasEscalation = false

	for _, it := range items {
		if it.IsEscalated {
			hasEscalation = true
		}
		if it.EffectiveClassification.IsHigherThan(overall) {
			overall = it.EffectiveClassification
			highestItem = it.RubriqueCode
		}
	}
	return overall, hasEscalation, highestItem
}

// CalculateVariationFee calcule la redevance réglementaire Décret 21-82
func CalculateVariationFee(classification VariationClassification) float64 {
	switch classification {
	case ClassMmaj:
		return 115000.00 // Décret 21-82 : 115 000 DA
	case ClassMmod:
		return 80000.00 // Décret 21-82 : 80 000 DA
	case ClassMmin:
		return 40000.00 // Décret 21-82 : 40 000 DA
	default:
		return 40000.00
	}
}

// GenerateChangeImpactTree projette les sections CTD impactées par les rubriques choisies
func GenerateChangeImpactTree(rubriqueCodes []string) ChangeImpactTree {
	directMap := make(map[string]CTDSectionImpact)
	rippleMap := make(map[string]CTDSectionImpact)

	for _, code := range rubriqueCodes {
		rub := FindRubriqueDefinition(code)
		if rub == nil {
			continue
		}
		for _, sec := range rub.ImpactedSections {
			if sec.Level == ImpactDirectRed {
				directMap[sec.SectionCode] = sec
				delete(rippleMap, sec.SectionCode) // Rouge l'emporte sur orange
			} else if sec.Level == ImpactRippleOrange {
				if _, exists := directMap[sec.SectionCode]; !exists {
					rippleMap[sec.SectionCode] = sec
				}
			}
		}
	}

	// Sections CTD majeures candidates pour la préservation (vert)
	standardSections := []CTDSectionImpact{
		{SectionCode: "1.A.1", Title: "Récépissé de dépôt", Module: 1, Level: ImpactPreservedGreen, Reason: "Non altérée"},
		{SectionCode: "1.A.6", Title: "Agrément MIPH", Module: 1, Level: ImpactPreservedGreen, Reason: "Non altéré"},
		{SectionCode: "1.A.7", Title: "Décision d'exercice du PDT", Module: 1, Level: ImpactPreservedGreen, Reason: "Non altérée"},
		{SectionCode: "1.H.3", Title: "Certificat CPP OMS", Module: 1, Level: ImpactPreservedGreen, Reason: "Préservé"},
		{SectionCode: "1.H.4", Title: "Structure de prix PCSU / FOB", Module: 1, Level: ImpactPreservedGreen, Reason: "Prix inchangé"},
		{SectionCode: "1.M.P.3", Title: "Étude de bioéquivalence initiale", Module: 1, Level: ImpactPreservedGreen, Reason: "Dossier clinique préservé"},
		{SectionCode: "2.3.P", Title: "QOS Produit Fini non modifié", Module: 2, Level: ImpactPreservedGreen, Reason: "Sans impact direct"},
		{SectionCode: "3.2.P.6", Title: "Étalons de référence secondaires", Module: 3, Level: ImpactPreservedGreen, Reason: "Monographies inchangées"},
		{SectionCode: "4.1", Title: "Études précliniques", Module: 4, Level: ImpactPreservedGreen, Reason: "Non applicable aux variations qualité"},
		{SectionCode: "5.1", Title: "Rapports cliniques complets", Module: 5, Level: ImpactPreservedGreen, Reason: "Non applicable aux variations qualité"},
	}

	preservedList := make([]CTDSectionImpact, 0)
	for _, sec := range standardSections {
		if _, inDirect := directMap[sec.SectionCode]; inDirect {
			continue
		}
		if _, inRipple := rippleMap[sec.SectionCode]; inRipple {
			continue
		}
		preservedList = append(preservedList, sec)
	}

	directList := make([]CTDSectionImpact, 0, len(directMap))
	for _, v := range directMap {
		directList = append(directList, v)
	}

	rippleList := make([]CTDSectionImpact, 0, len(rippleMap))
	for _, v := range rippleMap {
		rippleList = append(rippleList, v)
	}

	return ChangeImpactTree{
		DirectRed:      directList,
		RippleOrange:   rippleList,
		PreservedGreen: preservedList,
		Stats: map[string]int{
			"direct_red_count":      len(directList),
			"ripple_orange_count":   len(rippleList),
			"preserved_green_count": len(preservedList),
			"total_impacted_count":  len(directList) + len(rippleList),
		},
	}
}
