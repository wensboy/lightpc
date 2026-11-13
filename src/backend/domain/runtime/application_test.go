package runtime

import (
	"slices"
	"testing"
)

func TestAppRuntimeRunAndExitOrder(t *testing.T) {
	var order []string
	engineRuntime := NewEngineRuntime()
	engineRuntime.Register(&stubEngine{group: "core", name: "a", order: &order})
	serviceRuntime := NewServiceRuntime()
	serviceRuntime.Register(&stubService{group: "svc", order: &order})

	appRuntime := NewAppRuntime().
		WithEngineRuntime(engineRuntime).
		WithServiceRuntime(serviceRuntime)

	appRuntime.Run()
	if !slices.Equal(order, []string{"init:a", "start:a", "init:svc"}) {
		t.Fatalf("run order = %v", order)
	}

	order = nil
	appRuntime.Exit()
	if !slices.Equal(order, []string{"deinit:svc", "stop:a"}) {
		t.Fatalf("exit order = %v", order)
	}
}

func TestAppRuntimeWithoutRuntimes(t *testing.T) {
	NewAppRuntime().Run()
	NewAppRuntime().Exit()
}
