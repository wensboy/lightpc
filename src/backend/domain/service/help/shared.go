package serviceHelp

import (
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

const (
	// ErrKeyVersionUnknown reports a documentation version that no directory declares.
	ErrKeyVersionUnknown = "error::service.help::version_unknown"

	// ErrKeyPathInvalid reports a document path that leaves its documentation set.
	ErrKeyPathInvalid = "error::service.help::path_invalid"
)

// Errors lists every error this service publishes.
var Errors = []interError.Option{
	{Key: ErrKeyVersionUnknown, Code: interError.ErrCodeNotExist, Msg: "help: unknown version"},
	{Key: ErrKeyPathInvalid, Code: interError.ErrCodeInvalidParam, Msg: "help: invalid document path"},
}

// DocsGroupDto is one section of a documentation set, holding the documents it lists in the order the
// meta.json declares them.
type DocsGroupDto struct {
	Title string        `json:"title"`
	Items []DocsItemDto `json:"items"`
}

// DocsItemDto is one document of a group. Path is relative to the directory that carries the meta.json.
type DocsItemDto struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

// DocsVersionDto is one documentation set as the reader sees it.
type DocsVersionDto struct {
	Version string `json:"version"`
	// Index is the markdown shown before a document is chosen, relative to the meta.json directory.
	Index      string         `json:"index"`
	Deprecated bool           `json:"deprecated"`
	Tags       []string       `json:"tags"`
	Toc        []DocsGroupDto `json:"toc"`
}

// DocsPayloadDto lists every documentation set that ships with the application.
type DocsPayloadDto struct {
	Versions []DocsVersionDto `json:"versions"`
}
