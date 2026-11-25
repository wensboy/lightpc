package engineLogImpl

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	"github.com/wensboy/lightpc/backend/internal/cast"
	"github.com/wensboy/lightpc/backend/internal/config"
	interError "github.com/wensboy/lightpc/backend/internal/error"
	interLog "github.com/wensboy/lightpc/backend/internal/log"
)

// poolSource is the one capability the database handler needs from the store: the pool behind a key. Declaring it
// here keeps this file free of the concrete store type.
type poolSource interface {
	RawDB(key string) (*sqlx.DB, bool)
}

// handlersFor opens every output of one entry and returns the bound handlers.
//
// It runs during Init rather than at construction, because the store engine has created the root directory by
// then. Level filtering belongs to the outer handler, so the serializer keeps its own default.
func handlersFor(entry logEntry, pools poolSource) []slog.Handler {
	handlers := make([]slog.Handler, 0, len(entry.outputs))
	for _, output := range entry.outputs {
		serializer := interLog.SerializerOption{AddSource: output.serializer.EnableSource}

		if entry.handler == handlerDatabase {
			// The path names the pool, not a file: the store registry keys its connections that way.
			if pool, ok := pools.RawDB(output.path); ok {
				handlers = append(handlers, interLog.NewDatabaseHandler(interLog.DatabaseOption{
					SerializerOption: serializer,
					Serializer:       output.serializer.Type,
					Db:               pool,
					Statement:        logInsertStatement,
					Listener:         entry.name,
				}).WithLevel(entry.level))
			} else {
				slog.Default().Warn("log output skipped", "path", output.path, "err", interError.Retrieve(engineLog.ErrKeyHandlerFailed))
			}
			continue
		}

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

// poolSourceOf returns the pool accessor of the store engine, or nil when this build has no database.
//
// A log group that only writes to a file never asks for it, so a missing database is not an error here: the handler
// loop simply finds nothing to insert into and the entry falls back to its other outputs.
func poolSourceOf(engines map[string]domainEngine.Engine) poolSource {
	key := domainEngine.EngineDep(domainEngine.GroupStore, domainEngine.NameDb)
	engine, found := engines[key]
	if !found {
		return nil
	}
	source, _ := cast.As[poolSource](engine)
	return source
}
