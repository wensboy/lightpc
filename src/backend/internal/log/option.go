package log

import (
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	// SerializerText is the default serializer name.
	SerializerText = "text"
	// SerializerJSON selects the JSON serializer.
	SerializerJSON = "json"
)

// ParseLevel maps the configured level text to slog.Level. Unknown or empty text falls back to
// LevelInfo so a typo never silences a sink.
func ParseLevel(text string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(text)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// ParseMode maps the configured octal text such as "0o644" to os.FileMode. Invalid or empty text
// falls back to 0o644.
func ParseMode(text string) os.FileMode {
	if text == "" {
		return 0o644
	}
	mode, err := strconv.ParseUint(strings.TrimSpace(text), 0, 32)
	if err != nil {
		return 0o644
	}
	return os.FileMode(mode)
}

// NewSerializer builds the serializer named by kind. Anything other than SerializerJSON yields the
// text serializer, so an unknown value degrades to the default instead of failing.
func NewSerializer(kind string, option SerializerOption, out io.Writer) slog.Handler {
	if strings.EqualFold(strings.TrimSpace(kind), SerializerJSON) {
		return slog.NewJSONHandler(out, &option)
	}
	return slog.NewTextHandler(out, &option)
}
