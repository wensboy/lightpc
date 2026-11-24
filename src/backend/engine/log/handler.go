package engineLogImpl

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	engineLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	"github.com/wensboy/lightpc/backend/internal/config"
	interError "github.com/wensboy/lightpc/backend/internal/error"
	interLog "github.com/wensboy/lightpc/backend/internal/log"
)

// handlersFor opens every output of one entry and returns the bound handlers.
//
// It runs during Init rather than at construction, because the store engine has created the root directory by
// then. Level filtering belongs to the outer handler, so the serializer keeps its own default.
func handlersFor(entry logEntry) []slog.Handler {
	handlers := make([]slog.Handler, 0, len(entry.outputs))
	for _, output := range entry.outputs {
		serializer := interLog.SerializerOption{AddSource: output.serializer.EnableSource}

		if entry.handler == handlerFile {
			if filePath, ok := fileTarget(output.path, entry.name); ok {
				handlers = append(handlers, interLog.NewFileHandler(interLog.FileOption{
					SerializerOption: serializer,
					Serializer:       output.serializer.Type,
					Path:             filePath,
					Mode:             int(interLog.ParseMode(output.mode)),
				}).WithLevel(entry.level))
			}
			continue
		}

		out := io.Writer(os.Stdout)
		if strings.EqualFold(output.path, "stderr") {
			out = os.Stderr
		}
		handlers = append(handlers, interLog.NewConsoleHandler(interLog.ConsoleOption{
			SerializerOption: serializer,
			Serializer:       output.serializer.Type,
			Out:              out,
		}).WithLevel(entry.level))
	}
	return handlers
}

// fileTarget resolves the configured path of a file output to the absolute path the handler writes to.
//
// A path that names an embedded resource is refused: embedded content is read only, and a log file has to be
// written. The reason is logged here so a misconfigured entry is visible without the engine failing.
func fileTarget(path, name string) (string, bool) {
	if path == "" {
		slog.Default().Warn("log output skipped", "name", name, "err", interError.Retrieve(engineLog.ErrKeyInvalidItem))
		return "", false
	}

	resource, err := config.NewResource(path)
	if err != nil || resource.IsEmbed() {
		slog.Default().Warn("log output skipped", "path", path, "err", interError.Retrieve(engineLog.ErrKeyHandlerFailed).Wrap(err))
		return "", false
	}

	filePath, err := resource.Path()
	if err != nil {
		slog.Default().Warn("log output skipped", "path", path, "err", interError.Retrieve(engineLog.ErrKeyHandlerFailed).Wrap(err))
		return "", false
	}
	return filePath, true
}

// PrepareOutputs creates the directory of every file output.
//
// It is exported for the hook that wraps a log engine: deciding where a log file lives and creating the directory
// are bootstrap concerns, while building the handler stays with the engine.
func PrepareOutputs(items []engineLog.LogEntry) error {
	for _, entry := range entriesOf(items) {
		for _, output := range entry.outputs {
			if entry.handler != handlerFile {
				continue
			}
			filePath, ok := fileTarget(output.path, entry.name)
			if !ok {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
				slog.Default().Warn("log output directory skipped", "path", filePath, "err", err)
			}
		}
	}
	return nil
}
