package log

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jmoiron/sqlx"
)

// FileHandler wraps the built in text serializer and writes records into a file.
//
// The file is opened at construction and owned by the handler: Close releases it and is idempotent.
// Handlers derived by WithAttrs / WithGroup own no resource, so their Close is a no op.
type FileHandler struct {
	serializer slog.Handler
	out        io.Writer
	closer     io.Closer // nil 表示输出目标不由本 handler 拥有

	mu     sync.Mutex
	level  slog.Level
	closed atomic.Bool
}

// FileOption configures a FileHandler.
type FileOption struct {
	SerializerOption
	Serializer string
	Path       string
	Mode       int
}

// NewFileHandler opens Path and builds a handler on top of it. The handler never falls back to another
// writer: when the file cannot be opened it stays silent, and Enabled reports false so no record is
// handled at all.
func NewFileHandler(opt FileOption) *FileHandler {
	mode := os.FileMode(opt.Mode)
	if mode == 0 {
		mode = 0o644
	}
	fileHandler := &FileHandler{}
	if file, err := os.OpenFile(opt.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, mode); err == nil {
		fileHandler.out = file
		fileHandler.closer = file
		fileHandler.serializer = NewSerializer(opt.Serializer, opt.SerializerOption, file)
	}
	return fileHandler
}

// WithSerializer replaces the serializer; it must be called before concurrent use.
func (fh *FileHandler) WithSerializer(s slog.Handler) *FileHandler {
	fh.serializer = s
	return fh
}

// WithLevel sets the minimum level; it must be called before concurrent use.
func (fh *FileHandler) WithLevel(level slog.Level) *FileHandler {
	fh.level = level
	return fh
}

// Close releases the file opened by NewFileHandler and is idempotent.
func (fh *FileHandler) Close() error {
	if !fh.closed.CompareAndSwap(false, true) {
		return nil
	}
	if fh.closer == nil {
		return nil
	}
	return fh.closer.Close()
}

// Enabled filters here so Handle only ever sees records that can actually be written.
func (fh *FileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= fh.level && fh.out != nil && fh.serializer != nil && !fh.closed.Load()
}

// Handle writes record to the underlying file.
func (fh *FileHandler) Handle(ctx context.Context, record slog.Record) error {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	if fh.closed.Load() || fh.serializer == nil {
		return nil
	}
	return fh.serializer.Handle(ctx, record)
}

// WithAttrs derives a handler that owns no file resource.
func (fh *FileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FileHandler{
		serializer: fh.serializer.WithAttrs(attrs),
		out:        fh.out,
		level:      fh.level,
	}
}

// WithGroup derives a handler that owns no file resource.
func (fh *FileHandler) WithGroup(name string) slog.Handler {
	return &FileHandler{
		serializer: fh.serializer.WithGroup(name),
		out:        fh.out,
		level:      fh.level,
	}
}

// ConsoleHandler wraps the built in text serializer and writes records to a console stream. It never
// owns the stream, so Close only stops further writes.
type ConsoleHandler struct {
	serializer slog.Handler
	out        io.Writer

	mu     sync.Mutex
	level  slog.Level
	closed atomic.Bool
}

// ConsoleOption configures a ConsoleHandler.
type ConsoleOption struct {
	SerializerOption
	Serializer string
	// Out defaults to os.Stdout when nil; pass os.Stderr to route to the error stream.
	Out io.Writer
}

// NewConsoleHandler builds a handler on top of opt.Out, defaulting to standard output.
func NewConsoleHandler(opt ConsoleOption) *ConsoleHandler {
	out := opt.Out
	if out == nil {
		out = os.Stdout
	}
	consoleHandler := &ConsoleHandler{out: out}
	consoleHandler.serializer = NewSerializer(opt.Serializer, opt.SerializerOption, out)
	return consoleHandler
}

// WithSerializer replaces the serializer; it must be called before concurrent use.
func (ch *ConsoleHandler) WithSerializer(s slog.Handler) *ConsoleHandler {
	ch.serializer = s
	return ch
}

// WithLevel sets the minimum level; it must be called before concurrent use.
func (ch *ConsoleHandler) WithLevel(level slog.Level) *ConsoleHandler {
	ch.level = level
	return ch
}

// Close stops further writes and is idempotent; the stream itself is never closed.
func (ch *ConsoleHandler) Close() error {
	ch.closed.CompareAndSwap(false, true)
	return nil
}

// Enabled filters here so Handle only ever sees records that can actually be written.
func (ch *ConsoleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= ch.level && ch.out != nil && ch.serializer != nil && !ch.closed.Load()
}

// Handle writes record to the console stream.
func (ch *ConsoleHandler) Handle(ctx context.Context, record slog.Record) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	if ch.closed.Load() || ch.serializer == nil {
		return nil
	}
	return ch.serializer.Handle(ctx, record)
}

// WithAttrs derives a handler that shares the same stream.
func (ch *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ConsoleHandler{
		serializer: ch.serializer.WithAttrs(attrs),
		out:        ch.out,
		level:      ch.level,
	}
}

// WithGroup derives a handler that shares the same stream.
func (ch *ConsoleHandler) WithGroup(name string) slog.Handler {
	return &ConsoleHandler{
		serializer: ch.serializer.WithGroup(name),
		out:        ch.out,
		level:      ch.level,
	}
}

