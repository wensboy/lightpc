package runtime

import (
	"maps"
	"slices"
	"strings"
	"sync"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"

	"github.com/wensboy/lightpc/backend/internal/ebus"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// ErrKeyDependencyCycle reports that engines depend on each other in a cycle.
const ErrKeyDependencyCycle = "error::engine.runtime::dependency_cycle"

// EngineRuntimeErrors lists every error published when the engine runtime starts.
var EngineRuntimeErrors = []interError.Option{
	{Key: ErrKeyDependencyCycle, Code: interError.ErrCodeUnknown, Msg: "engine runtime: dependency cycle detected"},
}

// EngineRuntime 管理所有引擎的生命周期.
type EngineRuntime struct {
	mu sync.RWMutex

	registry map[string]domainEngine.Engine

	// runtime ref
	appRuntime *AppRuntime

	// event bus in engine runtime
	ebus   domainBus.Bus
	eFn    func(domainEngine.EngineEventContext) // engine 内总线监听
	appBus domainBus.Bus
	// 拓扑序列, 退出时逆序终止
	execStack []domainEngine.Engine
}

func NewEngineRuntime() *EngineRuntime {
	return &EngineRuntime{
		ebus:     ebus.NewEventBus(),
		registry: map[string]domainEngine.Engine{},
	}
}

func (er *EngineRuntime) WithAppRuntime(ar *AppRuntime) *EngineRuntime {
	er.appRuntime = ar
	return er
}

func (er *EngineRuntime) WithEventFn(efn func(domainEngine.EngineEventContext)) *EngineRuntime {
	er.eFn = efn
	return er
}

func (er *EngineRuntime) WithAppBus(appBus domainBus.Bus) *EngineRuntime {
	er.appBus = appBus
	return er
}

func (er *EngineRuntime) Register(e ...domainEngine.Engine) {
	er.mu.Lock()
	defer er.mu.Unlock()
	for _, engine := range e {
		if engine == nil {
			continue
		}
		er.registry[engine.Group()+"."+engine.Name()] = engine
	}
}

func (er *EngineRuntime) Engine(group, name string) []domainEngine.Engine {
	er.mu.RLock()
	defer er.mu.RUnlock()
	if group == "" {
		return nil
	}
	if name != "" {
		engine, ok := er.registry[group+"."+name]
		if !ok {
			return nil
		}
		return []domainEngine.Engine{engine}
	}
	prefix := group + "."
	keys := make([]string, 0, len(er.registry))
	for key := range er.registry {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)
	engines := make([]domainEngine.Engine, 0, len(keys))
	for _, key := range keys {
		engines = append(engines, er.registry[key])
	}
	return engines
}

func (er *EngineRuntime) Run() error {
	type engineEvent struct {
		Group string `json:"group"`
		Name  string `json:"name"`
		State string `json:"state"`
	}

	engineCtx := domainEngine.EngineContext{Ebus: er.ebus}
	if er.appRuntime != nil {
		engineCtx.AppCtx = &er.appRuntime.appCtx
	}
	interError.RegisterAll(engineCtx.AppCtx, EngineRuntimeErrors...)

	er.mu.RLock()
	snapshot := make(map[string]domainEngine.Engine, len(er.registry))
	maps.Copy(snapshot, er.registry)
	er.mu.RUnlock()

	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	ordered := make([]domainEngine.Engine, 0, len(keys))
	resolved := make(map[string]struct{}, len(keys))
	for len(ordered) < len(keys) {
		progressed := false
		for _, key := range keys {
			if _, ok := resolved[key]; ok {
				continue
			}
			ready := true
			for _, dependency := range snapshot[key].Deps() {
				if _, registered := snapshot[dependency]; !registered {
					continue
				}
				if _, ok := resolved[dependency]; !ok {
					ready = false
					break
				}
			}
			if !ready {
				continue
			}
			ordered = append(ordered, snapshot[key])
			resolved[key] = struct{}{}
			progressed = true
		}
		if !progressed {
			return interError.Retrieve(ErrKeyDependencyCycle)
		}
	}

	er.mu.Lock()
	er.execStack = ordered
	er.mu.Unlock()

	if er.eFn != nil {
		er.eFn(domainEngine.EngineEventContext{Registry: snapshot, Ebus: er.ebus})
	}

	for _, engine := range ordered {
		event := engineEvent{Group: engine.Group(), Name: engine.Name()}
		payload := map[string]any{"group": event.Group, "name": event.Name, "state": ""}

		if err := engine.Init(engineCtx); err != nil {
			event.State = string(engine.State())
			payload["state"] = event.State
			er.ebus.Emit(domainBus.EventEngineError, payload)
			return err
		}
		event.State = string(engine.State())
		payload["state"] = event.State
		er.ebus.Emit(domainBus.EventEngineInit, payload)

		if err := engine.Start(); err != nil {
			event.State = string(engine.State())
			payload["state"] = event.State
			er.ebus.Emit(domainBus.EventEngineError, payload)
			return err
		}
		event.State = string(engine.State())
		payload["state"] = event.State
		er.ebus.Emit(domainBus.EventEngineStart, payload)
		er.ebus.Emit(domainBus.EventEngineReady, payload)
	}
	return nil
}

func (er *EngineRuntime) Exit() error {
	er.mu.Lock()
	ordered := er.execStack
	er.execStack = nil
	er.mu.Unlock()

	var firstErr error
	for index := len(ordered) - 1; index >= 0; index-- {
		engine := ordered[index]
		payload := map[string]any{
			"group": engine.Group(),
			"name":  engine.Name(),
			"state": string(engine.State()),
		}
		if err := engine.Stop(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			payload["state"] = string(engine.State())
			er.ebus.Emit(domainBus.EventEngineError, payload)
			continue
		}
		payload["state"] = string(engine.State())
		er.ebus.Emit(domainBus.EventEngineStop, payload)
	}
	return firstErr
}
