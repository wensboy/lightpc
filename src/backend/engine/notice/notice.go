package engineNoticeImpl

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineNotice "github.com/wensboy/lightpc/backend/domain/engine/notice"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

var (
	_ domainEngine.Engine = (*Notice)(nil)
	_ engineNotice.Client = (*Notice)(nil)
)

// NoticeOption is a critical construction parameter for Notice.
type NoticeOption func(*Notice)

// WithNoticeSize sets the ring length. It is the caller's job to read that value from the document: the engine never
// touches the configuration, so it can be constructed from a literal in a test just as well as from a file.
func WithNoticeSize(size int) NoticeOption {
	return func(n *Notice) {
		if size <= 0 {
			return
		}
		n.size = size
	}
}

// Notice implements the notice engine: a fixed ring of short lived notifications plus the bus announcement that
// makes a toast appear.
//
// A ring rather than a map: notices are few and short lived, the interface only ever shows the newest handful, and a
// fixed length means the memory a busy process uses cannot grow without bound. When the ring is full the oldest entry
// is dropped, which is the behaviour a terminal scrollback has and what someone expects from a notification feed.
type Notice struct {
	ring []engineNotice.Record
	size int
	next uint64

	mu     sync.RWMutex
	ectx   domainEngine.EngineContext
	state  domainEngine.EngineState
	closed bool
}

// NewNotice builds a Notice. A size left at zero falls back to the package default.
func NewNotice(opts ...NoticeOption) *Notice {
	n := &Notice{state: domainEngine.EnginePending}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// Group implements domainEngine.Engine.
func (n *Notice) Group() string { return domainEngine.GroupNotice }

// Name implements domainEngine.Engine.
func (n *Notice) Name() string { return domainEngine.NameNotice }

// Deps implements domainEngine.Engine: the engine owns everything it needs, so it depends on nothing.
func (n *Notice) Deps() []string { return nil }

// State implements domainEngine.Engine.
func (n *Notice) State() domainEngine.EngineState {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.state
}

// Init implements domainEngine.Engine: it keeps the engine context, sizes the ring and becomes ready.
//
// The size is checked here rather than filled here: the value arrives through WithNoticeSize, and an unusable one is
// replaced by the default at the point of use instead of being repaired during setup.
func (n *Notice) Init(eCtx domainEngine.EngineContext) error {
	if eCtx.AppCtx != nil {
		interError.RegisterAll(eCtx.AppCtx.ErrorContext, engineNotice.Errors...)
	}

	n.mu.Lock()
	n.ectx = eCtx
	if n.size <= 0 {
		n.size = defaultBuf
	}
	n.ring = make([]engineNotice.Record, 0, n.size)
	n.next = 0
	n.closed = false
	n.state = domainEngine.EngineReady
	size := n.size
	n.mu.Unlock()

	slog.Default().Warn("notice engine ready", "buf", size)
	return nil
}

// Start implements domainEngine.Engine. The engine serves calls synchronously, so it only moves the state.
func (n *Notice) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.state = domainEngine.EngineRunning
	return nil
}

// Restart implements domainEngine.Engine by rebuilding the ring from the context the engine already holds.
func (n *Notice) Restart() error {
	n.mu.Lock()
	eCtx := n.ectx
	n.mu.Unlock()
	return n.Init(eCtx)
}

// Stop implements domainEngine.Engine. The ring is process state, so stopping drops what is in it.
func (n *Notice) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.closed = true
	n.ring = n.ring[:0]
	n.state = domainEngine.EngineStopped
	return nil
}

// Notify implements engineNotice.Client.
//
// The record is stored first and announced second. An announcement that arrives before the entry exists would leave
// the interface showing a toast it cannot find in the sheet, which is the one ordering that must not happen.
func (n *Notice) Notify(input engineNotice.Input) (engineNotice.Record, error) {
	if input.Title == "" {
		return engineNotice.Record{}, interError.Retrieve(engineNotice.ErrKeyInvalidArgument).
			Wrap(fmt.Errorf("notice: title is required"))
	}
	if input.Level == "" {
		input.Level = engineNotice.LevelInfo
	}
	if !input.Level.Known() {
		return engineNotice.Record{}, interError.Retrieve(engineNotice.ErrKeyInvalidArgument).
			Wrap(fmt.Errorf("notice: unknown level %q", input.Level))
	}

	n.mu.Lock()
	n.next++
	record := engineNotice.Record{
		ID:        n.next,
		Level:     input.Level,
		Title:     input.Title,
		Content:   input.Content,
		Source:    input.Source,
		Timestamp: time.Now().UnixMilli(),
	}
	if len(n.ring) >= n.size {
		// The ring is full, so the oldest entry gives way. A copy shift keeps the order and avoids a second index.
		copy(n.ring, n.ring[1:])
		n.ring[len(n.ring)-1] = record
	} else {
		n.ring = append(n.ring, record)
	}
	bus := n.ectx.Ebus
	n.mu.Unlock()

	if bus != nil {
		bus.Emit(domainBus.EventNoticeSend, record)
	}
	return record, nil
}

// History implements engineNotice.Client, newest first.
//
// The copy matters: a caller may hold the slice while another goroutine publishes, and the ring keeps mutating.
func (n *Notice) History() []engineNotice.Record {
	n.mu.RLock()
	defer n.mu.RUnlock()

	history := make([]engineNotice.Record, 0, len(n.ring))
	for index := len(n.ring) - 1; index >= 0; index-- {
		history = append(history, n.ring[index])
	}
	return history
}

// Remove implements engineNotice.Client.
func (n *Notice) Remove(id uint64) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	for index, record := range n.ring {
		if record.ID != id {
			continue
		}
		n.ring = append(n.ring[:index], n.ring[index+1:]...)
		return true
	}
	return false
}

// Clear implements engineNotice.Client.
func (n *Notice) Clear() int {
	n.mu.Lock()
	defer n.mu.Unlock()

	removed := len(n.ring)
	n.ring = n.ring[:0]
	return removed
}
