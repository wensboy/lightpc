// Package settingimpl provides the layered settings engine.
package settingimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	"github.com/wensboy/lightpc/backend/domain/engine/setting"
	"github.com/wensboy/lightpc/backend/domain/engine/store"
	"github.com/wensboy/lightpc/backend/internal/cast"
	interError "github.com/wensboy/lightpc/backend/internal/error"
	"github.com/wensboy/lightpc/backend/internal/monitor"
	"github.com/wensboy/lightpc/backend/internal/path"
)

// defaultWatchInterval is used when no explicit interval is configured.
const defaultWatchInterval = 3 * time.Second

// runtimeLayer names the in-memory layer that holds unsaved UI edits. It carries no path, always sits on
// top and is emptied by Flush.
const runtimeLayer = "runtime"

// defaultFlushTarget is the layer Flush writes into; it must be the layer that owns persistence.
const defaultFlushTarget = "user"

// defaultFlushInterval is how often a dirty runtime layer sinks into the flush target. The gap is long on
// purpose: the manual path flushes immediately, so this only catches edits nobody saved explicitly.
const defaultFlushInterval = 30 * time.Second

var _ domainEngine.Engine = (*Setting)(nil)

// Option is a critical construction parameter for Setting.
type Option func(*Setting)

// WithLayer appends one layer. Layers are ordered by priority: a higher index wins.
func WithLayer(layer setting.Layer) Option {
	return func(s *Setting) {
		if layer.Name == "" {
			return
		}
		s.layers = append(s.layers, layer)
	}
}

// WithRuntimeLayer names the in-memory layer that receives unsaved edits.
func WithRuntimeLayer(name string) Option {
	return func(s *Setting) {
		if name != "" {
			s.runtime = name
		}
	}
}

// WithFlushTarget names the layer Flush writes into.
func WithFlushTarget(name string) Option {
	return func(s *Setting) {
		if name != "" {
			s.flushTarget = name
		}
	}
}

// WithFlushInterval overrides how often a dirty runtime layer sinks into the flush target.
func WithFlushInterval(interval time.Duration) Option {
	return func(s *Setting) {
		if interval > 0 {
			s.flushEvery = interval
		}
	}
}

// WithWatchInterval overrides how often layer files are re-read.
func WithWatchInterval(interval time.Duration) Option {
	return func(s *Setting) {
		if interval > 0 {
			s.interval = interval
		}
	}
}

// Setting stores every layer contribution of a key and exposes the aggregated view. Contributions
// travel as setting.Entry[any] because values differ per key; the engine never inspects them.
type Setting struct {
	layers   []setting.Layer
	values   map[string][]any
	watches  []*monitor.Monitor
	stamps   map[int]time.Time
	interval time.Duration

	runtime     string
	flushTarget string
	flushEvery  time.Duration
	dirty       bool

	mu     sync.RWMutex
	ectx   domainEngine.EngineContext
	state  domainEngine.EngineState
	closed bool
}

// NewSetting builds a Setting with the given layers, ordered by ascending priority.
func NewSetting(opts ...Option) *Setting {
	s := &Setting{
		values:      map[string][]any{},
		stamps:      map[int]time.Time{},
		interval:    defaultWatchInterval,
		runtime:     runtimeLayer,
		flushTarget: defaultFlushTarget,
		flushEvery:  defaultFlushInterval,
		state:       domainEngine.EnginePending,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Group implements domainEngine.Engine.
func (s *Setting) Group() string { return setting.GroupSetting }

// Name implements domainEngine.Engine.
func (s *Setting) Name() string { return setting.NameSetting }

// Deps implements domainEngine.Engine: layers live in the store. Diagnostics use slog directly, so the
// engine must not force a log sink that the configuration may leave disabled.
func (s *Setting) Deps() []string { return []string{store.StoreFs} }

// State implements domainEngine.Engine.
func (s *Setting) State() domainEngine.EngineState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Init implements domainEngine.Engine: it publishes the setting errors, loads every file backed layer
// and starts watching it for changes.
func (s *Setting) Init(eCtx domainEngine.EngineContext) error {
	interError.RegisterAll(eCtx.AppCtx, setting.Errors...)

	s.mu.Lock()
	s.ectx = eCtx
	// The runtime layer is always the top one and is never configured: it only holds unsaved edits.
	if _, found := s.layerIndex(s.runtime); !found {
		s.layers = append(s.layers, setting.Layer{Name: s.runtime, Mode: setting.ModeReadWrite})
	}
	s.state = domainEngine.EngineReady
	layers := slices.Clone(s.layers)
	s.mu.Unlock()

	for index, layer := range layers {
		if layer.Path == "" {
			continue
		}
		// Every declared resource must exist before the layer is read: embedded content is compiled in
		// and only verified, while local layers get their directory and an empty document created.
		if err := s.ensureLayer(layer); err != nil {
			slog.Default().Warn("setting layer unavailable", "layer", layer.Name, "err", err)
			continue
		}
		if err := s.loadLayer(index); err != nil {
			slog.Default().Warn("setting layer load failed", "layer", layer.Name, "err", err)
		}
		s.watch(index, layer)
	}
	return nil
}

// Start implements domainEngine.Engine. Setting serves lookups synchronously, so it only moves the state.
func (s *Setting) Start() error {
	s.watchFlush()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = domainEngine.EngineRunning
	return nil
}

// Restart implements domainEngine.Engine by reloading with the retained context.
func (s *Setting) Restart() error {
	s.stopWatches()

	s.mu.Lock()
	s.closed = false
	eCtx := s.ectx
	s.mu.Unlock()
	return s.Init(eCtx)
}

// Stop implements domainEngine.Engine by releasing every watcher.
func (s *Setting) Stop() error {
	if err := s.Flush(); err != nil {
		slog.Default().Warn("setting flush on stop failed", "err", err)
	}
	s.stopWatches()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.state = domainEngine.EngineStopped
	return nil
}

// Layers implements setting.Setting.
func (s *Setting) Layers() []setting.Layer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.layers)
}

