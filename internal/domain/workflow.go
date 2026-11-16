package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// WorkflowStatus représente les 5 états stricts du cycle de vie rédactionnel BPF
type WorkflowStatus string

const (
	WorkflowStatusDraftPending    WorkflowStatus = "draft_pending"    // À Rédiger
	WorkflowStatusDrafting        WorkflowStatus = "drafting"         // En cours de rédaction (Maker)
	WorkflowStatusUnderReview     WorkflowStatus = "under_review"     // En revue Qualité/Pair (Checker - AQ)
	WorkflowStatusPendingApproval WorkflowStatus = "pending_approval" // Prêt pour Visa DT (Approver)
	WorkflowStatusApprovedSealed  WorkflowStatus = "approved_sealed"  // Approuvé & Scellé v1.0
)

// DepartmentCode représente le pôle ou département métier assigné à la rédaction
type DepartmentCode string

const (
	DeptRA       DepartmentCode = "RA"       // Affaires Réglementaires (Module 1, notices)
	DeptQC       DepartmentCode = "QC"       // Contrôle Qualité & Stabilités (3.2.S.4, 3.2.P.5, 3.2.P.8)
	DeptRD       DepartmentCode = "RD"       // R&D / Galénique (3.2.P.2, Dissolution, QOS 2.3)
	DeptPROD     DepartmentCode = "PROD"     // Production Industrielle (3.2.P.3, Procédés, Validation)
	DeptQA       DepartmentCode = "QA"       // Assurance Qualité (Revue Checker, SOP, PQR)
	DeptExternal DepartmentCode = "EXTERNAL" // Tiers / CRO (DMF, Bioéquivalence Module 5)
)

// ReviewType représente la typologie d'intervention du Checker ou du PDT
type ReviewType string

const (
	ReviewTypeAnnotation    ReviewType = "annotation"           // Annotation ou remarque
	ReviewTypeChangeRequest ReviewType = "demande_modification" // Demande d'ajustement formelle
	ReviewTypeQualityPassed ReviewType = "validation_qualite"   // Visa favorable du Checker
	ReviewTypeRejected      ReviewType = "rejet_motive"         // Rejet motivé renvoyé en rédaction
)

// DossierItemReview représente une annotation ou décision dans la piste d'audit du workflow
type DossierItemReview struct {
	ID              uuid.UUID  `json:"id"`
	DossierItemID   uuid.UUID  `json:"dossier_item_id"`
	AuthorID        uuid.UUID  `json:"author_id"`
	AuthorName      string     `json:"author_name,omitempty"`
	AuthorRole      string     `json:"author_role"`
	ReviewType      ReviewType `json:"review_type"`
	Comment         string     `json:"comment"`
	RevisionVersion string     `json:"revision_version"`
	CreatedAt       time.Time  `json:"created_at"`
}

// MasterDocumentTemplate représente un modèle maître institutionnel ANPP / ICH
type MasterDocumentTemplate struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	TitleFR         string    `json:"title_fr"`
	TitleAR         string    `json:"title_ar,omitempty"`
	Category        string    `json:"category"`
	ModuleNumber    int       `json:"module_number"`
	SubModule       string    `json:"sub_module"`
	TargetProcedure string    `json:"target_procedure"`
	ContentTemplate string    `json:"content_template"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// WatermarkStatus représente l'état d'opposabilité légale pour l'impression BPF
type WatermarkStatus string

const (
	WatermarkDraft    WatermarkStatus = "BROUILLON" // Document de travail / non opposable
	WatermarkOfficial WatermarkStatus = "OFFICIEL"  // Visé et scellé par le Pharmacien DT
	WatermarkArchived WatermarkStatus = "ARCHIVE"   // Périmé / version antérieure
)

// GeneratedReport représente un état de sortie officiel généré et scellé (Chapitre 10 CDC)
type GeneratedReport struct {
	ID                uuid.UUID              `json:"id"`
	OrganizationID    uuid.UUID              `json:"organization_id"`
	DossierID         uuid.UUID              `json:"dossier_id"`
	ReportCode        string                 `json:"report_code"` // ES-REG-01, ES-REG-02, ES-FIN-01, EI-AVT-01, etc.
	ReportTitle       string                 `json:"report_title"`
	Format            string                 `json:"format"` // pdf, xlsx, html
	WatermarkStatus   WatermarkStatus        `json:"watermark_status"`
	SHA256Hash        string                 `json:"sha256_hash"`
	FileSizeBytes     int64                  `json:"file_size_bytes"`
	SignedByPDTID     *uuid.UUID             `json:"signed_by_pdt_id,omitempty"`
	SignedByPDTName   string                 `json:"signed_by_pdt_name,omitempty"`
	SignedAt          *time.Time             `json:"signed_at,omitempty"`
	QRVerificationURL string                 `json:"qr_verification_url,omitempty"`
	ReportMetadata    map[string]interface{} `json:"report_metadata,omitempty"`
	CreatedBy         *uuid.UUID             `json:"created_by,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

// ValidateCheckout vérifie si un document peut être verrouillé en édition
func ValidateCheckout(isLocked bool, lockedByID *uuid.UUID, currentUserID uuid.UUID) error {
	if isLocked && lockedByID != nil && *lockedByID != currentUserID {
		return errors.New("section déjà verrouillée en édition par un autre utilisateur (conflit de Check-out)")
	}
	return nil
}

// ValidateCheckin vérifie si un document peut être restitué (Check-in)
func ValidateCheckin(isLocked bool, lockedByID *uuid.UUID, currentUserID uuid.UUID) error {
	if !isLocked {
		return errors.New("impossible d'effectuer un Check-in sur une section qui n'est pas verrouillée")
	}
	if lockedByID != nil && *lockedByID != currentUserID {
		return errors.New("seul l'utilisateur ayant posé le verrou (Check-out) est autorisé à effectuer le Check-in")
	}
	return nil
}

// NextMinorRevision incrémente une version de révision de travail (v0.1 -> v0.2 -> v0.3)
func NextMinorRevision(currentRev string) string {
	curr := strings.TrimSpace(currentRev)
	if curr == "" || !strings.HasPrefix(curr, "v0.") {
		return "v0.1"
	}
	sub := strings.TrimPrefix(curr, "v0.")
	var minor int
	for _, c := range sub {
		if c >= '0' && c <= '9' {
			minor = minor*10 + int(c-'0')
		} else {
			break
		}
	}
	return fmt.Sprintf("v0.%d", minor+1)
}
