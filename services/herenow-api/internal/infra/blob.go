// Package infra holds persistence adapters: the filesystem blob store and the
// file-backed metadata + hash-chained audit store. Both sit behind interfaces
// declared by their consumers (see internal/api). Postgres and S3-compatible
// adapters drop in behind the same interfaces for scale.
package infra

import (
	"os"
	"path/filepath"
)

// BlobFS is a filesystem-backed blob store. Bytes are returned only after the
// API layer's authorization allow — the store itself makes no access decisions.
type BlobFS struct{ dir string }

func NewBlobFS(dir string) (*BlobFS, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &BlobFS{dir: dir}, nil
}

func (b *BlobFS) path(slug string) string { return filepath.Join(b.dir, slug+".bundle") }

func (b *BlobFS) Put(slug string, content []byte) error {
	return os.WriteFile(b.path(slug), content, 0o600)
}

func (b *BlobFS) Get(slug string) ([]byte, error) {
	return os.ReadFile(b.path(slug))
}
