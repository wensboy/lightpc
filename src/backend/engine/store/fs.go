// Package storeimpl provides the concrete store implementations.
package storeimpl

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	"github.com/wensboy/lightpc/backend/domain/engine/store"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

var (
	_ store.FileStore     = (*FS)(nil)
	_ domainEngine.Engine = (*FS)(nil)
)

// FSOption is a critical construction parameter for FS.
type FSOption func(*FS)

// WithRoot sets the root directory of FS.
func WithRoot(root string) FSOption {
	return func(f *FS) { f.root = root }
}

// FS implements store.FileStore and confines every operation below root.
type FS struct {
	root     string
	fileMode fs.FileMode
	dirMode  fs.FileMode
	mu       sync.RWMutex
	state    domainEngine.EngineState
	closed   bool
}

// NewFS builds an FS. Callers must call Validate before concurrent use.
func NewFS(opts ...FSOption) *FS {
	f := &FS{fileMode: 0o644, dirMode: 0o755, state: domainEngine.EnginePending}
	for _, opt := range opts {
		opt(f)
	}
	if f.root != "" {
		if abs, err := filepath.Abs(f.root); err == nil {
			f.root = abs
		}
	}
	return f
}

// WithFileMode sets the file mode used by Put and returns *FS for chaining.
// It must be called before concurrent use.
func (f *FS) WithFileMode(m fs.FileMode) *FS {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fileMode = m
	return f
}

// WithDirMode sets the directory mode used by MkdirAll and returns *FS for chaining.
// It must be called before concurrent use.
func (f *FS) WithDirMode(m fs.FileMode) *FS {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dirMode = m
	return f
}

// Group implements domainEngine.Engine.
func (f *FS) Group() string { return store.GroupStore }

// Name implements domainEngine.Engine.
func (f *FS) Name() string { return store.NameFs }

// Deps implements domainEngine.Engine.
func (f *FS) Deps() []string { return []string{} }

// State implements domainEngine.Engine.
func (f *FS) State() domainEngine.EngineState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

// Init implements domainEngine.Engine: it publishes the store errors, then validates root.
func (f *FS) Init(eCtx domainEngine.EngineContext) error {
	f.setState(domainEngine.EnginePending)
	interError.RegisterAll(eCtx.AppCtx, store.Errors...)
	if err := f.Validate(); err != nil {
		f.setState(domainEngine.EngineError)
		return err
	}
	f.setState(domainEngine.EngineReady)
	return nil
}

// Start implements domainEngine.Engine.
func (f *FS) Start() error {
	f.setState(domainEngine.EngineRunning)
	return nil
}

// Restart implements domainEngine.Engine by reopening root and validating it again.
func (f *FS) Restart() error {
	f.mu.Lock()
	f.closed = false
	f.mu.Unlock()
	return f.Init(domainEngine.EngineContext{})
}

// Stop implements domainEngine.Engine by closing the store.
func (f *FS) Stop() error { return f.Close() }

// setState stores the engine state.
func (f *FS) setState(state domainEngine.EngineState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}

// Validate checks that root exists and is a directory, then normalizes it to a real absolute path.
// It must be called before concurrent use.
func (f *FS) Validate() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.root == "" {
		return interError.Retrieve(store.ErrKeyInvalid).Wrap(errors.New("root not configured"))
	}
	real, err := filepath.EvalSymlinks(f.root)
	if err != nil {
		return interError.Retrieve(store.ErrKeyInvalid).Wrap(fmt.Errorf("eval root: %w", err))
	}
	info, err := os.Stat(real)
	if err != nil {
		return interError.Retrieve(store.ErrKeyInvalid).Wrap(fmt.Errorf("stat root: %w", err))
	}
	if !info.IsDir() {
		return interError.Retrieve(store.ErrKeyInvalid).Wrap(fmt.Errorf("root is not a directory: %s", real))
	}
	f.root = real
	return nil
}

