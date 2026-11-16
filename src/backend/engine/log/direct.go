package logimpl

import (
	"log/slog"
	"sync"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	domainLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	"github.com/wensboy/lightpc/backend/domain/engine/store"
)

var (
	_ domainEngine.Engine    = (*DirectLog)(nil)
	_ domainLog.DirectClient = (*DirectLog)(nil)
	_ domainLog.DirectServer = (*DirectLog)(nil)
)

// DirectOption is a critical construction parameter for DirectLog.
type DirectOption func(*DirectLog)

// WithDirectHandler binds handler to name before the engine starts.
func WithDirectHandler(name string, handler slog.Handler) DirectOption {
	return func(dl *DirectLog) {
		if name == "" || handler == nil {
			return
		}
		dl.handlers[name] = handler
	}
}

// DirectLog implements the direct log engine: handlers are registered under an exact name and a
// record written through one logger is consumed by exactly one handler.
type DirectLog struct {
	handlers map[string]slog.Handler

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
func (dl *DirectLog) Group() string { return domainLog.GroupLog }

// Name implements domainEngine.Engine.
func (dl *DirectLog) Name() string { return domainLog.NameDirect }

// Deps implements domainEngine.Engine.
func (dl *DirectLog) Deps() []string { return []string{store.StoreFs} }

// State implements domainEngine.Engine.
func (dl *DirectLog) State() domainEngine.EngineState {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	return dl.state
}

// Init implements domainEngine.Engine: it keeps the engine context and becomes ready.
func (dl *DirectLog) Init(eCtx domainEngine.EngineContext) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.ectx = eCtx
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

// Logger implements domainLog.DirectClient by returning the logger bound to name.
func (dl *DirectLog) Logger(name string) (*slog.Logger, bool) {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	handler, found := dl.handlers[name]
	if !found {
		return nil, false
	}
	return slog.New(handler), true
}

// Register implements domainLog.DirectServer by binding handler to name.
func (dl *DirectLog) Register(name string, handler slog.Handler) {
	if name == "" || handler == nil {
		return
	}
	dl.mu.Lock()
	defer dl.mu.Unlock()
	dl.handlers[name] = handler
}
