package audit

import (
	"testing"
)

func TestResolveLocation(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		ua        string
		details   map[string]any
		expected  string
	}{
		{
			name:     "Loopback IPv4",
			ip:       "127.0.0.1",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "💻 Console Locale Serveur (Sur Site)",
		},
		{
			name:     "Loopback with port",
			ip:       "127.0.0.1:51234",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "💻 Console Locale Serveur (Sur Site)",
		},
		{
			name:     "Loopback IPv6",
			ip:       "::1",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "💻 Console Locale Serveur (Sur Site)",
		},
		{
			name:     "Empty IP",
			ip:       "",
			ua:       "",
			details:  nil,
			expected: "💻 Console Locale Serveur (Sur Site)",
		},
		{
			name:     "Siège Administratif & RA (Alger)",
			ip:       "192.168.1.45",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🏢 Siège Administratif & RA (Alger)",
		},
		{
			name:     "Usine & Production (Oued Smar)",
			ip:       "192.168.10.12",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🏭 Usine & Production (Oued Smar)",
		},
		{
			name:     "Labo LCQ (Alger)",
			ip:       "10.20.1.5",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🔬 Labo Contrôle Qualité LCQ (Alger)",
		},
		{
			name:     "Magasin Baraki",
			ip:       "10.30.2.10",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "📦 Magasin Central & Logistique (Baraki)",
		},
		{
			name:     "CERIST / Réseau Ministériel (Alger)",
			ip:       "193.194.66.1",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🇩🇿 CERIST / Réseau Ministériel (Alger)",
		},
		{
			name:     "Algérie Télécom Alger",
			ip:       "41.200.15.8",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🇩🇿 Alger (Algérie Télécom)",
		},
		{
			name:     "Oran",
			ip:       "41.220.50.2",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🇩🇿 Oran (Algérie)",
		},
		{
			name:     "Fibre/4G Alger",
			ip:       "105.100.20.4",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🇩🇿 Alger, Algérie (Fibre/4G)",
		},
		{
			name:     "Partenaire Paris B2B",
			ip:       "194.50.12.3",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🇫🇷 Paris, France (Partenaire B2B)",
		},
		{
			name:     "Titulaire API Suisse",
			ip:       "195.12.4.99",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🇨🇭 Bâle, Suisse (Titulaire API)",
		},
		{
			name:     "Auditeur Qualité Allemagne",
			ip:       "165.22.10.1",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🇩🇪 Francfort, Allemagne (Auditeur Qualité)",
		},
		{
			name:     "Fournisseur API Inde",
			ip:       "115.110.50.2",
			ua:       "Mozilla/5.0",
			details:  nil,
			expected: "🇮🇳 Hyderabad, Inde (Fournisseur API)",
		},
		{
			name:     "Explicit metadata location override",
			ip:       "127.0.0.1",
			ua:       "Mozilla/5.0",
			details:  map[string]any{"location": "Site Spécifique Constantine"},
			expected: "Site Spécifique Constantine",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := ResolveLocation(tc.ip, tc.ua, tc.details)
			if res != tc.expected {
				t.Errorf("ResolveLocation(%q) = %q, want %q", tc.ip, res, tc.expected)
			}
		})
	}
}
