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
	CategoryExportation  EstablishmentCategory = "EXPORTATION"
)

// ApplicationNature représente la nature de la démarche ministérielle auprès du MIPH
type ApplicationNature string

const (
	NaturePrealableRealisation ApplicationNature = "PREALABLE_REALISATION"
	NatureOuvertureInitiale    ApplicationNature = "OUVERTURE_INITIALE"
	NatureAgrementInitial      ApplicationNature = "AGREMENT_INITIAL"
	NatureSousTraitance        ApplicationNature = "SOUS_TRAITANCE"
	NatureExercicePDT          ApplicationNature = "EXERCICE_PDT"
	NatureExerciceAssistant    ApplicationNature = "EXERCICE_ASSISTANT"
	NatureRenouvellement       ApplicationNature = "RENOUVELLEMENT"
	NatureModifSubstantielle   ApplicationNature = "MODIF_SUBSTANTIELLE"
	NatureModificationSubst    ApplicationNature = "MODIFICATION_SUBSTANTIELLE"
	NatureModifNonSubst        ApplicationNature = "MODIF_NON_SUBSTANTIELLE"
	NatureExtensionLignes      ApplicationNature = "EXTENSION_LIGNES"
	NatureTransfertSite        ApplicationNature = "TRANSFERT_SITE"
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
	ID                     uuid.UUID              `json:"id"`
	OrganizationID         uuid.UUID              `json:"organization_id"`
	EstablishmentID        *uuid.UUID             `json:"establishment_id,omitempty"`
	Category               EstablishmentCategory  `json:"category"`
	Nature                 ApplicationNature      `json:"nature"`
	ApplicationNumber      string                 `json:"application_number"`
	InternalReference      string                 `json:"internal_reference,omitempty"`
	Title                  string                 `json:"title"`
	Status                 ApplicationStatus      `json:"status"`
	FormType               string                 `json:"form_type"`
	FormData               map[string]interface{} `json:"form_data"`
	CompletenessPercent    float64                `json:"completeness_percent"`
	PDTSignedBy            *uuid.UUID             `json:"pdt_signed_by,omitempty"`
	PDTSignedAt            *time.Time             `json:"pdt_signed_at,omitempty"`
	PDTSignatureSeal       string                 `json:"pdt_signature_seal,omitempty"`
	TabadolSubmissionRef   string                 `json:"tabadol_submission_ref,omitempty"`
	TabadolSubmittedAt     *time.Time             `json:"tabadol_submitted_at,omitempty"`
	TabadolSubmissionNotes string                 `json:"tabadol_submission_notes,omitempty"`
	MIPHReceiptRef         string                 `json:"miph_receipt_ref,omitempty"`
	MIPHReceiptDate        *string                `json:"miph_receipt_date,omitempty"`
	DecisionRef            string                 `json:"decision_ref,omitempty"`
	DecisionDate           *string                `json:"decision_date,omitempty"`
	LicenseExpiryDate      *string                `json:"license_expiry_date,omitempty"`
	CreatedBy              *uuid.UUID             `json:"created_by,omitempty"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`

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

		case NatureOuvertureInitiale, NatureAgrementInitial:
			// 1. Dossier Administratif & Juridique
			add("FAB.ADM.01", "Statuts notariés de la société enregistrés et certifiés conformes", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.02", "Registre du Commerce électronique CNRC (code activité fabrication pharmaceutique)", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.03", "Extrait de rôle fiscal apuré (validité < 3 mois, sans échéancier ni dette contestée)", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.04", "Carte d'immatriculation fiscale (NIF) et attestation d'identification statistique (NIS)", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.05", "Attestations de mise à jour des cotisations sociales CNAS et CASNOS (< 3 mois)", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.06", "Titre de propriété foncière notarié ou contrat de bail notarié d'une durée minimale de 10 ans enregistré", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.07", "Police d'assurance en responsabilité civile professionnelle couvrant les risques industriels et d'exploitation", "ADMINISTRATIF", "E", true, "")

			// 2. Conception Architecturale & Zonage BPF
			add("FAB.PLN.01", "Plan de masse général côté et plans de coupes architecturales as-built certifiés conformes", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.PLN.02", "Plan de zonage des classes de propreté particulaire et microbiologique (ISO 5 à 8 / Grades A à D)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.FLX.01", "Synoptique des flux des matières premières (API, excipients) et articles de conditionnement", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.FLX.02", "Synoptique des flux du personnel et gestion des sas vestiaires (habillage/déshabillage)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.FLX.03", "Synoptique d'évacuation des déchets industriels, effluents et produits sous quarantaine/rejet", "TECHNIQUE_LOCAUX", "E", true, "")

			// 3. Utilités Critiques & Équipements Industriels
			add("FAB.UTL.01", "Rapport de qualification des systèmes de traitement d'air CVC/HVAC (surpressions en cascade et filtration HEPA H14)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.UTL.02", "Rapports de qualification et de validation des boucles d'eau purifiée et d'eau PPI (Pharmacopée Européenne)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.EQP.01", "Dossiers de Qualification QI/QO/QP des équipements majeurs de production et de conditionnement", "TECHNIQUE_LOCAUX", "E", true, "")

			// 4. Système Qualité Pharmaceutique & BPF
			add("FAB.PQS.01", "Manuel du Système de Gestion de la Qualité Pharmaceutique (PQS) et procédures maîtresses de libération des lots", "ORGANISATION_BPF", "E", true, "")
			add("FAB.PQS.02", "Organigramme BPF nominatif garantissant l'indépendance hiérarchique du Contrôle Qualité vis-à-vis de la Production", "ORGANISATION_BPF", "E", true, "")

			// 5. Direction Technique & Personnel Clé (Art. 16 à 18 Décret 21-82)
			add("FAB.PDT.01", "Diplôme d'État de Docteur en Pharmacie du Pharmacien Directeur Technique (certifié conforme)", "PHARMACIEN_DT", "E", true, "")
			add("FAB.PDT.02", "Attestation d'inscription au tableau de l'Ordre des Pharmaciens (Conseil National de l'Ordre CNOP)", "PHARMACIEN_DT", "E", true, "")
			add("FAB.PDT.03", "Attestation d'expérience professionnelle préalable minimale en milieu industriel pharmaceutique (Art. 17 Décret 21-82)", "PHARMACIEN_DT", "E", true, "")
			add("FAB.PDT.04", "Contrat de travail à durée indéterminée exclusif à temps plein du Pharmacien DT (Art. 18 Décret 21-82)", "PHARMACIEN_DT", "E", true, "")
			add("FAB.PDT.05", "Contrats de travail et diplômes des Pharmaciens Assistants / Adjoints avec inscription à l'Ordre", "PHARMACIEN_DT", "C", true, "")

		case NatureRenouvellement:
			add("FAB.REN.01", "Rapport d'activité quinquennal de production et état récapitulatif des spécialités fabriquées", "ADMINISTRATIF", "E", true, "")
			add("FAB.REN.02", "Bilan périodique d'assurance qualité et revues annuelles de la qualité des produits (PQR)", "ORGANISATION_BPF", "E", true, "")
			add("FAB.REN.03", "Rapports de requalification périodique des utilités critiques (HVAC/CVC, boucles d'eau PPI et vapeur propre)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.REN.04", "Attestations de régularité fiscale et parafiscale (CNAS/CASNOS) en cours de validité", "ADMINISTRATIF", "E", true, "")
			add("FAB.REN.05", "Dossier d'exercice à jour du Pharmacien Directeur Technique et attestations d'inscription à l'Ordre", "PHARMACIEN_DT", "E", true, "")

		case NatureModifSubstantielle, NatureModificationSubst:
			add("MOD.SUB.01", "Rapport technico-réglementaire motivant la modification substantielle (Arrêté 22/09/2025)", "MODIFICATIONS", "E", true, "")
			add("MOD.SUB.02", "Plans comparatifs certifiés de l'état approuvé vs état modifié", "MODIFICATIONS", "E", true, "")
			add("MOD.SUB.03", "Rapports de requalification des utilités et locaux impactés (CVC/eau PPI)", "MODIFICATIONS", "E", true, "")
			add("MOD.SUB.04", "Analyse de risques BPF et plan de contrôle de changement (ICH Q9 / Q10)", "MODIFICATIONS", "E", true, "")

		case NatureExtensionLignes:
			add("FAB.EXT.01", "Dossier descriptif des nouvelles formes galéniques ou lignes de production additionnelles", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.EXT.02", "Plans de zonage et d'implantation des nouveaux équipements", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.EXT.03", "Protocoles et rapports de qualification QI/QO/QP des nouvelles lignes", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.EXT.04", "Mise à jour de l'organigramme technique et effectifs pharmaciens dédiés", "ORGANISATION_BPF", "E", true, "")

		case NatureTransfertSite:
			add("FAB.TRS.01", "Motifs du transfert et copie des actes notariés du nouveau site industriel", "ADMINISTRATIF", "E", true, "")
			add("FAB.TRS.02", "Dossier technique complet des nouveaux locaux et installations", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.TRS.03", "Plan de transfert et de validation des procédés sur le nouveau site", "ORGANISATION_BPF", "E", true, "")

		default:
			// Checklist granulaire Décret 21-82 par défaut
			add("FAB.ADM.01", "Statuts notariés de la société enregistrés et certifiés conformes", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.02", "Registre du Commerce électronique CNRC (code activité fabrication pharmaceutique)", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.03", "Extrait de rôle fiscal apuré (validité < 3 mois, sans échéancier ni dette contestée)", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.04", "Carte d'immatriculation fiscale (NIF) et attestation d'identification statistique (NIS)", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.05", "Attestations de mise à jour des cotisations sociales CNAS et CASNOS (< 3 mois)", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.06", "Titre de propriété foncière notarié ou contrat de bail notarié d'une durée minimale de 10 ans enregistré", "ADMINISTRATIF", "E", true, "")
			add("FAB.ADM.07", "Police d'assurance en responsabilité civile professionnelle couvrant les risques industriels et d'exploitation", "ADMINISTRATIF", "E", true, "")
			add("FAB.PLN.01", "Plan de masse général côté et plans de coupes architecturales as-built certifiés conformes", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.PLN.02", "Plan de zonage des classes de propreté particulaire et microbiologique (ISO 5 à 8 / Grades A à D)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.FLX.01", "Synoptique des flux des matières premières (API, excipients) et articles de conditionnement", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.FLX.02", "Synoptique des flux du personnel et gestion des sas vestiaires (habillage/déshabillage)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.FLX.03", "Synoptique d'évacuation des déchets industriels, effluents et produits sous quarantaine/rejet", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.UTL.01", "Rapport de qualification des systèmes de traitement d'air CVC/HVAC (surpressions en cascade et filtration HEPA H14)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.UTL.02", "Rapports de qualification et de validation des boucles d'eau purifiée et d'eau PPI (Pharmacopée Européenne)", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.EQP.01", "Dossiers de Qualification QI/QO/QP des équipements majeurs de production et de conditionnement", "TECHNIQUE_LOCAUX", "E", true, "")
			add("FAB.PQS.01", "Manuel du Système de Gestion de la Qualité Pharmaceutique (PQS) et procédures maîtresses de libération des lots", "ORGANISATION_BPF", "E", true, "")
			add("FAB.PQS.02", "Organigramme BPF nominatif garantissant l'indépendance hiérarchique du Contrôle Qualité vis-à-vis de la Production", "ORGANISATION_BPF", "E", true, "")
			add("FAB.PDT.01", "Diplôme d'État de Docteur en Pharmacie du Pharmacien Directeur Technique (certifié conforme)", "PHARMACIEN_DT", "E", true, "")
			add("FAB.PDT.02", "Attestation d'inscription au tableau de l'Ordre des Pharmaciens (Conseil National de l'Ordre CNOP)", "PHARMACIEN_DT", "E", true, "")
			add("FAB.PDT.03", "Attestation d'expérience professionnelle préalable minimale en milieu industriel pharmaceutique (Art. 17 Décret 21-82)", "PHARMACIEN_DT", "E", true, "")
			add("FAB.PDT.04", "Contrat de travail à durée indéterminée exclusif à temps plein du Pharmacien DT (Art. 18 Décret 21-82)", "PHARMACIEN_DT", "E", true, "")
			add("FAB.PDT.05", "Contrats de travail et diplômes des Pharmaciens Assistants / Adjoints avec inscription à l'Ordre", "PHARMACIEN_DT", "C", true, "")
		}

	case CategoryImportation:
		add("IMP.ADM.01", "Statuts notariés et Registre du Commerce électronique CNRC (code activité importation pharmaceutique)", "ADMINISTRATIF", "E", true, "")
		add("IMP.ADM.02", "Extrait de rôle fiscal apuré (validité < 3 mois) et attestations CNAS/CASNOS à jour", "ADMINISTRATIF", "E", true, "")
		add("IMP.ADM.03", "Programme prévisionnel d'importation annuel visé par la Direction de la Pharmacie MIPH", "ADMINISTRATIF", "E", true, "")
		add("IMP.ADM.04", "Titre de propriété notarié ou contrat de bail notarié enregistré des locaux et entrepôts de stockage (300 m² min. médicaments / 90 m² min. DM / 390 m² mixte)", "ADMINISTRATIF", "E", true, "")
		add("IMP.PDT.01", "Diplôme d'État de Docteur en Pharmacie et inscription Ordre CNOP du Pharmacien DT", "PHARMACIEN_DT", "E", true, "")
		add("IMP.PDT.02", "Contrat de travail à temps plein exclusif et décision nominative d'exercice du Pharmacien DT", "PHARMACIEN_DT", "E", true, "")
		add("IMP.TEC.01", "Descriptif des magasins de stockage sous température contrôlée et zones sécurisées (quarantaine/rejet sous clé)", "CONDITIONS_TECHNIQUES", "E", true, "")
		add("IMP.TEC.02", "Rapports de qualification et cartographies des chambres froides (+2°C à +8°C) et zones climatisées (+15°C à +25°C)", "CONDITIONS_TECHNIQUES", "E", true, "")
		add("IMP.BPF.01", "Procédure de contrôle documentaire, échantillonnage et libération des lots importés par le Pharmacien DT", "ORGANISATION_BPF", "E", true, "")
		add("IMP.LOG.01", "Conventions d'assurance qualité avec les transitaires et transporteurs sous chaîne du froid validée", "CONDITIONS_TECHNIQUES", "E", true, "")

	case CategoryDistribution:
		add("DIS.ADM.01", "Statuts notariés et Registre du Commerce électronique CNRC (code distribution en gros)", "ADMINISTRATIF", "E", true, "")
		add("DIS.ADM.02", "Extrait de rôle fiscal apuré (< 3 mois) et attestations CNAS/CASNOS à jour", "ADMINISTRATIF", "E", true, "")
		add("DIS.ADM.03", "Cahier des charges de distribution en gros paraphé et engagement d'approvisionnement continu du marché", "ADMINISTRATIF", "E", true, "")
		add("DIS.ADM.04", "Titre de propriété notarié ou contrat de bail notarié enregistré des locaux et entrepôts de distribution", "ADMINISTRATIF", "E", true, "")
		add("DIS.PDT.01", "Dossier nominatif du Pharmacien DT (Diplôme, Inscription Ordre CNOP, Contrat exclusif)", "PHARMACIEN_DT", "E", true, "")
		add("DIS.TEC.01", "Plans et descriptif des entrepôts au 1/100e selon les Bonnes Pratiques de Distribution en Gros (BPDG)", "CONDITIONS_TECHNIQUES", "E", true, "")
		add("DIS.TEC.02", "Inventaire de la flotte logistique et certificats d'étalonnage périodique des enregistreurs de température", "CONDITIONS_TECHNIQUES", "E", true, "")
		add("DIS.BPF.01", "Système informatique validé de traçabilité des lots et procédure de rappel d'urgence (ICH Q10)", "ORGANISATION_BPF", "E", true, "")

	case CategoryExploitation:
		add("EXP.ADM.01", "Statuts et Registre de Commerce d'exploitant titulaire de décisions d'enregistrement ANPP et/ou d'homologation DM", "ADMINISTRATIF", "E", true, "")
		add("EXP.ADM.02", "Extrait de rôle fiscal apuré (< 3 mois) et attestations CNAS/CASNOS en cours de validité", "ADMINISTRATIF", "E", true, "")
		add("EXP.ADM.03", "Cahier des conditions techniques à l'exploitation paraphé avec engagement déontologique", "ADMINISTRATIF", "E", true, "")
		add("EXP.ADM.04", "Titre de propriété notarié ou contrat de bail notarié enregistré du siège et des installations de l'exploitant", "ADMINISTRATIF", "E", true, "")
		add("EXP.PDT.01", "Dossier et décision d'exercice nominative du Pharmacien DT Responsable de l'Exploitation", "PHARMACIEN_DT", "E", true, "")
		add("EXP.PV.01", "Désignation officielle de la Personne Responsable de la Pharmacovigilance (QPPV) et système d'alerte 24/7", "ORGANISATION_BPF", "E", true, "")
		add("EXP.BPF.01", "Procédure de contrôle et validation de l'information médicale/scientifique par le Pharmacien DT", "ORGANISATION_BPF", "E", true, "")
		add("EXP.TRC.01", "Système de suivi de la traçabilité des lots sur le marché national et procédure de rappel de lots", "ORGANISATION_BPF", "E", true, "")

	case CategoryExportation:
		add("EXT.ADM.01", "Statuts notariés et Registre du Commerce électronique CNRC (codes 702103 médicaments / 702106 DM)", "ADMINISTRATIF", "E", true, "")
		add("EXT.ADM.02", "Extrait de rôle fiscal apuré (< 3 mois) et attestations CNAS/CASNOS à jour", "ADMINISTRATIF", "E", true, "")
		add("EXT.ADM.03", "Cahier des charges à l'exportation paraphé et engagement de conformité réglementaire (Décret 21-82 Art. 12)", "ADMINISTRATIF", "E", true, "")
		add("EXT.ADM.04", "Titre de propriété notarié ou contrat de bail notarié enregistré des locaux et aires d'entreposage sous douane/export", "ADMINISTRATIF", "E", true, "")
		add("EXT.PDT.01", "Dossier nominatif du Pharmacien DT (Diplôme d'État, Inscription CNOP, Contrat exclusif temps plein)", "PHARMACIEN_DT", "E", true, "")
		add("EXT.TEC.01", "Descriptif des magasins de stockage sous température contrôlée et conventions de sourcing avec les fabricants locaux", "CONDITIONS_TECHNIQUES", "E", true, "")
		add("EXT.TRC.01", "Système informatisé de traçabilité des lots exportés et procédure de délivrance des certificats d'exportation", "ORGANISATION_BPF", "E", true, "")
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
	case CategoryExportation:
		return "eForm-MIPH-EXPORTATION"
	default:
		return fmt.Sprintf("eForm-MIPH-%s", cat)
	}
}
