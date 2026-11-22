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

// Constantes pour l'évaluation de recevabilité DM (Arrêté du 10 mai 2021)
const (
	RecevabiliteStatusConforme    = "conforme"
	RecevabiliteStatusNonConforme = "non_conforme"
	RecevabiliteStatusReserve     = "avec_reserve"

	RecevabiliteDecisionComplet     = "complet_et_recevable"
	RecevabiliteDecisionReserves    = "incomplet_avec_reserves"
	RecevabiliteDecisionIrrecevable = "irrecevable"
)

// DMRecevabiliteEvaluation représente l'évaluation en 13 points de contrôle pour l'acte officiel ES-REG-03-DM
type DMRecevabiliteEvaluation struct {
	ID                    uuid.UUID `json:"id"`
	OrganizationID        uuid.UUID `json:"organization_id"`
	DossierID             uuid.UUID `json:"dossier_id"`
	MedicalDeviceID       uuid.UUID `json:"medical_device_id"`
	DepositOrderNumber    string    `json:"deposit_order_number"`
	DepositDate           time.Time `json:"deposit_date"`
	ElectronicCopiesCount int       `json:"electronic_copies_count"`
	ElectronicMediaType   string    `json:"electronic_media_type"`

	Item1Status  string `json:"item_1_status"`  // Lettre de demande 2 ex. + descriptif
	Item2Status  string `json:"item_2_status"`  // Formulaire 6 sections (ES-REG-02-DM)
	Item3Status  string `json:"item_3_status"`  // Agrément MIPH fabricant/exploitant
	Item4Status  string `json:"item_4_status"`  // Justification redevance BNA (300k DA)
	Item5Status  string `json:"item_5_status"`  // CLV / FSC pays d'origine
	Item6Status  string `json:"item_6_status"`  // Certificat CE & Déclaration
	Item7Status  string `json:"item_7_status"`  // BPF / ISO 13485 PF & LCQ
	Item8Status  string `json:"item_8_status"`  // Certifications matières premières
	Item9Status  string `json:"item_9_status"`  // Matériovigilance & PGR
	Item10Status string `json:"item_10_status"` // Partie III Stabilité & Stérilisation
	Item11Status string `json:"item_11_status"` // Partie IV ISO 14971 & ISO 10993
	Item12Status string `json:"item_12_status"` // Partie V Évaluation clinique ISO 14155
	Item13Status string `json:"item_13_status"` // 04 Exemplaires électroniques scellés

	ConformityScore  int        `json:"conformity_score"` // 0 à 13
	Decision         string     `json:"decision"`
	ReservesNotes    string     `json:"reserves_notes,omitempty"`
	ANPPAgentName    string     `json:"anpp_agent_name"`
	PDTUserID        *uuid.UUID `json:"pdt_user_id,omitempty"`
	PDTSignedAt      *time.Time `json:"pdt_signed_at,omitempty"`
	PDTSignatureSeal string     `json:"pdt_signature_seal,omitempty"`
	SHA256Hash       string     `json:"sha256_hash,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Données enrichies (jointures)
	MedicalDevice *MedicalDevice `json:"medical_device,omitempty"`
	DossierCode   string         `json:"dossier_code,omitempty"`
	PDTFullName   string         `json:"pdt_full_name,omitempty"`
}

// CalculateConformity recalcule le score de conformité sur 13 points et la décision
func (e *DMRecevabiliteEvaluation) CalculateConformity() {
	score := 0
	items := []string{
		e.Item1Status, e.Item2Status, e.Item3Status, e.Item4Status,
		e.Item5Status, e.Item6Status, e.Item7Status, e.Item8Status,
		e.Item9Status, e.Item10Status, e.Item11Status, e.Item12Status,
		e.Item13Status,
	}

	hasNonConforme := false
	hasReserve := false

	for _, s := range items {
		st := strings.ToLower(strings.TrimSpace(s))
		if st == RecevabiliteStatusConforme {
			score++
		} else if st == RecevabiliteStatusReserve {
			hasReserve = true
		} else {
			hasNonConforme = true
		}
	}

	e.ConformityScore = score

	if hasNonConforme || score < 10 {
		e.Decision = RecevabiliteDecisionIrrecevable
	} else if hasReserve || score < 13 {
		e.Decision = RecevabiliteDecisionReserves
	} else {
		e.Decision = RecevabiliteDecisionComplet
	}
}

// GenerateSeal génère le sceau cryptographique 21 CFR Part 11
func (e *DMRecevabiliteEvaluation) GenerateSeal(userEmail, decisionMinistérielle string) string {
	payload := fmt.Sprintf("EVAL_DM:%s|ORDER:%s|SCORE:%d/13|DECISION:%s|PDT:%s|DECISION_MIPH:%s|DATE:%s",
		e.ID.String(),
		e.DepositOrderNumber,
		e.ConformityScore,
		e.Decision,
		userEmail,
		decisionMinistérielle,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	hash := sha256.Sum256([]byte(payload))
	seal := fmt.Sprintf("DM-SEAL-REC-%s", hex.EncodeToString(hash[:16]))
	e.PDTSignatureSeal = seal
	return seal
}

// Constantes pour le statut de régularisation transitoire des DM
const (
	DMRegStatusCommercialiseDerogatoire = "commercialise_derogatoire"
	DMRegStatusDossierDeposeDDM         = "dossier_depose_ddm"
	DMRegStatusEnInstruction            = "en_instruction"
	DMRegStatusHomologueRegularise      = "homologue_regularise"
	DMRegStatusRetire                   = "retire"
)

// DMRegularization représente le suivi des DM non homologués sous mesures transitoires
type DMRegularization struct {
	ID                         uuid.UUID  `json:"id"`
	OrganizationID             uuid.UUID  `json:"organization_id"`
	MedicalDeviceID            uuid.UUID  `json:"medical_device_id"`
	RegularizationCode         string     `json:"regularization_code"`
	DerogationNumber           string     `json:"derogation_number"`
	InitialMarketDate          time.Time  `json:"initial_market_date"`
	Status                     string     `json:"status"`
	DDMAcknowledgementNumber   string     `json:"ddm_acknowledgement_number,omitempty"`
	DDMDepositDate             *time.Time `json:"ddm_deposit_date,omitempty"`
	LegalDeadline              time.Time  `json:"legal_deadline"`
	QuarterlyVolumeImported    float64    `json:"quarterly_volume_imported"`
	QuarterlyVolumeDistributed float64    `json:"quarterly_volume_distributed"`
	HospitalDistributionShare  int        `json:"hospital_distribution_share"`
	RetailDistributionShare    int        `json:"retail_distribution_share"`
	LastDeclarationDate        time.Time  `json:"last_declaration_date"`
	Observations               string     `json:"observations,omitempty"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`

	// Champs calculés / jointures
	DaysRemaining int            `json:"days_remaining"`
	UrgencyLevel  string         `json:"urgency_level"` // "critique", "urgent", "alerte", "dans_les_delais", "expire", "regularise"
	MedicalDevice *MedicalDevice `json:"medical_device,omitempty"`
}

// CalculateUrgency calcule le nombre de jours restants et le niveau d'urgence
func (r *DMRegularization) CalculateUrgency() {
	if r.Status == DMRegStatusHomologueRegularise {
		r.UrgencyLevel = "regularise"
		r.DaysRemaining = 0
		return
	}

	days := int(time.Until(r.LegalDeadline).Hours() / 24)
	r.DaysRemaining = days

	if days < 0 {
		r.UrgencyLevel = "expire"
	} else if days <= 30 {
		r.UrgencyLevel = "critique"
	} else if days <= 60 {
		r.UrgencyLevel = "urgent"
	} else if days <= 90 {
		r.UrgencyLevel = "alerte"
	} else {
		r.UrgencyLevel = "dans_les_delais"
	}
}

// Validate valide les attributs d'une régularisation DM
func (r *DMRegularization) Validate() error {
	if strings.TrimSpace(r.RegularizationCode) == "" {
		return errors.New("le code de dossier de régularisation est obligatoire")
	}
	if r.MedicalDeviceID == uuid.Nil {
		return errors.New("l'identifiant du dispositif médical est obligatoire")
	}
	if r.HospitalDistributionShare+r.RetailDistributionShare != 100 {
		return errors.New("la répartition Hôpital (%) + Officine (%) doit totaliser exactement 100%")
	}
	switch r.Status {
	case DMRegStatusCommercialiseDerogatoire, DMRegStatusDossierDeposeDDM,
		DMRegStatusEnInstruction, DMRegStatusHomologueRegularise, DMRegStatusRetire:
		// valide
	default:
		return errors.New("statut de régularisation non reconnu")
	}
	return nil
}

// DMRegularizationStats consolide les métriques du tableau de bord de régularisation
type DMRegularizationStats struct {
	TotalDevicesCount      int     `json:"total_devices_count"`
	DerogatoireCount       int     `json:"derogatoire_count"`
	DeposeDDMCount         int     `json:"depose_ddm_count"`
	EnInstructionCount     int     `json:"en_instruction_count"`
	RegulariseCount        int     `json:"regularise_count"`
	RetireCount            int     `json:"retire_count"`
	CriticalDeadlineCount  int     `json:"critical_deadline_count"` // J <= 30
	UrgentDeadlineCount    int     `json:"urgent_deadline_count"`   // J <= 60
	AlertDeadlineCount     int     `json:"alert_deadline_count"`    // J <= 90
	TotalVolumeDistributed float64 `json:"total_volume_distributed"`
}
