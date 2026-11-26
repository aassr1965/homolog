package domain

import (
	"time"

	"github.com/google/uuid"
)

// RegulatoryCategory domaine d'application du texte légal
type RegulatoryCategory string

const (
	RegCatEnregistrementAMM RegulatoryCategory = "enregistrement_amm"   // Décret 20-325, CTD, AMM
	RegCatDispositifsMed    RegulatoryCategory = "dispositifs_medicaux" // Décret 20-324, Homologation DM
	RegCatRedevancesTarifs  RegulatoryCategory = "redevances_tarifs"    // Arrêtés interministériels tarifs ANPP
	RegCatBPFInspection     RegulatoryCategory = "bpf_inspection"       // Arrêté 16/02/2021 BPF
	RegCatPharmacovigilance RegulatoryCategory = "pharmacovigilance"    // PV, Matériovigilance
	RegCatEssaisCliniquesBE RegulatoryCategory = "essais_cliniques_be"  // Bioéquivalence, Essais cliniques
)

// ImpactLevel criticité de l'impact réglementaire sur les dossiers
type ImpactLevel string

const (
	ImpactLevelCritique ImpactLevel = "critique" // Impacte immédiatement les dossiers en cours ou suspendus
	ImpactLevelMajeur   ImpactLevel = "majeur"   // Mise en conformité exigée sous 3 à 6 mois
	ImpactLevelMineur   ImpactLevel = "mineur"   // Aménagement technique ou précision d'application
)

// RegulatoryComplianceStatus statut de conformité de l'établissement vis-à-vis du texte
type RegulatoryComplianceStatus string

const (
	ComplianceStatusEnVigueur        RegulatoryComplianceStatus = "en_vigueur"
	ComplianceStatusEnCoursAnalyse   RegulatoryComplianceStatus = "en_cours_analyse"
	ComplianceStatusPlanActionRequis RegulatoryComplianceStatus = "plan_action_requis"
	ComplianceStatusConforme         RegulatoryComplianceStatus = "conforme"
)

// RegulatoryWatchItem texte officiel du corpus juridique pharmaceutique algérien
type RegulatoryWatchItem struct {
	ID                uuid.UUID                  `json:"id"`
	OrganizationID    uuid.UUID                  `json:"organization_id"`
	TitleFR           string                     `json:"title_fr"`
	TitleAR           string                     `json:"title_ar,omitempty"`
	LegalReference    string                     `json:"legal_reference"`         // ex: Décret Exécutif N° 20-325
	JORADPNumber      string                     `json:"joradp_number,omitempty"` // ex: JORADP N° 78 du 28/11/2020
	PublishedDate     time.Time                  `json:"published_date"`
	EffectiveDate     time.Time                  `json:"effective_date"`
	Category          RegulatoryCategory         `json:"category"`
	ImpactLevel       ImpactLevel                `json:"impact_level"`
	SummaryFR         string                     `json:"summary_fr"`
	ImpactAnalysis    string                     `json:"impact_analysis,omitempty"`
	ComplianceStatus  RegulatoryComplianceStatus `json:"compliance_status"`
	SourceURL         string                     `json:"source_url,omitempty"`
	DocumentFilename  string                     `json:"document_filename,omitempty"`
	DocumentSHA256    string                     `json:"document_sha256,omitempty"`
	DocumentSizeBytes int64                      `json:"document_size_bytes,omitempty"`
	HasEmbeddedPDF    bool                       `json:"has_embedded_pdf"`
	CreatedBy         *uuid.UUID                 `json:"created_by,omitempty"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
}

// IsEffective vérifie si le texte est entré en vigueur
func (r *RegulatoryWatchItem) IsEffective() bool {
	return !time.Now().UTC().Before(r.EffectiveDate)
}

// RequiresActionPlan indique si un plan d'action d'urgence est nécessaire
func (r *RegulatoryWatchItem) RequiresActionPlan() bool {
	return r.ImpactLevel == ImpactLevelCritique && r.ComplianceStatus != ComplianceStatusConforme
}

// FormatCategoryLabel libellé lisible en français
func (r *RegulatoryWatchItem) FormatCategoryLabel() string {
	switch r.Category {
	case RegCatEnregistrementAMM:
		return "Enregistrement & AMM (Médicaments)"
	case RegCatDispositifsMed:
		return "Homologation Dispositifs Médicaux"
	case RegCatRedevancesTarifs:
		return "Barèmes & Redevances ANPP"
	case RegCatBPFInspection:
		return "Bonnes Pratiques de Fabrication (BPF)"
	case RegCatPharmacovigilance:
		return "Pharmacovigilance & Matériovigilance"
	case RegCatEssaisCliniquesBE:
		return "Essais Cliniques & Bioéquivalence"
	default:
		return string(r.Category)
	}
}
