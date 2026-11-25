package serviceLogImpl

// LogDao maps a row of the log table.
//
// Level is the slog.Level value unchanged: debug -4, info 0, warn 4, error 8. Storing it as is means a record
// crosses the boundary without a translation, and the database compares it directly, so a floor is a range scan
// rather than a set of names.
//
// Detail holds key = value pairs, possibly over several lines. It is a string rather than a structure because the
// set of keys is open: the writer decides what to record, and a reader only displays it.
type LogDao struct {
	ID        int64  `json:"id" db:"id"`
	Listener  string `json:"listener" db:"listener"`
	Level     int64  `json:"level" db:"level"`
	Msg       string `json:"msg" db:"msg"`
	Detail    string `json:"detail" db:"detail"`
	CreatedAt int64  `json:"createdAt" db:"created_at"`
}
