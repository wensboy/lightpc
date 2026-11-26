// Package notice defines the notice engine contract: the shape of one notice and the operations a caller uses to
// publish, list and remove them.
package notice

import interError "github.com/wensboy/lightpc/backend/internal/error"

// Error keys published when the notice engine initialises.
const (
	// ErrKeyInvalidArgument reports a notice without a title, or with a level the interface cannot render.
	ErrKeyInvalidArgument = "error::engine.notice::invalid_argument"
)

// Errors lists every notice error published during Init.
var Errors = []interError.Option{
	{Key: ErrKeyInvalidArgument, Code: interError.ErrCodeInvalidParam, Msg: "notice: invalid argument"},
}

// Level is the severity a notice carries. It decides the icon and the colour the interface draws.
type Level string

const (
	LevelInfo    Level = "info"
	LevelSuccess Level = "success"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

// Known reports whether the level is one the interface can render.
func (l Level) Known() bool {
	switch l {
	case LevelInfo, LevelSuccess, LevelWarning, LevelError:
		return true
	default:
		return false
	}
}

// Record is one notice the engine keeps.
//
// It is a value the engine owns, not a dto: nothing here is designed around a binding. The interface receives the
// same fields over an event, and the notice service that follows reads them through the client.
type Record struct {
	ID        uint64 `json:"id"`
	Level     Level  `json:"level"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Source    string `json:"source"`
	Timestamp int64  `json:"timestamp"`
}

// Input is what a caller publishes. ID and Timestamp are the engine's to fill.
type Input struct {
	Level   Level  `json:"level"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Source  string `json:"source"`
}

// Config maps runtime.engines.notice completely.
//
// Buf is the ring length. Zero is not a usable value: the engine treats it as a setup mistake rather than silently
// substituting a size, so a caller that forgets to pass one finds out at start up.
type Config struct {
	Buf int `json:"buf"`
}
