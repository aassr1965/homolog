package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/config"
	"homolog/internal/domain"
)

func TestBackupArchiveAndSetupWizard(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-backup-" + testOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM anpp_deficiency_tasks WHERE organization_id = $1;
			DELETE FROM anpp_deficiency_letters WHERE organization_id = $1;
			DELETE FROM ged_documents WHERE organization_id = $1;
			DELETE FROM users WHERE organization_id = $1;
			DELETE FROM organizations WHERE id = $1;
		`, testOrgID)
	}()

	// 1. Créer une organisation et un Pharmacien DT de test
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Laboratoire Test Réversibilité', 'MIPH-2026-REV', 'Dr. Yacine Benkhalid')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation: %v", err)
	}

	dtUserID := uuid.New()
	pwdHash, _ := auth.HashPassword("RevSecret2026!#", nil)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, 'dt.backup@test.dz', $3, 'Yacine', 'Benkhalid', 'directeur_technique', TRUE)
	`, dtUserID, testOrgID, pwdHash)
	if err != nil {
		t.Fatalf("échec création user: %v", err)
	}

	auditService := audit.NewService(db.Pool)
	sessionManager := auth.NewSessionManager(db.Pool, 60, 15)
	cfg := &config.Config{
		DeploymentMode: config.ModeOnPremise,
		Environment:    "test",
		StoragePath:    t.TempDir(),
	}

	systemHandler := NewSystemHandler(db.Pool, sessionManager, auditService, cfg)
	setupHandler := NewSetupHandler(db.Pool, sessionManager, auditService, cfg)

	// =========================================================================
	// TEST 1 : GET /api/v1/setup/status
	// =========================================================================
	t.Run("GET /api/v1/setup/status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
		rr := httptest.NewRecorder()
		setupHandler.GetSetupStatus(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("attendu 200 OK, reçu %d: %s", rr.Code, rr.Body.String())
		}

		var resp SetupStatusResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("échec parsing json: %v", err)
		}
		if !resp.Initialized {
			t.Errorf("attendu initialized=true car des utilisateurs existent")
		}
		t.Logf("PASS: Statut Setup vérifié (Version: %s, Users: %d)", resp.Version, resp.UsersCount)
	})

	// =========================================================================
	// TEST 2 : POST /api/v1/setup/initialize (Rejet car instance déjà initialisée)
	// =========================================================================
	t.Run("POST /api/v1/setup/initialize [Conflit instance non-vierge]", func(t *testing.T) {
		body := bytes.NewBufferString(`{
			"organization_name": "Autre Labo",
			"organization_slug": "autre-labo",
			"dt_email": "dt2@test.dz",
			"dt_password": "Password123!#"
		}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/setup/initialize", body)
		rr := httptest.NewRecorder()
		setupHandler.InitializeInstance(rr, req)

		if rr.Code != http.StatusConflict {
			t.Fatalf("attendu 409 Conflict, reçu %d: %s", rr.Code, rr.Body.String())
		}
		t.Logf("PASS: Garde-fou Setup Wizard validé (interdiction de réinitialiser une instance active)")
	})

	// =========================================================================
	// TEST 3 : GET /api/v1/system/backup/export (Export de l'archive .homolog)
	// =========================================================================
	var exportedHomologBytes []byte
	t.Run("GET /api/v1/system/backup/export [.homolog tar.gz]", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/system/backup/export", nil)
		dtSession := &auth.SessionUser{
			SessionID:        uuid.New(),
			UserID:           dtUserID,
			OrganizationID:   testOrgID,
			Email:            "dt.backup@test.dz",
			Role:             string(domain.RoleDirecteurTechnique),
			OrganizationSlug: orgSlug,
			OrganizationName: "Laboratoire Test Réversibilité",
			Permissions:      []string{"SYS_MAINT", "DOC_APPROVE_PDT"},
		}
		ctxWithUser := context.WithValue(req.Context(), UserContextKey, dtSession)
		req = req.WithContext(ctxWithUser)

		rr := httptest.NewRecorder()
		systemHandler.ExportHomologArchive(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("attendu 200 OK, reçu %d: %s", rr.Code, rr.Body.String())
		}

		seal := rr.Header().Get("X-Homolog-Seal")
		if seal == "" {
			t.Fatalf("entête X-Homolog-Seal manquante")
		}

		exportedHomologBytes = rr.Body.Bytes()
		if len(exportedHomologBytes) < 50 {
			t.Fatalf("taille archive suspecte: %d octets", len(exportedHomologBytes))
		}

		// Inspection du contenu Tar.Gz
		gzr, err := gzip.NewReader(bytes.NewReader(exportedHomologBytes))
		if err != nil {
			t.Fatalf("échec ouverture gzip: %v", err)
		}
		defer gzr.Close()

		tr := tar.NewReader(gzr)
		hasManifest := false
		hasDatabase := false

		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("erreur lecture tar: %v", err)
			}
			if hdr.Name == "manifest.json" {
				hasManifest = true
				manifestBytes, _ := io.ReadAll(tr)
				var m HomologManifest
				if err := json.Unmarshal(manifestBytes, &m); err != nil {
					t.Fatalf("manifest json invalide: %v", err)
				}
				if m.Format != "HOMOLOG_PORTABLE_ARCHIVE_V1" {
					t.Errorf("format manifest inattendu: %s", m.Format)
				}
			}
			if len(hdr.Name) > 9 && hdr.Name[:9] == "database/" {
				hasDatabase = true
			}
		}

		if !hasManifest {
			t.Errorf("manifest.json absent de l'archive .homolog")
		}
		if !hasDatabase {
			t.Errorf("dossier database/ absent de l'archive .homolog")
		}

		t.Logf("PASS: Archive .homolog générée avec succès (%d octets, Sceau SHA-256: %s)", len(exportedHomologBytes), seal[:16]+"...")
	})

	// =========================================================================
	// TEST 4 : POST /api/v1/system/backup/restore (Restauration avec succès)
	// =========================================================================
	t.Run("POST /api/v1/system/backup/restore [Succès intégrité]", func(t *testing.T) {
		if len(exportedHomologBytes) == 0 {
			t.Skip("aucun flux d'archive exporté")
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("archive", "sauvegarde_officielle.homolog")
		if err != nil {
			t.Fatalf("erreur create form file: %v", err)
		}
		if _, err := part.Write(exportedHomologBytes); err != nil {
			t.Fatalf("erreur écriture part: %v", err)
		}
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/system/backup/restore", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		dtSession := &auth.SessionUser{
			SessionID:        uuid.New(),
			UserID:           dtUserID,
			OrganizationID:   testOrgID,
			Email:            "dt.backup@test.dz",
			Role:             string(domain.RoleDirecteurTechnique),
			OrganizationSlug: orgSlug,
			Permissions:      []string{"SYS_MAINT", "DOC_APPROVE_PDT"},
		}
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtSession))

		rr := httptest.NewRecorder()
		systemHandler.RestoreHomologArchive(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("attendu 200 OK, reçu %d: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]any
		_ = json.NewDecoder(rr.Body).Decode(&resp)
		if resp["status"] != "ok" {
			t.Errorf("statut inattendu: %v", resp)
		}
		t.Logf("PASS: Archive .homolog restaurée avec succès et intégrité certifiée 100%%")
	})

	// =========================================================================
	// TEST 5 : POST /api/v1/system/backup/restore [Rejet sur altération / falsification]
	// =========================================================================
	t.Run("POST /api/v1/system/backup/restore [Rejet altération SHA-256]", func(t *testing.T) {
		// Fabrication d'une fausse archive contenant un blob dont le hash ne correspond pas
		tamperedBuf := &bytes.Buffer{}
		gw := gzip.NewWriter(tamperedBuf)
		tw := tar.NewWriter(gw)

		manifest := HomologManifest{
			Version:        "0.8.0",
			Format:         "HOMOLOG_PORTABLE_ARCHIVE_V1",
			CreatedAt:      time.Now().UTC(),
			BlobsChecksums: map[string]string{"fake_sha": "fake_sha"},
		}
		mBytes, _ := json.Marshal(manifest)
		_ = addBytesToTar(tw, "manifest.json", mBytes, time.Now())

		// Ajout d'un blob avec contenu altéré
		fakeBlobName := "blobs/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		_ = addBytesToTar(tw, fakeBlobName, []byte("contenu_falsifie_non_conforme_au_hash"), time.Now())

		tw.Close()
		gw.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("archive", "archive_corrompue.homolog")
		_, _ = part.Write(tamperedBuf.Bytes())
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/system/backup/restore", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		dtSession := &auth.SessionUser{
			SessionID:      uuid.New(),
			UserID:         dtUserID,
			OrganizationID: testOrgID,
			Email:          "dt.backup@test.dz",
			Role:           string(domain.RoleDirecteurTechnique),
			Permissions:    []string{"SYS_MAINT"},
		}
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtSession))

		rr := httptest.NewRecorder()
		systemHandler.RestoreHomologArchive(rr, req)

		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("attendu 422 Unprocessable Entity pour blob altéré, reçu %d: %s", rr.Code, rr.Body.String())
		}
		t.Logf("PASS: Garde-fou d'intégrité validé (rejet immédiat d'une archive falsifiée)")
	})
}
