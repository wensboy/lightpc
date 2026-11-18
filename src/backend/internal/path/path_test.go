package path_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wensboy/lightpc/backend/internal/path"
)

func TestParseLocalExpandsHome(t *testing.T) {
	target, err := path.Parse("local://~/.config/lightpc/logs/debug.log")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if target.Proto != path.ProtoLocal || !target.Explicit {
		t.Fatalf("target = %+v", target)
	}
	if strings.HasPrefix(target.Path, "~") || !filepath.IsAbs(target.Path) {
		t.Fatalf("path = %q", target.Path)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(target.Path, home) {
		t.Fatalf("path = %q, home = %q", target.Path, home)
	}
}

func TestParseEmbedKeepsRelativePath(t *testing.T) {
	target, err := path.Parse("embed://data/conf/setting.default.json")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !target.IsEmbed() || target.Path != "data/conf/setting.default.json" {
		t.Fatalf("target = %+v", target)
	}
}

func TestParseWithoutProtoDefaultsToLocal(t *testing.T) {
	target, err := path.Parse("logs/debug.log")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if target.Explicit || target.Proto != path.ProtoLocal {
		t.Fatalf("target = %+v", target)
	}
	if !filepath.IsAbs(target.Path) {
		t.Fatalf("path = %q", target.Path)
	}
}

func TestParseRejectsUnknownProto(t *testing.T) {
	if _, err := path.Parse("http://example.com/a"); !errors.Is(err, path.ErrUnsupported) {
		t.Fatalf("err = %v", err)
	}
}

func TestParseRejectsEmptyEmbedPath(t *testing.T) {
	if _, err := path.Parse("embed:///"); !errors.Is(err, path.ErrUnsupported) {
		t.Fatalf("err = %v", err)
	}
}

func TestDirKeepsProto(t *testing.T) {
	cases := map[string]string{
		"local://~/.config/lightpc/setting.json": "local://~/.config/lightpc",
		"embed://data/conf/setting.default.json": "embed://data/conf",
		"logs/debug.log":                         "logs",
	}
	for value, want := range cases {
		if got := path.Dir(value); got != want {
			t.Fatalf("Dir(%q) = %q, want %q", value, got, want)
		}
	}
}

func TestRelativeAnchorsAtRoot(t *testing.T) {
	rel, ok := path.Relative("/root", "/root/logs/a.log")
	if !ok || rel != "logs/a.log" {
		t.Fatalf("rel = %q, ok = %v", rel, ok)
	}
	if _, ok := path.Relative("/root", "/other/a.log"); ok {
		t.Fatal("a path outside root must be rejected")
	}
	if _, ok := path.Relative("/root", "/root/../etc/passwd"); ok {
		t.Fatal("an escaping path must be rejected")
	}
}
