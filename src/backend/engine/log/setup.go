package logimpl

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	"github.com/wensboy/lightpc/backend/internal/cast"
	interLog "github.com/wensboy/lightpc/backend/internal/log"
)

// Setup builds the log engines described by runtime.log.<group>[].
//
// Each group yields at most one engine: every item of the group is registered into that engine, and a
// missing or empty group yields no engine at all. The first direct entry becomes the slog default.
func Setup(cfg domainContext.ConfigContext) []domainEngine.Engine {
	engines := make([]domainEngine.Engine, 0, 3)
	value, ok := cfg.Lookup("default.runtime.log")
	if !ok {
		return engines
	}
	groups, ok := cast.To[map[string]any](value)
	if !ok {
		return engines
	}

	if items, ok := cast.To[[]any](groups["direct"]); ok && len(items) > 0 {
		options := make([]DirectOption, 0, len(items))
		defaultName := ""
		for _, raw := range items {
			name, handlers := setupItem(raw)
			if name == "" {
				continue
			}
			if defaultName == "" {
				defaultName = name
			}
			for _, handler := range handlers {
				options = append(options, WithDirectHandler(name, handler))
			}
		}
		if len(options) > 0 {
			directLog := NewDirectLog(options...)
			engines = append(engines, directLog)
			if logger, ok := directLog.Logger(defaultName); ok {
				slog.SetDefault(logger)
			}
		}
	}

	if items, ok := cast.To[[]any](groups["fanout"]); ok && len(items) > 0 {
		options := make([]FanoutOption, 0, len(items))
		for _, raw := range items {
			name, handlers := setupItem(raw)
			if name == "" {
				continue
			}
			for _, handler := range handlers {
				options = append(options, WithFanoutHandler(name, handler))
			}
		}
		if len(options) > 0 {
			engines = append(engines, NewFanoutLog(options...))
		}
	}

	if items, ok := cast.To[[]any](groups["pattern"]); ok && len(items) > 0 {
		options := make([]PatternOption, 0, len(items))
		for _, raw := range items {
			name, handlers := setupItem(raw)
			if name == "" {
				continue
			}
			for _, handler := range handlers {
				options = append(options, WithPatternHandler(name, handler))
			}
		}
		if len(options) > 0 {
			engines = append(engines, NewPatternLog(options...))
		}
	}

	return engines
}

// setupItem parses one configured item and builds every handler it declares. An empty name is
// reported when the entry has no usable output, which lets the caller skip it.
func setupItem(raw any) (string, []slog.Handler) {
	record, ok := cast.To[map[string]any](raw)
	if !ok {
		return "", nil
	}
	item := domainConfig.LogItem{}
	cast.ToStruct(record, &item, "json")
	if item.Name == "" {
		return "", nil
	}

	level := interLog.ParseLevel(item.Level)

	// The direct group declares one output object while fanout and pattern declare an array, so both
	// shapes are normalised into a single list here.
	var outputs []any
	if list, ok := cast.To[[]any](item.Output); ok {
		outputs = list
	} else if item.Output != nil {
		outputs = []any{item.Output}
	}

	handlers := make([]slog.Handler, 0, len(outputs))
	for _, rawOutput := range outputs {
		source, ok := cast.To[map[string]any](rawOutput)
		if !ok {
			continue
		}
		output := domainConfig.LogOutput{}
		cast.ToStruct(source, &output, "json")

		// Level filtering belongs to the outer handler, so the serializer keeps its default level.
		serializerOption := interLog.SerializerOption{AddSource: output.Serializer.EnableSource}

		if item.Handler == "file" {
			path := output.Path
			if strings.HasPrefix(path, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					continue
				}
				path = filepath.Join(home, strings.TrimPrefix(path, "~"))
			}
			if path == "" {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				continue
			}
			handlers = append(handlers, interLog.NewFileHandler(interLog.FileOption{
				SerializerOption: serializerOption,
				Serializer:       output.Serializer.Type,
				Path:             path,
				Mode:             int(interLog.ParseMode(output.Mode)),
			}).WithLevel(level))
			continue
		}

		out := io.Writer(os.Stdout)
		if strings.EqualFold(output.Out, "stderr") {
			out = os.Stderr
		}
		handlers = append(handlers, interLog.NewConsoleHandler(interLog.ConsoleOption{
			SerializerOption: serializerOption,
			Serializer:       output.Serializer.Type,
			Out:              out,
		}).WithLevel(level))
	}
	if len(handlers) == 0 {
		return "", nil
	}
	return item.Name, handlers
}
