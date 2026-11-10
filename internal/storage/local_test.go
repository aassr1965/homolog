package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"testing"
)

func TestLocalCASStorage_SaveAndDeduplicate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cas_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cas, err := NewLocalCASStorage(tempDir)
	if err != nil {
		t.Fatalf("NewLocalCASStorage failed: %v", err)
	}

	content := []byte("ANPP_PARACETAMOL_500MG_CERTIFICAT_BPF_CONFORME_2026")
	expectedHashBytes := sha256.Sum256(content)
	expectedHash := hex.EncodeToString(expectedHashBytes[:])

	ctx := context.Background()

	// Premier enregistrement
	hash1, size1, path1, err := cas.Save(ctx, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("First Save failed: %v", err)
	}

	if hash1 != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, hash1)
	}
	if size1 != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), size1)
	}

	// Vérification de lecture
	rc, err := cas.Open(ctx, path1)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer rc.Close()

	readBytes, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if !bytes.Equal(readBytes, content) {
		t.Errorf("Read content does not match original")
	}

	// Deuxième enregistrement du même contenu (Test de DÉDUPLICATION)
	hash2, size2, path2, err := cas.Save(ctx, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("Second Save failed: %v", err)
	}

	if hash2 != hash1 || path2 != path1 || size2 != size1 {
		t.Errorf("Deduplication failed: got different hash or path for identical content")
	}
}
