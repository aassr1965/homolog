package reporting

import (
	"homolog/internal/domain"
	"strings"
)

// ReportKind qualifie la nature légale de l'état documenté (Chapitre 11 CDC)
type ReportKind string

const (
	KindEtatReglementaire ReportKind = "ES" // État de Sortie Réglementaire (opposable, visa PDT 2FA requis, QR-code scellé)
	KindEtatIntermediaire ReportKind = "EI" // État Intermédiaire de Pilotage (interne de travail, filigrane DOCUMENT DE TRAVAIL)
	KindMasterTemplate    ReportKind = "MT" // Template Maître Qualité / BPF (cadre d'épreuve et de libération)
)

// ReportCategory regroupe les états documentaires par famille fonctionnelle
type ReportCategory string

const (
	CategoryDossiersCTD          ReportCategory = "DOSSIERS_CTD"          // Famille 1 : Dossiers Réglementaires & Dépôts ANPP
	CategoryDispositifsMedicaux  ReportCategory = "DISPOSITIFS_MEDICAUX"  // Volet DDM : Homologation & Redevances DM
	CategoryFinancesRedevances   ReportCategory = "FINANCES_REDEVANCES"   // Famille 2 : États Financiers & Redevances ANPP
	CategoryVigilanceComplements ReportCategory = "VIGILANCE_COMPLEMENTS" // Famille 3 : Vigilance, Compléments & RDV ANPP
	CategoryQualiteBPF           ReportCategory = "QUALITE_BPF"           // Famille 4 : Assurance Qualité, Intégrité BPF & 21 CFR Part 11
	CategoryPilotageStrategique  ReportCategory = "PILOTAGE_STRATEGIQUE"  // Famille 5 : Pilotage Stratégique, Direction Générale
	CategoryControleLiberation   ReportCategory = "CONTROLE_LIBERATION"   // Famille 6 : Contrôle DCE, Métrologie & Libération
	CategorySuiviAvancement      ReportCategory = "SUIVI_AVANCEMENT"      // Volet EI : Pilotage interne & comités de projet
)

// ReportDefinition encapsule les métadonnées normatives d'un état réglementaire ou intermédiaire
type ReportDefinition struct {
	Code                 string                 `json:"code"`                   // Ex: ES-REG-01, ES-FIN-01, EI-AVT-01
	Kind                 ReportKind             `json:"kind"`                   // ES, EI, MT
	Category             ReportCategory         `json:"category"`               // Famille fonctionnelle
	Title                string                 `json:"title"`                  // Dénomination officielle
	Subtitle             string                 `json:"subtitle"`               // Sous-titre explicatif
	LegalReference       string                 `json:"legal_reference"`        // Base juridique (Loi, Décret, Arrêté, Guide CTD ANPP)
	DefaultFormat        string                 `json:"default_format"`         // pdf, html, xlsx
	SupportedFormats     []string               `json:"supported_formats"`      // Formats autorisés à l'export
	RequiresPDTSignature bool                   `json:"requires_pdt_signature"` // Nécessite le visa 2FA du Pharmacien DT
	IsPubliclyVerifiable bool                   `json:"is_publicly_verifiable"` // Vérifiable publiquement par QR-Code (/verify/{sha256})
	DefaultWatermark     domain.WatermarkStatus `json:"default_watermark"`      // BROUILLON, OFFICIEL, ARCHIVE
	FrequencyTrigger     string                 `json:"frequency_trigger"`      // Déclencheur & Fréquence d'émission
	IssuerRole           string                 `json:"issuer_role"`            // Rôle habilité à émettre
	Recipient            string                 `json:"recipient"`              // Destinataire officiel
	Description          string                 `json:"description"`            // Description détaillée CDC
}

