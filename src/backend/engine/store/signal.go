package storeimpl

import (
	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	"github.com/wensboy/lightpc/backend/domain/engine/store"
	"github.com/wensboy/lightpc/backend/internal/cast"
)

// watchStopSignal closes the engine when the bus reports a stop event addressed to it.
func watchStopSignal(ebus domainBus.Bus, name string, stop func() error) {
	if ebus == nil {
		return
	}
	ebus.On(domainBus.EventEngineStop, func(payload any) {
		source, ok := cast.To[map[string]any](payload)
		if !ok {
			return
		}
		event := struct {
			Group string `json:"group"`
			Name  string `json:"name"`
		}{}
		cast.ToStruct(source, &event, "json")
		if event.Group != store.GroupStore || event.Name != name {
			return
		}
		_ = stop()
	})
}
