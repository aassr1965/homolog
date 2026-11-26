package api

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
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
	"homolog/internal/storage"
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

func TestDownloadDVTIVArchivesGEDAndLinksChecklist(t *testing.T) {
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
	orgSlug := "lab-dvtiv-" + testOrgID.String()[:8]
	defer func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", testOrgID)
	}()

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, is_active)
		VALUES ($1, $2, 'SARL DVTIV Vigilance Pharma', true)
	`, testOrgID, orgSlug)
	if err != nil {
		t.Fatalf("échec création organisation : %v", err)
	}

	pdtUserID := uuid.New()
	hashedPwd, _ := auth.HashPassword("TestPdtPassword2026!", auth.DefaultArgon2Params)
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active)
		VALUES ($1, $2, 'pdt-dvtiv@test.dz', $3, 'Dr.', 'Samir Khelifi', 'directeur_technique', true)
	`, pdtUserID, testOrgID, hashedPwd)
	if err != nil {
		t.Fatalf("échec création utilisateur : %v", err)
	}

	corpRepo := repository.NewCorporateRepository(db.Pool)
	gedRepo := repository.NewGEDRepository(db.Pool)
	auditSvc := audit.NewService(db.Pool)
	store, err := storage.NewLocalCASStorage(t.TempDir())
	if err != nil {
		t.Fatalf("échec init storage : %v", err)
	}

	app := &domain.CorporateApplication{
		OrganizationID:    testOrgID,
		Category:          domain.CategoryExploitation,
		Nature:            domain.NatureAgrementInitial,
		ApplicationNumber: "MIPH-EXP-DVTIV-001",
		Title:             "Agrément Exploitation Test DVTIV",
		Status:            domain.StatusCorporateDraft,
		FormType:          "eForm-MIPH-EXPLOITATION",
		FormData:          map[string]interface{}{"raison_sociale": "SARL DVTIV Vigilance Pharma"},
		CreatedBy:         &pdtUserID,
	}
	if err := corpRepo.CreateApplication(ctx, app); err != nil {
		t.Fatalf("échec CreateApplication: %v", err)
	}

	lotHandler := NewLotReleaseHandler(db.Pool, auditSvc, gedRepo, corpRepo, store)
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			su := &auth.SessionUser{
				UserID:         pdtUserID,
				OrganizationID: testOrgID,
				Email:          "pdt-dvtiv@test.dz",
				Role:           "directeur_technique",
			}
			ctxWithUser := context.WithValue(req.Context(), UserContextKey, su)
			next.ServeHTTP(w, req.WithContext(ctxWithUser))
		})
	})
	r.Get("/api/v1/dvtiv/pdt-qppv-form", lotHandler.DownloadDVTIVPDTQPPVForm)
	r.Get("/api/v1/lot-releases/dce-forms/{formType}", lotHandler.DownloadDCEForm)

	req := httptest.NewRequest("GET", "/api/v1/dvtiv/pdt-qppv-form", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("statut attendu 200, reçu %d: %s", w.Code, w.Body.String())
	}
	gedDocIDStr := w.Header().Get("X-GED-Document-Id")
	if gedDocIDStr == "" {
		t.Fatalf("header X-GED-Document-Id manquant")
	}
	if w.Header().Get("X-Checklists-Linked") != "1" {
		t.Fatalf("X-Checklists-Linked attendu '1', reçu %q", w.Header().Get("X-Checklists-Linked"))
	}

	gedDocID, err := uuid.Parse(gedDocIDStr)
	if err != nil {
		t.Fatalf("UUID GED invalide %q: %v", gedDocIDStr, err)
	}
	doc, err := gedRepo.GetDocumentByID(ctx, testOrgID, gedDocID)
	if err != nil {
		t.Fatalf("document GED introuvable en base: %v", err)
	}
	if doc.Status != domain.GEDStatusBrouillon {
		t.Fatalf("statut GED attendu %q, reçu %q", domain.GEDStatusBrouillon, doc.Status)
	}
	if doc.TaxonomyPath != "04_Vigilance_Post_Market/04_01_Pharmacovigilance" {
		t.Fatalf("taxonomy_path inattendu: %q", doc.TaxonomyPath)
	}

	updatedApp, err := corpRepo.GetApplication(ctx, testOrgID, app.ID)
	if err != nil {
		t.Fatalf("GetApplication erreur: %v", err)
	}
	var foundExpPV01 bool
	for _, it := range updatedApp.ChecklistItems {
		if it.ItemCode == "EXP.PV.01" {
			foundExpPV01 = true
			if it.Status != "rattache" {
				t.Fatalf("EXP.PV.01 statut attendu 'rattache' (en attente de signature), reçu %q", it.Status)
			}
			if it.GEDDocumentID == nil || *it.GEDDocumentID != gedDocID {
				t.Fatalf("EXP.PV.01 ged_document_id attendu %s, reçu %v", gedDocID, it.GEDDocumentID)
			}
		}
	}
	if !foundExpPV01 {
		t.Fatalf("item EXP.PV.01 introuvable sur la demande Exploitation")
	}

	// Vérifier l'archivage GED des déclarations DCE et la non-archivage des simples modèles
	reqDecl := httptest.NewRequest("GET", "/api/v1/lot-releases/dce-forms/declaration-journaliere-dm", nil)
	wDecl := httptest.NewRecorder()
	r.ServeHTTP(wDecl, reqDecl)
	if wDecl.Code != http.StatusOK || wDecl.Header().Get("X-GED-Document-Id") == "" {
		t.Fatalf("declaration-journaliere-dm doit être archivée en GED: code=%d ged=%q", wDecl.Code, wDecl.Header().Get("X-GED-Document-Id"))
	}

	reqComp := httptest.NewRequest("GET", "/api/v1/lot-releases/dce-forms/composition-dossier", nil)
	wComp := httptest.NewRecorder()
	r.ServeHTTP(wComp, reqComp)
	if wComp.Code != http.StatusOK || wComp.Header().Get("X-GED-Document-Id") != "" {
		t.Fatalf("composition-dossier ne doit pas être archivé en GED: code=%d ged=%q", wComp.Code, wComp.Header().Get("X-GED-Document-Id"))
	}
}

