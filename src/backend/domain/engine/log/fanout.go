package log

import "log/slog"

// FanoutClient is the client side of the fanout log engine. One record is delivered to every
// handler bound to the resolved name.
type FanoutClient interface {
	// Logger returns a logger broadcasting to every handler bound to name.
	// An unknown name reports false.
	Logger(name string) (*slog.Logger, bool)
}

// FanoutServer is the server side of the fanout log engine. It owns the handler sets the client
// side broadcasts to.
type FanoutServer interface {
	// Register appends handler to the handler set bound to name.
	// An empty name or a nil handler is ignored.
	Register(name string, handler slog.Handler)
}
