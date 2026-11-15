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

func TestDeficiencyAPI_Workflow(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-deficiency-" + testOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM anpp_deficiency_tasks WHERE organization_id = $1;
			DELETE FROM anpp_deficiency_letters WHERE organization_id = $1;
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
		VALUES ($1, $2, 'Laboratoire Test Compléments ANPP', 'MIPH-2026-COMPL', 'Dr. Amine Khodja')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation test: %v", err)
	}

	// 2. Créer un Directeur Technique (PDT - 1.A.7) et un contributeur réglementaire
	dtUserID := uuid.New()
	raUserID := uuid.New()
	pwdHash, _ := auth.HashPassword("TestPass123!#", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES 
			($1, $2, 'dt@test.dz', $3, 'Amine', 'K.', 'directeur_technique', TRUE),
			($4, $2, 'ra@test.dz', $3, 'Sarah', 'B.', 'responsable_affaires_reglementaires', TRUE)
	`, dtUserID, testOrgID, pwdHash, raUserID)
	if err != nil {
		t.Fatalf("échec création utilisateurs: %v", err)
	}

	// 3. Créer un produit et un dossier d'enregistrement
	productID := uuid.New()
	medID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, manufacturer_name, marketing_status)
		VALUES ($1, $2, 'medicament', 'AmoxiSanad 500mg', 'Laboratoires Sanad', 'en_preparation')
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
			$1, $2, $3, 'generique_local', 'AmoxiSanad 500mg',
			'Amoxicilline', '500mg', 'Gélule', 'Orale',
			'pcsu_local', 350.00, 'DZD', 'en_constitution'
		)
	`, medID, testOrgID, productID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	dossierID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, medicine_id, product_type, dossier_code, procedure_code, procedure_type, internal_reference, status)
		VALUES ($1, $2, $3, 'medicine', 'DOS-2026-AMOX-01', 'NF', 'enregistrement_national', 'REF-INT-001', 'en_evaluation')
	`, dossierID, testOrgID, medID)
	if err != nil {
		t.Fatalf("échec création dossier: %v", err)
	}

	// Instancier les repositories et handler
	auditService := audit.NewService(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	deficiencyRepo := repository.NewDeficiencyRepository(db.Pool)
	handler := NewDeficiencyHandler(deficiencyRepo, dossierRepo, auditService, db.Pool)

	// Routeur Chi
	r := chi.NewRouter()
	r.Get("/api/v1/dossiers/{id}/deficiencies", handler.ListDossierDeficiencies)
	r.Post("/api/v1/dossiers/{id}/deficiencies", handler.CreateDossierDeficiency)
	r.Get("/api/v1/deficiencies/letters/{letterId}", handler.GetDeficiencyLetter)
	r.Post("/api/v1/deficiencies/letters/{letterId}/attach-document", handler.AttachDocumentToLetter)
	r.Get("/api/v1/deficiencies/dashboard", handler.GetDeficiencyDashboard)
	r.Put("/api/v1/deficiencies/tasks/{taskId}", handler.UpdateDeficiencyTask)
	r.Post("/api/v1/deficiencies/tasks/{taskId}/validate-dt", handler.ValidateDeficiencyTaskDT)

	// Contexte pour le responsable réglementaire
	raUser := &auth.SessionUser{
		UserID:         raUserID,
		OrganizationID: testOrgID,
		Email:          "ra@test.dz",
		FirstName:      "Sarah",
		LastName:       "B.",
		Role:           "responsable_affaires_reglementaires",
		Permissions:    []string{string(domain.PermDossierView), string(domain.PermDossierCreate), string(domain.PermDocDraftMaker)},
	}

	// Contexte pour le Pharmacien Directeur Technique
	dtUser := &auth.SessionUser{
		UserID:         dtUserID,
		OrganizationID: testOrgID,
		Email:          "dt@test.dz",
		FirstName:      "Amine",
		LastName:       "K.",
		Role:           "directeur_technique",
		Permissions:    []string{string(domain.PermDossierView), string(domain.PermDocApprovePDT)},
	}

	var createdLetterID string
	var createdTaskID string

	// Test 1: Création d'une lettre de compléments avec décomposition de questions
	t.Run("POST /api/v1/dossiers/{id}/deficiencies", func(t *testing.T) {
		reqBody := domain.CreateDeficiencyLetterRequest{
			ReferenceNumber:   "ANPP/DEPP/N°1428/2026",
			ReceivedDate:      time.Now().Format("2006-01-02"),
			LegalDeadlineDays: 90,
			DueDate:           time.Now().AddDate(0, 0, 90).Format("2006-01-02"),
			Notes:             "Compléments prioritaires reçus suite à la session de commission",
			Tasks: []domain.CreateDeficiencyTaskInput{
				{
					QuestionNumber: 1,
					QuestionText:   "Fournir le profil d'impuretés selon 3.2.S.3.2 pour le lot pilote",
					ModuleTarget:   "Module 3",
					Priority:       "haute",
				},
				{
					QuestionNumber: 2,
					QuestionText:   "Préciser la durée de conservation après reconstitution de la suspension",
					ModuleTarget:   "Module 1",
					Priority:       "urgente",
				},
			},
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/deficiencies", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, raUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("attendu code 201 Created, obtenu %d: %s", w.Code, w.Body.String())
		}

		var resp domain.ANPPDeficiencyLetter
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("échec décodage réponse: %v", err)
		}

		if resp.ReferenceNumber != "ANPP/DEPP/N°1428/2026" {
			t.Errorf("numéro de référence incorrect: %s", resp.ReferenceNumber)
		}
		if len(resp.Tasks) != 2 {
			t.Fatalf("attendu 2 tâches, obtenu %d", len(resp.Tasks))
		}

		createdLetterID = resp.ID
		createdTaskID = resp.Tasks[0].ID
	})

	// Test 2: Récupération de la liste des lettres pour le dossier
	t.Run("GET /api/v1/dossiers/{id}/deficiencies", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/deficiencies", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, raUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200 OK, obtenu %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		json.NewDecoder(w.Body).Decode(&resp)
		letters := resp["letters"].([]any)
		if len(letters) != 1 {
			t.Errorf("attendu 1 lettre, obtenu %d", len(letters))
		}
	})

	// Test 3: Mise à jour d'une tâche de complément (réponse rédigée)
	t.Run("PUT /api/v1/deficiencies/tasks/{taskId}", func(t *testing.T) {
		summary := "Le rapport analytique validant le seuil d'impuretés sous 0.05% a été joint"
		reqBody := domain.UpdateDeficiencyTaskRequest{
			Status:          domain.DeficiencyTaskStatusAnswered,
			ResponseSummary: summary,
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("PUT", "/api/v1/deficiencies/tasks/"+createdTaskID, bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, raUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200 OK, obtenu %d: %s", w.Code, w.Body.String())
		}

		var updated domain.ANPPDeficiencyTask
		json.NewDecoder(w.Body).Decode(&updated)
		if updated.Status != domain.DeficiencyTaskStatusAnswered {
			t.Errorf("statut attendu 'repondu', obtenu '%s'", updated.Status)
		}
		if updated.ResponseSummary != summary {
			t.Errorf("résumé de réponse incorrect: %s", updated.ResponseSummary)
		}
	})

	// Test 4: Visa Pharmacien Directeur Technique - Refus si utilisateur non habilité
	t.Run("POST /api/v1/deficiencies/tasks/{taskId}/validate-dt [Refus RA non-PDT]", func(t *testing.T) {
		reqBody := domain.ValidateDeficiencyTaskRequest{Comment: "Validation par RA non autorisée"}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/deficiencies/tasks/"+createdTaskID+"/validate-dt", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, raUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("attendu code 403 Forbidden pour non-PDT, obtenu %d: %s", w.Code, w.Body.String())
		}
	})

	// Test 5: Visa Pharmacien Directeur Technique - Succès avec compte Directeur Technique (1.A.7)
	t.Run("POST /api/v1/deficiencies/tasks/{taskId}/validate-dt [Succès PDT]", func(t *testing.T) {
		reqBody := domain.ValidateDeficiencyTaskRequest{Comment: "Approuvé par le Directeur Technique conformément au Guide CTD 2025"}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/deficiencies/tasks/"+createdTaskID+"/validate-dt", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200 OK, obtenu %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]any
		json.NewDecoder(w.Body).Decode(&resp)
		taskMap := resp["task"].(map[string]any)
		if taskMap["status"] != domain.DeficiencyTaskStatusValidatedDT {
			t.Errorf("statut attendu '%s', obtenu '%s'", domain.DeficiencyTaskStatusValidatedDT, taskMap["status"])
		}
	})

	// Test 6: Tableau de bord global des compléments
	t.Run("GET /api/v1/deficiencies/dashboard", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/deficiencies/dashboard", nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200 OK, obtenu %d: %s", w.Code, w.Body.String())
		}

		var summary domain.DeficiencyDashboardSummary
		if err := json.NewDecoder(w.Body).Decode(&summary); err != nil {
			t.Fatalf("échec décodage dashboard: %v", err)
		}

		if summary.TotalActiveLetters != 1 {
			t.Errorf("total lettres actives attendu 1, obtenu %d", summary.TotalActiveLetters)
		}
		if summary.ValidatedTasksCount != 1 {
			t.Errorf("tâches validées attendu 1, obtenu %d", summary.ValidatedTasksCount)
		}
	})

	// Test 7: Récupération unitaire de la lettre par ID
	t.Run("GET /api/v1/deficiencies/letters/{letterId}", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/deficiencies/letters/"+createdLetterID, nil)
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200 OK, obtenu %d: %s", w.Code, w.Body.String())
		}

		var letter domain.ANPPDeficiencyLetter
		if err := json.NewDecoder(w.Body).Decode(&letter); err != nil {
			t.Fatalf("échec décodage lettre: %v", err)
		}

		if letter.ID != createdLetterID {
			t.Errorf("ID de lettre attendu %s, obtenu %s", createdLetterID, letter.ID)
		}
		if len(letter.Tasks) != 2 {
			t.Errorf("attendu 2 tâches, obtenu %d", len(letter.Tasks))
		}
	})

	// Test 8: Rattachement du scan officiel ANPP archivé en GED
	t.Run("POST /api/v1/deficiencies/letters/{letterId}/attach-document [Rattachement Scan GED]", func(t *testing.T) {
		testDocID := uuid.New()
		_, err := db.Pool.Exec(context.Background(), `
			INSERT INTO ged_documents (id, organization_id, taxonomy_path, title_fr, document_code, status, current_version)
			VALUES ($1, $2, '01_Affaires_Reglementaires/01_06_Correspondances_ANPP', 'Notification Officielle ANPP Scan Numérisé', 'SCAN-ANPP-TEST', 'brouillon', 'v0.1')
		`, testDocID, testOrgID)
		if err != nil {
			t.Fatalf("échec insertion document test GED: %v", err)
		}

		attachPayload := map[string]string{
			"document_id": testDocID.String(),
		}
		bodyBytes, _ := json.Marshal(attachPayload)

		req := httptest.NewRequest("POST", "/api/v1/deficiencies/letters/"+createdLetterID+"/attach-document", bytes.NewReader(bodyBytes))
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, raUser))
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("attendu code 200 OK lors du rattachement scan, obtenu %d: %s", w.Code, w.Body.String())
		}

		var updatedLetter domain.ANPPDeficiencyLetter
		if err := json.NewDecoder(w.Body).Decode(&updatedLetter); err != nil {
			t.Fatalf("échec décodage lettre mise à jour: %v", err)
		}

		if updatedLetter.DocumentID == nil || *updatedLetter.DocumentID != testDocID.String() {
			t.Errorf("DocumentID attendu %s, obtenu %v", testDocID.String(), updatedLetter.DocumentID)
		}
		if updatedLetter.DocumentTitle != "Notification Officielle ANPP Scan Numérisé" {
			t.Errorf("DocumentTitle attendu 'Notification Officielle ANPP Scan Numérisé', obtenu %s", updatedLetter.DocumentTitle)
		}
	})
}