func TestReliancePathwayDossierAndLotRelease(t *testing.T) {
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
	defer func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", testOrgID)
	}()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO organizations (id, slug, legal_name, is_active) VALUES ($1, $2, 'Labo Reliance Test', true)`, testOrgID, "lab-rel-"+testOrgID.String()[:8]); err != nil {
		t.Fatalf("création org: %v", err)
	}

	userID := uuid.New()
	hashedPwd, _ := auth.HashPassword("TestReliance2026!", auth.DefaultArgon2Params)
	if _, err := db.Pool.Exec(ctx, `INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active) VALUES ($1, $2, 'pdt-rel@test.dz', $3, 'Dr.', 'Reliance', 'directeur_technique', true)`, userID, testOrgID, hashedPwd); err != nil {
		t.Fatalf("création user: %v", err)
	}

	dossierID := uuid.New()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO dossiers (id, organization_id, product_type, dossier_code, procedure_code, internal_reference, status) VALUES ($1, $2, 'medicament', 'DOS-REL-001', 'NI', 'REF-REL-001', 'en_constitution')`, dossierID, testOrgID); err != nil {
		t.Fatalf("création dossier: %v", err)
	}

	dossierRepo := repository.NewDossierRepository(db.Pool)
	pharmaRepo := repository.NewPharmaRepository(db.Pool)
	gedRepo := repository.NewGEDRepository(db.Pool)
	corpRepo := repository.NewCorporateRepository(db.Pool)
	auditSvc := audit.NewService(db.Pool)
	store, _ := storage.NewLocalCASStorage(t.TempDir())

	dossierHandler := NewDossierHandler(dossierRepo, pharmaRepo, gedRepo, auditSvc, db.Pool).WithStorage(store)
	lotHandler := NewLotReleaseHandler(db.Pool, auditSvc, gedRepo, corpRepo, store)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			su := &auth.SessionUser{UserID: userID, OrganizationID: testOrgID, Email: "pdt-rel@test.dz", Role: "directeur_technique"}
			next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), UserContextKey, su)))
		})
	})
	r.Put("/api/v1/dossiers/{id}/reliance", dossierHandler.UpdateDossierReliance)
	r.Post("/api/v1/lot-releases", lotHandler.CreateLotRelease)
	r.Get("/api/v1/lot-releases", lotHandler.ListLotReleases)

	// 1. Activer Reliance sur le dossier CTD
	req := httptest.NewRequest("PUT", "/api/v1/dossiers/"+dossierID.String()+"/reliance", bytes.NewBufferString(`{"is_reliance":true,"reliance_authority":"EMA","reliance_type":"enregistrement","reliance_reference":"EMEA/H/C/005432"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT reliance attendu 200, reçu %d: %s", w.Code, w.Body.String())
	}
	var updated domain.Dossier
	_ = json.Unmarshal(w.Body.Bytes(), &updated)
	if !updated.IsReliance || updated.RelianceAuthority != "EMA" || updated.RelianceReference != "EMEA/H/C/005432" {
		t.Fatalf("champs Reliance non persistés: %+v", updated)
	}

	// 2. Créer un lot DCE avec Reliance Tests Labo
	reqLot := httptest.NewRequest("POST", "/api/v1/lot-releases", bytes.NewBufferString(`{"product_kind":"pp","commercial_name":"Biomox Rel","lot_number":"LOT-REL-99","is_reliance_lab":true,"reliance_lab_authority":"EDQM","reliance_lab_reference":"OMCL-2026-77"}`))
	reqLot.Header.Set("Content-Type", "application/json")
	wLot := httptest.NewRecorder()
	r.ServeHTTP(wLot, reqLot)
	if wLot.Code != http.StatusCreated {
		t.Fatalf("POST lot-releases attendu 201, reçu %d: %s", wLot.Code, wLot.Body.String())
	}

	reqList := httptest.NewRequest("GET", "/api/v1/lot-releases", nil)
	wList := httptest.NewRecorder()
	r.ServeHTTP(wList, reqList)
	if !strings.Contains(wList.Body.String(), `"is_reliance_lab":true`) || !strings.Contains(wList.Body.String(), `"EDQM"`) {
		t.Fatalf("Reliance labo absent de la liste des lots: %s", wList.Body.String())
	}
}

func TestLotRecall24hWorkflowAndGEDClose(t *testing.T) {
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
	defer func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", testOrgID)
	}()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO organizations (id, slug, legal_name, is_active) VALUES ($1, $2, 'Labo Recall Test', true)`, testOrgID, "lab-rcl-"+testOrgID.String()[:8]); err != nil {
		t.Fatalf("création org: %v", err)
	}

	userID := uuid.New()
	hashedPwd, _ := auth.HashPassword("TestRecall2026!", auth.DefaultArgon2Params)
	if _, err := db.Pool.Exec(ctx, `INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active) VALUES ($1, $2, 'pdt-rcl@test.dz', $3, 'Dr.', 'Recall', 'directeur_technique', true)`, userID, testOrgID, hashedPwd); err != nil {
		t.Fatalf("création user: %v", err)
	}

	gedRepo := repository.NewGEDRepository(db.Pool)
	auditSvc := audit.NewService(db.Pool)
	store, _ := storage.NewLocalCASStorage(t.TempDir())
	recallHandler := NewRecallHandler(db.Pool, auditSvc, gedRepo, store)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			su := &auth.SessionUser{UserID: userID, OrganizationID: testOrgID, Email: "pdt-rcl@test.dz", Role: "directeur_technique"}
			next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), UserContextKey, su)))
		})
	})
	r.Post("/api/v1/recalls", recallHandler.CreateRecall)
	r.Post("/api/v1/recalls/{id}/notify-anpp", recallHandler.NotifyANPP)
	r.Post("/api/v1/recalls/{id}/close", recallHandler.CloseRecall)
	r.Get("/api/v1/recalls", recallHandler.ListRecalls)

	req := httptest.NewRequest("POST", "/api/v1/recalls", bytes.NewBufferString(`{"product_kind":"pp","product_name":"Biomox 1g","lot_numbers":"LOT-RCL-01","severity_class":"CLASSE_I","is_mock_drill":true,"reason":"Exercice à blanc annuel"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST recalls attendu 201, reçu %d: %s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	recallID, _ := created["id"].(string)

	reqN := httptest.NewRequest("POST", "/api/v1/recalls/"+recallID+"/notify-anpp", nil)
	wN := httptest.NewRecorder()
	r.ServeHTTP(wN, reqN)
	if wN.Code != http.StatusOK || !strings.Contains(wN.Body.String(), `"within_24h_sla":true`) {
		t.Fatalf("notify-anpp échoué: code=%d body=%s", wN.Code, wN.Body.String())
	}

	reqC := httptest.NewRequest("POST", "/api/v1/recalls/"+recallID+"/close", bytes.NewBufferString(`{"resolution_notes":"100% récupéré"}`))
	reqC.Header.Set("Content-Type", "application/json")
	wC := httptest.NewRecorder()
	r.ServeHTTP(wC, reqC)
	if wC.Code != http.StatusOK || !strings.Contains(wC.Body.String(), `"ged_pv_doc_code":"PV-RCL-`) {
		t.Fatalf("close recall échoué: code=%d body=%s", wC.Code, wC.Body.String())
	}
}

