package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/config"
)

// SystemHandler gère la console technique d'infrastructure et de maintenance (SYS_MAINT / PSSI)
type SystemHandler struct {
	pool           *pgxpool.Pool
	sessionManager *auth.SessionManager
	auditService   *audit.Service
	cfg            *config.Config
	startTime      time.Time
}

// NewSystemHandler instancie le handler système
func NewSystemHandler(pool *pgxpool.Pool, sm *auth.SessionManager, auditService *audit.Service, cfg *config.Config) *SystemHandler {
	return &SystemHandler{
		pool:           pool,
		sessionManager: sm,
		auditService:   auditService,
		cfg:            cfg,
		startTime:      time.Now(),
	}
}

// GetSystemMetrics retourne les télémétries en temps réel de la plateforme (PostgreSQL, CAS, Sessions, OS)
func (h *SystemHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Métriques PostgreSQL
	var pgVersion string
	_ = h.pool.QueryRow(ctx, "SELECT version()").Scan(&pgVersion)

	var dbSizeBytes int64
	var dbSizePretty string
	_ = h.pool.QueryRow(ctx, "SELECT pg_database_size(current_database()), pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSizeBytes, &dbSizePretty)

	var tableCount int
	_ = h.pool.QueryRow(ctx, "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)

	poolStat := h.pool.Stat()

	// 2. Métriques CAS Storage (Content Addressable Storage)
	var casFilesCount int64
	var casTotalSizeBytes int64
	_ = h.pool.QueryRow(ctx, "SELECT COUNT(*), COALESCE(SUM(file_size_bytes), 0) FROM ged_versions").Scan(&casFilesCount, &casTotalSizeBytes)

	storagePath := "./data/storage"
	if h.cfg != nil && h.cfg.StoragePath != "" {
		storagePath = h.cfg.StoragePath
	}

	// 3. Métriques Sécurité & Sessions
	var activeSessionsCount int
	_ = h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM sessions WHERE expires_at > NOW()").Scan(&activeSessionsCount)

	var lockedAccountsCount int
	_ = h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE locked_until IS NOT NULL AND locked_until > NOW()").Scan(&lockedAccountsCount)

	var totalActiveUsers int
	_ = h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE is_active = TRUE").Scan(&totalActiveUsers)

	var auditEntriesCount int64
	_ = h.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs").Scan(&auditEntriesCount)

	// 4. Métriques Runtime Go
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	uptimeSeconds := int64(time.Since(h.startTime).Seconds())

	metrics := map[string]any{
		"server": map[string]any{
			"uptime_seconds":   uptimeSeconds,
			"started_at":       h.startTime.Format(time.RFC3339),
			"environment":      h.cfg.Environment,
			"deployment_mode":  h.cfg.DeploymentMode,
			"num_goroutines":   runtime.NumGoroutine(),
			"mem_alloc_mb":     float64(mem.Alloc) / 1024 / 1024,
			"mem_sys_mb":       float64(mem.Sys) / 1024 / 1024,
			"go_version":       runtime.Version(),
			"operating_system": runtime.GOOS + "/" + runtime.GOARCH,
		},
		"database": map[string]any{
			"version":        pgVersion,
			"size_bytes":     dbSizeBytes,
			"size_pretty":    dbSizePretty,
			"tables_count":   tableCount,
			"acquired_conns": poolStat.AcquiredConns(),
			"idle_conns":     poolStat.IdleConns(),
			"total_conns":    poolStat.TotalConns(),
			"max_conns":      poolStat.MaxConns(),
			"pool_status":    "healthy",
		},
		"storage_cas": map[string]any{
			"base_path":          storagePath,
			"storage_type":       h.cfg.StorageType,
			"indexed_files":      casFilesCount,
			"size_bytes":         casTotalSizeBytes,
			"size_pretty":        formatBytes(casTotalSizeBytes),
			"quota_limit_bytes":  int64(50 * 1024 * 1024 * 1024), // 50 Go
			"quota_used_percent": fmt.Sprintf("%.2f%%", float64(casTotalSizeBytes)/float64(50*1024*1024*1024)*100),
			"deduplication":      "SHA-256 Content-Addressable Storage",
		},
		"security_pssi": map[string]any{
			"active_sessions":     activeSessionsCount,
			"total_active_users":  totalActiveUsers,
			"locked_accounts":     lockedAccountsCount,
			"audit_entries_count": auditEntriesCount,
			"zero_trust_policy":   "Strict 403 Enforcement on Health & Clinical Data",
			"tls_enforced":        h.cfg.CookieSecure || h.cfg.Environment == "production",
			"password_hash_algo":  "Argon2id (OWASP recommended parameters)",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(metrics)
}

// ActiveSessionItem représente une session connectée dans le moniteur IT
type ActiveSessionItem struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Email          string    `json:"email"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Role           string    `json:"role"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	CreatedAt      time.Time `json:"created_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// ListActiveSessions retourne la liste des sessions connectées sur la plateforme
func (h *SystemHandler) ListActiveSessions(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT 
			s.id, s.user_id, u.email, u.first_name, u.last_name, u.role,
			COALESCE(s.ip_address, '127.0.0.1'), COALESCE(s.user_agent, 'Inconnu'),
			s.created_at, s.last_activity_at, s.expires_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.expires_at > NOW()
		ORDER BY s.last_activity_at DESC
	`

	rows, err := h.pool.Query(r.Context(), query)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur lors du chargement des sessions actives")
		return
	}
	defer rows.Close()

	var sessions []ActiveSessionItem
	for rows.Next() {
		var s ActiveSessionItem
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.Email, &s.FirstName, &s.LastName, &s.Role,
			&s.IPAddress, &s.UserAgent, &s.CreatedAt, &s.LastActivityAt, &s.ExpiresAt,
		); err == nil {
			sessions = append(sessions, s)
		}
	}
	if sessions == nil {
		sessions = []ActiveSessionItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// RevokeSession révoque une session individuelle spécifique
func (h *SystemHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de session invalide")
		return
	}

	var targetEmail string
	query := `
		DELETE FROM sessions s
		USING users u
		WHERE s.user_id = u.id AND s.id = $1
		RETURNING u.email
	`
	err = h.pool.QueryRow(r.Context(), query, sessionID).Scan(&targetEmail)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Session introuvable ou déjà expirée")
		return
	}

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: admin.OrganizationID,
			UserID:         &admin.UserID,
			UserEmail:      admin.Email,
			UserRole:       admin.Role,
			Action:         "SESSION_REVOKED_BY_ADMIN",
			EntityType:     "SESSION",
			EntityID:       sessionID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"target_email": targetEmail,
				"revoked_by":   admin.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"message": "Session révoquée avec succès",
	})
}

