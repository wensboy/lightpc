package engineSetting

import (
	"testing"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
)

// configSource is a configuration context backed by a literal map, so a test does not need a file.
type configSource struct {
	values map[string]any
}

func (s configSource) Kind() uint8 { return domainConfig.SourceConfig }

func (s configSource) Lookup(key string) (any, bool) {
	value, ok := s.values[key]
	return value, ok
}

func TestLayersForReadsConfiguredLayers(t *testing.T) {
	cfg := domainContext.ConfigContext(configSource{values: map[string]any{
		LayersKey: []any{
			map[string]any{"layer": "default", "path": "embed://data/conf/setting.default.json", "mode": "ro"},
			map[string]any{"layer": "user", "path": "local://~/.config/lightpc/setting.json", "mode": "rw"},
			map[string]any{"path": "local:///tmp/unnamed.json", "mode": "rw"},
		},
	}})

	entries := LayersFor(cfg)
	// An entry without a name cannot be addressed, so it is dropped rather than producing a layer.
	if len(entries) != 2 {
		t.Fatalf("entries = %d", len(entries))
	}
	if entries[0].Layer != "default" || entries[1].Layer != "user" {
		t.Fatalf("entries = %+v", entries)
	}
}

func TestLayersForMissingConfig(t *testing.T) {
	cfg := domainContext.ConfigContext(configSource{values: map[string]any{}})

	if entries := LayersFor(cfg); len(entries) != 0 {
		t.Fatalf("entries = %d", len(entries))
	}
}

func TestModeOfFallsBackToReadOnly(t *testing.T) {
	cases := map[string]Mode{
		"rw":    ModeReadWrite,
		"wo":    ModeWriteOnly,
		"ro":    ModeReadOnly,
		"":      ModeReadOnly,
		"write": ModeReadOnly,
	}
	for text, want := range cases {
		if got := ModeOf(text); got != want {
			t.Fatalf("ModeOf(%q) = %v, want %v", text, got, want)
		}
	}
}