// Layered implements setting.Setting.
func (s *Setting) Layered(key string) []any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.values[key])
}

// Keys implements setting.Setting: every key that has at least one contribution, ordered.
func (s *Setting) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.values))
	for key := range s.values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// Resolve implements setting.Setting by returning the highest priority contribution.
func (s *Setting) Resolve(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries := s.values[key]
	if len(entries) == 0 {
		return nil, false
	}
	return entries[len(entries)-1], true
}

// Set implements setting.Setting by writing value into the named writable layer.
func (s *Setting) Set(layer, key string, value any) error {
	s.mu.Lock()
	index, found := s.layerIndex(layer)
	if !found {
		s.mu.Unlock()
		return interError.Retrieve(setting.ErrKeyLayerUnknown)
	}
	if !canWrite(s.layers[index].Mode) {
		s.mu.Unlock()
		return interError.Retrieve(setting.ErrKeyLayerReadOnly)
	}
	s.putContribution(index, key, value)
	if s.layers[index].Name == s.runtime {
		s.dirty = true
	}
	s.mu.Unlock()

	s.notify(key, value)
	return nil
}

// Load implements setting.Setting by replacing the whole content of the named layer.
func (s *Setting) Load(layer string, values map[string]any) error {
	s.mu.Lock()
	index, found := s.layerIndex(layer)
	if !found {
		s.mu.Unlock()
		return interError.Retrieve(setting.ErrKeyLayerUnknown)
	}
	if !canRead(s.layers[index].Mode) {
		s.mu.Unlock()
		return interError.Retrieve(setting.ErrKeyLayerReadOnly)
	}
	for key := range s.values {
		s.dropContribution(index, key)
	}
	for key, value := range values {
		s.putContribution(index, key, value)
	}
	s.mu.Unlock()

	for key, value := range values {
		s.notify(key, value)
	}
	return nil
}

// Flush implements setting.Setting: it sinks every runtime contribution into the flush target and empties
// the runtime layer. Only runtime keys move, so a concurrent reload of the user layer merges key by key
// while the higher layer still wins on collisions.
func (s *Setting) Flush() error {
	s.mu.Lock()
	if !s.dirty {
		s.mu.Unlock()
		return nil
	}
	runtimeIndex, found := s.layerIndex(s.runtime)
	if !found {
		s.mu.Unlock()
		return interError.Retrieve(setting.ErrKeyLayerUnknown)
	}
	targetIndex, found := s.layerIndex(s.flushTarget)
	if !found {
		s.mu.Unlock()
		return interError.Retrieve(setting.ErrKeyLayerUnknown)
	}

	pending := s.contributionsOf(runtimeIndex)
	if len(pending) == 0 {
		s.dirty = false
		s.mu.Unlock()
		return nil
	}
	for key, value := range pending {
		s.putContribution(targetIndex, key, value)
		s.dropContribution(runtimeIndex, key)
	}
	document := s.contributionsOf(targetIndex)
	layer := s.layers[targetIndex]
	s.dirty = false
	s.mu.Unlock()

	if layer.Path == "" {
		return nil
	}
	return s.writeLayer(layer, document)
}

