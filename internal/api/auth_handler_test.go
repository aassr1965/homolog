package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/database"
	"homolog/internal/domain"
)

func getTestDB(t *testing.T) *database.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://homolog:homolog_secret_2026@127.0.0.1:5434/homolog?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, dbURL)
	if err != nil {
		t.Skipf("PostgreSQL non disponible pour test d'intégration (%v), test ignoré", err)
		return nil
	}
	return db
}

func TestReauthSign_AuthenticationBypassPrevention(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	sm := auth.NewSessionManager(db.Pool, 60, 15)
	auditSvc := audit.NewService(db.Pool)
	handler := NewAuthHandler(db.Pool, sm, auditSvc)

	// Créer une organisation et un utilisateur DT de test SANS 2FA
	testOrgID := uuid.New()
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Labo Test Security', 'MIPH-TEST', 'Dr. Test')
		ON CONFLICT (id) DO NOTHING
	`, testOrgID, "org-test-"+testOrgID.String()[:8])
	if err != nil {
		t.Fatalf("Erreur création organisation test: %v", err)
	}

	testUserID := uuid.New()
	pwd := "P@ssw0rdPharma2026!"
	hash, err := auth.HashPassword(pwd, auth.DefaultArgon2Params)
	if err != nil {
		t.Fatalf("Erreur hachage mot de passe: %v", err)
	}

	testEmail := "dt-bypass-test-" + testUserID.String()[:8] + "@homolog.dz"
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, totp_secret, totp_enabled)
		VALUES ($1, $2, $3, $4, 'Amine', 'K.', $5, NULL, FALSE)
	`, testUserID, testOrgID, testEmail, hash, domain.RoleDirecteurTechnique)
	if err != nil {
		t.Fatalf("Erreur création utilisateur DT test: %v", err)
	}

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", testOrgID)
	}()

	sessionUser := &auth.SessionUser{
		SessionID:      uuid.New(),
		UserID:         testUserID,
		OrganizationID: testOrgID,
		Email:          testEmail,
		Role:           string(domain.RoleDirecteurTechnique),
		FirstName:      "Amine",
		LastName:       "K.",
		TOTPEnabled:    false,
	}

	sendReauth := func(reqBody ReauthSignRequest) *httptest.ResponseRecorder {
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reauth-sign", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctxWithUser := context.WithValue(req.Context(), UserContextKey, sessionUser)
		req = req.WithContext(ctxWithUser)

		rr := httptest.NewRecorder()
		handler.ReauthSign(rr, req)
		return rr
	}

	// 1. TENTATIVE D'EXPLOIT BYPASS : Mot de passe vide + code 2FA arbitraire alors que 2FA est inactif
	t.Run("Exploit Bypass: Password vide + TOTP sur compte sans 2FA doit être REJETÉ (401)", func(t *testing.T) {
		rec := sendReauth(ReauthSignRequest{
			Password:       "",
			TOTPCode:       "123456",
			DocumentName:   "Dossier-AMM-Test.pdf",
			DocumentSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			Reason:         "Test de sécurité",
			Declaration:    "Je tente de signer sans mot de passe",
		})

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("FAILLE DE SÉCURITÉ : La requête a reçu HTTP %d (attendu 401 Unauthorized). Réponse : %s", rec.Code, rec.Body.String())
		}

		var resp map[string]string
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		expectedErr := "Le second facteur 2FA n'est pas activé sur ce compte : la saisie du mot de passe est obligatoire"
		if resp["error"] != expectedErr {
			t.Errorf("Message d'erreur inattendu: attendu '%s', reçu '%s'", expectedErr, resp["error"])
		}
	})

	// 2. Requête vide : Ni mot de passe ni code 2FA
	t.Run("Requête vide rejetée (400)", func(t *testing.T) {
		rec := sendReauth(ReauthSignRequest{
			Password:     "",
			TOTPCode:     "",
			DocumentName: "Doc.pdf",
		})
		if rec.Code != http.StatusBadRequest {
			t.Errorf("Attendu 400 Bad Request, reçu %d", rec.Code)
		}
	})

	// 3. Faux mot de passe
	t.Run("Faux mot de passe rejeté (401)", func(t *testing.T) {
		rec := sendReauth(ReauthSignRequest{
			Password:       "MauvaisMotDePasse123!",
			DocumentName:   "Doc.pdf",
			DocumentSHA256: "abc",
		})
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Attendu 401 Unauthorized, reçu %d", rec.Code)
		}
	})

	// 4. Bon mot de passe : Signature autorisée (200)
	t.Run("Bon mot de passe autorisé (200)", func(t *testing.T) {
		rec := sendReauth(ReauthSignRequest{
			Password:       pwd,
			DocumentName:   "Doc-Valid.pdf",
			DocumentSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			Reason:         "Validation normale",
			Declaration:    "Certifié exact",
		})
		if rec.Code != http.StatusOK {
			t.Fatalf("Attendu 200 OK, reçu %d: %s", rec.Code, rec.Body.String())
		}

		var resp map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["status"] != "ok" {
			t.Errorf("Status inattendu: %v", resp["status"])
		}
	})
}

