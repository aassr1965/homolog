package domain

import (
	"time"

	"github.com/google/uuid"
)

// ImportLotRelease représente une demande / bordereau de libération de lot importé (DM ou PP - DCE ANPP)
type ImportLotRelease struct {
	ID                    uuid.UUID  `json:"id"`
	OrganizationID        uuid.UUID  `json:"organization_id"`
	ProductKind           string     `json:"product_kind"` // "dm" ou "pp"
	MedicalDeviceID       *uuid.UUID `json:"medical_device_id,omitempty"`
	MedicineID            *uuid.UUID `json:"medicine_id,omitempty"`
	CommercialName        string     `json:"commercial_name"`
	Designation           string     `json:"designation"`
	FormAspect            string     `json:"form_aspect"`
	LotNumber             string     `json:"lot_number"`
	SerialNumber          string     `json:"serial_number"`
	DDF                   *time.Time `json:"ddf,omitempty"`
	DDP                   *time.Time `json:"ddp,omitempty"`
	SalesUnits            string     `json:"sales_units"`
	HomologationOrDERef   string     `json:"homologation_or_de_ref"`
	DMClass               string     `json:"dm_class"`
	HolderOrExploiter     string     `json:"holder_or_exploiter"`
	ImportProgramAttached bool       `json:"import_program_attached"`
	ProductionSite        string     `json:"production_site"`
	QCSite                string     `json:"qc_site"`
	ReleaseSite           string     `json:"release_site"`
	CustomsDeclNumber     string     `json:"customs_decl_number"`
	InvoiceNumber         string     `json:"invoice_number"`
	OriginCertAttached    bool       `json:"origin_cert_attached"`
	TechFormAttached      bool       `json:"tech_form_attached"`
	SamplingPlanAttached  bool       `json:"sampling_plan_attached"`
	IsRelianceLab         bool       `json:"is_reliance_lab"`
	RelianceLabAuthority  string     `json:"reliance_lab_authority,omitempty"`
	RelianceLabReference  string     `json:"reliance_lab_reference,omitempty"`
	Status                string     `json:"status"`
	CreatedBy             *uuid.UUID `json:"created_by,omitempty"`
	SignedByPDTID         *uuid.UUID `json:"signed_by_pdt_id,omitempty"`
	SignedAt              *time.Time `json:"signed_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// CreateImportLotReleaseRequest payload de création d'un dossier de libération de lot importé
type CreateImportLotReleaseRequest struct {
	ProductKind           string     `json:"product_kind"`
	MedicalDeviceID       *uuid.UUID `json:"medical_device_id,omitempty"`
	MedicineID            *uuid.UUID `json:"medicine_id,omitempty"`
	CommercialName        string     `json:"commercial_name"`
	Designation           string     `json:"designation"`
	FormAspect            string     `json:"form_aspect"`
	LotNumber             string     `json:"lot_number"`
	SerialNumber          string     `json:"serial_number"`
	DDF                   string     `json:"ddf"`
	DDP                   string     `json:"ddp"`
	SalesUnits            string     `json:"sales_units"`
	HomologationOrDERef   string     `json:"homologation_or_de_ref"`
	DMClass               string     `json:"dm_class"`
	HolderOrExploiter     string     `json:"holder_or_exploiter"`
	ImportProgramAttached bool       `json:"import_program_attached"`
	ProductionSite        string     `json:"production_site"`
	QCSite                string     `json:"qc_site"`
	ReleaseSite           string     `json:"release_site"`
	CustomsDeclNumber     string     `json:"customs_decl_number"`
	InvoiceNumber         string     `json:"invoice_number"`
	OriginCertAttached    bool       `json:"origin_cert_attached"`
	TechFormAttached      bool       `json:"tech_form_attached"`
	SamplingPlanAttached  bool       `json:"sampling_plan_attached"`
	IsRelianceLab         bool       `json:"is_reliance_lab"`
	RelianceLabAuthority  string     `json:"reliance_lab_authority,omitempty"`
	RelianceLabReference  string     `json:"reliance_lab_reference,omitempty"`
	PDTOverrideReason     string     `json:"pdt_override_reason,omitempty"`
}
