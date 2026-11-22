package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/database"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

func TestDMAPI_FullWorkflow(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://homolog:homolog_secret_2026@127.0.0.1:5434/homolog?sslmode=disable"
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, dbURL)
	if err != nil {
		t.Skipf("PostgreSQL non disponible, test d'intégration ignoré : %v", err)
		return
	}
	defer db.Close()

	testOrgID := uuid.New()
	orgSlug := "lab-dm-test-" + testOrgID.String()[:8]

	// Nettoyage après test
	defer func() {
		_, _ = db.Pool.Exec(ctx, `
			DELETE FROM dm_regularizations WHERE organization_id = $1;
			DELETE FROM dm_recevabilite_evaluations WHERE organization_id = $1;
			DELETE FROM dossier_checklist_items WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id = $1);
			DELETE FROM dossiers WHERE organization_id = $1;
			DELETE FROM medical_devices WHERE organization_id = $1;
			DELETE FROM products WHERE organization_id = $1;
			DELETE FROM users WHERE organization_id = $1;
			DELETE FROM organizations WHERE id = $1;
		`, testOrgID)
	}()

	// 1. Provisionner l'organisation et utilisateurs
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Laboratoire Test DM Homologation', 'MIPH-2026-DM', 'Dr. Yacine Benkhalid')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation: %v", err)
	}

	dtUserID := uuid.New()
	raUserID := uuid.New()
	pwdHash, _ := auth.HashPassword("TestPass123!#", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES 
			($1, $2, 'dt-dm@test.dz', $3, 'Yacine', 'Benkhalid', 'directeur_technique', TRUE),
			($4, $2, 'ra-dm@test.dz', $3, 'Nadia', 'L.', 'responsable_reglementaire', TRUE)
	`, dtUserID, testOrgID, pwdHash, raUserID)
	if err != nil {
		t.Fatalf("échec création utilisateurs: %v", err)
	}

	// 2. Créer un produit et dispositif médical
	prodID := uuid.New()
	deviceID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, manufacturer_name, dm_class, is_sterile, created_at, updated_at)
		VALUES ($1, $2, 'dispositif_medical', 'Cathéter IV 18G Test', 'B. Braun Melsungen AG', 'IIa', TRUE, NOW(), NOW())
	`, prodID, testOrgID)
	if err != nil {
		t.Fatalf("échec création produit: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO medical_devices (
			id, organization_id, product_id, commercial_name_fr, commercial_name_ar,
			manufacturer_reference, dm_class, classification_rule, is_sterile, sterilization_method,
			is_single_use, intended_use_fr, origin_type,
			composition_details, has_latex, has_phthalates, has_active_pharma, has_biological_origin, has_cmr_substances,
			risk_management_iso14971_ref, risk_management_status,
			biocompatibility_iso10993_ref, biocompatibility_status,
			clinical_eval_iso14155_ref, clinical_eval_status,
			sterilization_validation_ref, packaging_stability_iso11607_ref,
			sample_batch_pf, sample_batch_mp, sample_units_count,
			status, created_at, updated_at
		) VALUES (
			$1, $2, $3, 'Cathéter Sécurisé 18G', 'قسطرة وريدية معقمة',
			'IVC-18G-TEST', 'IIa', 'Règle 5', TRUE, 'Oxyde d''éthylène (ETO)',
			TRUE, 'Perfusion intraveineuse sécurisée', 'importe',
			'Polyuréthane biocompatible 85%%, Acier inoxydable 10%%', FALSE, FALSE, FALSE, FALSE, FALSE,
			'PGR-14971-TEST', 'conforme',
			'BIO-10993-TEST', 'conforme',
			'CER-14155-TEST', 'conforme',
			'VAL-STERIL-ETO', 'STAB-TYVEK-36M',
			'LOT-PF-TEST01', 'LOT-MP-TEST02', 3,
			'en_cours', NOW(), NOW()
		)
	`, deviceID, testOrgID, prodID)
	if err != nil {
		t.Fatalf("échec création medical_devices: %v", err)
	}

	// 3. Créer un dossier DM lié
	dossierID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (
			id, organization_id, medical_device_id, dossier_code, internal_reference, product_type, procedure_type,
			status, created_at, updated_at
		) VALUES (
			$1, $2, $3, 'DOS-DM-TEST-01', 'REF-DM-TEST-01', 'dispositif_medical', 'HOMOLOG_DM',
			'en_constitution', NOW(), NOW()
		)
	`, dossierID, testOrgID, deviceID)
	if err != nil {
		t.Fatalf("échec création dossier: %v", err)
	}

	// Initialisation des handlers et repositories
	auditSvc := audit.NewService(db.Pool)
	dmRepo := repository.NewDMRepository(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	dmHandler := NewDMHandler(dmRepo, dossierRepo, auditSvc)

	r := chi.NewRouter()
	r.Get("/api/v1/medical-devices/{id}/homologation", dmHandler.GetMedicalDeviceHomologation)
	r.Put("/api/v1/medical-devices/{id}/homologation", dmHandler.UpdateMedicalDeviceHomologation)
	r.Get("/api/v1/dossiers/{id}/dm-recevabilite", dmHandler.GetRecevabiliteEvaluation)
	r.Put("/api/v1/dossiers/{id}/dm-recevabilite", dmHandler.UpdateRecevabiliteEvaluation)
	r.Post("/api/v1/dossiers/{id}/dm-recevabilite/seal", dmHandler.SealRecevabiliteEvaluation)
	r.Get("/api/v1/dm-regularizations", dmHandler.ListDMRegularizations)
	r.Get("/api/v1/dm-regularizations/stats", dmHandler.GetDMRegularizationStats)
	r.Post("/api/v1/dm-regularizations", dmHandler.CreateDMRegularization)
	r.Put("/api/v1/dm-regularizations/{id}", dmHandler.UpdateDMRegularization)
	r.Get("/api/v1/dm-regularizations/export/canevas", dmHandler.ExportDMCanevas)

	withAuth := func(req *http.Request, userID uuid.UUID, email, role string) *http.Request {
		user := &auth.SessionUser{
			UserID:         userID,
			OrganizationID: testOrgID,
			Email:          email,
			Role:           role,
		}
		return req.WithContext(context.WithValue(req.Context(), UserContextKey, user))
	}

	// TEST 1 : Consultation et mise à jour de la fiche 6 sections DM
	t.Run("Homologation 6 sections DM - Get & Put", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/medical-devices/"+deviceID.String()+"/homologation", nil)
		req = withAuth(req, raUserID, "ra-dm@test.dz", "responsable_reglementaire")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Statut attendu 200, reçu: %d (corps: %s)", rec.Code, rec.Body.String())
		}

		var resp struct {
			Success       bool                 `json:"success"`
			MedicalDevice domain.MedicalDevice `json:"medical_device"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Parsing JSON impossible: %v", err)
		}
		if resp.MedicalDevice.CommercialNameFR != "Cathéter Sécurisé 18G" {
			t.Errorf("CommercialNameFR = %s, attendu 'Cathéter Sécurisé 18G'", resp.MedicalDevice.CommercialNameFR)
		}
		if resp.MedicalDevice.RiskManagementISO14971Ref != "PGR-14971-TEST" {
			t.Errorf("ISO 14971 Ref = %s, attendu 'PGR-14971-TEST'", resp.MedicalDevice.RiskManagementISO14971Ref)
		}

		// Mise à jour de la composition
		updatePayload := resp.MedicalDevice
		updatePayload.CompositionDetails = "PUR médical 90%, Acier 304L 10% (Formulation optimisée)"
		updatePayload.SampleUnitsCount = 5

		bodyBytes, _ := json.Marshal(updatePayload)
		putReq := httptest.NewRequest("PUT", "/api/v1/medical-devices/"+deviceID.String()+"/homologation", bytes.NewReader(bodyBytes))
		putReq = withAuth(putReq, raUserID, "ra-dm@test.dz", "responsable_reglementaire")
		putRec := httptest.NewRecorder()
		r.ServeHTTP(putRec, putReq)

		if putRec.Code != http.StatusOK {
			t.Fatalf("Mise à jour attendue 200, reçu: %d", putRec.Code)
		}
	})

	// TEST 2 : Évaluation en 13 points de recevabilité (ES-REG-03-DM)
	t.Run("Recevabilite 13 points - Get, Put & Seal PDT", func(t *testing.T) {
		// Récupération initiale (auto-générée par défaut)
		req := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/dm-recevabilite", nil)
		req = withAuth(req, raUserID, "ra-dm@test.dz", "responsable_reglementaire")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Statut attendu 200 pour recevabilité, reçu: %d (corps: %s)", rec.Code, rec.Body.String())
		}

		var resp struct {
			Success    bool                            `json:"success"`
			Evaluation domain.DMRecevabiliteEvaluation `json:"evaluation"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Evaluation.ConformityScore != 13 {
			t.Errorf("Score attendu 13, reçu: %d", resp.Evaluation.ConformityScore)
		}
		if resp.Evaluation.Decision != domain.RecevabiliteDecisionComplet {
			t.Errorf("Décision attendue 'complet_et_recevable', reçu: %s", resp.Evaluation.Decision)
		}

		// Scellement officiel 21 CFR Part 11 par le Pharmacien DT
		sealBody := []byte(`{"solemn_declaration": true, "decision_miph": "DEC-MIPH-2022-7714"}`)
		sealReq := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/dm-recevabilite/seal", bytes.NewReader(sealBody))
		sealReq = withAuth(sealReq, dtUserID, "dt-dm@test.dz", "directeur_technique")
		sealRec := httptest.NewRecorder()
		r.ServeHTTP(sealRec, sealReq)

		if sealRec.Code != http.StatusOK {
			t.Fatalf("Scellement attendu 200, reçu: %d (corps: %s)", sealRec.Code, sealRec.Body.String())
		}

		var sealResp struct {
			Success bool   `json:"success"`
			Seal    string `json:"seal"`
		}
		_ = json.Unmarshal(sealRec.Body.Bytes(), &sealResp)
		if !strings.HasPrefix(sealResp.Seal, "DM-SEAL-REC-") {
			t.Errorf("Sceau généré invalide: %s", sealResp.Seal)
		}
	})

	// TEST 3 : Mesures Transitoires, Régularisation & Export Canevas
	t.Run("Régularisation DM - Create, List, Stats & Export", func(t *testing.T) {
		regCode := "REG-DM-TEST-" + testOrgID.String()[:8]
		// Création d'un dossier de régularisation transitoire
		reg := domain.DMRegularization{
			MedicalDeviceID:            deviceID,
			RegularizationCode:         regCode,
			DerogationNumber:           "DEROG-2024-TEST",
			Status:                     domain.DMRegStatusCommercialiseDerogatoire,
			QuarterlyVolumeImported:    15000,
			QuarterlyVolumeDistributed: 14200,
			HospitalDistributionShare:  70,
			RetailDistributionShare:    30,
			Observations:               "DM en régularisation d'urgence",
		}
		bodyBytes, _ := json.Marshal(reg)
		postReq := httptest.NewRequest("POST", "/api/v1/dm-regularizations", bytes.NewReader(bodyBytes))
		postReq = withAuth(postReq, raUserID, "ra-dm@test.dz", "responsable_reglementaire")
		postRec := httptest.NewRecorder()
		r.ServeHTTP(postRec, postReq)

		if postRec.Code != http.StatusCreated {
			t.Fatalf("Création régularisation attendue 201, reçu: %d (corps: %s)", postRec.Code, postRec.Body.String())
		}

		// Liste des régularisations
		listReq := httptest.NewRequest("GET", "/api/v1/dm-regularizations", nil)
		listReq = withAuth(listReq, raUserID, "ra-dm@test.dz", "responsable_reglementaire")
		listRec := httptest.NewRecorder()
		r.ServeHTTP(listRec, listReq)

		if listRec.Code != http.StatusOK {
			t.Fatalf("Liste attendue 200, reçu: %d", listRec.Code)
		}
		var listResp struct {
			Count           int                       `json:"count"`
			Regularizations []domain.DMRegularization `json:"regularizations"`
		}
		_ = json.Unmarshal(listRec.Body.Bytes(), &listResp)
		if listResp.Count < 1 {
			t.Errorf("Attendu au moins 1 régularisation, reçu: %d", listResp.Count)
		}

		// Statistiques
		statsReq := httptest.NewRequest("GET", "/api/v1/dm-regularizations/stats", nil)
		statsReq = withAuth(statsReq, raUserID, "ra-dm@test.dz", "responsable_reglementaire")
		statsRec := httptest.NewRecorder()
		r.ServeHTTP(statsRec, statsReq)

		if statsRec.Code != http.StatusOK {
			t.Fatalf("Stats attendues 200, reçu: %d", statsRec.Code)
		}
		var statsResp struct {
			Stats domain.DMRegularizationStats `json:"stats"`
		}
		_ = json.Unmarshal(statsRec.Body.Bytes(), &statsResp)
		if statsResp.Stats.TotalDevicesCount < 1 {
			t.Errorf("TotalDevicesCount attendu >= 1, reçu: %d", statsResp.Stats.TotalDevicesCount)
		}

		// Export Canevas officiel
		expReq := httptest.NewRequest("GET", "/api/v1/dm-regularizations/export/canevas", nil)
		expReq = withAuth(expReq, raUserID, "ra-dm@test.dz", "responsable_reglementaire")
		expRec := httptest.NewRecorder()
		r.ServeHTTP(expRec, expReq)

		if expRec.Code != http.StatusOK {
			t.Fatalf("Export attendu 200, reçu: %d", expRec.Code)
		}
		csvContent := expRec.Body.String()
		if !strings.Contains(csvContent, "Code Suivi Régularisation") {
			t.Errorf("En-tête manquant dans l'export: %s", csvContent)
		}
		if !strings.Contains(csvContent, regCode) {
			t.Errorf("Donnée %s manquante dans le CSV: %s", regCode, csvContent)
		}
	})
}
