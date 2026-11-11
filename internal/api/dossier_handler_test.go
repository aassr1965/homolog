package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func TestDossierAPI_Workflow(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := context.Background()
	testOrgID := uuid.New()
	orgSlug := "org-test-dossier-" + testOrgID.String()[:8]

	// 1. Créer l'organisation de test
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, $2, 'Laboratoire Test ANPP', 'MIPH-2026-TEST', 'Dr. Responsable Test')
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation test: %v", err)
	}

	// 2. Créer un utilisateur Directeur Technique pour la signature
	dtUserID := uuid.New()
	rawPassword := "SecurePassword2026!#"
	pwdHash, _ := auth.HashPassword(rawPassword, nil)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, $3, $4, 'Amine', 'K.', 'directeur_technique', TRUE)
	`, dtUserID, testOrgID, "dt-"+testOrgID.String()[:6]+"@sanad.dz", pwdHash)
	if err != nil {
		t.Fatalf("échec création utilisateur DT: %v", err)
	}

	// 3. Créer un médicament de test (Générique importé)
	medID := uuid.New()
	prodID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO products (id, organization_id, product_type, brand_name_fr, manufacturer_name, marketing_status)
		VALUES ($1, $2, 'medicament', 'Amoxicilline Sanad 1g', 'Laboratoires Sanad Pharma', 'en_preparation')
	`, prodID, testOrgID)
	if err != nil {
		t.Fatalf("échec création produit: %v", err)
	}

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO medicines (
			id, organization_id, product_id, origin_type, brand_name_fr,
			dci_fr, dosage_fr, pharmaceutical_form_fr, administration_route_fr,
			economic_file_type, fob_price, fob_currency, status
		) VALUES (
			$1, $2, $3, 'generique_importe', 'Amoxicilline Sanad 1g',
			'Amoxicilline trihydratée', '1000 mg', 'Comprimé pelliculé', 'Orale',
			'fob_benchmark', 2.45, 'EUR', 'en_constitution'
		)
	`, medID, testOrgID, prodID)
	if err != nil {
		t.Fatalf("échec création médicament: %v", err)
	}

	// 4. Créer un document GED test pour le rattachement
	gedDocID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO ged_documents (
			id, organization_id, taxonomy_path, title_fr, document_code,
			status, current_version, created_by
		) VALUES (
			$1, $2, '01_Affaires_Reglementaires/01_04_Dossiers_CTD_Reference',
			'Formulaire officiel ANPP visé', 'ANPP-FORM-TEST',
			'en_vigueur', 'v1.0', $3
		)
	`, gedDocID, testOrgID, dtUserID)
	if err != nil {
		t.Fatalf("échec création doc GED: %v", err)
	}

	// Instanciation des services et handlers
	auditSvc := audit.NewService(db.Pool)
	pharmaRepo := repository.NewPharmaRepository(db.Pool)
	gedRepo := repository.NewGEDRepository(db.Pool)
	dossierRepo := repository.NewDossierRepository(db.Pool)
	handler := NewDossierHandler(dossierRepo, pharmaRepo, gedRepo, auditSvc, db.Pool)

	dtSession := &auth.SessionUser{
		SessionID:      uuid.New(),
		UserID:         dtUserID,
		OrganizationID: testOrgID,
		Email:          "dt@sanad.dz",
		Role:           "directeur_technique",
		FirstName:      "Dr.",
		LastName:       "Pharmacien DT",
	}

	var createdDossierID uuid.UUID
	var firstItemID uuid.UUID

	// TEST A : Création d'un dossier CTD avec auto-génération de la checklist officielle ANPP
	t.Run("POST /api/v1/dossiers - Création et auto-génération de la Checklist CTD", func(t *testing.T) {
		reqBody := CreateDossierRequest{
			DossierCode:       "DOS-TEST-AMOX-NI",
			ProcedureCode:     domain.ProcedureNI,
			ProductType:       domain.ProductTypeMedicament,
			MedicineID:        &medID,
			InternalReference: "REF-TEST-001",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/api/v1/dossiers", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, dtSession))

		rr := httptest.NewRecorder()
		handler.CreateDossier(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("attendu code 201, reçu %d: %s", rr.Code, rr.Body.String())
		}

		var resp domain.Dossier
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("erreur désérialisation réponse: %v", err)
		}

		if resp.DossierCode != "DOS-TEST-AMOX-NI" {
			t.Fatalf("code dossier incorrect: %s", resp.DossierCode)
		}
		if len(resp.Items) < 30 {
			t.Fatalf("checklist incomplète, attendu > 30 items CTD, reçu: %d", len(resp.Items))
		}

		createdDossierID = resp.ID
		firstItemID = resp.Items[0].ID
	})

	// TEST B : Consultation du dossier complet avec ses exigences
	t.Run("GET /api/v1/dossiers/{id} - Consultation du dossier", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/dossiers/%s", createdDossierID), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", createdDossierID.String())
		ctxWithRoute := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		req = req.WithContext(context.WithValue(ctxWithRoute, UserContextKey, dtSession))

		rr := httptest.NewRecorder()
		handler.GetDossier(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("attendu code 200, reçu %d: %s", rr.Code, rr.Body.String())
		}

		var resp domain.Dossier
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		if resp.CompletenessPercent != 0 {
			t.Fatalf("complétude initiale attendue à 0%%, reçu %d%%", resp.CompletenessPercent)
		}
	})

	// TEST C : Rattachement d'un document GED à un emplacement CTD
	t.Run("POST /api/v1/dossiers/{id}/items/{itemId}/attach - Rattachement GED", func(t *testing.T) {
		notes := "Attestation officielle signée"
		reqBody := AttachDocumentRequest{
			GEDDocumentID: gedDocID,
			Notes:         &notes,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		url := fmt.Sprintf("/api/v1/dossiers/%s/items/%s/attach", createdDossierID, firstItemID)
		req := httptest.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", createdDossierID.String())
		rctx.URLParams.Add("itemId", firstItemID.String())
		ctxWithRoute := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		req = req.WithContext(context.WithValue(ctxWithRoute, UserContextKey, dtSession))

		rr := httptest.NewRecorder()
		handler.AttachDocument(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("attendu code 200, reçu %d: %s", rr.Code, rr.Body.String())
		}

		var item domain.ChecklistItem
		_ = json.Unmarshal(rr.Body.Bytes(), &item)
		if item.Status != domain.ItemStatusAttached {
			t.Fatalf("statut item attendu 'rattache', reçu: %s", item.Status)
		}
		if item.GEDDocumentID == nil || *item.GEDDocumentID != gedDocID {
			t.Fatalf("identifiant ged_document_id non renseigné")
		}
	})

	// TEST D : Verrouillage 21 CFR Part 11 si complétude < 100%
	t.Run("POST /api/v1/dossiers/{id}/sign-pdt - Blocage si complétude < 100%", func(t *testing.T) {
		reqBody := SignPDTRequest{
			Password: rawPassword,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		url := fmt.Sprintf("/api/v1/dossiers/%s/sign-pdt", createdDossierID)
		req := httptest.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", createdDossierID.String())
		ctxWithRoute := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		req = req.WithContext(context.WithValue(ctxWithRoute, UserContextKey, dtSession))

		rr := httptest.NewRecorder()
		handler.SignPDT(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("la signature d'un dossier incomplet aurait dû être refusée avec 400, reçu: %d (%s)", rr.Code, rr.Body.String())
		}
	})

	// TEST E : Refus de signature avec mot de passe erroné (21 CFR Part 11)
	t.Run("POST /api/v1/dossiers/{id}/sign-pdt - Refus re-authentification mot de passe faux", func(t *testing.T) {
		reqBody := SignPDTRequest{
			Password: "WrongPassword999!",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		url := fmt.Sprintf("/api/v1/dossiers/%s/sign-pdt", createdDossierID)
		req := httptest.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", createdDossierID.String())
		ctxWithRoute := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		req = req.WithContext(context.WithValue(ctxWithRoute, UserContextKey, dtSession))

		rr := httptest.NewRecorder()
		handler.SignPDT(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("attendu code 401 Unauthorized pour faux mot de passe, reçu: %d", rr.Code)
		}
	})

	// TEST F : Détachement d'un document GED
	t.Run("POST /api/v1/dossiers/{id}/items/{itemId}/detach - Détachement GED", func(t *testing.T) {
		url := fmt.Sprintf("/api/v1/dossiers/%s/items/%s/detach", createdDossierID, firstItemID)
		req := httptest.NewRequest("POST", url, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", createdDossierID.String())
		rctx.URLParams.Add("itemId", firstItemID.String())
		ctxWithRoute := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		req = req.WithContext(context.WithValue(ctxWithRoute, UserContextKey, dtSession))

		rr := httptest.NewRecorder()
		handler.DetachDocument(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("attendu code 200, reçu %d: %s", rr.Code, rr.Body.String())
		}

		var item domain.ChecklistItem
		_ = json.Unmarshal(rr.Body.Bytes(), &item)
		if item.Status != domain.ItemStatusMissing {
			t.Fatalf("statut attendu 'manquant', reçu: %s", item.Status)
		}
		if item.GEDDocumentID != nil {
			t.Fatalf("ged_document_id aurait dû être nil")
		}
	})
}
