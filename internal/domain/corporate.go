package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EstablishmentCategory représente la catégorie de l'établissement selon le Décret 21-82
type EstablishmentCategory string

const (
	CategoryFabrication  EstablishmentCategory = "FABRICATION"
	CategoryImportation  EstablishmentCategory = "IMPORTATION"
	CategoryDistribution EstablishmentCategory = "DISTRIBUTION"
	CategoryExploitation EstablishmentCategory = "EXPLOITATION"
)

// ApplicationNature représente la nature de la démarche ministérielle auprès du MIPH
type ApplicationNature string

const (
	NaturePrealableRealisation ApplicationNature = "PREALABLE_REALISATION"
	NatureOuvertureInitiale    ApplicationNature = "OUVERTURE_INITIALE"
	NatureSousTraitance        ApplicationNature = "SOUS_TRAITANCE"
	NatureExercicePDT          ApplicationNature = "EXERCICE_PDT"
	NatureExerciceAssistant    ApplicationNature = "EXERCICE_ASSISTANT"
	NatureRenouvellement       ApplicationNature = "RENOUVELLEMENT"
	NatureModifSubstantielle   ApplicationNature = "MODIF_SUBSTANTIELLE"
	NatureModifNonSubst        ApplicationNature = "MODIF_NON_SUBSTANTIELLE"
)

// ApplicationStatus représente le statut du dossier d'agrément
type ApplicationStatus string

const (
	StatusCorporateDraft            ApplicationStatus = "brouillon"
	StatusCorporateUnderReview      ApplicationStatus = "en_revue"
	StatusCorporatePDTSigned        ApplicationStatus = "vise_pdt"
	StatusCorporateTabadolSubmitted ApplicationStatus = "depose_tabadol"
	StatusCorporateBODeposited      ApplicationStatus = "depose_bo"
	StatusCorporateGranted          ApplicationStatus = "agrement_octroye"
	StatusCorporateRejected         ApplicationStatus = "rejete"
)

// CorporateApplication représente une demande d'agrément ministériel ou décision d'exercice
type CorporateApplication struct {
	ID                   uuid.UUID              `json:"id"`
	OrganizationID       uuid.UUID              `json:"organization_id"`
	EstablishmentID      *uuid.UUID             `json:"establishment_id,omitempty"`
	Category             EstablishmentCategory  `json:"category"`
	Nature               ApplicationNature      `json:"nature"`
	ApplicationNumber    string                 `json:"application_number"`
	InternalReference    string                 `json:"internal_reference,omitempty"`
	Title                string                 `json:"title"`
	Status               ApplicationStatus      `json:"status"`
	FormType             string                 `json:"form_type"`
	FormData             map[string]interface{} `json:"form_data"`
	CompletenessPercent  float64                `json:"completeness_percent"`
	PDTSignedBy          *uuid.UUID             `json:"pdt_signed_by,omitempty"`
	PDTSignedAt          *time.Time             `json:"pdt_signed_at,omitempty"`
	PDTSignatureSeal     string                 `json:"pdt_signature_seal,omitempty"`
	TabadolSubmissionRef string                 `json:"tabadol_submission_ref,omitempty"`
	TabadolSubmittedAt   *time.Time             `json:"tabadol_submitted_at,omitempty"`
	TabadolSubmissionNotes string               `json:"tabadol_submission_notes,omitempty"`
	MIPHReceiptRef       string                 `json:"miph_receipt_ref,omitempty"`
	MIPHReceiptDate      *string                `json:"miph_receipt_date,omitempty"`
	DecisionRef          string                 `json:"decision_ref,omitempty"`
	DecisionDate         *string                `json:"decision_date,omitempty"`
	LicenseExpiryDate    *string                `json:"license_expiry_date,omitempty"`
	CreatedBy            *uuid.UUID             `json:"created_by,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`

	// Champs enrichis
	EstablishmentName string                   `json:"establishment_name,omitempty"`
	ChecklistItems    []CorporateChecklistItem `json:"checklist_items,omitempty"`
}

