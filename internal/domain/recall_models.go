package domain

import (
	"time"

	"github.com/google/uuid"
)

// LotRecall représente un rappel de lot ou retrait de PP/DM non conforme (délai légal 24h ANPP).
type LotRecall struct {
	ID              uuid.UUID  `json:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id"`
	RecallReference string     `json:"recall_reference"`
	ProductKind     string     `json:"product_kind"` // pp, dm
	ProductName     string     `json:"product_name"`
	LotNumbers      string     `json:"lot_numbers"`
	SeverityClass   string     `json:"severity_class"` // CLASSE_I, CLASSE_II, CLASSE_III
	IsMockDrill     bool       `json:"is_mock_drill"`
	Reason          string     `json:"reason"`
	DetectedAt      time.Time  `json:"detected_at"`
	NotifiedANPPAt  *time.Time `json:"notified_anpp_at,omitempty"`
	Within24hSLA    bool       `json:"within_24h_sla"`
	Status          string     `json:"status"` // ouvert, notifie_anpp, cloture
	GEDPVDocumentID *uuid.UUID `json:"ged_pv_document_id,omitempty"`
	CreatedBy       *uuid.UUID `json:"created_by,omitempty"`
	SignedByPDTID   *uuid.UUID `json:"signed_by_pdt_id,omitempty"`
	SignedAt        *time.Time `json:"signed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ComputeSLA24h calcule si la notification ANPP a été effectuée dans le délai réglementaire de 24h.
func (r *LotRecall) ComputeSLA24h() bool {
	if r.NotifiedANPPAt != nil {
		return r.NotifiedANPPAt.Sub(r.DetectedAt) <= 24*time.Hour
	}
	return time.Since(r.DetectedAt) <= 24*time.Hour
}
