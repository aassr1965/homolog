package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/database"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

func TestCorporateAPI_FullWorkflow(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://homolog:homolog_secret_2026@127.0.0.1:5434/homolog?sslmode=disable"
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, dbURL)
	if err != nil {
		t.Skipf("PostgreSQL non disponible, test ignoré : %v", err)
		return
	}
	defer db.Close()

	testOrgID := uuid.New()
	orgSlug := "lab-corp-" + testOrgID.String()[:8]

	defer func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", testOrgID)
	}()

	// 1. Initialiser l'organisation et utilisateur Pharmacien DT
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, is_active)
		VALUES ($1, $2, 'SARL Corporate Test Pharma', true)
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation : %v", err)
	}

	pdtUserID := uuid.New()
	hashedPwd, _ := auth.HashPassword("TestPdtPassword2026!", auth.DefaultArgon2Params)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, $3, $4, 'Dr.', 'Pharmacien DT Corporate', 'pharmacien_directeur_technique', true)
	`, pdtUserID, testOrgID, "pdt-corp@test.dz", hashedPwd)
	if err != nil {
		t.Fatalf("échec création utilisateur : %v", err)
	}

	// 2. Initialiser un document GED pour test de rattachement
	gedDocID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO ged_documents (id, organization_id, taxonomy_path, title_fr, document_code, status, current_version)
		VALUES ($1, $2, '01_Affaires_Reglementaires', 'Statuts Notariés SARL', 'GED-CORP-001', 'approuve_dt', 'v1.0')
	`, gedDocID, testOrgID)
	if err != nil {
		t.Fatalf("échec création document GED : %v", err)
	}

	gedVersionID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO ged_versions (id, document_id, version_label, revision_summary, file_sha256, file_size_bytes, mime_type, original_filename, storage_path)
		VALUES ($1, $2, 'v1.0', 'Version initiale approuvée PDT', 'sha256-test-corporate-statuts', 204800, 'application/pdf', 'statuts.pdf', 'docs/statuts.pdf')
	`, gedVersionID, gedDocID)
	if err != nil {
		t.Fatalf("échec création version GED : %v", err)
	}

	auditSvc := audit.NewService(db.Pool)
	corpRepo := repository.NewCorporateRepository(db.Pool)
	corpHandler := NewCorporateHandler(corpRepo, auditSvc)

	mockUser := &auth.SessionUser{
		UserID:         pdtUserID,
		OrganizationID: testOrgID,
		Email:          "pdt-corp@test.dz",
		Role:           "pharmacien_directeur_technique",
		Permissions:    []string{string(domain.PermRefView), string(domain.PermRefEdit), string(domain.PermDocApprovePDT)},
	}

	// Routeur Chi de test
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), UserContextKey, mockUser)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})

	r.Get("/corporate/applications", corpHandler.ListApplications)
	r.Post("/corporate/applications", corpHandler.CreateApplication)
	r.Get("/corporate/applications/{id}", corpHandler.GetApplication)
	r.Put("/corporate/applications/{id}", corpHandler.UpdateApplication)
	r.Post("/corporate/applications/{id}/items/{itemId}/attach", corpHandler.AttachDocument)
	r.Post("/corporate/applications/{id}/items/{itemId}/detach", corpHandler.DetachDocument)
	r.Post("/corporate/applications/{id}/sign-pdt", corpHandler.SignPDT)
	r.Post("/corporate/applications/{id}/tabadol/submit", corpHandler.SubmitTabadol)
	r.Post("/corporate/applications/{id}/miph-receipt", corpHandler.RecordMIPHReceipt)
	r.Post("/corporate/applications/{id}/grant", corpHandler.GrantLicense)
	r.Get("/corporate/applications/{id}/tabadol/audit", corpHandler.AuditTabadol)
	r.Get("/corporate/applications/{id}/tabadol/export-package", corpHandler.ExportTabadolPackage)
	r.Get("/corporate/applications/{id}/bordereau-miph", corpHandler.GetMIPHBordereau)
	r.Get("/corporate/licenses", corpHandler.ListLicenses)

	// Étape A : Création d'une demande d'agrément préalable de réalisation (Fabrication)
	createReq := CreateApplicationRequest{
		Category:          domain.CategoryFabrication,
		Nature:            domain.NaturePrealableRealisation,
		ApplicationNumber: "MIPH-2026-FAB-TEST",
		Title:             "Usine de Formes Sèches & Liquides Oued Smar",
		InternalReference: "REF-CORP-2026-01",
		FormData: map[string]interface{}{
			"promoter_name":        "SARL Corporate Test Pharma",
			"site_address":         "Zone Industrielle Oued Smar, Lot 14",
			"site_wilaya":          "Alger",
			"forms":                []string{"Comprimés", "Gélules", "Sirops"},
			"estimated_investment": 850000000,
		},
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/corporate/applications", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("échec création demande (code %d) : %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success     bool                        `json:"success"`
		Application domain.CorporateApplication `json:"application"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal création: %v", err)
	}

	appID := resp.Application.ID
	if appID == uuid.Nil {
		t.Fatalf("ID d'application non généré")
	}
	if len(resp.Application.ChecklistItems) == 0 {
		t.Fatalf("checklist automatique non générée pour la fabrication")
	}

	t.Logf("✅ Demande d'agrément créée avec succès : %s (Checklist : %d exigences)", resp.Application.ApplicationNumber, len(resp.Application.ChecklistItems))

	// Étape B : Audit de pré-soumission bloquant (doit refuser le visa car 0 pièce [E] rattachée)
	req = httptest.NewRequest("GET", "/corporate/applications/"+appID.String()+"/tabadol/audit", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var auditResp struct {
		Success bool                      `json:"success"`
		Audit   domain.TabadolAuditReport `json:"audit"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &auditResp)
	if auditResp.Audit.CanSubmit {
		t.Fatalf("l'audit aurait dû bloquer car aucune pièce obligatoire [E] n'est rattachée")
	}
	t.Logf("✅ Audit bloquant vérifié : CanSubmit = false, %d pièces manquantes", len(auditResp.Audit.MissingMandatory))

	// Étape C : Rattachement de la pièce statutaire GED (FAB.PR.01)
	firstItemID := resp.Application.ChecklistItems[0].ID
	attachPayload := map[string]interface{}{"ged_document_id": gedDocID}
	attachBody, _ := json.Marshal(attachPayload)
	req = httptest.NewRequest("POST", "/corporate/applications/"+appID.String()+"/items/"+firstItemID.String()+"/attach", bytes.NewReader(attachBody))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("échec rattachement GED: %s", w.Body.String())
	}
	t.Logf("✅ Pièce GED rattachée à l'exigence FAB.PR.01 avec succès")

	// Étape D : Télésoumission Tabadol
	tabadolPayload := map[string]interface{}{
		"tabadol_reference": "TBD-2026-MIPH-89412",
		"notes":             "Télésoumission préalable effectuée avec succès sur gate.miph.gov.dz",
	}
	tabBody, _ := json.Marshal(tabadolPayload)
	req = httptest.NewRequest("POST", "/corporate/applications/"+appID.String()+"/tabadol/submit", bytes.NewReader(tabBody))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("échec soumission Tabadol: %s", w.Body.String())
	}
	t.Logf("✅ Télésoumission Tabadol enregistrée (Réf : TBD-2026-MIPH-89412)")

	// Étape E : Dépôt au bureau d'ordre central (Récépissé avec décharge)
	boPayload := map[string]interface{}{
		"miph_receipt_ref":  "REC-MIPH-BO-2026-00452",
		"miph_receipt_date": time.Now().Format("2006-01-02"),
	}
	boBody, _ := json.Marshal(boPayload)
	req = httptest.NewRequest("POST", "/corporate/applications/"+appID.String()+"/miph-receipt", bytes.NewReader(boBody))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("échec enregistrement récépissé BO: %s", w.Body.String())
	}
	t.Logf("✅ Récépissé physique Bureau d'Ordre enregistré (Réf : REC-MIPH-BO-2026-00452)")

	// Étape F : Octroi officiel de l'agrément ministériel & insertion au registre des licences
	grantPayload := map[string]interface{}{
		"decision_ref":  "ARR-MIPH-AGR-2026-088",
		"decision_date": time.Now().Format("2006-01-02"),
		"expiry_date":   time.Now().AddDate(5, 0, 0).Format("2006-01-02"),
	}
	grantBody, _ := json.Marshal(grantPayload)
	req = httptest.NewRequest("POST", "/corporate/applications/"+appID.String()+"/grant", bytes.NewReader(grantBody))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("échec octroi agrément: %s", w.Body.String())
	}
	t.Logf("✅ Arrêté d'agrément ministériel enregistré avec succès (ARR-MIPH-AGR-2026-088)")

	// Étape G : Consultation du Registre des Licences Corporate & Compte à rebours
	req = httptest.NewRequest("GET", "/corporate/licenses", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var licResp struct {
		Success  bool                      `json:"success"`
		Licenses []domain.CorporateLicense `json:"licenses"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &licResp)
	if len(licResp.Licenses) == 0 {
		t.Fatalf("aucune licence trouvée dans le registre")
	}
	t.Logf("✅ Registre Corporate interrogé : %d titre(s) actif(s), échéance calculée : %d jours restants",
		len(licResp.Licenses), *licResp.Licenses[0].DaysUntilExpiry,
	)

	// Étape H : Export du Submission Package Tabadol (.zip)
	req = httptest.NewRequest("GET", "/corporate/applications/"+appID.String()+"/tabadol/export-package", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("échec export submission package: %s", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "application/zip" {
		t.Fatalf("Content-Type attendu application/zip, reçu %s", w.Header().Get("Content-Type"))
	}
	t.Logf("✅ Submission Package Tabadol (.zip) exporté avec succès (%d octets)", w.Body.Len())

	// Étape I : Consultation et rendu du Bordereau officiel d'accompagnement Bureau d'Ordre MIPH
	req = httptest.NewRequest("GET", "/corporate/applications/"+appID.String()+"/bordereau-miph", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("échec génération bordereau MIPH: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Bordereau Récapitulatif") || !strings.Contains(w.Body.String(), "Bureau d'Ordre Central") {
		t.Fatalf("contenu du bordereau invalide: %s", w.Body.String()[:200])
	}
	t.Logf("✅ Bordereau officiel d'accompagnement Bureau d'Ordre MIPH généré avec succès (%d octets HTML)", w.Body.Len())
}
