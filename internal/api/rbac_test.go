package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"homolog/internal/auth"
	"homolog/internal/domain"
)

func TestRequirePermission_Allowed(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()

	sessionUser := &auth.SessionUser{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          "ra.lead@pharma.dz",
		Role:           string(domain.RoleResponsableAffairesReg),
		Permissions:    []string{string(domain.PermDossierView), string(domain.PermDocRead)},
	}

	handler := RequirePermission(nil, domain.PermDossierView)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))

	req := httptest.NewRequest("GET", "/api/v1/dossiers", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, sessionUser)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("Statut attendu HTTP 200, obtenu: %d (corps: %s)", rr.Code, rr.Body.String())
	}
}

func TestRequirePermission_Forbidden403(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()

	// Utilisateur sans la permission DOC_APPROVE_PDT
	sessionUser := &auth.SessionUser{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          "ra.assoc@pharma.dz",
		Role:           string(domain.RoleResponsableAffairesReg),
		Permissions:    []string{string(domain.PermDossierView), string(domain.PermDocDraftMaker)},
	}

	handler := RequirePermission(nil, domain.PermDocApprovePDT)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/v1/dossiers/123/sign-pdt", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, sessionUser)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("Statut attendu HTTP 403 Forbidden, obtenu: %d", rr.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Erreur de décodage JSON: %v", err)
	}

	if resp["code"] != "ERR_RBAC_FORBIDDEN" {
		t.Errorf("Code d'erreur attendu ERR_RBAC_FORBIDDEN, obtenu: %v", resp["code"])
	}
}

func TestRequirePermission_ZeroTrustEnforcement(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()

	// Administrateur système local disposant artificiellement de DOC_READ
	sessionUser := &auth.SessionUser{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          "admin.sys@cloud.dz",
		Role:           string(domain.RoleAdminSystemeLocal),
		Permissions:    []string{string(domain.PermDocRead), string(domain.PermSysMaint)},
	}

	handler := RequirePermission(nil, domain.PermDocRead)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"file":"confidential_data"}`))
	}))

	req := httptest.NewRequest("GET", "/api/v1/ged/documents/456/download", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, sessionUser)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req.WithContext(ctx))

	// Le garde-fou Zero-Trust DOIT bloquer avec HTTP 403
	if rr.Code != http.StatusForbidden {
		t.Fatalf("Garde-fou Zero-Trust échoué : Statut attendu HTTP 403, obtenu: %d", rr.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Erreur de décodage JSON: %v", err)
	}

	if resp["code"] != "ERR_ZERO_TRUST_DENIED" {
		t.Errorf("Code attendu ERR_ZERO_TRUST_DENIED, obtenu: %v", resp["code"])
	}
}

func TestRequireAnyPermission(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()

	sessionUser := &auth.SessionUser{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          "qa.lead@pharma.dz",
		Role:           string(domain.RoleContributeurQualiteBPF),
		Permissions:    []string{string(domain.PermDocCheckReview)},
	}

	// Route accessible par Maker, Checker ou Archiveur
	handler := RequireAnyPermission(nil, domain.PermDocDraftMaker, domain.PermDocCheckReview, domain.PermDocArchive)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"status_updated"}`))
	}))

	req := httptest.NewRequest("POST", "/api/v1/ged/documents/789/status", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, sessionUser)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("Statut attendu HTTP 200, obtenu: %d", rr.Code)
	}
}

func TestSessionUser_HasPermission(t *testing.T) {
	u := &auth.SessionUser{
		Permissions: []string{
			string(domain.PermRefView),
			string(domain.PermRefEdit),
			string(domain.PermDossierView),
		},
	}

	if !u.HasPermission(string(domain.PermRefView)) {
		t.Error("HasPermission(REF_VIEW) devrait être vrai")
	}
	if !u.HasPermission(string(domain.PermRefEdit)) {
		t.Error("HasPermission(REF_EDIT) devrait être vrai")
	}
	if u.HasPermission(string(domain.PermDocApprovePDT)) {
		t.Error("HasPermission(DOC_APPROVE_PDT) devrait être faux")
	}

	var nilUser *auth.SessionUser
	if nilUser.HasPermission(string(domain.PermRefView)) {
		t.Error("HasPermission sur utilisateur nil devrait être faux")
	}
}

func TestConsultantCloisonnement_TransverseGEDForbidden(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()

	sessionUser := &auth.SessionUser{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          "consultant.ext@cro.com",
		Role:           string(domain.RoleConsultantExterne),
		Permissions:    []string{string(domain.PermDossierView), string(domain.PermDocDraftMaker)},
	}

	gedHandler := &GEDHandler{}

	req := httptest.NewRequest("GET", "/api/v1/ged/documents", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, sessionUser)
	rr := httptest.NewRecorder()

	gedHandler.ListDocuments(rr, req.WithContext(ctx))

	if rr.Code != http.StatusForbidden {
		t.Fatalf("Statut attendu HTTP 403 Forbidden pour consultant sur GED transverse, obtenu: %d", rr.Code)
	}

	reqTaxo := httptest.NewRequest("GET", "/api/v1/ged/taxonomy", nil)
	rrTaxo := httptest.NewRecorder()
	gedHandler.GetTaxonomy(rrTaxo, reqTaxo.WithContext(ctx))

	if rrTaxo.Code != http.StatusForbidden {
		t.Fatalf("Statut attendu HTTP 403 Forbidden pour consultant sur taxonomie GED, obtenu: %d", rrTaxo.Code)
	}
}
