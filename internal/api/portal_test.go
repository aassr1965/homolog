package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"homolog/internal/config"
)

func TestPortalEnabled_EnvironmentAndDeploymentMode(t *testing.T) {
	cases := []struct {
		name string
		cfg  *config.Config
		want bool
	}{
		{"nil config", nil, true},
		{"dev on-premise", &config.Config{Environment: "development", DeploymentMode: config.ModeOnPremise}, true},
		{"dev saas", &config.Config{Environment: "development", DeploymentMode: config.ModeSaaS}, true},
		{"prod saas", &config.Config{Environment: "production", DeploymentMode: config.ModeSaaS}, true},
		{"prod on-premise", &config.Config{Environment: "production", DeploymentMode: config.ModeOnPremise}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := PortalEnabled(tc.cfg); got != tc.want {
				t.Fatalf("PortalEnabled(%v) = %v, want %v", tc.cfg, got, tc.want)
			}
		})
	}
}

func TestPortalAndApp_GateProtection(t *testing.T) {
	cfg := &config.Config{
		Environment:    "development",
		DeploymentMode: config.ModeSaaS,
		GatePasscode:   "Homolog2006!VIP",
		JWTSecret:      "test-secret-gate-portal-2026",
	}
	gate := NewGateHandler(cfg)

	// 1. Sas verrouillé : "/" et "/app" doivent servir gate.html (pas portal.html ni index.html)
	for _, path := range []string{"/", "/app"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		if path == "/" {
			ServePortalOrRedirect(cfg, gate)(rec, req)
		} else {
			ServeApp(cfg, gate)(rec, req)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s (locked) status = %d, want 200", path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "Homolog") && !strings.Contains(rec.Body.String(), "sas") {
			t.Fatalf("GET %s (locked) did not return gate.html", path)
		}
		if strings.Contains(rec.Body.String(), "sp-form-demo") || strings.Contains(rec.Body.String(), "screen-login") {
			t.Fatalf("GET %s (locked) leaked portal or app HTML!", path)
		}
	}

	// 2. Déverrouillage avec Homolog2006!VIP
	unlockBody := []byte(`{"passcode":"Homolog2006!VIP"}`)
	unlockReq := httptest.NewRequest(http.MethodPost, "/api/v1/gate/unlock", bytes.NewReader(unlockBody))
	unlockRec := httptest.NewRecorder()
	gate.Unlock(unlockRec, unlockReq)
	if unlockRec.Code != http.StatusOK {
		t.Fatalf("Gate unlock status = %d, want 200", unlockRec.Code)
	}
	cookies := unlockRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("Gate unlock did not set cookie")
	}

	// 3. Sas déverrouillé : "/" sert portal.html et "/app" sert index.html
	portalReq := httptest.NewRequest(http.MethodGet, "/", nil)
	portalReq.AddCookie(cookies[0])
	portalRec := httptest.NewRecorder()
	ServePortalOrRedirect(cfg, gate)(portalRec, portalReq)
	if portalRec.Code != http.StatusOK || !strings.Contains(portalRec.Body.String(), "sp-form-demo") {
		t.Fatalf("GET / (unlocked) failed to serve portal.html (status=%d)", portalRec.Code)
	}

	appReq := httptest.NewRequest(http.MethodGet, "/app", nil)
	appReq.AddCookie(cookies[0])
	appRec := httptest.NewRecorder()
	ServeApp(cfg, gate)(appRec, appReq)
	if appRec.Code != http.StatusOK || !strings.Contains(appRec.Body.String(), "screen-login") {
		t.Fatalf("GET /app (unlocked) failed to serve index.html (status=%d)", appRec.Code)
	}
}

func TestPortal_ProductionOnPremiseRedirectsToApp(t *testing.T) {
	cfg := &config.Config{
		Environment:    "production",
		DeploymentMode: config.ModeOnPremise,
	}
	gate := NewGateHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ServePortalOrRedirect(cfg, gate)(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("GET / in prod on-premise status = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/app" {
		t.Fatalf("Location = %q, want /app", loc)
	}
}

func TestPublicLeadHandler_ValidationAndSuccess(t *testing.T) {
	handler := NewPublicLeadHandler(nil)

	// Rejet email invalide
	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/lead", strings.NewReader(`{"lead_type":"demo","full_name":"Dr. Amine","email":"invalid"}`))
	badRec := httptest.NewRecorder()
	handler.SubmitLead(badRec, badReq)
	if badRec.Code != http.StatusBadRequest {
		t.Fatalf("invalid email status = %d, want 400", badRec.Code)
	}

	// Succès
	okReq := httptest.NewRequest(http.MethodPost, "/api/v1/public/lead", strings.NewReader(`{"lead_type":"demo","full_name":"Dr. Amine","email":"amine@pharma.dz","organization_name":"Labo Alger","language":"fr"}`))
	okRec := httptest.NewRecorder()
	handler.SubmitLead(okRec, okReq)
	if okRec.Code != http.StatusCreated {
		t.Fatalf("valid lead status = %d, want 201", okRec.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(okRec.Body.Bytes(), &resp); err != nil || resp["success"] != true {
		t.Fatalf("unexpected lead response: %s", okRec.Body.String())
	}
}

func TestPortal_SEO_RobotsAndSitemap(t *testing.T) {
	// Sas actif : robots.txt doit bloquer et sitemap.xml doit retourner 404
	lockedCfg := &config.Config{Environment: "production", DeploymentMode: config.ModeSaaS, GatePasscode: "Homolog2006!VIP"}
	recRobotsLocked := httptest.NewRecorder()
	ServeRobotsTxt(lockedCfg)(recRobotsLocked, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	if !strings.Contains(recRobotsLocked.Body.String(), "Disallow: /") {
		t.Fatalf("locked robots.txt = %q", recRobotsLocked.Body.String())
	}
	recMapLocked := httptest.NewRecorder()
	ServeSitemapXML(lockedCfg)(recMapLocked, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if recMapLocked.Code != http.StatusNotFound {
		t.Fatalf("locked sitemap status = %d, want 404", recMapLocked.Code)
	}

	// Public (sas désactivé en SaaS) : robots.txt autorise / et bloque /app, sitemap.xml retourne 200
	pubCfg := &config.Config{Environment: "production", DeploymentMode: config.ModeSaaS}
	recRobotsPub := httptest.NewRecorder()
	reqRobots := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	reqRobots.Host = "sanad-pharma.dz"
	ServeRobotsTxt(pubCfg)(recRobotsPub, reqRobots)
	if !strings.Contains(recRobotsPub.Body.String(), "Disallow: /app") {
		t.Fatalf("public robots.txt = %q", recRobotsPub.Body.String())
	}
	recMapPub := httptest.NewRecorder()
	reqMap := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	reqMap.Host = "sanad-pharma.dz"
	reqMap.Header.Set("X-Forwarded-Proto", "https")
	ServeSitemapXML(pubCfg)(recMapPub, reqMap)
	if recMapPub.Code != http.StatusOK || !strings.Contains(recMapPub.Body.String(), "https://sanad-pharma.dz/") {
		t.Fatalf("public sitemap status=%d body=%q", recMapPub.Code, recMapPub.Body.String())
	}
}
