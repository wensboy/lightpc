package serviceLog

import (
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// Error keys published by the log service during Init.
const (
	// ErrKeyDatabaseMissing reports that the configured database is not reachable.
	ErrKeyDatabaseMissing = "error::service.log::database_missing"
	// ErrKeyListenerUnknown reports a request for a source that is not registered.
	ErrKeyListenerUnknown = "error::service.log::listener_unknown"
)

// Errors lists every error this service publishes.
var Errors = []interError.Option{
	{Key: ErrKeyDatabaseMissing, Code: interError.ErrCodeNotExist, Msg: "log: database unavailable"},
	{Key: ErrKeyListenerUnknown, Code: interError.ErrCodeInvalidParam, Msg: "log: unknown listener"},
}

// LogRecordDto is one record the console shows.
//
// Level carries the slog.Level value unchanged, so a record crosses the boundary without a translation and the
// order of severity survives: debug -4, info 0, warn 4, error 8.
type LogRecordDto struct {
	ID        int64  `json:"id"`
	Listener  string `json:"listener"`
	Level     int64  `json:"level"`
	Msg       string `json:"msg"`
	Detail    string `json:"detail"`
	CreatedAt int64  `json:"createdAt"`
}

// ListenerDto describes one source the service can read.
type ListenerDto struct {
	// Name is the identifier the interface selects.
	Name string `json:"name"`
	// Title is the label rendered in the picker.
	Title string `json:"title"`
}

// LogRecordsDto is the answer of one read.
//
// Boundary is the instant the read started from. Returning it lets the interface show what "cleared" currently
// means without asking a second question.
type LogRecordsDto struct {
	Records  []LogRecordDto `json:"records"`
	Boundary int64          `json:"boundary"`
}

// LogQueryDto is what the interface sends with a read.
//
// Only the source is named. The start instant is not a client concern: each source keeps its own boundary in the
// service, so a request cannot widen a range the user already cleared.
type LogQueryDto struct {
	Listener string `json:"listener"`
	// Limit caps the answer at the newest records. Zero falls back to the service default.
	Limit int `json:"limit"`
}

// FilterParamsDto carries every filter parameter in one payload.
//
// The interface does not know which filter reads which field; it sends what the toolbar holds and the service
// hands the same value to every filter. A filter ignores the parts it does not own, which is what keeps one filter
// transparent to the others.
type FilterParamsDto struct {
	// Levels lists the levels allowed through. An empty list disables the level filter.
	Levels []int64 `json:"levels"`
	// Pattern is the search text. An empty string disables the pattern filter.
	Pattern string `json:"pattern"`
}
