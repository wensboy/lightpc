package logimpl

import (
	"log/slog"
	"slices"
	"sync"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	domainLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	"github.com/wensboy/lightpc/backend/domain/engine/store"
)

var (
	_ domainEngine.Engine    = (*FanoutLog)(nil)
	_ domainLog.FanoutClient = (*FanoutLog)(nil)
	_ domainLog.FanoutServer = (*FanoutLog)(nil)
)

// FanoutOption is a critical construction parameter for FanoutLog.
type FanoutOption func(*FanoutLog)

// WithFanoutHandler appends handler to the handler set bound to name before the engine starts.
func WithFanoutHandler(name string, handler slog.Handler) FanoutOption {
	return func(fl *FanoutLog) {
		if name == "" || handler == nil {
			return
		}
		fl.handlers[name] = append(fl.handlers[name], handler)
	}
}

// FanoutLog implements the fanout log engine: one record is delivered to every handler bound to a name.
type FanoutLog struct {
	handlers map[string][]slog.Handler

	mu     sync.RWMutex
	ectx   domainEngine.EngineContext
	state  domainEngine.EngineState
	closed bool
}

// NewFanoutLog builds a FanoutLog. Handlers may be preset with WithFanoutHandler or registered later.
func NewFanoutLog(opts ...FanoutOption) *FanoutLog {
	fl := &FanoutLog{handlers: map[string][]slog.Handler{}, state: domainEngine.EnginePending}
	for _, opt := range opts {
		opt(fl)
	}
	return fl
}

// Group implements domainEngine.Engine.
func (fl *FanoutLog) Group() string { return domainLog.GroupLog }

// Name implements domainEngine.Engine.
func (fl *FanoutLog) Name() string { return domainLog.NameFanout }

// Deps implements domainEngine.Engine.
func (fl *FanoutLog) Deps() []string { return []string{store.StoreFs} }

// State implements domainEngine.Engine.
func (fl *FanoutLog) State() domainEngine.EngineState {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	return fl.state
}

// Init implements domainEngine.Engine: it keeps the engine context and becomes ready.
func (fl *FanoutLog) Init(eCtx domainEngine.EngineContext) error {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.ectx = eCtx
	fl.state = domainEngine.EngineReady
	return nil
}

// Start implements domainEngine.Engine. FanoutLog serves lookups synchronously, so it only moves the state.
func (fl *FanoutLog) Start() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.state = domainEngine.EngineRunning
	return nil
}

// Restart implements domainEngine.Engine by reopening with the retained context.
func (fl *FanoutLog) Restart() error {
	fl.mu.Lock()
	fl.closed = false
	eCtx := fl.ectx
	fl.mu.Unlock()
	return fl.Init(eCtx)
}

// Stop implements domainEngine.Engine. Registered handlers are kept so a restart needs no rebinding.
func (fl *FanoutLog) Stop() error {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.closed = true
	fl.state = domainEngine.EngineStopped
	return nil
}

// Logger implements domainLog.FanoutClient. A non empty name narrows the broadcast to its own
// handlers, while an empty name addresses every registered handler, which is what the slog default
// logger relies on for development output.
func (fl *FanoutLog) Logger(name string) (*slog.Logger, bool) {
	fl.mu.RLock()
	var handlers []slog.Handler
	if name == "" {
		size := 0
		for _, list := range fl.handlers {
			size += len(list)
		}
		handlers = make([]slog.Handler, 0, size)
		for _, list := range fl.handlers {
			handlers = append(handlers, list...)
		}
	} else {
		handlers = slices.Clone(fl.handlers[name])
	}
	fl.mu.RUnlock()
	if len(handlers) == 0 {
		return nil, false
	}
	return slog.New(slog.NewMultiHandler(handlers...)), true
}

// Register implements domainLog.FanoutServer by appending handler to the set bound to name.
func (fl *FanoutLog) Register(name string, handler slog.Handler) {
	if name == "" || handler == nil {
		return
	}
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.handlers[name] = append(fl.handlers[name], handler)
}
