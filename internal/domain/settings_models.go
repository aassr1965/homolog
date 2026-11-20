package domain

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Constantes pour les clés de settings réglementaires du tenant (CDC Section 6.4.1)
const (
	// Circuit Médicaments
	SettingFinPaymentModeMedicine = "FIN_PAYMENT_MODE_MEDICINE"
	SettingFinTreasuryAccountMed  = "FIN_TREASURY_ACCOUNT_MEDICINE"
	SettingFinBeneficiaryOrderMed = "FIN_BENEFICIARY_ORDER_MEDICINE"

	// Circuit Dispositifs Médicaux
	SettingFinBankNameDM = "FIN_BANK_NAME_DM"
	SettingFinBankRibDM  = "FIN_BANK_RIB_DM"

	// Barèmes Médicaments Décret 21-82
	SettingFeeMedicineNI  = "FEE_MEDICINE_NI_DZD"
	SettingFeeMedicineNF  = "FEE_MEDICINE_NF_DZD"
	SettingFeeMedicineRI  = "FEE_MEDICINE_RI_DZD"
	SettingFeeMedicineMI  = "FEE_MEDICINE_MI_DZD"
	SettingFeeMedicineMin = "FEE_MEDICINE_MIN_DZD"

	// Barèmes Dispositifs Médicaux Ordonnance 21-07 Art. 31
	SettingFeeDMHomologation = "FEE_DM_HOMOLOGATION_DZD"
	SettingFeeDMRenewal      = "FEE_DM_RENEWAL_DZD"
	SettingFeeDMModification = "FEE_DM_MODIFICATION_DZD"
	SettingFeeDMTransfer     = "FEE_DM_TRANSFER_DZD"

	// Délais Légaux de Carence
	SettingRegDeficiencyStandardDays = "REG_DEFICIENCY_DEADLINE_STANDARD_DAYS"
	SettingRegDeficiencyUrgentDays   = "REG_DEFICIENCY_DEADLINE_URGENT_DAYS"

	// Guichets & Institutions ANPP
	SettingInstANPPDeppAddress = "INST_ANPP_DEPP_ADDRESS"
	SettingInstANPPDdmAddress  = "INST_ANPP_DDM_ADDRESS"
	SettingInstANPPDceAddress  = "INST_ANPP_DCE_ADDRESS"
)

// Catégories de settings
type TenantSettingCategory string

const (
	SettingCategoryFinanceMedicine    TenantSettingCategory = "finance_medicine"
	SettingCategoryFinanceDM          TenantSettingCategory = "finance_dm"
	SettingCategoryFeeSchedule        TenantSettingCategory = "fee_schedule"
	SettingCategoryRegulatoryDeadline TenantSettingCategory = "regulatory_deadlines"
	SettingCategoryInstitutionsANPP   TenantSettingCategory = "institutions_anpp"
)

// Statuts des settings
type TenantSettingStatus string

const (
	SettingStatusActive            TenantSettingStatus = "ACTIVE"
	SettingStatusPendingApprovalDT TenantSettingStatus = "PENDING_APPROVAL_DT"
	SettingStatusHistorical        TenantSettingStatus = "HISTORICAL"
)

// TenantRegulatorySetting représente une variable réglementaire administrable par organisation
type TenantRegulatorySetting struct {
	ID                  uuid.UUID             `json:"id"`
	OrganizationID      uuid.UUID             `json:"organization_id"`
	SettingKey          string                `json:"setting_key"`
	SettingValue        string                `json:"setting_value"`
	DataType            string                `json:"data_type"` // string, number, integer, boolean, json
	Category            TenantSettingCategory `json:"category"`
	LabelFR             string                `json:"label_fr"`
	LabelAR             *string               `json:"label_ar,omitempty"`
	Description         *string               `json:"description,omitempty"`
	LegalReference      *string               `json:"legal_reference,omitempty"`
	ValidFrom           time.Time             `json:"valid_from"`
	ValidTo             *time.Time            `json:"valid_to,omitempty"`
	IsActive            bool                  `json:"is_active"`
	PendingValue        *string               `json:"pending_value,omitempty"`
	PendingReason       *string               `json:"pending_reason,omitempty"`
	PendingProposedBy   *uuid.UUID            `json:"pending_proposed_by,omitempty"`
	PendingProposerName *string               `json:"pending_proposer_name,omitempty"`
	Status              TenantSettingStatus   `json:"status"`
	ApprovedBy          *uuid.UUID            `json:"approved_by,omitempty"`
	ApprovedByName      *string               `json:"approved_by_name,omitempty"`
	ApprovedAt          *time.Time            `json:"approved_at,omitempty"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
}

// Helpers pour récupérer les valeurs typées avec fallback sécurisé
func (s *TenantRegulatorySetting) Float64(defaultVal float64) float64 {
	if s == nil || s.SettingValue == "" {
		return defaultVal
	}
	val, err := strconv.ParseFloat(s.SettingValue, 64)
	if err != nil {
		return defaultVal
	}
	return val
}

func (s *TenantRegulatorySetting) Int(defaultVal int) int {
	if s == nil || s.SettingValue == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s.SettingValue)
	if err != nil {
		return defaultVal
	}
	return val
}

func (s *TenantRegulatorySetting) String(defaultVal string) string {
	if s == nil || s.SettingValue == "" {
		return defaultVal
	}
	return s.SettingValue
}

// UpdateTenantSettingRequest représente la requête de modification d'un paramètre
type UpdateTenantSettingRequest struct {
	NewValue          string `json:"new_value"`
	RegulatoryReason  string `json:"regulatory_reason"`
	RequireDTApproval bool   `json:"require_dt_approval"`
	TOTPCode          string `json:"totp_code,omitempty"` // Si le Pharmacien DT approuve directement
}

// ApproveTenantSettingRequest représente l'approbation formelle par le Pharmacien DT
type ApproveTenantSettingRequest struct {
	TOTPCode    string `json:"totp_code,omitempty"`
	SolemnCheck bool   `json:"solemn_check"` // Attestation sur l'honneur
}
