package engineSettingImpl

import (
	engineSetting "github.com/wensboy/lightpc/backend/domain/engine/setting"
)

// LayersFrom turns the configured entries into the layers the engine keeps, dropping the ones without a name.
//
// The document is read by the contract package, so this file only decides what a layer looks like once it is in
// hand: a name, a path and a mode. The runtime layer that holds unsaved edits is not part of the configuration; the
// engine adds it during Init.
func LayersFrom(entries []engineSetting.LayerConfig) []engineSetting.Layer {
	layers := make([]engineSetting.Layer, 0, len(entries))
	for _, entry := range entries {
		if entry.Layer == "" {
			continue
		}
		layers = append(layers, engineSetting.Layer{
			Name: entry.Layer,
			Mode: engineSetting.ModeOf(entry.Mode),
			Path: entry.Path,
		})
	}
	return layers
}
