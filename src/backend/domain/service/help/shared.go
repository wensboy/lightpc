package serviceHelp

import (
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// ErrKeyVersionUnknown reports a documentation version that no directory declares.
const ErrKeyVersionUnknown = "error::service.help::version_unknown"

// ErrKeyPathInvalid reports a document path that leaves its documentation set.
const ErrKeyPathInvalid = "error::service.help::path_invalid"

// Errors lists every error this service publishes.
var Errors = []interError.Option{
	{Key: ErrKeyVersionUnknown, Code: interError.ErrCodeNotExist, Msg: "help: unknown version"},
	{Key: ErrKeyPathInvalid, Code: interError.ErrCodeInvalidParam, Msg: "help: invalid document path"},
}