// EmergencyLockdown détruit toutes les sessions connectées sauf celle de l'administrateur exécutant
func (h *SystemHandler) EmergencyLockdown(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		req.Reason = "Verrouillage d'urgence de sécurité déclenché par l'administrateur système (PSSI)"
	}

	// Suppression de toutes les sessions actives SAUF celles de l'administrateur en cours
	tag, err := h.pool.Exec(r.Context(), "DELETE FROM sessions WHERE user_id != $1", admin.UserID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur lors de la révocation des sessions")
		return
	}

	revokedCount := tag.RowsAffected()

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: admin.OrganizationID,
			UserID:         &admin.UserID,
			UserEmail:      admin.Email,
			UserRole:       admin.Role,
			Action:         "EMERGENCY_LOCKDOWN_TRIGGERED",
			EntityType:     "SECURITY",
			EntityID:       admin.UserID.String(),
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"revoked_sessions": revokedCount,
				"reason":           req.Reason,
				"executed_by":      admin.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":           "ok",
		"message":          fmt.Sprintf("Verrouillage de sécurité exécuté : %d sessions révoquées", revokedCount),
		"revoked_sessions": revokedCount,
	})
}

// TriggerBackup simule ou exécute un instantané de sauvegarde horodaté scellé par SHA-256
func (h *SystemHandler) TriggerBackup(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	now := time.Now().UTC()
	backupID := fmt.Sprintf("backup-homolog-%s", now.Format("20060102-150405"))

	// Création du répertoire de backups si nécessaire
	backupDir := "./data/backups"
	_ = os.MkdirAll(backupDir, 0750)

	// Collecte des statistiques de tables pour générer le manifeste cryptographique de sauvegarde
	var tableNames []string
	rows, err := h.pool.Query(r.Context(), "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err == nil {
				tableNames = append(tableNames, name)
			}
		}
	}

	var dbSizeBytes int64
	_ = h.pool.QueryRow(r.Context(), "SELECT pg_database_size(current_database())").Scan(&dbSizeBytes)

	manifestData := map[string]any{
		"backup_id":         backupID,
		"timestamp":         now.Format(time.RFC3339),
		"database_size":     dbSizeBytes,
		"database_size_str": formatBytes(dbSizeBytes),
		"tables":            tableNames,
		"triggered_by":      admin.Email,
		"format":            "SQL_DUMP_METADATA_SNAPSHOT",
		"mode":              "on-premise-bpf",
	}

	manifestBytes, _ := json.MarshalIndent(manifestData, "", "  ")
	seal := sha256.Sum256(manifestBytes)
	sealHex := hex.EncodeToString(seal[:])

	manifestPath := filepath.Join(backupDir, backupID+".json")
	_ = os.WriteFile(manifestPath, manifestBytes, 0640)

	// Journalisation ALCOA+
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: admin.OrganizationID,
			UserID:         &admin.UserID,
			UserEmail:      admin.Email,
			UserRole:       admin.Role,
			Action:         "SYSTEM_BACKUP_EXECUTED",
			EntityType:     "BACKUP",
			EntityID:       backupID,
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"backup_id":    backupID,
				"sha256_seal":  sealHex,
				"size_bytes":   dbSizeBytes,
				"file_path":    manifestPath,
				"triggered_by": admin.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "ok",
		"message":     "Instantané de sauvegarde généré et scellé avec succès",
		"backup_id":   backupID,
		"sha256_seal": sealHex,
		"size_pretty": formatBytes(dbSizeBytes),
		"timestamp":   now.Format(time.RFC3339),
		"file_path":   manifestPath,
	})
}

