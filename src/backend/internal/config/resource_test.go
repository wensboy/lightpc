package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wensboy/lightpc/backend/internal/config"
)

func TestResourceLocalExpandsHome(t *testing.T) {
	resource, err := config.NewResource("local://~/.config/lightpc/logs/debug.log")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !resource.HasScheme() || !resource.IsLocal() {
		t.Fatalf("resource = %+v", resource)
	}

	resolved, err := resource.Path()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if strings.HasPrefix(resolved, "~") || !filepath.IsAbs(resolved) {
		t.Fatalf("path = %q", resolved)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(resolved, home) {
		t.Fatalf("path = %q, home = %q", resolved, home)
	}
}

func TestResourceEmbedDropsLeadingSlash(t *testing.T) {
	resource, err := config.NewResource("embed://data/conf/setting.default.json")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !resource.IsEmbed() {
		t.Fatalf("resource = %+v", resource)
	}
	resolved, err := resource.Path()
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	if resolved != "data/conf/setting.default.json" {
		t.Fatalf("path = %q", resolved)
	}
}

func TestResourceWithoutSchemeIsLocalRelative(t *testing.T) {
	resource, err := config.NewResource("logs/debug.log")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resource.HasScheme() || !resource.IsLocal() {
		t.Fatalf("resource = %+v", resource)
	}
	if resource.String() != "logs/debug.log" {
		t.Fatalf("String() = %q", resource.String())
	}
}

func TestResourceRejectsEmptyEmbedPath(t *testing.T) {
	resource, err := config.NewResource("embed:///")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := resource.Path(); err == nil {
		t.Fatal("an empty embed path must be rejected")
	}
}

func TestResourceDirKeepsTheScheme(t *testing.T) {
	cases := map[string]string{
		"local://~/.config/lightpc/setting.json": "local://~/.config/lightpc",
		"embed://data/conf/setting.default.json": "embed://data/conf",
		"logs/debug.log":                         "logs",
	}
	for value, want := range cases {
		resource, err := config.NewResource(value)
		if err != nil {
			t.Fatalf("%q: %v", value, err)
		}
		parent, found := resource.Dir()
		if !found {
			t.Fatalf("%q: no parent", value)
		}
		if parent.String() != want {
			t.Fatalf("Dir(%q) = %q, want %q", value, parent.String(), want)
		}
	}
}

func TestResourceJoinKeepsTheScheme(t *testing.T) {
	resource, err := config.NewResource("embed://data/docs")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	joined := resource.Join("beta/meta.json")
	if joined.String() != "embed://data/docs/beta/meta.json" {
		t.Fatalf("Join = %q", joined.String())
	}
	if joined.IsLocal() {
		t.Fatal("a joined embed resource must stay embedded")
	}
}

func TestResourceRelativeAnchorsAtRoot(t *testing.T) {
	inside := config.MustResource("/root/logs/a.log")
	if rel, ok := inside.Relative("/root"); !ok || rel != "logs/a.log" {
		t.Fatalf("rel = %q, ok = %v", rel, ok)
	}

	outside := config.MustResource("/other/a.log")
	if _, ok := outside.Relative("/root"); ok {
		t.Fatal("a path outside root must be rejected")
	}

	escaping := config.MustResource("/root/../etc/passwd")
	if _, ok := escaping.Relative("/root"); ok {
		t.Fatal("an escaping path must be rejected")
	}
}

func TestResourceDSNForSQLiteIsAFile(t *testing.T) {
	resource, err := config.NewResource("sqlite://~/.config/lightpc/db/dev.db")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	dsn, err := resource.DSN()
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	if !filepath.IsAbs(dsn) || strings.Contains(dsn, "~") {
		t.Fatalf("dsn = %q", dsn)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(dsn, home) {
		t.Fatalf("dsn = %q, home = %q", dsn, home)
	}
	if file, found := resource.DatabaseFile(); !found || file != dsn {
		t.Fatalf("FilePath = %q, %v", file, found)
	}
}

func TestResourceDSNForMySQLIsVerbatim(t *testing.T) {
	resource, err := config.NewResource("mysql://user:pass@tcp(127.0.0.1:3306)/lightpc")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	dsn, err := resource.DSN()
	if err != nil {
		t.Fatalf("dsn: %v", err)
	}
	if dsn != "user:pass@tcp(127.0.0.1:3306)/lightpc" {
		t.Fatalf("dsn = %q", dsn)
	}
	if _, found := resource.DatabaseFile(); found {
		t.Fatal("a mysql resource has no file")
	}
}

func TestResourceRejectsNonDatabaseScheme(t *testing.T) {
	resource, err := config.NewResource("embed://data/conf/default.json")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := resource.DSN(); err == nil {
		t.Fatal("a non database resource must not resolve to a DSN")
	}
}
