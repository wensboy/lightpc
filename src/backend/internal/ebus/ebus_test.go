package ebus

import (
	"sync"
	"testing"
	"time"
)

func TestEventBusOnAndEmit(t *testing.T) {
	bus := NewEventBus()
	var mu sync.Mutex
	var received []string
	bus.On("topic", func(payload any) {
		mu.Lock()
		received = append(received, "first:"+payload.(string))
		mu.Unlock()
	})
	bus.On("topic", func(payload any) {
		mu.Lock()
		received = append(received, "second:"+payload.(string))
		mu.Unlock()
	})

	bus.Emit("topic", "x")

	if len(received) != 2 || received[0] != "first:x" || received[1] != "second:x" {
		t.Fatalf("received = %v", received)
	}
}

func TestEventBusEmitWithoutHandlers(t *testing.T) {
	NewEventBus().Emit("missing", nil)
}

func TestEventBusHandlerCanRegisterDuringEmit(t *testing.T) {
	bus := NewEventBus()
	bus.On("topic", func(any) {
		bus.On("topic", func(any) {})
	})

	finished := make(chan struct{})
	go func() {
		bus.Emit("topic", nil)
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("Emit deadlocked while a handler registered a listener")
	}
}
