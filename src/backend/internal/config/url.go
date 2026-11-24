package config

import (
	"fmt"
	"net/url"
	"strings"
)

// Schemes a resource value may carry. A value without one is a local path relative to the store root.
const (
	SchemeLocal  = "local"
	SchemeEmbed  = "embed"
	SchemeSQLite = "sqlite"
	SchemeMySQL  = "mysql"
)

// ParseValue resolves a resource value into a standard url.
//
// net/url owns the parsing, so a value is read the way RFC 3986 describes it rather than by a private rule set:
// the document is the convention, this function only follows it.
//
// The document always writes "<scheme>://<path>", while authority syntax would turn the first path segment into
// a host. The separator is therefore folded to ":/" before parsing, which keeps the whole path in Path for both
// schemes and leaves a value without a scheme untouched.
func ParseValue(value string) (*url.URL, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, fmt.Errorf("config: empty resource value")
	}

	if scheme, rest, found := strings.Cut(trimmed, "://"); found {
		// An absolute body already starts with a slash, so it is reattached as is; otherwise a slash is put back.
		// Writing ":/" unconditionally would turn "local:///tmp/x" into "local://tmp/x", and net/url would read
		// "tmp" as a host.
		if strings.HasPrefix(rest, "/") {
			trimmed = scheme + ":" + rest
		} else {
			trimmed = scheme + ":/" + rest
		}
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("config: %q is not a resource: %w", value, err)
	}
	return parsed, nil
}
