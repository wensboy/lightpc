package engineLogImpl

import (
	"testing"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	"github.com/wensboy/lightpc/backend/internal/config"
)

type sharedSource struct {
	values map[string]any
}

func (s sharedSource) Kind() uint8 { return domainConfig.SourceConfig }

func (s sharedSource) Lookup(key string) (any, bool) {
	value, ok := s.values[key]
	return value, ok
}

func TestItemsForReadsConfiguredGroup(t *testing.T) {
	cfg := domainContext.ConfigContext(sharedSource{values: map[string]any{
		"default.runtime.engines.log": map[string]any{
			"direct": []any{
				map[string]any{"name": "dev.console", "level": "debug", "handler": ""},
				map[string]any{"level": "info"},
			},
		},
	}})

	items := ItemsFor(cfg, "direct")
	// The unnamed entry carries no sink to identify, so it is skipped instead of producing an engine.
	if len(items) != 1 {
		t.Fatalf("items = %d", len(items))
	}
	if items[0].Name != "dev.console" {
		t.Fatalf("name = %q", items[0].Name)
	}
}

func TestItemsForMissingGroup(t *testing.T) {
	cfg := domainContext.ConfigContext(sharedSource{values: map[string]any{}})

	if items := ItemsFor(cfg, "direct"); len(items) != 0 {
		t.Fatalf("items = %d", len(items))
	}
}

func TestItemsForEmptyGroup(t *testing.T) {
	cfg := domainContext.ConfigContext(sharedSource{values: map[string]any{
		"default.runtime.engines.log": map[string]any{"direct": []any{}},
	}})

	if items := ItemsFor(cfg, "direct"); len(items) != 0 {
		t.Fatalf("items = %d", len(items))
	}
}

func TestItemsForShippedConfig(t *testing.T) {
	fileSource := config.NewFileSource()
	fileSource.Parse(domainContext.FsContext{Paths: []string{"../../../data/conf/default.json"}})
	cfg := domainContext.ConfigContext(fileSource)

	if items := ItemsFor(cfg, "fanout"); len(items) == 0 {
		t.Fatal("the shipped configuration declares fanout items")
	}
	if items := ItemsFor(cfg, "direct"); len(items) != 0 {
		t.Fatalf("the shipped configuration leaves direct empty, got %d", len(items))
	}
}
