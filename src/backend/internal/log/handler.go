package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
)

// FileHandler 用于封装内置text或者json作为编码处理器, 将日志输出到文件中.
type FileHandler struct {
	serializer slog.Handler // 序列化器
	out        io.Writer    // out输出

	mu sync.Mutex

	level slog.Level
}

type FileOption struct {
	Path string
	Mode int
}

func NewFileHandler(opt FileOption) *FileHandler {
	out := io.Writer(os.Stdout)
	if opt.Path != "" {
		mode := os.FileMode(opt.Mode)
		if mode == 0 {
			mode = 0o644
		}
		file, err := os.OpenFile(opt.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, mode)
		if err == nil {
			out = file
		}
	}
	fileHandler := &FileHandler{out: out}
	fileHandler.serializer = slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug})
	return fileHandler
}

func (fh *FileHandler) WithSerializer(s slog.Handler) *FileHandler {
	fh.serializer = s
	return fh
}

func (fh *FileHandler) WithOut(o io.Writer) *FileHandler {
	fh.out = o
	return fh
}

func (fh *FileHandler) WithLevel(level slog.Level) *FileHandler {
	fh.level = level
	return fh
}

func (fh *FileHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= fh.level
}

func (fh *FileHandler) Handle(ctx context.Context, record slog.Record) error {
	fh.mu.Lock()
	defer fh.mu.Unlock()
	return fh.serializer.Handle(ctx, record)
}

func (fh *FileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &FileHandler{
		serializer: fh.serializer.WithAttrs(attrs),
		out:        fh.out,
		level:      fh.level,
	}
}

func (fh *FileHandler) WithGroup(name string) slog.Handler {
	return &FileHandler{
		serializer: fh.serializer.WithGroup(name),
		out:        fh.out,
		level:      fh.level,
	}
}
