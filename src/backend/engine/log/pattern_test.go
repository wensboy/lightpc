package logimpl

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	domainLog "github.com/wensboy/lightpc/backend/domain/engine/log"
)

func TestPatternLogRoutesByGlob(t *testing.T) {
	httpBuffer := &bytes.Buffer{}
	workerBuffer := &bytes.Buffer{}
	patternLog := NewPatternLog(
		WithPatternHandler("app.*", slog.NewTextHandler(httpBuffer, nil)),
		WithPatternHandler("worker", slog.NewTextHandler(workerBuffer, nil)),
	)

	logger, ok := patternLog.Logger("app.http")
	if !ok {
		t.Fatal("app.http must match app.*")
	}
	logger.Info("hello")

	if httpBuffer.Len() == 0 {
		t.Fatal("app.* must receive the record")
	}
	if workerBuffer.Len() != 0 {
		t.Fatal("worker must not receive the record")
	}
}

func TestPatternLogBroadcastsToEveryMatch(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	patternLog := NewPatternLog(
		WithPatternHandler("app.*", slog.NewTextHandler(first, nil)),
		WithPatternHandler("*.http", slog.NewTextHandler(second, nil)),
	)

	logger, ok := patternLog.Logger("app.http")
	if !ok {
		t.Fatal("app.http must match both patterns")
	}
	logger.Info("hello")

	if first.Len() == 0 || second.Len() == 0 {
		t.Fatalf("first = %d, second = %d", first.Len(), second.Len())
	}
}

func TestPatternLogIgnoresInvalidPattern(t *testing.T) {
	patternLog := NewPatternLog(WithPatternHandler("[", slog.NewTextHandler(io.Discard, nil)))

	if _, ok := patternLog.Logger("app"); ok {
		t.Fatal("an invalid pattern must not match")
	}
}

func TestPatternLogMatchesQueryWildcard(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	patternLog := NewPatternLog(
		WithPatternHandler("network.debug", slog.NewTextHandler(first, nil)),
		WithPatternHandler("network.warn", slog.NewTextHandler(second, nil)),
	)

	logger, ok := patternLog.Logger("network.*")
	if !ok {
		t.Fatal("a wildcard query must match every concrete registered name")
	}
	logger.Info("hello")

	if first.Len() == 0 || second.Len() == 0 {
		t.Fatalf("first = %d, second = %d", first.Len(), second.Len())
	}
}

func TestPatternLogIgnoresInvalidRegistration(t *testing.T) {
	patternLog := NewPatternLog()

	patternLog.Register("", slog.NewTextHandler(io.Discard, nil))
	patternLog.Register("app.*", nil)

	if len(patternLog.handlers) != 0 {
		t.Fatalf("handlers = %v", patternLog.handlers)
	}
}

func TestPatternLogLifecycle(t *testing.T) {
	patternLog := NewPatternLog(WithPatternHandler("app.*", slog.NewTextHandler(io.Discard, nil)))

	if patternLog.Group() != domainLog.GroupLog || patternLog.Name() != domainLog.NamePattern {
		t.Fatalf("identity = %s.%s", patternLog.Group(), patternLog.Name())
	}
	if patternLog.State() != domainEngine.EnginePending {
		t.Fatalf("state = %v", patternLog.State())
	}
	if err := patternLog.Init(domainEngine.EngineContext{}); err != nil {
		t.Fatal(err)
	}
	if patternLog.State() != domainEngine.EngineReady {
		t.Fatalf("state = %v", patternLog.State())
	}
	if err := patternLog.Start(); err != nil {
		t.Fatal(err)
	}
	if patternLog.State() != domainEngine.EngineRunning {
		t.Fatalf("state = %v", patternLog.State())
	}
	if err := patternLog.Stop(); err != nil {
		t.Fatal(err)
	}
	if patternLog.State() != domainEngine.EngineStopped {
		t.Fatalf("state = %v", patternLog.State())
	}
	if _, ok := patternLog.Logger("app.http"); !ok {
		t.Fatal("handlers must survive a stop")
	}
}
