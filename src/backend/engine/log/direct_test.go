package logimpl

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	"github.com/wensboy/lightpc/backend/domain/engine/log"
	"github.com/wensboy/lightpc/backend/internal/cast"
)

func TestDirectLogRegistersHandler(t *testing.T) {
	buffer := &bytes.Buffer{}
	directLog := NewDirectLog(WithDirectHandler("app", slog.NewTextHandler(buffer, nil)))

	logger, ok := directLog.Logger("app")
	if !ok {
		t.Fatal("app logger must be resolved")
	}
	logger.Info("hello")

	if buffer.Len() == 0 {
		t.Fatal("the handler must receive the record")
	}
}

func TestDirectLogMatchesExactName(t *testing.T) {
	directLog := NewDirectLog(WithDirectHandler("app", slog.NewTextHandler(io.Discard, nil)))

	if _, ok := directLog.Logger("ap"); ok {
		t.Fatal("a prefix must not match")
	}
	if _, ok := directLog.Logger(""); ok {
		t.Fatal("an empty name must not match")
	}
}

func TestDirectLogRegisterReplacesHandler(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	directLog := NewDirectLog(WithDirectHandler("app", slog.NewTextHandler(first, nil)))

	directLog.Register("app", slog.NewTextHandler(second, nil))
	logger, ok := directLog.Logger("app")
	if !ok {
		t.Fatal("app logger must be resolved")
	}
	logger.Info("hello")

	if first.Len() != 0 || second.Len() == 0 {
		t.Fatalf("first = %d, second = %d", first.Len(), second.Len())
	}
}

func TestDirectLogIgnoresInvalidRegistration(t *testing.T) {
	directLog := NewDirectLog()

	directLog.Register("", slog.NewTextHandler(io.Discard, nil))
	directLog.Register("app", nil)

	if len(directLog.handlers) != 0 {
		t.Fatalf("handlers = %v", directLog.handlers)
	}
}

func TestDirectLogLifecycle(t *testing.T) {
	directLog := NewDirectLog(WithDirectHandler("app", slog.NewTextHandler(io.Discard, nil)))

	if directLog.Group() != log.GroupLog || directLog.Name() != log.NameDirect {
		t.Fatalf("identity = %s.%s", directLog.Group(), directLog.Name())
	}
	if directLog.State() != domainEngine.EnginePending {
		t.Fatalf("state = %v", directLog.State())
	}
	if err := directLog.Init(domainEngine.EngineContext{}); err != nil {
		t.Fatal(err)
	}
	if directLog.State() != domainEngine.EngineReady {
		t.Fatalf("state = %v", directLog.State())
	}
	if err := directLog.Start(); err != nil {
		t.Fatal(err)
	}
	if directLog.State() != domainEngine.EngineRunning {
		t.Fatalf("state = %v", directLog.State())
	}
	if err := directLog.Stop(); err != nil {
		t.Fatal(err)
	}
	if directLog.State() != domainEngine.EngineStopped {
		t.Fatalf("state = %v", directLog.State())
	}
}

func TestDirectLogExposesCapabilitiesThroughCast(t *testing.T) {
	var engine domainEngine.Engine = NewDirectLog(WithDirectHandler("app", slog.NewTextHandler(io.Discard, nil)))

	client, ok := cast.As[log.DirectClient](engine)
	if !ok {
		t.Fatal("the engine must be convertible to DirectClient")
	}
	if _, ok := client.Logger("app"); !ok {
		t.Fatal("app logger must be resolved through the interface")
	}

	server, ok := cast.As[log.DirectServer](engine)
	if !ok {
		t.Fatal("the engine must be convertible to DirectServer")
	}
	server.Register("worker", slog.NewTextHandler(io.Discard, nil))
	if _, ok := client.Logger("worker"); !ok {
		t.Fatal("a runtime registration must be visible to the client")
	}
}

func TestDirectLogRestartKeepsHandlers(t *testing.T) {
	directLog := NewDirectLog(WithDirectHandler("app", slog.NewTextHandler(io.Discard, nil)))
	if err := directLog.Init(domainEngine.EngineContext{}); err != nil {
		t.Fatal(err)
	}
	if err := directLog.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := directLog.Restart(); err != nil {
		t.Fatal(err)
	}
	if directLog.State() != domainEngine.EngineReady {
		t.Fatalf("state = %v", directLog.State())
	}
	if _, ok := directLog.Logger("app"); !ok {
		t.Fatal("handlers must survive a restart")
	}
}
