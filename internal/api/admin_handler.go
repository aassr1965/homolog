package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/config"
	"homolog/internal/domain"
)

// AdminHandler gère l'administration du laboratoire, des utilisateurs et du RBAC
type AdminHandler struct {
	pool           *pgxpool.Pool
	sessionManager *auth.SessionManager
	auditService   *audit.Service
	cfg            *config.Config
}

// NewAdminHandler instancie le handler d'administration
func NewAdminHandler(pool *pgxpool.Pool, sm *auth.SessionManager, auditService *audit.Service, cfg ...*config.Config) *AdminHandler {
	var c *config.Config
	if len(cfg) > 0 {
		c = cfg[0]
	}
	return &AdminHandler{
		pool:           pool,
		sessionManager: sm,
		auditService:   auditService,
		cfg:            c,
	}
}

// UserPermOverride représente une dérogation de droit unitaire sur un profil
type UserPermOverride struct {
	PermissionID string `json:"permission_id"`
	Granted      bool   `json:"granted"`
}

// UserListItem représente un utilisateur dans la console d'administration
type UserListItem struct {
	ID                 uuid.UUID           `json:"id"`
	Email              string              `json:"email"`
	FirstName          string              `json:"first_name"`
	LastName           string              `json:"last_name"`
	Role               string              `json:"role"`
	RoleNameFR         string              `json:"role_name_fr"`
	Permissions        []domain.Permission `json:"permissions"`
	PermissionsVersion int                 `json:"permissions_version"`
	TOTPEnabled        bool                `json:"totp_enabled"`
	IsActive           bool                `json:"is_active"`
	LastLoginAt        *time.Time          `json:"last_login_at,omitempty"`
	ContractExpiresAt  *time.Time          `json:"contract_expires_at,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
	Overrides          []UserPermOverride  `json:"overrides"`
}

// ListUsers liste les collaborateurs du laboratoire avec leurs rôles et permissions effectives
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	query := `
		SELECT 
			u.id, u.email, u.first_name, u.last_name, u.role,
			COALESCE(r.name_fr, u.role) as role_name_fr,
			COALESCE(u.permissions_version, 1),
			u.totp_enabled, u.is_active, u.last_login_at, u.contract_expires_at, u.created_at
		FROM users u
		LEFT JOIN roles r ON u.role = r.id
		WHERE u.organization_id = $1
		ORDER BY u.created_at ASC
	`

	rows, err := h.pool.Query(r.Context(), query, user.OrganizationID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur lors du chargement des utilisateurs")
		return
	}
	defer rows.Close()

	var users []UserListItem
	for rows.Next() {
		var item UserListItem
		if err := rows.Scan(
			&item.ID, &item.Email, &item.FirstName, &item.LastName, &item.Role,
			&item.RoleNameFR, &item.PermissionsVersion, &item.TOTPEnabled,
			&item.IsActive, &item.LastLoginAt, &item.ContractExpiresAt, &item.CreatedAt,
		); err == nil {
			// Charger les permissions effectives et les dérogations individuelles
			item.Permissions = h.getUserPermissions(r.Context(), item.ID, item.Role)
			item.Overrides = h.getUserOverrides(r.Context(), item.ID)
			users = append(users, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"users": users,
		"count": len(users),
	})
}

// ListRoles retourne la liste des rôles avec leur matrice de permissions par défaut
func (h *AdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	rolesQuery := `SELECT id, name_fr, COALESCE(name_ar, ''), COALESCE(description, ''), is_system, created_at FROM roles ORDER BY is_system DESC, id ASC`
	rows, err := h.pool.Query(r.Context(), rolesQuery)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur base de données rôles")
		return
	}
	defer rows.Close()

	var roles []domain.RoleInfo
	for rows.Next() {
		var role domain.RoleInfo
		if err := rows.Scan(&role.ID, &role.NameFR, &role.NameAR, &role.Description, &role.IsSystem, &role.CreatedAt); err == nil {
			role.Permissions = h.getRolePermissions(r.Context(), role.ID)
			roles = append(roles, role)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"roles": roles,
		"count": len(roles),
	})
}

// ListPermissions retourne le catalogue des 14 permissions unitaires
func (h *AdminHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name_fr, COALESCE(description, ''), category, created_at FROM permissions ORDER BY category ASC, id ASC`
	rows, err := h.pool.Query(r.Context(), query)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur base de données permissions")
		return
	}
	defer rows.Close()

	var perms []domain.PermissionInfo
	for rows.Next() {
		var p domain.PermissionInfo
		if err := rows.Scan(&p.ID, &p.NameFR, &p.Description, &p.Category, &p.CreatedAt); err == nil {
			perms = append(perms, p)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"permissions": perms,
		"count":       len(perms),
	})
}

