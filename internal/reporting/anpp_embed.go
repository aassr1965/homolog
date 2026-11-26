package reporting

import (
	"embed"
	"fmt"
	"path"
	"sort"
)

//go:embed templates/anpp/*
var anppTemplatesFS embed.FS

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
