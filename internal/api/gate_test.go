package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"homolog/internal/api"
	"homolog/internal/config"
	"homolog/web"
)

func setupGateTestRouter(cfg *config.Config) (*chi.Mux, *api.GateHandler) {
	gateHandler := api.NewGateHandler(cfg)
	r := chi.NewRouter()
	r.Use(gateHandler.Middleware)

	// Route racine avec sas hermétique (Option B)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if cfg != nil && cfg.GateEnabled() && !gateHandler.IsUnlocked(r) {
			gateContent, err := web.Files.ReadFile("gate.html")
			if err != nil {
				http.Error(w, "gate.html introuvable", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(gateContent)
			return
		}

		content, err := web.Files.ReadFile("index.html")
		if err != nil {
			http.Error(w, "index.html introuvable", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	})

	gateAssetGuard := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg != nil && cfg.GateEnabled() && !gateHandler.IsUnlocked(r) {
				http.Error(w, "Accès refusé : Sas d'accès partenaire non déverrouillé", http.StatusForbidden)
				return
			}
			h.ServeHTTP(w, r)
		})
	}

	staticFS := http.FileServer(http.FS(web.Files))
	r.Handle("/js/*", gateAssetGuard(staticFS))
	r.Handle("/css/*", staticFS)

	r.Route("/api/v1/gate", func(r chi.Router) {
		r.Get("/status", gateHandler.GetStatus)
		r.Post("/unlock", gateHandler.Unlock)
		r.Post("/lock", gateHandler.Lock)
	})

	return r, gateHandler
}

func TestGateHermeticProtectionAndRateLimit(t *testing.T) {
	cfg := &config.Config{
		GatePasscode: "SecretPartenaire2026!",
		JWTSecret:    "test-jwt-secret-gate-12345",
	}

	router, _ := setupGateTestRouter(cfg)

	// 1. Visiteur sans cookie interrogeant GET /
	t.Run("01_Unauthenticated_Visitor_Receives_Only_GateHTML", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Attendu HTTP 200 pour gate.html, reçu: %d", rec.Code)
		}

		body := rec.Body.String()
		// Doit contenir le sas partenaire
		if !strings.Contains(body, "Sas d'Accès Partenaire Confidentiel") {
			t.Fatalf("Le corps devrait contenir le titre du sas, reçu:\n%s", body[:min(300, len(body))])
		}

		// NE DOIT STRICTEMENT PAS contenir les identifiants ou le mot de passe
		forbiddenTerms := []string{
			"screen-login",
			"pdt@pharma.dz",
			"dt@sanadpharma.dz",
			"Homolog2026!Dev",
			"Homolog2026!VIP",
			"Directeur Technique (1.A.7)",
		}
		for _, term := range forbiddenTerms {
			if strings.Contains(body, term) {
				t.Fatalf("ÉCHEC DE SÉCURITÉ : Le terme confidentiel [%s] a été trouvé dans gate.html !", term)
			}
		}
	})

	// 2. Tentative d'accès direct au JS sans déverrouillage préalable
	t.Run("02_Direct_Access_To_JS_Assets_Blocked_403", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/js/app.js", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("Attendu HTTP 403 Forbidden sur /js/app.js, reçu: %d", rec.Code)
		}
	})

	// 3. Brute Force Protection (Rate Limit après 5 échecs consécutifs)
	t.Run("03_Brute_Force_Protection_Rate_Limit", func(t *testing.T) {
		ip := "192.168.1.100:45678"

		// 5 tentatives erronées
		for i := 1; i <= 5; i++ {
			body, _ := json.Marshal(map[string]string{"passcode": "MauvaisCode123"})
			req := httptest.NewRequest("POST", "/api/v1/gate/unlock", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.RemoteAddr = ip

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("Tentative %d : Attendu HTTP 401 Unauthorized, reçu: %d", i, rec.Code)
			}
		}

		// 6ème tentative (même avec le bon code) -> Doit être bloquée en 429 Too Many Requests
		body, _ := json.Marshal(map[string]string{"passcode": cfg.GatePasscode})
		req := httptest.NewRequest("POST", "/api/v1/gate/unlock", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = ip

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusTooManyRequests {
			t.Fatalf("Tentative 6 : Attendu HTTP 429 Too Many Requests, reçu: %d (corps: %s)", rec.Code, rec.Body.String())
		}
	})

	// 4. Déverrouillage légitime avec le bon code partenaire (depuis une autre IP)
	var gateCookie *http.Cookie
	t.Run("04_Legitimate_Unlock_Sets_HttpOnly_Cookie", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"passcode": cfg.GatePasscode})
		req := httptest.NewRequest("POST", "/api/v1/gate/unlock", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "10.0.0.50:12345"

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Attendu HTTP 200 OK pour le déverrouillage, reçu: %d", rec.Code)
		}

		for _, cookie := range rec.Result().Cookies() {
			if cookie.Name == api.GateCookieName {
				gateCookie = cookie
				break
			}
		}

		if gateCookie == nil {
			t.Fatalf("Le cookie %s n'a pas été posé par le serveur !", api.GateCookieName)
		}
		if !gateCookie.HttpOnly {
			t.Fatalf("Le cookie de sas doit impérativement être HttpOnly")
		}
		if gateCookie.SameSite != http.SameSiteStrictMode {
			t.Fatalf("Le cookie de sas doit être SameSite=Strict")
		}
	})

	// 5. Requête GET / avec le cookie de sas -> Reçoit index.html
	t.Run("05_Authenticated_Request_Receives_IndexHTML", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(gateCookie)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Attendu HTTP 200 pour index.html, reçu: %d", rec.Code)
		}

		body := rec.Body.String()
		// Doit contenir le cockpit applicatif
		if !strings.Contains(body, "screen-login") {
			t.Fatalf("Le corps devrait contenir l'écran de login une fois déverrouillé")
		}

		// Vérification d'absence de mot de passe en clair dans le HTML
		if strings.Contains(body, "Homolog2026!Dev") {
			t.Fatalf("Le mot de passe de test en clair Homolog2026!Dev ne doit plus figurer dans index.html !")
		}
	})

	// 6. Requête /js/app.js avec le cookie de sas -> Autorisé
	t.Run("06_Authenticated_Request_Can_Download_JS", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/js/app.js", nil)
		req.AddCookie(gateCookie)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Attendu HTTP 200 pour /js/app.js avec cookie, reçu: %d", rec.Code)
		}
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
