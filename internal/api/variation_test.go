package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

func TestVariationAPI_RubriquesCatalogue(t *testing.T) {
	varRepo := repository.NewVariationRepository(nil)
	handler := NewVariationHandler(varRepo, nil, nil)

	r := chi.NewRouter()
	r.Get("/api/v1/variations/rubriques", handler.ListRubriques)
	r.Get("/api/v1/variations/rubriques/{code}", handler.GetRubrique)

	// 1. Liste complète des 34 rubriques
	req := httptest.NewRequest("GET", "/api/v1/variations/rubriques", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Statut attendu 200, obtenu: %d", rec.Code)
	}

	var resp struct {
		Success   bool                        `json:"success"`
		Total     int                         `json:"total"`
		Rubriques []domain.RubriqueDefinition `json:"rubriques"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Échec parsing JSON catalogue: %v", err)
	}
	if resp.Total < 25 {
		t.Errorf("Le catalogue doit contenir au moins 25 rubriques, obtenu: %d", resp.Total)
	}

	// 2. Recherche spécifique de la rubrique 7.d.1
	req2 := httptest.NewRequest("GET", "/api/v1/variations/rubriques/7.d.1", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("Statut attendu 200 pour rubrique 7.d.1, obtenu: %d", rec2.Code)
	}
	var resp2 struct {
		Success  bool                      `json:"success"`
		Rubrique domain.RubriqueDefinition `json:"rubrique"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("Échec parsing JSON rubrique 7.d.1: %v", err)
	}
	if resp2.Rubrique.Code != "7.d.1" {
		t.Errorf("Code attendu '7.d.1', obtenu: %s", resp2.Rubrique.Code)
	}
	if resp2.Rubrique.DefaultClassification != domain.ClassMmod {
		t.Errorf("Classification nominale attendue Mmod, obtenu: %s", resp2.Rubrique.DefaultClassification)
	}
	if resp2.Rubrique.EscalatedClassification != domain.ClassMmaj {
		t.Errorf("Classification escaladée attendue Mmaj, obtenu: %s", resp2.Rubrique.EscalatedClassification)
	}
	if len(resp2.Rubrique.Conditions) != 3 {
		t.Errorf("Rubrique 7.d.1 doit avoir 3 conditions préalables C1..C3, obtenu: %d", len(resp2.Rubrique.Conditions))
	}
}

