package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
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
