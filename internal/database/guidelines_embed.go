package database

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed guidelines_pdfs/*.pdf
var EmbeddedGuidelinePDFs embed.FS

// SyncGuidelinePDFsToDB stocke le contenu binaire complet (BYTEA) des 19 PDF officiels ANPP
// dans PostgreSQL (regulatory_watch_items) pour toute entrée dont document_pdf est NULL.
func SyncGuidelinePDFsToDB(ctx context.Context, pool *pgxpool.Pool) {
	if pool == nil {
		return
	}
	entries, err := fs.ReadDir(EmbeddedGuidelinePDFs, "guidelines_pdfs")
	if err != nil {
		log.Printf("[Guidelines] Lecture guidelines_pdfs impossible: %v", err)
		return
	}

	updatedRows := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".pdf" {
			continue
		}
		data, readErr := EmbeddedGuidelinePDFs.ReadFile("guidelines_pdfs/" + entry.Name())
		if readErr != nil || len(data) == 0 {
			continue
		}
		sum := sha256.Sum256(data)
		shaHex := hex.EncodeToString(sum[:])

		tag, execErr := pool.Exec(ctx, `
			INSERT INTO regulatory_watch_documents (document_filename, document_pdf, document_sha256, document_size_bytes, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
			ON CONFLICT (document_filename) DO UPDATE SET
				document_pdf = EXCLUDED.document_pdf,
				document_sha256 = EXCLUDED.document_sha256,
				document_size_bytes = EXCLUDED.document_size_bytes,
				updated_at = NOW()
			WHERE regulatory_watch_documents.document_sha256 IS DISTINCT FROM EXCLUDED.document_sha256
		`, entry.Name(), data, shaHex, int64(len(data)))
		if execErr == nil {
			updatedRows += int(tag.RowsAffected())
		} else {
			log.Printf("[Guidelines] Erreur sync PDF %s: %v", entry.Name(), execErr)
		}
	}
	if updatedRows > 0 {
		log.Printf("[Guidelines] %d fichier(s) PDF ANPP synchronisés dans PostgreSQL (regulatory_watch_items).", updatedRows)
	}
}

// GetEmbeddedGuidelinePDF lit un PDF officiel ANPP embarqué par son nom de fichier.
func GetEmbeddedGuidelinePDF(filename string) ([]byte, error) {
	data, err := EmbeddedGuidelinePDFs.ReadFile("guidelines_pdfs/" + filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("guideline PDF %q introuvable: %w", filename, err)
	}
	return data, nil
}