func TestVariationAPI_FullLifecycle_EscalationAndGrouping(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-var-" + testOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM anpp_variation_documents WHERE variation_request_id IN (SELECT id FROM anpp_variation_requests WHERE organization_id = $1);
			DELETE FROM anpp_variation_items WHERE variation_request_id IN (SELECT id FROM anpp_variation_requests WHERE organization_id = $1);
			DELETE FROM anpp_variation_requests WHERE organization_id = $1;
			DELETE FROM dossier_checklist_items WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id = $1);
			DELETE FROM dossiers WHERE organization_id = $1;
			DELETE FROM ged_documents WHERE organization_id = $1;
			DELETE FROM medicines WHERE organization_id = $1;
			DELETE FROM products WHERE organization_id = $1;
			DELETE FROM users WHERE organization_id = $1;
			DELETE FROM organizations WHERE id = $1;
		`, testOrgID)
	}()

	// 1. Créer l'organisation et utilisateurs
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Laboratoire Test Variations ANPP', 'MIPH-2026-VAR', 'Dr. Yacine Benalia')
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
			($1, $2, 'dt-var@test.dz', $3, 'Yacine', 'B.', 'directeur_technique', TRUE),
			($4, $2, 'ra-var@test.dz', $3, 'Nadia', 'L.', 'responsable_reglementaire', TRUE)
	`, dtUserID, testOrgID, pwdHash, raUserID)
	if err != nil {
		t.Fatalf("échec création utilisateurs: %v", err)
	}

	// 2. Créer un produit et médicament enregistré
	productID := uuid.New()
	medID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, manufacturer_name, marketing_status)
		VALUES ($1, $2, 'medicament', 'Amoxicilline Sanad 500mg', 'Laboratoire Sanad Pharma', 'commercialise')
	`, productID, testOrgID)
	if err != nil {
		t.Fatalf("échec création produit: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO medicines (
			id, organization_id, product_id, origin_type, brand_name_fr,
			dci_fr, dosage_fr, pharmaceutical_form_fr, administration_route_fr,
			economic_file_type, registration_decision_number, status
		) VALUES (
			$1, $2, $3, 'generique_local', 'Amoxicilline Sanad 500mg',
			'Amoxicilline trihydratée', '500 mg', 'Gélule', 'Voie orale',
			'pcsu_local', '2026/01/1234/DE', 'enregistre'
		)
	`, medID, testOrgID, productID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	// 3. Document GED pour rattachement
	gedDocID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO ged_documents (id, organization_id, title_fr, taxonomy_path, status, current_version, created_by)
		VALUES ($1, $2, 'Justificatif Synthèse API Gujarat', '01_Affaires_Reglementaires/01_04_Dossiers_CTD_Reference', 'valide_qualite', 'v1.0', $3)
	`, gedDocID, testOrgID, dtUserID)
	if err != nil {
		t.Fatalf("échec création doc GED: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO ged_versions (id, document_id, version_label, revision_summary, file_sha256, file_size_bytes, mime_type, original_filename, storage_path, author_id)
		VALUES (gen_random_uuid(), $1, 'v1.0', 'Dépôt initial', 'hash-test-sha256-var', 1024, 'application/pdf', 'justificatif_synthese_api.pdf', '/storage/test', $2)
	`, gedDocID, dtUserID)
	if err != nil {
		t.Fatalf("échec création version GED: %v", err)
	}

	varRepo := repository.NewVariationRepository(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	auditService := audit.NewService(db.Pool)
	handler := NewVariationHandler(varRepo, dossierRepo, auditService)

	// Router de test
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Simuler l'utilisateur DT par défaut
			user := &auth.SessionUser{
				UserID:         dtUserID,
				OrganizationID: testOrgID,
				Email:          "dt-var@test.dz",
				Role:           "directeur_technique",
			}
			ctx := context.WithValue(req.Context(), UserContextKey, user)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	r.Post("/api/v1/variations", handler.CreateVariation)
	r.Get("/api/v1/variations/{id}", handler.GetVariation)
	r.Post("/api/v1/variations/{id}/items", handler.AddVariationItem)
	r.Put("/api/v1/variations/{id}/items/{itemId}", handler.UpdateVariationItem)
	r.Post("/api/v1/variations/{id}/documents/{docId}/attach", handler.AttachDocument)
	r.Post("/api/v1/variations/{id}/seal", handler.SealVariationDT)
	r.Get("/api/v1/variations/{id}/comparative-table", handler.GetComparativeTableHTML)

	// -------------------------------------------------------------------------
	// Étape A : Créer la demande de variation
	// -------------------------------------------------------------------------
	createPayload := map[string]interface{}{
		"medicine_id":                medID.String(),
		"tracking_number":            "VAR-2026-AMOX-01",
		"title":                      "Changement de site de fabrication API et optimisation de synthèse",
		"description":                "Introduction d'un nouveau site de synthèse de l'Amoxicilline trihydratée",
		"current_approved_situation": "Synthèse par voie chimique classique sur le site d'origine (Site A, Hyderabad, Inde).",
		"proposed_situation":         "Synthèse enzymatique éco-compatible sur le nouveau site B (Gujarat, Inde).",
		"technical_justification":    "Réduction des impuretés connexes et augmentation des rendements conforme ICH Q11.",
		"non_modification_attested":  true,
		"is_grouping":                false,
	}
	bodyBytes, _ := json.Marshal(createPayload)
	reqA := httptest.NewRequest("POST", "/api/v1/variations", bytes.NewReader(bodyBytes))
	recA := httptest.NewRecorder()
	r.ServeHTTP(recA, reqA)

	if recA.Code != http.StatusCreated {
		t.Fatalf("Échec création demande variation, code: %d, body: %s", recA.Code, recA.Body.String())
	}

	var respA struct {
		Success   bool                    `json:"success"`
		Variation domain.VariationRequest `json:"variation"`
	}
	_ = json.Unmarshal(recA.Body.Bytes(), &respA)
	variationID := respA.Variation.ID
	if variationID == uuid.Nil {
		t.Fatalf("ID de variation invalide généré")
	}

	// -------------------------------------------------------------------------
	// Étape B : Vérifier l'initialisation automatique des 12 pièces de la LD-DEPP07
	// -------------------------------------------------------------------------
	reqB := httptest.NewRequest("GET", "/api/v1/variations/"+variationID.String(), nil)
	recB := httptest.NewRecorder()
	r.ServeHTTP(recB, reqB)

	var respB struct {
		Success   bool                    `json:"success"`
		Variation domain.VariationRequest `json:"variation"`
	}
	_ = json.Unmarshal(recB.Body.Bytes(), &respB)
	if len(respB.Variation.Documents) != 12 {
		t.Fatalf("Les 12 pièces obligatoires de la LD-DEPP07 doivent être initialisées, obtenu: %d", len(respB.Variation.Documents))
	}

	// -------------------------------------------------------------------------
	// Étape C : Ajouter la rubrique 7.d.1 avec condition C2 NON REMPLIE -> ESCALADE Mmaj !
	// -------------------------------------------------------------------------
	itemPayload := map[string]interface{}{
		"rubrique_code": "7.d.1",
		"conditions_evaluated": map[string]bool{
			"C1": true,
			"C2": false, // Stabilités Zone IVb 3 lots non conclues -> ESCALADE !
			"C3": true,
		},
		"current_state": "Site A (Hyderabad, Inde)",
		"proposed_state": "Site B (Gujarat, Inde)",
		"notes":          "Stabilités Zone IVb en cours d'achèvement",
	}
	bodyC, _ := json.Marshal(itemPayload)
	reqC := httptest.NewRequest("POST", "/api/v1/variations/"+variationID.String()+"/items", bytes.NewReader(bodyC))
	recC := httptest.NewRecorder()
	r.ServeHTTP(recC, reqC)

	if recC.Code != http.StatusCreated {
		t.Fatalf("Échec ajout rubrique 7.d.1, code: %d, body: %s", recC.Code, recC.Body.String())
	}

	var respC struct {
		Success     bool                  `json:"success"`
		Item        domain.VariationItem  `json:"item"`
		Variation   domain.VariationRequest `json:"variation"`
		IsEscalated bool                  `json:"is_escalated"`
		Escalation  string                `json:"escalation"`
	}
	_ = json.Unmarshal(recC.Body.Bytes(), &respC)

	if !respC.IsEscalated {
		t.Errorf("L'item 7.d.1 sans C2 doit être reclassé (is_escalated = true)")
	}
	if respC.Item.EffectiveClassification != domain.ClassMmaj {
		t.Errorf("La classification effective doit être Mmaj, obtenu: %s", respC.Item.EffectiveClassification)
	}
	if respC.Variation.OverallClassification != domain.ClassMmaj {
		t.Errorf("La classification globale de la demande doit s'ajuster à Mmaj, obtenu: %s", respC.Variation.OverallClassification)
	}
	if respC.Variation.FeeAmountDZD != 115000.00 {
		t.Errorf("La redevance pour Mmaj doit être 115 000 DA, obtenu: %f", respC.Variation.FeeAmountDZD)
	}

	// -------------------------------------------------------------------------
	// Étape D : Grouping (Article 09) - Ajouter une deuxième rubrique 1 (Mmin)
	// -------------------------------------------------------------------------
	itemPayload2 := map[string]interface{}{
		"rubrique_code": "1",
		"conditions_evaluated": map[string]bool{
			"C1": true,
			"C2": true,
		},
		"current_state": "Amoxicilline Sanad 500mg",
		"proposed_state": "Amoxisanad 500mg",
	}
	bodyD, _ := json.Marshal(itemPayload2)
	reqD := httptest.NewRequest("POST", "/api/v1/variations/"+variationID.String()+"/items", bytes.NewReader(bodyD))
	recD := httptest.NewRecorder()
	r.ServeHTTP(recD, reqD)

	if recD.Code != http.StatusCreated {
		t.Fatalf("Échec ajout rubrique grouping 1, code: %d", recD.Code)
	}

	// Recharger la demande pour vérifier le grouping
	reqD2 := httptest.NewRequest("GET", "/api/v1/variations/"+variationID.String(), nil)
	recD2 := httptest.NewRecorder()
	r.ServeHTTP(recD2, reqD2)

	var respD2 struct {
		Variation domain.VariationRequest `json:"variation"`
	}
	_ = json.Unmarshal(recD2.Body.Bytes(), &respD2)

	if !respD2.Variation.IsGrouping {
		t.Errorf("Avec 2 rubriques, is_grouping doit être true (Grouping Article 09)")
	}
	if respD2.Variation.OverallClassification != domain.ClassMmaj {
		t.Errorf("Le Grouping contenant une Mmaj doit rester au niveau global Mmaj, obtenu: %s", respD2.Variation.OverallClassification)
	}

	// -------------------------------------------------------------------------
	// Étape E : Rattacher un document GED à la pièce DOC-02 (Tableau Comparatif)
	// -------------------------------------------------------------------------
	doc2ID := respB.Variation.Documents[1].ID // Document index 2
	attachPayload := map[string]interface{}{
		"ged_document_id": gedDocID.String(),
	}
	bodyE, _ := json.Marshal(attachPayload)
	reqE := httptest.NewRequest("POST", "/api/v1/variations/"+variationID.String()+"/documents/"+doc2ID.String()+"/attach", bytes.NewReader(bodyE))
	recE := httptest.NewRecorder()
	r.ServeHTTP(recE, reqE)

	if recE.Code != http.StatusOK {
		t.Fatalf("Échec rattachement pièce justificative DOC-02, code: %d, body: %s", recE.Code, recE.Body.String())
	}

	// -------------------------------------------------------------------------
	// Étape F : Scellement 21 CFR Part 11 par le Pharmacien DT
	// -------------------------------------------------------------------------
	// F.1 : Refus sans engagement solennel
	sealPayloadNoSolemn := map[string]interface{}{
		"solemn_declaration": false,
		"notes":              "Tentative sans engagement",
	}
	bodyF1, _ := json.Marshal(sealPayloadNoSolemn)
	reqF1 := httptest.NewRequest("POST", "/api/v1/variations/"+variationID.String()+"/seal", bytes.NewReader(bodyF1))
	recF1 := httptest.NewRecorder()
	r.ServeHTTP(recF1, reqF1)
	if recF1.Code != http.StatusBadRequest {
		t.Errorf("Le scellement sans engagement solennel doit être rejeté (HTTP 400), obtenu: %d", recF1.Code)
	}

	// F.2 : Succès avec engagement solennel du Pharmacien DT
	sealPayloadOK := map[string]interface{}{
		"solemn_declaration": true,
		"notes":              "Dossier de variation vérifié et scellé pour soumission ANPP.",
	}
	bodyF2, _ := json.Marshal(sealPayloadOK)
	reqF2 := httptest.NewRequest("POST", "/api/v1/variations/"+variationID.String()+"/seal", bytes.NewReader(bodyF2))
	recF2 := httptest.NewRecorder()
	r.ServeHTTP(recF2, reqF2)

	if recF2.Code != http.StatusOK {
		t.Fatalf("Échec scellement DT, code: %d, body: %s", recF2.Code, recF2.Body.String())
	}

	var respF2 struct {
		Success  bool   `json:"success"`
		Seal     string `json:"seal"`
		SealedAt string `json:"sealed_at"`
	}
	_ = json.Unmarshal(recF2.Body.Bytes(), &respF2)
	if respF2.Seal == "" {
		t.Errorf("Un sceau cryptographique officiel doit être retourné")
	}

	// -------------------------------------------------------------------------
	// Étape G : Rendu Haute Fidélité A4 du Tableau Comparatif ES-REG-04
	// -------------------------------------------------------------------------
	reqG := httptest.NewRequest("GET", "/api/v1/variations/"+variationID.String()+"/comparative-table", nil)
	recG := httptest.NewRecorder()
	r.ServeHTTP(recG, reqG)

	if recG.Code != http.StatusOK {
		t.Fatalf("Statut attendu 200 pour rendu HTML ES-REG-04, obtenu: %d", recG.Code)
	}

	htmlOut := recG.Body.String()
	if !bytes.Contains([]byte(htmlOut), []byte("ES-REG-04")) {
		t.Errorf("Le rendu doit comporter le code officiel ES-REG-04")
	}
	if !bytes.Contains([]byte(htmlOut), []byte("LD-DEPP07/V01")) {
		t.Errorf("Le rendu doit mentionner la Ligne Directrice LD-DEPP07/V01")
	}
	if !bytes.Contains([]byte(htmlOut), []byte("Amoxicilline Sanad 500mg")) {
		t.Errorf("Le rendu doit contenir la dénomination de la spécialité")
	}
	if !bytes.Contains([]byte(htmlOut), []byte("7.d.1")) {
		t.Errorf("Le rendu doit afficher la rubrique 7.d.1")
	}
	if !bytes.Contains([]byte(htmlOut), []byte("115 000,00 DZD")) && !bytes.Contains([]byte(htmlOut), []byte("115000.00 DZD")) {
		t.Errorf("Le montant de la redevance 115 000 DZD doit apparaître")
	}
}
