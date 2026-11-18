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

func TestWorkflowGatingIntegration(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-gating-" + testOrgID.String()[:8]

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
		VALUES ($1, $2, 'Laboratoire Gating DAG BPF', 'MIPH-2026-GAT', 'Dr. Yacine Benali')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation: %v", err)
	}

	// 2. Utilisateurs : Maker (Rédacteur), Checker (AQ), PDT (Pharmacien DT)
	makerID := uuid.New()
	checkerID := uuid.New()
	pdtID := uuid.New()
	pwdHash, _ := auth.HashPassword("Pass123!#", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES 
		($1, $4, 'maker@gating.dz', $5, 'Karim', 'Maker', 'contributeur_rd_prod', TRUE),
		($2, $4, 'checker@gating.dz', $5, 'Sarah', 'Checker', 'contributeur_qualite_bpf', TRUE),
		($3, $4, 'pdt@gating.dz', $5, 'Yacine', 'Benali', 'directeur_technique', TRUE);
	`, makerID, checkerID, pdtID, testOrgID, pwdHash)
	if err != nil {
		t.Fatalf("échec création utilisateurs: %v", err)
	}

	makerUser := &auth.SessionUser{UserID: makerID, OrganizationID: testOrgID, Email: "maker@gating.dz", Role: "contributeur_rd_prod"}
	pdtUser := &auth.SessionUser{UserID: pdtID, OrganizationID: testOrgID, Email: "pdt@gating.dz", Role: "directeur_technique"}

	// 3. Produit & Dossier
	prodID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, dci, dosage, pharmaceutical_form, manufacturer_name)
		VALUES ($1, $2, 'medicament', 'Amoxicilline Biocare', 'Amoxicilline', '500 mg', 'Gélule', 'Biocare Algérie')
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
			$1, $2, $3, 'generique_local', 'Amoxicilline Biocare 500mg',
			'Amoxicilline', '500 mg', 'Gélule', 'Orale',
			'J01CA04', 'pcsu'
		)
	`, medID, testOrgID, prodID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	dossierID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, medicine_id, product_id, product_type, dossier_code, internal_reference, procedure_code, status, completeness_percent)
		VALUES ($1, $2, $3, $4, 'medicament', 'DOS-2026-AMOX-GAT', 'REF-GAT-001', 'NF', 'en_preparation', 50)
	`, dossierID, testOrgID, medID, prodID)
	if err != nil {
		t.Fatalf("échec création dossier: %v", err)
	}

	// 4. Création des pièces de checklist du DAG :
	// - 3.2.P.8 (Stabilités Zone IVb) : initialement draft_pending
	// - 1.P.2.1 (Notice Patient FR) : initialement pending_approval
	// - 1.P.2.2 (Notice Patient AR) : initialement drafting
	stabID := uuid.New()
	noticeFRID := uuid.New()
	noticeARID := uuid.New()

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossier_checklist_items (
			id, dossier_id, module_number, sub_module, item_code, title_fr, requirement_level, is_mandatory,
			status, workflow_status, assigned_department, current_revision, sort_order
		) VALUES 
		($1, $4, 3, '3.2.P', '3.2.P.8', 'Stabilité et études en conditions Zone IVb (30°C/75% RH)', 'E', TRUE, 'en_attente', 'draft_pending', 'RD', 'v0.1', 80),
		($2, $4, 1, '1.P', '1.P.2.1', 'Notice patient pour le public en langue française', 'E', TRUE, 'en_attente', 'pending_approval', 'RA', 'v0.2', 21),
		($3, $4, 1, '1.P', '1.P.2.2', 'Notice patient pour le public en langue arabe', 'E', TRUE, 'en_attente', 'drafting', 'RA', 'v0.1', 22);
	`, stabID, noticeFRID, noticeARID, dossierID)
	if err != nil {
		t.Fatalf("échec création items checklist: %v", err)
	}

	// 5. Initialisation Repositories & Handlers
	auditService := audit.NewService(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	workflowRepo := repository.NewWorkflowRepository(db.Pool)

	wfHandler := NewWorkflowHandler(workflowRepo, dossierRepo, auditService, db.Pool)

	buildRouter := func(activeUser *auth.SessionUser) *chi.Mux {
		r := chi.NewRouter()
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				ctx := context.WithValue(req.Context(), UserContextKey, activeUser)
				next.ServeHTTP(w, req.WithContext(ctx))
			})
		})
		r.Post("/api/v1/dossiers/{id}/items/{itemId}/submit-review", wfHandler.SubmitReview)
		r.Post("/api/v1/dossiers/{id}/items/{itemId}/approve", wfHandler.ApproveSection)
		r.Get("/api/v1/dossiers/{id}/items/{itemId}/gating", wfHandler.GetItemGating)
		r.Get("/api/v1/dossiers/{id}/gating-summary", wfHandler.GetDossierGatingSummary)
		return r
	}

	// =========================================================================
	// TEST 1 : Blocage Gating 3.2.P.8 ➔ 1.P.2.1 (Notice FR bloquée sans stabilités)
	// =========================================================================
	pdtRouter := buildRouter(pdtUser)
	approveBody, _ := json.Marshal(map[string]any{
		"solemn_declaration_accepted": true,
	})

	req := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+noticeFRID.String()+"/approve", bytes.NewReader(approveBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	pdtRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Attendu HTTP 422 (GATING_BLOCKED) pour l'approbation de la Notice sans stabilités, obtenu HTTP %d (body: %s)", w.Code, w.Body.String())
	}

	var blockedResp struct {
		Success            bool                       `json:"success"`
		Code               string                     `json:"code"`
		Error              string                     `json:"error"`
		UnmetPrerequisites []domain.UnmetPrerequisite `json:"unmet_prerequisites"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &blockedResp); err != nil {
		t.Fatalf("échec décodage JSON réponse 422: %v", err)
	}
	if blockedResp.Code != "GATING_BLOCKED" {
		t.Fatalf("Code d'erreur attendu GATING_BLOCKED, obtenu: %s", blockedResp.Code)
	}
	if len(blockedResp.UnmetPrerequisites) == 0 {
		t.Fatalf("Attendu au moins un prérequis non satisfait")
	}
	if blockedResp.UnmetPrerequisites[0].RequiredItemCode != "3.2.P.8" {
		t.Fatalf("Attendu blocage par 3.2.P.8, obtenu: %s", blockedResp.UnmetPrerequisites[0].RequiredItemCode)
	}

	// =========================================================================
	// TEST 2 : Endpoint d'inspection UI GET /items/{itemId}/gating
	// =========================================================================
	reqInspect := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/items/"+noticeFRID.String()+"/gating", nil)
	wInspect := httptest.NewRecorder()
	pdtRouter.ServeHTTP(wInspect, reqInspect)

	if wInspect.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour inspection gating, obtenu HTTP %d", wInspect.Code)
	}
	var inspectResp struct {
		Success    bool                          `json:"success"`
		Evaluation domain.GatingEvaluationResult `json:"evaluation"`
	}
	if err := json.Unmarshal(wInspect.Body.Bytes(), &inspectResp); err != nil {
		t.Fatalf("échec décodage inspection: %v", err)
	}
	if !inspectResp.Evaluation.IsBlocked {
		t.Fatalf("L'évaluation UI aurait dû marquer la Notice FR comme bloquée")
	}

	// =========================================================================
	// TEST 3 : Endpoint Gating Summary GET /dossiers/{id}/gating-summary
	// =========================================================================
	reqSummary := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/gating-summary", nil)
	wSummary := httptest.NewRecorder()
	pdtRouter.ServeHTTP(wSummary, reqSummary)

	if wSummary.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour summary gating, obtenu HTTP %d", wSummary.Code)
	}
	var sumResp struct {
		Success bool                            `json:"success"`
		Summary repository.DossierGatingSummary `json:"summary"`
	}
	if err := json.Unmarshal(wSummary.Body.Bytes(), &sumResp); err != nil {
		t.Fatalf("échec décodage summary: %v", err)
	}
	if sumResp.Summary.BlockedCount == 0 {
		t.Fatalf("Le résumé du dossier devrait contenir au moins 1 item bloqué")
	}

	// =========================================================================
	// TEST 4 : Scellement de 3.2.P.8 ➔ Déverrouillage de la Notice FR
	// =========================================================================
	_, err = db.Pool.Exec(ctx, `
		UPDATE dossier_checklist_items
		SET workflow_status = 'approved_sealed',
		    current_revision = 'v1.0',
		    status = 'approuve_dt',
		    approver_user_id = $1,
		    approver_signed_at = NOW(),
		    approver_signature_seal = 'SEAL-STAB-OK'
		WHERE id = $2
	`, pdtID, stabID)
	if err != nil {
		t.Fatalf("échec mise à jour 3.2.P.8: %v", err)
	}

	// Tentative d'approbation de Notice FR maintenant que Stabilités sont scellées
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+noticeFRID.String()+"/approve", bytes.NewReader(approveBody))
	req2.Header.Set("Content-Type", "application/json")
	pdtRouter.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour l'approbation de Notice FR après déverrouillage stabilités, obtenu HTTP %d: %s", w2.Code, w2.Body.String())
	}

	// =========================================================================
	// TEST 5 : Parité Linguistique Notice AR (1.P.2.2)
	// =========================================================================
	makerRouter := buildRouter(makerUser)
	submitARBody, _ := json.Marshal(map[string]any{
		"notes": "Tentative de soumission notice arabe",
	})

	// Notice AR est en drafting, Notice FR est maintenant approved_sealed ➔ La soumission de Notice AR en revue Checker doit réussir !
	wAR := httptest.NewRecorder()
	reqAR := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierID.String()+"/items/"+noticeARID.String()+"/submit-review", bytes.NewReader(submitARBody))
	reqAR.Header.Set("Content-Type", "application/json")
	makerRouter.ServeHTTP(wAR, reqAR)

	if wAR.Code != http.StatusOK {
		t.Fatalf("Attendu HTTP 200 pour soumission notice AR quand Notice FR est scellée, obtenu HTTP %d: %s", wAR.Code, wAR.Body.String())
	}
}