// DatabaseHandler writes records into a table through the caller's own statement.
//
// It talks to *sqlx.DB rather than to an engine interface: the handler only needs a pool and a statement, and
// depending on the store engine would make the logging layer depend on the storage layer it may be describing.
//
// The statement is supplied by the caller and receives a map, so a table with extra columns needs no change here.
// The three fields every record has are named level, msg and time; everything else the record carries is folded
// into detail, which is the one column whose shape is deliberately open.
type DatabaseHandler struct {
	db         *sqlx.DB
	statement  string
	serializer string
	listener   string
	source     bool
	attrs      []slog.Attr
	group      string

	mu     sync.Mutex
	level  slog.Level
	closed atomic.Bool
}

// DatabaseOption configures a DatabaseHandler.
type DatabaseOption struct {
	SerializerOption
	// Serializer selects how the extra attributes are folded into detail: "json" or "text".
	Serializer string
	// Db is the pool the handler inserts through. The handler never opens or closes it.
	Db *sqlx.DB
	// Statement is the insert to run. It is named, so its parameters are the keys of the map the handler builds.
	Statement string
	// Listener names the source a record belongs to. It travels in the map, so the statement can fill the column.
	Listener string
}

// NewDatabaseHandler builds a handler on top of opt.Db and opt.Statement.
//
// A missing pool or statement leaves the handler disabled rather than falling back somewhere else: a log line that
// cannot be stored is better dropped than written into the wrong table.
func NewDatabaseHandler(opt DatabaseOption) *DatabaseHandler {
	return &DatabaseHandler{
		db:         opt.Db,
		statement:  opt.Statement,
		serializer: opt.Serializer,
		listener:   opt.Listener,
		source:     opt.AddSource,
	}
}

// WithLevel sets the minimum level; it must be called before concurrent use.
func (dh *DatabaseHandler) WithLevel(level slog.Level) *DatabaseHandler {
	dh.level = level
	return dh
}

// Close stops further writes and is idempotent; the pool is owned by the caller and is never closed here.
func (dh *DatabaseHandler) Close() error {
	dh.closed.CompareAndSwap(false, true)
	return nil
}

// Enabled filters here so Handle only ever sees records that can actually be stored.
func (dh *DatabaseHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= dh.level && dh.db != nil && dh.statement != "" && !dh.closed.Load()
}

// Handle inserts record through the configured statement.
//
// The map is built per record, so a statement may name any subset of the keys; sqlx rejects a placeholder it cannot
// resolve, which turns a mismatched statement into an error at the first record instead of a silent empty column.
func (dh *DatabaseHandler) Handle(ctx context.Context, record slog.Record) error {
	dh.mu.Lock()
	defer dh.mu.Unlock()
	if dh.closed.Load() || dh.db == nil || dh.statement == "" {
		return nil
	}

	params := map[string]any{
		"listener": dh.listener,
		"level":    int64(record.Level),
		"msg":      record.Message,
		"time":     record.Time.UnixMilli(),
		"detail":   dh.detailOf(record),
	}
	if dh.source && record.PC != 0 {
		frame, _ := runtime.CallersFrames([]uintptr{record.PC}).Next()
		params["source"] = frame.File + ":" + strconv.Itoa(frame.Line)
	}

	_, err := dh.db.NamedExecContext(ctx, dh.statement, params)
	return err
}

// detailOf folds every attribute of record into one string.
//
// The two serializers differ only in shape: json keeps the attributes addressable, text keeps the line readable in
// a terminal. An empty result is returned as an empty string rather than omitted, because the column is not null.
func (dh *DatabaseHandler) detailOf(record slog.Record) string {
	attributes := map[string]any{}
	for _, attr := range dh.attrs {
		attributes[dh.group+attr.Key] = attr.Value.Any()
	}
	record.Attrs(func(attr slog.Attr) bool {
		attributes[attr.Key] = attr.Value.Any()
		return true
	})
	if len(attributes) == 0 {
		return ""
	}

	if strings.EqualFold(dh.serializer, "json") {
		encoded, err := json.Marshal(attributes)
		if err != nil {
			return ""
		}
		return string(encoded)
	}

	keys := make([]string, 0, len(attributes))
	for key := range attributes {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for index, key := range keys {
		if index > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(key)
		builder.WriteString("=")
		fmt.Fprint(&builder, attributes[key])
	}
	return builder.String()
}

// WithAttrs derives a handler that stores the attributes with every record.
//
// The attributes are kept on the derived value and merged in detailOf, so a logger built with a group writes the
// same rows as one that is not.
func (dh *DatabaseHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Constructed field by field rather than copied: the struct holds a mutex, and a copy would carry its state.
	return &DatabaseHandler{
		db:         dh.db,
		statement:  dh.statement,
		serializer: dh.serializer,
		listener:   dh.listener,
		source:     dh.source,
		attrs:      append(append([]slog.Attr(nil), dh.attrs...), attrs...),
		group:      dh.group,
		level:      dh.level,
	}
}

// WithGroup derives a handler that prefixes the attributes it stores.
func (dh *DatabaseHandler) WithGroup(name string) slog.Handler {
	return &DatabaseHandler{
		db:         dh.db,
		statement:  dh.statement,
		serializer: dh.serializer,
		listener:   dh.listener,
		source:     dh.source,
		attrs:      append([]slog.Attr(nil), dh.attrs...),
		group:      dh.group + name + ".",
		level:      dh.level,
	}
}
