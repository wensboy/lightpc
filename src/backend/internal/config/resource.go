package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Resource adapts a parsed url to the operations a caller needs: the scheme it resolves against, a path it can
// join or walk up from, and a text form that round trips back into the document.
//
// One schema, one adapter: every value the document writes is a resource, so there is a single type rather than a
// family of them. Anything that is not a file resource is still a resource here; the caller decides whether the
// scheme is one it can serve.
type Resource struct {
	u *url.URL
}

// NewResource parses value into a Resource.
func NewResource(value string) (Resource, error) {
	parsed, err := ParseValue(value)
	if err != nil {
		return Resource{}, err
	}
	return Resource{u: parsed}, nil
}

// MustResource parses value and panics on a malformed one. It is for constants the package itself declares.
func MustResource(value string) Resource {
	resource, err := NewResource(value)
	if err != nil {
		panic(err)
	}
	return resource
}

// Scheme reports the scheme, lower cased. An empty result means the value carried none.
func (r Resource) Scheme() string { return strings.ToLower(r.u.Scheme) }

// HasScheme reports whether the value named a scheme. It distinguishes a configured local:// value, which is
// anchored at the store root, from a plain relative store path.
func (r Resource) HasScheme() bool { return r.u.Scheme != "" }

// IsEmbed reports whether the resource lives in an embedded file system.
func (r Resource) IsEmbed() bool { return r.Scheme() == SchemeEmbed }

// IsLocal reports whether the resource is a file on disk. A value without a scheme counts as local.
func (r Resource) IsLocal() bool {
	scheme := r.Scheme()
	return scheme == "" || scheme == SchemeLocal
}

// Path returns the path the caller acts on, resolved according to the scheme: a local target is expanded and made
// absolute, an embedded one is slash separated without a leading slash, which is what fs.FS expects.
func (r Resource) Path() (string, error) {
	switch {
	case r.IsEmbed():
		body := strings.TrimPrefix(r.u.Path, "/")
		if body == "" {
			return "", fmt.Errorf("config: %q carries an empty embed path", r.String())
		}
		return body, nil
	case r.IsLocal():
		// A value written "local://~/x" folds to "/~/x"; that slash belongs to the fold rather than to the path,
		// and leaving it would stop the home expansion from recognising the tilde.
		body := strings.TrimPrefix(r.u.Path, "/~")
		if body != r.u.Path {
			return localPath("~" + body)
		}
		return localPath(r.u.Path)
	default:
		return "", fmt.Errorf("config: unknown scheme %q in %q", r.u.Scheme, r.String())
	}
}

// String restores the text form the document uses, so a resource can be handed back to a caller unchanged.
//
// The path keeps a leading slash once parsed, while the document writes "<scheme>://<body>"; dropping that slash
// is what keeps the two forms identical instead of producing "local:///~".
func (r Resource) String() string {
	if !r.HasScheme() {
		return r.u.Path
	}
	return r.u.Scheme + "://" + strings.TrimPrefix(r.u.Path, "/")
}

// Dir returns the parent as a resource, keeping the scheme. The second result is false when the value has no
// parent, so a caller can skip the directory step instead of creating one at the root.
func (r Resource) Dir() (Resource, bool) {
	parent := filepath.Dir(r.u.Path)
	if parent == "." || parent == "/" {
		return Resource{}, false
	}
	child := *r.u
	child.Path = parent
	return Resource{u: &child}, true
}

// Join appends a relative path, keeping the scheme. path.Join is not used because it cleans the value and would
// fold "embed://" into "embed:/".
func (r Resource) Join(relative string) Resource {
	child := *r.u
	child.Path = strings.TrimSuffix(r.u.Path, "/") + "/" + strings.TrimPrefix(relative, "/")
	return Resource{u: &child}
}

// Relative expresses r as a path below root, which is a plain absolute directory rather than a resource: the
// store root is always on disk. The second result is false when r escapes root, which is what keeps a hand
// edited document from reaching outside the store.
func (r Resource) Relative(root string) (string, bool) {
	resolved, err := r.Path()
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(filepath.Clean(root), filepath.Clean(resolved))
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return rel, true
}

// DSN returns the data source name the driver expects, with the scheme removed.
//
// Each scheme decides what its own DSN means: a sqlite value names a file, so it is expanded and made absolute,
// while a mysql value is handed on as written because the driver parses it itself. The caller only sees a string,
// which is what keeps the migration runner free of driver specific introspection.
func (r Resource) DSN() (string, error) {
	body := strings.TrimPrefix(r.u.Path, "/")

	switch r.Scheme() {
	case SchemeSQLite:
		// "sqlite://~/x" folds to "/~/x"; that slash belongs to the fold rather than to the path.
		if strings.HasPrefix(body, "~") {
			return localPath(body)
		}
		return localPath("/" + body)
	case SchemeMySQL:
		return body, nil
	default:
		return "", fmt.Errorf("config: %q is not a database resource", r.String())
	}
}

// DatabaseFile reports the file a sqlite database lives at, and false for any other scheme: a remote database has
// no file for the caller to create a directory for.
func (r Resource) DatabaseFile() (string, bool) {
	if r.Scheme() != SchemeSQLite {
		return "", false
	}
	file, err := r.DSN()
	if err != nil {
		return "", false
	}
	return file, true
}

// localPath expands a leading ~ and makes the result absolute.
func localPath(value string) (string, error) {
	if strings.HasPrefix(value, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("config: resolve home for %q: %w", value, err)
		}
		value = filepath.Join(home, strings.TrimPrefix(value, "~"))
	}
	abs, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("config: absolute for %q: %w", value, err)
	}
	return abs, nil
}
