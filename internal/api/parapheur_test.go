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
