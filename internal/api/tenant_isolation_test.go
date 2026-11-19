package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/repository"
)

// TestGlobalTenantIsolationSecurity vérifie l'étanchéité multi-tenant absolue sur Workflow, GED et Membres
func TestGlobalTenantIsolationSecurity(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()

	// Création de 2 organisations distinctes et concurrentes
	orgA := uuid.New()
	orgB := uuid.New()
	slugA := "lab-a-" + orgA.String()[:8]
	slugB := "lab-b-" + orgB.String()[:8]

	defer func() {
		cleanup := func(orgID uuid.UUID) {
			_, _ = db.Pool.Exec(context.Background(), `
				DELETE FROM dossier_item_reviews WHERE dossier_item_id IN (
					SELECT id FROM dossier_checklist_items WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id = $1)
				);
				DELETE FROM dossier_assignments WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id = $1);
				DELETE FROM dossier_checklist_items WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id = $1);
				DELETE FROM dossiers WHERE organization_id = $1;
				DELETE FROM ged_versions WHERE document_id IN (SELECT id FROM ged_documents WHERE organization_id = $1);
				DELETE FROM ged_documents WHERE organization_id = $1;
				DELETE FROM medicines WHERE organization_id = $1;
				DELETE FROM products WHERE organization_id = $1;
				DELETE FROM users WHERE organization_id = $1;
				DELETE FROM organizations WHERE id = $1;
			`, orgID)
		}
		cleanup(orgA)
		cleanup(orgB)
	}()

	// 1. Enregistrement des 2 organisations
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES 
			($1, $2, 'Laboratoire Alpha Pharma', 'MIPH-2026-ALPHA', 'Dr. Alpha DT'),
			($3, $4, 'Laboratoire Beta Pharma', 'MIPH-2026-BETA', 'Dr. Beta DT');
	`, orgA, slugA, orgB, slugB)
	if err != nil {
		t.Fatalf("échec création organisations: %v", err)
	}

	// 2. Utilisateurs : UserA (Alpha) et UserB (Beta)
	userAID := uuid.New()
	userBID := uuid.New()
	pwdHash, _ := auth.HashPassword("PassSec123!#", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES 
			($1, $2, 'dt@alpha.dz', $3, 'Directeur', 'Alpha', 'directeur_technique', TRUE),
			($4, $5, 'dt@beta.dz', $3, 'Directeur', 'Beta', 'directeur_technique', TRUE);
	`, userAID, orgA, pwdHash, userBID, orgB)
	if err != nil {
		t.Fatalf("échec création utilisateurs: %v", err)
	}

	sessionUserB := &auth.SessionUser{
		UserID:         userBID,
		OrganizationID: orgB,
		Role:           "directeur_technique",
		Email:          "dt@beta.dz",
	}

	// 3. Produit & Dossier appartenant au Laboratoire Alpha
	prodAID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, dci, dosage, pharmaceutical_form, manufacturer_name)
		VALUES ($1, $2, 'medicament', 'Amoxicilline Alpha', 'Amoxicilline', '500 mg', 'Gélule', 'Alpha Lab Algiers');
	`, prodAID, orgA)
	if err != nil {
		t.Fatalf("échec insertion produit: %v", err)
	}

	medAID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO medicines (
			id, organization_id, product_id, origin_type, brand_name_fr,
			dci_fr, dosage_fr, pharmaceutical_form_fr, administration_route_fr,
			atc_code, economic_file_type
		) VALUES (
			$1, $2, $3, 'generique_local', 'Amoxicilline Alpha 500mg',
			'Amoxicilline', '500 mg', 'Gélule', 'Orale',
			'J01CA04', 'pcsu'
		)
	`, medAID, orgA, prodAID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	dossierAID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, medicine_id, product_id, product_type, dossier_code, procedure_code, internal_reference, status)
		VALUES ($1, $2, $3, $4, 'medicament', 'DOS-ALPHA-01', 'NI', 'REF-ALPHA-01', 'en_cours');
	`, dossierAID, orgA, medAID, prodAID)
	if err != nil {
		t.Fatalf("échec insertion dossier: %v", err)
	}

	// 4. Checklist Item appartenant au Laboratoire Alpha
	itemAID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossier_checklist_items (
			id, dossier_id, module_number, sub_module, item_code,
			title_fr, requirement_level, is_mandatory, status, workflow_status, current_revision
		) VALUES (
			$1, $2, 1, '1.P', '1.P.1',
			'Résumé des Caractéristiques du Produit (RCP)', 'E', TRUE, 'manquant', 'draft_pending', 'v0.1'
		);
	`, itemAID, dossierAID)
	if err != nil {
		t.Fatalf("échec insertion item: %v", err)
	}

	// 5. Document GED & Version appartenant au Laboratoire Alpha
	docAID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO ged_documents (id, organization_id, taxonomy_path, title_fr, document_type, status)
		VALUES ($1, $2, 'm1_admin/1_0_lettre_demande', 'Lettre Demande Alpha', 'formulaire_anpp', 'brouillon');
	`, docAID, orgA)
	if err != nil {
		t.Fatalf("échec insertion document GED: %v", err)
	}

	versionAID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO ged_versions (
			id, document_id, version_label, is_major_version, revision_summary, file_sha256,
			file_size_bytes, mime_type, original_filename, storage_path, author_id
		) VALUES (
			$1, $2, 'v1.0', TRUE, 'Version initiale', 'sha256-alpha-secret-data',
			1024, 'application/pdf', 'lettre_alpha.pdf', 'storage/alpha/lettre.pdf', $3
		);
	`, versionAID, docAID, userAID)
	if err != nil {
		t.Fatalf("échec insertion version GED: %v", err)
	}

	// Repositories & Handlers
	wfRepo := repository.NewWorkflowRepository(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	pharmaRepo := repository.NewPharmaRepository(db.Pool)
	gedRepo := repository.NewGEDRepository(db.Pool)
	auditService := audit.NewService(db.Pool)

	wfHandler := NewWorkflowHandler(wfRepo, dossierRepo, auditService, db.Pool)
	dossierHandler := NewDossierHandler(dossierRepo, pharmaRepo, gedRepo, auditService, db.Pool)
	gedHandler := NewGEDHandler(gedRepo, nil, auditService, db.Pool)

	// Router configuré pour simuler l'utilisateur du Laboratoire Beta (Tenant B)
	routerBeta := chi.NewRouter()
	routerBeta.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), UserContextKey, sessionUserB))
			next.ServeHTTP(w, r)
		})
	})

	// Enregistrement des routes testées
	routerBeta.Post("/api/v1/dossiers/{id}/items/{itemId}/checkout", wfHandler.CheckoutSection)
	routerBeta.Post("/api/v1/dossiers/{id}/items/{itemId}/checkin", wfHandler.CheckinSection)
	routerBeta.Post("/api/v1/dossiers/{id}/items/{itemId}/assign", wfHandler.AssignSection)
	routerBeta.Post("/api/v1/dossiers/{id}/items/{itemId}/submit-review", wfHandler.SubmitReview)
	routerBeta.Post("/api/v1/dossiers/{id}/items/{itemId}/review", wfHandler.ReviewDecision)
	routerBeta.Post("/api/v1/dossiers/{id}/items/{itemId}/approve", wfHandler.ApproveSection)
	routerBeta.Get("/api/v1/dossiers/{id}/items/{itemId}/gating", wfHandler.GetItemGating)
	routerBeta.Get("/api/v1/dossiers/{id}/items/{itemId}/reviews", wfHandler.GetItemReviews)

	routerBeta.Get("/api/v1/dossiers/{id}/members", dossierHandler.ListMembers)
	routerBeta.Post("/api/v1/dossiers/{id}/members", dossierHandler.AssignMember)
	routerBeta.Delete("/api/v1/dossiers/{id}/members/{userId}", dossierHandler.RemoveMember)

	routerBeta.Get("/api/v1/ged/documents/{id}/download", gedHandler.DownloadVersion)
	routerBeta.Post("/api/v1/ged/documents/{id}/sign-dt", gedHandler.SignDocumentDT)

	// =========================================================================
	// TEST 1 : Tentatives de piratage Workflow Cross-Tenant (Beta -> Alpha)
	// =========================================================================

	// 1.1 Checkout cross-tenant -> 404
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierAID.String()+"/items/"+itemAID.String()+"/checkout", bytes.NewReader([]byte(`{"reason":"attaque"}`)))
	routerBeta.ServeHTTP(w1, r1)
	if w1.Code != http.StatusNotFound {
		t.Fatalf("Checkout cross-tenant : attendu HTTP 404, obtenu %d: %s", w1.Code, w1.Body.String())
	}

	// 1.2 Checkin cross-tenant -> 404
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierAID.String()+"/items/"+itemAID.String()+"/checkin", bytes.NewReader([]byte(`{"comment":"attaque"}`)))
	routerBeta.ServeHTTP(w2, r2)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("Checkin cross-tenant : attendu HTTP 404, obtenu %d: %s", w2.Code, w2.Body.String())
	}

	// 1.3 Assignation cross-tenant -> 404
	w3 := httptest.NewRecorder()
	r3 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierAID.String()+"/items/"+itemAID.String()+"/assign", bytes.NewReader([]byte(`{"assigned_department":"RA"}`)))
	routerBeta.ServeHTTP(w3, r3)
	if w3.Code != http.StatusNotFound {
		t.Fatalf("Assign cross-tenant : attendu HTTP 404, obtenu %d: %s", w3.Code, w3.Body.String())
	}

	// 1.4 Soumission en revue cross-tenant -> 404
	w4 := httptest.NewRecorder()
	r4 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierAID.String()+"/items/"+itemAID.String()+"/submit-review", bytes.NewReader([]byte(`{"notes":"attaque"}`)))
	routerBeta.ServeHTTP(w4, r4)
	if w4.Code != http.StatusNotFound {
		t.Fatalf("Submit-review cross-tenant : attendu HTTP 404, obtenu %d: %s", w4.Code, w4.Body.String())
	}

	// 1.5 Approbation cross-tenant -> 404
	w5 := httptest.NewRecorder()
	r5 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierAID.String()+"/items/"+itemAID.String()+"/approve", bytes.NewReader([]byte(`{"solemn_declaration_accepted":true}`)))
	routerBeta.ServeHTTP(w5, r5)
	if w5.Code != http.StatusNotFound {
		t.Fatalf("Approve cross-tenant : attendu HTTP 404, obtenu %d: %s", w5.Code, w5.Body.String())
	}

	// 1.6 Inspection Gating cross-tenant -> 404
	w6 := httptest.NewRecorder()
	r6 := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierAID.String()+"/items/"+itemAID.String()+"/gating", nil)
	routerBeta.ServeHTTP(w6, r6)
	if w6.Code != http.StatusNotFound {
		t.Fatalf("Gating cross-tenant : attendu HTTP 404, obtenu %d: %s", w6.Code, w6.Body.String())
	}

	// 1.7 Consultation des revues qualité confidentielles cross-tenant -> 404
	w7 := httptest.NewRecorder()
	r7 := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierAID.String()+"/items/"+itemAID.String()+"/reviews", nil)
	routerBeta.ServeHTTP(w7, r7)
	if w7.Code != http.StatusNotFound {
		t.Fatalf("Reviews cross-tenant : attendu HTTP 404, obtenu %d: %s", w7.Code, w7.Body.String())
	}

	// =========================================================================
	// TEST 2 : Tentatives de violation Membres & Habilitations (Beta -> Alpha)
	// =========================================================================

	// 2.1 Listage des membres d'un dossier Alpha par Beta -> 404
	wM1 := httptest.NewRecorder()
	rM1 := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierAID.String()+"/members", nil)
	routerBeta.ServeHTTP(wM1, rM1)
	if wM1.Code != http.StatusNotFound {
		t.Fatalf("ListMembers cross-tenant : attendu HTTP 404, obtenu %d: %s", wM1.Code, wM1.Body.String())
	}

	// 2.2 Affectation d'un membre sur dossier Alpha par Beta -> 404
	bodyAssign, _ := json.Marshal(map[string]any{
		"user_id":                    userBID,
		"role_in_dossier":            "espion",
		"can_view_financials":        true,
		"can_view_industrial_secret": true,
	})
	wM2 := httptest.NewRecorder()
	rM2 := httptest.NewRequest("POST", "/api/v1/dossiers/"+dossierAID.String()+"/members", bytes.NewReader(bodyAssign))
	routerBeta.ServeHTTP(wM2, rM2)
	if wM2.Code != http.StatusNotFound {
		t.Fatalf("AssignMember cross-tenant : attendu HTTP 404, obtenu %d: %s", wM2.Code, wM2.Body.String())
	}

	// 2.3 Révocation d'un membre sur dossier Alpha par Beta -> 404
	wM3 := httptest.NewRecorder()
	rM3 := httptest.NewRequest("DELETE", "/api/v1/dossiers/"+dossierAID.String()+"/members/"+userAID.String(), nil)
	routerBeta.ServeHTTP(wM3, rM3)
	if wM3.Code != http.StatusNotFound {
		t.Fatalf("RemoveMember cross-tenant : attendu HTTP 404, obtenu %d: %s", wM3.Code, wM3.Body.String())
	}

	// =========================================================================
	// TEST 3 : Tentatives d'accès aux versions GED (Beta -> Alpha)
	// =========================================================================

	// 3.1 Téléchargement d'une version spécifique d'Alpha par Beta -> 404
	wG1 := httptest.NewRecorder()
	rG1 := httptest.NewRequest("GET", "/api/v1/ged/documents/"+docAID.String()+"/download?version_id="+versionAID.String(), nil)
	routerBeta.ServeHTTP(wG1, rG1)
	if wG1.Code != http.StatusNotFound {
		t.Fatalf("DownloadDocument version cross-tenant : attendu HTTP 404, obtenu %d: %s", wG1.Code, wG1.Body.String())
	}

	// 3.2 Signature d'une version spécifique d'Alpha par Beta -> 404
	bodySign, _ := json.Marshal(map[string]any{
		"version_id":            versionAID,
		"signature_declaration": "Tentative de visa illégitime",
		"password":              "PassSec123!#",
	})
	wG2 := httptest.NewRecorder()
	rG2 := httptest.NewRequest("POST", "/api/v1/ged/documents/"+docAID.String()+"/sign-dt", bytes.NewReader(bodySign))
	routerBeta.ServeHTTP(wG2, rG2)
	if wG2.Code != http.StatusNotFound {
		t.Fatalf("SignDocument cross-tenant : attendu HTTP 404, obtenu %d: %s", wG2.Code, wG2.Body.String())
	}

	// =========================================================================
	// TEST 4 : Vérification directe au niveau Repository
	// =========================================================================

	// 4.1 Workflow Repository : GetChecklistItemByID avec orgB
	_, errWf := wfRepo.GetChecklistItemByID(ctx, orgB, itemAID)
	if !errors.Is(errWf, repository.ErrWorkflowItemNotFound) {
		t.Fatalf("wfRepo.GetChecklistItemByID avec autre org : attendu ErrWorkflowItemNotFound, obtenu %v", errWf)
	}

	// 4.2 GED Repository : GetVersionByID avec orgB
	_, errGed := gedRepo.GetVersionByID(ctx, orgB, versionAID)
	if !errors.Is(errGed, repository.ErrNotFound) {
		t.Fatalf("gedRepo.GetVersionByID avec autre org : attendu ErrNotFound, obtenu %v", errGed)
	}

	// 4.3 Dossier Repository : ListDossierMembers avec orgB
	_, errDossierMembers := dossierRepo.ListDossierMembers(ctx, orgB, dossierAID)
	if !errors.Is(errDossierMembers, repository.ErrNotFound) {
		t.Fatalf("dossierRepo.ListDossierMembers avec autre org : attendu ErrNotFound, obtenu %v", errDossierMembers)
	}
}
