package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

func TestFeeAPI_Volet1A4_VerificationAndFX(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-1a4-" + testOrgID.String()[:8]

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
		VALUES ($1, $2, 'Laboratoire Test 1A4 ANPP', 'MIPH-2026-1A4', 'Dr. Amine Khodja')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation test: %v", err)
	}

	// 2. Créer un DT (PDT), un RA (Responsable Réglementaire) et un contributeur R&D (non habilité SoD)
	dtUserID := uuid.New()
	raUserID := uuid.New()
	rdUserID := uuid.New()
	pwdHash, _ := auth.HashPassword("TestPass123!#", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES 
			($1, $2, 'dt-1a4@test.dz', $3, 'Amine', 'K.', 'directeur_technique', TRUE),
			($4, $2, 'ra-1a4@test.dz', $3, 'Sarah', 'B.', 'responsable_reglementaire', TRUE),
			($5, $2, 'rd-1a4@test.dz', $3, 'Karim', 'M.', 'contributeur_rd_production', TRUE)
	`, dtUserID, testOrgID, pwdHash, raUserID, rdUserID)
	if err != nil {
		t.Fatalf("échec création utilisateurs: %v", err)
	}

	// 3. Créer un produit importé et un dossier médicament en procédure NI
	productID := uuid.New()
	medID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, manufacturer_name, marketing_status)
		VALUES ($1, $2, 'medicament', 'ImportaCilline 1g', 'Pharma Suisse SA', 'en_preparation')
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
			$1, $2, $3, 'generique_importe', 'ImportaCilline 1g',
			'Amoxicilline', '1g', 'Poudre pour injection', 'IV/IM',
			'fob_import', 4.50, 'EUR', 'en_constitution'
		)
	`, medID, testOrgID, productID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	dossierID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, medicine_id, product_type, dossier_code, procedure_code, procedure_type, internal_reference, status)
		VALUES ($1, $2, $3, 'medicine', 'DOS-2026-IMP-01', 'NI', 'enregistrement_national', 'REF-INT-IMP', 'en_constitution')
	`, dossierID, testOrgID, medID)
	if err != nil {
		t.Fatalf("échec création dossier: %v", err)
	}

	// Créer l'élément de checklist 1.A.4
	checklistItemID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossier_checklist_items (id, dossier_id, module_number, sub_module, item_code, title_fr, status)
		VALUES ($1, $2, 1, '1.A', '1.A.4', 'Bordereau de versement + document bancaire / quittance ou autres + Attestation de taux de change si applicable', 'manquant')
	`, checklistItemID, dossierID)
	if err != nil {
		t.Fatalf("échec création item checklist 1.A.4: %v", err)
	}

	// Créer deux documents dans la GED (Quittance bancaire et Attestation Banque d'Algérie)
	receiptDocID := uuid.New()
	fxDocID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO ged_documents (id, organization_id, taxonomy_path, document_code, title_fr, status)
		VALUES 
			($1, $2, '01_Affaires_Reglementaires/1.A', 'GED-QUIT-01', 'Avis d''opéré bancaire BNA - Redevance ANPP', 'valide'),
			($3, $2, '01_Affaires_Reglementaires/1.A', 'GED-FX-01', 'Attestation de cours officiel Banque d''Algérie', 'valide')
	`, receiptDocID, testOrgID, fxDocID)
	if err != nil {
		t.Fatalf("échec création documents GED: %v", err)
	}

	// 4. Instancier le routeur et handler
	auditService := audit.NewService(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	feeRepo := repository.NewFeeRepository(db.Pool)
	handler := NewFeeHandler(feeRepo, dossierRepo, auditService, db.Pool)

	r := chi.NewRouter()
	r.Get("/api/v1/dossiers/{id}/fees/1a4-summary", handler.GetSection1A4Summary)
	r.Post("/api/v1/dossiers/{id}/fees/generate-slip", handler.GenerateDossierPaymentSlip)
	r.Post("/api/v1/dossiers/{id}/fees/slips/{slipId}/verify-1a4", handler.VerifyPaymentSlip1A4)

	raUser := &auth.SessionUser{
		UserID:         raUserID,
		OrganizationID: testOrgID,
		Email:          "ra-1a4@test.dz",
		FirstName:      "Sarah",
		LastName:       "B.",
		Role:           "responsable_reglementaire",
		Permissions:    []string{string(domain.PermDossierView), string(domain.PermDossierCreate), string(domain.PermDocCheckReview)},
	}

	rdUser := &auth.SessionUser{
		UserID:         rdUserID,
		OrganizationID: testOrgID,
		Email:          "rd-1a4@test.dz",
		FirstName:      "Karim",
		LastName:       "M.",
		Role:           "contributeur_rd_production",
		Permissions:    []string{string(domain.PermDossierView)},
	}

	// TEST 1 : Vérifier la synthèse 1.A.4 initiale (Bordereau manquant)
	t.Run("01_Initial_1A4_Summary_Missing_Slip", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dossiers/"+dossierID.String()+"/fees/1a4-summary", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, raUser))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu 200, reçu %d: %s", w.Code, w.Body.String())
		}

		var summary domain.Section1A4Summary
		if err := json.NewDecoder(w.Body).Decode(&summary); err != nil {
			t.Fatalf("décodage json synthèse échoué: %v", err)
		}

		if summary.OverallCompliant {
			t.Errorf("summary.OverallCompliant devrait être false au début")
		}
		if len(summary.BlockingIssues) == 0 {
			t.Errorf("devrait avoir des points bloquants signalés")
		}
	})

	// TEST 2 : Générer le bordereau BV1 pour ce dossier
	var generatedSlipID string
	t.Run("02_Generate_Payment_Slip_BV1", func(t *testing.T) {
		payload := map[string]any{
			"slip_type":       "BV1",
			"payment_purpose": "Dépôt légal dossier enregistrement princeps / importé",
		}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/dossiers/"+dossierID.String()+"/fees/generate-slip", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, raUser))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("attendu 201 Created, reçu %d: %s", w.Code, w.Body.String())
		}

		var slip domain.ANPPPaymentSlip
		if err := json.NewDecoder(w.Body).Decode(&slip); err != nil {
			t.Fatalf("décodage bordereau échoué: %v", err)
		}
		generatedSlipID = slip.ID
		if generatedSlipID == "" {
			t.Fatalf("slip.ID vide")
		}
	})

	// TEST 3 : Tentative de vérification 1.A.4 par un utilisateur R&D non habilité (SoD Failure)
	t.Run("03_SoD_Enforcement_Blocks_Unauthorized_User", func(t *testing.T) {
		receiptDocStr := receiptDocID.String()
		verifyPayload := domain.VerifyPaymentSlip1A4Request{
			BankName:          "Banque Nationale d'Algérie (BNA Agence Chéraga)",
			PaymentMode:       "virement",
			BankReceiptRef:    "BNA-VIR-2026-9988",
			BankReceiptDate:   time.Now().Format("2006-01-02"),
			ReceiptDocumentID: &receiptDocStr,
			Currency:          "DZD",
		}
		b, _ := json.Marshal(verifyPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/dossiers/"+dossierID.String()+"/fees/slips/"+generatedSlipID+"/verify-1a4", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, rdUser)) // R&D user
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("attendu 403 Forbidden pour SoD violation, reçu %d: %s", w.Code, w.Body.String())
		}
	})

	// TEST 4 : Contrôle de l'attestation Banque d'Algérie (Dossier importé en EUR sans scan BA -> non conforme)
	t.Run("04_FX_Attestation_Required_For_Imported_Currency", func(t *testing.T) {
		receiptDocStr := receiptDocID.String()
		eurAmt := 1720.50
		verifyPayload := domain.VerifyPaymentSlip1A4Request{
			BankName:              "Banque Nationale d'Algérie (BNA Agence Chéraga)",
			PaymentMode:           "virement",
			BankReceiptRef:        "BNA-VIR-2026-9988",
			BankReceiptDate:       time.Now().Format("2006-01-02"),
			ReceiptDocumentID:     &receiptDocStr,
			Currency:              "EUR",
			CurrencyAmount:        &eurAmt,
			RequiresFXAttestation: true,
			// Volontairement omission de FXRate et FXAttestationDocID
		}
		b, _ := json.Marshal(verifyPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/dossiers/"+dossierID.String()+"/fees/slips/"+generatedSlipID+"/verify-1a4", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, raUser))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu 200, reçu %d: %s", w.Code, w.Body.String())
		}

		var res struct {
			Slip    domain.ANPPPaymentSlip   `json:"slip"`
			Summary domain.Section1A4Summary `json:"summary"`
		}
		if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
			t.Fatalf("décodage json échoué: %v", err)
		}

		if res.Slip.VerificationStatus != "non_conforme" {
			t.Errorf("attendu verification_status='non_conforme' car scan BA manquant, reçu '%s'", res.Slip.VerificationStatus)
		}
		if res.Summary.OverallCompliant {
			t.Errorf("summary.OverallCompliant devrait être false")
		}
	})

	// TEST 5 : Vérification conforme complète (Quittance BNA + Attestation Taux de Change Banque d'Algérie)
	t.Run("05_Full_Verification_Compliant_With_FX_Attestation", func(t *testing.T) {
		receiptDocStr := receiptDocID.String()
		fxDocStr := fxDocID.String()
		fxRate := 145.3050
		fxDate := time.Now().Format("2006-01-02")
		eurAmt := 1720.50

		verifyPayload := domain.VerifyPaymentSlip1A4Request{
			BankName:              "Banque Nationale d'Algérie (BNA Agence Chéraga)",
			PaymentMode:           "virement",
			BankReceiptRef:        "BNA-VIR-2026-9988",
			BankReceiptDate:       time.Now().Format("2006-01-02"),
			ReceiptDocumentID:     &receiptDocStr,
			Currency:              "EUR",
			CurrencyAmount:        &eurAmt,
			RequiresFXAttestation: true,
			FXRate:                &fxRate,
			FXRateDate:            &fxDate,
			FXAttestationRef:      "BA-DGC-2026-11-8874",
			FXAttestationDocID:    &fxDocStr,
			VerificationNotes:     "Paiement redevance ANPP vérifié et rapproché du cours officiel Banque d'Algérie du 21/11/2026.",
		}
		b, _ := json.Marshal(verifyPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/dossiers/"+dossierID.String()+"/fees/slips/"+generatedSlipID+"/verify-1a4", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, raUser))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu 200 OK, reçu %d: %s", w.Code, w.Body.String())
		}

		var res struct {
			Slip    domain.ANPPPaymentSlip   `json:"slip"`
			Summary domain.Section1A4Summary `json:"summary"`
		}
		if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
			t.Fatalf("décodage json échoué: %v", err)
		}

		if res.Slip.VerificationStatus != "conforme" {
			t.Errorf("attendu verification_status='conforme', reçu '%s'", res.Slip.VerificationStatus)
		}
		if !res.Slip.AmountConcordanceOK {
			t.Errorf("attendu AmountConcordanceOK=true")
		}
		if !res.Summary.OverallCompliant {
			t.Errorf("attendu Summary.OverallCompliant=true")
		}

		// Vérifier en base que dossier_checklist_items pour 1.A.4 est bien passé à 'valide_qualite'
		var itemStatus, itemNotes string
		var gedDocID *uuid.UUID
		err := db.Pool.QueryRow(ctx, `
			SELECT status, validation_notes, ged_document_id
			FROM dossier_checklist_items
			WHERE dossier_id = $1 AND item_code = '1.A.4'
		`, dossierID).Scan(&itemStatus, &itemNotes, &gedDocID)
		if err != nil {
			t.Fatalf("lecture checklist 1.A.4 en base échouée: %v", err)
		}

		if itemStatus != "valide_qualite" {
			t.Errorf("attendu itemStatus='valide_qualite', reçu '%s'", itemStatus)
		}
		if gedDocID == nil || *gedDocID != receiptDocID {
			t.Errorf("gedDocID sur checklist 1.A.4 non rattaché à la quittance: %v", gedDocID)
		}
	})
}
