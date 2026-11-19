package log

import (
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

const (
	GroupLog = "log"

	NameDirect  = "direct"
	NameFanout  = "fanout"
	NamePattern = "pattern"
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

var (
	LogDirect  = domainEngine.EngineDep(GroupLog, NameDirect)
	LogFanout  = domainEngine.EngineDep(GroupLog, NameFanout)
	LogPattern = domainEngine.EngineDep(GroupLog, NamePattern)
)

type LogClient struct {
	Direct  DirectLog
	Fanout  FanoutLog
	Pattern PatternLog
}
