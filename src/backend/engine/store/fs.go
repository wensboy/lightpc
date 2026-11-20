// Package storeimpl provides the concrete store implementations.
package engineStoreImpl

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	interError "github.com/wensboy/lightpc/backend/internal/error"
	"github.com/wensboy/lightpc/backend/internal/path"
)

var (
	_ engineStore.FileStore = (*FS)(nil)
	_ domainEngine.Engine   = (*FS)(nil)
)

// FSOption is a critical construction parameter for FS.
type FSOption func(*FS)

// WithRoot sets the root directory of FS.
func WithRoot(root string) FSOption {
	return func(f *FS) { f.root = root }
}

// FS implements engineStore.FileStore and confines every operation below root.
type FS struct {
	root     string
	fileMode fs.FileMode
	dirMode  fs.FileMode

	mu     sync.RWMutex
	ectx   domainEngine.EngineContext
	state  domainEngine.EngineState
	closed bool
}

// NewFS builds an FS. Every field starts at its zero value; defaults come from configuration during Init.
func NewFS(opts ...FSOption) *FS {
	f := &FS{state: domainEngine.EnginePending}
	for _, opt := range opts {
		opt(f)
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
func (f *FS) Group() string { return domainEngine.GroupStore }

// Name implements domainEngine.Engine.
func (f *FS) Name() string { return domainEngine.NameFs }

// Deps implements domainEngine.Engine.
func (f *FS) Deps() []string { return []string{} }

// State implements domainEngine.Engine.
func (f *FS) State() domainEngine.EngineState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

// Init implements domainEngine.Engine: it keeps the engine context, publishes the store errors,
// applies configuration defaults for fields left at their zero value, ensures root and watches signals.
func (f *FS) Init(eCtx domainEngine.EngineContext) error {
	f.setState(domainEngine.EnginePending)
	interError.RegisterAll(eCtx.AppCtx, engineStore.Errors...)

	f.mu.Lock()
	f.ectx = eCtx
	if eCtx.AppCtx != nil && eCtx.AppCtx.ConfigContext != nil {
		domainContext.AssignIfZero(eCtx.AppCtx.ConfigContext, "default.runtime.store.root", &f.root, "", func(text string) (string, bool) {
			target, err := path.Parse(text)
			if err != nil || target.IsEmbed() {
				return "", false
			}
			return target.Path, true
		})
		domainContext.AssignIfZero(eCtx.AppCtx.ConfigContext, "default.runtime.store.fs.fileMode", &f.fileMode, 0, func(text string) (fs.FileMode, bool) {
			mode, err := strconv.ParseUint(text, 0, 32)
			if err != nil {
				return 0, false
			}
			return fs.FileMode(mode), true
		})
		domainContext.AssignIfZero(eCtx.AppCtx.ConfigContext, "default.runtime.store.fs.dirMode", &f.dirMode, 0, func(text string) (fs.FileMode, bool) {
			mode, err := strconv.ParseUint(text, 0, 32)
			if err != nil {
				return 0, false
			}
			return fs.FileMode(mode), true
		})
	}
	root, dirMode := f.root, f.dirMode
	f.mu.Unlock()

	if root != "" && dirMode != 0 {
		if err := os.MkdirAll(root, dirMode); err != nil {
			f.setState(domainEngine.EngineError)
			return interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(fmt.Errorf("create root: %w", err))
		}
	}

	if err := f.Validate(); err != nil {
		f.setState(domainEngine.EngineError)
		return err
	}

	watchStopSignal(eCtx.Ebus, domainEngine.NameFs, f.Close)
	f.setState(domainEngine.EngineReady)
	return nil
}

// Start implements domainEngine.Engine. FS serves requests synchronously, so it only moves the state.
func (f *FS) Start() error {
	f.setState(domainEngine.EngineRunning)
	return nil
}

// Restart implements domainEngine.Engine by reopening root with the retained context and validating it again.
func (f *FS) Restart() error {
	f.mu.Lock()
	f.closed = false
	eCtx := f.ectx
	f.mu.Unlock()
	return f.Init(eCtx)
}

// Stop implements domainEngine.Engine by closing the engineStore.
func (f *FS) Stop() error { return f.Close() }

// Validate checks that root exists and is a directory, then normalizes it to a real absolute path.
// It must be called before concurrent use.
func (f *FS) Validate() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.root == "" {
		return interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(errors.New("root not configured"))
	}
	real, err := filepath.EvalSymlinks(f.root)
	if err != nil {
		return interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(fmt.Errorf("eval root: %w", err))
	}
	info, err := os.Stat(real)
	if err != nil {
		return interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(fmt.Errorf("stat root: %w", err))
	}
	if !info.IsDir() {
		return interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(fmt.Errorf("root is not a directory: %s", real))
	}
	f.root = real
	return nil
}

// Get implements engineStore.FileStore.Get.
func (f *FS) Get(ctx context.Context, p string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if fsys, inner, ok := f.embedTarget(p); ok {
		data, err := fsys.ReadFile(inner)
		if err != nil {
			return nil, mapEmbedErr(err)
		}
		return data, nil
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

// Put implements engineStore.FileStore.Put.
func (f *FS) Put(ctx context.Context, p string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, _, ok := f.embedTarget(p); ok {
		return interError.Retrieve(engineStore.ErrKeyReadOnly)
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

// Delete implements engineStore.FileStore.Delete.
func (f *FS) Delete(ctx context.Context, p string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, _, ok := f.embedTarget(p); ok {
		return interError.Retrieve(engineStore.ErrKeyReadOnly)
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

// Stat implements engineStore.FileStore.Stat.
func (f *FS) Stat(ctx context.Context, p string) (engineStore.FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return engineStore.FileInfo{}, err
	}
	if fsys, inner, ok := f.embedTarget(p); ok {
		info, err := fs.Stat(fsys, inner)
		if err != nil {
			return engineStore.FileInfo{}, mapEmbedErr(err)
		}
		return engineStore.FileInfo{
			Path:    inner,
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		}, nil
	}
	root, _, _, err := f.rootAndModes()
	if err != nil {
		return engineStore.FileInfo{}, err
	}
	abs, err := resolve(root, p)
	if err != nil {
		return engineStore.FileInfo{}, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return engineStore.FileInfo{}, mapFSErr(err)
	}
	rel, _ := filepath.Rel(root, abs)
	return engineStore.FileInfo{
		Path:    rel,
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

// List implements engineStore.FileStore.List.
func (f *FS) List(ctx context.Context, prefix string) ([]engineStore.FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if fsys, inner, ok := f.embedTarget(prefix); ok {
		var out []engineStore.FileInfo
		walkErr := fs.WalkDir(fsys, inner, func(p string, d fs.DirEntry, walkErr error) error {
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
			out = append(out, engineStore.FileInfo{
				Path:    p,
				Size:    info.Size(),
				Mode:    info.Mode(),
				ModTime: info.ModTime(),
				IsDir:   info.IsDir(),
			})
			return nil
		})
		if walkErr != nil {
			return nil, mapEmbedErr(walkErr)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
		return out, nil
	}
	root, _, _, err := f.rootAndModes()
	if err != nil {
		return nil, err
	}
	base, err := resolve(root, prefix)
	if err != nil {
		return nil, err
	}

	var out []engineStore.FileInfo
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
		out = append(out, engineStore.FileInfo{
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

// MkdirAll implements engineStore.FileStore.MkdirAll.
func (f *FS) MkdirAll(ctx context.Context, p string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, _, ok := f.embedTarget(p); ok {
		return interError.Retrieve(engineStore.ErrKeyReadOnly)
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

// Close implements engineStore.FileStore.Close and is idempotent.
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
		return "", 0, 0, interError.Retrieve(engineStore.ErrKeyClosed)
	}
	if f.root == "" {
		return "", 0, 0, interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(errors.New("FS not validated, call Validate first"))
	}
	return f.root, f.fileMode, f.dirMode, nil
}

// embedTarget resolves p to an embedded file system and the path inside it. The longest matching
// EmbedContext key wins, so "/frontend/dist" is preferred over "/frontend". Reads use the returned
// pair, writes must be rejected because embedded content is immutable.
func (f *FS) embedTarget(p string) (*embed.FS, string, bool) {
	target, err := path.Parse(p)
	if err != nil || !target.IsEmbed() {
		return nil, "", false
	}

	f.mu.RLock()
	appCtx := f.ectx.AppCtx
	f.mu.RUnlock()
	if appCtx == nil {
		return nil, "", false
	}

	rest := target.Path
	var (
		found   *embed.FS
		inner   string
		matched int
	)
	for key, fsys := range appCtx.EmbedContext {
		if fsys == nil {
			continue
		}
		// The context key selects the file system while the whole remaining path stays intact: an
		// embed directive such as "all:data" keeps "data/" as the first segment inside the FS.
		prefix := strings.Trim(key, "/")
		if rest != prefix && !strings.HasPrefix(rest, prefix+"/") {
			continue
		}
		if len(prefix) > matched {
			found, inner, matched = fsys, rest, len(prefix)
		}
	}
	return found, inner, found != nil
}

// mapEmbedErr translates io/fs errors into store errors so callers can use errors.Is.
func mapEmbedErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return interError.Retrieve(engineStore.ErrKeyNotFound).Wrap(err)
	}
	return err
}

// resolve maps the virtual path p to an absolute path below root. A p that carries an explicit proto
// is anchored at root first, so configured local:// values and plain store paths share one entry point.
func resolve(root, p string) (string, error) {
	target, err := path.Parse(p)
	if err != nil {
		return "", interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(err)
	}
	if target.IsEmbed() {
		return "", interError.Retrieve(engineStore.ErrKeyInvalid)
	}

	rel := p
	if target.Explicit {
		var ok bool
		rel, ok = path.Relative(root, target.Path)
		if !ok {
			return "", interError.Retrieve(engineStore.ErrKeyEscape)
		}
	}

	clean := filepath.Clean("/" + rel)
	rel = strings.TrimPrefix(clean, string(filepath.Separator))
	abs := filepath.Join(root, rel)

	if !withinRoot(root, abs) {
		return "", interError.Retrieve(engineStore.ErrKeyEscape)
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		if !withinRoot(root, real) {
			return "", interError.Retrieve(engineStore.ErrKeyEscape)
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
		return interError.Retrieve(engineStore.ErrKeyNotFound).Wrap(err)
	case errors.Is(err, os.ErrExist):
		return interError.Retrieve(engineStore.ErrKeyExist).Wrap(err)
	default:
		return err
	}
}

// setState stores the engine state.
func (f *FS) setState(state domainEngine.EngineState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}
