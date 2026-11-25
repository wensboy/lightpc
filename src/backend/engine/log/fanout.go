package engineLogImpl

import (
	"log/slog"
	"slices"
	"sync"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

var (
	_ domainEngine.Engine    = (*FanoutLog)(nil)
	_ engineLog.FanoutClient = (*FanoutLog)(nil)
	_ engineLog.FanoutServer = (*FanoutLog)(nil)
)

// FanoutOption is a critical construction parameter for FanoutLog.
type FanoutOption func(*FanoutLog)

// WithFanoutHandler appends handler to the set bound to name immediately; it is meant for programmatic
// wiring.
func WithFanoutHandler(name string, handler slog.Handler) FanoutOption {
	return func(fl *FanoutLog) {
		if name == "" || handler == nil {
			return
		}
		fl.handlers[name] = append(fl.handlers[name], handler)
	}
}

// WithFanoutEntry records one configured entry. Its handlers are opened during Init, after the store
// engine has created the root directory the log files live in.
func WithFanoutEntry(entry engineLog.LogEntry) FanoutOption {
	return func(fl *FanoutLog) {
		if entry.Name == "" {
			return
		}
		fl.entries = append(fl.entries, entry)
	}
}

// FanoutLog implements the fanout log engine: one record is delivered to every handler bound to a name.
type FanoutLog struct {
	handlers map[string][]slog.Handler
	entries  []engineLog.LogEntry

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
func (fl *FanoutLog) Group() string { return domainEngine.GroupLog }

// Name implements domainEngine.Engine.
func (fl *FanoutLog) Name() string { return domainEngine.NameFanout }

// Deps implements domainEngine.Engine: log files live below the store root.
func (fl *FanoutLog) Deps() []string { return []string{domainEngine.GroupStore} }

// State implements domainEngine.Engine.
func (fl *FanoutLog) State() domainEngine.EngineState {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	return fl.state
}

// Init implements domainEngine.Engine: it keeps the engine context, opens every configured handler,
// becomes ready and publishes the empty name aggregate as the slog default.
func (fl *FanoutLog) Init(eCtx domainEngine.EngineContext) error {
	interError.RegisterAll(eCtx.AppCtx, engineLog.Errors...)

	fl.mu.Lock()
	fl.ectx = eCtx
	pools := poolSourceOf(eCtx.Engines)
	for _, entry := range entriesOf(fl.entries) {
		fl.handlers[entry.name] = append(fl.handlers[entry.name], handlersFor(entry, pools)...)
	}
	fl.state = domainEngine.EngineReady
	fl.mu.Unlock()

	if logger, ok := fl.Logger(""); ok {
		slog.SetDefault(logger)
	}
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

// Logger implements engineLog.FanoutClient. A non empty name narrows the broadcast to its own
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

// Register implements engineLog.FanoutServer by appending handler to the set bound to name.
func (fl *FanoutLog) Register(name string, handler slog.Handler) {
	if name == "" || handler == nil {
		return
	}
	fl.mu.Lock()
	defer fl.mu.Unlock()
	fl.handlers[name] = append(fl.handlers[name], handler)
}

// HookContext implements the log hook contract: it reports the configured entries so a hook can prepare
// their outputs.
func (fl *FanoutLog) HookContext() engineLog.FanoutHookContext {
	fl.mu.RLock()
	defer fl.mu.RUnlock()
	return engineLog.FanoutHookContext{
		LogEntries:    fl.entries,
		EngineContext: fl.ectx,
	}
}