// layerIndex returns the position of the named layer; it must be called with the lock held.
func (s *Setting) layerIndex(name string) (int, bool) {
	for index := range s.layers {
		if s.layers[index].Name == name {
			return index, true
		}
	}
	return 0, false
}

// putContribution stores one setting.Entry[any] for key, keeping the slice ordered by layer position
// so the last element stays the winner. It must be called with the lock held.
func (s *Setting) putContribution(index int, key string, value any) {
	layer := s.layers[index]
	entry := setting.Entry[any]{Layer: layer.Name, Mode: layer.Mode, Value: value}

	entries := s.values[key]
	for position := range entries {
		existing, ok := cast.As[setting.Entry[any]](entries[position])
		if !ok || existing.Layer != layer.Name {
			continue
		}
		entries[position] = entry
		s.values[key] = entries
		return
	}

	position := len(entries)
	for position > 0 {
		previous, ok := cast.As[setting.Entry[any]](entries[position-1])
		if !ok {
			break
		}
		if previousIndex, found := s.layerIndex(previous.Layer); !found || previousIndex <= index {
			break
		}
		position--
	}
	entries = append(entries, nil)
	copy(entries[position+1:], entries[position:])
	entries[position] = entry
	s.values[key] = entries
}

// dropContribution removes the contribution of one layer from key. It must be called with the lock held.
func (s *Setting) dropContribution(index int, key string) {
	layer := s.layers[index].Name
	entries := s.values[key]
	kept := entries[:0]
	for _, raw := range entries {
		entry, ok := cast.As[setting.Entry[any]](raw)
		if ok && entry.Layer == layer {
			continue
		}
		kept = append(kept, raw)
	}
	if len(kept) == 0 {
		delete(s.values, key)
		return
	}
	s.values[key] = kept
}

// loadLayer reads one file backed layer through the store engine and replaces its contributions.
func (s *Setting) loadLayer(index int) error {
	fileStore, found := s.fileStore()
	if !found {
		return interError.Retrieve(setting.ErrKeyLayerUnknown)
	}

	layer := s.layers[index]
	ctx := context.Background()
	info, err := fileStore.Stat(ctx, layer.Path)
	if err != nil {
		return err
	}
	data, err := fileStore.Get(ctx, layer.Path)
	if err != nil {
		return err
	}

	stored := map[string]any{}
	if err := json.Unmarshal(data, &stored); err != nil {
		return fmt.Errorf("decode layer %s: %w", layer.Name, err)
	}
	values := flatten(stored)

	s.mu.Lock()
	for key := range s.values {
		s.dropContribution(index, key)
	}
	for key, value := range values {
		s.putContribution(index, key, value)
	}
	s.stamps[index] = info.ModTime
	s.mu.Unlock()

	// A reload replaces the whole layer, so every key it carries is republished: the UI hears about a
	// file edit without waiting for the next request.
	for key, value := range values {
		s.notify(key, value)
	}
	return nil
}

// watch re-reads the layer whenever its modification time changes.
func (s *Setting) watch(index int, layer setting.Layer) {
	watcher := monitor.New(s.interval, func() (any, error) {
		fileStore, found := s.fileStore()
		if !found {
			return nil, nil
		}
		info, err := fileStore.Stat(context.Background(), layer.Path)
		if err != nil {
			return nil, err
		}
		return info.ModTime, nil
	}, func(value any) {
		modTime, ok := cast.As[time.Time](value)
		if !ok {
			return
		}
		s.mu.RLock()
		previous := s.stamps[index]
		s.mu.RUnlock()
		if !modTime.After(previous) {
			return
		}
		if err := s.loadLayer(index); err != nil {
			slog.Default().Warn("setting layer reload failed", "layer", layer.Name, "err", err)
		}
	}, func(err error) {
		slog.Default().Warn("setting layer watch failed", "layer", layer.Name, "err", err)
	})
	watcher.Start()

	s.mu.Lock()
	s.watches = append(s.watches, watcher)
	s.mu.Unlock()
}

// stopWatches releases every watcher; it is safe to call when nothing is running.
func (s *Setting) stopWatches() {
	s.mu.Lock()
	watches := s.watches
	s.watches = nil
	s.mu.Unlock()

	for _, watcher := range watches {
		watcher.Stop()
	}
}

// fileStore resolves the file store dependency through the engine context.
func (s *Setting) fileStore() (store.FileStore, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	engine, found := s.ectx.Engines[store.StoreFs]
	if !found {
		return nil, false
	}
	return cast.As[store.FileStore](engine)
}

