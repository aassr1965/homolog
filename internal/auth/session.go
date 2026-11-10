package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSessionNotFound = errors.New("session inexistante ou révoquée")
	ErrSessionExpired  = errors.New("session expirée (durée maximale atteinte)")
	ErrSessionIdle     = errors.New("session verrouillée pour inactivité (dépassement du délai BPF de 15 min)")
	ErrUserInactive    = errors.New("le compte utilisateur est désactivé")
	ErrUserLocked      = errors.New("le compte utilisateur est temporairement verrouillé suite à plusieurs échecs de connexion")
)

// SessionUser représente le contexte d'authentification injecté dans chaque requête
type SessionUser struct {
	SessionID            uuid.UUID
	UserID               uuid.UUID
	OrganizationID       uuid.UUID
	Email                string
	FirstName            string
	LastName             string
	Role                 string
	TOTPEnabled          bool
	FailedLoginAttempts  int
	LockedUntil          *time.Time
	OrganizationSlug     string
	OrganizationName     string
}

// SessionManager gère le cycle de vie des sessions en base PostgreSQL
type SessionManager struct {
	pool               *pgxpool.Pool
	sessionExpiry      time.Duration
	idleTimeout        time.Duration
}

// NewSessionManager instancie le gestionnaire de sessions
func NewSessionManager(pool *pgxpool.Pool, sessionExpiryMinutes, idleTimeoutMinutes int) *SessionManager {
	return &SessionManager{
		pool:          pool,
		sessionExpiry: time.Duration(sessionExpiryMinutes) * time.Minute,
		idleTimeout:   time.Duration(idleTimeoutMinutes) * time.Minute,
	}
}

// CreateSession génère un token opaque sécurisé, stocke son hash SHA-256 en base et retourne le token clair
func (sm *SessionManager) CreateSession(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) (string, error) {
	// Génération de 32 octets aléatoires cryptographiques
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("impossible de générer le jeton de session: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)

	// Hachage SHA-256 pour stockage sécurisé (même si la base est compromise, les tokens réels sont inviolables)
	tokenHash := hashSessionToken(rawToken)

	now := time.Now().UTC()
	expiresAt := now.Add(sm.sessionExpiry)

	query := `
		INSERT INTO sessions (user_id, token_hash, ip_address, user_agent, expires_at, last_activity_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := sm.pool.Exec(ctx, query, userID, tokenHash, ipAddress, userAgent, expiresAt, now, now)
	if err != nil {
		return "", fmt.Errorf("impossible d'enregistrer la session en base: %w", err)
	}

	return rawToken, nil
}

// ValidateSession vérifie la validité du token, le timeout d'inactivité BPF et rafraîchit l'activité
func (sm *SessionManager) ValidateSession(ctx context.Context, rawToken string) (*SessionUser, error) {
	if rawToken == "" {
		return nil, ErrSessionNotFound
	}

	tokenHash := hashSessionToken(rawToken)
	now := time.Now().UTC()

	query := `
		SELECT 
			s.id, s.user_id, s.expires_at, s.last_activity_at,
			u.organization_id, u.email, u.first_name, u.last_name, u.role, 
			u.totp_enabled, u.failed_login_attempts, u.locked_until, u.is_active,
			o.slug, o.legal_name
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		JOIN organizations o ON u.organization_id = o.id
		WHERE s.token_hash = $1
	`

	var (
		sessionID      uuid.UUID
		userID         uuid.UUID
		expiresAt      time.Time
		lastActivityAt time.Time
		orgID          uuid.UUID
		email          string
		firstName      string
		lastName       string
		role           string
		totpEnabled    bool
		failedAttempts int
		lockedUntil    *time.Time
		isActive       bool
		orgSlug        string
		orgName        string
	)

	err := sm.pool.QueryRow(ctx, query, tokenHash).Scan(
		&sessionID, &userID, &expiresAt, &lastActivityAt,
		&orgID, &email, &firstName, &lastName, &role,
		&totpEnabled, &failedAttempts, &lockedUntil, &isActive,
		&orgSlug, &orgName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("erreur de vérification de session: %w", err)
	}

	// 1. Contrôle compte actif
	if !isActive {
		_ = sm.RevokeSession(ctx, rawToken)
		return nil, ErrUserInactive
	}

	// 2. Contrôle verrouillage de sécurité
	if lockedUntil != nil && lockedUntil.After(now) {
		return nil, ErrUserLocked
	}

	// 3. Contrôle expiration absolue de la session
	if now.After(expiresAt) {
		_ = sm.RevokeSession(ctx, rawToken)
		return nil, ErrSessionExpired
	}

	// 4. Contrôle verrouillage sur inactivité (Idle Timeout BPF 15 min)
	if now.Sub(lastActivityAt) > sm.idleTimeout {
		_ = sm.RevokeSession(ctx, rawToken)
		return nil, ErrSessionIdle
	}

	// 5. Mise à jour de la dernière activité
	_, _ = sm.pool.Exec(ctx, "UPDATE sessions SET last_activity_at = $1 WHERE id = $2", now, sessionID)

	return &SessionUser{
		SessionID:           sessionID,
		UserID:              userID,
		OrganizationID:      orgID,
		Email:               email,
		FirstName:           firstName,
		LastName:            lastName,
		Role:                role,
		TOTPEnabled:         totpEnabled,
		FailedLoginAttempts: failedAttempts,
		LockedUntil:         lockedUntil,
		OrganizationSlug:    orgSlug,
		OrganizationName:    orgName,
	}, nil
}

// RevokeSession supprime une session unique (déconnexion)
func (sm *SessionManager) RevokeSession(ctx context.Context, rawToken string) error {
	tokenHash := hashSessionToken(rawToken)
	_, err := sm.pool.Exec(ctx, "DELETE FROM sessions WHERE token_hash = $1", tokenHash)
	return err
}

// RevokeAllUserSessions détruit toutes les sessions d'un utilisateur (changement mot de passe ou alerte sécurité)
func (sm *SessionManager) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := sm.pool.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
	return err
}

// CleanExpiredSessions supprime les sessions échues
func (sm *SessionManager) CleanExpiredSessions(ctx context.Context) (int64, error) {
	tag, err := sm.pool.Exec(ctx, "DELETE FROM sessions WHERE expires_at < NOW()")
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// hashSessionToken calcule le hash SHA-256 du jeton
func hashSessionToken(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return hex.EncodeToString(hasher.Sum(nil))
}
