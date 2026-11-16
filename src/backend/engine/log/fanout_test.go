package logimpl

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	domainLog "github.com/wensboy/lightpc/backend/domain/engine/log"
)

func TestFanoutLogBroadcastsToEveryHandler(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	fanoutLog := NewFanoutLog(
		WithFanoutHandler("app", slog.NewTextHandler(first, nil)),
		WithFanoutHandler("app", slog.NewTextHandler(second, nil)),
	)

	logger, ok := fanoutLog.Logger("app")
	if !ok {
		t.Fatal("app logger must be resolved")
	}
	logger.Info("hello")

	if first.Len() == 0 || second.Len() == 0 {
		t.Fatalf("first = %d, second = %d", first.Len(), second.Len())
	}
}

func TestFanoutLogMatchesExactName(t *testing.T) {
	fanoutLog := NewFanoutLog(WithFanoutHandler("app", slog.NewTextHandler(io.Discard, nil)))

	if _, ok := fanoutLog.Logger("ap"); ok {
		t.Fatal("a prefix must not match")
	}
	if _, ok := fanoutLog.Logger(""); ok {
		t.Fatal("an empty name must not match")
	}
}

func TestFanoutLogRegisterAppends(t *testing.T) {
	fanoutLog := NewFanoutLog(WithFanoutHandler("app", slog.NewTextHandler(io.Discard, nil)))

	fanoutLog.Register("app", slog.NewTextHandler(io.Discard, nil))

	if count := len(fanoutLog.handlers["app"]); count != 2 {
		t.Fatalf("count = %d", count)
	}
}

func TestFanoutLogIgnoresInvalidRegistration(t *testing.T) {
	fanoutLog := NewFanoutLog()

	fanoutLog.Register("", slog.NewTextHandler(io.Discard, nil))
	fanoutLog.Register("app", nil)

	if len(fanoutLog.handlers) != 0 {
		t.Fatalf("handlers = %v", fanoutLog.handlers)
	}
}

func TestFanoutLogLifecycle(t *testing.T) {
	fanoutLog := NewFanoutLog(WithFanoutHandler("app", slog.NewTextHandler(io.Discard, nil)))

	if fanoutLog.Group() != domainLog.GroupLog || fanoutLog.Name() != domainLog.NameFanout {
		t.Fatalf("identity = %s.%s", fanoutLog.Group(), fanoutLog.Name())
	}
	if fanoutLog.State() != domainEngine.EnginePending {
		t.Fatalf("state = %v", fanoutLog.State())
	}
	if err := fanoutLog.Init(domainEngine.EngineContext{}); err != nil {
		t.Fatal(err)
	}
	if fanoutLog.State() != domainEngine.EngineReady {
		t.Fatalf("state = %v", fanoutLog.State())
	}
	if err := fanoutLog.Start(); err != nil {
		t.Fatal(err)
	}
	if fanoutLog.State() != domainEngine.EngineRunning {
		t.Fatalf("state = %v", fanoutLog.State())
	}
	if err := fanoutLog.Stop(); err != nil {
		t.Fatal(err)
	}
	if fanoutLog.State() != domainEngine.EngineStopped {
		t.Fatalf("state = %v", fanoutLog.State())
	}
	if _, ok := fanoutLog.Logger("app"); !ok {
		t.Fatal("handlers must survive a stop")
	}
}
