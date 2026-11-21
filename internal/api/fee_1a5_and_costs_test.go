package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

func TestFeeAPI_Volet1A5_PreSubmissionAndCostConsolidation(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-1a5-" + testOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM anpp_pre_submissions WHERE organization_id = $1;
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
		VALUES ($1, $2, 'Laboratoire Test 1A5 ANPP', 'MIPH-2026-1A5', 'Dr. Farida Benali')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation: %v", err)
	}

	// 2. Créer un DT, un RA et un R&D
	dtUserID := uuid.New()
	raUserID := uuid.New()
	rdUserID := uuid.New()
	pwdHash, _ := auth.HashPassword("TestPass123!#", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES 
			($1, $2, 'dt-1a5@test.dz', $3, 'Farida', 'B.', 'directeur_technique', TRUE),
			($4, $2, 'ra-1a5@test.dz', $3, 'Nassim', 'L.', 'responsable_reglementaire', TRUE),
			($5, $2, 'rd-1a5@test.dz', $3, 'Karim', 'S.', 'contributeur_rd_production', TRUE)
	`, dtUserID, testOrgID, pwdHash, raUserID, rdUserID)
	if err != nil {
		t.Fatalf("échec création utilisateurs: %v", err)
	}

	// 3. Créer un produit et deux dossiers : un en NI (soumis à 1.A.5) et un en RI (renouvellement non soumis)
	productID := uuid.New()
	medID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, manufacturer_name, marketing_status)
		VALUES ($1, $2, 'medicament', 'CardioGuard 50mg', 'Laboratoires El Djazair', 'en_preparation')
	`, productID, testOrgID)
	if err != nil {
		t.Fatalf("échec création produit: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO medicines (
			id, organization_id, product_id, origin_type, brand_name_fr,
			dci_fr, dosage_fr, pharmaceutical_form_fr, administration_route_fr,
			economic_file_type, pcsu_amount, status
		) VALUES (
			$1, $2, $3, 'generique_local', 'CardioGuard 50mg',
			'Losartan', '50mg', 'Comprimé pelliculé', 'Orale',
			'pcsu', 450.00, 'en_constitution'
		)
	`, medID, testOrgID, productID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	dossierNIID := uuid.New()
	dossierRIID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, medicine_id, product_type, dossier_code, procedure_code, procedure_type, internal_reference, status)
		VALUES 
			($1, $2, $3, 'medicine', 'DOS-2026-NI-01', 'NI', 'enregistrement_national', 'REF-NI-CARDIO', 'en_constitution'),
			($4, $2, $3, 'medicine', 'DOS-2026-RI-01', 'RI', 'renouvellement', 'REF-RI-CARDIO', 'en_constitution')
	`, dossierNIID, testOrgID, medID, dossierRIID)
	if err != nil {
		t.Fatalf("échec création dossiers: %v", err)
	}

	// Checklist 1.A.5 pour dossier NI
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossier_checklist_items (
			id, dossier_id, module_number, sub_module, item_code,
			title_fr, is_mandatory, status
		) VALUES (
			gen_random_uuid(), $1, 1, '1.A', '1.A.5',
			'Frais de pré-soumission ANPP', TRUE, 'manquant'
		)
	`, dossierNIID)
	if err != nil {
		t.Fatalf("échec création item checklist 1.A.5: %v", err)
	}

	// Créer deux documents GED (accord préalable et quittance redevance pré-soumission)
	accordDocID := uuid.New()
	quittanceDocID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO ged_documents (id, organization_id, taxonomy_path, document_code, title_fr, status)
		VALUES 
			($1, $2, '01_Affaires_Reglementaires/1.A', 'GED-ACCORD-01', 'Lettre Accord Préalable ANPP 2026', 'valide'),
			($3, $2, '01_Affaires_Reglementaires/1.A', 'GED-QUIT-1A5', 'Quittance Frais Pré-soumission Trésor', 'valide')
	`, accordDocID, testOrgID, quittanceDocID)
	if err != nil {
		t.Fatalf("échec création docs GED: %v", err)
	}

	// Créer un bordereau BV1 pour tester la consolidation des coûts
	var scheduleID uuid.UUID
	err = db.Pool.QueryRow(ctx, "SELECT id FROM anpp_fee_schedules WHERE procedure_code = 'generique_local' LIMIT 1").Scan(&scheduleID)
	if err != nil {
		t.Fatalf("échec récupération barème fee_schedules: %v", err)
	}

	slipBV1ID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO anpp_payment_slips (
			id, organization_id, dossier_id, fee_schedule_id, slip_type,
			slip_number, payment_purpose, amount_dzd, amount_words_fr, amount_words_ar, status
		) VALUES (
			$1, $2, $3, $4, 'BV1',
			'BV1-2026-CARDIO-01', 'Dépôt légal dossier enregistrement générique local',
			300000.00, 'Trois cent mille dinars algériens', 'ثلاثمائة ألف دينار جزائري', 'paye'
		)
	`, slipBV1ID, testOrgID, dossierNIID, scheduleID)
	if err != nil {
		t.Fatalf("échec création bordereau BV1: %v", err)
	}

	// Setup Handlers
	auditService := audit.NewService(db.Pool)
	feeRepo := repository.NewFeeRepository(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	handler := NewFeeHandler(feeRepo, dossierRepo, auditService, db.Pool)

	r := chi.NewRouter()
	r.Get("/api/v1/dossiers/{id}/fees/1a5-summary", handler.GetSection1A5Summary)
	r.Post("/api/v1/dossiers/{id}/fees/verify-1a5", handler.VerifyPreSubmission1A5)
	r.Get("/api/v1/fees/analytics/dashboard", handler.GetCostConsolidationDashboard)
	r.Get("/api/v1/fees/analytics/export", handler.ExportCostConsolidationCSV)

	dtUser := &auth.SessionUser{
		UserID:         dtUserID,
		OrganizationID: testOrgID,
		Email:          "dt-1a5@test.dz",
		FirstName:      "Farida",
		LastName:       "B.",
		Role:           "directeur_technique",
		Permissions:    []string{string(domain.PermDossierView), string(domain.PermDossierCreate), string(domain.PermDocApprovePDT), string(domain.PermDocCheckReview)},
	}

	t.Run("1. Consultation 1.A.5 pour dossier NI sans accord enregistré", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dossiers/"+dossierNIID.String()+"/fees/1a5-summary", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200, reçu %d: %s", w.Code, w.Body.String())
		}

		var summary domain.Section1A5Summary
		if err := json.NewDecoder(w.Body).Decode(&summary); err != nil {
			t.Fatalf("échec unmarshal json: %v", err)
		}

		if !summary.IsPreSubmissionRequired {
			t.Errorf("attendu IsPreSubmissionRequired=true pour procédure NI")
		}
		if summary.Section1A5Status != "manquant" {
			t.Errorf("statut inattendu: %s", summary.Section1A5Status)
		}
		if summary.OverallCompliant {
			t.Errorf("OverallCompliant devrait être faux avant enregistrement")
		}
	})

	t.Run("2. Consultation 1.A.5 pour dossier RI (Renouvellement)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dossiers/"+dossierRIID.String()+"/fees/1a5-summary", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200, reçu %d: %s", w.Code, w.Body.String())
		}

		var summary domain.Section1A5Summary
		if err := json.NewDecoder(w.Body).Decode(&summary); err != nil {
			t.Fatalf("échec unmarshal json: %v", err)
		}

		if summary.IsPreSubmissionRequired {
			t.Errorf("attendu IsPreSubmissionRequired=false pour procédure RI")
		}
		if summary.Section1A5Status != "non_applicable" {
			t.Errorf("attendu statut 'non_applicable', reçu %s", summary.Section1A5Status)
		}
		if !summary.OverallCompliant {
			t.Errorf("OverallCompliant devrait être vrai car non applicable")
		}
	})

	t.Run("3. Enregistrement & Validation 1.A.5 par DT", func(t *testing.T) {
		accordDocIDStr := accordDocID.String()
		quittanceDocIDStr := quittanceDocID.String()
		payload := domain.VerifyPreSubmission1A5Request{
			ANPPReference:        "ANPP/DREC/2026/045",
			ANPPNotificationDate: "2026-03-15",
			DecisionType:         "favorable",
			FeeAmountDZD:         50000.00,
			SlipReference:        "QUITT-ANPP-2026-089",
			PaymentDate:          "2026-03-12",
			ResponseDocumentID:   &accordDocIDStr,
			PaymentDocumentID:    &quittanceDocIDStr,
			VerificationNotes:    "Contrôle de conformité de l'accord 1.A.5 certifié par Pharmacien Directeur Technique.",
		}

		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/dossiers/"+dossierNIID.String()+"/fees/verify-1a5", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200, reçu %d: %s", w.Code, w.Body.String())
		}

		// Re-consulter 1.A.5 summary
		req2 := httptest.NewRequest(http.MethodGet, "/api/v1/dossiers/"+dossierNIID.String()+"/fees/1a5-summary", nil)
		req2 = req2.WithContext(context.WithValue(req2.Context(), UserContextKey, dtUser))
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Fatalf("attendu code 200, reçu %d: %s", w2.Code, w2.Body.String())
		}

		var updatedSummary domain.Section1A5Summary
		if err := json.NewDecoder(w2.Body).Decode(&updatedSummary); err != nil {
			t.Fatalf("échec unmarshal json: %v", err)
		}

		if updatedSummary.Section1A5Status != "valide_qualite" {
			t.Errorf("attendu statut 'valide_qualite', reçu %s", updatedSummary.Section1A5Status)
		}
		if !updatedSummary.OverallCompliant {
			t.Errorf("attendu OverallCompliant=true")
		}
		if updatedSummary.PreSubmission == nil || updatedSummary.PreSubmission.FeeAmountDZD != 50000.00 {
			t.Errorf("montant redevance non conforme: %+v", updatedSummary.PreSubmission)
		}

		// Vérifier que l'item de checklist 1.A.5 a été mis à jour dans la base
		var checklistStatus string
		err = db.Pool.QueryRow(ctx, `
			SELECT status FROM dossier_checklist_items WHERE dossier_id = $1 AND item_code = '1.A.5'
		`, dossierNIID).Scan(&checklistStatus)
		if err != nil {
			t.Fatalf("erreur lecture checklist: %v", err)
		}
		if checklistStatus != "valide_qualite" {
			t.Errorf("attendu item checklist 'valide_qualite', reçu %s", checklistStatus)
		}
	})

	t.Run("4. Consultation Tableau de Bord Consolidé des Coûts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/fees/analytics/dashboard", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200, reçu %d: %s", w.Code, w.Body.String())
		}

		var dash domain.CostConsolidationSummary
		if err := json.NewDecoder(w.Body).Decode(&dash); err != nil {
			t.Fatalf("échec unmarshal dashboard: %v", err)
		}

		// Total émis : 50,000 (1.A.5) + 300,000 (BV1) = 350,000 DZD
		if dash.TotalFeesEmittedDZD < 350000.00 {
			t.Errorf("total émis attendu >= 350000, reçu %f", dash.TotalFeesEmittedDZD)
		}
		if dash.TotalFeesPaidDZD < 350000.00 {
			t.Errorf("total payé attendu >= 350000, reçu %f", dash.TotalFeesPaidDZD)
		}
		if len(dash.ProductBreakdown) == 0 {
			t.Errorf("attendu au moins 1 produit dans ProductBreakdown")
		}
		if len(dash.YearlyBreakdown) == 0 {
			t.Errorf("attendu au moins 1 année dans YearlyBreakdown")
		}
	})

	t.Run("5. Export CSV Comptabilité & DAF", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/fees/analytics/export", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200, reçu %d: %s", w.Code, w.Body.String())
		}

		csvContent := w.Body.String()
		if !strings.Contains(csvContent, "Produit;Type;Dossier;Procédure") {
			t.Errorf("entête CSV manquante ou invalide: %s", csvContent)
		}
		if !strings.Contains(csvContent, "DOS-2026-NI-01") {
			t.Errorf("dossier DOS-2026-NI-01 manquant dans l'export CSV")
		}
		if !strings.Contains(csvContent, "CardioGuard 50mg") {
			t.Errorf("produit CardioGuard 50mg manquant dans l'export CSV")
		}
	})
}
