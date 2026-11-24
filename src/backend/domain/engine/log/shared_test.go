package engineLog

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

func TestEntriesForReadsConfiguredGroup(t *testing.T) {
	cfg := domainContext.ConfigContext(configSource{values: map[string]any{
		EntriesKey: map[string]any{
			SubDirect: []any{
				map[string]any{"name": "dev.console", "level": "debug", "handler": "console"},
				map[string]any{"level": "info"},
			},
		},
	}})

	entries := EntriesFor(cfg, SubDirect)
	// The unnamed entry cannot be addressed, so it is dropped rather than producing an engine.
	if len(entries) != 1 {
		t.Fatalf("entries = %d", len(entries))
	}
	if entries[0].Name != "dev.console" {
		t.Fatalf("name = %q", entries[0].Name)
	}
}

func TestEntriesForMissingGroup(t *testing.T) {
	cfg := domainContext.ConfigContext(configSource{values: map[string]any{}})

	if entries := EntriesFor(cfg, SubDirect); len(entries) != 0 {
		t.Fatalf("entries = %d", len(entries))
	}
}

func TestEntriesForEmptyGroup(t *testing.T) {
	cfg := domainContext.ConfigContext(configSource{values: map[string]any{
		EntriesKey: map[string]any{SubDirect: []any{}},
	}})

	if entries := EntriesFor(cfg, SubDirect); len(entries) != 0 {
		t.Fatalf("entries = %d", len(entries))
	}
}

func TestEntriesForKeepsTheOutputRaw(t *testing.T) {
	cfg := domainContext.ConfigContext(configSource{values: map[string]any{
		EntriesKey: map[string]any{
			SubFanout: []any{
				map[string]any{
					"name":    "dev.file",
					"handler": "file",
					"output":  []any{map[string]any{"path": "local://~/logs/a.log"}},
				},
			},
		},
	}})

	entries := EntriesFor(cfg, SubFanout)
	if len(entries) != 1 {
		t.Fatalf("entries = %d", len(entries))
	}
	// The shape depends on the group, so the contract only promises the bytes are preserved.
	if len(entries[0].Output) == 0 {
		t.Fatal("output must be kept verbatim")
	}
}
