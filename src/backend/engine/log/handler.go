package engineLogImpl

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	engineLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	"github.com/wensboy/lightpc/backend/internal/cast"
	interError "github.com/wensboy/lightpc/backend/internal/error"
	interLog "github.com/wensboy/lightpc/backend/internal/log"
	"github.com/wensboy/lightpc/backend/internal/path"
)

// logEntry is one configured logger: a name, its level and the outputs it writes to.
type logEntry struct {
	name    string
	level   slog.Level
	handler string
	outputs []domainConfig.LogOutput
}

// parseItem reads one configured item. It reports false when the item carries no name.
func parseItem(raw any) (domainConfig.LogItem, bool) {
	record, ok := cast.To[map[string]any](raw)
	if !ok {
		return domainConfig.LogItem{}, false
	}
	item := domainConfig.LogItem{}
	cast.ToStruct(record, &item, "json")
	if item.Name == "" {
		return domainConfig.LogItem{}, false
	}
	return item, true
}

// parseEntry normalises the output shape of item. The direct group declares one output object while
// fanout and pattern declare an array, so both are folded into a single list here.
func parseEntry(item domainConfig.LogItem) (logEntry, bool) {
	var outputs []domainConfig.LogOutput
	if list, ok := cast.To[[]any](item.Output); ok {
		for _, rawOutput := range list {
			if output, ok := parseOutput(rawOutput); ok {
				outputs = append(outputs, output)
			}
		}
	} else if item.Output != nil {
		if output, ok := parseOutput(item.Output); ok {
			outputs = append(outputs, output)
		}
	}
	if len(outputs) == 0 {
		return logEntry{}, false
	}
	return logEntry{
		name:    item.Name,
		level:   interLog.ParseLevel(item.Level),
		handler: item.Handler,
		outputs: outputs,
	}, true
}

// parseOutput reads one output object.
func parseOutput(raw any) (domainConfig.LogOutput, bool) {
	source, ok := cast.To[map[string]any](raw)
	if !ok {
		return domainConfig.LogOutput{}, false
	}
	output := domainConfig.LogOutput{}
	cast.ToStruct(source, &output, "json")
	return output, true
}

// entriesOf parses every configured item, reporting the ones that cannot be used.
func entriesOf(items []domainConfig.LogItem) []logEntry {
	entries := make([]logEntry, 0, len(items))
	for _, item := range items {
		entry, ok := parseEntry(item)
		if !ok {
			slog.Default().Warn("log item skipped", "name", item.Name, "err", interError.Retrieve(engineLog.ErrKeyInvalidItem))
			continue
		}
		entries = append(entries, entry)
	}
	return entries
}

// buildHandlers opens every output of entry. It runs during Init instead of construction so the store
// engine has already created the root directory the log files live in.
func buildHandlers(entry logEntry) []slog.Handler {
	handlers := make([]slog.Handler, 0, len(entry.outputs))
	for _, output := range entry.outputs {
		// Level filtering belongs to the outer handler, so the serializer keeps its default level.
		serializerOption := interLog.SerializerOption{AddSource: output.Serializer.EnableSource}

		if entry.handler == "file" {
			if output.Path == "" {
				slog.Default().Warn("log output skipped", "name", entry.name, "err", interError.Retrieve(engineLog.ErrKeyInvalidItem))
				continue
			}
			target, err := path.Parse(output.Path)
			if err != nil || target.IsEmbed() {
				slog.Default().Warn("log output skipped", "path", output.Path, "err", interError.Retrieve(engineLog.ErrKeyHandlerFailed).Wrap(err))
				continue
			}
			if err := os.MkdirAll(filepath.Dir(target.Path), 0o755); err != nil {
				slog.Default().Warn("log output skipped", "path", target.Path, "err", interError.Retrieve(engineLog.ErrKeyHandlerFailed).Wrap(err))
				continue
			}
			handlers = append(handlers, interLog.NewFileHandler(interLog.FileOption{
				SerializerOption: serializerOption,
				Serializer:       output.Serializer.Type,
				Path:             target.Path,
				Mode:             int(interLog.ParseMode(output.Mode)),
			}).WithLevel(entry.level))
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
		}).WithLevel(entry.level))
	}
	return handlers
}
