package engineSettingImpl_test

import (
	"os"
	"path/filepath"
	"testing"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineSetting "github.com/wensboy/lightpc/backend/domain/engine/setting"
	engineSettingImpl "github.com/wensboy/lightpc/backend/engine/setting"
	engineStoreImpl "github.com/wensboy/lightpc/backend/engine/store"
)

// newStore builds a ready file store rooted at root.
func newStore(t *testing.T, root string) *engineStoreImpl.FS {
	t.Helper()
	fsStore := engineStoreImpl.NewFS(engineStoreImpl.WithRoot(root)).WithFileMode(0o644).WithDirMode(0o755)
	if err := fsStore.Init(domainEngine.EngineContext{}); err != nil {
		t.Fatalf("init store: %v", err)
	}
	t.Cleanup(func() { _ = fsStore.Close() })
	return fsStore
}

func TestInitCreatesMissingLocalLayer(t *testing.T) {
	root := t.TempDir()
	fsStore := newStore(t, root)
	layerPath := filepath.Join(root, "settings", "engineSetting.json")

	instance := engineSettingImpl.NewSetting(engineSettingImpl.WithLayer(engineSetting.Layer{
		Name: "user",
		Mode: engineSetting.ModeReadWrite,
		Path: "local://" + layerPath,
	}))
	if err := instance.Init(domainEngine.EngineContext{
		Engines: map[string]domainEngine.Engine{domainEngine.StoreFs: fsStore},
	}); err != nil {
		t.Fatalf("init setting: %v", err)
	}
	t.Cleanup(func() { _ = instance.Stop() })

	data, err := os.ReadFile(layerPath)
	if err != nil {
		t.Fatalf("layer file must be created: %v", err)
	}
	if string(data) != "{}\n" {
		t.Fatalf("content = %q", data)
	}
}

func TestInitKeepsExistingLocalLayer(t *testing.T) {
	root := t.TempDir()
	fsStore := newStore(t, root)
	layerPath := filepath.Join(root, "engineSetting.json")
	if err := os.WriteFile(layerPath, []byte(`{"general":{"language":"Chinese"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	instance := engineSettingImpl.NewSetting(engineSettingImpl.WithLayer(engineSetting.Layer{
		Name: "user",
		Mode: engineSetting.ModeReadWrite,
		Path: "local://" + layerPath,
	}))
	if err := instance.Init(domainEngine.EngineContext{
		Engines: map[string]domainEngine.Engine{domainEngine.StoreFs: fsStore},
	}); err != nil {
		t.Fatalf("init setting: %v", err)
	}
	t.Cleanup(func() { _ = instance.Stop() })

	// Leaves are addressed by their dotted path, so the nested document resolves one level down.
	raw, found := instance.Resolve("general.language")
	if !found {
		t.Fatal("the stored document must be loaded")
	}
	entry, ok := raw.(engineSetting.Entry[any])
	if !ok {
		t.Fatalf("entry = %T", raw)
	}
	if entry.Layer != "user" {
		t.Fatalf("layer = %q", entry.Layer)
	}
	if entry.Value != "Chinese" {
		t.Fatalf("value = %v", entry.Value)
	}
}

func TestInitSkipsLayerWithoutStore(t *testing.T) {
	instance := engineSettingImpl.NewSetting(engineSettingImpl.WithLayer(engineSetting.Layer{
		Name: "default",
		Mode: engineSetting.ModeReadOnly,
		Path: "embed://data/conf/missing.json",
	}))

	// Without the file store the layer cannot be verified; initialisation must still succeed.
	if err := instance.Init(domainEngine.EngineContext{}); err != nil {
		t.Fatalf("init setting: %v", err)
	}
	t.Cleanup(func() { _ = instance.Stop() })
}
