package api

import (
	"bytes"
	"context"
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

func TestFeeAPI_Workflow(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-fees-" + testOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM anpp_payment_slips WHERE organization_id = $1;
			DELETE FROM dossier_checklist_items WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id = $1);
			DELETE FROM dossiers WHERE organization_id = $1;
			DELETE FROM ged_documents WHERE organization_id = $1;
			DELETE FROM medicines WHERE organization_id = $1;
			DELETE FROM products WHERE organization_id = $1;
			DELETE FROM users WHERE organization_id = $1;
			DELETE FROM organizations WHERE id = $1;
		`, testOrgID)
	}()

	// 1. Créer l'organisation de test
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Laboratoire Test Redevances ANPP', 'MIPH-2026-FEES', 'Dr. Amine Khodja')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation test: %v", err)
	}

	// 2. Créer un Directeur Technique (PDT) et un utilisateur RA
	dtUserID := uuid.New()
	pwdHash, _ := auth.HashPassword("TestPass123!#", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, 'dt-fees@test.dz', $3, 'Amine', 'K.', 'directeur_technique', TRUE)
	`, dtUserID, testOrgID, pwdHash)
	if err != nil {
		t.Fatalf("échec création utilisateur DT: %v", err)
	}

	// 3. Créer un produit et un dossier médicament générique local (Procédure NF)
	productID := uuid.New()
	medID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, manufacturer_name, marketing_status)
		VALUES ($1, $2, 'medicament', 'CiproSanad 500mg', 'Laboratoires Sanad', 'en_preparation')
	`, productID, testOrgID)
	if err != nil {
		t.Fatalf("échec création produit: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO medicines (
			id, organization_id, product_id, origin_type, brand_name_fr,
			dci_fr, dosage_fr, pharmaceutical_form_fr, administration_route_fr,
			economic_file_type, fob_price, fob_currency, status
		) VALUES (
			$1, $2, $3, 'generique_local', 'CiproSanad 500mg',
			'Ciprofloxacine', '500mg', 'Comprimé', 'Orale',
			'pcsu_local', 420.00, 'DZD', 'en_constitution'
		)
	`, medID, testOrgID, productID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	dossierID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, medicine_id, product_type, dossier_code, procedure_code, procedure_type, internal_reference, status)
		VALUES ($1, $2, $3, 'medicine', 'DOS-2026-CIPRO-01', 'NF', 'enregistrement_national', 'REF-INT-CIPRO', 'en_constitution')
	`, dossierID, testOrgID, medID)
	if err != nil {
		t.Fatalf("échec création dossier: %v", err)
	}

	// 4. Instancier le routeur et handler
	auditService := audit.NewService(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	feeRepo := repository.NewFeeRepository(db.Pool)
	handler := NewFeeHandler(feeRepo, dossierRepo, auditService, db.Pool)

	r := chi.NewRouter()
	r.Get("/api/v1/fees/schedules", handler.ListFeeSchedules)
	r.Post("/api/v1/fees/schedules", handler.CreateOrUpdateFeeSchedule)
	r.Get("/api/v1/dossiers/{id}/fees", handler.ListDossierPaymentSlips)
	r.Post("/api/v1/dossiers/{id}/fees/generate-slip", handler.GenerateDossierPaymentSlip)
	r.Get("/api/v1/fees/slips/{slipId}", handler.GetPaymentSlip)
	r.Get("/api/v1/fees/slips/{slipId}/print-html", handler.GetPaymentSlipPrintHTML)
	r.Put("/api/v1/fees/slips/{slipId}/status", handler.UpdatePaymentSlipStatus)

	dtUser := &auth.SessionUser{
		UserID:         dtUserID,
		OrganizationID: testOrgID,
		Email:          "dt-fees@test.dz",
		FirstName:      "Amine",
		LastName:       "K.",
		Role:           "directeur_technique",
		Permissions:    []string{string(domain.PermDossierView), string(domain.PermDossierCreate), string(domain.PermUserManage), string(domain.PermRefEdit)},
	}

	var createdSlipID string

	// Test 1: Liste des barèmes réglementaires injectés par la migration 011
	t.Run("GET /api/v1/fees/schedules", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/fees/schedules?active=true", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200 OK, obtenu %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		json.NewDecoder(w.Body).Decode(&resp)
		schedules := resp["schedules"].([]any)
		if len(schedules) < 5 {
			t.Errorf("attendu au moins 5 barèmes actifs, obtenu %d", len(schedules))
		}
	})

	// Test 2: Génération automatique d'un bordereau BV1 pour le dossier générique local
	t.Run("POST /api/v1/dossiers/{id}/fees/generate-slip [BV1 Générique Local 150 000 DZD]", func(t *testing.T) {
		reqBody := domain.GeneratePaymentSlipRequest{
			SlipType:       domain.SlipTypeBV1,
			PaymentPurpose: "Redevance officielle de dépôt initial ANPP - CiproSanad 500mg",
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/fees/generate-slip", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("attendu code 201 Created, obtenu %d: %s", w.Code, w.Body.String())
		}

		var slip domain.ANPPPaymentSlip
		if err := json.NewDecoder(w.Body).Decode(&slip); err != nil {
			t.Fatalf("échec décodage bordereau: %v", err)
		}

		if slip.SlipType != "BV1" {
			t.Errorf("slip_type attendu 'BV1', obtenu '%s'", slip.SlipType)
		}
		if slip.AmountDZD != 150000.00 {
			t.Errorf("montant DZD attendu 150000.00, obtenu %.2f", slip.AmountDZD)
		}
		if !strings.Contains(strings.ToLower(slip.AmountWordsFR), "cent cinquante mille") {
			t.Errorf("montant en toutes lettres FR incorrect: %s", slip.AmountWordsFR)
		}
		if !strings.Contains(slip.AmountWordsAR, "دينار جزائري") {
			t.Errorf("montant en toutes lettres AR incorrect: %s", slip.AmountWordsAR)
		}

		createdSlipID = slip.ID
	})

	// Test 3: Récupération de la liste des bordereaux du dossier
	t.Run("GET /api/v1/dossiers/{id}/fees", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/fees", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200 OK, obtenu %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		json.NewDecoder(w.Body).Decode(&resp)
		slips := resp["slips"].([]any)
		if len(slips) != 1 {
			t.Errorf("attendu 1 bordereau pour le dossier, obtenu %d", len(slips))
		}
	})

	// Test 4: Génération de l'HTML imprimable officiel du bordereau BV1
	t.Run("GET /api/v1/fees/slips/{slipId}/print-html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/fees/slips/"+createdSlipID+"/print-html", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200 OK, obtenu %d: %s", w.Code, w.Body.String())
		}

		body := w.Body.String()
		if !strings.Contains(body, "RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE") {
			t.Errorf("en-tête république manquant dans le bordereau")
		}
		if !strings.Contains(body, "150000.00 DZD") {
			t.Errorf("montant numérique manquant dans le bordereau")
		}
		if !strings.Contains(body, "دينار جزائري") {
			t.Errorf("montant en arabe manquant dans le bordereau")
		}
	})

	// Test 5: Mise à jour du statut (marquage payé avec référence de quittance)
	t.Run("PUT /api/v1/fees/slips/{slipId}/status", func(t *testing.T) {
		reqBody := domain.UpdatePaymentSlipStatusRequest{
			Status:           domain.PaymentStatusPaye,
			PaymentReference: "QUIT-TRESOR-2026-9812",
			PaidAt:           time.Now().Format("2006-01-02"),
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/api/v1/fees/slips/"+createdSlipID+"/status", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200 OK, obtenu %d: %s", w.Code, w.Body.String())
		}

		var updated domain.ANPPPaymentSlip
		json.NewDecoder(w.Body).Decode(&updated)
		if updated.Status != domain.PaymentStatusPaye {
			t.Errorf("statut attendu 'paye', obtenu '%s'", updated.Status)
		}
		if updated.PaymentReference != "QUIT-TRESOR-2026-9812" {
			t.Errorf("référence quittance incorrecte: %s", updated.PaymentReference)
		}
	})

	// Test 6: Mise à jour dynamique d'un barème (ex: nouveau décret augmentant le tarif princeps à 400 000 DZD)
	t.Run("POST /api/v1/fees/schedules [Révision Tarifaire]", func(t *testing.T) {
		reqBody := domain.CreateFeeScheduleRequest{
			Category:       "medicament",
			ProcedureCode:  "princeps",
			LabelFR:        "Enregistrement Princeps (Nouveau Tarif Décret 2027)",
			LegalReference: "Arrêté Ministériel du 01 Janvier 2027 / JORADP N°01",
			AmountDZD:      400000.00,
			ValidFrom:      "2027-01-01",
			Notes:          "Révision quinquennale des redevances ANPP",
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/fees/schedules", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("attendu code 201 Created, obtenu %d: %s", w.Code, w.Body.String())
		}

		var schedule domain.ANPPFeeSchedule
		json.NewDecoder(w.Body).Decode(&schedule)
		if schedule.AmountDZD != 400000.00 {
			t.Errorf("nouveau montant attendu 400000.00, obtenu %.2f", schedule.AmountDZD)
		}
	})
}