func TestInspectionCAPAAndProductStatusLotGating(t *testing.T) {
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
	defer func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", testOrgID)
	}()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO organizations (id, slug, legal_name, is_active) VALUES ($1, $2, 'Labo Insp Status Test', true)`, testOrgID, "lab-insp-"+testOrgID.String()[:8]); err != nil {
		t.Fatalf("création org: %v", err)
	}

	pdtID := uuid.New()
	hashedPwd, _ := auth.HashPassword("TestInsp2026!", auth.DefaultArgon2Params)
	if _, err := db.Pool.Exec(ctx, `INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active) VALUES ($1, $2, 'pdt-insp@test.dz', $3, 'Dr.', 'Insp', 'directeur_technique', true)`, pdtID, testOrgID, hashedPwd); err != nil {
		t.Fatalf("création user: %v", err)
	}

	dmID := uuid.New()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO medical_devices (id, organization_id, commercial_name_fr, intended_use_fr, dm_class, manufacturer_reference) VALUES ($1, $2, 'Cathéter Test', 'Cardio', 'III', 'REF-CATH-001')`, dmID, testOrgID); err != nil {
		t.Fatalf("création DM: %v", err)
	}

	gedRepo := repository.NewGEDRepository(db.Pool)
	corpRepo := repository.NewCorporateRepository(db.Pool)
	auditSvc := audit.NewService(db.Pool)
	store, _ := storage.NewLocalCASStorage(t.TempDir())
	inspHandler := NewInspectionHandler(db.Pool, auditSvc, gedRepo, store)
	lotHandler := NewLotReleaseHandler(db.Pool, auditSvc, gedRepo, corpRepo, store)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			su := &auth.SessionUser{UserID: pdtID, OrganizationID: testOrgID, Email: "pdt-insp@test.dz", Role: "directeur_technique"}
			next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), UserContextKey, su)))
		})
	})
	r.Post("/api/v1/inspections/findings", inspHandler.CreateFinding)
	r.Post("/api/v1/inspections/findings/{id}/capa", inspHandler.SubmitCAPA)
	r.Post("/api/v1/inspections/findings/{id}/close", inspHandler.CloseFinding)
	r.Put("/api/v1/regulatory-status/{kind}/{id}", inspHandler.UpdateProductRegulatoryStatus)
	r.Post("/api/v1/lot-releases", lotHandler.CreateLotRelease)

	// 1. Phase 4 : Observation CRITIQUE -> CAPA -> Clôture GED 02.04
	reqF := httptest.NewRequest("POST", "/api/v1/inspections/findings", bytes.NewBufferString(`{"inspection_reference":"INSP-2026-01","severity":"CRITIQUE","description":"Ecart traçabilité lot"}`))
	reqF.Header.Set("Content-Type", "application/json")
	wF := httptest.NewRecorder()
	r.ServeHTTP(wF, reqF)
	if wF.Code != http.StatusCreated {
		t.Fatalf("POST finding attendu 201, reçu %d: %s", wF.Code, wF.Body.String())
	}
	var fRes map[string]any
	_ = json.Unmarshal(wF.Body.Bytes(), &fRes)
	fid, _ := fRes["id"].(string)

	reqCAPA := httptest.NewRequest("POST", "/api/v1/inspections/findings/"+fid+"/capa", bytes.NewBufferString(`{"capa_plan":"SOP révisée","responsible_name":"QA Lead"}`))
	reqCAPA.Header.Set("Content-Type", "application/json")
	wCAPA := httptest.NewRecorder()
	r.ServeHTTP(wCAPA, reqCAPA)
	if wCAPA.Code != http.StatusOK {
		t.Fatalf("POST capa attendu 200, reçu %d", wCAPA.Code)
	}

	reqCl := httptest.NewRequest("POST", "/api/v1/inspections/findings/"+fid+"/close", nil)
	wCl := httptest.NewRecorder()
	r.ServeHTTP(wCl, reqCl)
	if wCl.Code != http.StatusOK || !strings.Contains(wCl.Body.String(), `"ged_report_code":"CAPA-INSP-`) {
		t.Fatalf("close finding échoué: code=%d body=%s", wCl.Code, wCl.Body.String())
	}

	// 2. Phase 5 : Passer le DM à 'suspendu' -> blocage 409 sans motif -> succès 201 avec motif PDT
	reqSt := httptest.NewRequest("PUT", "/api/v1/regulatory-status/dm/"+dmID.String(), bytes.NewBufferString(`{"regulatory_status":"suspendu","status_decision_ref":"DEC-SUSP-2026","status_reason":"Suspension temporaire"}`))
	reqSt.Header.Set("Content-Type", "application/json")
	wSt := httptest.NewRecorder()
	r.ServeHTTP(wSt, reqSt)
	if wSt.Code != http.StatusOK {
		t.Fatalf("PUT regulatory-status attendu 200, reçu %d: %s", wSt.Code, wSt.Body.String())
	}

	reqBlocked := httptest.NewRequest("POST", "/api/v1/lot-releases", bytes.NewBufferString(`{"product_kind":"dm","medical_device_id":"`+dmID.String()+`","commercial_name":"Cathéter Test","lot_number":"LOT-SUSP-1"}`))
	reqBlocked.Header.Set("Content-Type", "application/json")
	wBlocked := httptest.NewRecorder()
	r.ServeHTTP(wBlocked, reqBlocked)
	if wBlocked.Code != http.StatusConflict {
		t.Fatalf("lot sur DM suspendu sans motif attendu 409, reçu %d: %s", wBlocked.Code, wBlocked.Body.String())
	}

	reqOverride := httptest.NewRequest("POST", "/api/v1/lot-releases", bytes.NewBufferString(`{"product_kind":"dm","medical_device_id":"`+dmID.String()+`","commercial_name":"Cathéter Test","lot_number":"LOT-SUSP-1","pdt_override_reason":"Lot importé avant suspension"}`))
	reqOverride.Header.Set("Content-Type", "application/json")
	wOverride := httptest.NewRecorder()
	r.ServeHTTP(wOverride, reqOverride)
	if wOverride.Code != http.StatusCreated {
		t.Fatalf("contournement PDT avec motif attendu 201, reçu %d: %s", wOverride.Code, wOverride.Body.String())
	}
}

