package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalCASStorage implements Content-Addressable Storage on local filesystem with SHA-256 deduplication
type LocalCASStorage struct {
	baseDir string
}

// NewLocalCASStorage creates a new storage instance ensuring root directories exist
func NewLocalCASStorage(baseDir string) (*LocalCASStorage, error) {
	if err := os.MkdirAll(baseDir, 0750); err != nil {
		return nil, fmt.Errorf("impossible de créer le répertoire racine de stockage: %w", err)
	}
	return &LocalCASStorage{baseDir: baseDir}, nil
}

// Save streams content to a temporary file, calculates SHA-256, and moves to content-addressed path
func (s *LocalCASStorage) Save(ctx context.Context, r io.Reader) (string, int64, string, error) {
	// Création d'un fichier temporaire
	tempFile, err := os.CreateTemp(s.baseDir, "upload_*.tmp")
	if err != nil {
		return "", 0, "", fmt.Errorf("création du fichier temporaire échouée: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath) // Nettoyé si non renommé
	}()

	hasher := sha256.New()
	tee := io.TeeReader(r, hasher)

	size, err := io.Copy(tempFile, tee)
	if err != nil {
		return "", 0, "", fmt.Errorf("erreur lors de l'écriture du flux: %w", err)
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	// Arborescence découpée par préfixe : baseDir/ab/cd/abcdef...
	dir := filepath.Join(s.baseDir, hash[:2], hash[2:4])
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", 0, "", fmt.Errorf("création des répertoires de stockage échouée: %w", err)
	}

	finalPath := filepath.Join(dir, hash)

	// Déduplication : si le fichier existe déjà avec ce hash exact, on ne fait que supprimer le tmp
	if _, err := os.Stat(finalPath); err == nil {
		// Le fichier identique existe déjà, économie de disque (Déduplication BPF)
		return hash, size, finalPath, nil
	}

	// Déplacement atomique
	_ = tempFile.Close()
	if err := os.Rename(tempPath, finalPath); err != nil {
		return "", 0, "", fmt.Errorf("déplacement du fichier vers le stockage final échoué: %w", err)
	}

	// Rendre le fichier en lecture seule pour empêcher toute altération
	_ = os.Chmod(finalPath, 0440)

	return hash, size, finalPath, nil
}

// Open opens the file from storage
func (s *LocalCASStorage) Open(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	file, err := os.Open(storagePath)
	if err != nil {
		return nil, fmt.Errorf("ouverture du document échouée: %w", err)
	}
	return file, nil
}

// Stat returns metadata about the file
func (s *LocalCASStorage) Stat(ctx context.Context, storagePath string) (int64, bool, error) {
	fi, err := os.Stat(storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return fi.Size(), true, nil
}

// Delete removes a file (utilisé uniquement pour nettoyages spécifiques)
func (s *LocalCASStorage) Delete(ctx context.Context, storagePath string) error {
	return os.Remove(storagePath)
}
