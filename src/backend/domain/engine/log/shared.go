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
const ()

// Errors lists every log error published when a log engine initialises.
// Kind is resolved from the application error context at registration time.
var Errors = []interError.Option{}

var (
	LogDirect  = domainEngine.EngineDep(GroupLog, NameDirect)
	LogFanout  = domainEngine.EngineDep(GroupLog, NameFanout)
	LogPattern = domainEngine.EngineDep(GroupLog, NamePattern)
)
