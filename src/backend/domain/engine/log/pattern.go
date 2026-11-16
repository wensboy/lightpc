package log

import "log/slog"

// PatternClient is the client side of the pattern log engine. A name is matched against the
// registered patterns and one record is delivered to every handler of every matching pattern.
type PatternClient interface {
	// Logger returns a logger broadcasting to every handler whose pattern matches name.
	// Matching is glob based; when nothing matches it reports false.
	Logger(name string) (*slog.Logger, bool)
}

// PatternServer is the server side of the pattern log engine. It owns the pattern registry the
// client side matches against.
type PatternServer interface {
	// Register appends handler to the handler set bound to pattern.
	// An empty pattern or a nil handler is ignored.
	Register(pattern string, handler slog.Handler)
}
