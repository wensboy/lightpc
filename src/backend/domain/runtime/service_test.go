package runtime

import (
	"slices"
	"testing"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
)

type stubService struct {
	group   string
	order   *[]string
	initErr error
	engines map[string]domainEngine.Engine
	appCtx  *domainContext.AppContext
	deinits int
}

func (s *stubService) Group() string { return s.group }

func (s *stubService) Init(serviceCtx domainService.ServiceContext) error {
	s.engines = serviceCtx.Engines
	s.appCtx = serviceCtx.AppCtx
	if s.initErr != nil {
		return s.initErr
	}
	if s.order != nil {
		*s.order = append(*s.order, "init:"+s.group)
	}
	return nil
}

func (s *stubService) Deinit() error {
	s.deinits++
	if s.order != nil {
		*s.order = append(*s.order, "deinit:"+s.group)
	}
	return nil
}

func TestServiceRuntimeRegisterByGroup(t *testing.T) {
	serviceRuntime := NewServiceRuntime()
	serviceRuntime.Register(&stubService{group: "svc.a"}, nil, &stubService{group: "svc.b"})

	if len(serviceRuntime.registry) != 2 {
		t.Fatalf("registry = %d, want 2", len(serviceRuntime.registry))
	}
	if _, ok := serviceRuntime.registry["svc.a"]; !ok {
		t.Fatal("a service must be registered by its group")
	}

	replaced := &stubService{group: "svc.a"}
	serviceRuntime.Register(replaced)
	if serviceRuntime.registry["svc.a"] != replaced {
		t.Fatal("the latest service must replace the previous one in the same group")
	}
}

func TestServiceRuntimeRunAndExit(t *testing.T) {
	var order []string
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(&stubEngine{group: "core", name: "a"})
	appRuntime := NewAppRuntime().WithEngineRuntime(engineRuntime)

	serviceRuntime := NewServiceRuntime().WithAppRuntime(appRuntime)
	service := &stubService{group: "svc", order: &order}
	serviceRuntime.Register(service)

	serviceRuntime.Run()

	if len(service.engines) != 1 || service.engines["core.a"] == nil {
		t.Fatalf("engines = %v", service.engines)
	}
	if service.appCtx == nil {
		t.Fatal("the app context must be passed to the service")
	}

	serviceRuntime.Exit()
	if !slices.Equal(order, []string{"init:svc", "deinit:svc"}) {
		t.Fatalf("order = %v", order)
	}
	if service.deinits != 1 {
		t.Fatalf("deinits = %d, want 1", service.deinits)
	}
}

func TestServiceRuntimeCallsEventFn(t *testing.T) {
	called := false
	serviceRuntime := NewServiceRuntime().WithEventFn(func(domainService.ServiceEventContext) {
		called = true
	})
	serviceRuntime.Run()
	if !called {
		t.Fatal("the event function must be called on run")
	}
}
