package logimpl

import (
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

	engines := Setup(domainContext.ConfigContext(fileSource))
	if len(engines) != 3 {
		t.Fatalf("engines = %d", len(engines))
	}

	directClient, ok := cast.As[domainLog.DirectClient](engines[0])
	if !ok {
		t.Fatal("the first engine must expose DirectClient")
	}
	for _, name := range []string{"dev.console", "dev.file"} {
		if _, ok := directClient.Logger(name); !ok {
			t.Fatalf("direct logger %q must be resolved", name)
		}
	}

	fanoutClient, ok := cast.As[domainLog.FanoutClient](engines[1])
	if !ok {
		t.Fatal("the second engine must expose FanoutClient")
	}
	if _, ok := fanoutClient.Logger("dev.test"); !ok {
		t.Fatal("fanout logger dev.test must be resolved")
	}

	patternClient, ok := cast.As[domainLog.PatternClient](engines[2])
	if !ok {
		t.Fatal("the third engine must expose PatternClient")
	}
	for _, name := range []string{"network.debug", "network.info"} {
		if _, ok := patternClient.Logger(name); !ok {
			t.Fatalf("pattern logger %q must be resolved", name)
		}
	}
}
