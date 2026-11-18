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

func TestWorkflowMakerCheckerCycle(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-wf-" + testOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM generated_reports WHERE organization_id = $1;
			DELETE FROM dossier_item_reviews WHERE dossier_item_id IN (
				SELECT id FROM dossier_checklist_items WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id = $1)
			);
			DELETE FROM dossier_checklist_items WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id = $1);
			DELETE FROM dossiers WHERE organization_id = $1;
			DELETE FROM medicines WHERE organization_id = $1;
			DELETE FROM products WHERE organization_id = $1;
			DELETE FROM users WHERE organization_id = $1;
			DELETE FROM organizations WHERE id = $1;
		`, testOrgID)
	}()

	// 1. Organisation de test
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Laboratoire Workflow BPF Test', 'MIPH-2026-WF', 'Dr. Yacine Benali')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation: %v", err)
	}

	// 2. Trois collaborateurs distincts pour tester la Ségrégation des Fonctions (SoD)
	makerID := uuid.New()
	checkerID := uuid.New()
	pdtID := uuid.New()
	pwdHash, _ := auth.HashPassword("Pass123!#", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES 
		($1, $4, 'maker@test.dz', $5, 'Karim', 'Maker', 'contributeur_rd_prod', TRUE),
		($2, $4, 'checker@test.dz', $5, 'Sarah', 'Checker', 'contributeur_qualite_bpf', TRUE),
		($3, $4, 'pdt@test.dz', $5, 'Yacine', 'Benali', 'directeur_technique', TRUE);
	`, makerID, checkerID, pdtID, testOrgID, pwdHash)
	if err != nil {
		t.Fatalf("échec création utilisateurs: %v", err)
	}

	// 3. Produit & Dossier CTD
	prodID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, dci, dosage, pharmaceutical_form, manufacturer_name)
		VALUES ($1, $2, 'medicament', 'Paracétamol Biocare', 'Paracétamol', '500 mg', 'Comprimé', 'Biocare Algérie')
	`, prodID, testOrgID)
	if err != nil {
		t.Fatalf("échec création produit: %v", err)
	}

	medID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO medicines (
			id, organization_id, product_id, origin_type, brand_name_fr,
			dci_fr, dosage_fr, pharmaceutical_form_fr, administration_route_fr,
			atc_code, economic_file_type
		) VALUES (
			$1, $2, $3, 'generique_local', 'Paracétamol Biocare 500mg',
			'Paracétamol', '500 mg', 'Comprimé', 'Orale',
			'N02BE01', 'pcsu'
		)
	`, medID, testOrgID, prodID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	dossierID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, medicine_id, product_id, product_type, dossier_code, internal_reference, procedure_code, status, completeness_percent)
		VALUES ($1, $2, $3, $4, 'medicament', 'DOS-2026-PARAC-WF', 'REF-WF-001', 'NF', 'en_preparation', 40)
	`, dossierID, testOrgID, medID, prodID)
	if err != nil {
		t.Fatalf("échec création dossier: %v", err)
	}

	// 4. Section de test (Module 3.2.P.2 - Développement Pharmaceutique)
	itemID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossier_checklist_items (
			id, dossier_id, module_number, sub_module, item_code, title_fr, requirement_level, is_mandatory,
			status, workflow_status, assigned_department, current_revision, sort_order
		) VALUES (
			$1, $2, 3, '3.2.P', '3.2.P.2', 'Développement pharmaceutique et justification de la formulation',
			'E', TRUE, 'manquant', 'draft_pending', 'RD', 'v0.1', 20
		)
	`, itemID, dossierID)
	if err != nil {
		t.Fatalf("échec création item checklist: %v", err)
	}

	// 5. Initialisation Repositories & Handlers
	auditService := audit.NewService(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	workflowRepo := repository.NewWorkflowRepository(db.Pool)
	pharmaRepo := repository.NewPharmaRepository(db.Pool)

	wfHandler := NewWorkflowHandler(workflowRepo, dossierRepo, auditService, db.Pool)
	reportHandler := NewReportHandler(dossierRepo, workflowRepo, pharmaRepo, auditService, db.Pool)

	// Helper pour injecter l'utilisateur connecté dans le contexte
	buildRouter := func(activeUser *auth.SessionUser) *chi.Mux {
		r := chi.NewRouter()
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				ctx := context.WithValue(req.Context(), UserContextKey, activeUser)
				next.ServeHTTP(w, req.WithContext(ctx))
			})
		})
		r.Post("/api/v1/dossiers/{id}/items/{itemId}/checkout", wfHandler.CheckoutSection)
		r.Post("/api/v1/dossiers/{id}/items/{itemId}/checkin", wfHandler.CheckinSection)
		r.Post("/api/v1/dossiers/{id}/items/{itemId}/submit-review", wfHandler.SubmitReview)
		r.Post("/api/v1/dossiers/{id}/items/{itemId}/review", wfHandler.ReviewDecision)
		r.Post("/api/v1/dossiers/{id}/items/{itemId}/approve", wfHandler.ApproveSection)
		r.Get("/api/v1/dossiers/{id}/items/{itemId}/reviews", wfHandler.GetItemReviews)
		r.Get("/api/v1/templates", wfHandler.GetMasterTemplates)
		r.Get("/api/v1/templates/{code}", wfHandler.GetMasterTemplateByCode)
		r.Get("/api/v1/dossiers/{id}/reports/{code}", reportHandler.GenerateDossierReport)
		r.Post("/api/v1/dossiers/{id}/reports/{code}/seal", reportHandler.SealDossierReport)
		r.Get("/api/v1/public/verify/{sha256}", reportHandler.VerifyReportPublic)
		return r
	}

	makerUser := &auth.SessionUser{
		UserID:         makerID,
		OrganizationID: testOrgID,
		Email:          "maker@test.dz",
		Role:           "contributeur_rd_prod",
	}
	checkerUser := &auth.SessionUser{
		UserID:         checkerID,
		OrganizationID: testOrgID,
		Email:          "checker@test.dz",
		Role:           "contributeur_qualite_bpf",
	}
	pdtUser := &auth.SessionUser{
		UserID:         pdtID,
		OrganizationID: testOrgID,
		Email:          "pdt@test.dz",
		Role:           "directeur_technique",
	}

	// =========================================================================
	// Test 1 : Check-out par le Maker (Verrouillage pour rédaction)
	// =========================================================================
	rMaker := buildRouter(makerUser)
	coPayload, _ := json.Marshal(map[string]string{"reason": "Rédaction justification galénique"})
	req := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+itemID.String()+"/checkout", bytes.NewReader(coPayload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	rMaker.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Test 1: Checkout attendu 200, reçu %d: %s", rr.Code, rr.Body.String())
	}

	// =========================================================================
	// Test 2 : Tentative de Check-out par un tiers pendant que c'est verrouillé (Conflit 409)
	// =========================================================================
	rChecker := buildRouter(checkerUser)
	req2 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+itemID.String()+"/checkout", bytes.NewReader(coPayload))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	rChecker.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusConflict {
		t.Fatalf("Test 2: Conflit attendu 409, reçu %d", rr2.Code)
	}

	// =========================================================================
	// Test 3 : Check-in par le Maker avec nouveau contenu (v0.2 et déverrouillage)
	// =========================================================================
	ciPayload, _ := json.Marshal(map[string]interface{}{
		"content": "# Justification galénique\nLe choix des excipients garantit une dissolution > 85% en 15 min.",
		"comment": "Finalisation du projet initial",
	})
	req3 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+itemID.String()+"/checkin", bytes.NewReader(ciPayload))
	req3.Header.Set("Content-Type", "application/json")
	rr3 := httptest.NewRecorder()
	rMaker.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Fatalf("Test 3: Checkin attendu 200, reçu %d: %s", rr3.Code, rr3.Body.String())
	}

	// =========================================================================
	// Test 4 : Soumission pour revue Qualité (Checker / AQ)
	// =========================================================================
	subPayload, _ := json.Marshal(map[string]string{"notes": "Prêt pour contrôle qualité"})
	req4 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+itemID.String()+"/submit-review", bytes.NewReader(subPayload))
	req4.Header.Set("Content-Type", "application/json")
	rr4 := httptest.NewRecorder()
	rMaker.ServeHTTP(rr4, req4)

	if rr4.Code != http.StatusOK {
		t.Fatalf("Test 4: Soumission revue attendue 200, reçu %d: %s", rr4.Code, rr4.Body.String())
	}

	// =========================================================================
	// Test 5 : Garde-fou SoD : Le Maker tente d'approuver sa propre section (Rejeté 403)
	// =========================================================================
	revPayload, _ := json.Marshal(map[string]string{"decision": "approved", "notes": "Auto-validation interdite"})
	req5 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+itemID.String()+"/review", bytes.NewReader(revPayload))
	req5.Header.Set("Content-Type", "application/json")
	rr5 := httptest.NewRecorder()
	rMaker.ServeHTTP(rr5, req5)

	if rr5.Code != http.StatusForbidden {
		t.Fatalf("Test 5: Auto-revue Maker devrait être interdite (403), reçu %d", rr5.Code)
	}

	// =========================================================================
	// Test 6 : Revue légitime par le Checker AQ (Validation qualité -> pending_approval)
	// =========================================================================
	revOkPayload, _ := json.Marshal(map[string]string{"decision": "approved", "notes": "Conforme ICH Q8(R2)"})
	req6 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+itemID.String()+"/review", bytes.NewReader(revOkPayload))
	req6.Header.Set("Content-Type", "application/json")
	rr6 := httptest.NewRecorder()
	rChecker.ServeHTTP(rr6, req6)

	if rr6.Code != http.StatusOK {
		t.Fatalf("Test 6: Revue Checker attendue 200, reçu %d: %s", rr6.Code, rr6.Body.String())
	}

	// =========================================================================
	// Test 7 : Tentative d'approbation finale par un non-DT (Rejeté 403)
	// =========================================================================
	appPayload, _ := json.Marshal(map[string]interface{}{
		"solemn_declaration": true,
	})
	req7 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+itemID.String()+"/approve", bytes.NewReader(appPayload))
	req7.Header.Set("Content-Type", "application/json")
	rr7 := httptest.NewRecorder()
	rChecker.ServeHTTP(rr7, req7)

	if rr7.Code != http.StatusForbidden {
		t.Fatalf("Test 7: Seul le PDT peut approuver (403 attendu pour Checker), reçu %d", rr7.Code)
	}

	// =========================================================================
	// Test 8 : Visa officiel 21 CFR Part 11 par le Pharmacien DT (Approbation & Scellement v1.0)
	// =========================================================================
	rPDT := buildRouter(pdtUser)
	req8 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+itemID.String()+"/approve", bytes.NewReader(appPayload))
	req8.Header.Set("Content-Type", "application/json")
	rr8 := httptest.NewRecorder()
	rPDT.ServeHTTP(rr8, req8)

	if rr8.Code != http.StatusOK {
		t.Fatalf("Test 8: Approbation PDT attendue 200, reçu %d: %s", rr8.Code, rr8.Body.String())
	}

	var appResp struct {
		Success       bool                 `json:"success"`
		SignatureSeal string               `json:"signature_seal"`
		Item          domain.ChecklistItem `json:"item"`
	}
	_ = json.Unmarshal(rr8.Body.Bytes(), &appResp)

	if appResp.Item.WorkflowStatus != domain.WorkflowStatusApprovedSealed {
		t.Fatalf("Statut workflow attendu approved_sealed, reçu: %s", appResp.Item.WorkflowStatus)
	}
	if appResp.Item.CurrentRevision != "v1.0" {
		t.Fatalf("Version scellée attendue v1.0, reçu: %s", appResp.Item.CurrentRevision)
	}
	if appResp.SignatureSeal == "" {
		t.Fatal("Sceau de signature cryptographique SHA-256 manquant")
	}

	// =========================================================================
	// Test 9 : Consultation de l'historique complet des annotations (Audit Trail)
	// =========================================================================
	req9 := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/items/"+itemID.String()+"/reviews", nil)
	rr9 := httptest.NewRecorder()
	rPDT.ServeHTTP(rr9, req9)

	if rr9.Code != http.StatusOK {
		t.Fatalf("Test 9: Historique attendu 200, reçu %d", rr9.Code)
	}

	var reviewsResp struct {
		Total   int                        `json:"total"`
		Reviews []domain.DossierItemReview `json:"reviews"`
	}
	_ = json.Unmarshal(rr9.Body.Bytes(), &reviewsResp)
	if reviewsResp.Total < 4 {
		t.Fatalf("Audit trail incomplet : attendu au moins 4 traces de revue, reçu %d", reviewsResp.Total)
	}

	// =========================================================================
	// Test 10 : Génération dynamique de l'État de Sortie ES-REG-01 (Bordereau CTD)
	// =========================================================================
	req10 := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/reports/es-reg-01", nil)
	rr10 := httptest.NewRecorder()
	rPDT.ServeHTTP(rr10, req10)

	if rr10.Code != http.StatusOK {
		t.Fatalf("Test 10: Rendu HTML ES-REG-01 attendu 200, reçu %d", rr10.Code)
	}
	htmlBody := rr10.Body.String()
	if !strings.Contains(htmlBody, "ES-REG-01") {
		t.Fatal("Rendu ES-REG-01 ne contient pas le code de référence")
	}
	if !strings.Contains(htmlBody, "Paracétamol") {
		t.Fatal("Rendu ES-REG-01 ne contient pas les données réelles du produit (Paracétamol)")
	}
	if !strings.Contains(htmlBody, "https://homolog.labo.dz/verify/") {
		t.Fatal("Rendu ES-REG-01 ne contient pas le lien QR-code de vérification cryptographique")
	}

	// =========================================================================
	// Test 11 : Génération en format JSON et scellement formel
	// =========================================================================
	req11 := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/reports/es-reg-01?format=json", nil)
	rr11 := httptest.NewRecorder()
	rPDT.ServeHTTP(rr11, req11)

	if rr11.Code != http.StatusOK {
		t.Fatalf("Test 11: Format JSON attendu 200, reçu %d", rr11.Code)
	}
	var repJSON struct {
		Success bool                   `json:"success"`
		Report  domain.GeneratedReport `json:"report"`
	}
	_ = json.Unmarshal(rr11.Body.Bytes(), &repJSON)
	if repJSON.Report.SHA256Hash == "" {
		t.Fatal("Empreinte SHA-256 du rapport vide")
	}

	// =========================================================================
	// Test 12 : Vérification publique de l'authenticité par scan QR-Code
	// =========================================================================
	req12 := httptest.NewRequest("GET", "/api/v1/public/verify/"+repJSON.Report.SHA256Hash, nil)
	rr12 := httptest.NewRecorder()
	rPDT.ServeHTTP(rr12, req12)

	if rr12.Code != http.StatusOK {
		t.Fatalf("Test 12: Portail public attendu 200, reçu %d", rr12.Code)
	}
	if !strings.Contains(rr12.Body.String(), "Certificat d'Authenticité Réglementaire") {
		t.Fatal("Certificat public d'authenticité manquant dans la réponse")
	}

	// =========================================================================
	// Test 13 : Modèles Maîtres ANPP / ICH (Engagements)
	// =========================================================================
	req13 := httptest.NewRequest("GET", "/api/v1/templates", nil)
	rr13 := httptest.NewRecorder()
	rPDT.ServeHTTP(rr13, req13)

	if rr13.Code != http.StatusOK {
		t.Fatalf("Test 13: Templates attendu 200, reçu %d", rr13.Code)
	}
	var tmplResp struct {
		Total     int                             `json:"total"`
		Templates []domain.MasterDocumentTemplate `json:"templates"`
	}
	_ = json.Unmarshal(rr13.Body.Bytes(), &tmplResp)
	if tmplResp.Total < 3 {
		t.Fatalf("Attendu au moins 3 templates maîtres pré-enregistrés, reçu: %d", tmplResp.Total)
	}

	// =========================================================================
	// Test 14 : Génération des Nouveaux États Réglementaires BPF (Jalon 2.2)
	// =========================================================================
	extendedCodes := []struct {
		code       string
		expectedSub string
	}{
		{"ES-REG-03", "Table des Matières Exhaustive & Index Paginé CTD"},
		{"ES-REG-04", "Fiche Récapitulative de Modification"},
		{"ES-REG-05", "Bordereau d'Envoi des Échantillons & Réactifs (1.P)"},
		{"ES-FIN-02", "Bordereau de Versement Variations, Renouvellements & Transferts"},
		{"1.P.8", "Module 1.P.8 — Engagement de Disponibilité Continue"},
		{"1.P.9", "Module 1.P.9 — Engagement de Pharmacovigilance"},
		{"1.H.10", "Module 1.H.10 — Déclaration Solennelle de Non-Modification"},
	}

	for _, tc := range extendedCodes {
		reqExt := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/reports/"+tc.code, nil)
		rrExt := httptest.NewRecorder()
		rPDT.ServeHTTP(rrExt, reqExt)

		if rrExt.Code != http.StatusOK {
			t.Fatalf("Test 14: Échec génération rapport %s: attendu 200, reçu %d", tc.code, rrExt.Code)
		}
		bodyExt := rrExt.Body.String()
		if !strings.Contains(bodyExt, tc.expectedSub) {
			t.Fatalf("Test 14: Rapport %s ne contient pas le sous-titre attendu '%s'", tc.code, tc.expectedSub)
		}
		if !strings.Contains(bodyExt, "data:image/png;base64,") {
			t.Fatalf("Test 14: Rapport %s ne contient pas le QR-code cryptographique", tc.code)
		}
		if !strings.Contains(bodyExt, "Page 1 sur 1") {
			t.Fatalf("Test 14: Rapport %s ne contient pas la pagination indélébile BPF", tc.code)
		}
	}

	// =========================================================================
	// Test 15 : Génération des 8 États Intermédiaires de Pilotage (EI) (Jalon 2.2)
	// =========================================================================
	eiCodes := []struct {
		code        string
		expectedSub string
	}{
		{"EI-AVT-01", "Fiche Navette d'Avancement CTD / DM"},
		{"EI-AVT-02", "Relevé des Écarts & Pièces Manquantes"},
		{"EI-AVT-03", "Fiche de Revue & d'Annotation Checker"},
		{"EI-CHR-01", "Suivi des Stabilités Zone IVb"},
		{"EI-CHR-02", "Rapport d'Étape de l'Étude de Bioéquivalence"},
		{"EI-VIG-01", "Instruction des Compléments & Plan d'Action Interne"},
		{"EI-CHG-01", "Évaluation d'Impact Réglementaire"},
		{"EI-PIP-01", "Point de Situation Hebdomadaire du Pipeline"},
	}

	for _, tc := range eiCodes {
		reqEI := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/reports/"+tc.code, nil)
		rrEI := httptest.NewRecorder()
		rPDT.ServeHTTP(rrEI, reqEI)

		if rrEI.Code != http.StatusOK {
			t.Fatalf("Test 15: Échec génération rapport EI %s: attendu 200, reçu %d", tc.code, rrEI.Code)
		}
		bodyEI := rrEI.Body.String()
		if !strings.Contains(bodyEI, tc.expectedSub) {
			t.Fatalf("Test 15: Rapport %s ne contient pas le sous-titre attendu '%s'", tc.code, tc.expectedSub)
		}
		if !strings.Contains(bodyEI, "DOCUMENT INTERNE DE TRAVAIL") {
			t.Fatalf("Test 15: Rapport %s ne contient pas le filigrane confidentiel", tc.code)
		}
	}
}
