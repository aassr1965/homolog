package api

import (
	"context"
	"net/http"
	"strings"

	"homolog/internal/auth"
	"homolog/internal/domain"
)

type contextKey string

const (
	UserContextKey contextKey = "homolog_session_user"
	CookieName                = "homolog_session"
)

// AuthMiddleware vérifie le cookie HttpOnly ou l'en-tête Authorization
func AuthMiddleware(sm *auth.SessionManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, `{"error":"Authentification requise"}`, http.StatusUnauthorized)
				return
			}

			user, err := sm.ValidateSession(r.Context(), token)
			if err != nil {
				// Gestion des erreurs spécifiques
				switch err {
				case auth.ErrSessionIdle:
					http.Error(w, `{"error":"Session verrouillée pour inactivité (15 min BPF)","code":"SESSION_IDLE"}`, http.StatusUnauthorized)
				case auth.ErrSessionExpired:
					http.Error(w, `{"error":"Session expirée","code":"SESSION_EXPIRED"}`, http.StatusUnauthorized)
				case auth.ErrUserLocked:
					http.Error(w, `{"error":"Compte temporairement verrouillé","code":"USER_LOCKED"}`, http.StatusForbidden)
				default:
					http.Error(w, `{"error":"Session invalide ou révoquée"}`, http.StatusUnauthorized)
				}
				return
			}

			// Injection du SessionUser dans le contexte
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole interdit l'accès si l'utilisateur ne possède pas l'un des rôles requis
func RequireRole(allowedRoles ...domain.UserRole) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserContextKey).(*auth.SessionUser)
			if !ok || user == nil {
				http.Error(w, `{"error":"Non authentifié"}`, http.StatusUnauthorized)
				return
			}

			for _, role := range allowedRoles {
				if string(role) == user.Role {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error":"Accès refusé : privilège réglementaire insuffisant pour votre rôle"}`, http.StatusForbidden)
		})
	}
}

// GetSessionUser extrait le SessionUser du contexte de la requête
func GetSessionUser(ctx context.Context) *auth.SessionUser {
	if val, ok := ctx.Value(UserContextKey).(*auth.SessionUser); ok {
		return val
	}
	return nil
}

// extractToken extrait le token du cookie ou du header Authorization
func extractToken(r *http.Request) string {
	// 1. Priorité au Cookie HttpOnly
	if cookie, err := r.Cookie(CookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 2. Repli sur Header Authorization: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}