// CorporateChecklistItem représente une exigence documentaire pour le dossier MIPH
type CorporateChecklistItem struct {
	ID               uuid.UUID  `json:"id"`
	ApplicationID    uuid.UUID  `json:"application_id"`
	ItemCode         string     `json:"item_code"`
	Title            string     `json:"title"`
	Section          string     `json:"section"`
	RequirementLevel string     `json:"requirement_level"` // E, C, F, S
	IsMandatory      bool       `json:"is_mandatory"`
	ConditionRule    string     `json:"condition_rule,omitempty"`
	SortOrder        int        `json:"sort_order"`
	GEDDocumentID    *uuid.UUID `json:"ged_document_id,omitempty"`
	Status           string     `json:"status"` // manquant, rattache, valide_qualite, rejete
	Notes            string     `json:"notes,omitempty"`
	UploadedAt       *time.Time `json:"uploaded_at,omitempty"`
	VerifiedBy       *uuid.UUID `json:"verified_by,omitempty"`
	VerifiedAt       *time.Time `json:"verified_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Champs enrichis GED
	DocumentCode string `json:"document_code,omitempty"`
	DocumentName string `json:"document_name,omitempty"`
	FileSHA256   string `json:"file_sha256,omitempty"`
}

// CorporateLicense représente un titre réglementaire, agrément ou décision d'exercice
type CorporateLicense struct {
	ID                    uuid.UUID  `json:"id"`
	OrganizationID        uuid.UUID  `json:"organization_id"`
	EstablishmentID       *uuid.UUID `json:"establishment_id,omitempty"`
	ApplicationID         *uuid.UUID `json:"application_id,omitempty"`
	LicenseType           string     `json:"license_type"` // AGREMENT_REALISATION, AGREMENT_OUVERTURE, DECISION_PDT, DECISION_ASSISTANT, AGREMENT_MODIF
	LicenseNumber         string     `json:"license_number"`
	DecisionDate          string     `json:"decision_date"`
	EffectiveDate         string     `json:"effective_date"`
	ExpiryDate            *string    `json:"expiry_date,omitempty"`
	PharmacistName        string     `json:"pharmacist_name,omitempty"`
	PharmacistOrderNumber string     `json:"pharmacist_order_number,omitempty"`
	Status                string     `json:"status"` // actif, a_renouveler_j180, alerte_j90, critique_j30, expire, modifie, abroge
	DocumentID            *uuid.UUID `json:"document_id,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`

	// Calculs d'échéances
	DaysUntilExpiry *int `json:"days_until_expiry,omitempty"`
}

// TabadolAuditReport représente le résultat de l'audit bloquant de pré-soumission Tabadol
type TabadolAuditReport struct {
	CanSubmit          bool     `json:"can_submit"`
	TotalItems         int      `json:"total_items"`
	AttachedItems      int      `json:"attached_items"`
	MissingMandatory   []string `json:"missing_mandatory"`
	CompletenessRatio  float64  `json:"completeness_ratio"`
	IsPDTSigned        bool     `json:"is_pdt_signed"`
	BlockingErrors     []string `json:"blocking_errors"`
	Warnings           []string `json:"warnings"`
	SubmissionManifest string   `json:"submission_manifest,omitempty"`
}

