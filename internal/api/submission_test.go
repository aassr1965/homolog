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

func TestSubmissionAPI_Workflow(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-subm-" + testOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM anpp_payment_slips WHERE organization_id = $1;
			DELETE FROM dossier_checklist_items WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id = $1);
			DELETE FROM dossiers WHERE organization_id = $1;
			DELETE FROM establishments WHERE organization_id = $1;
			DELETE FROM medicines WHERE organization_id = $1;
			DELETE FROM products WHERE organization_id = $1;
			DELETE FROM users WHERE organization_id = $1;
			DELETE FROM organizations WHERE id = $1;
		`, testOrgID)
	}()

	// 1. Créer l'organisation de test
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Laboratoires Soumission E-Services Test', 'MIPH-2026-SUBM', 'Dr. Yacine Benkhalid')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation test: %v", err)
	}

	// 2. Créer l'établissement demandeur
	estID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO establishments (id, organization_id, name_fr, legal_status, miph_license_number, technical_director_name, technical_director_decision_ref, wilaya, is_active)
		VALUES ($1, $2, 'Laboratoires Soumission E-Services Test', 'SPA', 'MIPH-2026-SUBM', 'Dr. Yacine Benkhalid', 'DEC-MIPH-DT-2026/099', 'Alger', TRUE)
	`, estID, testOrgID)
	if err != nil {
		t.Fatalf("échec création établissement: %v", err)
	}

	// 3. Créer un Directeur Technique (PDT)
	dtUserID := uuid.New()
	pwdHash, _ := auth.HashPassword("TestPass123!#", nil)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, 'dt-subm@test.dz', $3, 'Yacine', 'B.', 'directeur_technique', TRUE)
	`, dtUserID, testOrgID, pwdHash)
	if err != nil {
		t.Fatalf("échec création utilisateur DT: %v", err)
	}

	// 4. Créer un produit et un dossier
	productID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, manufacturer_name, marketing_status)
		VALUES ($1, $2, 'medicament', 'PARACETAL TEST 500mg', 'Laboratoires Sanad', 'en_preparation')
	`, productID, testOrgID)
	if err != nil {
		t.Fatalf("échec création produit: %v", err)
	}

	medID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO medicines (
			id, organization_id, product_id, origin_type, brand_name_fr,
			dci_fr, dosage_fr, pharmaceutical_form_fr, administration_route_fr,
			economic_file_type, fob_price, fob_currency, status
		) VALUES (
			$1, $2, $3, 'generique_local', 'PARACETAL TEST 500mg',
			'PARACETAMOL', '500 mg', 'Comprimé', 'Orale',
			'pcsu_local', 180.00, 'DZD', 'en_constitution'
		)
	`, medID, testOrgID, productID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	dossierID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, medicine_id, product_type, dossier_code, procedure_code, internal_reference, status, responsible_id, completeness_percent)
		VALUES ($1, $2, $3, 'medicament', 'DOS-2026-SUBM-01', 'NF', 'INT-SUBM-01', 'en_preparation', $4, 0)
	`, dossierID, testOrgID, medID, dtUserID)
	if err != nil {
		t.Fatalf("échec création dossier: %v", err)
	}

	// Ajouter 1 item obligatoire et 1 item facultatif
	itemReqID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossier_checklist_items (id, dossier_id, module_number, sub_module, item_code, title_fr, requirement_level, is_mandatory, status, sort_order)
		VALUES ($1, $2, 1, '1.A', '1.A.1', 'Formulaire de demande AMM', 'E', TRUE, 'manquant', 10)
	`, itemReqID, dossierID)
	if err != nil {
		t.Fatalf("échec création item obligatoire: %v", err)
	}

	itemFacID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossier_checklist_items (id, dossier_id, module_number, sub_module, item_code, title_fr, requirement_level, is_mandatory, status, sort_order)
		VALUES ($1, $2, 1, '1.F', '1.F.1', 'Échantillons de conditionnement', 'F', FALSE, 'manquant', 20)
	`, itemFacID, dossierID)
	if err != nil {
		t.Fatalf("échec création item facultatif: %v", err)
	}

	// Initialiser les dépendances
	auditService := audit.NewService(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	feeRepo := repository.NewFeeRepository(db.Pool)
	pharmaRepo := repository.NewPharmaRepository(db.Pool)

	submissionHandler := NewSubmissionHandler(dossierRepo, feeRepo, pharmaRepo, auditService)

	dtUser := &auth.SessionUser{
		UserID:         dtUserID,
		OrganizationID: testOrgID,
		Email:          "dt-subm@test.dz",
		FirstName:      "Yacine",
		LastName:       "B.",
		Role:           "directeur_technique",
		Permissions:    []string{string(domain.PermDossierView), string(domain.PermDossierCreate), string(domain.PermDocApprovePDT)},
	}

	// Routeur de test
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), UserContextKey, dtUser)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	submissionHandler.RegisterRoutes(r)

	// =========================================================================
	// Étape 1 : Audit pré-dépôt sur dossier incomplet -> doit être BLOQUÉ
	// =========================================================================
	req := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/pre-submission-audit", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Étape 1: statut attendu 200, reçu %d: %s", rr.Code, rr.Body.String())
	}

	var report domain.PreSubmissionValidationReport
	if err := json.Unmarshal(rr.Body.Bytes(), &report); err != nil {
		t.Fatalf("Étape 1: échec parsing rapport audit: %v", err)
	}

	if report.IsEligible {
		t.Errorf("Étape 1: le dossier incomplet ne devrait PAS être éligible au télé-dépôt")
	}
	if report.CriticalBlockersCount == 0 {
		t.Errorf("Étape 1: des bloqueurs critiques auraient dû être détectés")
	}
	t.Logf("Étape 1 PASS: Audit pré-dépôt dossier incomplet détecté comme non éligible (%d bloqueurs, %d avertissements)", report.CriticalBlockersCount, report.WarningsCount)

	// =========================================================================
	// Étape 2 : Tentative de verrouillage pré-dépôt sur dossier incomplet -> 422
	// =========================================================================
	lockPayload := `{"submission_notes": "Tentative prématurée"}`
	req = httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/lock-pre-submission", strings.NewReader(lockPayload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Étape 2: statut attendu 422 Unprocessable Entity, reçu %d: %s", rr.Code, rr.Body.String())
	}
	t.Logf("Étape 2 PASS: Tentative de verrouillage rejetée avec 422 Unprocessable Entity")

	// =========================================================================
	// Étape 3 : Résolution des exigences (Rattachement document, Paiement BV1, Visa PDT)
	// =========================================================================
	// 3a. Créer et régler un bordereau BV1
	now := time.Now().UTC()
	var feeSchedID string
	err = db.Pool.QueryRow(ctx, "SELECT id FROM anpp_fee_schedules WHERE procedure_code = 'generique_local' AND is_active = TRUE LIMIT 1").Scan(&feeSchedID)
	if err != nil {
		t.Fatalf("Étape 3a: échec sélection barème: %v", err)
	}

	slipID := uuid.New().String()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO anpp_payment_slips (
			id, organization_id, dossier_id, fee_schedule_id, slip_type, slip_number,
			payment_purpose, amount_dzd, amount_words_fr, amount_words_ar,
			bank_account_rib, status, payment_reference, paid_at
		) VALUES (
			$1, $2, $3, $4, 'BV1', 'BV1-2026-TEST-001',
			'Redevance enregistrement initial', 150000,
			'Cent cinquante mille dinars algériens', 'مائة وخمسون ألف دينار جزائري',
			'007.99999.0001234567.89', 'paye', 'QUIT-TRESOR-SUBM-991', $5
		)
	`, slipID, testOrgID, dossierID.String(), feeSchedID, now)
	if err != nil {
		t.Fatalf("Étape 3a: échec création BV1 payé: %v", err)
	}

	// 3b. Créer un document GED et rattacher à l'item 1.A.1
	docID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO ged_documents (id, organization_id, title_fr, document_code, current_version, status, taxonomy_path, created_by)
		VALUES ($1, $2, 'Formulaire AMM Officiel Dûment Rempli', 'FORM-AMM-001', 'v1.0', 'en_vigueur', '01_Affaires_Reglementaires', $3)
	`, docID, testOrgID, dtUserID)
	if err != nil {
		t.Fatalf("Étape 3b: échec création doc GED: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		UPDATE dossier_checklist_items
		SET ged_document_id = $1, status = 'approuve_dt', updated_at = $2
		WHERE id = $3
	`, docID, now, itemReqID)
	if err != nil {
		t.Fatalf("Étape 3b: échec update item: %v", err)
	}

	// 3c. Apposer le Visa formel du Pharmacien DT
	seal := "pdt-seal-sha256-test-valid"
	pdtRole := "Directeur Technique (1.A.7)"
	_, err = db.Pool.Exec(ctx, `
		UPDATE dossiers
		SET pdt_signed_at = $1, pdt_signed_by = $2, pdt_signature_seal = $3, pdt_signature_role = $4, completeness_percent = 100
		WHERE id = $5
	`, now, dtUserID, seal, pdtRole, dossierID)
	if err != nil {
		t.Fatalf("Étape 3c: échec visa DT: %v", err)
	}

	// =========================================================================
	// Étape 4 : Ré-audit pré-dépôt -> Doit être CONFORME et ÉLIGIBLE
	// =========================================================================
	req = httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/pre-submission-audit", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Étape 4: statut attendu 200, reçu %d: %s", rr.Code, rr.Body.String())
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &report); err != nil {
		t.Fatalf("Étape 4: échec parsing rapport audit: %v", err)
	}

	if !report.IsEligible {
		t.Fatalf("Étape 4: le dossier complet DEVRAIT être éligible, bloqueurs : %v", report.Checks)
	}
	if report.CriticalBlockersCount != 0 {
		t.Errorf("Étape 4: attendu 0 bloqueurs critiques, reçu %d", report.CriticalBlockersCount)
	}
	if report.CompletenessScore != 100 {
		t.Errorf("Étape 4: attendu 100%% complétude, reçu %d%%", report.CompletenessScore)
	}
	t.Logf("Étape 4 PASS: Audit pré-dépôt 100%% CONFORME et ÉLIGIBLE au télé-dépôt E-Services ANPP")

	// =========================================================================
	// Étape 5 : Génération Manifeste E-Services JSON
	// =========================================================================
	req = httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/eservices-manifest?format=json", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Étape 5: statut attendu 200, reçu %d: %s", rr.Code, rr.Body.String())
	}

	var manifest domain.ANPPEservicesManifest
	if err := json.Unmarshal(rr.Body.Bytes(), &manifest); err != nil {
		t.Fatalf("Étape 5: échec parsing manifeste JSON: %v", err)
	}

	if manifest.PortalStandard != "ANPP_ESERVICES_REGULATORY_PACKAGE_V2.4" {
		t.Errorf("Étape 5: standard attendu ANPP_ESERVICES_REGULATORY_PACKAGE_V2.4, reçu %s", manifest.PortalStandard)
	}
	if manifest.FinancialClearance.PaymentReference != "QUIT-TRESOR-SUBM-991" {
		t.Errorf("Étape 5: quittance Trésor BV1 mismatch")
	}
	if manifest.IntegritySeal.ManifestDigestHex == "" {
		t.Errorf("Étape 5: sceau d'intégrité SHA-256 manquant")
	}
	t.Logf("Étape 5 PASS: Manifeste JSON généré avec succès (Digest: %s)", manifest.IntegritySeal.ManifestDigestHex)

	// =========================================================================
	// Étape 6 : Génération Manifeste E-Services XML
	// =========================================================================
	req = httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/eservices-manifest?format=xml", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Étape 6: statut attendu 200, reçu %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "application/xml") {
		t.Errorf("Étape 6: Content-Type attendu application/xml, reçu %s", rr.Header().Get("Content-Type"))
	}
	xmlBody := rr.Body.String()
	if !strings.Contains(xmlBody, "<ANPPEservicesManifest>") || !strings.Contains(xmlBody, "QUIT-TRESOR-SUBM-991") {
		t.Errorf("Étape 6: contenu XML invalide ou incomplet: %s", xmlBody)
	}
	t.Logf("Étape 6 PASS: Manifeste XML interopérable conforme généré")

	// =========================================================================
	// Étape 7 : Verrouillage et Certification pré-dépôt officielle
	// =========================================================================
	validLockPayload := `{"submission_notes": "Vérifié conforme pour télé-transmission E-Services session Novembre 2026"}`
	req = httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/lock-pre-submission", bytes.NewBufferString(validLockPayload))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Étape 7: statut attendu 200, reçu %d: %s", rr.Code, rr.Body.String())
	}

	var lockResp struct {
		Success        bool   `json:"success"`
		Message        string `json:"message"`
		ManifestDigest string `json:"manifest_digest"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &lockResp); err != nil {
		t.Fatalf("Étape 7: échec unmarshal réponse verrouillage: %v", err)
	}

	if !lockResp.Success || lockResp.ManifestDigest == "" {
		t.Errorf("Étape 7: résultat de verrouillage invalide")
	}
	t.Logf("Étape 7 PASS: Dossier verrouillé et certifié avec succès (Digest: %s)", lockResp.ManifestDigest)
}
