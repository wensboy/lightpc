package engineLogImpl

import (
	"log/slog"
	"sync"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

var (
	_ domainEngine.Engine    = (*DirectLog)(nil)
	_ engineLog.DirectClient = (*DirectLog)(nil)
	_ engineLog.DirectServer = (*DirectLog)(nil)
)

// DirectOption is a critical construction parameter for DirectLog.
type DirectOption func(*DirectLog)

// WithDirectHandler binds handler to name immediately; it is meant for programmatic wiring.
func WithDirectHandler(name string, handler slog.Handler) DirectOption {
	return func(dl *DirectLog) {
		if name == "" || handler == nil {
			return
		}
		dl.handlers[name] = handler
	}
}

// WithDirectEntry records one configured entry. Its handlers are opened during Init, after the store
// engine has created the root directory the log files live in.
func WithDirectEntry(entry engineLog.LogEntry) DirectOption {
	return func(dl *DirectLog) {
		if entry.Name == "" {
			return
		}
		dl.entries = append(dl.entries, entry)
	}
}

// DirectLog implements the direct log engine: handlers are registered under an exact name and a
// record written through one logger is consumed by exactly one handler.
type DirectLog struct {
	handlers map[string]slog.Handler
	entries  []engineLog.LogEntry

	mu     sync.RWMutex
	ectx   domainEngine.EngineContext
	state  domainEngine.EngineState
	closed bool
}

// NewDirectLog builds a DirectLog. Handlers may be preset with WithDirectHandler or registered later.
func NewDirectLog(opts ...DirectOption) *DirectLog {
	dl := &DirectLog{handlers: map[string]slog.Handler{}, state: domainEngine.EnginePending}
	for _, opt := range opts {
		opt(dl)
	}
	return dl
}

// Group implements domainEngine.Engine.
func (dl *DirectLog) Group() string { return domainEngine.GroupLog }

// Name implements domainEngine.Engine.
func (dl *DirectLog) Name() string { return domainEngine.NameDirect }

// Deps implements domainEngine.Engine: log files live below the store root.
func (dl *DirectLog) Deps() []string { return []string{domainEngine.StoreFs} }

// State implements domainEngine.Engine.
func (dl *DirectLog) State() domainEngine.EngineState {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	return dl.state
}

// Init implements domainEngine.Engine: it keeps the engine context, opens every configured handler
// and becomes ready.
func (dl *DirectLog) Init(eCtx domainEngine.EngineContext) error {
	interError.RegisterAll(eCtx.AppCtx, engineLog.Errors...)

	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.ectx = eCtx
	for _, entry := range entriesOf(dl.entries) {
		for _, handler := range handlersFor(entry) {
			dl.handlers[entry.name] = handler
		}
	}
	dl.state = domainEngine.EngineReady
	return nil
}

// Start implements domainEngine.Engine. DirectLog serves lookups synchronously, so it only moves the state.
func (dl *DirectLog) Start() error {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.state = domainEngine.EngineRunning
	return nil
}

// Restart implements domainEngine.Engine by reopening with the retained context.
func (dl *DirectLog) Restart() error {
	dl.mu.Lock()
	dl.closed = false
	eCtx := dl.ectx
	dl.mu.Unlock()
	return dl.Init(eCtx)
}

// Stop implements domainEngine.Engine. Registered handlers are kept so a restart needs no rebinding.
func (dl *DirectLog) Stop() error {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.closed = true
	dl.state = domainEngine.EngineStopped
	return nil
}

// Logger implements engineLog.DirectClient by returning the logger bound to name.
func (dl *DirectLog) Logger(name string) (*slog.Logger, bool) {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	handler, found := dl.handlers[name]
	if !found {
		return nil, false
	}
	return slog.New(handler), true
}

// Register implements engineLog.DirectServer by binding handler to name.
func (dl *DirectLog) Register(name string, handler slog.Handler) {
	if name == "" || handler == nil {
		return
	}
	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.handlers[name] = handler
}

// HookContext implements the log hook contract: it reports the configured entries so a hook can prepare
// their outputs.
func (dl *DirectLog) HookContext() engineLog.DirectHookContext {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	return engineLog.DirectHookContext{
		LogEntries:    dl.entries,
		EngineContext: dl.ectx,
	}
}
