package engineSetting

import (
	"strings"

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

// Config maps runtime.engines.setting completely: the schema the engine validates against and the ordered layers.
type Config struct {
	Schema string `json:"schema"`
	// Layers is what the document declares, not what the engine ends up holding: the runtime layer is appended during
	// Init and never appears here.
	Layers []LayerConfig `json:"layers"`
}

// LayerConfig maps one entry of runtime.engines.setting.layers.
type LayerConfig struct {
	Layer string `json:"layer"`
	Path  string `json:"path"`
	Mode  string `json:"mode"`
}

// GenLayers generates the layer list from the configuration: the declared entries in ascending priority, with the
// mode translated and the unnamed ones dropped.
//
// Generating rather than reading, because the result is not what the document holds. Layers is the persisted list;
// the runtime layer that carries unsaved edits has no path and no configured mode, so the engine appends it during
// Init and it never appears here. The name keeps the difference visible at the call site: a caller asking for the
// layers the engine will actually use has to add the runtime one itself.
func (c Config) GenLayers() []Layer {
	layers := make([]Layer, 0, len(c.Layers))
	for _, entry := range c.Layers {
		if entry.Layer == "" {
			continue
		}
		layers = append(layers, Layer{
			Name: entry.Layer,
			Mode: ModeOf(entry.Mode),
			Path: entry.Path,
		})
	}
	return layers
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
