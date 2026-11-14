package storeimpl

import (
	"context"
	"errors"
	"testing"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	"github.com/wensboy/lightpc/backend/domain/engine/store"
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
