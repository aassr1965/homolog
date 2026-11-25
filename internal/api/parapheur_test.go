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

func TestParapheurPendingItemsAndBatchApprove(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-parapheur-" + testOrgID.String()[:8]

	otherOrgID := uuid.New()
	otherSlug := "org-other-parapheur-" + otherOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM dossier_item_reviews WHERE dossier_item_id IN (
				SELECT id FROM dossier_checklist_items WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id IN ($1, $2))
			);
			DELETE FROM dossier_checklist_items WHERE dossier_id IN (SELECT id FROM dossiers WHERE organization_id IN ($1, $2));
			DELETE FROM dossiers WHERE organization_id IN ($1, $2);
			DELETE FROM medicines WHERE organization_id IN ($1, $2);
			DELETE FROM products WHERE organization_id IN ($1, $2);
			DELETE FROM users WHERE organization_id IN ($1, $2);
			DELETE FROM organizations WHERE id IN ($1, $2);
		`, testOrgID, otherOrgID)
	}()

	// 1. Création des 2 organisations pour tester l'isolation multi-tenant
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES 
		($1, $2, 'Biocare Algérie Parapheur Test', 'MIPH-2026-PARAPH', 'Dr. Yacine Benali'),
		($3, $4, 'Autre Labo Parapheur Test', 'MIPH-2026-OTHER', 'Dr. Samir Chérif');
	`, testOrgID, orgSlug, otherOrgID, otherSlug)
	if err != nil {
		t.Fatalf("échec création organisations: %v", err)
	}

	// 2. Utilisateurs : Maker, Checker, PDT
	makerID := uuid.New()
	checkerID := uuid.New()
	pdtID := uuid.New()
	pwdHash, _ := auth.HashPassword("Pass123!#", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES 
		($1, $4, 'maker.para@biocare.dz', $5, 'Karim', 'Maker', 'contributeur_rd_prod', TRUE),
		($2, $4, 'checker.para@biocare.dz', $5, 'Sarah', 'Checker', 'contributeur_qualite_bpf', TRUE),
		($3, $4, 'pdt.para@biocare.dz', $5, 'Yacine', 'Benali', 'directeur_technique', TRUE);
	`, makerID, checkerID, pdtID, testOrgID, pwdHash)
	if err != nil {
		t.Fatalf("échec création utilisateurs: %v", err)
	}

	// 3. Produits & Dossiers
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
	targetDeposit := time.Now().AddDate(0, 3, 0) // J+90
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, product_id, medicine_id, dossier_code, procedure_code, internal_reference, status, target_deposit_date)
		VALUES ($1, $2, $3, $4, 'DOS-PARA-2026', 'NF', 'REF-PARA-001', 'en_cours', $5)
	`, dossierID, testOrgID, prodID, medID, targetDeposit)
	if err != nil {
		t.Fatalf("échec création dossier: %v", err)
	}

	// Dossier de l'autre organisation
	otherDossierID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, dossier_code, procedure_code, internal_reference, status)
		VALUES ($1, $2, 'DOS-OTHER-2026', 'NI', 'REF-OTHER-001', 'en_cours')
	`, otherDossierID, otherOrgID)
	if err != nil {
		t.Fatalf("échec création autre dossier: %v", err)
	}

	// 4. Création des items dans le dossier testOrgID
	// Item 1 : 1.P.8 (pending_approval, maker = makerID, checker = checkerID) -> DOIT être dans le parapheur et éligible au visa
	item1ID := uuid.New()
	// Item 2 : 1.P.9 (pending_approval, maker = makerID, checker = checkerID) -> DOIT être dans le parapheur et éligible au visa
	item2ID := uuid.New()
	// Item 3 : 1.H.10 (pending_approval mais Maker == PDT -> SoD Violée !)
	item3ID := uuid.New()
	// Item 4 : 3.2.P.5 (drafting -> Ne doit PAS être dans le parapheur)
	item4ID := uuid.New()
	// Item 5 (autre org) : pending_approval -> Ne doit JAMAIS être retourné (Tenant Isolation)
	itemOtherID := uuid.New()

	checkerReviewTime := time.Now().Add(-1 * time.Hour)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossier_checklist_items (
			id, dossier_id, module_number, sub_module, item_code,
			title_fr, requirement_level, is_mandatory, status,
			workflow_status, current_revision, assigned_department,
			maker_user_id, checker_user_id, checker_reviewed_at, checker_decision, checker_notes
		) VALUES 
		($1, $6, 1, '1.P', '1.P.8', 'Engagement de validation industrielle 3 premiers lots', 'E', TRUE, 'valide_qualite', 'pending_approval', 'v0.2', 'RA', $7, $8, $9, 'approved', 'Conforme aux normes ANPP'),
		($2, $6, 1, '1.P', '1.P.9', 'Engagement de pharmacovigilance et stabilité Zone IVb', 'E', TRUE, 'valide_qualite', 'pending_approval', 'v0.2', 'RA', $7, $8, $9, 'approved', 'Revue AQ favorable'),
		($3, $6, 1, '1.H', '1.H.10', 'Déclaration solennelle de conformité', 'E', TRUE, 'valide_qualite', 'pending_approval', 'v0.2', 'DT', $10, $8, $9, 'approved', 'Auto-approbation SoD test'),
		($4, $6, 3, '3.2.P', '3.2.P.5', 'Contrôle du produit fini', 'E', TRUE, 'manquant', 'drafting', 'v0.1', 'QC', $7, NULL, NULL, NULL, NULL),
		($5, $11, 1, '1.A', '1.A.1', 'Formulaire autre organisation', 'E', TRUE, 'valide_qualite', 'pending_approval', 'v0.2', 'RA', NULL, NULL, NULL, NULL, NULL);
	`, item1ID, item2ID, item3ID, item4ID, itemOtherID, dossierID, makerID, checkerID, checkerReviewTime, pdtID, otherDossierID)
	if err != nil {
		t.Fatalf("échec insertion checklist items: %v", err)
	}

	// 5. Initialisation des composants API
	auditSvc := audit.NewService(db.Pool)
	wfRepo := repository.NewWorkflowRepository(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	handler := NewWorkflowHandler(wfRepo, dossierRepo, auditSvc, db.Pool)

	// Router de test avec middleware de session
	pdtSession := &auth.SessionUser{
		UserID:         pdtID,
		OrganizationID: testOrgID,
		Email:          "pdt.para@biocare.dz",
		Role:           "directeur_technique",
		Permissions:    []string{string(domain.PermDocApprovePDT), string(domain.PermDossierView)},
	}

	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), UserContextKey, pdtSession)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	router.Get("/api/v1/parapheur/pending", handler.GetParapheurPendingItems)
	router.Post("/api/v1/parapheur/batch-approve", handler.BatchApproveSections)

	// -------------------------------------------------------------
	// TEST A : Consultation du Parapheur Unifié (GET /api/v1/parapheur/pending)
	// -------------------------------------------------------------
	t.Run("Consultation du Parapheur Unifié", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/parapheur/pending", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Attendu HTTP 200, reçu %d: %s", rr.Code, rr.Body.String())
		}

		var resp struct {
			Count int                            `json:"count"`
			Items []*domain.ParapheurPendingItem `json:"items"`
		}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("Décodage JSON impossible: %v", err)
		}

		// Attendu : exactement 3 items (item1, item2, item3) de testOrgID
		// L'item 4 est drafting, et l'item 5 appartient à otherOrgID !
		if resp.Count != 3 {
			t.Fatalf("Nombre d'items attendu dans le parapheur: 3, reçu: %d", resp.Count)
		}

		for _, it := range resp.Items {
			if it.WorkflowStatus != "pending_approval" {
				t.Errorf("Item %s a le statut %s au lieu de pending_approval", it.ItemCode, it.WorkflowStatus)
			}
			if it.DossierCode != "DOS-PARA-2026" {
				t.Errorf("Isolation tenant violée : dossier code reçu %s", it.DossierCode)
			}
			if it.ProductName == "" {
				t.Errorf("ProductName attendu non vide pour %s", it.ItemCode)
			}
			if it.CheckerNotes == nil || *it.CheckerNotes == "" {
				t.Errorf("Notes du Checker attendues non vides pour %s", it.ItemCode)
			}
		}
	})

	// -------------------------------------------------------------
	// TEST B : Signature par Lot Refusée sans Déclaration Solennelle
	// -------------------------------------------------------------
	t.Run("Rejet Signature Lot sans Déclaration Solennelle", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]any{
			"item_ids":                    []string{item1ID.String(), item2ID.String()},
			"password":                    "Pass123!#",
			"solemn_declaration_accepted": false,
		})
		req := httptest.NewRequest("POST", "/api/v1/parapheur/batch-approve", bytes.NewReader(payload))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Attendu HTTP 400 pour omission déclaration solennelle, reçu %d", rr.Code)
		}
	})

	// -------------------------------------------------------------
	// TEST C : Signature par Lot avec SoD (Item 1 & 2 OK, Item 3 Rejeté car Maker == PDT)
	// -------------------------------------------------------------
	t.Run("Exécution Signature par Lot avec Garde-Fou SoD", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]any{
			"item_ids":                    []string{item1ID.String(), item2ID.String(), item3ID.String(), itemOtherID.String()},
			"password":                    "Pass123!#",
			"solemn_declaration_accepted": true,
		})
		req := httptest.NewRequest("POST", "/api/v1/parapheur/batch-approve", bytes.NewReader(payload))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Attendu HTTP 200, reçu %d: %s", rr.Code, rr.Body.String())
		}

		var batchResp domain.BatchApproveResponse
		if err := json.NewDecoder(rr.Body).Decode(&batchResp); err != nil {
			t.Fatalf("Décodage réponse batch impossible: %v", err)
		}

		if batchResp.TotalRequested != 4 {
			t.Fatalf("TotalRequested attendu 4, reçu: %d", batchResp.TotalRequested)
		}
		if batchResp.TotalApproved != 2 {
			t.Fatalf("TotalApproved attendu 2 (items 1 et 2), reçu: %d", batchResp.TotalApproved)
		}
		if batchResp.TotalFailed != 2 {
			t.Fatalf("TotalFailed attendu 2 (item 3 SoD et itemOther cross-tenant), reçu: %d", batchResp.TotalFailed)
		}

		// Vérifier le scellement en base des items 1 et 2
		for _, targetID := range []uuid.UUID{item1ID, item2ID} {
			var st, rev, seal string
			err := db.Pool.QueryRow(ctx, `SELECT workflow_status, current_revision, approver_signature_seal FROM dossier_checklist_items WHERE id = $1`, targetID).Scan(&st, &rev, &seal)
			if err != nil {
				t.Fatalf("Erreur lecture item %s: %v", targetID, err)
			}
			if st != "approved_sealed" {
				t.Errorf("Statut attendu approved_sealed, reçu: %s", st)
			}
			if rev != "v1.0" {
				t.Errorf("Révision attendue v1.0, reçu: %s", rev)
			}
			if len(seal) < 32 {
				t.Errorf("Sceau SHA-256 attendu, reçu: %s", seal)
			}
		}

		// Vérifier que l'item 3 (Maker == PDT) n'a PAS été approuvé (SoD préservée)
		var st3 string
		_ = db.Pool.QueryRow(ctx, `SELECT workflow_status FROM dossier_checklist_items WHERE id = $1`, item3ID).Scan(&st3)
		if st3 != "pending_approval" {
			t.Errorf("L'item 3 devrait rester pending_approval en raison de la violation SoD, reçu: %s", st3)
		}
	})

	// -------------------------------------------------------------
	// TEST D : Parapheur Mis à Jour (Ne reste que l'item 3 non approuvé)
	// -------------------------------------------------------------
	t.Run("Mise à Jour Dynamique du Parapheur après Signature", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/parapheur/pending", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		var resp struct {
			Count int                            `json:"count"`
			Items []*domain.ParapheurPendingItem `json:"items"`
		}
		_ = json.NewDecoder(rr.Body).Decode(&resp)

		if resp.Count != 1 {
			t.Fatalf("Après signature de 2 items, attendu 1 item restant dans le parapheur, reçu: %d", resp.Count)
		}
		if resp.Items[0].ItemID != item3ID {
			t.Errorf("L'item restant attendu est item3, reçu: %s", resp.Items[0].ItemCode)
		}
	})
}

