package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
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
