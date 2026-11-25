package serviceLog

import "context"

// LogClient is the surface the front end depends on.
type LogClient interface {
	// Listeners returns every source the service can read, ordered by name.
	Listeners() ([]ListenerDto, error)
	// Records returns the records of one source from its boundary onwards, oldest first.
	Records(query LogQueryDto) (LogRecordsDto, error)
	// Clear moves the boundary of one source to now, so a later read starts after the records just shown.
	Clear(listener string) error
	// SetFilter stores the parameters the filters read on the next call.
	SetFilter(params FilterParamsDto) error
}

// Listener is one log source.
//
// A source hides how its records are produced: reading from a table, draining a buffer or tailing a file are all
// the same to the service, which only asks for what came after an instant.
type Listener interface {
	// Name is the identifier the interface selects.
	Name() string
	// Title is the label rendered in the picker.
	Title() string
	// Records returns the records strictly after since, oldest first, at most limit of them.
	Records(ctx context.Context, since int64, limit int) ([]LogRecordDto, error)
}

// Filter narrows a record list.
//
// A filter is a step in a pipeline: it receives the list the previous step produced and returns the part that
// passes. The order of the surviving records is preserved, because the caller renders them top to bottom and the
// list is already sorted by time.
//
// Parameters live inside the implementation. Nothing outside a filter decides what its settings mean, which is
// what keeps one filter invisible to the others.
type Filter interface {
	// Name identifies the step in a log line; it is not part of the contract with the interface.
	Name() string
	// Apply returns the records that pass, in their original order.
	Apply(records []LogRecordDto) []LogRecordDto
}
