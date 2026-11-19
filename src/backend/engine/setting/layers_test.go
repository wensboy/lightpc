package settingimpl_test

import (
	"strings"
	"testing"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	"github.com/wensboy/lightpc/backend/domain/engine/setting"
	settingimpl "github.com/wensboy/lightpc/backend/engine/setting"
	"github.com/wensboy/lightpc/backend/internal/config"
)

func TestLayersFromShippedConfig(t *testing.T) {
	fileSource := config.NewFileSource()
	fileSource.Parse(domainContext.FsContext{Paths: []string{"../../../data/conf/default.json"}})

	layers := settingimpl.LayersFrom(domainContext.ConfigContext(fileSource))
	if len(layers) != 2 {
		t.Fatalf("layers = %d", len(layers))
	}

	// Ascending priority: the embedded default comes first and the user document overrides it.
	if layers[0].Name != "default" || layers[0].Mode != setting.ModeReadOnly {
		t.Fatalf("first layer = %+v", layers[0])
	}
	if !strings.HasPrefix(layers[0].Path, "embed://") {
		t.Fatalf("first layer path = %q", layers[0].Path)
	}
	if layers[1].Name != "user" || layers[1].Mode != setting.ModeReadWrite {
		t.Fatalf("second layer = %+v", layers[1])
	}
	if !strings.HasPrefix(layers[1].Path, "local://") {
		t.Fatalf("second layer path = %q", layers[1].Path)
	}

	// The runtime layer is created by the engine, never configured.
	for _, layer := range layers {
		if layer.Name == "runtime" {
			t.Fatal("the runtime layer must not be configured")
		}
	}
}

func TestLayersFromMissingConfig(t *testing.T) {
	fileSource := config.NewFileSource()

	if layers := settingimpl.LayersFrom(domainContext.ConfigContext(fileSource)); len(layers) != 0 {
		t.Fatalf("layers = %d", len(layers))
	}
}
