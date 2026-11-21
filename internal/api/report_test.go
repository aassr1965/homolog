package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"homolog/internal/domain"
	"homolog/internal/reporting"
	"homolog/internal/repository"
)

func TestReportHandler_GetReportRegistry(t *testing.T) {
	handler := NewReportHandler(nil, nil, nil, nil, nil)

	// 1. Appel sans filtre (catalogue complet)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/registry", nil)
	rr := httptest.NewRecorder()

	handler.GetReportRegistry(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Statut inattendu: attendu %d, reçu %d", http.StatusOK, rr.Code)
	}

	var resp struct {
		Success bool                         `json:"success"`
		Count   int                          `json:"count"`
		Reports []reporting.ReportDefinition `json:"reports"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Erreur décodage JSON: %v", err)
	}

	if !resp.Success {
		t.Errorf("Success attendu true")
	}
	if resp.Count < 28 {
		t.Errorf("Attendu au moins 28 rapports dans le registre, reçu %d", resp.Count)
	}

	// 2. Filtre par nature (kind=EI)
	reqEI := httptest.NewRequest(http.MethodGet, "/api/v1/reports/registry?kind=EI", nil)
	rrEI := httptest.NewRecorder()
	handler.GetReportRegistry(rrEI, reqEI)

	var respEI struct {
		Success bool                         `json:"success"`
		Count   int                          `json:"count"`
		Reports []reporting.ReportDefinition `json:"reports"`
	}
	_ = json.Unmarshal(rrEI.Body.Bytes(), &respEI)
	if respEI.Count != 8 {
		t.Errorf("Attendu 8 états intermédiaires (EI), reçu %d", respEI.Count)
	}

	// 3. Filtre par catégorie (category=DOSSIERS_CTD)
	reqCTD := httptest.NewRequest(http.MethodGet, "/api/v1/reports/registry?category=DOSSIERS_CTD", nil)
	rrCTD := httptest.NewRecorder()
	handler.GetReportRegistry(rrCTD, reqCTD)

	var respCTD struct {
		Success bool                         `json:"success"`
		Count   int                          `json:"count"`
		Reports []reporting.ReportDefinition `json:"reports"`
	}
	_ = json.Unmarshal(rrCTD.Body.Bytes(), &respCTD)
	if respCTD.Count < 5 {
		t.Errorf("Attendu au moins 5 rapports pour DOSSIERS_CTD, reçu %d", respCTD.Count)
	}
}

func TestReportHandler_VerifyReportPublic_NotFound(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Skip("PostgreSQL non disponible pour test d'intégration")
	}
	defer db.Close()

	workflowRepo := repository.NewWorkflowRepository(db.Pool)
	handler := NewReportHandler(nil, workflowRepo, nil, nil, db.Pool)

	r := chi.NewRouter()
	r.Get("/verify/{sha256}", handler.VerifyReportPublic)

	req := httptest.NewRequest(http.MethodGet, "/verify/nonexistent-hash-1234567890abcdef", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Doit renvoyer 404 avec page HTML d'erreur claire
	if rr.Code != http.StatusNotFound {
		t.Errorf("Attendu statut %d pour un hash inexistant, reçu %d", http.StatusNotFound, rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Document Non Reconnu") {
		t.Errorf("Message d'erreur attendu dans le HTML d'invalidation")
	}
}

func TestReportHandler_TabularCSVRendering(t *testing.T) {
	handler := NewReportHandler(nil, nil, nil, nil, nil)

	d := &domain.Dossier{
		DossierCode:       "DOS-2026-AMOX-NI",
		InternalReference: "REF-AMOX-01",
		ProcedureCode:     domain.ProcedureNI,
		ProductType:       domain.ProductTypeMedicine,
		Medicine: &domain.Medicine{
			BrandNameFR: "Amoxicilline LPA 1 g",
			DCIFR:       "Amoxicilline",
		},
	}

	now := time.Now().UTC()
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	// 1. Test ES-FIN-01 (Bordereau BV1)
	rows, filename := handler.generateTabularReportRows(d, "ES-FIN-01", domain.WatermarkOfficial, hash, now, nil)
	if len(rows) < 4 {
		t.Fatalf("Attendu au moins 4 lignes pour ES-FIN-01 (en-tête + 3 articles + total), reçu %d", len(rows))
	}
	if !strings.HasPrefix(filename, "es-fin-01_dos-2026-amox-ni") {
		t.Errorf("Nom de fichier inattendu: %s", filename)
	}

	// Vérifier présence du montant total 355000.00
	foundTotal := false
	for _, row := range rows {
		if len(row) > 11 && row[11] == "355000.00" {
			foundTotal = true
			break
		}
	}
	if !foundTotal {
		t.Errorf("Ligne de total 355000.00 DZD non trouvée dans le CSV")
	}

	// 2. Test ES-FIN-02 (Bordereau BV2)
	rows2, _ := handler.generateTabularReportRows(d, "ES-FIN-02", domain.WatermarkDraft, hash, now, nil)
	if len(rows2) < 3 {
		t.Fatalf("Attendu au moins 3 lignes pour ES-FIN-02, reçu %d", len(rows2))
	}

	// 3. Test ES-REG-01 (5 Modules CTD)
	rows3, _ := handler.generateTabularReportRows(d, "ES-REG-01", domain.WatermarkOfficial, hash, now, nil)
	if len(rows3) != 6 { // en-tête + 5 modules
		t.Errorf("Attendu 6 lignes (en-tête + M1..M5) pour ES-REG-01, reçu %d", len(rows3))
	}
}
