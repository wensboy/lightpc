package bus

const (
	// engine event
	EventEngineStop  = "engine::stop"
	EventEngineStart = "engine::start"
	EventEngineReady = "engine::ready"
	EventEngineError = "engine::error"
	EventEnginePend  = "engine::pend"
	EventEngineInit  = "engine::init"

	// setting event
	//
	// Two events rather than one with an optional list: a single edit and a layer swap are different things to a
	// listener. A layer load replaces many keys at once, and sending it through the single key event would make every
	// listener run its per-key path once per key.
	EventSettingChanged  = "setting::changed"  // one key, from the service bus
	EventSettingReloaded = "setting::reloaded" // one layer, from the engine bus

	// session event
	EventSessionExpired = "session::expired"

	// notice event
	EventNoticeSend = "notice::send"
)

// UiEvents lists the bus events the interface re-renders from. Keeping the names in one place means a new
// event is forwarded by adding a name, not by wiring another subscription in the layout.
var UiEvents = []string{EventSettingChanged, EventSettingReloaded, EventSessionExpired, EventNoticeSend}

var EngineOpaqueEvents = []string{EventSettingReloaded, EventNoticeSend}

var ServiceOpaqueEvents = []string{EventSettingChanged, EventSessionExpired, EventNoticeSend}
