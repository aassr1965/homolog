package web_test

import (
	"testing"

	"homolog/web"
)

func TestEmbeddedFiles(t *testing.T) {
	requiredFiles := []string{
		"index.html",
		"css/style.css",
		"js/app.js",
		"js/auth.js",
		"js/referential.js",
		"js/variations.js",
		"js/ged.js",
		"js/dossiers.js",
	}

	for _, file := range requiredFiles {
		content, err := web.Files.ReadFile(file)
		if err != nil {
			t.Errorf("Impossible de lire le fichier embarqué '%s': %v", file, err)
			continue
		}
		if len(content) == 0 {
			t.Errorf("Le fichier embarqué '%s' est vide", file)
		}
	}
}
