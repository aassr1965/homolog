package domain

import (
	"time"

	"github.com/google/uuid"
)

// InspectionFinding représente une observation / écart d'inspection BPF (LD-DVIV04/V01) et son plan CAPA.
type InspectionFinding struct {
	ID                  uuid.UUID  `json:"id"`
	OrganizationID      uuid.UUID  `json:"organization_id"`
	InspectionReference string     `json:"inspection_reference"`
	InspectionDate      time.Time  `json:"inspection_date"`
	InspectorAuthority  string     `json:"inspector_authority"`
	Severity            string     `json:"severity"` // CRITIQUE, MAJEUR, AUTRE
	Description         string     `json:"description"`
	CAPAPlan            string     `json:"capa_plan,omitempty"`
	ResponsibleName     string     `json:"responsible_name,omitempty"`
	DueDate             *time.Time `json:"due_date,omitempty"`
	Status              string     `json:"status"` // ouvert, capa_soumis, cloture
	GEDReportDocumentID *uuid.UUID `json:"ged_report_document_id,omitempty"`
	CreatedBy           *uuid.UUID `json:"created_by,omitempty"`
	ClosedBy            *uuid.UUID `json:"closed_by,omitempty"`
	ClosedAt            *time.Time `json:"closed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
