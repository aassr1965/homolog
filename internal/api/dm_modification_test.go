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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

func TestDMModification_Workflow(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://homolog:homolog_secret_2026@127.0.0.1:5434/homolog?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("PostgreSQL non disponible: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("PostgreSQL inaccessible: %v", err)
	}

	// 1. Setup organisation & user
	orgID := uuid.New()
	userID := uuid.New()
	orgSlug := "org-dm-mod-" + orgID.String()[:8]

	_, err = pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, created_at, updated_at)
		VALUES ($1, $2, 'Laboratoire Test DM Mod', NOW(), NOW())
	`, orgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation: %v", err)
	}
	defer pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", orgID)

	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, 'dt@mod.dz', 'hash', 'Yacine', 'Benkhalid', 'pharmacien_directeur_technique', TRUE, NOW(), NOW())
	`, userID, orgID)
	if err != nil {
		t.Fatalf("échec création user: %v", err)
	}

	// 2. Créer un DM de test
	devID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO medical_devices (
			id, organization_id, commercial_name_fr, manufacturer_reference, dm_class, origin_type,
			is_sterile, is_single_use, sterilization_method, status, created_at, updated_at
		) VALUES (
			$1, $2, 'Cathéter Test Modification', 'IVC-MOD-TEST', 'IIa', 'fabrication_locale',
			TRUE, TRUE, 'Oxyde d''éthylène (ETO)', 'en_cours', NOW(), NOW()
		)
	`, devID, orgID)
	if err != nil {
		t.Fatalf("échec création DM: %v", err)
	}

	// Handlers
	auditSvc := audit.NewService(pool)
	modRepo := repository.NewDMModificationRepository(pool)
	dmRepo := repository.NewDMRepository(pool)
	handler := NewDMModificationHandler(modRepo, dmRepo, auditSvc)

	mockUser := &auth.SessionUser{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          "dt@mod.dz",
		Role:           "pharmacien_directeur_technique",
	}

	// Router setup
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), UserContextKey, mockUser)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	r.Get("/api/v1/dm-modifications", handler.ListModifications)
	r.Get("/api/v1/dm-modifications/stats", handler.GetStats)
	r.Get("/api/v1/dm-modifications/categories", handler.GetCategories)
	r.Get("/api/v1/dm-modifications/{id}", handler.GetModification)
	r.Post("/api/v1/dm-modifications", handler.CreateModification)
	r.Put("/api/v1/dm-modifications/{id}", handler.UpdateModification)
	r.Post("/api/v1/dm-modifications/{id}/seal", handler.SealModification)
	r.Post("/api/v1/dm-modifications/{id}/avenant", handler.IssueAvenant)

	// A. Test GET Categories
	req := httptest.NewRequest("GET", "/api/v1/dm-modifications/categories", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("attendu 200 pour categories, reçu %d", w.Code)
	}

	// B. Test POST Create Modification (Substantielle)
	createPayload := map[string]any{
		"medical_device_id":             devID.String(),
		"modification_code":             "MOD-DM-TEST-001",
		"modification_type":             "SUBSTANTIELLE",
		"category_code":                 "DM_SUB_03",
		"category_label":                "Changement du procédé de stérilisation",
		"title":                         "Substitution ETO par Rayonnement Gamma",
		"description":                   "Passage à la stérilisation par irradiation Gamma",
		"justification":                 "Meilleure innocuité environnementale",
		"current_state":                 "Stérilisation Oxyde d'Éthylène",
		"proposed_state":                "Stérilisation Rayonnement Gamma 25 kGy",
		"impact_sterilization_iso11135": true,
		"impact_risk_iso14971":          true,
	}
	body, _ := json.Marshal(createPayload)
	req = httptest.NewRequest("POST", "/api/v1/dm-modifications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("attendu 201 pour création modification, reçu %d: %s", w.Code, w.Body.String())
	}

	var createdResp struct {
		Modification domain.DMModification `json:"modification"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createdResp)
	modID := createdResp.Modification.ID
	if modID == uuid.Nil {
		t.Fatal("ID de modification vide")
	}
	if createdResp.Modification.FeeAmount != domain.DMFeeSubstantielle {
		t.Fatalf("attendu redevance %.2f, reçu %.2f", domain.DMFeeSubstantielle, createdResp.Modification.FeeAmount)
	}

	// C. Test GET List Modifications
	req = httptest.NewRequest("GET", "/api/v1/dm-modifications", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("attendu 200 pour liste, reçu %d", w.Code)
	}

	// D. Test POST Seal Modification (PDT)
	sealPayload := map[string]any{
		"solemn_declaration": true,
		"decision_miph":      "DEC-MIPH-PDT-2022-7714",
		"notes":              "Visa BPF conforme",
	}
	sealBody, _ := json.Marshal(sealPayload)
	req = httptest.NewRequest("POST", "/api/v1/dm-modifications/"+modID.String()+"/seal", bytes.NewReader(sealBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("attendu 200 pour scellement, reçu %d: %s", w.Code, w.Body.String())
	}

	// E. Test POST Issue Avenant
	avnPayload := map[string]any{
		"avenant_number": "AVN-DDM-2026-TEST",
		"decision_notes": "Avenant favorable accordé",
	}
	avnBody, _ := json.Marshal(avnPayload)
	req = httptest.NewRequest("POST", "/api/v1/dm-modifications/"+modID.String()+"/avenant", bytes.NewReader(avnBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("attendu 200 pour délivrance avenant, reçu %d: %s", w.Code, w.Body.String())
	}

	// F. Test GET Stats
	req = httptest.NewRequest("GET", "/api/v1/dm-modifications/stats", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("attendu 200 pour stats, reçu %d", w.Code)
	}
	var statsResp struct {
		Stats domain.DMModificationStats `json:"stats"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &statsResp)
	if statsResp.Stats.TotalCount < 1 {
		t.Errorf("attendu au moins 1 modification dans stats, reçu %d", statsResp.Stats.TotalCount)
	}
	if statsResp.Stats.ApprouveAvenantCount < 1 {
		t.Errorf("attendu au moins 1 avenant approuvé, reçu %d", statsResp.Stats.ApprouveAvenantCount)
	}
}