// MasterRegistry répertorie les 28 états documentaires BPF officiels (20 ES + 8 EI) et le template maître MT-LCQ-01
var MasterRegistry = []ReportDefinition{
	// =========================================================================
	// FAMILLE 1 : DOSSIERS RÉGLEMENTAIRES & DÉPÔTS ANPP (PIÈCES MAÎTRESSES)
	// =========================================================================
	{
		Code:                 "ES-REG-01",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryDossiersCTD,
		Title:                "Bordereau de Dépôt / Transmission Officiel CTD",
		Subtitle:             "Inventaire exhaustif des 5 modules CTD et décompte des volumes physiques et numériques",
		LegalReference:       "Guide CTD ANPP Juillet 2025 • Bonnes Pratiques de Fabrication Algériennes",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html", "json"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "À chaque soumission physique ou dépôt de dossier",
		IssuerRole:           "Directeur Technique / Pharmacien Responsable",
		Recipient:            "Bureau d'Ordre ANPP / Direction de l'Enregistrement (DEPP)",
		Description:          "Inventaire officiel obligatoire accompagnant le dépôt physique et numérique du dossier CTD à l'ANPP avec ventilation des pièces par module M1-M5.",
	},
	{
		Code:                 "ES-REG-02",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryDossiersCTD,
		Title:                "Formulaire Officiel de Demande d'Enregistrement (1.A.3)",
		Subtitle:             "Formulaire administratif et réglementaire d'Autorisation de Mise sur le Marché Médicaments",
		LegalReference:       "Arrêté du 16 décembre 2020 fixant la forme et le contenu de la demande d'enregistrement",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html", "json"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Constitution initiale d'un dossier NI, NF, RI ou RF",
		IssuerRole:           "Directeur Technique & Direction Générale",
		Recipient:            "Direction de l'Enregistrement des Produits Pharmaceutiques (DEPP)",
		Description:          "Formulaire officiel 1.A.3 normalisé comprenant les caractéristiques administratives, pharmaceutiques et cliniques du médicament soumis à l'ANPP.",
	},
	{
		Code:                 "ES-REG-02-DM",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryDispositifsMedicaux,
		Title:                "Formulaire Officiel de Demande d'Homologation DM (6 sections)",
		Subtitle:             "Dossier technique et administratif d'homologation de Dispositif Médical",
		LegalReference:       "Décret exécutif n° 20-324 du 22 novembre 2020 & Arrêté du 10 mai 2021",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html", "json"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Constitution d'un dossier d'homologation DM (Classes I, IIa, IIb, III)",
		IssuerRole:           "Directeur Technique & Pharmacien Responsable",
		Recipient:            "Direction des Dispositifs Médicaux ANPP (DDM)",
		Description:          "Formulaire complet en 6 sections : identification, nature de demande, composition technique, fabricant/sites, sécurité/performances, et déclaration solennelle du Pharmacien DT.",
	},
	{
		Code:                 "ES-REG-03",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryDossiersCTD,
		Title:                "Table des Matières Exhaustive & Index Paginé CTD",
		Subtitle:             "Sommaire interactif M1-M5 avec références croisées et empreintes SHA-256",
		LegalReference:       "Guide CTD ANPP Juillet 2025 • ICH M4",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Scellement final d'un Submission Package avant remise à l'ANPP",
		IssuerRole:           "Responsable Affaires Réglementaires / Chef de Projet",
		Recipient:            "Commissions d'Évaluation & Experts ANPP",
		Description:          "Index intégral paginé des pièces constitutives du dossier CTD avec vérification bit-à-bit des documents versés au coffre-fort BPF.",
	},
	{
		Code:                 "ES-REG-03-DM",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryDispositifsMedicaux,
		Title:                "Récépissé Officiel de Dépôt des Dossiers soumis à Homologation DM",
		Subtitle:             "Grille de recevabilité matérielle en 13 points et double visa officiel",
		LegalReference:       "Décret exécutif n° 20-324 & Circulaire DDM ANPP 2021",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html", "json"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Dépôt physique et enregistrement guichet Direction DM",
		IssuerRole:           "Directeur Technique & Agent Récepteur ANPP",
		Recipient:            "Direction des Dispositifs Médicaux ANPP / Archives Labo",
		Description:          "Acte officiel de recevabilité matérielle vérifiant les 13 points de contrôle exigés par l'ANPP (exemplaires électroniques, BV BNA, ISO 13485, etc.).",
	},
	{
		Code:                 "ES-REG-04",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryDossiersCTD,
		Title:                "Fiche Récapitulative de Modification (LD-DEPP07/V01)",
		Subtitle:             "Tableau comparatif Approuvé vs Proposé des rubriques 1 à 34 de la Ligne Directrice Variations",
		LegalReference:       "Ligne Directrice LD-DEPP07/V01 ANPP • Décret exécutif n° 21-82",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html", "xlsx"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Dépôt d'une variation Mmin, Mmod, Mmaj ou Grouping",
		IssuerRole:           "Directeur Technique / Responsable Réglementaire",
		Recipient:            "Sous-Direction des Variations ANPP",
		Description:          "Relevé formel détaillant l'état approuvé, la modification proposée, la justification technique, le classement réglementaire et l'impact sur le dossier AMM.",
	},
	{
		Code:                 "ES-REG-05",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryDossiersCTD,
		Title:                "Bordereau d'Envoi des Échantillons & Réactifs (1.P)",
		Subtitle:             "Fiche de traçabilité LNCPP avec N° de scellés inviolables et relevé datalogger chaîne du froid (+2°C/+8°C)",
		LegalReference:       "Procédure Fr-DCE02-F/V01 ANPP • Norme LNCPP",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Convocation officielle au dépôt des échantillons physiques par l'ANPP",
		IssuerRole:           "Responsable Contrôle Qualité / Pharmacien BPF",
		Recipient:            "Laboratoire National de Contrôle des Produits Pharmaceutiques (LNCPP / DCE)",
		Description:          "Bordereau opposable d'expédition des lots de produit fini, substances de référence, colonnes et réactifs spécifiques requis pour l'expertise analytique.",
	},
	{
		Code:                 "ES-ENG-1P8",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryDossiersCTD,
		Title:                "Engagement de Disponibilité Continue sur le Marché National (1.P.8)",
		Subtitle:             "Attestation solennelle de constitution de stock de sécurité 3 mois et approvisionnement ininterrompu",
		LegalReference:       "Décret exécutif n° 20-325 du 22 novembre 2020 • Guide CTD ANPP Juillet 2025",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Constitution de tout dossier de fabrication locale (NF) ou importation (NI)",
		IssuerRole:           "Pharmacien Directeur Technique Titulaire",
		Recipient:            "Direction de la Production & de la Régulation MIPH / ANPP",
		Description:          "Lettre d'engagement légal obligatoire par laquelle le Pharmacien DT garantit sous serment la disponibilité permanente du médicament et un stock de 3 mois.",
	},
	{
		Code:                 "ES-ENG-1P9",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryDossiersCTD,
		Title:                "Engagement de Pharmacovigilance & Notification Immédiate (1.P.9)",
		Subtitle:             "Maintien du système permanent 24/7, désignation du QPPV et délais de déclaration 15j / 48h",
		LegalReference:       "Loi sanitaire n° 18-11 • Lignes Directrices de Pharmacovigilance ANPP / CNPM",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Dépôt de toute demande d'enregistrement (NI, NF, RI, RF)",
		IssuerRole:           "Pharmacien Directeur Technique Titulaire",
		Recipient:            "Centre National de Pharmacovigilance & Matériovigilance (CNPM) / ANPP",
		Description:          "Acte officiel certifiant l'organisation de la pharmacovigilance, la nomination du QPPV et l'engagement de transmission sans délai de toute alerte sanitaire.",
	},
	{
		Code:                 "ES-ENG-1H10",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryDossiersCTD,
		Title:                "Déclaration Solennelle de Non-Modification (1.H.10)",
		Subtitle:             "Attestation formelle de conformité bit-à-bit du produit pour renouvellement quinquennal",
		LegalReference:       "Arrêté ministériel du 16 décembre 2020 fixant les modalités de renouvellement",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Dossier de renouvellement d'enregistrement (RI / RF / DM)",
		IssuerRole:           "Pharmacien Directeur Technique Titulaire",
		Recipient:            "Direction de l'Enregistrement des Produits Pharmaceutiques (DEPP)",
		Description:          "Déclaration solennelle sous la foi du serment certifiant qu'aucune modification non autorisée n'a été apportée depuis la décision d'enregistrement précédente.",
	},

	// =========================================================================
	// FAMILLE 2 : ÉTATS FINANCIERS & REDEVANCES ANPP
	// =========================================================================
	{
		Code:                 "ES-FIN-01",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryFinancesRedevances,
		Title:                "Bordereau de Versement BV1 / BV DM (Enregistrement Initial)",
		Subtitle:             "Liquidation officielle des droits d'enregistrement (Trésor Public / BNA Chéraga)",
		LegalReference:       "Ordonnance n° 21-07 du 8 juin 2021 & Décret exécutif n° 21-82 du 23 février 2021",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html", "xlsx"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Préalable obligatoire au dépôt NI, NF ou Homologation DM",
		IssuerRole:           "Directeur Financier & Pharmacien Directeur Technique",
		Recipient:            "Régie des Recettes ANPP / Banque Nationale d'Algérie / Trésor Public",
		Description:          "Décompte certifié des redevances réglementaires : 355 000 DZD pour médicament ou 300 000 DZD sur le compte BNA Al Qods Chéraga RIB 00100622030000064395 pour Dispositifs Médicaux.",
	},
	{
		Code:                 "ES-FIN-02",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryFinancesRedevances,
		Title:                "Bordereau de Versement BV Variations, Renouvellements & Transferts",
		Subtitle:             "Calcul automatique selon barème réglementaire Décret 21-82 et grilles DDM",
		LegalReference:       "Décret exécutif n° 21-82 • Tarification officielle ANPP",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html", "xlsx"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Dépôt de modifications (MI/MF), renouvellements (RI/RF/DM) ou transferts",
		IssuerRole:           "Directeur Financier & Pharmacien DT",
		Recipient:            "Régie des Recettes ANPP",
		Description:          "Bordereau de liquidation pour renouvellement (205k DZD PP / 300k DZD DM), variations majeures (115k DZD PP / 150k DZD DM) ou transferts de titulaire (100k DZD DM).",
	},
	{
		Code:                 "ES-FIN-03",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryFinancesRedevances,
		Title:                "Fiche d'Imputation & Rapprochement Bancaire des Redevances",
		Subtitle:             "Tableau de concordance Avis de Débit, Bordereau BV et Quittance officielle ANPP",
		LegalReference:       "Règlement comptable et financier pharmaceutique",
		DefaultFormat:        "xlsx",
		SupportedFormats:     []string{"xlsx", "pdf", "csv"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Mensuel ou à réception de la quittance définitive ANPP",
		IssuerRole:           "Responsable Comptable / Contrôle de Gestion",
		Recipient:            "Direction Financière & Audit Interne",
		Description:          "Rapprochement comptable bit-à-bit des flux bancaires avec les quittances émises par la régie des recettes de l'ANPP pour apurement des comptes d'attente.",
	},

	// =========================================================================
	// FAMILLE 3 : VIGILANCE RÉGLEMENTAIRE, COMPLÉMENTS & RENDEZ-VOUS ANPP
	// =========================================================================
	{
		Code:                 "ES-VIG-01",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryVigilanceComplements,
		Title:                "Bordereau de Réponse aux Compléments (Deficiency Reply)",
		Subtitle:             "Tableau croisé officiel Question ANPP / Argumentaire / Pièces justificatives GED",
		LegalReference:       "Règlement intérieur des commissions d'enregistrement ANPP (Délais 60j / 90j)",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Notification d'une lettre de carence ou demande d'informations complémentaires",
		IssuerRole:           "Pharmacien Directeur Technique",
		Recipient:            "Commission Technique d'Enregistrement / Direction de l'Enregistrement",
		Description:          "Bordereau formel de levée des réserves formulées par les commissions ANPP, liant chaque item à son livrable probant déposé dans le coffre-fort BPF.",
	},
	{
		Code:                 "ES-VIG-02",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryVigilanceComplements,
		Title:                "Fiche de Convocation & Préparation aux Rendez-vous ANPP",
		Subtitle:             "Synthèse du créneau confirmé (rdvdm.anpp.dz / E-Tasjil) et dossier préparatoire",
		LegalReference:       "Dispositif de gestion des audiences et dépôts ANPP",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Confirmation d'un rendez-vous sur le portail ANPP",
		IssuerRole:           "Responsable Affaires Réglementaires",
		Recipient:            "Délégation Laboratoire / Accueil Sécurité ANPP",
		Description:          "Fiche officielle de mission accréditant la délégation du laboratoire pharmaceutique auprès du guichet de dépôt ANPP avec l'ordre du jour exhaustif.",
	},
	{
		Code:                 "ES-VIG-03",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryVigilanceComplements,
		Title:                "Rapport Périodique de Veille & Échéancier Réglementaire",
		Subtitle:             "Cartographie des AMM à renouveler (J-180, J-90, J-30) et textes légaux parus au JO",
		LegalReference:       "Loi n° 18-11 relative à la santé • Normes de Veille Réglementaire MIPH",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "xlsx", "html"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Mensuel / Trimestriel",
		IssuerRole:           "Responsable Affaires Réglementaires / Pharmacien Veille",
		Recipient:            "Comité de Direction / Direction Générale",
		Description:          "Tableau de bord prospectif assurant la pérennité du portefeuille AMM/DM et anticipant les impacts des nouvelles dispositions législatives algériennes.",
	},

	// =========================================================================
	// FAMILLE 4 : ASSURANCE QUALITÉ, INTÉGRITÉ BPF & 21 CFR PART 11
	// =========================================================================
	{
		Code:                 "ES-QUA-01",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryQualiteBPF,
		Title:                "Registre d'Audit Trail BPF Scellé & Traçabilité ALCOA+",
		Subtitle:             "Journal immuable append-only des événements système, horodatages UTC+1 et empreintes SHA-256",
		LegalReference:       "21 CFR Part 11 • Annexe 11 des BPF • ALCOA+",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "csv", "json"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkOfficial,
		FrequencyTrigger:     "Sur demande d'inspection ANPP, audit qualité ou revue annuelle de lot",
		IssuerRole:           "Responsable Assurance Qualité / Pharmacien BPF",
		Recipient:            "Inspecteurs BPF ANPP / Auditeurs Qualité / Direction Générale",
		Description:          "Extraction certifiée du grand livre d'audit trail prouvant l'intégrité intégrale des données pharmaceutiques et l'absence totale d'altération rétrospective.",
	},
	{
		Code:                 "ES-QUA-02",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryQualiteBPF,
		Title:                "Manifeste de Signature Électronique BPF (Maker-Checker)",
		Subtitle:             "Preuve juridique de double visa conforme 21 CFR Part 11, certificats et tokens 2FA",
		LegalReference:       "21 CFR Part 11 Subpart B • GAMP 5 • Décret exécutif n° 21-82",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkOfficial,
		FrequencyTrigger:     "Scellement de toute version majeure d'un document ou dossier CTD",
		IssuerRole:           "Pharmacien Directeur Technique",
		Recipient:            "Dossier Maître AMM / Archives Centrales",
		Description:          "Certificat probant d'authentification attestant de la séparation des tâches (Maker ≠ Checker) et de la validité de la signature électronique qualifiée du DT.",
	},
	{
		Code:                 "ES-QUA-03",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryQualiteBPF,
		Title:                "Fiche de Contrôle Documentaire Pré-Dépôt (Pre-Submission Audit)",
		Subtitle:             "Procès-verbal d'audit d'exhaustivité et de conformité réglementaire avant remise ANPP",
		LegalReference:       "Système de Management de la Qualité Pharmaceutique (ICH Q10)",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Préalable absolu à la signature et au scellement de soumission PDT",
		IssuerRole:           "Responsable Assurance Qualité & Directeur Technique",
		Recipient:            "Comité d'Approbation Réglementaire Interne",
		Description:          "Bilan d'audit pré-dépôt validant l'absence de pièces manquantes, l'exactitude des métadonnées et la complétude documentaire requise par l'ANPP.",
	},

	// =========================================================================
	// FAMILLE 5 : PILOTAGE STRATÉGIQUE, DIRECTION GÉNÉRALE & RÉVERSIBILITÉ
	// =========================================================================
	{
		Code:                 "ES-DIR-01",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryPilotageStrategique,
		Title:                "Tableau de Bord Exécutif du Portefeuille Produits",
		Subtitle:             "Synthèse directionnelle des AMM, homologations DM, délais d'évaluation et indicateurs clés",
		LegalReference:       "Gouvernance pharmaceutique industrielle",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "xlsx", "html"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Mensuel / Revue de Direction Trimestrielle",
		IssuerRole:           "Directeur Technique & Direction Affaires Réglementaires",
		Recipient:            "Directeur Général & Conseil d'Administration",
		Description:          "Indicateurs consolidés de performance : délais de traitement moyen ANPP, ratios de complétude, alertes de renouvellement et prévisions de commercialisation.",
	},
	{
		Code:                 "ES-EXP-01",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryPilotageStrategique,
		Title:                "Procès-Verbal de Réversibilité & Manifeste d'Archive (.sanad)",
		Subtitle:             "Inventaire cryptographique des bases relationnelles, coffre-fort CAS et chaînes de hachage",
		LegalReference:       "Norme de Réversibilité Totale & Pérennité Documentaire BPF (CDC Chapitre 14)",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "json"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkOfficial,
		FrequencyTrigger:     "À chaque export de réversibilité ou sauvegarde légale annuelle",
		IssuerRole:           "Directeur Technique & Responsable Informatique (DSI)",
		Recipient:            "Direction Générale / Tiers de Confiance Archivage / Huissier de Justice",
		Description:          "Manifeste technique certifié garantissant l'indépendance de l'établissement et certifiant l'exhaustivité des 19 tables de données et des binaires CAS exportés.",
	},

	// =========================================================================
	// FAMILLE 6 : CONTRÔLE DE CONFORMITÉ, MÉTROLOGIE & LIBÉRATION DE LOT DCE
	// =========================================================================
	{
		Code:                 "ES-DCE-01",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryControleLiberation,
		Title:                "Certificat Officiel de Libération de Lot du Pharmacien DT",
		Subtitle:             "Filiation Lot PF ↔ Lot SA ↔ Lot Semi-fini, agréments et conformité Arrêté du 10 mai 2021",
		LegalReference:       "Arrêté du 10 mai 2021 fixant les conditions et modalités de libération de lot",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "À chaque lot de produit pharmaceutique ou DM libéré pour commercialisation",
		IssuerRole:           "Pharmacien Directeur Technique Libérateur",
		Recipient:            "Direction du Contrôle et de l'Expertise (DCE) / Marché National",
		Description:          "Acte régalien par lequel le Pharmacien DT engage sa responsabilité personnelle sur la conformité analytique et réglementaire du lot avant mise sur le marché.",
	},
	{
		Code:                 "ES-DCE-02",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryControleLiberation,
		Title:                "Formulaires Officiels de Réception des Produits Importés soumis à Libération",
		Subtitle:             "Dossier de réception sous douane : décision d'enregistrement, dédouanement et plans de contrôle",
		LegalReference:       "Arrêté du 14 février 2021 pour PP • Arrêtés 14/02/2021 & 20/02/2025 pour DM",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "À chaque arrivage de marchandise importée sous douane",
		IssuerRole:           "Pharmacien DT Importateur",
		Recipient:            "Bureau de Libération de Lot DCE / Services Douaniers",
		Description:          "Formulaire officiel de notification d'arrivage permettant le prélèvement d'échantillons conservatoires et la vérification de la chaîne logistique d'importation.",
	},
	{
		Code:                 "ES-DCE-03",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryControleLiberation,
		Title:                "Exportateur Automatisé des Canevas Journaliers de Libération ANPP (.xlsx)",
		Subtitle:             "Classeurs normalisés unifiés : PP-DM-Revision, PP-Routine et DM-Non-Homologues",
		LegalReference:       "Circulaire DCE/ANPP relative à la transmission dématérialisée des libérations de lots",
		DefaultFormat:        "xlsx",
		SupportedFormats:     []string{"xlsx", "csv"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkOfficial,
		FrequencyTrigger:     "Quotidien / Hebdomadaire (Obligation légale de déclaration)",
		IssuerRole:           "Responsable Libération de Lot / Pharmacien Délégué",
		Recipient:            "Service Déclarations Journalières DCE ANPP",
		Description:          "Générateur natif de fichiers Excel conformes aux maquettes imposées par la DCE pour l'enregistrement continu des lots libérés en Algérie.",
	},
	{
		Code:                 "ES-DCE-04",
		Kind:                 KindEtatReglementaire,
		Category:             CategoryControleLiberation,
		Title:                "Bordereau de Composition du Dossier de Libération de Lot",
		Subtitle:             "Checklist de contrôle des 8 pièces PDF indispensables avec mention du 1er lot commercialisé",
		LegalReference:       "Ligne directrice DCE-LB-01 ANPP",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Constitution de chaque dossier de libération physique à soumettre à la DCE",
		IssuerRole:           "Responsable Assurance Qualité / BPF",
		Recipient:            "Direction du Contrôle et de l'Expertise (DCE)",
		Description:          "Bordereau récapitulatif vérifiant les 8 éléments probants (CoA PF, CoA PA, fiche fabrication, rapport analytique, etc.) avant délivrance du visa DCE.",
	},
	{
		Code:                 "MT-LCQ-01",
		Kind:                 KindMasterTemplate,
		Category:             CategoryControleLiberation,
		Title:                "Template Maître BPF de Certificat d'Analyse pour Standard de Référence Secondaire",
		Subtitle:             "Traçabilité ISO 17034, FTIR, chromatographie HPLC/UV 3 essais ± incertitude U et teneur en eau",
		LegalReference:       "Pharmacopée Européenne / Algérienne • Conforme canevas TEMPLATE-CA-SR-secondaire.docx",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "docx", "html"},
		RequiresPDTSignature: true,
		IsPubliclyVerifiable: true,
		DefaultWatermark:     domain.WatermarkOfficial,
		FrequencyTrigger:     "Qualification et re-test de tout étalon secondaire de travail au laboratoire",
		IssuerRole:           "Responsable Contrôle Qualité / Responsable Métrologie",
		Recipient:            "Laboratoire Contrôle Qualité (LCQ) & Auditeurs BPF",
		Description:          "Canevas d'analyse analytique de haute précision pour étalons de travail garantissant la traçabilité métrologique par rapport aux standards primaires OMS/EDQM.",
	},

	// =========================================================================
	// MATRICE DES 8 ÉTATS INTERMÉDIAIRES (EI) : PILOTAGE & SUIVI D'AVANCEMENT
	// =========================================================================
	{
		Code:                 "EI-AVT-01",
		Kind:                 KindEtatIntermediaire,
		Category:             CategorySuiviAvancement,
		Title:                "Fiche Navette d'Avancement CTD / DM (Dossier Progress Sheet)",
		Subtitle:             "Vue 360° du montage : % par module M1-M5, statuts Maker-Checker et verrous actifs",
		LegalReference:       "Système interne de pilotage de projet BPF",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html", "json"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Hebdomadaire / Revue de projet",
		IssuerRole:           "Chef de Projet Affaires Réglementaires",
		Recipient:            "Pharmacien Directeur Technique & Comité de Pilotage",
		Description:          "Instrument de liaison hebdomadaire récapitulant l'état d'avancement des sections M1 à M5, les baux d'édition en cours et les pièces en attente de visa.",
	},
	{
		Code:                 "EI-AVT-02",
		Kind:                 KindEtatIntermediaire,
		Category:             CategorySuiviAvancement,
		Title:                "Relevé des Écarts & Pièces Manquantes (Gap Analysis Matrix)",
		Subtitle:             "Inventaire des pièces exigées manquantes ou rejetées, ventilé par service responsable et dates butoirs",
		LegalReference:       "Management Qualité et Gestion des Risques (ICH Q9)",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "xlsx", "html"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "À chaque sprint de montage / Revue mensuelle",
		IssuerRole:           "Responsable Affaires Réglementaires",
		Recipient:            "Rédacteurs (Makers) & Chefs de Départements (RA, QC, QA, RD, PROD)",
		Description:          "Matrice opérationnelle alertant sur les goulots d'étranglement documentaires avant franchissement des jalons critiques de dépôt.",
	},
	{
		Code:                 "EI-AVT-03",
		Kind:                 KindEtatIntermediaire,
		Category:             CategorySuiviAvancement,
		Title:                "Fiche de Revue & d'Annotation Checker (Quality Review Card)",
		Subtitle:             "Relevé d'arbitrage émis lors d'un rejet : écarts constatés vs BPF/ICH et actions correctives",
		LegalReference:       "Atelier Maker-Checker BPF • Ségrégation des tâches",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "À chaque rejet ou demande d'ajustement formulée par un Relecteur",
		IssuerRole:           "Relecteur Qualité (Checker)",
		Recipient:            "Rédacteur (Maker) & Responsable Assurance Qualité",
		Description:          "Fiche d'annotation contradictoire consignant le motif exact du rejet qualité et prescrivant les corrections indispensables pour mise en conformité.",
	},
	{
		Code:                 "EI-CHR-01",
		Kind:                 KindEtatIntermediaire,
		Category:             CategorySuiviAvancement,
		Title:                "Fiche de Suivi des Stabilités Zone IVb (Stability Milestone Card)",
		Subtitle:             "Suivi des échéances de prélèvement en enceinte climatique M0, M3, M6 (seuil légal dépôt), M9, M12, M24 et conformité CoA",
		LegalReference:       "Ligne Directrice ANPP Stabilités Zone Climatique IVb (30°C / 75%% HR)",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "xlsx", "html"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Mensuel / À chaque échéance de prélèvement de stabilité",
		IssuerRole:           "Responsable Stabilités / Contrôle Qualité",
		Recipient:            "Directeur Technique & Responsable R&D Galénique",
		Description:          "Chronomètre qualité suivant les points de contrôle en enceintes climatiques de stabilité et validant le respect du seuil minimum de 6 mois d'études accélérées exigé au dépôt.",
	},
	{
		Code:                 "EI-CHR-02",
		Kind:                 KindEtatIntermediaire,
		Category:             CategorySuiviAvancement,
		Title:                "Rapport d'Étape de l'Étude de Bioéquivalence (BE Study Tracker)",
		Subtitle:             "Suivi des jalons CRO : éthique, phase clinique, dosage bioanalytique BPL et calculs IC 90%%",
		LegalReference:       "Ligne Directrice Bioéquivalence ANPP • Bonnes Pratiques de Laboratoire (BPL)",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Mensuel / À réception d'un livrable d'étape d'un centre d'essais cliniques (CRO)",
		IssuerRole:           "Chef de Projet Galénique / Responsable R&D",
		Recipient:            "Pharmacien Directeur Technique",
		Description:          "Suivi critique des études in vivo comparatives (générique vs princeps) indispensables pour la recevabilité du Module 5 CTD.",
	},
	{
		Code:                 "EI-VIG-01",
		Kind:                 KindEtatIntermediaire,
		Category:             CategorySuiviAvancement,
		Title:                "Fiche d'Instruction des Compléments & Plan d'Action Interne",
		Subtitle:             "Découpage des questions de la lettre de carence ANPP, jalons internes J-15/J-5 et compte à rebours 60j/90j",
		LegalReference:       "Procédure interne de réponse aux demandes de renseignements de l'ANPP",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Dès notification d'une lettre de carence ou demande de compléments ANPP",
		IssuerRole:           "Coordinateur Réglementaire",
		Recipient:            "Équipe Projet Compléments ANPP & Directeur Technique",
		Description:          "Feuille de route interne affectant chaque question soulevée par les experts à un pôle métier avec calendrier impératif de rendu.",
	},
	{
		Code:                 "EI-CHG-01",
		Kind:                 KindEtatIntermediaire,
		Category:             CategorySuiviAvancement,
		Title:                "Fiche d'Évaluation d'Impact Réglementaire (Change Control Assessment)",
		Subtitle:             "Évaluation préalable avant changement usine : rubriques 1-34 LD-DEPP07/V01, conditions et budget",
		LegalReference:       "Système de Contrôle des Changements BPF (Change Control) • ICH Q10",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "html"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "À l'initiation d'un Change Control interne en production ou contrôle qualité",
		IssuerRole:           "Assurance Qualité & Affaires Réglementaires",
		Recipient:            "Comité de Contrôle des Changements (Change Control Board - CCB)",
		Description:          "Analyse prévisionnelle de l'impact réglementaire d'une modification technique sur les autorisations d'AMM détenues.",
	},
	{
		Code:                 "EI-PIP-01",
		Kind:                 KindEtatIntermediaire,
		Category:             CategorySuiviAvancement,
		Title:                "Point de Situation Hebdomadaire du Pipeline (Weekly Pipeline Digest)",
		Subtitle:             "Synthèse Comité de Direction : dossiers en attente ANPP, délais écoulés vs cibles 120j/180j et alertes",
		LegalReference:       "Gouvernance du portefeuille et planification industrielle",
		DefaultFormat:        "pdf",
		SupportedFormats:     []string{"pdf", "xlsx", "html"},
		RequiresPDTSignature: false,
		IsPubliclyVerifiable: false,
		DefaultWatermark:     domain.WatermarkDraft,
		FrequencyTrigger:     "Hebdomadaire (Comité de Direction)",
		IssuerRole:           "Responsable Affaires Réglementaires",
		Recipient:            "Directeur Technique & Direction Générale",
		Description:          "Vue synthétique macroscopique du portefeuille des dépôts en cours, des délais administratifs consommés et des actions de relance requises.",
	},
}

// GetAllReportDefinitions retourne la liste complète des 28 états documentaires
func GetAllReportDefinitions() []ReportDefinition {
	result := make([]ReportDefinition, len(MasterRegistry))
	copy(result, MasterRegistry)
	return result
}

// GetReportDefinition recherche une définition par son code ou alias réglementaire (insensible à la casse)
func GetReportDefinition(code string) (ReportDefinition, bool) {
	upperCode := strings.ToUpper(strings.TrimSpace(code))

	// Résolution des alias usuels ANPP
	switch upperCode {
	case "1.P.8", "1P8", "ENG-1P8":
		upperCode = "ES-ENG-1P8"
	case "1.P.9", "1P9", "ENG-1P9":
		upperCode = "ES-ENG-1P9"
	case "1.H.10", "1H10", "ENG-1H10":
		upperCode = "ES-ENG-1H10"
	case "BV1", "BV-1":
		upperCode = "ES-FIN-01"
	case "BV2", "BV-2":
		upperCode = "ES-FIN-02"
	case "BV-DM":
		upperCode = "ES-FIN-01"
	}

	for _, def := range MasterRegistry {
		if strings.EqualFold(def.Code, upperCode) {
			return def, true
		}
	}
	return ReportDefinition{}, false
}

// GetReportsByCategory filtre les états documentaires par catégorie
func GetReportsByCategory(category ReportCategory) []ReportDefinition {
	var filtered []ReportDefinition
	for _, def := range MasterRegistry {
		if def.Category == category {
			filtered = append(filtered, def)
		}
	}
	return filtered
}

// GetReportsByKind filtre les états documentaires par nature légale (ES, EI, MT)
func GetReportsByKind(kind ReportKind) []ReportDefinition {
	var filtered []ReportDefinition
	for _, def := range MasterRegistry {
		if def.Kind == kind {
			filtered = append(filtered, def)
		}
	}
	return filtered
}
