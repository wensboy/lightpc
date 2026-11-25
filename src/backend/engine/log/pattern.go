package engineLogImpl

import (
	"log/slog"
	"path"
	"sync"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

var (
	_ domainEngine.Engine     = (*PatternLog)(nil)
	_ engineLog.PatternClient = (*PatternLog)(nil)
	_ engineLog.PatternServer = (*PatternLog)(nil)
)

// PatternOption is a critical construction parameter for PatternLog.
type PatternOption func(*PatternLog)

// WithPatternHandler appends handler to the set bound to pattern immediately; it is meant for
// programmatic wiring.
func WithPatternHandler(pattern string, handler slog.Handler) PatternOption {
	return func(pl *PatternLog) {
		if pattern == "" || handler == nil {
			return
		}
		pl.handlers[pattern] = append(pl.handlers[pattern], handler)
	}
}

// WithPatternEntry records one configured entry. Its handlers are opened during Init, after the store
// engine has created the root directory the log files live in.
func WithPatternEntry(entry engineLog.LogEntry) PatternOption {
	return func(pl *PatternLog) {
		if entry.Name == "" {
			return
		}
		pl.entries = append(pl.entries, entry)
	}
}

// PatternLog implements the pattern log engine: one record is delivered to every handler whose glob
// pattern matches the resolved name.
type PatternLog struct {
	handlers map[string][]slog.Handler
	entries  []engineLog.LogEntry

	mu     sync.RWMutex
	ectx   domainEngine.EngineContext
	state  domainEngine.EngineState
	closed bool
}

// NewPatternLog builds a PatternLog. Handlers may be preset with WithPatternHandler or registered later.
func NewPatternLog(opts ...PatternOption) *PatternLog {
	pl := &PatternLog{handlers: map[string][]slog.Handler{}, state: domainEngine.EnginePending}
	for _, opt := range opts {
		opt(pl)
	}
	return pl
}

// Group implements domainEngine.Engine.
func (pl *PatternLog) Group() string { return domainEngine.GroupLog }

// Name implements domainEngine.Engine.
func (pl *PatternLog) Name() string { return domainEngine.NamePattern }

// Deps implements domainEngine.Engine: log files live below the store root.
func (pl *PatternLog) Deps() []string { return []string{domainEngine.GroupStore} }

// State implements domainEngine.Engine.
func (pl *PatternLog) State() domainEngine.EngineState {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return pl.state
}

// Init implements domainEngine.Engine: it keeps the engine context, opens every configured handler
// and becomes ready.
func (pl *PatternLog) Init(eCtx domainEngine.EngineContext) error {
	if eCtx.AppCtx != nil {
		interError.RegisterAll(eCtx.AppCtx.ErrorContext, engineLog.Errors...)
	}

	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.ectx = eCtx
	pools := poolSourceOf(eCtx.Engines)
	for _, entry := range entriesOf(pl.entries) {
		pl.handlers[entry.name] = append(pl.handlers[entry.name], handlersFor(entry, pools)...)
	}
	pl.state = domainEngine.EngineReady
	return nil
}

// Start implements domainEngine.Engine. PatternLog serves lookups synchronously, so it only moves the state.
func (pl *PatternLog) Start() error {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.state = domainEngine.EngineRunning
	return nil
}

// Restart implements domainEngine.Engine by reopening with the retained context.
func (pl *PatternLog) Restart() error {
	pl.mu.Lock()
	pl.closed = false
	eCtx := pl.ectx
	pl.mu.Unlock()
	return pl.Init(eCtx)
}

// Stop implements domainEngine.Engine. Registered handlers are kept so a restart needs no rebinding.
func (pl *PatternLog) Stop() error {
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.closed = true
	pl.state = domainEngine.EngineStopped
	return nil
}

// Logger implements engineLog.PatternClient by broadcasting to every handler whose pattern matches
// name. Matching is attempted in both directions, so a concrete name can be looked up by a wildcard
// query and a wildcard registry key can still match a concrete name.
func (pl *PatternLog) Logger(name string) (*slog.Logger, bool) {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	matched := make([]slog.Handler, 0, len(pl.handlers))
	for pattern, handlers := range pl.handlers {
		forward, _ := path.Match(pattern, name)
		backward, _ := path.Match(name, pattern)
		if !forward && !backward {
			continue
		}
		matched = append(matched, handlers...)
	}
	if len(matched) == 0 {
		return nil, false
	}
	return slog.New(slog.NewMultiHandler(matched...)), true
}

// Register implements engineLog.PatternServer by appending handler to the set bound to pattern.
func (pl *PatternLog) Register(pattern string, handler slog.Handler) {
	if pattern == "" || handler == nil {
		return
	}
	pl.mu.Lock()
	defer pl.mu.Unlock()
	pl.handlers[pattern] = append(pl.handlers[pattern], handler)
}

// HookContext implements the log hook contract: it reports the configured entries so a hook can prepare
// their outputs.
func (pl *PatternLog) HookContext() engineLog.PatternHookContext {
	pl.mu.RLock()
	defer pl.mu.RUnlock()
	return engineLog.PatternHookContext{
		LogEntries:    pl.entries,
		EngineContext: pl.ectx,
	}
}
