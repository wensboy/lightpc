package serviceHelp

// DocsGroup is one section of a documentation set, holding the documents it lists in the order the
// meta.json declares them.
type DocsGroup struct {
	Title string     `json:"title"`
	Items []DocsItem `json:"items"`
}

// DocsItem is one document of a group. Path is relative to the directory that carries the meta.json.
type DocsItem struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

// DocsVersion is one documentation set as the reader sees it.
type DocsVersion struct {
	Version string `json:"version"`
	// Index is the markdown shown before a document is chosen, relative to the meta.json directory.
	Index      string      `json:"index"`
	Deprecated bool        `json:"deprecated"`
	Tags       []string    `json:"tags"`
	Toc        []DocsGroup `json:"toc"`
}

// DocsPayload lists every documentation set that ships with the application.
type DocsPayload struct {
	Versions []DocsVersion `json:"versions"`
}

// HelpClient is the surface the front end consumes to browse the documentation.
type HelpClient interface {
	// Payload returns every documentation set, ordered by version.
	Payload() (DocsPayload, error)
	// Read returns the markdown of one document of the given version.
	Read(version, path string) (string, error)
}
