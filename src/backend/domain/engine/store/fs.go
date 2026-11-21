// Package store defines the storage engine contracts: interfaces and command or response messages.
package engineStore

import (
	"context"
	"io/fs"
	"time"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
)

// FileInfo describes one entry below root.
type FileInfo struct {
	Path    string
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
	IsDir   bool
}

// FileStore is a synchronous file store with chroot semantics.
type FileStore interface {
	// Get reads the whole content of path and reports ErrKeyNotFound when it is missing.
	Get(ctx context.Context, path string) ([]byte, error)
	// Put writes path and replaces existing content; it reports ErrKeyNotFound when the parent directory is missing.
	Put(ctx context.Context, path string, data []byte) error
	// Delete removes path and reports ErrKeyNotFound when it is missing.
	Delete(ctx context.Context, path string) error
	// Stat returns the metadata of path and reports ErrKeyNotFound when it is missing.
	Stat(ctx context.Context, path string) (FileInfo, error)
	// List returns every entry whose Path carries the given prefix, ordered by Path; an empty prefix lists everything.
	List(ctx context.Context, prefix string) ([]FileInfo, error)
	// MkdirAll ensures path exists as a directory, similar to mkdir -p.
	MkdirAll(ctx context.Context, path string) error
	// Close releases resources and is idempotent.
	Close() error
}

// FsHookContext is what a store hook may inspect between the engine phases. The engine owns the fields: a
// hook reads them to decide whether to prepare the root or to run a check, and changes state through the
// engine's own methods so it stays in one place.
type FsHookContext struct {
	// Root is the absolute directory the store is chrooted to.
	Root string
	// FileMode is applied to files the store creates.
	FileMode fs.FileMode
	// DirMode is applied to directories the store creates.
	DirMode fs.FileMode
	// EngineContext is the context the engine was initialised with.
	EngineContext domainEngine.EngineContext
}