// notify publishes the change so services can react to it.
func (s *Setting) notify(key string, value any) {
	s.mu.RLock()
	ebus := s.ectx.Ebus
	s.mu.RUnlock()
	if ebus == nil {
		return
	}
	ebus.Emit(domainBus.EventSettingChanged, map[string]any{"key": key, "value": value})
}

// canRead reports whether the mode permits reading.
func canRead(mode setting.Mode) bool {
	return mode == setting.ModeReadOnly || mode == setting.ModeReadWrite
}

// canWrite reports whether the mode permits writing.
func canWrite(mode setting.Mode) bool {
	return mode == setting.ModeWriteOnly || mode == setting.ModeReadWrite
}

// ensureLayer makes sure the resource behind a layer exists and can be read.
func (s *Setting) ensureLayer(layer setting.Layer) error {
	target, err := path.Parse(layer.Path)
	if err != nil {
		return err
	}

	fileStore, found := s.fileStore()
	if !found {
		return interError.Retrieve(setting.ErrKeyLayerUnknown)
	}
	ctx := context.Background()

	if target.IsEmbed() {
		// Embedded content is immutable and shipped with the binary, so only its presence is checked.
		_, err := fileStore.Stat(ctx, layer.Path)
		return err
	}

	if _, err := fileStore.Stat(ctx, layer.Path); err == nil {
		return nil
	}
	if parent := path.Dir(layer.Path); parent != "" {
		if err := fileStore.MkdirAll(ctx, parent); err != nil {
			return err
		}
	}
	if !canWrite(layer.Mode) {
		return nil
	}
	return fileStore.Put(ctx, layer.Path, []byte("{}\n"))
}

// contributionsOf collects the values the layer at index currently contributes.
func (s *Setting) contributionsOf(index int) map[string]any {
	values := map[string]any{}
	if index < 0 || index >= len(s.layers) {
		return values
	}
	name := s.layers[index].Name
	for key, entries := range s.values {
		for _, raw := range entries {
			entry, ok := cast.As[setting.Entry[any]](raw)
			if !ok || entry.Layer != name {
				continue
			}
			values[key] = entry.Value
		}
	}
	return values
}

// writeLayer serialises values into the file behind layer, creating the parent directory when needed.
func (s *Setting) writeLayer(layer setting.Layer, values map[string]any) error {
	fileStore, found := s.fileStore()
	if !found {
		return interError.Retrieve(setting.ErrKeyLayerUnknown)
	}
	data, err := json.MarshalIndent(nest(values), "", "  ")
	if err != nil {
		return interError.Retrieve(setting.ErrKeyLayerUnknown).Wrap(err)
	}
	if parent := path.Dir(layer.Path); parent != "" {
		if err := fileStore.MkdirAll(context.Background(), parent); err != nil {
			return err
		}
	}
	return fileStore.Put(context.Background(), layer.Path, append(data, '\n'))
}

// watchFlush sinks a dirty runtime layer every flushEvery. The manual path already flushes on demand, so
// this only catches edits nobody saved explicitly and is deliberately infrequent.
func (s *Setting) watchFlush() {
	watcher := monitor.New(s.flushEvery, func() (any, error) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.dirty, nil
	}, func(value any) {
		dirty, ok := cast.As[bool](value)
		if !ok || !dirty {
			return
		}
		if err := s.Flush(); err != nil {
			slog.Default().Warn("setting flush failed", "err", err)
		}
	}, func(err error) {
		slog.Default().Warn("setting flush watch failed", "err", err)
	})
	watcher.Start()

	s.mu.Lock()
	s.watches = append(s.watches, watcher)
	s.mu.Unlock()
}

// nest expands dotted keys into nested maps so the stored document mirrors the schema shape while the
// engine keeps addressing every leaf directly.
func nest(flat map[string]any) map[string]any {
	document := map[string]any{}
	for key, value := range flat {
		segments := strings.Split(key, ".")
		cursor := document
		for _, segment := range segments[:len(segments)-1] {
			child, ok := cursor[segment].(map[string]any)
			if !ok {
				child = map[string]any{}
				cursor[segment] = child
			}
			cursor = child
		}
		cursor[segments[len(segments)-1]] = value
	}
	return document
}

// flatten is the inverse of nest: every leaf becomes a dotted key. Arrays stay whole because the engine
// treats them as single values.
func flatten(document map[string]any) map[string]any {
	flat := map[string]any{}
	var walk func(prefix string, node map[string]any)
	walk = func(prefix string, node map[string]any) {
		for key, value := range node {
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}
			if child, ok := value.(map[string]any); ok {
				walk(path, child)
				continue
			}
			flat[path] = value
		}
	}
	walk("", document)
	return flat
}
