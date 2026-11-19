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
)
