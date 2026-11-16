// Package log defines the logging engine contracts: engine names and the direct log capabilities.
package log

import "log/slog"

// DirectClient is the client side of the direct log engine. A caller resolves a logger by an exact
// name and writes records through it, so one record is always consumed by exactly one handler.
type DirectClient interface {
	// Logger returns the logger bound to name. Matching is exact; an unknown name reports false.
	Logger(name string) (*slog.Logger, bool)
}

// DirectServer is the server side of the direct log engine. It owns the handler registry that the
// client side resolves against.
type DirectServer interface {
	// Register binds handler to name, replacing any handler already bound to it.
	// An empty name or a nil handler is ignored.
	Register(name string, handler slog.Handler)
}
