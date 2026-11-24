package audit

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ipLocationRule struct {
	cidr     *net.IPNet
	location string
}

var (
	parsedRules []ipLocationRule
	geoIPCache  sync.Map
	httpClient  = &http.Client{
		Timeout: 2 * time.Second,
	}
)

var frenchCountryNames = map[string]string{
	"Algeria":              "Algérie",
	"France":               "France",
	"Netherlands":          "Pays-Bas",
	"United States":        "États-Unis",
	"Germany":              "Allemagne",
	"Switzerland":          "Suisse",
	"United Kingdom":       "Royaume-Uni",
	"Spain":                "Espagne",
	"Italy":                "Italie",
	"Belgium":              "Belgique",
	"Canada":               "Canada",
	"India":                "Inde",
	"Tunisia":              "Tunisie",
	"Morocco":              "Maroc",
	"Turkey":               "Turquie",
	"Sweden":               "Suède",
	"Norway":               "Norvège",
	"Japan":                "Japon",
	"China":                "Chine",
	"Singapore":            "Singapour",
	"Ukraine":              "Ukraine",
	"Poland":               "Pologne",
	"Ireland":              "Irlande",
	"Portugal":             "Portugal",
	"Austria":              "Autriche",
	"Australia":            "Australie",
	"United Arab Emirates": "Émirats Arabes Unis",
	"Saudi Arabia":         "Arabie Saoudite",
}

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

		// Plages de test et validation BPF (RFC 5737)
		{"192.0.2.0/24", "🔬 Réseau Test BPF / Validation"},
		{"198.51.100.0/24", "🔬 Réseau Test BPF / Validation"},
		{"203.0.113.0/24", "🔬 Réseau Test BPF / Validation"},

		// Réseaux Partenaires B2B / Titulaires API / Auditeurs Étrangers connus
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

func countryCodeToFlag(countryCode string) string {
	cc := strings.ToUpper(strings.TrimSpace(countryCode))
	if len(cc) != 2 || cc[0] < 'A' || cc[0] > 'Z' || cc[1] < 'A' || cc[1] > 'Z' {
		return "🌍"
	}
	r1 := rune(0x1F1E6 + int(cc[0]-'A'))
	r2 := rune(0x1F1E6 + int(cc[1]-'A'))
	return string([]rune{r1, r2})
}

func isVPNOrHosting(isp, org string) bool {
	combined := strings.ToLower(isp + " " + org)
	keywords := []string{
		"vpn", "datacenter", "data center", "hosting", "datacamp", "m247",
		"ovh", "hetzner", "digitalocean", "linode", "aws", "amazon", "google cloud",
		"azure", "microsoft", "leaseweb", "choopa", "vultr", "packetexchange",
		"private internet", "nord", "proton", "expressvpn", "surfshark", "cyberghost",
		"tunnel", "proxy", "tor", "cloud", "server",
	}
	for _, kw := range keywords {
		if strings.Contains(combined, kw) {
			return true
		}
	}
	return false
}

type geoIPResponse struct {
	Status      string `json:"status"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	City        string `json:"city"`
	ISP         string `json:"isp"`
	Org         string `json:"org"`
}

type bdcLocalityResponse struct {
	CountryName          string `json:"countryName"`
	CountryCode          string `json:"countryCode"`
	PrincipalSubdivision string `json:"principalSubdivision"`
	City                 string `json:"city"`
	Locality             string `json:"locality"`
}

func queryGeoIP(ipStr string) string {
	if val, ok := geoIPCache.Load(ipStr); ok {
		return val.(string)
	}

	// 1. Tenter la résolution haute précision (Commune / Daïra / Ville via BigDataCloud)
	bdcURL := fmt.Sprintf("https://api.bigdatacloud.net/data/reverse-geocode-client?ip=%s&localityLanguage=fr", ipStr)
	var (
		countryCode string
		countryName string
		locality    string
		city        string
	)

	bdcResp, err := httpClient.Get(bdcURL)
	if err == nil && bdcResp.StatusCode == http.StatusOK {
		var bdc bdcLocalityResponse
		if json.NewDecoder(bdcResp.Body).Decode(&bdc) == nil && bdc.CountryCode != "" {
			countryCode = bdc.CountryCode
			countryName = bdc.CountryName
			locality = strings.TrimSpace(bdc.Locality)
			city = strings.TrimSpace(bdc.City)
			if city == "" {
				city = strings.TrimSpace(bdc.PrincipalSubdivision)
			}
			city = strings.TrimPrefix(city, "wilaya d'")
			city = strings.TrimPrefix(city, "wilaya de ")
			city = strings.TrimPrefix(city, "Wilaya d'")
			city = strings.TrimPrefix(city, "Wilaya de ")
		}
		bdcResp.Body.Close()
	}

	// 2. Interroger ip-api pour ISP et détection VPN
	apiURL := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,city,countryCode,isp,org", ipStr)
	var (
		isVPN   bool
		ispName string
	)
	apiResp, err := httpClient.Get(apiURL)
	if err == nil && apiResp.StatusCode == http.StatusOK {
		var data geoIPResponse
		if json.NewDecoder(apiResp.Body).Decode(&data) == nil && data.Status == "success" {
			if countryCode == "" {
				countryCode = data.CountryCode
				countryName = data.Country
			}
			if city == "" {
				city = strings.TrimSpace(data.City)
			}
			ispName = data.ISP
			isVPN = isVPNOrHosting(data.ISP, data.Org)
		}
		apiResp.Body.Close()
	}

	if countryCode == "" && city == "" {
		fallback := fmt.Sprintf("🌍 Accès Distant (%s)", ipStr)
		geoIPCache.Store(ipStr, fallback)
		return fallback
	}

	flag := countryCodeToFlag(countryCode)
	country := countryName
	if frName, exists := frenchCountryNames[country]; exists {
		country = frName
	}
	if country == "Algerie" {
		country = "Algérie"
	}

	var parts []string
	if locality != "" {
		parts = append(parts, locality)
	}
	if city != "" && !strings.EqualFold(city, locality) {
		parts = append(parts, city)
	}
	if country != "" {
		parts = append(parts, country)
	}

	var locationBase string
	if len(parts) > 0 {
		locationBase = strings.Join(parts, ", ")
	} else {
		locationBase = cleanISP(ispName)
	}

	var resolved string
	if isVPN {
		resolved = fmt.Sprintf("%s %s (VPN / %s)", flag, locationBase, cleanISP(ispName))
	} else {
		resolved = fmt.Sprintf("%s %s", flag, locationBase)
	}

	geoIPCache.Store(ipStr, resolved)
	return resolved
}

func cleanISP(isp string) string {
	isp = strings.TrimSpace(isp)
	if len(isp) > 20 {
		return isp[:18] + ".."
	}
	if isp == "" {
		return "Accès Distant"
	}
	return isp
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
		if clientLoc, ok := details["client_location"].(string); ok && strings.TrimSpace(clientLoc) != "" {
			return strings.TrimSpace(clientLoc)
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

	// Correspondance avec les règles CIDR ordonnées (sites internes et partenaires fixes)
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

	// Résolution dynamique GeoIP pour toute adresse IP publique (VPN, télétravail, auditeurs internationaux)
	return queryGeoIP(cleanIP)
}
