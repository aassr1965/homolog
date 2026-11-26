package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ANPPFormDraft représente un brouillon éditable persistant pour compléter un formulaire ANPP.
type ANPPFormDraft struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	FormKind       string          `json:"form_kind"` // structure_prix, ca_sr_secondaire, dvtiv_qppv, dce_sa_complement
	EntityID       uuid.UUID       `json:"entity_id"`
	Payload        json.RawMessage `json:"payload"`
	Status         string          `json:"status"` // toujours 'brouillon' (éditable)
	UpdatedBy      *uuid.UUID      `json:"updated_by,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// StructurePrixDraft données du formulaire Structure de prix à la fabrication locale.
type StructurePrixDraft struct {
	SubstanceActive   string `json:"substance_active"`
	MatieresImportees string `json:"matieres_importees"`
	MatieresLocales   string `json:"matieres_locales"`
	ServicesImportes  string `json:"services_importes"`
	ServicesLocaux    string `json:"services_locaux"`
	FraisPersonnel    string `json:"frais_personnel"`
	Taxes             string `json:"taxes"`
	FraisFinanciers   string `json:"frais_financiers"`
	Amortissements    string `json:"amortissements"`
	AutresCharges     string `json:"autres_charges"`
	PrixRevient       string `json:"prix_revient"`
	PCSU              string `json:"pcsu"`
	TauxExportation   string `json:"taux_exportation"`
	TauxIntegration   string `json:"taux_integration"`
	TauxChange        string `json:"taux_change"`
}

// CASRSecondaireDraft données du prototype Certificat d'analyse SR secondaire.
type CASRSecondaireDraft struct {
	StandardName       string `json:"standard_name"`
	LotCode            string `json:"lot_code"`
	ExpiryDate         string `json:"expiry_date"`
	Manufacturer       string `json:"manufacturer"`
	MethodRef          string `json:"method_ref"`
	Weighing           string `json:"weighing"`
	StorageConditions  string `json:"storage_conditions"`
	SafetyInstructions string `json:"safety_instructions"`
	PrimaryLotNumber   string `json:"primary_lot_number"`
	StabilityValidated bool   `json:"stability_validated"`
	EquipQualified     bool   `json:"equip_qualified"`
	DescSpec           string `json:"desc_spec"`
	DescResult         string `json:"desc_result"`
	IdentSpec          string `json:"ident_spec"`
	IdentResult        string `json:"ident_result"`
	AssaySpec          string `json:"assay_spec"`
	AssayResult        string `json:"assay_result"`
	PuritySpec         string `json:"purity_spec"`
	PurityResult       string `json:"purity_result"`
	WaterSpec          string `json:"water_spec"`
	WaterResult        string `json:"water_result"`
	PreparedBy         string `json:"prepared_by"`
	VerifiedBy         string `json:"verified_by"`
	ApprovedBy         string `json:"approved_by"`
}

// DVTIVQPPVDraft données complémentaires QPPV / PSMF pour le formulaire DVTIV.
type DVTIVQPPVDraft struct {
	QPPVName        string `json:"qppv_name"`
	QPPVEmail       string `json:"qppv_email"`
	QPPVPhone       string `json:"qppv_phone"`
	QPPVDeputy      string `json:"qppv_deputy"`
	HasPSMF         bool   `json:"has_psmf"`
	HasPSMFNational bool   `json:"has_psmf_national"`
}

// DCESAComplementDraft données complémentaires Substance Active & commercialisation pour déclarations DCE.
type DCESAComplementDraft struct {
	SAName            string `json:"sa_name"`
	SALotNumber       string `json:"sa_lot_number"`
	SADDF             string `json:"sa_ddf"`
	SADDP             string `json:"sa_ddp"`
	SAManufacturer    string `json:"sa_manufacturer"`
	FirstMarketedLot  bool   `json:"first_marketed_lot"`
	ChannelAllocation string `json:"channel_allocation"` // Hopital, Officine, Hopital / Officine
}