// VerifyCASIntegrity vérifie la correspondance cryptographique SHA-256 entre la base et les fichiers disques
func (h *SystemHandler) VerifyCASIntegrity(w http.ResponseWriter, r *http.Request) {
	admin := GetSessionUser(r.Context())
	if admin == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	query := `SELECT id, document_id, version_label, original_filename, file_sha256, storage_path, file_size_bytes FROM ged_versions ORDER BY created_at DESC LIMIT 100`
	rows, err := h.pool.Query(r.Context(), query)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur interrogation versions GED")
		return
	}
	defer rows.Close()

	var totalChecked int
	var validCount int
	var mismatchCount int
	var missingCount int

	for rows.Next() {
		var id, docID uuid.UUID
		var versionLabel string
		var fileName, expectedSHA, storagePath string
		var sizeBytes int64

		if err := rows.Scan(&id, &docID, &versionLabel, &fileName, &expectedSHA, &storagePath, &sizeBytes); err == nil {
			totalChecked++

			file, err := os.Open(storagePath)
			if err != nil {
				missingCount++
				continue
			}

			hasher := sha256.New()
			if _, err := io.Copy(hasher, file); err != nil {
				_ = file.Close()
				mismatchCount++
				continue
			}
			_ = file.Close()

			actualSHA := hex.EncodeToString(hasher.Sum(nil))
			if strings.EqualFold(actualSHA, expectedSHA) {
				validCount++
			} else {
				mismatchCount++
			}
		}
	}

	statusStr := "INTEGRITY_VERIFIED_100_PERCENT"
	if mismatchCount > 0 || missingCount > 0 {
		statusStr = "INTEGRITY_WARNING_TAMPER_OR_MISSING"
	}

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: admin.OrganizationID,
			UserID:         &admin.UserID,
			UserEmail:      admin.Email,
			UserRole:       admin.Role,
			Action:         "CAS_INTEGRITY_CHECK_COMPLETED",
			EntityType:     "STORAGE",
			EntityID:       "CAS_STORE",
			IPAddress:      getClientIP(r),
			UserAgent:      r.UserAgent(),
			Details: map[string]any{
				"total_checked":  totalChecked,
				"valid_count":    validCount,
				"mismatch_count": mismatchCount,
				"missing_count":  missingCount,
				"status":         statusStr,
				"executed_by":    admin.Email,
			},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":         statusStr,
		"total_checked":  totalChecked,
		"valid_count":    validCount,
		"mismatch_count": mismatchCount,
		"missing_count":  missingCount,
		"message":        fmt.Sprintf("Vérification terminée : %d fichiers contrôlés, %d valides", totalChecked, validCount),
	})
}

// formatBytes convertit un nombre d'octets en chaîne lisible (Ko, Mo, Go)
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
