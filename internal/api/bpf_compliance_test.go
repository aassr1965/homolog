package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// TestBPF_ALCOA_AuditLogImmutability valide l'inaltérabilité stricte de l'audit trail (ALCOA+ Enduring)
func TestBPF_ALCOA_AuditLogImmutability(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	auditSvc := audit.NewService(db.Pool)

	orgID := uuid.New()
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Labo Audit Immuable', 'MIPH-IMMUTABLE', 'Dr. Immuable')
		ON CONFLICT (id) DO NOTHING
	`, orgID, "org-audit-"+uuid.New().String()[:8])

	// 1. Insertion d'une trace d'audit
	entry, err := auditSvc.Log(ctx, &audit.Entry{
		OrganizationID: orgID,
		Action:         "TEST_BPF_IMMUTABILITY",
		EntityType:     "compliance",
		EntityID:       "ALCOA-01",
		Details:        map[string]any{"check": "tamper_protection"},
	})
	if err != nil {
		t.Fatalf("Échec création trace audit: %v", err)
	}

	// 2. Tentative de falsification (UPDATE) sur l'Audit Trail -> Doit lever une exception SQL BPF
	_, err = db.Pool.Exec(ctx, "UPDATE audit_logs SET action = 'FALSIFIED_ACTION' WHERE id = $1", entry.ID)
	if err == nil {
		t.Fatal("CONFORMITÉ BPF VIOLÉE : L'UPDATE sur audit_logs a réussi alors qu'il doit être strictement interdit !")
	}
	if !strings.Contains(err.Error(), "ALERTE CONFORMITÉ BPF") && !strings.Contains(err.Error(), "immuable") {
		t.Logf("Message de rejet trigger reçu : %v", err)
	}

	// 3. Tentative de suppression (DELETE) sur l'Audit Trail -> Doit également être rejetée
	_, err = db.Pool.Exec(ctx, "DELETE FROM audit_logs WHERE id = $1", entry.ID)
	if err == nil {
		t.Fatal("CONFORMITÉ BPF VIOLÉE : Le DELETE sur audit_logs a réussi alors qu'il doit être strictement interdit !")
	}
}

// TestBPF_ALCOA_SHA256CryptographicSeal valide le scellement cryptographique SHA-256 de chaque entrée d'audit
func TestBPF_ALCOA_SHA256CryptographicSeal(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	auditSvc := audit.NewService(db.Pool)

	orgID := uuid.New()
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Labo Test SHA256', 'MIPH-SHA', 'Dr. SHA')
		ON CONFLICT (id) DO NOTHING
	`, orgID, "org-sha-"+uuid.New().String()[:8])

	userID := uuid.New()
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, 'sha-test@pharma.dz', 'dummy-hash', 'Dr.', 'SHA', 'directeur_technique', true)
		ON CONFLICT (id) DO NOTHING
	`, userID, orgID)

	details := map[string]any{"lot": "L-2026-99", "conformite": "conforme"}

	entry, err := auditSvc.Log(ctx, &audit.Entry{
		OrganizationID: orgID,
		UserID:         &userID,
		UserEmail:      "pdt@pharma.dz",
		UserRole:       string(domain.RoleDirecteurTechnique),
		Action:         "LOT_RELEASE",
		EntityType:     "batch",
		EntityID:       "BATCH-2026-99",
		Details:        details,
		IPAddress:      "192.168.1.100",
		UserAgent:      "Homolog-BPF-Agent/1.0",
	})
	if err != nil {
		t.Fatalf("Échec création trace audit: %v", err)
	}

	if entry.SHA256Hash == "" || len(entry.SHA256Hash) != 64 {
		t.Fatalf("Scellement SHA-256 invalide : %s", entry.SHA256Hash)
	}

	// Recalcul indépendant du hash
	detailsBytes, _ := json.Marshal(details)
	hasher := sha256.New()
	hasher.Write([]byte(entry.CreatedAt.Format(time.RFC3339Nano)))
	hasher.Write([]byte(orgID.String()))
	hasher.Write([]byte(userID.String()))
	hasher.Write([]byte("LOT_RELEASE"))
	hasher.Write([]byte("batch"))
	hasher.Write([]byte("BATCH-2026-99"))
	hasher.Write(detailsBytes)
	hasher.Write([]byte("192.168.1.100"))
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	if entry.SHA256Hash != expectedHash {
		t.Fatalf("Empreinte SHA-256 non concordante !\nAttendu: %s\nObtenu:  %s", expectedHash, entry.SHA256Hash)
	}
}

// TestBPF_21CFRPart11_ElectronicSignatureEnforcement teste la signature électronique nominative sur l'archivage
func TestBPF_21CFRPart11_ElectronicSignatureEnforcement(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	sm := auth.NewSessionManager(db.Pool, 60, 15)
	auditSvc := audit.NewService(db.Pool)
	pharmaRepo := repository.NewPharmaRepository(db.Pool)
	handler := NewPharmaHandler(pharmaRepo, auditSvc)

	// Création organisation et utilisateur test avec mot de passe connu
	orgID := uuid.New()
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Labo Test Part11', 'MIPH-PART11', 'Dr. Part11')
		ON CONFLICT (id) DO NOTHING
	`, orgID, "org-p11-"+uuid.New().String()[:8])

	userID := uuid.New()
	password := "SecretP11!2026"
	hash, _ := auth.HashPassword(password, nil)
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, $3, $4, 'Dr.', 'PDT', 'directeur_technique', true)
		ON CONFLICT (id) DO NOTHING
	`, userID, orgID, "pdt-test-p11@pharma.dz", hash)

	// Créer un médicament de test
	medID := uuid.New()
	prodID := uuid.New()
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, manufacturer_name)
		VALUES ($1, $2, 'medicament', 'Spécialité Part11 Test', 'Labo Test')
	`, prodID, orgID)
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO medicines (id, organization_id, product_id, origin_type, brand_name_fr, dci_fr, dosage_fr, pharmaceutical_form_fr, administration_route_fr, status)
		VALUES ($1, $2, $3, 'generique_local', 'Spécialité Part11 Test', 'DCI Test', '500 mg', 'Comprimé', 'Voie orale', 'en_preparation')
	`, medID, orgID, prodID)

	sessionUser := &auth.SessionUser{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          "pdt-test-p11@pharma.dz",
		Role:           string(domain.RoleDirecteurTechnique),
		Permissions:    []string{string(domain.PermRefEdit), string(domain.PermRefView)},
	}

	// Helper pour ajouter le contexte Chi et la session
	buildReq := func(body string) *http.Request {
		req := httptest.NewRequest("POST", "/api/v1/medicines/"+medID.String()+"/archive", bytes.NewReader([]byte(body)))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", medID.String())
		ctxWithRoute := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		return req.WithContext(context.WithValue(ctxWithRoute, UserContextKey, sessionUser))
	}

	// 1. Test rejet si absence de mot de passe (signature non fournie)
	noPwdReq := buildReq(`{"reason": "Arrêt commercial"}`)
	noPwdRR := httptest.NewRecorder()
	handler.ArchiveMedicine(noPwdRR, noPwdReq)
	if noPwdRR.Code != http.StatusUnauthorized {
		t.Fatalf("Attendu HTTP 401 pour signature absente, obtenu: %d", noPwdRR.Code)
	}

	// 2. Test rejet si mot de passe erroné
	badPwdReq := buildReq(`{"reason": "Arrêt commercial", "password": "MauvaisMotDePasse"}`)
	badPwdRR := httptest.NewRecorder()
	handler.ArchiveMedicine(badPwdRR, badPwdReq)
	if badPwdRR.Code != http.StatusUnauthorized {
		t.Fatalf("Attendu HTTP 401 pour signature invalide, obtenu: %d", badPwdRR.Code)
	}

	// 3. Test rejet si justification réglementaire absente
	noReasonReq := buildReq(`{"reason": "   ", "password": "` + password + `"}`)
	noReasonRR := httptest.NewRecorder()
	handler.ArchiveMedicine(noReasonRR, noReasonReq)
	if noReasonRR.Code != http.StatusBadRequest {
		t.Fatalf("Attendu HTTP 400 pour motif absent, obtenu: %d", noReasonRR.Code)
	}

	// 4. Test succès avec signature électronique nominative valide
	goodReq := buildReq(`{"reason": "Arrêt définitif de commercialisation ordonné par le MIPH", "password": "` + password + `"}`)
	goodRR := httptest.NewRecorder()
	handler.ArchiveMedicine(goodRR, goodReq)
	if goodRR.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour signature valide, obtenu: %d (corps: %s)", goodRR.Code, goodRR.Body.String())
	}

	// 5. Vérification ALCOA+ : Le médicament n'est PAS détruit en base, il a le statut 'archive'
	var dbStatus string
	err := db.Pool.QueryRow(ctx, "SELECT status FROM medicines WHERE id = $1", medID).Scan(&dbStatus)
	if err != nil {
		t.Fatalf("Le médicament doit exister en base (interdiction du hard delete): %v", err)
	}
	if dbStatus != "archive" {
		t.Fatalf("Statut attendu 'archive', obtenu: '%s'", dbStatus)
	}

	_ = sm
}