// UpdateUserRole modifie le rôle d'un collaborateur et incrémente le versioning de permissions
func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant utilisateur invalide")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.Role = strings.TrimSpace(req.Role)
	if req.Role == "" {
		writeJSONError(w, http.StatusBadRequest, "Le rôle est obligatoire")
		return
	}

	// Vérifier l'existence du rôle
	var roleExists bool
	_ = h.pool.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`, req.Role).Scan(&roleExists)
	if !roleExists {
		writeJSONError(w, http.StatusBadRequest, "Rôle non reconnu dans le système")
		return
	}

	// Mise à jour et incrémentation de permissions_version pour invalidation immédiate du cache
	updateQuery := `
		UPDATE users 
		SET role = $1, permissions_version = permissions_version + 1, updated_at = NOW() 
		WHERE id = $2 AND organization_id = $3
		RETURNING email
	`
	var targetEmail string
	err = h.pool.QueryRow(r.Context(), updateQuery, req.Role, targetUserID, user.OrganizationID).Scan(&targetEmail)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Impossible de modifier le rôle")
		return
	}

	// Audit trail
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "USER_ROLE_UPDATED",
			EntityType:     "USER",
			EntityID:       targetUserID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"target_email": targetEmail,
				"new_role":     req.Role,
				"updated_by":   user.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"message": "Rôle utilisateur mis à jour avec succès",
		"user_id": targetUserID,
		"role":    req.Role,
	})
}

// DevSwitchRole permet de basculer instantanément de profil en mode développement ou preview protégée sans mot de passe.
// Strictement verrouillé et interdit en environnement de production sans sas confidentiel (CodeRabbit PR #5 review).
func (h *AdminHandler) DevSwitchRole(w http.ResponseWriter, r *http.Request) {
	// Guard défense en profondeur : vérification de l'activation du Dev Switcher
	devEnabled := false
	if h.cfg != nil {
		devEnabled = h.cfg.DevSwitcherEnabled()
	} else {
		devEnabled = config.Load().DevSwitcherEnabled()
	}
	if !devEnabled {
		writeJSONError(w, http.StatusNotFound, "Endpoint désactivé en environnement de production")
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	targetEmail := strings.ToLower(strings.TrimSpace(req.Email))
	if targetEmail == "" {
		writeJSONError(w, http.StatusBadRequest, "L'adresse email est obligatoire")
		return
	}

	// Trouver l'utilisateur
	var (
		userID      uuid.UUID
		orgID       uuid.UUID
		firstName   string
		lastName    string
		role        string
		totpEnabled bool
		orgSlug     string
		orgName     string
	)

	query := `
		SELECT u.id, u.organization_id, u.first_name, u.last_name, u.role, u.totp_enabled, o.slug, o.legal_name
		FROM users u
		JOIN organizations o ON u.organization_id = o.id
		WHERE u.email = $1 AND u.is_active = TRUE
	`
	err := h.pool.QueryRow(r.Context(), query, targetEmail).Scan(
		&userID, &orgID, &firstName, &lastName, &role, &totpEnabled, &orgSlug, &orgName,
	)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Utilisateur introuvable pour la bascule de test")
		return
	}

	// Créer une nouvelle session active
	token, err := h.sessionManager.CreateSession(r.Context(), userID, getClientIP(r), r.UserAgent())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Impossible de créer la session de test")
		return
	}

	now := time.Now().UTC()
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		Expires:  now.Add(8 * time.Hour),
		MaxAge:   8 * 3600,
		HttpOnly: true,
		Secure:   DetermineCookieSecure(r, h.cfg),
		SameSite: http.SameSiteStrictMode,
	})

	sessionUser, _ := h.sessionManager.ValidateSession(r.Context(), token)
	var perms []string
	var permVersion int = 1
	if sessionUser != nil {
		perms = sessionUser.Permissions
		permVersion = sessionUser.PermissionsVersion
	}

	// Audit trail
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: orgID,
			UserID:         &userID,
			UserEmail:      targetEmail,
			UserRole:       role,
			Action:         "DEV_ROLE_SWITCH",
			EntityType:     "SESSION",
			EntityID:       userID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"switched_to_email": targetEmail,
				"switched_to_role":  role,
				"mode":              "development_role_switcher",
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"message": "Bascule de profil effectuée avec succès (Environnement Dev)",
		"user": map[string]any{
			"id":                    userID,
			"email":                 targetEmail,
			"first_name":            firstName,
			"last_name":             lastName,
			"role":                  role,
			"permissions":           perms,
			"permissions_version":   permVersion,
			"is_technical_director": role == string(domain.RoleDirecteurTechnique),
			"has_2fa":               totpEnabled,
		},
		"organization": map[string]any{
			"id":   orgID,
			"slug": orgSlug,
			"name": orgName,
		},
	})
}

// getUserPermissions charge les permissions d'un utilisateur
func (h *AdminHandler) getUserPermissions(ctx context.Context, userID uuid.UUID, role string) []domain.Permission {
	query := `
		SELECT permission_id FROM role_permissions WHERE role_id = $1
		UNION
		SELECT permission_id FROM user_permissions WHERE user_id = $2 AND granted = TRUE
		EXCEPT
		SELECT permission_id FROM user_permissions WHERE user_id = $2 AND granted = FALSE
	`
	rows, err := h.pool.Query(ctx, query, role, userID)
	if err != nil {
		return []domain.Permission{}
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p); err == nil {
			perms = append(perms, p)
		}
	}
	if perms == nil {
		perms = []domain.Permission{}
	}
	return perms
}

// getRolePermissions charge les permissions d'un rôle
func (h *AdminHandler) getRolePermissions(ctx context.Context, roleID string) []domain.Permission {
	query := `SELECT permission_id FROM role_permissions WHERE role_id = $1 ORDER BY permission_id ASC`
	rows, err := h.pool.Query(ctx, query, roleID)
	if err != nil {
		return []domain.Permission{}
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p); err == nil {
			perms = append(perms, p)
		}
	}
	if perms == nil {
		perms = []domain.Permission{}
	}
	return perms
}

// getUserOverrides charge les dérogations accordées ou révoquées pour un collaborateur
func (h *AdminHandler) getUserOverrides(ctx context.Context, userID uuid.UUID) []UserPermOverride {
	query := `SELECT permission_id, granted FROM user_permissions WHERE user_id = $1 ORDER BY permission_id ASC`
	rows, err := h.pool.Query(ctx, query, userID)
	if err != nil {
		return []UserPermOverride{}
	}
	defer rows.Close()

	var overrides []UserPermOverride
	for rows.Next() {
		var o UserPermOverride
		if err := rows.Scan(&o.PermissionID, &o.Granted); err == nil {
			overrides = append(overrides, o)
		}
	}
	if overrides == nil {
		overrides = []UserPermOverride{}
	}
	return overrides
}

// CreateUserRequest représente la charge utile pour créer un collaborateur
type CreateUserRequest struct {
	Email             string     `json:"email"`
	FirstName         string     `json:"first_name"`
	LastName          string     `json:"last_name"`
	Role              string     `json:"role"`
	Password          string     `json:"password"`
	ContractExpiresAt *time.Time `json:"contract_expires_at,omitempty"`
}

// CreateUser crée un nouveau collaborateur au sein de l'organisation
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Role = strings.TrimSpace(req.Role)

	if req.Email == "" || !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		writeJSONError(w, http.StatusBadRequest, "Adresse email invalide")
		return
	}
	if req.FirstName == "" || req.LastName == "" {
		writeJSONError(w, http.StatusBadRequest, "Le prénom et le nom sont obligatoires")
		return
	}
	if req.Role == "" {
		writeJSONError(w, http.StatusBadRequest, "Le rôle BPF est obligatoire")
		return
	}

	// Vérification de l'existence du rôle
	var roleExists bool
	_ = h.pool.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`, req.Role).Scan(&roleExists)
	if !roleExists {
		writeJSONError(w, http.StatusBadRequest, "Rôle non reconnu dans le système")
		return
	}

	// Consultant externe : date de fin de contrat obligatoire et dans le futur (PSSI / Annexe 11 BPF)
	if req.Role == string(domain.RoleConsultantExterne) {
		if req.ContractExpiresAt == nil || req.ContractExpiresAt.Before(time.Now()) {
			writeJSONError(w, http.StatusBadRequest, "Pour un consultant externe, une date d'expiration de contrat future est obligatoire (PSSI / Annexe 11 BPF)")
			return
		}
	}

	// Validation de la politique de mot de passe BPF / 21 CFR Part 11
	if err := auth.ValidatePassword(req.Password); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("Mot de passe non conforme : %v", err))
		return
	}

	// Unicité de l'email au sein de l'organisation
	var emailTaken bool
	_ = h.pool.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM users WHERE organization_id = $1 AND email = $2)`, admin.OrganizationID, req.Email).Scan(&emailTaken)
	if emailTaken {
		writeJSONError(w, http.StatusConflict, "Un collaborateur avec cette adresse email existe déjà dans le laboratoire")
		return
	}

	// Hachage Argon2id
	hashedPassword, err := auth.HashPassword(req.Password, nil)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur lors du hachage de sécurité du mot de passe")
		return
	}

	var newUserID uuid.UUID
	var createdAt time.Time
	insertQuery := `
		INSERT INTO users (
			organization_id, email, password_hash, first_name, last_name, role,
			contract_expires_at, is_active, permissions_version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE, 1, NOW(), NOW())
		RETURNING id, created_at
	`
	err = h.pool.QueryRow(r.Context(), insertQuery,
		admin.OrganizationID, req.Email, hashedPassword, req.FirstName, req.LastName, req.Role, req.ContractExpiresAt,
	).Scan(&newUserID, &createdAt)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur base de données lors de la création de l'utilisateur")
		return
	}

	// Audit trail ALCOA+
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: admin.OrganizationID,
			UserID:         &admin.UserID,
			UserEmail:      admin.Email,
			UserRole:       admin.Role,
			Action:         "USER_CREATED",
			EntityType:     "USER",
			EntityID:       newUserID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"created_email": req.Email,
				"created_name":  req.FirstName + " " + req.LastName,
				"role":          req.Role,
				"has_contract":  req.ContractExpiresAt != nil,
				"created_by":    admin.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":     "ok",
		"message":    "Collaborateur créé avec succès",
		"user_id":    newUserID,
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
		"role":       req.Role,
		"is_active":  true,
		"created_at": createdAt,
	})
}

// UpdateUserStatusRequest représente la charge utile pour modifier le statut d'un collaborateur
type UpdateUserStatusRequest struct {
	IsActive bool   `json:"is_active"`
	Reason   string `json:"reason"`
}

// UpdateUserStatus active ou suspend un compte utilisateur sans suppression physique
func (h *AdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant utilisateur invalide")
		return
	}

	var req UpdateUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.Reason = strings.TrimSpace(req.Reason)
	if len(req.Reason) < 3 {
		writeJSONError(w, http.StatusBadRequest, "Un motif de modification de statut est requis (BPF / 21 CFR Part 11, min 3 caractères)")
		return
	}

	// Auto-désactivation interdite
	if targetUserID == admin.UserID && !req.IsActive {
		writeJSONError(w, http.StatusBadRequest, "Auto-désactivation interdite : vous ne pouvez pas suspendre votre propre compte d'administrateur")
		return
	}

	updateQuery := `
		UPDATE users 
		SET is_active = $1, permissions_version = permissions_version + 1, updated_at = NOW() 
		WHERE id = $2 AND organization_id = $3
		RETURNING email, role
	`
	var targetEmail, targetRole string
	err = h.pool.QueryRow(r.Context(), updateQuery, req.IsActive, targetUserID, admin.OrganizationID).Scan(&targetEmail, &targetRole)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Utilisateur introuvable dans votre laboratoire")
		return
	}

	// Si désactivation : révocation immédiate de toutes les sessions actives (21 CFR Part 11)
	if !req.IsActive && h.sessionManager != nil {
		_ = h.sessionManager.RevokeAllUserSessions(r.Context(), targetUserID)
	}

	// Audit trail ALCOA+
	if h.auditService != nil {
		action := "USER_ACTIVATED"
		if !req.IsActive {
			action = "USER_DEACTIVATED"
		}
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: admin.OrganizationID,
			UserID:         &admin.UserID,
			UserEmail:      admin.Email,
			UserRole:       admin.Role,
			Action:         action,
			EntityType:     "USER",
			EntityID:       targetUserID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"target_email": targetEmail,
				"target_role":  targetRole,
				"is_active":    req.IsActive,
				"reason":       req.Reason,
				"updated_by":   admin.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":    "ok",
		"message":   fmt.Sprintf("Statut utilisateur mis à jour : is_active=%v", req.IsActive),
		"user_id":   targetUserID,
		"is_active": req.IsActive,
	})
}

// ResetPasswordRequest représente la charge utile pour réinitialiser un mot de passe
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
	Reason      string `json:"reason"`
}

// ResetUserPassword réinitialise le mot de passe d'un utilisateur par l'administrateur
func (h *AdminHandler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant utilisateur invalide")
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.Reason = strings.TrimSpace(req.Reason)
	if len(req.Reason) < 3 {
		writeJSONError(w, http.StatusBadRequest, "Un motif de réinitialisation est requis (BPF / 21 CFR Part 11)")
		return
	}

	if err := auth.ValidatePassword(req.NewPassword); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("Nouveau mot de passe non conforme : %v", err))
		return
	}

	hashedPassword, err := auth.HashPassword(req.NewPassword, nil)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur lors du hachage du mot de passe")
		return
	}

	updateQuery := `
		UPDATE users 
		SET password_hash = $1, failed_login_attempts = 0, locked_until = NULL, permissions_version = permissions_version + 1, updated_at = NOW() 
		WHERE id = $2 AND organization_id = $3
		RETURNING email
	`
	var targetEmail string
	err = h.pool.QueryRow(r.Context(), updateQuery, hashedPassword, targetUserID, admin.OrganizationID).Scan(&targetEmail)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Utilisateur introuvable")
		return
	}

	// Révocation de toutes les sessions actives pour forcer la reconnexion avec le nouveau mot de passe
	if h.sessionManager != nil {
		_ = h.sessionManager.RevokeAllUserSessions(r.Context(), targetUserID)
	}

	// Audit trail ALCOA+
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: admin.OrganizationID,
			UserID:         &admin.UserID,
			UserEmail:      admin.Email,
			UserRole:       admin.Role,
			Action:         "USER_PASSWORD_RESET",
			EntityType:     "USER",
			EntityID:       targetUserID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"target_email": targetEmail,
				"reason":       req.Reason,
				"reset_by":     admin.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"message": "Mot de passe réinitialisé avec succès",
		"user_id": targetUserID,
	})
}

// ResetTOTPRequest représente la charge utile pour réinitialiser le 2FA
type ResetTOTPRequest struct {
	Reason string `json:"reason"`
}

// ResetUserTOTP réinitialise l'authentification à deux facteurs en cas de perte de terminal
func (h *AdminHandler) ResetUserTOTP(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant utilisateur invalide")
		return
	}

	var req ResetTOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.Reason = strings.TrimSpace(req.Reason)
	if len(req.Reason) < 3 {
		writeJSONError(w, http.StatusBadRequest, "Un motif de réinitialisation 2FA est requis")
		return
	}

	updateQuery := `
		UPDATE users 
		SET totp_enabled = FALSE, totp_secret = NULL, recovery_codes = NULL, permissions_version = permissions_version + 1, updated_at = NOW() 
		WHERE id = $1 AND organization_id = $2
		RETURNING email
	`
	var targetEmail string
	err = h.pool.QueryRow(r.Context(), updateQuery, targetUserID, admin.OrganizationID).Scan(&targetEmail)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Utilisateur introuvable")
		return
	}

	// Audit trail ALCOA+
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: admin.OrganizationID,
			UserID:         &admin.UserID,
			UserEmail:      admin.Email,
			UserRole:       admin.Role,
			Action:         "USER_TOTP_RESET",
			EntityType:     "USER",
			EntityID:       targetUserID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"target_email": targetEmail,
				"reason":       req.Reason,
				"reset_by":     admin.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"message": "Double facteur (2FA / TOTP) réinitialisé avec succès",
		"user_id": targetUserID,
	})
}

// OverridePermissionItem représente une dérogation de droit unitaire
type OverridePermissionItem struct {
	PermissionID string `json:"permission_id"`
	Granted      *bool  `json:"granted"` // true: accordé, false: révoqué, nil: suppression dérogation
}

// UpdateUserPermissionsRequest représente la charge utile pour modifier les surcharges de permissions
type UpdateUserPermissionsRequest struct {
	Overrides []OverridePermissionItem `json:"overrides"`
	Reason    string                   `json:"reason"`
}

// UpdateUserPermissions configure les surcharges de permissions unitaires pour un collaborateur
func (h *AdminHandler) UpdateUserPermissions(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	targetUserIDStr := chi.URLParam(r, "id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant utilisateur invalide")
		return
	}

	var req UpdateUserPermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.Reason = strings.TrimSpace(req.Reason)
	if len(req.Reason) < 3 {
		writeJSONError(w, http.StatusBadRequest, "Un motif de dérogation de droits est requis (BPF / 21 CFR Part 11)")
		return
	}

	// Vérifier l'existence de l'utilisateur dans l'organisation
	var targetEmail, targetRole string
	err = h.pool.QueryRow(r.Context(), `SELECT email, role FROM users WHERE id = $1 AND organization_id = $2`, targetUserID, admin.OrganizationID).Scan(&targetEmail, &targetRole)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Utilisateur introuvable dans votre laboratoire")
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur début de transaction")
		return
	}
	defer tx.Rollback(r.Context())

	for _, item := range req.Overrides {
		item.PermissionID = strings.TrimSpace(item.PermissionID)
		if item.PermissionID == "" {
			continue
		}

		if item.Granted == nil {
			// Suppression de la surcharge (retour au droit par défaut du rôle)
			_, err = tx.Exec(r.Context(), `DELETE FROM user_permissions WHERE user_id = $1 AND permission_id = $2`, targetUserID, item.PermissionID)
		} else {
			// Insertion ou mise à jour de la dérogation
			upsertQuery := `
				INSERT INTO user_permissions (user_id, permission_id, granted, assigned_by, created_at)
				VALUES ($1, $2, $3, $4, NOW())
				ON CONFLICT (user_id, permission_id) 
				DO UPDATE SET granted = EXCLUDED.granted, assigned_by = EXCLUDED.assigned_by, created_at = NOW()
			`
			_, err = tx.Exec(r.Context(), upsertQuery, targetUserID, item.PermissionID, *item.Granted, admin.UserID)
		}
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur lors de l'enregistrement de la permission %s", item.PermissionID))
			return
		}
	}

	// Bump version pour invalidation immédiate
	_, err = tx.Exec(r.Context(), `UPDATE users SET permissions_version = permissions_version + 1, updated_at = NOW() WHERE id = $1`, targetUserID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur mise à jour version permissions")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur validation de la transaction")
		return
	}

	// Audit trail ALCOA+
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: admin.OrganizationID,
			UserID:         &admin.UserID,
			UserEmail:      admin.Email,
			UserRole:       admin.Role,
			Action:         "USER_PERMISSIONS_OVERRIDDEN",
			EntityType:     "USER",
			EntityID:       targetUserID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"target_email": targetEmail,
				"reason":       req.Reason,
				"overrides":    req.Overrides,
				"updated_by":   admin.Email,
			},
		})
	}

	effectivePerms := h.getUserPermissions(r.Context(), targetUserID, targetRole)
	overrides := h.getUserOverrides(r.Context(), targetUserID)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "ok",
		"message":     "Permissions personnalisées enregistrées avec succès",
		"user_id":     targetUserID,
		"permissions": effectivePerms,
		"overrides":   overrides,
	})
}

// DisallowDeleteUser rejette formellement toute tentative de suppression physique (Hard Delete)
func (h *AdminHandler) DisallowDeleteUser(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusMethodNotAllowed, "Suppression physique interdite par la réglementation 21 CFR Part 11 et les BPF algériennes. Veuillez utiliser la désactivation / suspension de compte.")
}

// OrganizationSettingsResponse représente les paramètres globaux de l'organisation
type OrganizationSettingsResponse struct {
	ID                 uuid.UUID `json:"id"`
	Slug               string    `json:"slug"`
	LegalName          string    `json:"legal_name"`
	MIPHLicenseNumber  string    `json:"miph_license_number"`
	WorkingDaysPerWeek int       `json:"working_days_per_week"`
}

// UpdateOrganizationSettingsRequest payload de mise à jour des paramètres de l'organisation
type UpdateOrganizationSettingsRequest struct {
	WorkingDaysPerWeek int `json:"working_days_per_week"`
}

// GetOrganizationSettings retourne les paramètres généraux et réglementaires de l'organisation
func (h *AdminHandler) GetOrganizationSettings(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	query := `
		SELECT id, slug, legal_name, miph_license_number, COALESCE(working_days_per_week, 5)
		FROM organizations
		WHERE id = $1
	`
	var resp OrganizationSettingsResponse
	err := h.pool.QueryRow(r.Context(), query, admin.OrganizationID).Scan(
		&resp.ID, &resp.Slug, &resp.LegalName, &resp.MIPHLicenseNumber, &resp.WorkingDaysPerWeek,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Échec chargement des paramètres de l'organisation")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// UpdateOrganizationSettings met à jour les paramètres de l'organisation (ex: base de calcul des jours ouvrables)
func (h *AdminHandler) UpdateOrganizationSettings(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var req UpdateOrganizationSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Payload JSON invalide")
		return
	}

	if req.WorkingDaysPerWeek != 5 && req.WorkingDaysPerWeek != 6 {
		req.WorkingDaysPerWeek = 5
	}

	query := `
		UPDATE organizations
		SET working_days_per_week = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := h.pool.Exec(r.Context(), query, admin.OrganizationID, req.WorkingDaysPerWeek)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Échec enregistrement des paramètres de l'organisation")
		return
	}

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: admin.OrganizationID,
			UserID:         &admin.UserID,
			UserEmail:      admin.Email,
			UserRole:       admin.Role,
			Action:         "ORGANIZATION_SETTINGS_UPDATED",
			EntityType:     "ORGANIZATION",
			EntityID:       admin.OrganizationID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"working_days_per_week": req.WorkingDaysPerWeek,
				"updated_by":            admin.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":                "ok",
		"message":               "Paramètres de l'organisation mis à jour avec succès",
		"working_days_per_week": req.WorkingDaysPerWeek,
	})
}