func TestANPPFormDraftsEditableAndDOCXInjection(t *testing.T) {
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
	defer func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", testOrgID)
	}()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO organizations (id, slug, legal_name, is_active) VALUES ($1, $2, 'Labo Drafts Test', true)`, testOrgID, "lab-draft-"+testOrgID.String()[:8]); err != nil {
		t.Fatalf("création org: %v", err)
	}

	pdtID := uuid.New()
	hashedPwd, _ := auth.HashPassword("TestDraft2026!", auth.DefaultArgon2Params)
	if _, err := db.Pool.Exec(ctx, `INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active) VALUES ($1, $2, 'pdt-draft@test.dz', $3, 'Dr.', 'Draft', 'directeur_technique', true)`, pdtID, testOrgID, hashedPwd); err != nil {
		t.Fatalf("création user: %v", err)
	}

	dossierID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO dossiers (id, organization_id, internal_reference, dossier_code, procedure_code, product_type, status)
		VALUES ($1, $2, 'REF-DRAFT-01', 'REF-DRAFT-01', 'NF', 'medicament', 'brouillon')
	`, dossierID, testOrgID)
	if err != nil {
		t.Fatalf("création dossier test: %v", err)
	}

	dossierRepo := repository.NewDossierRepository(db.Pool)
	pharmaRepo := repository.NewPharmaRepository(db.Pool)
	gedRepo := repository.NewGEDRepository(db.Pool)
	auditSvc := audit.NewService(db.Pool)
	store, _ := storage.NewLocalCASStorage(t.TempDir())
	dossierHandler := NewDossierHandler(dossierRepo, pharmaRepo, gedRepo, auditSvc, db.Pool).WithStorage(store)
	draftHandler := NewANPPDraftHandler(db.Pool, auditSvc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			su := &auth.SessionUser{UserID: pdtID, OrganizationID: testOrgID, Email: "pdt-draft@test.dz", Role: "directeur_technique"}
			next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), UserContextKey, su)))
		})
	})
	r.Get("/api/v1/anpp-drafts/{kind}", draftHandler.GetDraft)
	r.Put("/api/v1/anpp-drafts/{kind}", draftHandler.SaveDraft)
	r.Get("/api/v1/dossiers/{id}/official-form", dossierHandler.DownloadOfficialDEPPForm)

	// 1. Enregistrer un brouillon Structure de Prix
	draftPayload := `{"substance_active":"Paracétamol 450 DZD/kg","prix_revient":"285.50 DZD","pcsu":"340.00 DZD","taux_change":"1 EUR = 145.20 DZD"}`
	reqPut := httptest.NewRequest("PUT", "/api/v1/anpp-drafts/structure_prix?entity_id="+dossierID.String(), bytes.NewBufferString(draftPayload))
	reqPut.Header.Set("Content-Type", "application/json")
	wPut := httptest.NewRecorder()
	r.ServeHTTP(wPut, reqPut)
	if wPut.Code != http.StatusOK {
		t.Fatalf("PUT anpp-drafts attendu 200, reçu %d: %s", wPut.Code, wPut.Body.String())
	}

	// 2. Relire et vérifier que le statut reste 'brouillon' et ré-éditable
	reqGet := httptest.NewRequest("GET", "/api/v1/anpp-drafts/structure_prix?entity_id="+dossierID.String(), nil)
	wGet := httptest.NewRecorder()
	r.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK {
		t.Fatalf("GET anpp-drafts attendu 200, reçu %d: %s", wGet.Code, wGet.Body.String())
	}
	if !bytes.Contains(wGet.Body.Bytes(), []byte(`"status":"brouillon"`)) || !bytes.Contains(wGet.Body.Bytes(), []byte(`"pcsu":"340.00 DZD"`)) {
		t.Fatalf("brouillon inattendu: %s", wGet.Body.String())
	}

	// 3. Mettre à jour le brouillon pour prouver qu'il reste éditable
	reqPut2 := httptest.NewRequest("PUT", "/api/v1/anpp-drafts/structure_prix?entity_id="+dossierID.String(), bytes.NewBufferString(`{"substance_active":"Paracétamol 460 DZD/kg","prix_revient":"290.00 DZD","pcsu":"355.00 DZD"}`))
	reqPut2.Header.Set("Content-Type", "application/json")
	wPut2 := httptest.NewRecorder()
	r.ServeHTTP(wPut2, reqPut2)
	if wPut2.Code != http.StatusOK {
		t.Fatalf("2e PUT anpp-drafts attendu 200, reçu %d: %s", wPut2.Code, wPut2.Body.String())
	}

	// 4. Télécharger le DOCX Structure de Prix et vérifier que le ZIP DOCX est généré (PK...)
	reqDoc := httptest.NewRequest("GET", "/api/v1/dossiers/"+dossierID.String()+"/official-form?type=structure-prix", nil)
	wDoc := httptest.NewRecorder()
	r.ServeHTTP(wDoc, reqDoc)
	if wDoc.Code != http.StatusOK {
		t.Fatalf("GET official-form structure-prix attendu 200, reçu %d: %s", wDoc.Code, wDoc.Body.String())
	}
	if wDoc.Body.Len() < 100 || wDoc.Body.Bytes()[0] != 'P' || wDoc.Body.Bytes()[1] != 'K' {
		t.Fatalf("fichier DOCX invalide (taille=%d)", wDoc.Body.Len())
	}
}

