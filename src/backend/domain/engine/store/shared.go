package store

import interError "github.com/wensboy/lightpc/backend/internal/error"

const (
	// GroupStore is the engine group shared by every store implementation.
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
	{Key: ErrKeyUnsupportedCommand, Code: interError.ErrCodeInvalidParam, Msg: "store: unsupported command"},
}
