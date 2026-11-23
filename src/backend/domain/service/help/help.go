package serviceHelp

// HelpClient is the surface the front end consumes to browse the documentation.
type HelpClient interface {
	// Payload returns every documentation set, ordered by version.
	Payload() (DocsPayloadDto, error)
	// Read returns the markdown of one document of the given version.
	Read(version, path string) (string, error)
}
