package engineStore

import (
	"io/fs"
	"strconv"
	"strings"

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

// Config maps runtime.engines.store completely: the root the file store is chrooted to, the modes it applies, and
// every connection the database store opens.
type Config struct {
	Root string   `json:"root"`
	Fs   FsConfig `json:"fs"`
	Db   DbConfig `json:"db"`
}

// FsConfig maps runtime.engines.store.fs.
type FsConfig struct {
	FileMode string `json:"fileMode"`
	DirMode  string `json:"dirMode"`
}

// DbConfig maps runtime.engines.store.db.
type DbConfig struct {
	SQL []ConnConfig `json:"sql"`
}

// ConnConfig maps one entry of runtime.engines.store.db.sql.
//
// Options and Migrations are fixed lists, so they decode as themselves. Nothing here is narrowed by a runtime choice:
// the shape of a connection does not depend on another field of the same entry.
type ConnConfig struct {
	Name       string   `json:"name"`
	Driver     string   `json:"driver"`
	Skipped    bool     `json:"skipped"`
	Dsn        string   `json:"dsn"`
	Options    []string `json:"options"`
	Migrations []string `json:"migrations"`
}

// Usable reports whether a connection can be served at all: it needs a name, a DSN and a driver this build links.
//
// The check is a conversion of the config into a decision, not a read: nothing here touches the document, and the
// caller still decides what to do with an entry that fails.
func (c ConnConfig) Usable() bool {
	if c.Name == "" || c.Dsn == "" || c.Skipped {
		return false
	}
	_, known := DriverOf(c.Driver)
	return known
}

// ParseMode reads a mode written either as an octal literal ("0o644") or as the bare octal digits ("644").
//
// It is exported because the bootstrap is what reads the document, and the engine only receives the result.
func ParseMode(text string) fs.FileMode {
	if text == "" {
		return 0
	}
	mode, err := strconv.ParseUint(strings.TrimPrefix(text, "0o"), 8, 32)
	if err != nil {
		return 0
	}
	return fs.FileMode(mode)
}
