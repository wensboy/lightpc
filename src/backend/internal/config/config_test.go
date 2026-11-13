package config

import (
	"os"
	"path/filepath"
	"testing"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
)

func TestFileSourceParseAndLookup(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "default.json"), []byte(`{"server":{"port":8080,"host":"127.0.0.1"},"name":"lightpc"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "broken.json"), []byte(`{`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "skip.yaml"), []byte("name: lightpc"), 0o600); err != nil {
		t.Fatal(err)
	}

	source := NewFileSource()
	source.Parse(domainContext.FsContext{Paths: []string{
		filepath.Join(directory, "default.json"),
		filepath.Join(directory, "broken.json"),
		filepath.Join(directory, "skip.yaml"),
	}})

	if value, ok := source.Lookup("default.server.port"); !ok || value != float64(8080) {
		t.Fatalf("default.server.port = %v, %v", value, ok)
	}
	if value, ok := source.Lookup("default.name"); !ok || value != "lightpc" {
		t.Fatalf("default.name = %v, %v", value, ok)
	}
	if value, ok := source.Lookup("default.server.host"); !ok || value != "127.0.0.1" {
		t.Fatalf("default.server.host = %v, %v", value, ok)
	}
	if _, ok := source.Lookup("broken.name"); ok {
		t.Fatal("broken.json should be ignored")
	}
	if _, ok := source.Lookup("skip.name"); ok {
		t.Fatal("skip.yaml should be ignored")
	}
}

func TestFileSourceFlat(t *testing.T) {
	source := &FileSource{raw: map[string]any{
		"default": map[string]any{
			"server": map[string]any{"port": 8080},
			"tags":   []any{"a", "b"},
		},
	}}
	source.Flat()

	if value, ok := source.flat["default.server.port"]; !ok || value != 8080 {
		t.Fatalf("default.server.port = %v, %v", value, ok)
	}
	if _, ok := source.flat["default.server"]; ok {
		t.Fatal("an object node must not be flattened into the result")
	}
	if _, ok := source.flat["default.tags"]; !ok {
		t.Fatal("an array must be treated as a leaf")
	}
}

func TestFileSourceLookupMissing(t *testing.T) {
	source := &FileSource{raw: map[string]any{"default": map[string]any{"name": "lightpc"}}}
	if _, ok := source.Lookup("default.server.port"); ok {
		t.Fatal("a missing path must not be found")
	}
	if _, ok := source.Lookup("default.name.value"); ok {
		t.Fatal("a leaf must not be traversed any deeper")
	}
}

func TestEnvSourceLookup(t *testing.T) {
	t.Setenv("LIGHTPC_TEST_ENV_AUTHOR", "env")
	source := NewEnvSource().
		WithPrefix("LIGHTPC_TEST_ENV_").
		WithExtra(map[string]any{"LIGHTPC_TEST_ENV_REPO": "extra"})

	if value, ok := source.Lookup("author"); !ok || value != "env" {
		t.Fatalf("author = %v, %v", value, ok)
	}
	if value, ok := source.Lookup("repo"); !ok || value != "extra" {
		t.Fatalf("repo = %v, %v", value, ok)
	}
	if _, ok := source.Lookup("missing"); ok {
		t.Fatal("a missing key must not be found")
	}
}

func TestEnvSourceExport(t *testing.T) {
	directory := t.TempDir()
	envPath := filepath.Join(directory, ".env")
	content := "# comment\n\nLIGHTPC_TEST_EXPORT_AUTHOR=\"wendisx\"\nLIGHTPC_TEST_EXPORT_LICENSE=MIT\ninvalid line\n"
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	source := NewEnvSource().WithPrefix("LIGHTPC_TEST_EXPORT_")
	source.Export(domainContext.FsContext{Paths: []string{envPath}})

	if value, ok := source.Lookup("author"); !ok || value != "wendisx" {
		t.Fatalf("author = %v, %v", value, ok)
	}
	if value, ok := source.Lookup("license"); !ok || value != "MIT" {
		t.Fatalf("license = %v, %v", value, ok)
	}
}

func TestEnvSourceExportMissingFile(t *testing.T) {
	source := NewEnvSource()
	source.Export(domainContext.FsContext{Paths: []string{filepath.Join(t.TempDir(), "absent.env")}})
	if len(source.extra) != 0 {
		t.Fatalf("extra = %v, want empty", source.extra)
	}
}

func TestConfigLookup(t *testing.T) {
	fileSource := &FileSource{raw: map[string]any{"default": map[string]any{"name": "file"}}}
	config := NewConfig().
		WithSources(fileSource).
		WithRegistry(map[string]any{"default.name": "dynamic", "runtime.only": 1})

	if value, ok := config.Lookup("default.name"); !ok || value != "file" {
		t.Fatalf("default.name = %v, %v", value, ok)
	}
	if value, ok := config.Lookup("runtime.only"); !ok || value != 1 {
		t.Fatalf("runtime.only = %v, %v", value, ok)
	}
	if _, ok := config.Lookup("runtime.missing"); ok {
		t.Fatal("a missing key must not be found")
	}
}
