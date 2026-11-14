package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// AppointmentHandler gère les rendez-vous et séances de dépôt / commission ANPP
type AppointmentHandler struct {
	repo         *repository.AppointmentRepository
	dossierRepo  *repository.DossierRepository
	auditService *audit.Service
}

// NewAppointmentHandler instancie le gestionnaire de rendez-vous ANPP
func NewAppointmentHandler(
	repo *repository.AppointmentRepository,
	dossierRepo *repository.DossierRepository,
	auditService *audit.Service,
) *AppointmentHandler {
	return &AppointmentHandler{
		repo:         repo,
		dossierRepo:  dossierRepo,
		auditService: auditService,
	}
}

// ListAppointments retourne la liste des créneaux officiels
func (h *AppointmentHandler) ListAppointments(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var dossierID *uuid.UUID
	dossierParam := r.URL.Query().Get("dossier_id")
	if dossierParam != "" {
		if dID, err := uuid.Parse(dossierParam); err == nil {
			dossierID = &dID
		}
	}

	statusFilter := r.URL.Query().Get("status")
	upcomingOnly := r.URL.Query().Get("upcoming") == "true"

	list, err := h.repo.ListAppointments(r.Context(), user.OrganizationID, dossierID, statusFilter, upcomingOnly)
	if err != nil {
		log.Printf("[ERROR] ListAppointments: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Erreur lecture des rendez-vous")
		return
	}

	// Calcul des métadonnées enrichies (urgences J <= 7)
	urgentCount := 0
	for i := range list {
		if list[i].IsUrgent() {
			urgentCount++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"appointments": list,
		"total":        len(list),
		"urgent_count": urgentCount,
	})
}

// CreateAppointmentRequest payload pour planifier une séance ou créneau ANPP
type CreateAppointmentRequest struct {
	DossierID          *uuid.UUID             `json:"dossier_id,omitempty"`
	AppointmentType    domain.AppointmentType `json:"appointment_type"`
	ScheduledAt        string                 `json:"scheduled_at"` // ISO8601 ou YYYY-MM-DDTHH:MM
	Location           string                 `json:"location"`
	SlotReference      string                 `json:"slot_reference"`
	ContactOfficer     string                 `json:"contact_officer,omitempty"`
	AssignedDelegateID *uuid.UUID             `json:"assigned_delegate_id,omitempty"`
	Notes              string                 `json:"notes,omitempty"`
}

// CreateAppointment crée un nouveau rendez-vous
func (h *AppointmentHandler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var req CreateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	if req.AppointmentType == "" {
		req.AppointmentType = domain.AppointmentTypeDepotInitial
	}
	if req.Location == "" {
		req.Location = "ANPP - Direction de l'Enregistrement, Annexe Dely Ibrahim, Alger"
	}
	if req.SlotReference == "" {
		req.SlotReference = "RDV-ANPP-" + time.Now().Format("2006-0102-1504")
	}

	scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		scheduledTime, err = time.Parse("2006-01-02T15:04", req.ScheduledAt)
		if err != nil {
			scheduledTime, err = time.Parse("2006-01-02", req.ScheduledAt)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "Format de date invalide (attendu: YYYY-MM-DDTHH:MM ou ISO8601)")
				return
			}
		}
	}

	app := &domain.ANPPAppointment{
		OrganizationID:     user.OrganizationID,
		DossierID:          req.DossierID,
		AppointmentType:    req.AppointmentType,
		ScheduledAt:        scheduledTime,
		Location:           req.Location,
		SlotReference:      req.SlotReference,
		ContactOfficer:     req.ContactOfficer,
		AssignedDelegateID: req.AssignedDelegateID,
		Status:             domain.AppointmentStatusPlanifie,
		Notes:              req.Notes,
		CreatedBy:          &user.UserID,
	}

	created, err := h.repo.CreateAppointment(r.Context(), app)
	if err != nil {
		log.Printf("[ERROR] CreateAppointment: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de créer le rendez-vous ANPP")
		return
	}

	// Audit Trail BPF
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionAppointmentCreate,
		EntityType:     "anpp_appointment",
		EntityID:       created.ID.String(),
		Details: map[string]any{
			"slot_reference":   created.SlotReference,
			"appointment_type": created.AppointmentType,
			"scheduled_at":     created.ScheduledAt.Format(time.RFC3339),
			"location":         created.Location,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, created)
}

// GetAppointment retourne un rendez-vous par son ID
func (h *AppointmentHandler) GetAppointment(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	appIDStr := chi.URLParam(r, "appId")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant rendez-vous invalide")
		return
	}

	app, err := h.repo.GetAppointment(r.Context(), user.OrganizationID, appID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Rendez-vous introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur lecture rendez-vous")
		return
	}

	writeJSON(w, http.StatusOK, app)
}

// UpdateAppointmentStatusRequest payload de changement d'état
type UpdateAppointmentStatusRequest struct {
	Status domain.AppointmentStatus `json:"status"`
	Notes  *string                  `json:"notes,omitempty"`
}

// UpdateAppointmentStatus met à jour le statut du créneau
func (h *AppointmentHandler) UpdateAppointmentStatus(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	appIDStr := chi.URLParam(r, "appId")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant rendez-vous invalide")
		return
	}

	var req UpdateAppointmentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	updated, err := h.repo.UpdateAppointmentStatus(r.Context(), user.OrganizationID, appID, req.Status, req.Notes)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Rendez-vous introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur mise à jour statut")
		return
	}

	// Audit Trail BPF
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionAppointmentStatus,
		EntityType:     "anpp_appointment",
		EntityID:       appID.String(),
		Details: map[string]any{
			"new_status": updated.Status,
			"slot_ref":   updated.SlotReference,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, updated)
}
