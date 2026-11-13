package ebus

import (
	"sync"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
)

// EventBus 是一个标准的事件总线.
type EventBus struct {
	mu sync.RWMutex

	registry map[string][]domainBus.EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		registry: make(map[string][]domainBus.EventHandler),
	}
}

// On 用于对事件 e 进行监听.
func (eb *EventBus) On(e string, eh domainBus.EventHandler) {
	eb.mu.Lock()
	eb.registry[e] = append(eb.registry[e], eh)
	eb.mu.Unlock()
}

// Emit 用于触发事件 e, 并传递负载.
func (eb *EventBus) Emit(e string, payload any) {
	eb.mu.RLock()
	handlers := append([]domainBus.EventHandler(nil), eb.registry[e]...)
	eb.mu.RUnlock()
	for _, eh := range handlers {
		eh(payload)
	}
}
