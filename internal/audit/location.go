package audit

import (
	"net"
	"strings"
)

type ipLocationRule struct {
	cidr     *net.IPNet
	location string
}

var parsedRules []ipLocationRule

func init() {
	rules := []struct {
		cidrStr  string
		location string
	}{
		// Sites Pharmaceutiques Privés & Locaux (Intranet GxP)
		{"192.168.1.0/24", "🏢 Siège Administratif & RA (Alger)"},
		{"192.168.10.0/24", "🏭 Usine & Production (Oued Smar)"},
		{"192.168.20.0/24", "🔬 Labo Contrôle Qualité LCQ (Alger)"},
		{"192.168.30.0/24", "📦 Magasin Central & Logistique (Baraki)"},
		{"192.168.0.0/16", "🏢 Réseau Intranet Siège (Alger)"},
		{"10.10.0.0/16", "🏭 Usine & Production (Oued Smar)"},
		{"10.20.0.0/16", "🔬 Labo Contrôle Qualité LCQ (Alger)"},
		{"10.30.0.0/16", "📦 Magasin Central & Logistique (Baraki)"},
		{"10.0.0.0/8", "🏢 Réseau Intranet GxP (Site Industriel)"},
		{"172.16.0.0/12", "🏢 Réseau Sécurisé Datacenter GxP"},

		// Opérateurs et Réseaux Institutionnels Algériens
		{"193.194.64.0/18", "🇩🇿 CERIST / Réseau Ministériel (Alger)"},
		{"41.200.0.0/15", "🇩🇿 Alger (Algérie Télécom)"},
		{"41.220.0.0/16", "🇩🇿 Oran (Algérie)"},
		{"41.221.0.0/16", "🇩🇿 Constantine (Algérie)"},
		{"41.0.0.0/8", "🇩🇿 Alger, Algérie"},
		{"105.96.0.0/12", "🇩🇿 Alger, Algérie (Fibre/4G)"},
		{"105.0.0.0/8", "🇩🇿 Alger, Algérie"},
		{"154.0.0.0/8", "🇩🇿 Alger, Algérie (Mobilis/Djezzy)"},
		{"197.112.0.0/12", "🇩🇿 Alger, Algérie (Ooredoo)"},
		{"197.0.0.0/8", "🇩🇿 Alger, Algérie"},

		// Réseaux Partenaires B2B / Titulaires API / Auditeurs Étrangers
		{"194.0.0.0/8", "🇫🇷 Paris, France (Partenaire B2B)"},
		{"195.0.0.0/8", "🇨🇭 Bâle, Suisse (Titulaire API)"},
		{"165.0.0.0/8", "🇩🇪 Francfort, Allemagne (Auditeur Qualité)"},
		{"115.0.0.0/8", "🇮🇳 Hyderabad, Inde (Fournisseur API)"},
		{"103.0.0.0/8", "🇮🇳 Hyderabad, Inde (Fournisseur API)"},
		{"185.0.0.0/8", "🇪🇺 Europe (Accès Distant Sécurisé)"},
	}

	for _, r := range rules {
		_, ipNet, err := net.ParseCIDR(r.cidrStr)
		if err == nil {
			parsedRules = append(parsedRules, ipLocationRule{
				cidr:     ipNet,
				location: r.location,
			})
		}
	}
}

// ResolveLocation détermine le lieu géographique et physique à partir de l'IP,
// du User-Agent et des métadonnées contextuelles.
func ResolveLocation(ipAddress, userAgent string, details map[string]any) string {
	// 1. Vérifier si un lieu explicite est présent dans les métadonnées
	if details != nil {
		if loc, ok := details["location"].(string); ok && strings.TrimSpace(loc) != "" {
			return strings.TrimSpace(loc)
		}
		if site, ok := details["site"].(string); ok && strings.TrimSpace(site) != "" {
			return strings.TrimSpace(site)
		}
	}

	cleanIP := strings.TrimSpace(ipAddress)
	if cleanIP == "" {
		return "💻 Console Locale Serveur (Sur Site)"
	}

	// Supprimer le port éventuel (ex: "127.0.0.1:54321" ou "[::1]:8080")
	if host, _, err := net.SplitHostPort(cleanIP); err == nil {
		cleanIP = host
	}

	// Vérifier bouclage local
	if cleanIP == "127.0.0.1" || cleanIP == "::1" || cleanIP == "localhost" || cleanIP == "0.0.0.0" {
		return "💻 Console Locale Serveur (Sur Site)"
	}

	parsedIP := net.ParseIP(cleanIP)
	if parsedIP == nil {
		return "📍 Emplacement Inconnu"
	}

	// Correspondance avec les règles CIDR ordonnées
	for _, rule := range parsedRules {
		if rule.cidr.Contains(parsedIP) {
			return rule.location
		}
	}

	// Détection générale selon classe d'adresses
	if parsedIP.IsLoopback() {
		return "💻 Console Locale Serveur (Sur Site)"
	}
	if parsedIP.IsPrivate() {
		return "🏢 Réseau Intranet Privé GxP"
	}

	return "🌍 Connexion Distante Sécurisée"
}
