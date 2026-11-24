package engineSetting

import (
	"encoding/json"
	"strings"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

const (
	// ModeReadOnly allows resolving values from the layer.
	ModeReadOnly Mode = iota + 1
	// ModeWriteOnly allows writing values into the layer, such as a user overlay.
	ModeWriteOnly
	// ModeReadWrite allows both.
	ModeReadWrite
)

// Error keys published when the settings engine initialises.
const (
	ErrKeyLayerUnknown  = "error::engine.setting::unknown_layer"
	ErrKeyLayerReadOnly = "error::engine.setting::read_only_layer"
)

// Errors lists every setting error published during Init.
var Errors = []interError.Option{
	{Key: ErrKeyLayerUnknown, Code: interError.ErrCodeNotExist, Msg: "setting: unknown layer"},
	{Key: ErrKeyLayerReadOnly, Code: interError.ErrCodeOpDenied, Msg: "setting: layer is read only"},
}

// Mode controls how a layer may be used. The zero value is intentionally invalid.
type Mode uint8

// LayersKey is the configuration path that holds the layer list.
const LayersKey = "default.runtime.engines.setting.layers"

// LayerConfig maps one entry of runtime.engines.setting.layers.
type LayerConfig struct {
	Layer string `json:"layer"`
	Path  string `json:"path"`
	Mode  string `json:"mode"`
}

// LayersFor reads the configured layers in ascending priority: the first entry has the lowest priority and every
// following entry overrides the ones before it.
//
// The runtime layer is deliberately absent from the document. It is the top layer that holds unsaved edits, so the
// engine creates it during Init instead of asking the bootstrap to wire it.
func LayersFor(cfg domainContext.ConfigContext) []LayerConfig {
	value, found := cfg.Lookup(LayersKey)
	if !found {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	entries := make([]LayerConfig, 0)
	if err := json.Unmarshal(encoded, &entries); err != nil {
		return nil
	}

	kept := entries[:0]
	for _, entry := range entries {
		if entry.Layer == "" {
			continue
		}
		kept = append(kept, entry)
	}
	return kept
}

// ModeOf maps a configured mode onto Mode. An unknown value falls back to read only, so a typo never grants write
// access to a layer.
func ModeOf(text string) Mode {
	switch strings.ToLower(text) {
	case "rw":
		return ModeReadWrite
	case "wo":
		return ModeWriteOnly
	default:
		return ModeReadOnly
	}
}
