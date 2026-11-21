package engineSettingImpl_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineSetting "github.com/wensboy/lightpc/backend/domain/engine/setting"
	engineSettingImpl "github.com/wensboy/lightpc/backend/engine/setting"
	engineStoreImpl "github.com/wensboy/lightpc/backend/engine/store"
)

// newLayered builds a settings engine with a writable user layer at dir and an in-memory runtime layer.
func newLayered(t *testing.T, dir string) *engineSettingImpl.Setting {
	t.Helper()
	fsStore := engineStoreImpl.NewFS(engineStoreImpl.WithRoot(dir)).WithDirMode(0o755).WithFileMode(0o644)
	if err := fsStore.Init(domainEngine.EngineContext{}); err != nil {
		t.Fatalf("init store: %v", err)
	}
	t.Cleanup(func() { _ = fsStore.Close() })

	instance := engineSettingImpl.NewSetting(
		engineSettingImpl.WithLayer(engineSetting.Layer{Name: "user", Mode: engineSetting.ModeReadWrite, Path: "user.json"}),
		engineSettingImpl.WithLayer(engineSetting.Layer{Name: "runtime", Mode: engineSetting.ModeReadWrite}),
	)
	if err := instance.Init(domainEngine.EngineContext{
		Engines: map[string]domainEngine.Engine{domainEngine.StoreFs: fsStore},
	}); err != nil {
		t.Fatalf("init setting: %v", err)
	}
	if err := instance.Prepare(); err != nil {
		t.Fatalf("prepare setting: %v", err)
	}
	if err := instance.Prepare(); err != nil {
		t.Fatalf("prepare setting: %v", err)
	}
	t.Cleanup(func() { _ = instance.Stop() })
	return instance
}

func readDocument(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	values := map[string]any{}
	if err := json.Unmarshal(data, &values); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return values
}

func TestFlushSinksRuntimeIntoUserLayer(t *testing.T) {
	dir := t.TempDir()
	instance := newLayered(t, dir)

	if err := instance.Set("runtime", "appearence.mode", "dark"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if raw, found := instance.Resolve("appearence.mode"); !found {
		t.Fatal("runtime value must resolve before the flush")
	} else if entry, ok := raw.(engineSetting.Entry[any]); !ok || entry.Value != "dark" {
		t.Fatalf("entry = %+v", raw)
	}

	if err := instance.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}

	stored := readDocument(t, filepath.Join(dir, "user.json"))
	appearence, ok := stored["appearence"].(map[string]any)
	if !ok {
		t.Fatalf("dotted keys must be nested on disk, got %v", stored)
	}
	if appearence["mode"] != "dark" {
		t.Fatalf("user layer = %v", stored)
	}
}

func TestFlushKeepsSiblingKeys(t *testing.T) {
	dir := t.TempDir()
	instance := newLayered(t, dir)

	if err := instance.Set("runtime", "one", 1); err != nil {
		t.Fatalf("set one: %v", err)
	}
	if err := instance.Flush(); err != nil {
		t.Fatalf("first flush: %v", err)
	}
	if err := instance.Set("runtime", "two", 2); err != nil {
		t.Fatalf("set two: %v", err)
	}
	if err := instance.Flush(); err != nil {
		t.Fatalf("second flush: %v", err)
	}

	stored := readDocument(t, filepath.Join(dir, "user.json"))
	if stored["one"] != float64(1) || stored["two"] != float64(2) {
		t.Fatalf("user layer = %v", stored)
	}
}

func TestFlushWithoutEditsLeavesDocumentEmpty(t *testing.T) {
	dir := t.TempDir()
	instance := newLayered(t, dir)

	// Init creates the layer file so a fresh install always has one; a clean flush must not add to it.
	if err := instance.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	stored := readDocument(t, filepath.Join(dir, "user.json"))
	if len(stored) != 0 {
		t.Fatalf("a clean flush must not touch the document, got %v", stored)
	}
}
