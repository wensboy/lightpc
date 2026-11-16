package storeimpl

import (
	"context"
	"embed"
	"errors"
	"os"
	"path/filepath"
	"testing"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	"github.com/wensboy/lightpc/backend/domain/engine/store"
	"github.com/wensboy/lightpc/backend/internal/config"
	"github.com/wensboy/lightpc/backend/internal/ebus"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

func newEngineContext() domainEngine.EngineContext {
	return domainEngine.EngineContext{AppCtx: &domainContext.AppContext{
		Context:      context.Background(),
		ErrorContext: domainContext.ErrorContext{interError.ErrCodeNotExist: "Not Exist"},
	}}
}

func TestFSInitRegistersStoreErrors(t *testing.T) {
	fsStore := NewFS(WithRoot(t.TempDir()))

	if err := fsStore.Init(newEngineContext()); err != nil {
		t.Fatal(err)
	}

	notFound := interError.Retrieve(store.ErrKeyNotFound)
	if notFound.Code != interError.ErrCodeNotExist || notFound.Kind != "Not Exist" {
		t.Fatalf("not found = %+v", notFound)
	}
	if state := fsStore.State(); state != domainEngine.EngineReady {
		t.Fatalf("state = %v", state)
	}
}

func TestFSMissingFileReusesRegisteredError(t *testing.T) {
	fsStore := NewFS(WithRoot(t.TempDir()))
	if err := fsStore.Init(newEngineContext()); err != nil {
		t.Fatal(err)
	}

	if _, err := fsStore.Get(context.Background(), "missing.txt"); !errors.Is(err, interError.Retrieve(store.ErrKeyNotFound)) {
		t.Fatalf("err = %v", err)
	}
}

func TestFSInitWithoutRootReportsInvalid(t *testing.T) {
	fsStore := NewFS()

	err := fsStore.Init(newEngineContext())
	if !errors.Is(err, interError.Retrieve(store.ErrKeyInvalid)) {
		t.Fatalf("err = %v", err)
	}
	if state := fsStore.State(); state != domainEngine.EngineError {
		t.Fatalf("state = %v", state)
	}
}

func TestSQLInitWithoutPoolReportsInvalid(t *testing.T) {
	sqlStore := NewSQL()

	err := sqlStore.Init(newEngineContext())
	if !errors.Is(err, interError.Retrieve(store.ErrKeyInvalid)) {
		t.Fatalf("err = %v", err)
	}
	if state := sqlStore.State(); state != domainEngine.EngineError {
		t.Fatalf("state = %v", state)
	}
}

func newConfiguredEngineContext(t *testing.T) domainEngine.EngineContext {
	t.Helper()
	fileSource := config.NewFileSource()
	fileSource.Parse(domainContext.FsContext{Paths: []string{"../../../data/conf/default.json"}})
	return domainEngine.EngineContext{AppCtx: &domainContext.AppContext{
		Context:       context.Background(),
		ConfigContext: fileSource,
		ErrorContext:  domainContext.ErrorContext{interError.ErrCodeNotExist: "Not Exist"},
	}}
}

func TestFSInitAppliesDefaultsFromConfig(t *testing.T) {
	fsStore := NewFS(WithRoot(t.TempDir()))

	if err := fsStore.Init(newConfiguredEngineContext(t)); err != nil {
		t.Fatal(err)
	}
	if fsStore.fileMode != 0o644 || fsStore.dirMode != 0o755 {
		t.Fatalf("modes = %v/%v", fsStore.fileMode, fsStore.dirMode)
	}
}

func TestFSInitKeepsConfiguredOverrides(t *testing.T) {
	fsStore := NewFS(WithRoot(t.TempDir())).WithFileMode(0o600).WithDirMode(0o700)

	if err := fsStore.Init(newConfiguredEngineContext(t)); err != nil {
		t.Fatal(err)
	}
	if fsStore.fileMode != 0o600 || fsStore.dirMode != 0o700 {
		t.Fatalf("modes = %v/%v", fsStore.fileMode, fsStore.dirMode)
	}
}

var testEmbed embed.FS

func newEmbeddedEngineContext(t *testing.T) domainEngine.EngineContext {
	t.Helper()
	return domainEngine.EngineContext{AppCtx: &domainContext.AppContext{
		Context:      context.Background(),
		EmbedContext: domainContext.EmbedContext{"/testdata": &testEmbed},
	}}
}

func TestFSReportsMissingEmbeddedContent(t *testing.T) {
	fsStore := NewFS(WithRoot(t.TempDir()))
	if err := fsStore.Init(newEmbeddedEngineContext(t)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	missing := interError.Retrieve(store.ErrKeyNotFound)

	if _, err := fsStore.Get(ctx, "embed://testdata/hello.txt"); !errors.Is(err, missing) {
		t.Fatalf("get err = %v", err)
	}
	if _, err := fsStore.Stat(ctx, "embed://testdata"); !errors.Is(err, missing) {
		t.Fatalf("stat err = %v", err)
	}
}

func TestFSRejectsEmbeddedWrites(t *testing.T) {
	fsStore := NewFS(WithRoot(t.TempDir()))
	if err := fsStore.Init(newEmbeddedEngineContext(t)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	readOnly := interError.Retrieve(store.ErrKeyReadOnly)

	if err := fsStore.Put(ctx, "embed://testdata/hello.txt", []byte("x")); !errors.Is(err, readOnly) {
		t.Fatalf("put err = %v", err)
	}
	if err := fsStore.Delete(ctx, "embed://testdata/hello.txt"); !errors.Is(err, readOnly) {
		t.Fatalf("delete err = %v", err)
	}
	if err := fsStore.MkdirAll(ctx, "embed://testdata/sub"); !errors.Is(err, readOnly) {
		t.Fatalf("mkdir err = %v", err)
	}
}

func TestFSInitCreatesRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "nested", "store")
	fsStore := NewFS(WithRoot(root))

	if err := fsStore.Init(newConfiguredEngineContext(t)); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		t.Fatalf("root stat = %v, %v", info, err)
	}
}

func TestFSInitWatchesStopSignal(t *testing.T) {
	bus := ebus.NewEventBus()
	engineCtx := domainEngine.EngineContext{
		AppCtx: &domainContext.AppContext{Context: context.Background(), ErrorContext: domainContext.ErrorContext{}},
		Ebus:   bus,
	}
	fsStore := NewFS(WithRoot(t.TempDir()))
	if err := fsStore.Init(engineCtx); err != nil {
		t.Fatal(err)
	}
	if err := fsStore.Start(); err != nil {
		t.Fatal(err)
	}

	bus.Emit(domainBus.EventEngineStop, map[string]any{"group": store.GroupStore, "name": store.NameFs})

	if state := fsStore.State(); state != domainEngine.EngineStopped {
		t.Fatalf("state = %v", state)
	}
}

func TestFSRestartRetainsContext(t *testing.T) {
	fsStore := NewFS(WithRoot(t.TempDir()))
	if err := fsStore.Init(newEngineContext()); err != nil {
		t.Fatal(err)
	}
	if err := fsStore.Stop(); err != nil {
		t.Fatal(err)
	}

	if err := fsStore.Restart(); err != nil {
		t.Fatal(err)
	}
	if state := fsStore.State(); state != domainEngine.EngineReady {
		t.Fatalf("state = %v", state)
	}
}
