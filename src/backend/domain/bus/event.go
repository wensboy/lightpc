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
	EventSettingChanged = "setting::changed"

	// session event
	EventSessionExpired = "session::expired"
)

// UiEvents lists the bus events the interface re-renders from. Keeping the names in one place means a new
// event is forwarded by adding a name, not by wiring another subscription in the layout.
var UiEvents = []string{EventSettingChanged, EventSessionExpired}