// GenerateCorporateChecklist génère le jeu d'exigences normé selon la catégorie et la nature de la demande
func GenerateCorporateChecklist(category EstablishmentCategory, nature ApplicationNature) []CorporateChecklistItem {
	var items []CorporateChecklistItem
	order := 10

	add := func(code, title, section, reqLevel string, mandatory bool, condition string) {
		items = append(items, CorporateChecklistItem{
			ItemCode:         code,
			Title:            title,
			Section:          section,
			RequirementLevel: reqLevel,
			IsMandatory:      mandatory,
			ConditionRule:    condition,
			SortOrder:        order,
			Status:           "manquant",
		})
		order += 10
	}

	switch category {
	case CategoryFabrication:
		switch nature {
		case NaturePrealableRealisation:
			add("FAB.PR.01", "Statuts de la société et registre du commerce avec code fabrication pharmaceutique", "ADMINISTRATIF", "E", true, "")
			add("FAB.PR.02", "Acte de propriété notarié ou contrat de bail notarié d'une durée minimale de 10 ans enregistré", "ADMINISTRATIF", "E", true, "")
			add("FAB.PR.03", "Certificat d'urbanisme et permis de construire", "ADMINISTRATIF", "E", true, "")
			add("FAB.PR.04", "Étude d'impact sur l'environnement approuvée par les services compétents", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.PR.05", "Plans de masse et plans de zonage (niveaux de propreté ISO 5 à ISO 8)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.PR.06", "Schémas des flux matières, personnels et déchets (élimination stricte des flux croisés)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.PR.07", "Liste des formes galéniques envisagées et procédés de fabrication prévisionnels", "ORGANISATION_BPF", "E", true, "")
			add("FAB.PR.08", "Étude technico-économique et plan de financement validé", "ADMINISTRATIF", "E", true, "")

		case NatureOuvertureInitiale:
			add("FAB.OUV.01", "Dossier administratif, registre de commerce, statuts et assurances RC professionnelle", "ADMINISTRATIF", "E", true, "")
			add("FAB.OUV.02", "Plans finaux as-built certifiés conformes", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.OUV.03", "Rapport de qualification des systèmes de traitement d'air CVC/HVAC (surpressions, filtration HEPA)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.OUV.04", "Rapport de qualification des boucles d'eau purifiée et d'eau PPI (Ph. Eur.)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.OUV.05", "Rapports de Qualification d'Installation (QI) et Opérationnelle (QO) des équipements", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.OUV.06", "Organigramme BPF nominatif et plan de formation continue du personnel", "ORGANISATION_BPF", "E", true, "")
			add("FAB.OUV.07", "Manuel Qualité (PQS) et procédures opératoires maîtresses de libération et d'hygiène", "ORGANISATION_BPF", "E", true, "")
			add("FAB.OUV.08", "Dossier d'agrément nominatif du Pharmacien Directeur Technique (Article 16 Décret 21-82)", "PHARMACIEN_DT", "E", true, "")

		case NatureSousTraitance:
			add("FAB.ST.01", "Contrat de sous-traitance et de façon BPF opposable définissant les obligations respectives", "ADMINISTRATIF", "E", true, "")
			add("FAB.ST.02", "Partage des responsabilités de contrôle analytique et de libération exclusive par le PDT", "ORGANISATION_BPF", "E", true, "")
			add("FAB.ST.03", "Rapport d'audit qualité contradictoire des installations du façonnier", "ORGANISATION_BPF", "E", true, "")
			add("FAB.ST.04", "Certificats de conformité BPF en vigueur des installations du fabricant tiers", "ORGANISATION_BPF", "E", true, "")

		case NatureExercicePDT:
			add("PDT.01", "Diplôme d'État de Docteur en Pharmacie ou équivalent légalement reconnu", "PHARMACIEN_DT", "E", true, "")
			add("PDT.02", "Attestation d'inscription au Conseil National de l'Ordre des Pharmaciens (Section B) < 3 mois", "PHARMACIEN_DT", "E", true, "")
			add("PDT.03", "Justificatif d'expérience professionnelle 3 ou 5 ans en fabrication/contrôle (attestations CNAS)", "PHARMACIEN_DT", "E", true, "")
			add("PDT.04", "Contrat de travail à durée indéterminée (CDI) exclusif à temps plein pour les fonctions de DT", "PHARMACIEN_DT", "E", true, "")
			add("PDT.05", "Extrait de casier judiciaire (Bulletin n° 3) datant de moins de 3 mois", "PHARMACIEN_DT", "E", true, "")
			add("PDT.06", "Certificat médical d'aptitude physique et mentale", "PHARMACIEN_DT", "E", true, "")
			add("PDT.07", "Déclaration sur l'honneur de non-cumul et engagement de respect des règles déontologiques et BPF", "PHARMACIEN_DT", "E", true, "")

		case NatureExerciceAssistant:
			add("AST.01", "Diplôme de pharmacien reconnu", "PHARMACIEN_DT", "E", true, "")
			add("AST.02", "Attestation d'inscription à l'Ordre National des Pharmaciens", "PHARMACIEN_DT", "E", true, "")
			add("AST.03", "Contrat de travail sous l'autorité directe du Pharmacien Directeur Technique", "PHARMACIEN_DT", "E", true, "")
			add("AST.04", "Fiche de poste et désignation précise du champ d'activité délégué", "PHARMACIEN_DT", "E", true, "")

		case NatureModifSubstantielle:
			add("MOD.SUB.01", "Rapport technico-réglementaire motivant la modification substantielle (Arrêté 22/09/2025)", "MODIFICATIONS", "E", true, "")
			add("MOD.SUB.02", "Plans comparatifs certifiés de l'état approuvé vs état modifié", "MODIFICATIONS", "E", true, "")
			add("MOD.SUB.03", "Rapports de requalification des utilités et locaux impactés (CVC/eau PPI)", "MODIFICATIONS", "E", true, "")
			add("MOD.SUB.04", "Analyse de risques BPF et plan de contrôle de changement (ICH Q9 / Q10)", "MODIFICATIONS", "E", true, "")

		default:
			add("FAB.GEN.01", "Dossier réglementaire d'information et fiches descriptives d'exploitation", "ADMINISTRATIF", "E", true, "")
		}

	case CategoryImportation:
		add("IMP.01", "Registre du commerce avec code d'activité importation pharmaceutique et statuts", "ADMINISTRATIF", "E", true, "")
		add("IMP.02", "Attestations de régularité fiscale et parafiscale (CNAS/CASNOS) en cours de validité", "ADMINISTRATIF", "E", true, "")
		add("IMP.03", "Programme prévisionnel d'importation annuel visé", "ADMINISTRATIF", "E", true, "")
		add("IMP.04", "Dossier et décision d'exercice nominative du Pharmacien Directeur Technique", "PHARMACIEN_DT", "E", true, "")
		add("IMP.05", "Descriptif des magasins de stockage et séparation stricte quarantaine/rejet sous clé", "CONDITIONS_TECHNIQUES", "E", true, "")
		add("IMP.06", "Rapports de qualification des chambres froides (+2°C à +8°C) et zones climatisées (+15°C à +25°C)", "CONDITIONS_TECHNIQUES", "E", true, "")
		add("IMP.07", "Procédure de contrôle documentaire et de libération des lots importés par le PDT", "ORGANISATION_BPF", "E", true, "")
		add("IMP.08", "Conventions d'assurance qualité avec les transitaires et transporteurs sous température dirigée", "CONDITIONS_TECHNIQUES", "E", true, "")

	case CategoryDistribution:
		add("DIS.01", "Plans et descriptif des entrepôts selon les Bonnes Pratiques de Distribution en Gros (BPDG)", "CONDITIONS_TECHNIQUES", "E", true, "")
		add("DIS.02", "Inventaire de la flotte logistique et certificats d'étalonnage des enregistreurs de température", "CONDITIONS_TECHNIQUES", "E", true, "")
		add("DIS.03", "Système informatique de traçabilité des lots et plan de rappel d'urgence", "ORGANISATION_BPF", "E", true, "")
		add("DIS.04", "Cahier des charges de distribution paraphé et engagement d'approvisionnement continu", "ADMINISTRATIF", "E", true, "")
		add("DIS.05", "Dossier et décision d'exercice du Pharmacien Directeur Technique", "PHARMACIEN_DT", "E", true, "")

	case CategoryExploitation:
		add("EXP.01", "Désignation de la personne responsable de la Pharmacovigilance (QPPV) et système d'alerte", "ORGANISATION_BPF", "E", true, "")
		add("EXP.02", "Procédure de contrôle et validation de l'information médicale/scientifique par le PDT", "ORGANISATION_BPF", "E", true, "")
		add("EXP.03", "Système de suivi de la traçabilité des lots sur le marché et procédure de rappel", "ORGANISATION_BPF", "E", true, "")
		add("EXP.04", "Cahier des conditions techniques à l'exploitation paraphé avec engagement déontologique", "ADMINISTRATIF", "E", true, "")
		add("EXP.05", "Dossier et décision d'exercice nominative du Pharmacien Directeur Technique", "PHARMACIEN_DT", "E", true, "")
	}

	return items
}

// GetFormTypeForApplication détermine le type d'e-Form officiel correspondant
func GetFormTypeForApplication(cat EstablishmentCategory, nat ApplicationNature) string {
	switch cat {
	case CategoryFabrication:
		switch nat {
		case NaturePrealableRealisation:
			return "eForm-MIPH-FAB-PREALABLE"
		case NatureOuvertureInitiale:
			return "eForm-MIPH-FAB-OUVERTURE"
		case NatureSousTraitance:
			return "eForm-MIPH-FAB-SOUS-TRAITANCE"
		case NatureExercicePDT:
			return "eForm-MIPH-PDT"
		case NatureExerciceAssistant:
			return "eForm-MIPH-ASSISTANT"
		default:
			return "eForm-MIPH-FAB-MODIF"
		}
	case CategoryImportation:
		return "eForm-MIPH-IMPORTATION"
	case CategoryDistribution:
		return "eForm-MIPH-DISTRIBUTION"
	case CategoryExploitation:
		return "eForm-MIPH-EXPLOITATION"
	default:
		return fmt.Sprintf("eForm-MIPH-%s", cat)
	}
}
