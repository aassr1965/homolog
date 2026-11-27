package reporting

import (
	"embed"
	"fmt"
	"path"
	"sort"
	"sync"
)

//go:embed templates/anpp/*
var anppTemplatesFS embed.FS

var (
	templateOverrideMu       sync.RWMutex
	templateOverrideResolver func(orgID, filename string) ([]byte, error)
)

// RegisterTemplateOverrideResolver enregistre une fonction de résolution dynamique (base/CAS) prioritaire sur go:embed.
func RegisterTemplateOverrideResolver(fn func(orgID, filename string) ([]byte, error)) {
	templateOverrideMu.Lock()
	defer templateOverrideMu.Unlock()
	templateOverrideResolver = fn
}

// GetANPPTemplateForOrg retourne la surcharge dynamique si disponible (établissement ou globale), sinon le gabarit embarqué.
func GetANPPTemplateForOrg(orgID, filename string) ([]byte, error) {
	clean := path.Base(filename)
	templateOverrideMu.RLock()
	resolver := templateOverrideResolver
	templateOverrideMu.RUnlock()

	if resolver != nil {
		if custom, err := resolver(orgID, clean); err == nil && len(custom) > 0 {
			return custom, nil
		}
	}
	return GetANPPTemplate(clean)
}

// GetANPPTemplate retourne les octets bruts d'un gabarit officiel ANPP embarqué (.docx ou .xlsx).
func GetANPPTemplate(filename string) ([]byte, error) {
	clean := path.Base(filename)
	data, err := anppTemplatesFS.ReadFile("templates/anpp/" + clean)
	if err != nil {
		return nil, fmt.Errorf("gabarit ANPP introuvable (%s): %w", clean, err)
	}
	return data, nil
}

// ListANPPTemplates liste tous les gabarits officiels ANPP embarqués dans le binaire.
func ListANPPTemplates() []string {
	entries, err := anppTemplatesFS.ReadDir("templates/anpp")
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}
