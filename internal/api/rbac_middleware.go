package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"homolog/internal/audit"
	"homolog/internal/domain"
)

// RequirePermission interdit l'accès si l'utilisateur ne possède pas la permission requise
// Conforme CDC Section 7.2.1 : Blocage hermétique HTTP 403 et journalisation d'alerte dans audit_logs
func RequirePermission(auditService *audit.Service, requiredPerm domain.Permission) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetSessionUser(r.Context())
			if user == nil {
				writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
				return
			}

			// 1. Vérification de la permission
			if !user.HasPermission(string(requiredPerm)) {
				// Journalisation d'alerte dans l'Audit Trail immuable
				if auditService != nil {
					_, _ = auditService.Log(r.Context(), &audit.Entry{
						OrganizationID: user.OrganizationID,
						UserID:         &user.UserID,
						UserEmail:      user.Email,
						UserRole:       user.Role,
						Action:         "AUDIT_SECURITY_403_BLOCKED",
						EntityType:     "SECURITY_RBAC",
						EntityID:       string(requiredPerm),
						IPAddress:      getClientIP(r),
						UserAgent:      r.UserAgent(),
						Details: map[string]any{
							"reason":              "Tentative d'accès non autorisée à une ressource protégée",
							"required_permission": requiredPerm,
							"user_permissions":    user.Permissions,
							"http_method":         r.Method,
							"request_uri":         r.URL.Path,
						},
					})
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":               "Accès réglementaire refusé : permissions insuffisantes",
					"message":             fmt.Sprintf("Votre rôle actuel (%s) ne dispose pas de la permission requise [%s] pour effectuer cette action réglementaire (Ségrégation des Fonctions BPF / 21 CFR Part 11).", user.Role, requiredPerm),
					"code":                "ERR_RBAC_FORBIDDEN",
					"required_permission": requiredPerm,
					"user_role":           user.Role,
				})
				return
			}

			// 2. Garde-fou BPF Zero-Trust : Interdiction de lecture documentaire pour admin technique
			if requiredPerm == domain.PermDocRead {
				if user.Role == string(domain.RoleAdminSystemeLocal) || user.Role == string(domain.RoleSuperAdminEditeur) {
					if auditService != nil {
						_, _ = auditService.Log(r.Context(), &audit.Entry{
							OrganizationID: user.OrganizationID,
							UserID:         &user.UserID,
							UserEmail:      user.Email,
							UserRole:       user.Role,
							Action:         "AUDIT_SECURITY_ZEROTRUST_BLOCKED",
							EntityType:     "DOCUMENT_CONFIDENTIAL",
							EntityID:       r.URL.Path,
							IPAddress:      getClientIP(r),
							UserAgent:      r.UserAgent(),
							Details: map[string]any{
								"reason": "Règle Zero-Trust : Les administrateurs techniques n'ont pas accès aux données de santé et documents pharmaceutiques",
								"role":   user.Role,
							},
						})
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"error": "Accès Zero-Trust refusé : Les administrateurs techniques n'ont pas accès au contenu des pièces pharmaceutiques",
						"code":  "ERR_ZERO_TRUST_DENIED",
					})
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission autorise l'accès si l'utilisateur possède au moins une des permissions listées
func RequireAnyPermission(auditService *audit.Service, requiredPerms ...domain.Permission) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetSessionUser(r.Context())
			if user == nil {
				writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
				return
			}

			for _, perm := range requiredPerms {
				if user.HasPermission(string(perm)) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Aucune permission correspondante
			if auditService != nil {
				_, _ = auditService.Log(r.Context(), &audit.Entry{
					OrganizationID: user.OrganizationID,
					UserID:         &user.UserID,
					UserEmail:      user.Email,
					UserRole:       user.Role,
					Action:         "AUDIT_SECURITY_403_BLOCKED",
					EntityType:     "SECURITY_RBAC",
					EntityID:       fmt.Sprintf("%v", requiredPerms),
					IPAddress:      getClientIP(r),
					UserAgent:      r.UserAgent(),
					Details: map[string]any{
						"reason":         "Aucune des permissions requises n'est détenue",
						"required_perms": requiredPerms,
						"http_method":    r.Method,
						"request_uri":    r.URL.Path,
					},
				})
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":          "Accès réglementaire refusé : permissions insuffisantes",
				"code":           "ERR_RBAC_FORBIDDEN",
				"required_perms": requiredPerms,
			})
		})
	}
}
