package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/config"
	"homolog/internal/domain"
)

// SetupHandler gère l'assistant de premier démarrage (Setup Wizard BPF)
type SetupHandler struct {
	pool           *pgxpool.Pool
	sessionManager *auth.SessionManager
	auditService   *audit.Service
	cfg            *config.Config
}

// NewSetupHandler instancie le handler de premier démarrage
func NewSetupHandler(pool *pgxpool.Pool, sm *auth.SessionManager, auditService *audit.Service, cfg *config.Config) *SetupHandler {
	return &SetupHandler{
		pool:           pool,
		sessionManager: sm,
		auditService:   auditService,
		cfg:            cfg,
	}
}

// SetupStatusResponse décrit l'état d'initialisation de l'instance
type SetupStatusResponse struct {
	SetupRequired    bool   `json:"setup_required"`
	Initialized      bool   `json:"initialized"`
	OrganizationName string `json:"organization_name,omitempty"`
	UsersCount       int    `json:"users_count"`
	Version          string `json:"version"`
	DeploymentMode   string `json:"deployment_mode"`
}

// GetSetupStatus vérifie si l'instance requiert un paramétrage initial
func (h *SetupHandler) GetSetupStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var usersCount int
	err := h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&usersCount)
	if err != nil {
		usersCount = 0
	}

	var orgName string
	_ = h.pool.QueryRow(ctx, "SELECT legal_name FROM organizations ORDER BY created_at ASC LIMIT 1").Scan(&orgName)

	resp := SetupStatusResponse{
		SetupRequired:    usersCount == 0,
		Initialized:      usersCount > 0,
		OrganizationName: orgName,
		UsersCount:       usersCount,
		Version:          "0.8.0",
		DeploymentMode:   "on-premise",
	}
	if h.cfg != nil && h.cfg.DeploymentMode != "" {
		resp.DeploymentMode = string(h.cfg.DeploymentMode)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// SetupInitRequest porte les données de configuration initiale de l'établissement et du Pharmacien DT
type SetupInitRequest struct {
	OrganizationName  string `json:"organization_name"`
	OrganizationSlug  string `json:"organization_slug"`
	MIPHLicenseNumber string `json:"miph_license_number"`
	DTFirstName       string `json:"dt_first_name"`
	DTLastName        string `json:"dt_last_name"`
	DTEmail           string `json:"dt_email"`
	DTPassword        string `json:"dt_password"`
	DTDecisionNumber  string `json:"dt_decision_number"`
	ActivityType      string `json:"activity_type"` // 'fabrication_locale', 'importation', 'dispositifs_medicaux'
}

// InitializeInstance configure le premier laboratoire et crée le compte régalien du Pharmacien Directeur Technique
func (h *SetupHandler) InitializeInstance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Contrôle préalable : interdiction de réinitialiser si des comptes existent déjà
	var existingCount int
	_ = h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&existingCount)
	if existingCount > 0 {
		writeJSONError(w, http.StatusConflict, "L'instance a déjà été initialisée. Le Setup Wizard est désactivé.")
		return
	}

	var req SetupInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.OrganizationName = strings.TrimSpace(req.OrganizationName)
	req.OrganizationSlug = strings.TrimSpace(req.OrganizationSlug)
	req.MIPHLicenseNumber = strings.TrimSpace(req.MIPHLicenseNumber)
	req.DTEmail = strings.TrimSpace(strings.ToLower(req.DTEmail))
	req.DTPassword = strings.TrimSpace(req.DTPassword)
	req.DTFirstName = strings.TrimSpace(req.DTFirstName)
	req.DTLastName = strings.TrimSpace(req.DTLastName)
	req.DTDecisionNumber = strings.TrimSpace(req.DTDecisionNumber)

	if req.OrganizationName == "" || req.DTEmail == "" || len(req.DTPassword) < 8 {
		writeJSONError(w, http.StatusBadRequest, "Raison sociale, email valide et mot de passe robuste (≥ 8 caractères) obligatoires")
		return
	}

	if req.OrganizationSlug == "" {
		req.OrganizationSlug = "laboratoire-principal"
	}
	if req.MIPHLicenseNumber == "" {
		req.MIPHLicenseNumber = "MIPH-2026-ENR-01"
	}

	// 2. Hash du mot de passe avec Argon2id (OWASP & 21 CFR Part 11)
	passwordHash, err := auth.HashPassword(req.DTPassword, nil)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur de chiffrement du mot de passe")
		return
	}

	// 3. Création transactionnelle de l'Organisation et du compte PDT
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur d'initialisation de transaction")
		return
	}
	defer tx.Rollback(ctx)

	orgID := uuid.New()
	pdtName := fmt.Sprintf("Dr. %s %s", req.DTFirstName, req.DTLastName)

	_, err = tx.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (slug) DO UPDATE SET
			legal_name = EXCLUDED.legal_name,
			miph_license_number = EXCLUDED.miph_license_number,
			technical_director_name = EXCLUDED.technical_director_name
	`, orgID, req.OrganizationSlug, req.OrganizationName, req.MIPHLicenseNumber, pdtName)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Échec de création de l'établissement: %v", err))
		return
	}

	dtUserID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
	`, dtUserID, orgID, req.DTEmail, passwordHash, req.DTFirstName, req.DTLastName, string(domain.RoleDirecteurTechnique))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Échec de création du Pharmacien DT: %v", err))
		return
	}

	// Création de l'établissement pharmaceutique lié dans le référentiel
	estID := uuid.New()
	_, _ = tx.Exec(ctx, `
		INSERT INTO establishments (id, organization_id, legal_name_fr, establishment_type, miph_license_number, technical_director_name, technical_director_mandate)
		VALUES ($1, $2, $3, 'laboratoire_fabricant', $4, $5, $6)
	`, estID, orgID, req.OrganizationName, req.MIPHLicenseNumber, pdtName, req.DTDecisionNumber)

	if err := tx.Commit(ctx); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Échec de validation de la configuration")
		return
	}

	// 4. Scellement du premier jalon dans l'Audit Trail immuable BPF
	if h.auditService != nil {
		_, _ = h.auditService.Log(ctx, &audit.Entry{
			OrganizationID: orgID,
			UserID:         &dtUserID,
			UserEmail:      req.DTEmail,
			UserRole:       string(domain.RoleDirecteurTechnique),
			Action:         "SYSTEM_SETUP_WIZARD_COMPLETED",
			EntityType:     "ORGANIZATION",
			EntityID:       orgID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"legal_name":          req.OrganizationName,
				"miph_license":        req.MIPHLicenseNumber,
				"pdt_email":           req.DTEmail,
				"pdt_decision_number": req.DTDecisionNumber,
				"activity_type":       req.ActivityType,
				"completed_at":        time.Now().UTC().Format(time.RFC3339),
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":          "ok",
		"message":         "Instance initialisée avec succès. Le compte du Pharmacien Directeur Technique est opérationnel.",
		"organization_id": orgID.String(),
		"pdt_user_id":     dtUserID.String(),
		"pdt_email":       req.DTEmail,
		"pdt_name":        pdtName,
	})
}
