package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// DeficiencyHandler gère les requêtes HTTP relatives aux compléments ANPP
type DeficiencyHandler struct {
	deficiencyRepo *repository.DeficiencyRepository
	dossierRepo    *repository.DossierRepository
	auditService   *audit.Service
	pool           *pgxpool.Pool
}

// NewDeficiencyHandler instancie le contrôleur REST des compléments ANPP
func NewDeficiencyHandler(
	deficiencyRepo *repository.DeficiencyRepository,
	dossierRepo *repository.DossierRepository,
	auditService *audit.Service,
	pool *pgxpool.Pool,
) *DeficiencyHandler {
	return &DeficiencyHandler{
		deficiencyRepo: deficiencyRepo,
		dossierRepo:    dossierRepo,
		auditService:   auditService,
		pool:           pool,
	}
}

// ListDossierDeficiencies retourne l'ensemble des lettres et questions de compléments d'un dossier
func (h *DeficiencyHandler) ListDossierDeficiencies(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	if _, err := uuid.Parse(dossierIDStr); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de dossier invalide")
		return
	}

	letters, err := h.deficiencyRepo.ListLettersByDossier(r.Context(), user.OrganizationID.String(), dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur lors de la récupération des compléments: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"dossier_id": dossierIDStr,
		"letters":    letters,
		"total":      len(letters),
	})
}

// CreateDossierDeficiency enregistre une nouvelle lettre officielle ANPP et décompose ses questions
func (h *DeficiencyHandler) CreateDossierDeficiency(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	if _, err := uuid.Parse(dossierIDStr); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de dossier invalide")
		return
	}

	var req domain.CreateDeficiencyLetterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.ReferenceNumber = strings.TrimSpace(req.ReferenceNumber)
	if req.ReferenceNumber == "" {
		writeJSONError(w, http.StatusBadRequest, "Le numéro de référence officielle ANPP (ex: ANPP/DEPP/N°...) est obligatoire")
		return
	}

	// Parsing de la date de notification officielle
	var receivedDate time.Time
	var err error
	if req.ReceivedDate != "" {
		receivedDate, err = time.Parse("2006-01-02", req.ReceivedDate)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Format de date de réception invalide (attendu: AAAA-MM-JJ)")
			return
		}
	} else {
		receivedDate = time.Now().UTC()
	}

	if req.LegalDeadlineDays <= 0 {
		req.LegalDeadlineDays = 90 // Délai légal par défaut selon décret 20-325
	}

	// Calcul ou parsing de la date limite légale
	var dueDate time.Time
	if req.DueDate != "" {
		dueDate, err = time.Parse("2006-01-02", req.DueDate)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Format de date limite invalide (attendu: AAAA-MM-JJ)")
			return
		}
	} else {
		dueDate = receivedDate.AddDate(0, 0, req.LegalDeadlineDays)
	}

	letter := domain.ANPPDeficiencyLetter{
		OrganizationID:    user.OrganizationID.String(),
		DossierID:         dossierIDStr,
		ReferenceNumber:   req.ReferenceNumber,
		ReceivedDate:      receivedDate,
		LegalDeadlineDays: req.LegalDeadlineDays,
		DueDate:           dueDate,
		Status:            domain.DeficiencyLetterStatusOpen,
		DocumentID:        req.DocumentID,
		Notes:             strings.TrimSpace(req.Notes),
	}

	tasks := make([]domain.ANPPDeficiencyTask, 0, len(req.Tasks))
	for i, tInput := range req.Tasks {
		qText := strings.TrimSpace(tInput.QuestionText)
		if qText == "" {
			continue
		}

		tDueDate := dueDate
		if tInput.DueDate != "" {
			parsedTDue, err := time.Parse("2006-01-02", tInput.DueDate)
			if err == nil {
				tDueDate = parsedTDue
			}
		}

		qNum := tInput.QuestionNumber
		if qNum <= 0 {
			qNum = i + 1
		}

		tasks = append(tasks, domain.ANPPDeficiencyTask{
			QuestionNumber: qNum,
			QuestionText:   qText,
			ModuleTarget:   strings.TrimSpace(tInput.ModuleTarget),
			Priority:       strings.TrimSpace(tInput.Priority),
			AssignedTo:     tInput.AssignedTo,
			DueDate:        tDueDate,
			Status:         domain.DeficiencyTaskStatusPending,
		})
	}

	created, err := h.deficiencyRepo.CreateLetterWithTasks(r.Context(), &letter, tasks)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Création lettre compléments impossible: %v", err))
		return
	}

	// Audit Trail ALCOA+
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         audit.ActionDeficiencyLetterCreate,
			EntityType:     "anpp_deficiency_letters",
			EntityID:       created.ID,
			Details: map[string]any{
				"reference_number": created.ReferenceNumber,
				"dossier_id":       dossierIDStr,
				"tasks_count":      len(tasks),
				"due_date":         created.DueDate.Format("2006-01-02"),
			},
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusCreated, created)
}

