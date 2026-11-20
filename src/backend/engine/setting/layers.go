package engineSettingImpl

import (
	"strings"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	engineSetting "github.com/wensboy/lightpc/backend/domain/engine/setting"
	"github.com/wensboy/lightpc/backend/internal/cast"
)

// LayersKey is the configuration path that holds the layer list.
const LayersKey = "default.runtime.setting.layers"

// LayersFrom reads the configured layers in ascending priority: the first entry has the lowest priority
// and every following entry overrides the ones before it.
//
// The runtime layer is deliberately absent from the configuration. It is the top layer that holds unsaved
// edits, so the engine creates it during Init instead of asking the bootstrap to wire it.
func LayersFrom(cfg domainContext.ConfigContext) []engineSetting.Layer {
	value, ok := cfg.Lookup(LayersKey)
	if !ok {
		return nil
	}
	records, ok := cast.To[[]any](value)
	if !ok {
		return nil
	}

	layers := make([]engineSetting.Layer, 0, len(records))
	for _, record := range records {
		fields, ok := cast.To[map[string]any](record)
		if !ok {
			continue
		}
		name, _ := cast.To[string](fields["layer"])
		if name == "" {
			continue
		}
		path, _ := cast.To[string](fields["path"])
		mode, _ := cast.To[string](fields["mode"])
		layers = append(layers, engineSetting.Layer{Name: name, Path: path, Mode: parseMode(mode)})
	}
	return layers
}

// parseMode maps the configured mode onto engineSetting.Mode; an unknown value falls back to read only so a
// typo never grants write access.
func parseMode(text string) engineSetting.Mode {
	switch strings.ToLower(text) {
	case "rw":
		return engineSetting.ModeReadWrite
	case "wo":
		return engineSetting.ModeWriteOnly
	default:
		return engineSetting.ModeReadOnly
	}
}
