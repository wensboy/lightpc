package store

import (
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

const (
	GroupStore = "store"

	NameFs = "fs"
	NameDb = "db"
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

var (
	StoreFs = domainEngine.EngineDep(GroupStore, NameFs)
	StoreDb = domainEngine.EngineDep(GroupStore, NameDb)
)
