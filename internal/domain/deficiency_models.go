package domain

import (
	"math"
	"time"
)

// Constantes pour le cycle de vie des lettres de compléments ANPP
const (
	DeficiencyLetterStatusOpen              = "open"
	DeficiencyLetterStatusPartiallyAnswered = "partially_answered"
	DeficiencyLetterStatusSubmitted         = "submitted"
	DeficiencyLetterStatusOverdue           = "overdue"
)

// Constantes pour le statut des questions/tâches de compléments
const (
	DeficiencyTaskStatusPending     = "a_traiter"
	DeficiencyTaskStatusInProgress  = "en_cours"
	DeficiencyTaskStatusAnswered    = "repondu"
	DeficiencyTaskStatusValidatedDT = "valide_dt"
)

// ANPPDeficiencyLetter représente une lettre officielle de notification de compléments de l'ANPP
type ANPPDeficiencyLetter struct {
	ID                 string               `json:"id"`
	OrganizationID     string               `json:"organization_id"`
	DossierID          string               `json:"dossier_id"`
	DossierReference   string               `json:"dossier_reference,omitempty"`
	DossierTitle       string               `json:"dossier_title,omitempty"`
	ProductBrandName   string               `json:"product_brand_name,omitempty"`
	ReferenceNumber    string               `json:"reference_number"`    // Ex: ANPP/DEPP/N°1428/2026
	ReceivedDate       time.Time            `json:"received_date"`
	LegalDeadlineDays  int                  `json:"legal_deadline_days"` // 30, 60, 90, 180 jours
	DueDate            time.Time            `json:"due_date"`
	Status             string               `json:"status"`              // open, partially_answered, submitted, overdue
	DocumentID         *string              `json:"document_id,omitempty"` // Scan GED de la lettre
	DocumentTitle      string               `json:"document_title,omitempty"`
	Notes              string               `json:"notes"`
	Tasks              []ANPPDeficiencyTask `json:"tasks,omitempty"`
	TotalTasksCount    int                  `json:"total_tasks_count"`
	AnsweredTasksCount int                  `json:"answered_tasks_count"`
	ValidatedTasksCount int                 `json:"validated_tasks_count"`
	DaysRemaining      int                  `json:"days_remaining"`
	UrgencyBadge       string               `json:"urgency_badge"` // critical, urgent, warning, normal, expired
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

// ANPPDeficiencyTask représente une question unitaire décomposée issue de la lettre
type ANPPDeficiencyTask struct {
	ID                   string     `json:"id"`
	LetterID             *string    `json:"letter_id,omitempty"`
	DossierID            string     `json:"dossier_id"`
	OrganizationID       string     `json:"organization_id"`
	QuestionNumber       int        `json:"question_number"`
	QuestionText         string     `json:"question_text"`
	ModuleTarget         string     `json:"module_target"` // Module 1, Module 3, Module 4, Module 5, etc.
	Priority             string     `json:"priority"`      // normale, haute, urgente
	AssignedTo           *string    `json:"assigned_to,omitempty"`
	AssignedUserName     string     `json:"assigned_user_name,omitempty"`
	AssignedUserRole     string     `json:"assigned_user_role,omitempty"`
	DueDate              time.Time  `json:"due_date"`
	Status               string     `json:"status"` // a_traiter, en_cours, repondu, valide_dt
	ResponseDocumentID   *string    `json:"response_document_id,omitempty"`
	ResponseDocumentTitle string    `json:"response_document_title,omitempty"`
	ResponseSummary      string     `json:"response_summary"`
	ValidatedByDTID      *string    `json:"validated_by_dt_id,omitempty"`
	ValidatedByName      string     `json:"validated_by_name,omitempty"`
	ValidatedAt          *time.Time `json:"validated_at,omitempty"`
	DaysRemaining        int        `json:"days_remaining"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// CalculateDeadlineAndUrgency calcule le nombre de jours restants et le niveau d'urgence
func CalculateDeadlineAndUrgency(dueDate time.Time, now time.Time) (int, string) {
	diff := dueDate.Sub(now)
	days := int(math.Ceil(diff.Hours() / 24.0))

	var badge string
	switch {
	case days < 0:
		badge = "expired"
	case days <= 15:
		badge = "critical" // Rouge clignotant / alerte maximale
	case days <= 30:
		badge = "urgent"   // Orange vif
	case days <= 60:
		badge = "warning"  // Ambre
	default:
		badge = "normal"   // Vert
	}
	return days, badge
}

// ComputeLetterSummary recalcule les compteurs et l'urgence pour une lettre
func (l *ANPPDeficiencyLetter) ComputeMetrics(now time.Time) {
	l.DaysRemaining, l.UrgencyBadge = CalculateDeadlineAndUrgency(l.DueDate, now)
	l.TotalTasksCount = len(l.Tasks)
	l.AnsweredTasksCount = 0
	l.ValidatedTasksCount = 0

	for i := range l.Tasks {
		l.Tasks[i].DaysRemaining, _ = CalculateDeadlineAndUrgency(l.Tasks[i].DueDate, now)
		if l.Tasks[i].Status == DeficiencyTaskStatusAnswered || l.Tasks[i].Status == DeficiencyTaskStatusValidatedDT {
			l.AnsweredTasksCount++
		}
		if l.Tasks[i].Status == DeficiencyTaskStatusValidatedDT {
			l.ValidatedTasksCount++
		}
	}

	// Ajuster le statut global si nécessaire
	if l.Status != DeficiencyLetterStatusSubmitted {
		if l.DaysRemaining < 0 {
			l.Status = DeficiencyLetterStatusOverdue
		} else if l.TotalTasksCount > 0 && l.ValidatedTasksCount == l.TotalTasksCount {
			l.Status = DeficiencyLetterStatusPartiallyAnswered // Prêt pour soumission
		} else if l.AnsweredTasksCount > 0 {
			l.Status = DeficiencyLetterStatusPartiallyAnswered
		}
	}
}

// CreateDeficiencyLetterRequest représente la charge utile de création d'une lettre
type CreateDeficiencyLetterRequest struct {
	ReferenceNumber   string                      `json:"reference_number"`
	ReceivedDate      string                      `json:"received_date"`      // "YYYY-MM-DD"
	LegalDeadlineDays int                         `json:"legal_deadline_days"` // 30, 60, 90, 180
	DueDate           string                      `json:"due_date"`           // "YYYY-MM-DD"
	DocumentID        *string                     `json:"document_id,omitempty"`
	Notes             string                      `json:"notes"`
	Tasks             []CreateDeficiencyTaskInput `json:"tasks"`
}

// CreateDeficiencyTaskInput représente une question à insérer avec la lettre
type CreateDeficiencyTaskInput struct {
	QuestionNumber int     `json:"question_number"`
	QuestionText   string  `json:"question_text"`
	ModuleTarget   string  `json:"module_target"`
	Priority       string  `json:"priority"`
	AssignedTo     *string `json:"assigned_to,omitempty"`
	DueDate        string  `json:"due_date"` // "YYYY-MM-DD"
}

// UpdateDeficiencyTaskRequest représente la charge utile de mise à jour d'une tâche
type UpdateDeficiencyTaskRequest struct {
	Status             string  `json:"status"` // a_traiter, en_cours, repondu
	ResponseDocumentID *string `json:"response_document_id,omitempty"`
	ResponseSummary    string  `json:"response_summary"`
	AssignedTo         *string `json:"assigned_to,omitempty"`
	DueDate            *string `json:"due_date,omitempty"`
	Priority           *string `json:"priority,omitempty"`
}

// ValidateDeficiencyTaskRequest représente la demande de visa Pharmacien Directeur Technique
type ValidateDeficiencyTaskRequest struct {
	Comment string `json:"comment"`
}

// DeficiencyDashboardSummary regroupe les KPI pour le pilotage d'affaires réglementaires
type DeficiencyDashboardSummary struct {
	TotalActiveLetters  int                    `json:"total_active_letters"`
	CriticalLetters     int                    `json:"critical_letters"` // J <= 15
	UrgentLetters       int                    `json:"urgent_letters"`   // 15 < J <= 30
	PendingTasksCount   int                    `json:"pending_tasks_count"`
	ValidatedTasksCount int                    `json:"validated_tasks_count"`
	RecentLetters       []ANPPDeficiencyLetter `json:"recent_letters"`
}