func TestReauthSign_With2FA(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	sm := auth.NewSessionManager(db.Pool, 60, 15)
	auditSvc := audit.NewService(db.Pool)
	handler := NewAuthHandler(db.Pool, sm, auditSvc)

	testOrgID := uuid.New()
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Labo Test 2FA', 'MIPH-2FA', 'Dr. 2FA')
		ON CONFLICT (id) DO NOTHING
	`, testOrgID, "org-2fa-"+testOrgID.String()[:8])
	if err != nil {
		t.Fatalf("Erreur création organisation test: %v", err)
	}

	testUserID := uuid.New()
	pwd := "P@ssw0rdPharma2026!"
	hash, err := auth.HashPassword(pwd, auth.DefaultArgon2Params)
	if err != nil {
		t.Fatalf("Erreur hachage mot de passe: %v", err)
	}

	testEmail := "dt-2fa-" + testUserID.String()[:8] + "@homolog.dz"
	totpSetup, err := auth.GenerateTOTPSetup(testEmail)
	if err != nil {
		t.Fatalf("Erreur génération 2FA: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, totp_secret, totp_enabled, recovery_codes)
		VALUES ($1, $2, $3, $4, 'Amine', 'K.', $5, $6, TRUE, $7)
	`, testUserID, testOrgID, testEmail, hash, domain.RoleDirecteurTechnique, totpSetup.Secret, totpSetup.HashedBackupCodes)
	if err != nil {
		t.Fatalf("Erreur création utilisateur DT 2FA: %v", err)
	}

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", testOrgID)
	}()

	sessionUser := &auth.SessionUser{
		SessionID:      uuid.New(),
		UserID:         testUserID,
		OrganizationID: testOrgID,
		Email:          testEmail,
		Role:           string(domain.RoleDirecteurTechnique),
		FirstName:      "Amine",
		LastName:       "K.",
		TOTPEnabled:    true,
	}

	sendReauth := func(reqBody ReauthSignRequest) *httptest.ResponseRecorder {
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reauth-sign", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		ctxWithUser := context.WithValue(req.Context(), UserContextKey, sessionUser)
		req = req.WithContext(ctxWithUser)

		rr := httptest.NewRecorder()
		handler.ReauthSign(rr, req)
		return rr
	}

	// 1. Faux code 2FA rejeté (401)
	t.Run("Faux code 2FA rejeté (401)", func(t *testing.T) {
		rec := sendReauth(ReauthSignRequest{
			Password:       "",
			TOTPCode:       "000000",
			DocumentName:   "Doc-2FA.pdf",
			DocumentSHA256: "abc",
		})
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("Attendu 401 Unauthorized, reçu %d", rec.Code)
		}
	})

	// 2. Bon code TOTP accepté (200)
	t.Run("Bon code TOTP accepté (200)", func(t *testing.T) {
		code, err := totp.GenerateCode(totpSetup.Secret, time.Now().UTC())
		if err != nil {
			t.Fatalf("Erreur génération code: %v", err)
		}

		rec := sendReauth(ReauthSignRequest{
			Password:       "",
			TOTPCode:       code,
			DocumentName:   "Doc-2FA.pdf",
			DocumentSHA256: "abc",
			Reason:         "Visa avec TOTP",
			Declaration:    "Signature conforme",
		})
		if rec.Code != http.StatusOK {
			t.Errorf("Attendu 200 OK, reçu %d: %s", rec.Code, rec.Body.String())
		}
	})

	// 3. Code de secours usage unique
	t.Run("Code de secours usage unique", func(t *testing.T) {
		backupCode := totpSetup.BackupCodes[0]

		rec1 := sendReauth(ReauthSignRequest{
			Password:       "",
			TOTPCode:       backupCode,
			DocumentName:   "Doc-2FA-Backup.pdf",
			DocumentSHA256: "abc",
			Reason:         "Visa secours",
			Declaration:    "Signature conforme",
		})
		if rec1.Code != http.StatusOK {
			t.Errorf("Première utilisation du code de secours: attendu 200, reçu %d: %s", rec1.Code, rec1.Body.String())
		}

		rec2 := sendReauth(ReauthSignRequest{
			Password:       "",
			TOTPCode:       backupCode,
			DocumentName:   "Doc-2FA-Backup2.pdf",
			DocumentSHA256: "abc",
			Reason:         "Visa secours rejeu",
			Declaration:    "Signature conforme",
		})
		if rec2.Code != http.StatusUnauthorized {
			t.Errorf("Rejeu du code de secours déjà consommé: attendu 401, reçu %d", rec2.Code)
		}
	})
}
