package engineLog

import (
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// Error keys published by the log engines during Init.
const (
	// ErrKeyInvalidItem reports a configured item without a name or without any usable output.
	ErrKeyInvalidItem = "error::engine.log::invalid_item"
	// ErrKeyHandlerFailed reports an output whose handler could not be created.
	ErrKeyHandlerFailed = "error::engine.log::handler_failed"
)

// Errors lists every log error published when a log engine initialises.
// Kind is resolved from the application error context at registration time.
var Errors = []interError.Option{
	{Key: ErrKeyInvalidItem, Code: interError.ErrCodeInvalidParam, Msg: "log: invalid item"},
	{Key: ErrKeyHandlerFailed, Code: interError.ErrCodeOpFailed, Msg: "log: handler creation failed"},
}

type LogClient struct {
	Direct  DirectLog
	Fanout  FanoutLog
	Pattern PatternLog
}
