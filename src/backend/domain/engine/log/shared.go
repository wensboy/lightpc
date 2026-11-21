package engineLog

import (
	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
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

// LogHookContext is what a log hook may inspect between the engine phases. All three log engines share the
// same shape: they differ in how a record is routed, not in what they need to prepare.
type LogHookContext struct {
	// Items lists the configured entries of the group, in configuration order.
	Items []domainConfig.LogItem
	// EngineContext is the context the engine was initialised with.
	EngineContext domainEngine.EngineContext
}

// DirectHookContext is the direct log engine's hook context.
type DirectHookContext = LogHookContext

// FanoutHookContext is the fanout log engine's hook context.
type FanoutHookContext = LogHookContext

// PatternHookContext is the pattern log engine's hook context.
type PatternHookContext = LogHookContext
