package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserRole defines the 8 strict roles in the pharmaceutical platform
type UserRole string

const (
	RoleSuperAdminEditeur       UserRole = "super_admin_editeur"
	RoleAdminSystemeLocal       UserRole = "admin_systeme_local"
	RoleAdminLaboratoire        UserRole = "admin_laboratoire"
	RoleResponsableAffairesReg  UserRole = "responsable_affaires_reglementaires"
	RoleDirecteurTechnique      UserRole = "directeur_technique"
	RoleContributeurTechnique   UserRole = "contributeur_technique"
	RoleConsultantExterne       UserRole = "consultant_externe"
	RoleLecteur                 UserRole = "lecteur"
	RoleAuditeur                UserRole = "auditeur"
)

// ProductType differentiates medicines and medical devices
type ProductType string

const (
	ProductTypeMedicament       ProductType = "medicament"
	ProductTypeDispositifMedical ProductType = "dispositif_medical"
)

// ProcedureType defines regulatory submission procedures in Algeria
type ProcedureType string

const (
	ProcEnregistrementInitial ProcedureType = "enregistrement_initial"
	ProcPreSoumission         ProcedureType = "pre_soumission"
	ProcRenouvellement        ProcedureType = "renouvellement"
	ProcModification          ProcedureType = "modification"
	ProcATU                   ProcedureType = "atu"
	ProcHomologationDM        ProcedureType = "homologation_dm"
	ProcRegularisationDM      ProcedureType = "regularisation_dm"
	ProcHomologationExportDM  ProcedureType = "homologation_export_dm"
)

// DossierStatus defines the workflow status of an ANPP dossier
type DossierStatus string

const (
	DossierStatusBrouillon          DossierStatus = "brouillon"
	DossierStatusConstitution       DossierStatus = "constitution"
	DossierStatusValidationDT       DossierStatus = "validation_dt"
	DossierStatusDeposeANPP         DossierStatus = "depose_anpp"
	DossierStatusEvaluation         DossierStatus = "evaluation"
	DossierStatusAttenteComplements DossierStatus = "attente_complements"
	DossierStatusAvisCommission     DossierStatus = "avis_commission"
	DossierStatusDecisionEnregistre DossierStatus = "decision_enregistre"
	DossierStatusDecisionRefuse     DossierStatus = "decision_refuse"
	DossierStatusArchive            DossierStatus = "archive"
)

// GEDDocumentStatus defines the GxP lifecycle status of an EDMS document
type GEDDocumentStatus string

const (
	GEDStatusBrouillon      GEDDocumentStatus = "brouillon"
	GEDStatusEnRevue        GEDDocumentStatus = "en_revue"
	GEDStatusEnApprobationDT GEDDocumentStatus = "en_approbation_dt"
	GEDStatusEnVigueur      GEDDocumentStatus = "en_vigueur"
	GEDStatusObsolete       GEDDocumentStatus = "obsolete"
	GEDStatusArchive        GEDDocumentStatus = "archive"
)

