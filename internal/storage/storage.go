package storage

import (
	"context"
	"io"
)

// StorageProvider abstracts physical storage (Local Disk CAS or S3/MinIO)
type StorageProvider interface {
	// Save stores the content, calculates its SHA-256 hash on the fly, and returns metadata
	Save(ctx context.Context, r io.Reader) (sha256Hash string, sizeBytes int64, storagePath string, err error)

	// Open opens the file for streaming (resumable downloads, preview)
	Open(ctx context.Context, storagePath string) (io.ReadCloser, error)

	// Stat returns file information
	Stat(ctx context.Context, storagePath string) (sizeBytes int64, exists bool, err error)

	// Delete marks or removes a file (for temporary files)
	Delete(ctx context.Context, storagePath string) error
}
