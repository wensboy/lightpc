package runtime

import (
	"errors"
	"slices"
	"testing"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	"github.com/wensboy/lightpc/backend/internal/cast"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

type stubEngine struct {
	group    string
	name     string
	deps     []string
	state    domainEngine.EngineState
	order    *[]string
	initCtx  domainEngine.EngineContext
	initErr  error
	startErr error
	stopErr  error
}

func (s *stubEngine) Group() string                   { return s.group }
func (s *stubEngine) Name() string                    { return s.name }
func (s *stubEngine) Deps() []string                  { return s.deps }
func (s *stubEngine) State() domainEngine.EngineState { return s.state }

func (s *stubEngine) Init(eCtx domainEngine.EngineContext) error {
	s.initCtx = eCtx
	if s.initErr != nil {
		s.state = domainEngine.EngineError
		return s.initErr
	}
	s.state = domainEngine.EngineReady
	if s.order != nil {
		*s.order = append(*s.order, "init:"+s.name)
	}
	return nil
}

func (s *stubEngine) Start() error {
	if s.startErr != nil {
		s.state = domainEngine.EngineError
		return s.startErr
	}
	s.state = domainEngine.EngineRunning
	if s.order != nil {
		*s.order = append(*s.order, "start:"+s.name)
	}
	return nil
}

func (s *stubEngine) Restart() error { return nil }

func (s *stubEngine) Stop() error {
	if s.stopErr != nil {
		s.state = domainEngine.EngineError
		return s.stopErr
	}
	s.state = domainEngine.EngineStopped
	if s.order != nil {
		*s.order = append(*s.order, "stop:"+s.name)
	}
	return nil
}

func TestEngineRuntimeRegisterAndEngine(t *testing.T) {
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(
		&stubEngine{group: "core", name: "b"},
		&stubEngine{group: "core", name: "a"},
		&stubEngine{group: "extra", name: "c"},
		nil,
	)

	if engines := engineRuntime.Engine("", ""); len(engines) != 0 {
		t.Fatalf("empty group returned %d engines", len(engines))
	}
	core := engineRuntime.Engine("core", "")
	if len(core) != 2 || core[0].Name() != "a" || core[1].Name() != "b" {
		t.Fatalf("core = %v", core)
	}
	single := engineRuntime.Engine("core", "b")
	if len(single) != 1 || single[0].Name() != "b" {
		t.Fatalf("core.b = %v", single)
	}
	if engines := engineRuntime.Engine("core", "missing"); len(engines) != 0 {
		t.Fatalf("core.missing returned %d engines", len(engines))
	}
	if engines := engineRuntime.Engine("corex", ""); len(engines) != 0 {
		t.Fatalf("corex returned %d engines", len(engines))
	}
}

func TestEngineRuntimeRunOrder(t *testing.T) {
	var order []string
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(
		&stubEngine{group: "core", name: "b", deps: []string{"core.a"}, order: &order},
		&stubEngine{group: "core", name: "a", order: &order},
	)

	if err := engineRuntime.Run(); err != nil {
		t.Fatal(err)
	}
	want := []string{"init:a", "start:a", "init:b", "start:b"}
	if !slices.Equal(order, want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
}

func TestEngineRuntimeRunCycle(t *testing.T) {
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(
		&stubEngine{group: "core", name: "a", deps: []string{"core.b"}},
		&stubEngine{group: "core", name: "b", deps: []string{"core.a"}},
	)

	if err := engineRuntime.Run(); err == nil {
		t.Fatal("a dependency cycle must return an error")
	}
}

func TestEngineRuntimeRunInitError(t *testing.T) {
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(&stubEngine{group: "core", name: "a", initErr: errors.New("init failed")})

	if err := engineRuntime.Run(); err == nil {
		t.Fatal("an init failure must be returned")
	}
}

func TestEngineRuntimeRunStartError(t *testing.T) {
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(&stubEngine{group: "core", name: "a", startErr: errors.New("start failed")})

	if err := engineRuntime.Run(); err == nil {
		t.Fatal("a start failure must be returned")
	}
}

func TestEngineRuntimeRunEmitsParsablePayload(t *testing.T) {
	type engineEvent struct {
		Group string `json:"group"`
		Name  string `json:"name"`
		State string `json:"state"`
	}
	var received []engineEvent
	engineRuntime := NewEngineRuntime()
	engineRuntime.ebus.On(domainBus.EventEngineReady, func(payload any) {
		event := engineEvent{}
		cast.ToStruct(payload.(map[string]any), &event, "json")
		received = append(received, event)
	})
	engineRuntime.Register(&stubEngine{group: "core", name: "a"})

	if err := engineRuntime.Run(); err != nil {
		t.Fatal(err)
	}
	if len(received) != 1 || received[0].Group != "core" || received[0].Name != "a" || received[0].State != string(domainEngine.EngineRunning) {
		t.Fatalf("received = %+v", received)
	}
}

func TestEngineRuntimeExitReverseOrder(t *testing.T) {
	var order []string
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(
		&stubEngine{group: "core", name: "a", order: &order},
		&stubEngine{group: "core", name: "b", deps: []string{"core.a"}, order: &order},
	)

	if err := engineRuntime.Run(); err != nil {
		t.Fatal(err)
	}
	order = nil
	if err := engineRuntime.Exit(); err != nil {
		t.Fatal(err)
	}
	want := []string{"stop:b", "stop:a"}
	if !slices.Equal(order, want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
}

func TestEngineRuntimeExitReturnsFirstError(t *testing.T) {
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(
		&stubEngine{group: "core", name: "a", stopErr: errors.New("stop failed")},
		&stubEngine{group: "core", name: "b", deps: []string{"core.a"}},
	)

	if err := engineRuntime.Run(); err != nil {
		t.Fatal(err)
	}
	if err := engineRuntime.Exit(); err == nil {
		t.Fatal("a stop failure must be returned")
	}
}

func TestEngineRuntimeCallsEventFn(t *testing.T) {
	var registry map[string]domainEngine.Engine
	engineRuntime := NewEngineRuntime()
	engineRuntime.WithEventFn(func(eec domainEngine.EngineEventContext) {
		registry = eec.Registry
	})
	engineRuntime.Register(&stubEngine{group: "core", name: "a"})

	if err := engineRuntime.Run(); err != nil {
		t.Fatal(err)
	}
	if len(registry) != 1 || registry["core.a"] == nil {
		t.Fatalf("registry = %v", registry)
	}
}

func TestEngineRuntimeRegisterSkipsEmptyIdentity(t *testing.T) {
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(
		&stubEngine{group: "", name: "a"},
		&stubEngine{group: "core", name: ""},
		&stubEngine{group: "", name: ""},
		&stubEngine{group: "core", name: "ok"},
	)

	if engines := engineRuntime.Engine("core", ""); len(engines) != 1 || engines[0].Name() != "ok" {
		t.Fatalf("engines = %v", engines)
	}
	if engines := engineRuntime.Engine("", ""); engines != nil {
		t.Fatalf("engines = %v", engines)
	}
}

func TestEngineRuntimeRunFailsOnMissingDependency(t *testing.T) {
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(&stubEngine{group: "core", name: "a", deps: []string{"core.missing"}})

	if err := engineRuntime.Run(); !errors.Is(err, interError.Retrieve(ErrKeyMissingDependency)) {
		t.Fatalf("err = %v", err)
	}
}

func TestEngineRuntimeRunInjectsDependencies(t *testing.T) {
	dependency := &stubEngine{group: "core", name: "b"}
	dependent := &stubEngine{group: "core", name: "a", deps: []string{"core.b"}}
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(dependent, dependency)

	if err := engineRuntime.Run(); err != nil {
		t.Fatal(err)
	}
	if dependent.initCtx.Engines["core.b"] != dependency {
		t.Fatalf("engines = %v", dependent.initCtx.Engines)
	}
	if len(dependency.initCtx.Engines) != 0 {
		t.Fatalf("engines = %v", dependency.initCtx.Engines)
	}
}

func TestEngineRuntimeRunInjectsGroupDependencies(t *testing.T) {
	first := &stubEngine{group: "core", name: "b"}
	second := &stubEngine{group: "core", name: "c"}
	dependent := &stubEngine{group: "app", name: "a", deps: []string{"core"}}
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(first, second, dependent)

	if err := engineRuntime.Run(); err != nil {
		t.Fatal(err)
	}
	if len(dependent.initCtx.Engines) != 2 {
		t.Fatalf("engines = %v", dependent.initCtx.Engines)
	}
	if dependent.initCtx.Engines["core.b"] != first || dependent.initCtx.Engines["core.c"] != second {
		t.Fatalf("engines = %v", dependent.initCtx.Engines)
	}
}

func TestEngineRuntimeRunFailsOnEmptyGroupDependency(t *testing.T) {
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(&stubEngine{group: "app", name: "a", deps: []string{"missing"}})

	if err := engineRuntime.Run(); !errors.Is(err, interError.Retrieve(ErrKeyMissingDependency)) {
		t.Fatalf("err = %v", err)
	}
}
