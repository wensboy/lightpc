// Package path parses the proto://<source-path> values used by configuration for file resources.
//
// It is pure and stateless: the same value always resolves to the same target, no matter which engine
// consumes it. Local paths expand a leading ~ and become absolute; embedded paths stay slash separated
// without a leading slash.
package path

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Supported schemes of a resource path.
const (
	ProtoLocal = "local"
	ProtoEmbed = "embed"
)

// schemeSeparator splits the proto from the rest of the value.
const schemeSeparator = "://"

// ErrUnsupported reports a value whose scheme or body cannot be resolved.
var ErrUnsupported = errors.New("path: unsupported value")

// Target is the parsed form of a resource path.
type Target struct {
	// Proto is the scheme; values without a scheme resolve to ProtoLocal.
	Proto string
	// Path is absolute for local targets, and slash separated without a leading slash for embed ones.
	Path string
	// Explicit reports whether the value carried an explicit proto. It distinguishes a configured
	// local:// path, which is anchored at the store root, from a plain relative store path.
	Explicit bool
}

// IsEmbed reports whether the target lives in an embedded file system.
func (t Target) IsEmbed() bool { return t.Proto == ProtoEmbed }

// Parse resolves value into a Target.
func Parse(value string) (Target, error) {
	target := Target{Proto: ProtoLocal, Path: value}

	proto, rest, found := strings.Cut(value, schemeSeparator)
	if found {
		target.Proto = strings.ToLower(proto)
		target.Path = rest
		target.Explicit = true
	}

	switch target.Proto {
	case ProtoLocal:
		resolved, err := localPath(target.Path)
		if err != nil {
			return Target{}, err
		}
		target.Path = resolved
	case ProtoEmbed:
		target.Path = strings.TrimPrefix(target.Path, "/")
		if target.Path == "" {
			return Target{}, fmt.Errorf("%w: empty embed path in %q", ErrUnsupported, value)
		}
	default:
		return Target{}, fmt.Errorf("%w: unknown proto %q", ErrUnsupported, proto)
	}
	return target, nil
}

// Join appends a relative path to a resource while preserving the proto separator. path.Join is not used
// because it cleans the value and would fold "embed://" into "embed:/".
func Join(resource, relative string) string {
	return strings.TrimSuffix(resource, "/") + "/" + strings.TrimPrefix(relative, "/")
}

// Dir returns the parent of a resource value, keeping its proto. It returns an empty string when the
// value has no parent, so callers can skip the mkdir step.
func Dir(value string) string {
	proto, rest, found := strings.Cut(value, schemeSeparator)
	if !found {
		parent := filepath.Dir(value)
		if parent == "." || parent == string(filepath.Separator) {
			return ""
		}
		return parent
	}
	parent := filepath.Dir(rest)
	if parent == "." || parent == string(filepath.Separator) {
		return ""
	}
	return proto + schemeSeparator + parent
}

// Relative returns target relative to root and reports false when target escapes root.
func Relative(root, target string) (string, bool) {
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return rel, true
}

// localPath expands a leading ~ and makes the result absolute.
func localPath(value string) (string, error) {
	if strings.HasPrefix(value, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("path: resolve home for %q: %w", value, err)
		}
		value = filepath.Join(home, strings.TrimPrefix(value, "~"))
	}
	abs, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("path: absolute for %q: %w", value, err)
	}
	return abs, nil
}
