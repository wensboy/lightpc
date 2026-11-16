package log

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestFileHandler(t *testing.T) (*FileHandler, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "app.log")
	return NewFileHandler(FileOption{Path: path, Mode: 0o644}), path
}

func TestNewFileHandlerWritesToFile(t *testing.T) {
	handler, path := newTestFileHandler(t)

	slog.New(handler).Info("hello", "name", "lightpc")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "hello") || !strings.Contains(string(content), "name=lightpc") {
		t.Fatalf("content = %q", content)
	}
}

func TestFileHandlerEnabledFiltersLevel(t *testing.T) {
	handler, _ := newTestFileHandler(t)
	handler.WithLevel(slog.LevelWarn)

	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("info must be disabled at warn level")
	}
	if !handler.Enabled(context.Background(), slog.LevelError) {
		t.Fatal("error must be enabled at warn level")
	}
}

func TestFileHandlerHandleDoesNotFilterByLevel(t *testing.T) {
	handler, path := newTestFileHandler(t)
	handler.WithLevel(slog.LevelError)

	record := slog.NewRecord(time.Time{}, slog.LevelDebug, "debug record", 0)
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "debug record") {
		t.Fatalf("content = %q", content)
	}
}

func TestFileHandlerSilentWhenOpenFails(t *testing.T) {
	handler := NewFileHandler(FileOption{Path: filepath.Join(t.TempDir(), "missing", "app.log")})

	if handler.Enabled(context.Background(), slog.LevelError) {
		t.Fatal("a handler that failed to open the file must stay disabled")
	}

	record := slog.NewRecord(time.Time{}, slog.LevelInfo, "silent", 0)
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("err = %v", err)
	}
	if err := handler.Close(); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestFileHandlerWithAttrs(t *testing.T) {
	handler, path := newTestFileHandler(t)

	logger := slog.New(handler.WithAttrs([]slog.Attr{slog.String("component", "test")}))
	logger.Info("with attrs")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "component=test") {
		t.Fatalf("content = %q", content)
	}
}

func TestFileHandlerWithGroup(t *testing.T) {
	handler, path := newTestFileHandler(t)

	logger := slog.New(handler.WithGroup("grp"))
	logger.Info("grouped", "key", "value")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "grp.key=value") {
		t.Fatalf("content = %q", content)
	}
}

func TestFileHandlerWithSerializer(t *testing.T) {
	handler, _ := newTestFileHandler(t)
	var builder strings.Builder
	handler.WithSerializer(slog.NewJSONHandler(&builder, nil))

	slog.New(handler).Info("json record")

	if !strings.Contains(builder.String(), `"msg":"json record"`) {
		t.Fatalf("content = %q", builder.String())
	}
}

func TestFileHandlerCloseReleasesFile(t *testing.T) {
	handler, path := newTestFileHandler(t)

	slog.New(handler).Info("before close")
	if err := handler.Close(); err != nil {
		t.Fatal(err)
	}
	if err := handler.Close(); err != nil {
		t.Fatalf("second close = %v", err)
	}

	slog.New(handler).Info("after close")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "before close") {
		t.Fatalf("content = %q", content)
	}
	if strings.Contains(string(content), "after close") {
		t.Fatalf("content = %q", content)
	}
}

func TestFileHandlerDerivedHandlerOwnsNoResource(t *testing.T) {
	handler, path := newTestFileHandler(t)

	derived, ok := handler.WithAttrs(nil).(*FileHandler)
	if !ok {
		t.Fatal("the derived handler must be a *FileHandler")
	}
	if err := derived.Close(); err != nil {
		t.Fatal(err)
	}

	slog.New(handler).Info("still open")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "still open") {
		t.Fatalf("content = %q", content)
	}
}

func TestNewConsoleHandlerWritesToOut(t *testing.T) {
	buffer := &bytes.Buffer{}
	handler := NewConsoleHandler(ConsoleOption{Out: buffer})

	slog.New(handler).Info("hello", "name", "lightpc")

	if !strings.Contains(buffer.String(), "hello") || !strings.Contains(buffer.String(), "name=lightpc") {
		t.Fatalf("content = %q", buffer.String())
	}
}

func TestNewConsoleHandlerDefaultsToStdout(t *testing.T) {
	if NewConsoleHandler(ConsoleOption{}).out != os.Stdout {
		t.Fatal("the console handler must default to standard output")
	}
}

func TestConsoleHandlerEnabledFiltersLevel(t *testing.T) {
	handler := NewConsoleHandler(ConsoleOption{Out: &bytes.Buffer{}}).WithLevel(slog.LevelWarn)

	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("info must be disabled at warn level")
	}
	if !handler.Enabled(context.Background(), slog.LevelError) {
		t.Fatal("error must be enabled at warn level")
	}
}

func TestConsoleHandlerWithAttrsAndGroup(t *testing.T) {
	buffer := &bytes.Buffer{}
	handler := NewConsoleHandler(ConsoleOption{Out: buffer})

	logger := slog.New(handler.WithAttrs([]slog.Attr{slog.String("component", "test")}).WithGroup("grp"))
	logger.Info("grouped", "key", "value")

	content := buffer.String()
	if !strings.Contains(content, "component=test") || !strings.Contains(content, "grp.key=value") {
		t.Fatalf("content = %q", content)
	}
}

func TestConsoleHandlerCloseStopsWrites(t *testing.T) {
	buffer := &bytes.Buffer{}
	handler := NewConsoleHandler(ConsoleOption{Out: buffer})

	if err := handler.Close(); err != nil {
		t.Fatal(err)
	}
	if err := handler.Close(); err != nil {
		t.Fatalf("second close = %v", err)
	}

	slog.New(handler).Info("after close")
	if buffer.Len() != 0 {
		t.Fatalf("content = %q", buffer.String())
	}
}