func TestPharmacienAssistantAndDTDelegations(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-deleg-" + testOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), `
			DELETE FROM dt_delegations WHERE organization_id = $1;
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
		VALUES ($1, $2, 'Biocare Algérie Intérim Test', 'MIPH-2026-INT', 'Dr. Yacine Benali');
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation: %v", err)
	}

	// 2. Utilisateurs : PDT et Pharmacien Assistant
	pdtID := uuid.New()
	assistantID := uuid.New()
	makerID := uuid.New()
	checkerID := uuid.New()
	pwdHash, _ := auth.HashPassword("Pass123!#", nil)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES 
		($1, $5, 'pdt.interim@test.dz', $6, 'Yacine', 'Benali', 'directeur_technique', TRUE),
		($2, $5, 'assistant.interim@test.dz', $6, 'Lynda', 'Amrani', 'pharmacien_assistant', TRUE),
		($3, $5, 'maker.interim@test.dz', $6, 'Karim', 'Maker', 'contributeur_rd_prod', TRUE),
		($4, $5, 'checker.interim@test.dz', $6, 'Sarah', 'Checker', 'contributeur_qualite_bpf', TRUE);
	`, pdtID, assistantID, makerID, checkerID, testOrgID, pwdHash)
	if err != nil {
		t.Fatalf("échec création utilisateurs: %v", err)
	}

	// 3. Produit & Dossier avec 1 pièce prête pour visa
	prodID := uuid.New()
	medID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, dci, dosage, pharmaceutical_form, manufacturer_name)
		VALUES ($1, $2, 'medicament', 'Amoxicilline Intérim', 'Amoxicilline', '500 mg', 'Gélule', 'Biocare Lab')
	`, prodID, testOrgID)
	if err != nil {
		t.Fatalf("échec création produit: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO medicines (
			id, organization_id, product_id, origin_type, brand_name_fr,
			dci_fr, dosage_fr, pharmaceutical_form_fr, administration_route_fr,
			atc_code, economic_file_type
		) VALUES (
			$1, $2, $3, 'generique_local', 'Amoxicilline Intérim',
			'Amoxicilline', '500 mg', 'Gélule', 'Orale',
			'J01CA04', 'pcsu'
		)
	`, medID, testOrgID, prodID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	dossierID := uuid.New()
	targetDeposit := time.Now().AddDate(0, 3, 0)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, product_id, medicine_id, dossier_code, procedure_code, internal_reference, status, target_deposit_date)
		VALUES ($1, $2, $3, $4, 'DOS-2026-INT-01', 'NF', 'REF-INT-001', 'en_cours', $5);
	`, dossierID, testOrgID, prodID, medID, targetDeposit)
	if err != nil {
		t.Fatalf("échec création dossier: %v", err)
	}

	itemID := uuid.New()
	subTime := time.Now().Add(-2 * time.Hour)
	revTime := time.Now().Add(-1 * time.Hour)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossier_checklist_items (
			id, dossier_id, module_number, sub_module, item_code, title_fr, requirement_level,
			status, workflow_status, current_revision, assigned_department, is_mandatory, sort_order,
			maker_user_id, maker_submitted_at, checker_user_id, checker_reviewed_at, checker_decision, checker_notes
		)
		VALUES (
			$1, $2, 1, '1.A', '1.A.1', 'Lettre d’engagement de validation', 'E',
			'valide_qualite', 'pending_approval', 'v0.2', 'RA', TRUE, 10,
			$3, $4, $5, $6, 'approved', 'Revue qualité favorable, conforme BPF.'
		);
	`, itemID, dossierID, makerID, subTime, checkerID, revTime)
	if err != nil {
		t.Fatalf("échec création item: %v", err)
	}

	auditSvc := audit.NewService(db.Pool)
	wfRepo := repository.NewWorkflowRepository(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	handler := NewWorkflowHandler(wfRepo, dossierRepo, auditSvc, db.Pool)

	// Routeur de test
	var currentUser *auth.SessionUser
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if currentUser != nil {
				ctx := context.WithValue(r.Context(), UserContextKey, currentUser)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})
	router.Get("/api/v1/parapheur/pending", handler.GetParapheurPendingItems)
	router.Post("/api/v1/parapheur/batch-approve", handler.BatchApproveSections)
	router.Get("/api/v1/parapheur/delegations", handler.ListDTDelegations)
	router.Get("/api/v1/parapheur/delegation/active", handler.GetActiveDTDelegation)
	router.Post("/api/v1/parapheur/delegations", handler.CreateDTDelegation)
	router.Post("/api/v1/parapheur/delegations/{id}/revoke", handler.RevokeDTDelegation)
	router.Get("/api/v1/parapheur/assistants", handler.GetAvailableAssistants)

	// TEST 1 : Pharmacien Assistant SANS délégation active -> Rejet 403
	t.Run("Assistant_Sans_Delegation_Refus", func(t *testing.T) {
		currentUser = &auth.SessionUser{
			UserID:         assistantID,
			OrganizationID: testOrgID,
			Email:          "assistant.interim@test.dz",
			Role:           "pharmacien_assistant",
		}

		req := httptest.NewRequest("GET", "/api/v1/parapheur/pending", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("Attendu 403 Forbidden pour Assistant sans mandat d'intérim, reçu: %d", rr.Code)
		}
	})

	// TEST 2 : Création d'une délégation dépassant la limite légale de 3 mois (Décret 21-82 Art. 24) -> Rejet 400
	t.Run("Creation_Delegation_Depassement_3_Mois_Refus", func(t *testing.T) {
		currentUser = &auth.SessionUser{
			UserID:         pdtID,
			OrganizationID: testOrgID,
			Email:          "pdt.interim@test.dz",
			Role:           "directeur_technique",
		}

		// Durée de 120 jours (> 92 jours)
		start := time.Now()
		end := start.Add(120 * 24 * time.Hour)

		body, _ := json.Marshal(domain.CreateDTDelegationRequest{
			DelegateID:          assistantID,
			MIPHNotificationRef: "NOTIF-MIPH-INT-2026/999",
			Reason:              "Mission longue",
			StartDate:           start,
			EndDate:             end,
		})

		req := httptest.NewRequest("POST", "/api/v1/parapheur/delegations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Attendu 400 Bad Request pour mandat > 3 mois, reçu: %d", rr.Code)
		}
	})

	// TEST 2b : Tentative d'auto-désignation du DT titulaire -> Rejet 400
	t.Run("Creation_Delegation_Auto_Designation_Refus", func(t *testing.T) {
		currentUser = &auth.SessionUser{
			UserID:         pdtID,
			OrganizationID: testOrgID,
			Email:          "pdt.interim@test.dz",
			Role:           "directeur_technique",
		}

		start := time.Now()
		end := start.Add(30 * 24 * time.Hour)

		body, _ := json.Marshal(domain.CreateDTDelegationRequest{
			DelegateID:          pdtID, // Auto-désignation
			MIPHNotificationRef: "NOTIF-MIPH-INT-2026/001",
			Reason:              "Auto-délégation impossible",
			StartDate:           start,
			EndDate:             end,
		})

		req := httptest.NewRequest("POST", "/api/v1/parapheur/delegations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Attendu 400 Bad Request pour auto-désignation, reçu: %d", rr.Code)
		}
	})

	// TEST 2c : Tentative de désignation d'un rôle NON Pharmacien Assistant -> Rejet 400
	t.Run("Creation_Delegation_Non_Assistant_Refus", func(t *testing.T) {
		currentUser = &auth.SessionUser{
			UserID:         pdtID,
			OrganizationID: testOrgID,
			Email:          "pdt.interim@test.dz",
			Role:           "directeur_technique",
		}

		start := time.Now()
		end := start.Add(30 * 24 * time.Hour)

		// Tentative avec le contributeur qualité BPF (checkerID)
		body, _ := json.Marshal(domain.CreateDTDelegationRequest{
			DelegateID:          checkerID,
			MIPHNotificationRef: "NOTIF-MIPH-INT-2026/002",
			Reason:              "Délégation invalide",
			StartDate:           start,
			EndDate:             end,
		})

		req := httptest.NewRequest("POST", "/api/v1/parapheur/delegations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("Attendu 400 Bad Request pour délégataire non-assistant, reçu: %d", rr.Code)
		}
	})

	// TEST 2d : Endpoint Assistants éligibles -> Filtrage strict rôle Pharmacien Assistant & exclusion du DT
	t.Run("GetAvailableAssistants_Uniquement_Pharmacien_Assistant", func(t *testing.T) {
		currentUser = &auth.SessionUser{
			UserID:         pdtID,
			OrganizationID: testOrgID,
			Email:          "pdt.interim@test.dz",
			Role:           "directeur_technique",
		}

		req := httptest.NewRequest("GET", "/api/v1/parapheur/assistants", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Attendu 200 OK pour liste assistants, reçu: %d", rr.Code)
		}

		type AssistantItem struct {
			ID        uuid.UUID `json:"id"`
			Email     string    `json:"email"`
			FirstName string    `json:"first_name"`
			LastName  string    `json:"last_name"`
			Role      string    `json:"role"`
		}
		var list []AssistantItem
		if err := json.NewDecoder(rr.Body).Decode(&list); err != nil {
			t.Fatalf("Erreur décodage assistants: %v", err)
		}

		if len(list) != 1 {
			t.Fatalf("Attendu exactement 1 assistant (Lynda Amrani), reçu: %d", len(list))
		}
		if list[0].ID != assistantID {
			t.Errorf("Attendu assistantID %s, reçu: %s", assistantID, list[0].ID)
		}
		if list[0].Role != "pharmacien_assistant" {
			t.Errorf("Attendu rôle pharmacien_assistant, reçu: %s", list[0].Role)
		}
	})

	// TEST 3 : Création d'une délégation légale valide (45 jours, notification MIPH) -> Succès 201
	var delegationID uuid.UUID
	t.Run("Creation_Delegation_Valide_Succes", func(t *testing.T) {
		currentUser = &auth.SessionUser{
			UserID:         pdtID,
			OrganizationID: testOrgID,
			Email:          "pdt.interim@test.dz",
			Role:           "directeur_technique",
		}

		start := time.Now().Add(-1 * time.Hour)
		end := start.Add(45 * 24 * time.Hour)

		body, _ := json.Marshal(domain.CreateDTDelegationRequest{
			DelegateID:          assistantID,
			MIPHNotificationRef: "NOTIF-MIPH-INT-2026/088",
			Reason:              "Mission d'audit BPF internationale",
			StartDate:           start,
			EndDate:             end,
		})

		req := httptest.NewRequest("POST", "/api/v1/parapheur/delegations", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("Attendu 201 Created pour création délégation, reçu: %d, body: %s", rr.Code, rr.Body.String())
		}

		var created domain.DTDelegation
		_ = json.NewDecoder(rr.Body).Decode(&created)
		if created.ID == uuid.Nil {
			t.Fatal("L'ID de la délégation créée ne doit pas être nil")
		}
		delegationID = created.ID
	})

	// TEST 4 : Pharmacien Assistant AVEC mandat d'intérim actif -> Consultation & Visa par lot autorisés
	t.Run("Assistant_Avec_Mandat_Actif_Autorise", func(t *testing.T) {
		currentUser = &auth.SessionUser{
			UserID:         assistantID,
			OrganizationID: testOrgID,
			Email:          "assistant.interim@test.dz",
			Role:           "pharmacien_assistant",
		}

		// 4a. Consultation du parapheur
		req := httptest.NewRequest("GET", "/api/v1/parapheur/pending", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Attendu 200 OK pour Assistant sous mandat d'intérim, reçu: %d", rr.Code)
		}

		// 4b. Signature par lot avec mention d'intérim
		batchReq, _ := json.Marshal(domain.BatchApproveRequest{
			ItemIDs:                   []uuid.UUID{itemID},
			SolemnDeclarationAccepted: true,
		})
		reqSign := httptest.NewRequest("POST", "/api/v1/parapheur/batch-approve", bytes.NewReader(batchReq))
		reqSign.Header.Set("Content-Type", "application/json")
		rrSign := httptest.NewRecorder()
		router.ServeHTTP(rrSign, reqSign)

		if rrSign.Code != http.StatusOK {
			t.Fatalf("Attendu 200 OK pour signature par lot assistant délégué, reçu: %d body: %s", rrSign.Code, rrSign.Body.String())
		}

		var signResp domain.BatchApproveResponse
		_ = json.NewDecoder(rrSign.Body).Decode(&signResp)
		if signResp.TotalApproved != 1 {
			errMsg := "aucun item"
			if len(signResp.Items) > 0 {
				errMsg = signResp.Items[0].ErrorMessage
			}
			t.Fatalf("Attendu 1 pièce approuvée, reçu: %d (erreur item: %s)", signResp.TotalApproved, errMsg)
		}

		// Vérifier la mention d'intérim dans dossier_item_reviews
		var reviewComment string
		err := db.Pool.QueryRow(ctx, `SELECT comment FROM dossier_item_reviews WHERE dossier_item_id = $1 ORDER BY created_at DESC LIMIT 1`, itemID).Scan(&reviewComment)
		if err != nil {
			t.Fatalf("Erreur lecture review audit: %v", err)
		}
		if !bytes.Contains([]byte(reviewComment), []byte("Pharmacien DT par intérim")) {
			t.Errorf("Le commentaire d'audit doit mentionner l'intérim, reçu: %s", reviewComment)
		}
	})

	// TEST 5 : Révocation du mandat d'intérim
	t.Run("Revocation_Delegation_Reprise_Fonctions", func(t *testing.T) {
		currentUser = &auth.SessionUser{
			UserID:         pdtID,
			OrganizationID: testOrgID,
			Email:          "pdt.interim@test.dz",
			Role:           "directeur_technique",
		}

		revokeReq, _ := json.Marshal(domain.RevokeDTDelegationRequest{
			Reason: "Reprise des fonctions",
		})
		req := httptest.NewRequest("POST", "/api/v1/parapheur/delegations/"+delegationID.String()+"/revoke", bytes.NewReader(revokeReq))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Attendu 200 OK pour révocation, reçu: %d", rr.Code)
		}

		// L'assistant ne doit plus avoir accès au parapheur
		currentUser = &auth.SessionUser{
			UserID:         assistantID,
			OrganizationID: testOrgID,
			Email:          "assistant.interim@test.dz",
			Role:           "pharmacien_assistant",
		}
		reqCheck := httptest.NewRequest("GET", "/api/v1/parapheur/pending", nil)
		rrCheck := httptest.NewRecorder()
		router.ServeHTTP(rrCheck, reqCheck)

		if rrCheck.Code != http.StatusForbidden {
			t.Fatalf("Après révocation, attendu 403 Forbidden pour Assistant, reçu: %d", rrCheck.Code)
		}
	})

	t.Run("AuditTrail_Assistant_Uniquement_Durant_Interim", func(t *testing.T) {
		sm := auth.NewSessionManager(db.Pool, 60, 15)
		rawToken, err := sm.CreateSession(ctx, assistantID, "127.0.0.1", "TestAgent")
		if err != nil {
			t.Fatalf("Erreur création session Assistant: %v", err)
		}

		// 1. Hors intérim (révoqué) : AUDIT_VIEW doit être refusé
		uNoInterim, err := sm.ValidateSession(ctx, rawToken)
		if err != nil {
			t.Fatalf("Erreur validation session: %v", err)
		}
		if uNoInterim.HasPermission(string(domain.PermAuditView)) {
			t.Fatalf("Hors période d'intérim, le Pharmacien Assistant ne doit pas avoir AUDIT_VIEW")
		}

		// 2. Durant un mandat d'intérim actif : AUDIT_VIEW doit être accordé
		delID := uuid.New()
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO dt_delegations (id, organization_id, delegator_id, delegate_id, miph_notification_ref, reason, start_date, end_date, is_active)
			VALUES ($1, $2, $3, $4, 'MIPH-TEST-AUDIT', 'Intérim test', NOW() - INTERVAL '1 hour', NOW() + INTERVAL '24 hours', TRUE)
		`, delID, testOrgID, pdtID, assistantID)
		if err != nil {
			t.Fatalf("Erreur insertion délégation test: %v", err)
		}

		uActiveInterim, err := sm.ValidateSession(ctx, rawToken)
		if err != nil {
			t.Fatalf("Erreur validation session pendant intérim: %v", err)
		}
		if !uActiveInterim.HasPermission(string(domain.PermAuditView)) {
			t.Fatalf("Durant un intérim actif, le Pharmacien Assistant doit avoir AUDIT_VIEW")
		}

		// 3. Après révocation du mandat : AUDIT_VIEW doit être immédiatement retiré
		_, err = db.Pool.Exec(ctx, `UPDATE dt_delegations SET is_active = FALSE, revoked_at = NOW() WHERE id = $1`, delID)
		if err != nil {
			t.Fatalf("Erreur révocation délégation test: %v", err)
		}

		uRevoked, err := sm.ValidateSession(ctx, rawToken)
		if err != nil {
			t.Fatalf("Erreur validation session après révocation: %v", err)
		}
		if uRevoked.HasPermission(string(domain.PermAuditView)) {
			t.Fatalf("Après révocation de l'intérim, le Pharmacien Assistant ne doit plus avoir AUDIT_VIEW")
		}
	})

	t.Run("GED_SignDocumentDT_Interim_Qualite_Production", func(t *testing.T) {
		gedRepo := repository.NewGEDRepository(db.Pool)
		gedH := NewGEDHandler(gedRepo, nil, audit.NewService(db.Pool), db.Pool)

		docRegID := uuid.New()
		docQualID := uuid.New()
		verRegID := uuid.New()
		verQualID := uuid.New()

		_, err := db.Pool.Exec(ctx, `
			INSERT INTO ged_documents (id, organization_id, taxonomy_path, title_fr, document_code, document_type, status, current_version, created_by)
			VALUES
			  ($1, $3, '01_Affaires_Reglementaires/01_01_Agrements_Etablissement', 'Note Affaires Réglementaires', 'DOC-REG-INT-01', 'document_administratif', 'en_approbation_dt', 'v1.0', $4),
			  ($2, $3, '02_Assurance_Qualite/02_01_SOP', 'SOP Assurance Qualité BPF', 'DOC-QA-INT-01', 'procedure_sop', 'en_approbation_dt', 'v1.0', $4)
		`, docRegID, docQualID, testOrgID, makerID)
		if err != nil {
			t.Fatalf("Erreur insertion documents GED test: %v", err)
		}

		_, err = db.Pool.Exec(ctx, `
			INSERT INTO ged_versions (id, document_id, version_label, revision_summary, file_sha256, file_size_bytes, mime_type, original_filename, storage_path, author_id)
			VALUES
			  ($1, $3, 'v1.0', 'Version initiale test', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', 1200, 'application/pdf', 'reg.pdf', '/tmp/reg.pdf', $5),
			  ($2, $4, 'v1.0', 'Version initiale test', 'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb', 1200, 'application/pdf', 'qa.pdf', '/tmp/qa.pdf', $5)
		`, verRegID, verQualID, docRegID, docQualID, makerID)
		if err != nil {
			t.Fatalf("Erreur insertion versions GED test: %v", err)
		}

		gedRouter := chi.NewRouter()
		gedRouter.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				u := &auth.SessionUser{
					UserID:         assistantID,
					OrganizationID: testOrgID,
					Email:          "assistant.interim@test.dz",
					Role:           "pharmacien_assistant",
					Permissions:    []string{string(domain.PermDocCheckReview)},
				}
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserContextKey, u)))
			})
		})
		gedRouter.With(RequireAnyPermission(nil, domain.PermDocApprovePDT, domain.PermDocCheckReview)).Post("/api/v1/ged/documents/{id}/sign-dt", gedH.SignDocumentDT)

		signBody, _ := json.Marshal(map[string]string{"password": "Pass123!#"})

		// 1. Hors intérim -> 403
		req1 := httptest.NewRequest("POST", "/api/v1/ged/documents/"+docQualID.String()+"/sign-dt", bytes.NewReader(signBody))
		rr1 := httptest.NewRecorder()
		gedRouter.ServeHTTP(rr1, req1)
		if rr1.Code != http.StatusForbidden {
			t.Fatalf("Hors intérim, attendu 403 Forbidden pour signature GED Qualité, reçu: %d (%s)", rr1.Code, rr1.Body.String())
		}

		// 2. Activer un mandat d'intérim
		delID := uuid.New()
		_, err = db.Pool.Exec(ctx, `
			INSERT INTO dt_delegations (id, organization_id, delegator_id, delegate_id, miph_notification_ref, reason, start_date, end_date, is_active)
			VALUES ($1, $2, $3, $4, 'MIPH-INT-GED-01', 'Intérim Qualité/Prod', NOW() - INTERVAL '1 hour', NOW() + INTERVAL '24 hours', TRUE)
		`, delID, testOrgID, pdtID, assistantID)
		if err != nil {
			t.Fatalf("Erreur insertion délégation test: %v", err)
		}

		// 3. Durant intérim sur document Affaires Réglementaires (01_) -> 403
		req2 := httptest.NewRequest("POST", "/api/v1/ged/documents/"+docRegID.String()+"/sign-dt", bytes.NewReader(signBody))
		rr2 := httptest.NewRecorder()
		gedRouter.ServeHTTP(rr2, req2)
		if rr2.Code != http.StatusForbidden {
			t.Fatalf("Durant intérim sur 01_Affaires_Reglementaires, attendu 403 Forbidden, reçu: %d (%s)", rr2.Code, rr2.Body.String())
		}

		// 4. Durant intérim sur document Assurance Qualité (02_) -> 200 OK
		req3 := httptest.NewRequest("POST", "/api/v1/ged/documents/"+docQualID.String()+"/sign-dt", bytes.NewReader(signBody))
		rr3 := httptest.NewRecorder()
		gedRouter.ServeHTTP(rr3, req3)
		if rr3.Code != http.StatusOK {
			t.Fatalf("Durant intérim sur 02_Assurance_Qualite, attendu 200 OK, reçu: %d (%s)", rr3.Code, rr3.Body.String())
		}
	})

	t.Run("GED_SignDocumentDT_Permanent_Delegation_Hors_Interim", func(t *testing.T) {
		// Révoquer tout mandat d'intérim actif pour tester le mode HORS INTÉRIM
		_, _ = db.Pool.Exec(ctx, `UPDATE dt_delegations SET is_active = FALSE, revoked_at = NOW() WHERE organization_id = $1 AND delegate_id = $2`, testOrgID, assistantID)

		gedRepo := repository.NewGEDRepository(db.Pool)
		gedH := NewGEDHandler(gedRepo, nil, audit.NewService(db.Pool), db.Pool)

		docRegID := uuid.New()
		docQualID := uuid.New()
		verRegID := uuid.New()
		verQualID := uuid.New()

		_, err := db.Pool.Exec(ctx, `
			INSERT INTO ged_documents (id, organization_id, taxonomy_path, title_fr, document_code, document_type, status, current_version, created_by)
			VALUES
			  ($1, $3, '01_Affaires_Reglementaires/01_01_Agrements_Etablissement', 'Note Réglementaire Hors Intérim', 'DOC-REG-PERM-01', 'document_administratif', 'en_approbation_dt', 'v1.0', $4),
			  ($2, $3, '02_Assurance_Qualite/02_01_SOP', 'SOP Assurance Qualité Hors Intérim', 'DOC-QA-PERM-01', 'procedure_sop', 'en_approbation_dt', 'v1.0', $4)
		`, docRegID, docQualID, testOrgID, makerID)
		if err != nil {
			t.Fatalf("Erreur insertion documents GED permanent test: %v", err)
		}

		_, err = db.Pool.Exec(ctx, `
			INSERT INTO ged_versions (id, document_id, version_label, revision_summary, file_sha256, file_size_bytes, mime_type, original_filename, storage_path, author_id)
			VALUES
			  ($1, $3, 'v1.0', 'Version initiale', 'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc', 1200, 'application/pdf', 'reg2.pdf', '/tmp/reg2.pdf', $5),
			  ($2, $4, 'v1.0', 'Version initiale', 'dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd', 1200, 'application/pdf', 'qa2.pdf', '/tmp/qa2.pdf', $5)
		`, verRegID, verQualID, docRegID, docQualID, makerID)
		if err != nil {
			t.Fatalf("Erreur insertion versions GED permanent test: %v", err)
		}

		// Router permanent delegations (DT pour création/révocation)
		permRouter := chi.NewRouter()
		permRouter.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				u := &auth.SessionUser{
					UserID:         pdtID,
					OrganizationID: testOrgID,
					Email:          "pdt.interim@test.dz",
					Role:           "directeur_technique",
					Permissions:    []string{string(domain.PermDocApprovePDT)},
				}
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserContextKey, u)))
			})
		})
		permRouter.Post("/api/v1/parapheur/permanent-delegations", handler.CreateDTPermanentDelegation)
		permRouter.Post("/api/v1/parapheur/permanent-delegations/{id}/revoke", handler.RevokeDTPermanentDelegation)

		// 1. Tentative d'inclure 01_ (Affaires Réglementaires) -> 400 Bad Request
		badBody, _ := json.Marshal(domain.CreateDTPermanentDelegationRequest{
			DelegateID:              assistantID,
			HabilitationRef:         "NOTE-HAB-BAD",
			AllowedTaxonomyPrefixes: []string{"01_", "02_"},
		})
		badReq := httptest.NewRequest("POST", "/api/v1/parapheur/permanent-delegations", bytes.NewReader(badBody))
		badRR := httptest.NewRecorder()
		permRouter.ServeHTTP(badRR, badReq)
		if badRR.Code != http.StatusBadRequest {
			t.Fatalf("Attendu 400 pour périmètre 01_, reçu %d (%s)", badRR.Code, badRR.Body.String())
		}

		// 2. Création valide sur 02_ et 03_ -> 201 Created
		goodBody, _ := json.Marshal(domain.CreateDTPermanentDelegationRequest{
			DelegateID:              assistantID,
			HabilitationRef:         "NOTE-HAB-DT-2026-01",
			AllowedTaxonomyPrefixes: []string{"02_", "03_"},
			Notes:                   "Délégation permanente AQ et Production",
		})
		goodReq := httptest.NewRequest("POST", "/api/v1/parapheur/permanent-delegations", bytes.NewReader(goodBody))
		goodRR := httptest.NewRecorder()
		permRouter.ServeHTTP(goodRR, goodReq)
		if goodRR.Code != http.StatusCreated {
			t.Fatalf("Attendu 201 Created pour délégation permanente, reçu %d (%s)", goodRR.Code, goodRR.Body.String())
		}
		var createdPerm domain.DTPermanentDelegation
		if err := json.Unmarshal(goodRR.Body.Bytes(), &createdPerm); err != nil {
			t.Fatalf("Erreur décodage délégation permanente: %v", err)
		}

		// Router GED en tant que Pharmacien Assistant (hors intérim)
		gedRouter := chi.NewRouter()
		gedRouter.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				u := &auth.SessionUser{
					UserID:         assistantID,
					OrganizationID: testOrgID,
					Email:          "assistant.interim@test.dz",
					Role:           "pharmacien_assistant",
					Permissions:    []string{string(domain.PermDocCheckReview)},
				}
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserContextKey, u)))
			})
		})
		gedRouter.With(RequireAnyPermission(nil, domain.PermDocApprovePDT, domain.PermDocCheckReview)).Post("/api/v1/ged/documents/{id}/sign-dt", gedH.SignDocumentDT)

		signBody, _ := json.Marshal(map[string]string{"password": "Pass123!#"})

		// 3. Signature sur 01_Affaires_Reglementaires -> 403 Forbidden
		reqReg := httptest.NewRequest("POST", "/api/v1/ged/documents/"+docRegID.String()+"/sign-dt", bytes.NewReader(signBody))
		rrReg := httptest.NewRecorder()
		gedRouter.ServeHTTP(rrReg, reqReg)
		if rrReg.Code != http.StatusForbidden {
			t.Fatalf("Hors intérim sur 01_Affaires_Reglementaires, attendu 403, reçu %d (%s)", rrReg.Code, rrReg.Body.String())
		}

		// 4. Signature sur 02_Assurance_Qualite -> 200 OK
		reqQual := httptest.NewRequest("POST", "/api/v1/ged/documents/"+docQualID.String()+"/sign-dt", bytes.NewReader(signBody))
		rrQual := httptest.NewRecorder()
		gedRouter.ServeHTTP(rrQual, reqQual)
		if rrQual.Code != http.StatusOK {
			t.Fatalf("Hors intérim avec délégation permanente sur 02_, attendu 200 OK, reçu %d (%s)", rrQual.Code, rrQual.Body.String())
		}

		// 5. Révocation -> signature ultérieure bloquée (403)
		revReq := httptest.NewRequest("POST", "/api/v1/parapheur/permanent-delegations/"+createdPerm.ID.String()+"/revoke", bytes.NewReader([]byte(`{"reason":"Fin test"}`)))
		revRR := httptest.NewRecorder()
		permRouter.ServeHTTP(revRR, revReq)
		if revRR.Code != http.StatusOK {
			t.Fatalf("Attendu 200 OK pour révocation délégation permanente, reçu %d (%s)", revRR.Code, revRR.Body.String())
		}

		reqAfterRev := httptest.NewRequest("POST", "/api/v1/ged/documents/"+docQualID.String()+"/sign-dt", bytes.NewReader(signBody))
		rrAfterRev := httptest.NewRecorder()
		gedRouter.ServeHTTP(rrAfterRev, reqAfterRev)
		if rrAfterRev.Code != http.StatusForbidden {
			t.Fatalf("Après révocation, attendu 403 Forbidden, reçu %d (%s)", rrAfterRev.Code, rrAfterRev.Body.String())
		}
	})
}
