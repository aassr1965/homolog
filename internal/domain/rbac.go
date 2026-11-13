package domain

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents an individual system permission (CDC Section 7.2)
type Permission string

const (
	PermSysMaint        Permission = "SYS_MAINT"
	PermUserManage      Permission = "USER_MANAGE"
	PermRefView         Permission = "REF_VIEW"
	PermRefEdit         Permission = "REF_EDIT"
	PermDossierView     Permission = "DOSSIER_VIEW"
	PermDossierCreate   Permission = "DOSSIER_CREATE"
	PermDocRead         Permission = "DOC_READ"
	PermDocDraftMaker   Permission = "DOC_DRAFT_MAKER"
	PermDocCheckReview  Permission = "DOC_CHECK_REVIEW"
	PermDocLockCheckout Permission = "DOC_LOCK_CHECKOUT"
	PermDocApprovePDT   Permission = "DOC_APPROVE_PDT"
	PermDocArchive      Permission = "DOC_ARCHIVE"
	PermExportPackage   Permission = "EXPORT_PACKAGE"
	PermAuditView       Permission = "AUDIT_VIEW"
)

// RoleInfo represents a system or custom role
type RoleInfo struct {
	ID          string       `json:"id"`
	NameFR      string       `json:"name_fr"`
	NameAR      string       `json:"name_ar,omitempty"`
	Description string       `json:"description,omitempty"`
	IsSystem    bool         `json:"is_system"`
	Permissions []Permission `json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}

// PermissionInfo represents a permission definition in the catalog
type PermissionInfo struct {
	ID          Permission `json:"id"`
	NameFR      string     `json:"name_fr"`
	Description string     `json:"description"`
	Category    string     `json:"category"` // 'SYSTEM', 'REFERENTIEL', 'DOSSIER', 'GED', 'AUDIT'
	CreatedAt   time.Time  `json:"created_at"`
}

// UserProfileSummary represents public/authenticated user information with permissions
type UserProfileSummary struct {
	ID                  uuid.UUID    `json:"id"`
	OrganizationID      uuid.UUID    `json:"organization_id"`
	Email               string       `json:"email"`
	FirstName           string       `json:"first_name"`
	LastName            string       `json:"last_name"`
	Role                UserRole     `json:"role"`
	RoleNameFR          string       `json:"role_name_fr"`
	Permissions         []Permission `json:"permissions"`
	PermissionsVersion  int          `json:"permissions_version"`
	TOTPEnabled         bool         `json:"totp_enabled"`
	IsTechnicalDirector bool         `json:"is_technical_director"`
}
