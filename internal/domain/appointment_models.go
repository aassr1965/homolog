package domain

import (
	"time"

	"github.com/google/uuid"
)

// AppointmentType type de séance ou de créneau auprès de l'ANPP
type AppointmentType string

const (
	AppointmentTypeDepotInitial    AppointmentType = "depot_initial"     // Dépôt initial Direction de l'Enregistrement
	AppointmentTypeCommission      AppointmentType = "seance_commission" // Séance plénière de la Commission d'Enregistrement
	AppointmentTypeExpertiseLNCM   AppointmentType = "expertise_lncm"    // Dépôt d'échantillons et étalons au LNCM
	AppointmentTypeRecoursAudition AppointmentType = "recours_audition"  // Audition contradictoire / Recours technique
	AppointmentTypeComplements     AppointmentType = "depot_complements" // Remise physique des réponses aux compléments
)

// AppointmentStatus état du rendez-vous
type AppointmentStatus string

const (
	AppointmentStatusPlanifie     AppointmentStatus = "planifie"      // Planifié en interne
	AppointmentStatusConfirmeANPP AppointmentStatus = "confirme_anpp" // Confirmé officiellement par l'ANPP
	AppointmentStatusEffectue     AppointmentStatus = "effectue"      // Réalisé avec récépissé de dépôt
	AppointmentStatusReporte      AppointmentStatus = "reporte"       // Ajourné / Reporté
	AppointmentStatusAnnule       AppointmentStatus = "annule"        // Annulé
)

// ANPPAppointment créneau officiel de rendez-vous ou séance de commission ANPP
type ANPPAppointment struct {
	ID                     uuid.UUID         `json:"id"`
	OrganizationID         uuid.UUID         `json:"organization_id"`
	DossierID              *uuid.UUID        `json:"dossier_id,omitempty"`
	DossierCode            string            `json:"dossier_code,omitempty"`
	ProductBrandName       string            `json:"product_brand_name,omitempty"`
	AppointmentType        AppointmentType   `json:"appointment_type"`
	ScheduledAt            time.Time         `json:"scheduled_at"`
	Location               string            `json:"location"`
	SlotReference          string            `json:"slot_reference"`
	ContactOfficer         string            `json:"contact_officer,omitempty"`
	AssignedDelegateID     *uuid.UUID        `json:"assigned_delegate_id,omitempty"`
	AssignedDelegateName   string            `json:"assigned_delegate_name,omitempty"`
	Status                 AppointmentStatus `json:"status"`
	Notes                  string            `json:"notes,omitempty"`
	ConfirmationDocumentID *uuid.UUID        `json:"confirmation_document_id,omitempty"`
	CreatedBy              *uuid.UUID        `json:"created_by,omitempty"`
	CreatedAt              time.Time         `json:"created_at"`
	UpdatedAt              time.Time         `json:"updated_at"`
}

// DaysUntil calcule le nombre de jours calendaires restants avant la séance
func (a *ANPPAppointment) DaysUntil() int {
	now := time.Now().UTC()
	diff := a.ScheduledAt.Sub(now)
	days := int(diff.Hours() / 24)
	if diff > 0 && days == 0 {
		return 1
	}
	return days
}

// IsUrgent indique si le rendez-vous a lieu dans les 7 jours à venir
func (a *ANPPAppointment) IsUrgent() bool {
	days := a.DaysUntil()
	return days >= 0 && days <= 7 && (a.Status == AppointmentStatusPlanifie || a.Status == AppointmentStatusConfirmeANPP)
}

// IsPast indique si la date est échue
func (a *ANPPAppointment) IsPast() bool {
	return a.ScheduledAt.Before(time.Now().UTC())
}

// FormatTypeLabel libellé lisible en français
func (a *ANPPAppointment) FormatTypeLabel() string {
	switch a.AppointmentType {
	case AppointmentTypeDepotInitial:
		return "Dépôt Initial de Dossier (Direction Enregistrement)"
	case AppointmentTypeCommission:
		return "Séance de Commission d'Enregistrement"
	case AppointmentTypeExpertiseLNCM:
		return "Dépôt Échantillons / Expertise LNCM"
	case AppointmentTypeRecoursAudition:
		return "Audition Technique & Recours"
	case AppointmentTypeComplements:
		return "Dépôt des Réponses aux Compléments"
	default:
		return string(a.AppointmentType)
	}
}
