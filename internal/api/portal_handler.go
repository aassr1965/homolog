package api

import (
	"net/http"
	"strings"

	"homolog/internal/config"
	"homolog/web"
)

// PortalEnabled indique si le portail vitrine SPA à la racine "/" doit être affiché.
// En développement, il est toujours actif (SaaS et On-Premise).
// En production On-Premise, il est désactivé et "/" redirige directement vers "/app".
func PortalEnabled(cfg *config.Config) bool {
	if cfg == nil {
		return true
	}
	if cfg.Environment == "production" && cfg.DeploymentMode == config.ModeOnPremise {
		return false
	}
	return true
}

func serveGateHTML(w http.ResponseWriter) {
	gateContent, err := web.Files.ReadFile("gate.html")
	if err != nil {
		http.Error(w, "Page du sas confidentiel introuvable", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	_, _ = w.Write(gateContent)
}

// ServePortalOrRedirect sert le portail SPA SANAD Pharma à la racine "/",
// protégé par le sas confidentiel si actif, ou redirige vers "/app" en production On-Premise.
func ServePortalOrRedirect(cfg *config.Config, gateHandler *GateHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg != nil && cfg.GateEnabled() && gateHandler != nil && !gateHandler.IsUnlocked(r) {
			serveGateHTML(w)
			return
		}

		if !PortalEnabled(cfg) {
			http.Redirect(w, r, "/app", http.StatusFound)
			return
		}

		content, err := web.Files.ReadFile("portal.html")
		if err != nil {
			http.Error(w, "Portail SANAD Pharma introuvable", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		_, _ = w.Write(content)
	}
}

// ServeApp sert l'application GxP 'Pharma Explorer' (écran d'authentification & cockpit) sous "/app",
// protégée par le sas confidentiel si actif.
func ServeApp(cfg *config.Config, gateHandler *GateHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg != nil && cfg.GateEnabled() && gateHandler != nil && !gateHandler.IsUnlocked(r) {
			serveGateHTML(w)
			return
		}

		content, err := web.Files.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Interface web non trouvée", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		if cfg != nil && !cfg.PitchDeckEnabled() {
			htmlStr := string(content)
			// En production, masquer immédiatement le bouton du pitch commercial (Défense en profondeur)
			htmlStr = strings.ReplaceAll(htmlStr, `id="btn-pitch-commercial"`, `id="btn-pitch-commercial" style="display:none !important;"`)
			_, _ = w.Write([]byte(htmlStr))
			return
		}
		_, _ = w.Write(content)
	}
}

// ServeRobotsTxt génère dynamiquement /robots.txt selon l'état du sas confidentiel et du portail.
func ServeRobotsTxt(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if cfg != nil && (cfg.GateEnabled() || !PortalEnabled(cfg)) {
			_, _ = w.Write([]byte("User-agent: *\nDisallow: /\n"))
			return
		}
		scheme := "https"
		if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
			scheme = "http"
		}
		host := r.Host
		if host == "" {
			host = "sanad-pharma.dz"
		}
		_, _ = w.Write([]byte("User-agent: *\nAllow: /\nDisallow: /app\nDisallow: /api/\nDisallow: /verify/\n\nSitemap: " + scheme + "://" + host + "/sitemap.xml\n"))
	}
}

// ServeSitemapXML génère le sitemap XML du portail public SANAD PHARMA.
func ServeSitemapXML(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg != nil && (cfg.GateEnabled() || !PortalEnabled(cfg)) {
			http.NotFound(w, r)
			return
		}
		scheme := "https"
		if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
			scheme = "http"
		}
		host := r.Host
		if host == "" {
			host = "sanad-pharma.dz"
		}
		base := scheme + "://" + host
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>` + base + `/</loc><changefreq>weekly</changefreq><priority>1.0</priority></url>
</urlset>
`))
	}
}
