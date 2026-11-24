package engineLogImpl

import (
	"encoding/json"
	"log/slog"

	engineLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	interError "github.com/wensboy/lightpc/backend/internal/error"
	interLog "github.com/wensboy/lightpc/backend/internal/log"
)

// Handler names a configured entry may declare.
const (
	handlerFile    = "file"
	handlerConsole = "console"
)

// logEntry is what the engine actually works with: the configured entry with its level parsed and its outputs
// unfolded into one list.
//
// The configuration keeps the output raw because the shape depends on the entry - direct declares one object,
// fanout and pattern declare an array, and the fields differ per handler. Resolving that once here means the
// engine never looks at configuration while a record is being written.
type logEntry struct {
	name    string
	level   slog.Level
	handler string
	outputs []logOutput
}

// logOutput is one resolved output of an entry. Both handlers share the serializer and the path, so a single
// shape carries a console target and a file target alike.
type logOutput struct {
	serializer engineLog.LogSerializer
	path       string
	mode       string
}

// entriesOf resolves every configured entry, reporting the ones that cannot be used.
//
// It lives in shared.go because all three engines prepare entries the same way; only the routing of a finished
// record differs between them.
func entriesOf(items []engineLog.LogEntry) []logEntry {
	entries := make([]logEntry, 0, len(items))
	for _, item := range items {
		outputs := outputsOf(item)
		if len(outputs) == 0 {
			slog.Default().Warn("log item skipped", "name", item.Name, "err", interError.Retrieve(engineLog.ErrKeyInvalidItem))
			continue
		}
		entries = append(entries, logEntry{
			name:    item.Name,
			level:   interLog.ParseLevel(item.Level),
			handler: item.Handler,
			outputs: outputs,
		})
	}
	return entries
}

// outputsOf unfolds the output of an entry into a list.
//
// The declared shape follows the sub-group: direct writes "output": {...} while fanout and pattern write
// "output": [{...}]. Trying the array first and falling back to the object handles both without the caller having
// to say which group it read the entry from.
func outputsOf(item engineLog.LogEntry) []logOutput {
	if len(item.Output) == 0 {
		return nil
	}

	var list []engineLog.FileOutput
	if err := json.Unmarshal(item.Output, &list); err == nil {
		return convert(list)
	}

	var single engineLog.FileOutput
	if err := json.Unmarshal(item.Output, &single); err != nil {
		return nil
	}
	return convert([]engineLog.FileOutput{single})
}

// convert maps the configured outputs onto the shape the engine keeps. FileOutput and ConsoleOutput differ only in
// the mode field, which a console target simply leaves empty.
func convert(outputs []engineLog.FileOutput) []logOutput {
	converted := make([]logOutput, 0, len(outputs))
	for _, output := range outputs {
		converted = append(converted, logOutput{
			serializer: output.Serializer,
			path:       output.Path,
			mode:       output.Mode,
		})
	}
	return converted
}
