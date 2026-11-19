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

func TestSmartPlanningAPIIntegration(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-sp-" + testOrgID.String()[:8]

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

	// 1. Organisation
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Laboratoires Test SmartPlanning', 'MIPH-2026-SP', 'Dr. Yacine Benali')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation: %v", err)
	}

	// 2. Utilisateurs
	pdtID := uuid.New()
	pwdHash, _ := auth.HashPassword("Pass123!#", nil)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, 'pdt@smartplanning.dz', $3, 'Yacine', 'Benali', 'directeur_technique', TRUE);
	`, pdtID, testOrgID, pwdHash)
	if err != nil {
		t.Fatalf("échec création utilisateur: %v", err)
	}

	pdtUser := &auth.SessionUser{UserID: pdtID, OrganizationID: testOrgID, Email: "pdt@smartplanning.dz", Role: "directeur_technique"}

	// 3. Produit & Dossier
	prodID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, dci, dosage, pharmaceutical_form, manufacturer_name)
		VALUES ($1, $2, 'medicament', 'Ciprofloxacine Biocare', 'Ciprofloxacine', '500 mg', 'Comprimé', 'Biocare Algérie')
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
			$1, $2, $3, 'generique_local', 'Ciprofloxacine Biocare 500mg',
			'Ciprofloxacine', '500 mg', 'Comprimé', 'Orale',
			'J01MA02', 'pcsu'
		)
	`, medID, testOrgID, prodID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	dossierID := uuid.New()
	// Créer un dossier avec une date cible dans 10 mois
	targetDate := time.Now().AddDate(0, 10, 0).Format("2006-01-02")
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (
			id, organization_id, medicine_id, product_type, dossier_code,
			procedure_code, internal_reference, model_template_id, model_version,
			status, completeness_percent, target_deposit_date
		) VALUES (
			$1, $2, $3, 'medicament', 'DOS-2026-CIPRO-NI',
			'NI', 'DOS-2026-CIPRO-NI', 'ctd_generique_v1', '1.0',
			'en_preparation', 25, $4
		)
	`, dossierID, testOrgID, medID, targetDate)
	if err != nil {
		t.Fatalf("échec création dossier: %v", err)
	}

	// 4. Checklist items
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossier_checklist_items (
			id, dossier_id, module_number, sub_module, item_code,
			title_fr, requirement_level, is_mandatory, status, workflow_status, sort_order
		) VALUES
		(gen_random_uuid(), $1, 3, '3.2.P', '3.2.P.8', 'Études de Stabilité Zone IVb', 'E', TRUE, 'manquant', 'draft_pending', 1),
		(gen_random_uuid(), $1, 2, '2.3', '2.3', 'Résumé Global de la Qualité (QOS)', 'E', TRUE, 'manquant', 'draft_pending', 2),
		(gen_random_uuid(), $1, 1, '1.A', '1.A.1', 'Formulaire de demande d''enregistrement', 'E', TRUE, 'manquant', 'draft_pending', 3),
		(gen_random_uuid(), $1, 1, '1.P', '1.P.1', 'Résumé des Caractéristiques du Produit (RCP)', 'E', TRUE, 'manquant', 'draft_pending', 4),
		(gen_random_uuid(), $1, 1, '1.P', '1.P.2.1', 'Notice pour le Patient (Français)', 'E', TRUE, 'manquant', 'draft_pending', 5),
		(gen_random_uuid(), $1, 1, '1.P', '1.P.2.2', 'Notice pour le Patient (Arabe)', 'E', TRUE, 'manquant', 'draft_pending', 6),
		(gen_random_uuid(), $1, 1, '1.A', '1.A.7', 'Visa et Engagement du Pharmacien DT', 'E', TRUE, 'manquant', 'draft_pending', 7)
	`, dossierID)
	if err != nil {
		t.Fatalf("échec insertion items: %v", err)
	}

	// 5. Instanciation du handler
	wfRepo := repository.NewWorkflowRepository(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	auditService := audit.NewService(db.Pool)
	wfHandler := NewWorkflowHandler(wfRepo, dossierRepo, auditService, db.Pool)

	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), UserContextKey, pdtUser))
			next.ServeHTTP(w, r)
		})
	})
	router.Get("/api/v1/dossiers/{id}/smart-planning", wfHandler.GetDossierSmartPlanning)
	router.Post("/api/v1/dossiers/{id}/smart-planning/set-target", wfHandler.SetDossierTargetDates)
	router.Post("/api/v1/dossiers/{id}/smart-planning/sync-due-dates", wfHandler.SyncSmartPlanningDueDates)

	// =========================================================================
	// TEST 1 : GET /smart-planning (Lecture du rétroplanning et des 6 jalons)
	// =========================================================================
	reqGet := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/smart-planning", nil)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)

	if wGet.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour GET smart-planning, obtenu %d: %s", wGet.Code, wGet.Body.String())
	}

	var getResp struct {
		Success  bool                         `json:"success"`
		Planning domain.SmartPlanningAnalysis `json:"planning"`
	}
	if err := json.Unmarshal(wGet.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("Erreur décodage JSON GET smart-planning: %v", err)
	}

	if !getResp.Success {
		t.Fatalf("Attendu success=true")
	}
	if len(getResp.Planning.Milestones) != 6 {
		t.Fatalf("Attendu 6 jalons critiques, obtenu %d", len(getResp.Planning.Milestones))
	}
	if getResp.Planning.OverallHealth != domain.HealthOnTrack {
		t.Fatalf("Pour 10 mois de marge, attendu HealthOnTrack, obtenu %s", getResp.Planning.OverallHealth)
	}

	// =========================================================================
	// TEST 2 : POST /smart-planning/set-target (Modification date cible de dépôt)
	// =========================================================================
	newTargetDate := time.Now().AddDate(0, 12, 0).Format("2006-01-02")
	bodySet, _ := json.Marshal(map[string]interface{}{
		"target_deposit_date": newTargetDate,
	})
	reqSet := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/smart-planning/set-target", bytes.NewReader(bodySet))
	reqSet.Header.Set("Content-Type", "application/json")
	wSet := httptest.NewRecorder()
	router.ServeHTTP(wSet, reqSet)

	if wSet.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour set-target, obtenu %d: %s", wSet.Code, wSet.Body.String())
	}

	// =========================================================================
	// TEST 3 : POST /smart-planning/sync-due-dates (Propagation des dates aux tâches)
	// =========================================================================
	reqSync := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/smart-planning/sync-due-dates", nil)
	wSync := httptest.NewRecorder()
	router.ServeHTTP(wSync, reqSync)

	if wSync.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour sync-due-dates, obtenu %d: %s", wSync.Code, wSync.Body.String())
	}

	var syncResp struct {
		Success           bool `json:"success"`
		UpdatedItemsCount int  `json:"updated_items_count"`
	}
	if err := json.Unmarshal(wSync.Body.Bytes(), &syncResp); err != nil {
		t.Fatalf("Erreur décodage sync-due-dates: %v", err)
	}
	if syncResp.UpdatedItemsCount == 0 {
		t.Fatalf("Attendu au moins 1 item mis à jour, obtenu 0")
	}

	// Vérifier en base que 3.2.P.8 a bien reçu une due_date non vide
	var dueDateCheck *time.Time
	err = db.Pool.QueryRow(ctx, `SELECT due_date FROM dossier_checklist_items WHERE dossier_id = $1 AND item_code = '3.2.P.8'`, dossierID).Scan(&dueDateCheck)
	if err != nil {
		t.Fatalf("Échec vérification due_date en base: %v", err)
	}
	if dueDateCheck == nil {
		t.Fatalf("due_date pour 3.2.P.8 n'a pas été mise à jour en base")
	}
}
