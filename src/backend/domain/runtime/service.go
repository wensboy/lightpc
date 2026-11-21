package runtime

import (
	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	"github.com/wensboy/lightpc/backend/internal/ebus"
)

// ServiceRuntime 用于统一管理运行时 Service 相关处理.
// Service 无状态, 编排 engines 实现逻辑.
type ServiceRuntime struct {
	registry map[string]domainService.Service

	// runtime
	appRuntime *AppRuntime

	// event bus in service runtime
	ebus   domainBus.Bus
	eFn    func(domainService.ServiceEventContext) // service 内总线监听
	appBus domainBus.Bus
}

func NewServiceRuntime() *ServiceRuntime {
	return &ServiceRuntime{
		registry: make(map[string]domainService.Service),
		ebus:     ebus.NewEventBus(),
	}
}

func (sr *ServiceRuntime) WithAppRuntime(ar *AppRuntime) *ServiceRuntime {
	sr.appRuntime = ar
	return sr
}

func (sr *ServiceRuntime) WithEventFn(efn func(domainService.ServiceEventContext)) *ServiceRuntime {
	sr.eFn = efn
	return sr
}

func (sr *ServiceRuntime) WithAppBus(appBus domainBus.Bus) *ServiceRuntime {
	sr.appBus = appBus
	return sr
}

func (sr *ServiceRuntime) Register(s ...domainService.Service) {
	for _, service := range s {
		if service == nil {
			continue
		}
		sr.registry[service.Group()] = service
	}
}

func (sr *ServiceRuntime) Run() {
	if sr.eFn != nil {
		sr.eFn(domainService.ServiceEventContext{Services: sr.registry, Ebus: sr.ebus})
	}
	engines := make(map[string]domainEngine.Engine)
	if sr.appRuntime != nil && sr.appRuntime.er != nil {
		sr.appRuntime.er.mu.RLock()
		// Services resolve concrete engines, so the hook wrappers are stripped here: the wrapper only owns
		// the lifecycle, never the behaviour a service depends on.
		for key, engine := range sr.appRuntime.er.registry {
			engines[key] = domainEngine.Unwrap(engine)
		}
		sr.appRuntime.er.mu.RUnlock()
	}
	serviceCtx := domainService.ServiceContext{Ebus: sr.ebus, Engines: engines}
	if sr.appRuntime != nil {
		serviceCtx.AppCtx = &sr.appRuntime.appCtx
	}
	for _, service := range sr.registry {
		_ = service.Init(serviceCtx)
	}
}

func (sr *ServiceRuntime) Exit() {
	for _, service := range sr.registry {
		_ = service.Deinit()
	}
}