// GetDeficiencyLetter récupère les détails d'une lettre de compléments et ses tâches
func (h *DeficiencyHandler) GetDeficiencyLetter(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	letterIDStr := chi.URLParam(r, "letterId")
	if _, err := uuid.Parse(letterIDStr); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de lettre invalide")
		return
	}

	letter, err := h.deficiencyRepo.GetLetter(r.Context(), user.OrganizationID.String(), letterIDStr)
	if err != nil {
		if errors.Is(err, repository.ErrLetterNotFound) {
			writeJSONError(w, http.StatusNotFound, "Lettre de compléments introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur récupération lettre: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, letter)
}

// GetDeficiencyDashboard retourne les alertes de délai légal et KPI pour l'établissement
func (h *DeficiencyHandler) GetDeficiencyDashboard(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dashboard, err := h.deficiencyRepo.ListAllActiveDeficiencies(r.Context(), user.OrganizationID.String())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur tableau de bord compléments: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, dashboard)
}

// UpdateDeficiencyTask met à jour l'avancement, le document réponse ou l'attribution d'une tâche
func (h *DeficiencyHandler) UpdateDeficiencyTask(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	taskIDStr := chi.URLParam(r, "taskId")
	if _, err := uuid.Parse(taskIDStr); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de tâche invalide")
		return
	}

	task, err := h.deficiencyRepo.GetTask(r.Context(), user.OrganizationID.String(), taskIDStr)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeJSONError(w, http.StatusNotFound, "Tâche de complément introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur chargement tâche: %v", err))
		return
	}

	var req domain.UpdateDeficiencyTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	if req.Status != "" {
		task.Status = req.Status
	}
	if req.ResponseSummary != "" {
		task.ResponseSummary = strings.TrimSpace(req.ResponseSummary)
	}
	if req.ResponseDocumentID != nil {
		task.ResponseDocumentID = req.ResponseDocumentID
	}
	if req.AssignedTo != nil {
		task.AssignedTo = req.AssignedTo
	}
	if req.Priority != nil && *req.Priority != "" {
		task.Priority = *req.Priority
	}
	if req.DueDate != nil && *req.DueDate != "" {
		if dDate, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			task.DueDate = dDate
		}
	}

	// Si une réponse est fournie et qu'aucun statut n'a été spécifié, marquer comme répondu
	if task.ResponseSummary != "" && task.Status == domain.DeficiencyTaskStatusPending {
		task.Status = domain.DeficiencyTaskStatusAnswered
	}

	if err := h.deficiencyRepo.UpdateTask(r.Context(), task); err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur mise à jour tâche: %v", err))
		return
	}

	// Audit Trail ALCOA+
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         audit.ActionDeficiencyTaskUpdate,
			EntityType:     "anpp_deficiency_tasks",
			EntityID:       task.ID,
			Details: map[string]any{
				"status":           task.Status,
				"response_summary": task.ResponseSummary,
				"has_document":     task.ResponseDocumentID != nil,
			},
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
		})
	}

	updated, _ := h.deficiencyRepo.GetTask(r.Context(), user.OrganizationID.String(), taskIDStr)
	writeJSON(w, http.StatusOK, updated)
}

// ValidateDeficiencyTaskDT appose le visa du Pharmacien Directeur Technique (PDT - Mandat 1.A.7)
func (h *DeficiencyHandler) ValidateDeficiencyTaskDT(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	// Vérification stricte du rôle Pharmacien Directeur Technique ou de la permission d'approbation DT
	if user.Role != string(domain.RoleDirecteurTechnique) && user.Role != string(domain.RoleAdminLaboratoire) && !user.HasPermission(string(domain.PermDocApprovePDT)) {
		writeJSONError(w, http.StatusForbidden, "Le visa réglementaire d'une réponse de complément est réservé exclusivement au Pharmacien Directeur Technique (Mandat Légal 1.A.7)")
		return
	}

	taskIDStr := chi.URLParam(r, "taskId")
	if _, err := uuid.Parse(taskIDStr); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de tâche invalide")
		return
	}

	var req domain.ValidateDeficiencyTaskRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.deficiencyRepo.ValidateTaskDT(r.Context(), user.OrganizationID.String(), taskIDStr, user.UserID.String(), strings.TrimSpace(req.Comment)); err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeJSONError(w, http.StatusNotFound, "Tâche de complément introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur visa PDT: %v", err))
		return
	}

	// Audit Trail ALCOA+
	if h.auditService != nil {
		fullName := strings.TrimSpace(user.FirstName + " " + user.LastName)
		if fullName == "" {
			fullName = user.Email
		}
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         audit.ActionDeficiencyTaskSignDT,
			EntityType:     "anpp_deficiency_tasks",
			EntityID:       taskIDStr,
			Details: map[string]any{
				"validated_by": fullName,
				"comment":      req.Comment,
			},
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
		})
	}

	validated, _ := h.deficiencyRepo.GetTask(r.Context(), user.OrganizationID.String(), taskIDStr)
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Visa Pharmacien Directeur Technique apposé avec succès",
		"task":    validated,
	})
}