// Get implements store.FileStore.Get.
func (f *FS) Get(ctx context.Context, p string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, _, _, err := f.rootAndModes()
	if err != nil {
		return nil, err
	}
	abs, err := resolve(root, p)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, mapFSErr(err)
	}
	return data, nil
}

// Put implements store.FileStore.Put.
func (f *FS) Put(ctx context.Context, p string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	root, fileMode, _, err := f.rootAndModes()
	if err != nil {
		return err
	}
	abs, err := resolve(root, p)
	if err != nil {
		return err
	}
	return mapFSErr(os.WriteFile(abs, data, fileMode))
}

// Delete implements store.FileStore.Delete.
func (f *FS) Delete(ctx context.Context, p string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	root, _, _, err := f.rootAndModes()
	if err != nil {
		return err
	}
	abs, err := resolve(root, p)
	if err != nil {
		return err
	}
	return mapFSErr(os.Remove(abs))
}

// Stat implements store.FileStore.Stat.
func (f *FS) Stat(ctx context.Context, p string) (store.FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return store.FileInfo{}, err
	}
	root, _, _, err := f.rootAndModes()
	if err != nil {
		return store.FileInfo{}, err
	}
	abs, err := resolve(root, p)
	if err != nil {
		return store.FileInfo{}, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return store.FileInfo{}, mapFSErr(err)
	}
	rel, _ := filepath.Rel(root, abs)
	return store.FileInfo{
		Path:    rel,
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

// List implements store.FileStore.List.
func (f *FS) List(ctx context.Context, prefix string) ([]store.FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, _, _, err := f.rootAndModes()
	if err != nil {
		return nil, err
	}
	base, err := resolve(root, prefix)
	if err != nil {
		return nil, err
	}

	var out []store.FileInfo
	walkErr := filepath.WalkDir(base, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		out = append(out, store.FileInfo{
			Path:    rel,
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})
		return nil
	})
	if walkErr != nil {
		return nil, mapFSErr(walkErr)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// MkdirAll implements store.FileStore.MkdirAll.
func (f *FS) MkdirAll(ctx context.Context, p string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	root, _, dirMode, err := f.rootAndModes()
	if err != nil {
		return err
	}
	abs, err := resolve(root, p)
	if err != nil {
		return err
	}
	return mapFSErr(os.MkdirAll(abs, dirMode))
}

// Close implements store.FileStore.Close and is idempotent.
func (f *FS) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	f.state = domainEngine.EngineStopped
	return nil
}

// rootAndModes snapshots root and both modes, and reports an error when FS is invalid or closed.
func (f *FS) rootAndModes() (string, fs.FileMode, fs.FileMode, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.closed {
		return "", 0, 0, interError.Retrieve(store.ErrKeyClosed)
	}
	if f.root == "" {
		return "", 0, 0, interError.Retrieve(store.ErrKeyInvalid).Wrap(errors.New("FS not validated, call Validate first"))
	}
	return f.root, f.fileMode, f.dirMode, nil
}

// resolve maps the virtual path p to an absolute path below root.
func resolve(root, p string) (string, error) {
	clean := filepath.Clean("/" + p)
	rel := strings.TrimPrefix(clean, string(filepath.Separator))
	abs := filepath.Join(root, rel)

	if !withinRoot(root, abs) {
		return "", interError.Retrieve(store.ErrKeyEscape)
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		if !withinRoot(root, real) {
			return "", interError.Retrieve(store.ErrKeyEscape)
		}
		return real, nil
	}
	return abs, nil
}

// withinRoot reports whether target stays below root. Both paths must be absolute.
func withinRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// mapFSErr translates os level errors into store errors so callers can use errors.Is.
func mapFSErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, os.ErrNotExist):
		return interError.Retrieve(store.ErrKeyNotFound).Wrap(err)
	case errors.Is(err, os.ErrExist):
		return interError.Retrieve(store.ErrKeyExist).Wrap(err)
	default:
		return err
	}
}
