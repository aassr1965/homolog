package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"homolog/internal/config"
)

// GateCookieName est le nom du cookie HttpOnly mémorisant l'autorisation du sas confidentiel
const GateCookieName = "homolog_gate_access"

// GateHandler gère le sas d'accès confidentiel (Private Preview Gate)
type GateHandler struct {
	cfg *config.Config
}

// NewGateHandler instancie le gestionnaire du sas confidentiel
func NewGateHandler(cfg *config.Config) *GateHandler {
	return &GateHandler{cfg: cfg}
}

// computeGateToken génère un jeton cryptographique HMAC-SHA256 scellé par le secret d'instance
func (h *GateHandler) computeGateToken() string {
	if h.cfg == nil || !h.cfg.GateEnabled() {
		return ""
	}
	secret := h.cfg.JWTSecret
	if secret == "" {
		secret = "homolog-gate-private-key-2026"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("homolog_gate_v1:" + strings.TrimSpace(h.cfg.GatePasscode)))
	return hex.EncodeToString(mac.Sum(nil))
}

// IsUnlocked vérifie si la requête entrante possède le cookie d'autorisation valide
func (h *GateHandler) IsUnlocked(r *http.Request) bool {
	if h.cfg == nil || !h.cfg.GateEnabled() {
		return true
	}
	cookie, err := r.Cookie(GateCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	expected := h.computeGateToken()
	if expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(expected)) == 1
}

// GateStatusResponse décrit l'état d'activation et de déverrouillage du sas
type GateStatusResponse struct {
	GateRequired bool   `json:"gate_required"`
	Unlocked     bool   `json:"unlocked"`
	AppName      string `json:"app_name"`
	OrgName      string `json:"org_name"`
	Environment  string `json:"environment"`
}

// GetStatus retourne l'état actuel du sas pour l'interface client
func (h *GateHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	resp := GateStatusResponse{
		GateRequired: h.cfg.GateEnabled(),
		Unlocked:     h.IsUnlocked(r),
		AppName:      "Homolog",
		OrgName:      "Sanad Pharma",
		Environment:  "development",
	}
	if h.cfg != nil {
		if h.cfg.OnPremOrgName != "" {
			resp.OrgName = h.cfg.OnPremOrgName
		}
		if h.cfg.Environment != "" {
			resp.Environment = h.cfg.Environment
		}
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// GateUnlockRequest porte le mot de passe unique saisi par le client
type GateUnlockRequest struct {
	Passcode string `json:"passcode"`
}

// Unlock valide le mot de passe partenaire et pose le cookie sécurisé (30 jours)
func (h *GateHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.cfg == nil || !h.cfg.GateEnabled() {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"message": "Le sas d'accès confidentiel n'est pas activé",
		})
		return
	}

	var req GateUnlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	inputPasscode := strings.TrimSpace(req.Passcode)
	expectedPasscode := strings.TrimSpace(h.cfg.GatePasscode)

	// Comparaison sécurisée en temps constant (résistance aux attaques par canal auxiliaire)
	if subtle.ConstantTimeCompare([]byte(inputPasscode), []byte(expectedPasscode)) != 1 {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": "Code d'accès partenaire invalide. Veuillez vérifier le code auprès de votre correspondant.",
		})
		return
	}

	// Code valide -> Pose du Cookie HttpOnly (30 jours)
	token := h.computeGateToken()
	http.SetCookie(w, &http.Cookie{
		Name:     GateCookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		MaxAge:   30 * 24 * 3600,
		HttpOnly: true,
		Secure:   DetermineCookieSecure(r, h.cfg),
		SameSite: http.SameSiteStrictMode,
	})

	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Accès partenaire déverrouillé avec succès",
	})
}

// Lock révoque l'autorisation du sas en supprimant le cookie
func (h *GateHandler) Lock(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     GateCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   DetermineCookieSecure(r, h.cfg),
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Sas confidentiel reverrouillé avec succès",
	})
}

// Middleware bloque l'accès aux API métier si le sas est activé et verrouillé
func (h *GateHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.cfg != nil && h.cfg.GateEnabled() && !h.IsUnlocked(r) {
			path := r.URL.Path

			// Autoriser les endpoints du sas lui-même
			if strings.HasPrefix(path, "/api/v1/gate/") {
				next.ServeHTTP(w, r)
				return
			}

			// Pour toute autre requête sous /api/v1/*, bloquer immédiatement en 403 Forbidden
			if strings.HasPrefix(path, "/api/v1/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": "Accès restreint : Le sas d'accès confidentiel est actif. Veuillez entrer votre code d'accès partenaire.",
					"code":  "GATE_LOCKED",
				})
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
