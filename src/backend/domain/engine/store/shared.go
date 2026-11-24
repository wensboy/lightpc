package engineStore

import (
	"encoding/json"
	"log/slog"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// Error keys published by the store engines during Init.
const (
	ErrKeyNotFound = "error::engine.store::not_found"
	ErrKeyExist    = "error::engine.store::already_exists"
	ErrKeyInvalid  = "error::engine.store::invalid_argument"
	ErrKeyEscape   = "error::engine.store::path_escapes_root"
	ErrKeyClosed   = "error::engine.store::closed"
	// ErrKeyReadOnly reports a write against a path served from an embedded file system.
	ErrKeyReadOnly = "error::engine.store::read_only"
	// ErrKeyUnsupportedCommand reports a command type the dispatcher does not know.
	ErrKeyUnsupportedCommand = "error::engine.store::unsupported_command"
)

// Errors lists every store error published when a store engine initialises.
// Kind is resolved from the application error context at registration time.
var Errors = []interError.Option{
	{Key: ErrKeyNotFound, Code: interError.ErrCodeNotExist, Msg: "store: not found"},
	{Key: ErrKeyExist, Code: interError.ErrCodeOpDenied, Msg: "store: already exists"},
	{Key: ErrKeyInvalid, Code: interError.ErrCodeUnknown, Msg: "store: invalid argument"},
	{Key: ErrKeyEscape, Code: interError.ErrCodeOpDenied, Msg: "store: path escapes root"},
	{Key: ErrKeyClosed, Code: interError.ErrCodeFatal, Msg: "store: closed"},
	{Key: ErrKeyReadOnly, Code: interError.ErrCodeOpDenied, Msg: "store: read only"},
	{Key: ErrKeyUnsupportedCommand, Code: interError.ErrCodeInvalidParam, Msg: "store: unsupported command"},
}

// Configuration paths of the store group. Root and the two modes belong to the file store, while the connection
// list belongs to the database store; both live under the same engine group.
const (
	KeyFsRoot     = "default.runtime.engines.store.root"
	KeyFsFileMode = "default.runtime.engines.store.fs.fileMode"
	KeyFsDirMode  = "default.runtime.engines.store.fs.dirMode"
	KeyDbSQL      = "default.runtime.engines.store.db.sql"
)

// FsConfig maps runtime.engines.store: the root the file store is chrooted to, together with the modes it applies
// to the files and directories it creates.
type FsConfig struct {
	Root     string `json:"root"`
	FileMode string `json:"fileMode"`
	DirMode  string `json:"dirMode"`
}

// DbConfig maps one entry of runtime.engines.store.db.sql.
type DbConfig struct {
	Name       string   `json:"name"`
	Driver     string   `json:"driver"`
	Skipped    bool     `json:"skipped"`
	Dsn        string   `json:"dsn"`
	Options    []string `json:"options"`
	Migrations []string `json:"migrations"`
}

// FsConfigFor reads the file store settings.
//
// The root sits beside the fs object rather than inside it, so the three keys are read separately and assembled
// here. A missing value stays empty, leaving the engine to fall back to its own default.
func FsConfigFor(cfg domainContext.ConfigContext) FsConfig {
	return FsConfig{
		Root:     textAt(cfg, KeyFsRoot),
		FileMode: textAt(cfg, KeyFsFileMode),
		DirMode:  textAt(cfg, KeyFsDirMode),
	}
}

// DbConfigsFor reads every configured connection.
//
// Driver names this build does not link stay in the document and are dropped here with a warning, so a shared
// configuration file still runs on a binary that carries fewer drivers.
func DbConfigsFor(cfg domainContext.ConfigContext) []DbConfig {
	value, found := cfg.Lookup(KeyDbSQL)
	if !found {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	entries := make([]DbConfig, 0)
	if err := json.Unmarshal(encoded, &entries); err != nil {
		return nil
	}

	kept := entries[:0]
	for _, entry := range entries {
		if entry.Name == "" || entry.Dsn == "" {
			continue
		}
		if _, known := DriverOf(entry.Driver); !known {
			slog.Default().Warn("db driver skipped", "name", entry.Name, "driver", entry.Driver)
			continue
		}
		kept = append(kept, entry)
	}
	return kept
}

// textAt reads a text setting, reporting an empty string when it is absent or of another type.
func textAt(cfg domainContext.ConfigContext, key string) string {
	value, found := cfg.Lookup(key)
	if !found {
		return ""
	}
	text, _ := value.(string)
	return text
}
