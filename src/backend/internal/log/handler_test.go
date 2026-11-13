package log

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewFileHandlerWritesToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	logger := slog.New(NewFileHandler(FileOption{Path: path, Mode: 0o644}))

	logger.Info("hello", "name", "lightpc")

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "hello") || !strings.Contains(string(content), "name=lightpc") {
		t.Fatalf("content = %q", content)
	}
}

func TestFileHandlerEnabledFiltersLevel(t *testing.T) {
	handler := NewFileHandler(FileOption{}).WithLevel(slog.LevelWarn)

	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("info must be disabled at warn level")
	}
	if !handler.Enabled(context.Background(), slog.LevelError) {
		t.Fatal("error must be enabled at warn level")
	}
}

func TestFileHandlerHandleDoesNotFilterByLevel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "level.log")
	handler := NewFileHandler(FileOption{Path: path, Mode: 0o644}).WithLevel(slog.LevelError)

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

func TestFileHandlerWithAttrs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "attrs.log")
	handler := NewFileHandler(FileOption{Path: path, Mode: 0o644})

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
	path := filepath.Join(t.TempDir(), "group.log")
	handler := NewFileHandler(FileOption{Path: path, Mode: 0o644})

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
	var builder strings.Builder
	handler := NewFileHandler(FileOption{}).WithSerializer(slog.NewJSONHandler(&builder, nil))

	slog.New(handler).Info("json record")

	if !strings.Contains(builder.String(), `"msg":"json record"`) {
		t.Fatalf("content = %q", builder.String())
	}
}