// Organization represents a pharmaceutical laboratory (Tenant)
type Organization struct {
	ID                    uuid.UUID `json:"id"`
	Slug                  string    `json:"slug"`
	LegalName             string    `json:"legal_name"`
	MIPHLicenseNumber     string    `json:"miph_license_number"`
	TechnicalDirectorName string    `json:"technical_director_name"`
	Address               string    `json:"address"`
	Phone                 string    `json:"phone"`
	Email                 string    `json:"email"`
	StorageQuotaBytes     int64     `json:"storage_quota_bytes"`
	StorageUsedBytes      int64     `json:"storage_used_bytes"`
	IsActive              bool      `json:"is_active"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// User represents an individual identified user (nominative)
type User struct {
	ID                   uuid.UUID  `json:"id"`
	OrganizationID       uuid.UUID  `json:"organization_id"`
	Email                string     `json:"email"`
	PasswordHash         string     `json:"-"`
	FirstName            string     `json:"first_name"`
	LastName             string     `json:"last_name"`
	Role                 UserRole   `json:"role"`
	TOTPSecret           string     `json:"-"`
	TOTPEnabled          bool       `json:"totp_enabled"`
	RecoveryCodes        []string   `json:"-"`
	FailedLoginAttempts  int        `json:"failed_login_attempts"`
	LockedUntil          *time.Time `json:"locked_until,omitempty"`
	ContractExpiresAt    *time.Time `json:"contract_expires_at,omitempty"`
	IsActive             bool       `json:"is_active"`
	LastLoginAt          *time.Time `json:"last_login_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// Session represents a server-side tracked session cookie
type Session struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	TokenHash      string    `json:"-"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	ExpiresAt      time.Time `json:"expires_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// Product represents a Medicine or Medical Device
type Product struct {
	ID                     uuid.UUID   `json:"id"`
	OrganizationID         uuid.UUID   `json:"organization_id"`
	ProductType            ProductType `json:"product_type"`
	BrandNameFR            string      `json:"brand_name_fr"`
	BrandNameAR            string      `json:"brand_name_ar"`
	DCI                    string      `json:"dci,omitempty"`
	Dosage                 string      `json:"dosage,omitempty"`
	PharmaceuticalForm     string      `json:"pharmaceutical_form,omitempty"`
	AdministrationRoute    string      `json:"administration_route,omitempty"`
	DMClass                string      `json:"dm_class,omitempty"` // "I", "IIa", "IIb", "III", "DMDIV"
	IsSterile              bool        `json:"is_sterile"`
	DMClassificationRule   string      `json:"dm_classification_rule,omitempty"`
	ManufacturerName       string      `json:"manufacturer_name"`
	MarketingStatus        string      `json:"marketing_status"`
	IsHorsNomenclature     bool        `json:"is_hors_nomenclature"`
	CreatedAt              time.Time   `json:"created_at"`
	UpdatedAt              time.Time   `json:"updated_at"`
}

// GEDDocument represents an entry in the pharmaceutical EDMS
type GEDDocument struct {
	ID             uuid.UUID         `json:"id"`
	OrganizationID uuid.UUID         `json:"organization_id"`
	TaxonomyPath   string            `json:"taxonomy_path"` // e.g. "01_Affaires_Reglementaires/01_04_Dossiers_CTD_Reference"
	TitleFR        string            `json:"title_fr"`
	TitleAR        string            `json:"title_ar"`
	DocumentCode   string            `json:"document_code"`
	Status         GEDDocumentStatus `json:"status"`
	CurrentVersion string            `json:"current_version"` // e.g. "v1.0"
	ExpiryDate     *time.Time        `json:"expiry_date,omitempty"`
	CreatedBy      *uuid.UUID        `json:"created_by,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// GEDVersion represents an immutable version of a document with its physical file
type GEDVersion struct {
	ID                   uuid.UUID  `json:"id"`
	DocumentID           uuid.UUID  `json:"document_id"`
	VersionLabel         string     `json:"version_label"` // "v0.1", "v1.0", "v2.0"
	RevisionSummary      string     `json:"revision_summary"`
	FileSHA256           string     `json:"file_sha256"`
	FileSizeBytes        int64      `json:"file_size_bytes"`
	MimeType             string     `json:"mime_type"`
	OriginalFilename     string     `json:"original_filename"`
	StoragePath          string     `json:"storage_path"`
	IsSignedDT           bool       `json:"is_signed_dt"`
	SignedByDTID         *uuid.UUID `json:"signed_by_dt_id,omitempty"`
	SignedAt             *time.Time `json:"signed_at,omitempty"`
	SignatureDeclaration string     `json:"signature_declaration,omitempty"`
	AuthorID             *uuid.UUID `json:"author_id,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// ChecklistItem represents a single required or conditional document slot in an ANPP dossier
type ChecklistItem struct {
	SlotID              string     `json:"slot_id"`              // e.g. "CTD_3.2.P.5"
	TitleFR             string     `json:"title_fr"`
	TitleAR             string     `json:"title_ar"`
	ModuleNumber        string     `json:"module_number"`        // "Module 1", "Module 3"
	IsMandatory         bool       `json:"is_mandatory"`
	ConditionRule       string     `json:"condition_rule,omitempty"`
	Status              string     `json:"status"`               // "manquant", "en_revue", "valide"
	AttachedDocID       *uuid.UUID `json:"attached_doc_id,omitempty"`
	AttachedVersionID   *uuid.UUID `json:"attached_version_id,omitempty"`
	AttachedFilename    string     `json:"attached_filename,omitempty"`
	AttachedSHA256      string     `json:"attached_sha256,omitempty"`
	IsValidatedByDT     bool       `json:"is_validated_by_dt"`
}

// Dossier represents an official submission dossier for ANPP
type Dossier struct {
	ID                    uuid.UUID       `json:"id"`
	OrganizationID        uuid.UUID       `json:"organization_id"`
	ProductID             uuid.UUID       `json:"product_id"`
	ProcedureType         ProcedureType   `json:"procedure_type"`
	InternalReference     string          `json:"internal_reference"`
	ANPPOfficialReference string          `json:"anpp_official_reference,omitempty"`
	ANPPDepositDate       *time.Time      `json:"anpp_deposit_date,omitempty"`
	ANPPDecisionDate      *time.Time      `json:"anpp_decision_date,omitempty"`
	ANPPValidityEndDate   *time.Time      `json:"anpp_validity_end_date,omitempty"`
	Status                DossierStatus   `json:"status"`
	ResponsibleID         *uuid.UUID      `json:"responsible_id,omitempty"`
	ModelTemplateID       string          `json:"model_template_id"`
	ModelVersion          string          `json:"model_version"`
	ChecklistData         []ChecklistItem `json:"checklist_data"`
	CompletenessPercent   int             `json:"completeness_percent"`
	IsLockedForDeposit    bool            `json:"is_locked_for_deposit"`
	LockedByDTID          *uuid.UUID      `json:"locked_by_dt_id,omitempty"`
	LockedAt              *time.Time      `json:"locked_at,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// AuditLog represents an immutable audit trail entry (append-only)
type AuditLog struct {
	ID             int64      `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	UserEmail      string     `json:"user_email"`
	UserRole       string     `json:"user_role"`
	Action         string     `json:"action"` // e.g. "AUTH_LOGIN", "DOC_APPROVE_DT"
	EntityType     string     `json:"entity_type"`
	EntityID       string     `json:"entity_id"`
	Details        string     `json:"details"` // JSON payload
	IPAddress      string     `json:"ip_address"`
	UserAgent      string     `json:"user_agent"`
	SHA256Hash     string     `json:"sha256_hash"`
	CreatedAt      time.Time  `json:"created_at"`
}
