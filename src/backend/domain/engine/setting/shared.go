package setting

import interError "github.com/wensboy/lightpc/backend/internal/error"

const (
	// GroupSetting is the engine group of the settings engine.
	GroupSetting = "setting"

	// NameSetting is the engine name of the settings engine.
	NameSetting = "setting"
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
