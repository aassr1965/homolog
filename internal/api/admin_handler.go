package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
)

// AdminHandler gère l'administration du laboratoire, des utilisateurs et du RBAC
type AdminHandler struct {
	pool           *pgxpool.Pool
	sessionManager *auth.SessionManager
	auditService   *audit.Service
}

// NewAdminHandler instancie le handler d'administration
func NewAdminHandler(pool *pgxpool.Pool, sm *auth.SessionManager, auditService *audit.Service) *AdminHandler {
	return &AdminHandler{
		pool:           pool,
		sessionManager: sm,
		auditService:   auditService,
	}
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
	CreatedAt          time.Time           `json:"created_at"`
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
			u.totp_enabled, u.is_active, u.last_login_at, u.created_at
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
			&item.IsActive, &item.LastLoginAt, &item.CreatedAt,
		); err == nil {
			// Charger les permissions effectives
			item.Permissions = h.getUserPermissions(r.Context(), item.ID, item.Role)
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

// DevSwitchRole permet de basculer instantanément de profil en mode développement sans mot de passe
func (h *AdminHandler) DevSwitchRole(w http.ResponseWriter, r *http.Request) {
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
		Secure:   false,
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