func TestANPPTemplateUploadOverrideAndRevert(t *testing.T) {
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
	defer func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM organizations WHERE id = $1", testOrgID)
	}()
	if _, err := db.Pool.Exec(ctx, `INSERT INTO organizations (id, slug, legal_name, is_active) VALUES ($1, $2, 'Labo Template Test', true)`, testOrgID, "lab-tpl-"+testOrgID.String()[:8]); err != nil {
		t.Fatalf("création org: %v", err)
	}

	raID := uuid.New()
	hashedPwd, _ := auth.HashPassword("TestTpl2026!", auth.DefaultArgon2Params)
	if _, err := db.Pool.Exec(ctx, `INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role, is_active) VALUES ($1, $2, 'ra-tpl@test.dz', $3, 'RA', 'Tpl', 'responsable_affaires_reglementaires', true)`, raID, testOrgID, hashedPwd); err != nil {
		t.Fatalf("création user: %v", err)
	}

	auditSvc := audit.NewService(db.Pool)
	tplHandler := NewANPPTemplateHandler(db.Pool, auditSvc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			su := &auth.SessionUser{UserID: raID, OrganizationID: testOrgID, Email: "ra-tpl@test.dz", Role: "responsable_affaires_reglementaires"}
			next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), UserContextKey, su)))
		})
	})
	r.Get("/api/v1/anpp-templates", tplHandler.ListTemplates)
	r.Get("/api/v1/anpp-templates/{name}/download", tplHandler.DownloadTemplate)
	r.Post("/api/v1/anpp-templates/{name}", tplHandler.UploadTemplate)
	r.Delete("/api/v1/anpp-templates/{name}", tplHandler.DeleteOverride)

	// 1. Télécharger le gabarit embarqué par défaut
	reqDl := httptest.NewRequest("GET", "/api/v1/anpp-templates/depp_modele_structure.docx/download", nil)
	wDl := httptest.NewRecorder()
	r.ServeHTTP(wDl, reqDl)
	if wDl.Code != http.StatusOK || wDl.Body.Len() < 100 {
		t.Fatalf("téléchargement défaut échoué: code=%d len=%d", wDl.Code, wDl.Body.Len())
	}

	// 2. Uploader une surcharge établissement via multipart/form-data
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", "depp_modele_structure.docx")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	customBytes := append([]byte{}, wDl.Body.Bytes()...)
	_, _ = fw.Write(customBytes)
	_ = mw.WriteField("notes", "MAJ ANPP Test 2026")
	_ = mw.WriteField("scope", "organization")
	_ = mw.Close()

	reqUp := httptest.NewRequest("POST", "/api/v1/anpp-templates/depp_modele_structure.docx", &body)
	reqUp.Header.Set("Content-Type", mw.FormDataContentType())
	wUp := httptest.NewRecorder()
	r.ServeHTTP(wUp, reqUp)
	if wUp.Code != http.StatusOK || !strings.Contains(wUp.Body.String(), `"source":"organization"`) {
		t.Fatalf("upload surcharge attendu 200 organization, reçu %d: %s", wUp.Code, wUp.Body.String())
	}

	// 3. Vérifier la liste et la restauration sur le défaut embarqué
	reqList := httptest.NewRequest("GET", "/api/v1/anpp-templates", nil)
	wList := httptest.NewRecorder()
	r.ServeHTTP(wList, reqList)
	if wList.Code != http.StatusOK || !strings.Contains(wList.Body.String(), `"MAJ ANPP Test 2026"`) {
		t.Fatalf("liste gabarits sans surcharge: %s", wList.Body.String())
	}

	reqDel := httptest.NewRequest("DELETE", "/api/v1/anpp-templates/depp_modele_structure.docx?scope=organization", nil)
	wDel := httptest.NewRecorder()
	r.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusOK {
		t.Fatalf("DELETE surcharge attendu 200, reçu %d: %s", wDel.Code, wDel.Body.String())
	}
}
