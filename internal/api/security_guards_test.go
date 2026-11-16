package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/config"
)

func TestDetermineCookieSecure(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		reqSetup func() *http.Request
		expected bool
	}{
		{
			name: "Plain HTTP in Development returns false",
			cfg: &config.Config{
				Environment:  "development",
				CookieSecure: false,
			},
			reqSetup: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "http://localhost:8090/", nil)
			},
			expected: false,
		},
		{
			name: "Direct TLS/HTTPS in Development returns true",
			cfg: &config.Config{
				Environment:  "development",
				CookieSecure: false,
			},
			reqSetup: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "https://localhost:8090/", nil)
				req.TLS = &tls.ConnectionState{}
				return req
			},
			expected: true,
		},
		{
			name: "X-Forwarded-Proto https in Development returns true",
			cfg: &config.Config{
				Environment:  "development",
				CookieSecure: false,
			},
			reqSetup: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "http://localhost:8090/", nil)
				req.Header.Set("X-Forwarded-Proto", "https")
				return req
			},
			expected: true,
		},
		{
			name: "X-Forwarded-Ssl on in Development returns true",
			cfg: &config.Config{
				Environment:  "development",
				CookieSecure: false,
			},
			reqSetup: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "http://localhost:8090/", nil)
				req.Header.Set("X-Forwarded-Ssl", "on")
				return req
			},
			expected: true,
		},
		{
			name: "Production environment ALWAYS returns true even over plain HTTP",
			cfg: &config.Config{
				Environment:  "production",
				CookieSecure: false,
			},
			reqSetup: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "http://homolog.pharma.dz/", nil)
			},
			expected: true,
		},
		{
			name: "CookieSecure explicitly true returns true",
			cfg: &config.Config{
				Environment:  "development",
				CookieSecure: true,
			},
			reqSetup: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "http://localhost:8090/", nil)
			},
			expected: true,
		},
		{
			name: "Nil config with X-Forwarded-Proto https returns true",
			cfg:  nil,
			reqSetup: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "http://localhost:8090/", nil)
				req.Header.Set("X-Forwarded-Proto", "https")
				return req
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.reqSetup()
			actual := DetermineCookieSecure(req, tc.cfg)
			if actual != tc.expected {
				t.Errorf("Expected DetermineCookieSecure = %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestDevSwitchRole_ProductionBlocking(t *testing.T) {
	// 1. Configured as production
	prodCfg := &config.Config{
		Environment:       "production",
		EnableDevSwitcher: false,
	}
	handlerProd := NewAdminHandler(nil, nil, nil, prodCfg)

	reqBody := []byte(`{"email":"admin.lab@pharma.dz"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/dev-switch", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()

	handlerProd.DevSwitchRole(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected HTTP 404 in production, got %d", rr.Code)
	}

	var resp map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if !strings.Contains(resp["error"], "production") {
		t.Errorf("Expected error to mention production, got %q", resp["error"])
	}

	// 2. Configured as development but EnableDevSwitcher=false
	devDisabledCfg := &config.Config{
		Environment:       "development",
		EnableDevSwitcher: false,
	}
	handlerDevDisabled := NewAdminHandler(nil, nil, nil, devDisabledCfg)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/dev-switch", bytes.NewReader(reqBody))
	rr2 := httptest.NewRecorder()

	handlerDevDisabled.DevSwitchRole(rr2, req2)

	if rr2.Code != http.StatusNotFound {
		t.Errorf("Expected HTTP 404 when EnableDevSwitcher=false, got %d", rr2.Code)
	}

	// 3. Environment variable APP_ENV=production fallback guard
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")

	handlerFallback := NewAdminHandler(nil, nil, nil, nil)
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/dev-switch", bytes.NewReader(reqBody))
	rr3 := httptest.NewRecorder()

	handlerFallback.DevSwitchRole(rr3, req3)

	if rr3.Code != http.StatusNotFound {
		t.Errorf("Expected HTTP 404 when APP_ENV=production, got %d", rr3.Code)
	}
}

func TestLoginAndLogout_CookieAttributes(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	ctx := t.Context()
	sm := auth.NewSessionManager(db.Pool, 60, 15)
	auditSvc := audit.NewService(db.Pool)

	// Create test organization and user
	testOrgID := "11111111-2222-3333-4444-555555555555"
	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, slug, legal_name, miph_license_number, technical_director_name)
		VALUES ($1, 'org-cookie-sec-test', 'Labo Cookie Security', 'MIPH-COOKIE', 'Dr. Secure')
		ON CONFLICT (id) DO NOTHING
	`, testOrgID)

	testUserID := "22222222-3333-4444-5555-666666666666"
	testEmail := "secure-cookie-test@homolog.dz"
	pwd := "P@ssw0rdPharma2026!"
	hash, err := auth.HashPassword(pwd, auth.DefaultArgon2Params)
	if err != nil {
		t.Fatalf("Erreur hash: %v", err)
	}

	_, _ = db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, role)
		VALUES ($1, $2, $3, $4, 'Dr', 'Secure', 'directeur_technique')
		ON CONFLICT (id) DO UPDATE SET password_hash = $4, is_active = TRUE, failed_login_attempts = 0, locked_until = NULL
	`, testUserID, testOrgID, testEmail, hash)

	defer func() {
		_, _ = db.Pool.Exec(ctx, "DELETE FROM users WHERE id = $1", testUserID)
		_, _ = db.Pool.Exec(ctx, "DELETE FROM organizations WHERE id = $1", testOrgID)
	}()

	loginBody, _ := json.Marshal(LoginRequest{
		Email:    testEmail,
		Password: pwd,
	})

	// 1. Test Login in Production -> Must have Secure=true in Cookie
	t.Run("Login in Production sets Secure cookie", func(t *testing.T) {
		prodCfg := &config.Config{
			Environment:  "production",
			CookieSecure: true,
		}
		handler := NewAuthHandler(db.Pool, sm, auditSvc, prodCfg)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
		rr := httptest.NewRecorder()
		handler.Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Login failed with status %d: %s", rr.Code, rr.Body.String())
		}

		cookies := rr.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == CookieName {
				sessionCookie = c
				break
			}
		}

		if sessionCookie == nil {
			t.Fatalf("Expected session cookie %q to be set", CookieName)
		}
		if !sessionCookie.HttpOnly {
			t.Errorf("Expected HttpOnly = true, got %v", sessionCookie.HttpOnly)
		}
		if !sessionCookie.Secure {
			t.Errorf("Expected Secure = true in production, got %v", sessionCookie.Secure)
		}
		if sessionCookie.SameSite != http.SameSiteStrictMode {
			t.Errorf("Expected SameSite = Strict, got %v", sessionCookie.SameSite)
		}

		// Verify Logout also maintains Secure attribute
		logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		logoutReq.AddCookie(sessionCookie)
		logoutRR := httptest.NewRecorder()
		handler.Logout(logoutRR, logoutReq)

		logoutCookies := logoutRR.Result().Cookies()
		var logoutCookie *http.Cookie
		for _, c := range logoutCookies {
			if c.Name == CookieName {
				logoutCookie = c
				break
			}
		}
		if logoutCookie == nil {
			t.Fatalf("Expected session cookie in logout response")
		}
		if !logoutCookie.Secure {
			t.Errorf("Expected Secure = true on logout cookie in production")
		}
	})

	// 2. Test Login in Development over plain HTTP -> Secure=false
	t.Run("Login in Development over plain HTTP sets Secure=false", func(t *testing.T) {
		devCfg := &config.Config{
			Environment:  "development",
			CookieSecure: false,
		}
		handler := NewAuthHandler(db.Pool, sm, auditSvc, devCfg)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
		rr := httptest.NewRecorder()
		handler.Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Login failed: %d", rr.Code)
		}

		cookies := rr.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == CookieName {
				sessionCookie = c
				break
			}
		}

		if sessionCookie == nil {
			t.Fatalf("Expected session cookie")
		}
		if sessionCookie.Secure {
			t.Errorf("Expected Secure = false on plain HTTP in dev, got true")
		}
	})

	// 3. Test Login in Development behind HTTPS reverse proxy -> Secure=true
	t.Run("Login in Development behind HTTPS reverse proxy sets Secure=true", func(t *testing.T) {
		devCfg := &config.Config{
			Environment:  "development",
			CookieSecure: false,
		}
		handler := NewAuthHandler(db.Pool, sm, auditSvc, devCfg)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
		req.Header.Set("X-Forwarded-Proto", "https")
		rr := httptest.NewRecorder()
		handler.Login(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Login failed: %d", rr.Code)
		}

		cookies := rr.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == CookieName {
				sessionCookie = c
				break
			}
		}

		if sessionCookie == nil {
			t.Fatalf("Expected session cookie")
		}
		if !sessionCookie.Secure {
			t.Errorf("Expected Secure = true when X-Forwarded-Proto is https")
		}
	})
}

func TestPitchDeck_ProductionGuard(t *testing.T) {
	// 1. Production config must never enable pitch deck
	prodCfg := &config.Config{
		Environment:     "production",
		EnablePitchDeck: false,
	}
	if prodCfg.PitchDeckEnabled() {
		t.Errorf("Expected PitchDeckEnabled() = false in production")
	}

	// Even if someone mistakenly sets EnablePitchDeck = true in production, PitchDeckEnabled() must block it
	prodDangerousCfg := &config.Config{
		Environment:     "production",
		EnablePitchDeck: true,
	}
	if prodDangerousCfg.PitchDeckEnabled() {
		t.Errorf("Expected PitchDeckEnabled() = false in production even if EnablePitchDeck is true")
	}

	prodAliasCfg := &config.Config{
		Environment:     "prod",
		EnablePitchDeck: true,
	}
	if prodAliasCfg.PitchDeckEnabled() {
		t.Errorf("Expected PitchDeckEnabled() = false when Environment is 'prod'")
	}

	// 2. Development config enables pitch deck when allowed
	devCfg := &config.Config{
		Environment:     "development",
		EnablePitchDeck: true,
	}
	if !devCfg.PitchDeckEnabled() {
		t.Errorf("Expected PitchDeckEnabled() = true in development")
	}

	devDisabledCfg := &config.Config{
		Environment:     "development",
		EnablePitchDeck: false,
	}
	if devDisabledCfg.PitchDeckEnabled() {
		t.Errorf("Expected PitchDeckEnabled() = false when EnablePitchDeck is false")
	}

	// 3. Fallback via APP_ENV environment variable
	t.Run("APP_ENV=production fallback", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		var nilCfg *config.Config
		if nilCfg.PitchDeckEnabled() {
			t.Errorf("Expected nil config with APP_ENV=production to return false")
		}
	})

	// 4. Test Route level blocking in production
	t.Run("Route level blocking in production", func(t *testing.T) {
		r := chi.NewRouter()
		if prodCfg.PitchDeckEnabled() {
			r.Get("/presentation", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("presentation"))
			})
		} else {
			r.Get("/presentation", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Ressource désactivée ou indisponible en environnement de production", http.StatusNotFound)
			})
			r.Get("/pitch", func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Ressource désactivée ou indisponible en environnement de production", http.StatusNotFound)
			})
		}

		req1 := httptest.NewRequest(http.MethodGet, "/presentation", nil)
		rr1 := httptest.NewRecorder()
		r.ServeHTTP(rr1, req1)
		if rr1.Code != http.StatusNotFound {
			t.Errorf("Expected /presentation to return 404 in production, got %d", rr1.Code)
		}

		req2 := httptest.NewRequest(http.MethodGet, "/pitch", nil)
		rr2 := httptest.NewRecorder()
		r.ServeHTTP(rr2, req2)
		if rr2.Code != http.StatusNotFound {
			t.Errorf("Expected /pitch to return 404 in production, got %d", rr2.Code)
		}
	})
}
