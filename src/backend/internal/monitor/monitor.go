// Package monitor runs a handler on a fixed interval, mirroring setInterval.
package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// defaultInterval is used when New receives a non positive interval.
const defaultInterval = time.Second

// Handler runs on every tick. It returns an arbitrary payload and reports whether this round
// succeeded; the monitor only inspects the error to pick the result callback.
type Handler func() (any, error)

// Success receives the payload of a successful tick.
type Success func(value any)

// Failure receives the error of a failed tick, including a recovered handler panic.
type Failure func(err error)

// Monitor invokes Handler on a fixed interval until Stop. Calls never overlap: a tick that arrives
// while the previous call is still running is dropped.
//
// The result callbacks are fixed at construction and cannot be chained or replaced, so every tick
// reaches exactly one of them: Success on a nil error, Failure otherwise.
type Monitor struct {
	interval  time.Duration
	handler   Handler
	onSuccess Success
	onFailure Failure

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

// New builds a Monitor. A non positive interval falls back to one second, and a nil callback simply
// ignores that outcome.
func New(interval time.Duration, handler Handler, onSuccess Success, onFailure Failure) *Monitor {
	if interval <= 0 {
		interval = defaultInterval
	}
	return &Monitor{
		interval:  interval,
		handler:   handler,
		onSuccess: onSuccess,
		onFailure: onFailure,
	}
}

// Start begins ticking. Calling Start again while the monitor is running has no effect.
func (m *Monitor) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	m.cancel = cancel
	m.done = done

	go func() {
		defer close(done)
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				func() {
					defer func() {
						if recovered := recover(); recovered != nil {
							if m.onFailure != nil {
								m.onFailure(fmt.Errorf("monitor handler panicked: %v", recovered))
							}
						}
					}()
					value, err := m.handler()
					if err != nil {
						if m.onFailure != nil {
							m.onFailure(err)
						}
						return
					}
					if m.onSuccess != nil {
						m.onSuccess(value)
					}
				}()
			}
		}
	}()
}

// Stop cancels ticking and waits for the in flight handler call to return. It is idempotent, and a
// handler that blocks delays Stop for as long as it runs.
func (m *Monitor) Stop() {
	m.mu.Lock()
	cancel, done := m.cancel, m.done
	m.cancel, m.done = nil, nil
	m.mu.Unlock()

	if cancel == nil {
		return
	}
	cancel()
	<-done
}
