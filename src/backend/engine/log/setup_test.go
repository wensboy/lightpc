package logimpl

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	"github.com/wensboy/lightpc/backend/internal/cast"
	"github.com/wensboy/lightpc/backend/internal/config"
)

type setupSource struct {
	values map[string]any
}

func (s setupSource) Kind() uint8 { return domainConfig.SourceConfig }

func (s setupSource) Lookup(key string) (any, bool) {
	value, ok := s.values[key]
	return value, ok
}

func TestSetupDirectAndFanout(t *testing.T) {
	directory := t.TempDir()
	debugPath := filepath.Join(directory, "logs", "debug.log")

	cfg := domainContext.ConfigContext(setupSource{values: map[string]any{
		"default.runtime.log": map[string]any{
			"direct": []any{
				map[string]any{
					"name":    "dev.console",
					"level":   "debug",
					"handler": "",
					"output": map[string]any{
						"serializer": map[string]any{"type": "text", "enableSource": true},
						"out":        "stdout",
					},
				},
				map[string]any{
					"name":    "dev.file",
					"level":   "info",
					"handler": "file",
					"output": map[string]any{
						"serializer": map[string]any{"type": "json", "enableSource": false},
						"path":       debugPath,
						"mode":       "0o644",
					},
				},
			},
			"fanout": []any{
				map[string]any{
					"name":    "dev.test",
					"level":   "info",
					"handler": "file",
					"output": []any{
						map[string]any{
							"serializer": map[string]any{"type": "json"},
							"path":       filepath.Join(directory, "logs", "process.log"),
							"mode":       "0o644",
						},
						map[string]any{
							"serializer": map[string]any{"type": "json"},
							"path":       filepath.Join(directory, "logs", "error.log"),
							"mode":       "0o644",
						},
					},
				},
			},
		},
	}})

	engines := Setup(cfg)

	// The pattern group is absent, so only direct and fanout are created.
	if len(engines) != 2 {
		t.Fatalf("engines = %d", len(engines))
	}

	directClient, ok := cast.As[domainLog.DirectClient](engines[0])
	if !ok {
		t.Fatal("the first engine must expose DirectClient")
	}
	for _, name := range []string{"dev.console", "dev.file"} {
		if _, ok := directClient.Logger(name); !ok {
			t.Fatalf("engine %q must be registered in a single instance", name)
		}
	}

	fanoutClient, ok := cast.As[domainLog.FanoutClient](engines[1])
	if !ok {
		t.Fatal("the second engine must expose FanoutClient")
	}
	if _, ok := fanoutClient.Logger("dev.test"); !ok {
		t.Fatal("dev.test must be registered in the fanout instance")
	}

	if _, err := os.Stat(debugPath); err != nil {
		t.Fatalf("the log file must be created: %v", err)
	}

	if _, ok := slog.Default().Handler().(*slog.MultiHandler); !ok {
		t.Fatalf("the slog default must fan out to every handler, got %T", slog.Default().Handler())
	}
}

func TestSetupWithoutConfig(t *testing.T) {
	cfg := domainContext.ConfigContext(setupSource{values: map[string]any{}})

	if engines := Setup(cfg); len(engines) != 0 {
		t.Fatalf("engines = %d", len(engines))
	}
}

func TestSetupWithEmptyGroups(t *testing.T) {
	cfg := domainContext.ConfigContext(setupSource{values: map[string]any{
		"default.runtime.log": map[string]any{
			"direct":  []any{},
			"fanout":  []any{},
			"pattern": []any{},
		},
	}})

	if engines := Setup(cfg); len(engines) != 0 {
		t.Fatalf("engines = %d", len(engines))
	}
}

func TestSetupFromConfigFile(t *testing.T) {
	fileSource := config.NewFileSource()
	fileSource.Parse(domainContext.FsContext{Paths: []string{"../../../data/conf/default.json"}})

	// The shipped configuration only declares the fanout group.
	engines := Setup(domainContext.ConfigContext(fileSource))
	if len(engines) != 1 {
		t.Fatalf("engines = %d", len(engines))
	}

	fanoutClient, ok := cast.As[domainLog.FanoutClient](engines[0])
	if !ok {
		t.Fatal("the engine must expose FanoutClient")
	}
	for _, name := range []string{"dev.file", "dev.console"} {
		if _, ok := fanoutClient.Logger(name); !ok {
			t.Fatalf("fanout logger %q must be resolved", name)
		}
	}
	if _, ok := fanoutClient.Logger(""); !ok {
		t.Fatal("the slog default must resolve to every fanout handler")
	}
}
