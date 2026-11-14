package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/config"
	"homolog/internal/domain"
)

func TestSystemHandler_Metrics_And_Backup(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	cfg := config.Load()
	auditSvc := audit.NewService(db.Pool)
	sm := auth.NewSessionManager(db.Pool, 60, 15)
	sysHandler := NewSystemHandler(db.Pool, sm, auditSvc, cfg)

	orgID := uuid.New()
	adminID := uuid.New()

	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Labo Test Sys', 'MIPH-SYS-01', 'Dr. Sys')
		ON CONFLICT (id) DO NOTHING
	`, orgID, "org-sys-"+uuid.New().String()[:8])

	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active, permissions_version)
		VALUES ($1, $2, 'sysadmin@labo-test.dz', 'sys-hash', 'Karim', 'SysAdmin', 'admin_systeme_local', TRUE, 1)
	`, adminID, orgID)

	sysAdminUser := &auth.SessionUser{
		UserID:         adminID,
		Email:          "sysadmin@labo-test.dz",
		Role:           string(domain.RoleAdminSystemeLocal),
		OrganizationID: orgID,
		Permissions:    []string{string(domain.PermSysMaint)},
	}

	// 1. Consultation des télémétries système
	reqMetrics := httptest.NewRequest(http.MethodGet, "/api/v1/system/metrics", nil)
	reqMetrics = reqMetrics.WithContext(context.WithValue(reqMetrics.Context(), UserContextKey, sysAdminUser))
	wMetrics := httptest.NewRecorder()

	sysHandler.GetSystemMetrics(wMetrics, reqMetrics)
	if wMetrics.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour télémétries système, obtenu: %d (%s)", wMetrics.Code, wMetrics.Body.String())
	}

	var metricsRes map[string]any
	_ = json.Unmarshal(wMetrics.Body.Bytes(), &metricsRes)
	if metricsRes["database"] == nil || metricsRes["storage_cas"] == nil || metricsRes["server"] == nil {
		t.Fatalf("Composants manquants dans les métriques système : %v", metricsRes)
	}

	// 2. Déclenchement d'un instantané de sauvegarde
	reqBackup := httptest.NewRequest(http.MethodPost, "/api/v1/system/backup", nil)
	reqBackup = reqBackup.WithContext(context.WithValue(reqBackup.Context(), UserContextKey, sysAdminUser))
	wBackup := httptest.NewRecorder()

	sysHandler.TriggerBackup(wBackup, reqBackup)
	if wBackup.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour sauvegarde système, obtenu: %d (%s)", wBackup.Code, wBackup.Body.String())
	}

	var backupRes map[string]any
	_ = json.Unmarshal(wBackup.Body.Bytes(), &backupRes)
	seal, ok := backupRes["sha256_seal"].(string)
	if !ok || len(seal) != 64 {
		t.Fatalf("Empreinte SHA-256 de sauvegarde invalide : %v", seal)
	}

	// 3. Contrôle d'intégrité CAS
	reqCAS := httptest.NewRequest(http.MethodPost, "/api/v1/system/verify-integrity", nil)
	reqCAS = reqCAS.WithContext(context.WithValue(reqCAS.Context(), UserContextKey, sysAdminUser))
	wCAS := httptest.NewRecorder()

	sysHandler.VerifyCASIntegrity(wCAS, reqCAS)
	if wCAS.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour intégrité CAS, obtenu: %d (%s)", wCAS.Code, wCAS.Body.String())
	}

	// 4. Liste des sessions actives
	reqSessions := httptest.NewRequest(http.MethodGet, "/api/v1/system/sessions", nil)
	reqSessions = reqSessions.WithContext(context.WithValue(reqSessions.Context(), UserContextKey, sysAdminUser))
	wSessions := httptest.NewRecorder()

	sysHandler.ListActiveSessions(wSessions, reqSessions)
	if wSessions.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour liste des sessions, obtenu: %d (%s)", wSessions.Code, wSessions.Body.String())
	}

	// 5. Verrouillage d'urgence
	lockdownBody, _ := json.Marshal(map[string]string{"reason": "Test audit intrusion"})
	reqLockdown := httptest.NewRequest(http.MethodPost, "/api/v1/system/sessions/emergency-lockdown", bytes.NewReader(lockdownBody))
	reqLockdown = reqLockdown.WithContext(context.WithValue(reqLockdown.Context(), UserContextKey, sysAdminUser))
	wLockdown := httptest.NewRecorder()

	sysHandler.EmergencyLockdown(wLockdown, reqLockdown)
	if wLockdown.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour verrouillage d'urgence, obtenu: %d (%s)", wLockdown.Code, wLockdown.Body.String())
	}
}
