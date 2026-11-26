package engineStoreImpl

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	"github.com/wensboy/lightpc/backend/internal/config"
)

// dbConfig is one entry of runtime.store.db.sql. Options are appended to the DSN as query parameters, which
// is how the sqlite driver receives its pragmas.
type DBConfig struct {
	Name       string             `json:"name"`
	Driver     engineStore.Driver `json:"driver"`
	Skipped    bool               `json:"skipped"`
	DSN        string             `json:"dsn"`
	Options    []string           `json:"options"`
	Migrations []string           `json:"migrations"`
}

// DBConfigs turns the configured entries into the shape the runner works with, dropping the ones it cannot serve.
//
// The document is read by the contract package, so this file only decides what a usable entry looks like: a name,
// a driver this build links, and a DSN the adapter can resolve.
func DBConfigs(entries []engineStore.ConnConfig) []DBConfig {
	configs := make([]DBConfig, 0, len(entries))
	for _, entry := range entries {
		if !entry.Usable() {
			continue
		}
		driver, _ := engineStore.DriverOf(entry.Driver)
		configs = append(configs, DBConfig{
			Name:       entry.Name,
			Driver:     driver,
			DSN:        entry.Dsn,
			Options:    entry.Options,
			Migrations: entry.Migrations,
		})
	}
	return configs
}

// DSNString turns the configured resource into the text the driver expects and appends the options.
//
// What the value means is the scheme's business, not this file's: the adapter resolves "sqlite://~/db" to a file
// and hands a mysql value on as written. This function only joins the options, which is the one step every driver
// shares.
func (c DBConfig) DSNString() (string, error) {
	resource, err := c.resource()
	if err != nil {
		return "", err
	}
	dsn, err := resource.DSN()
	if err != nil {
		return "", err
	}
	if len(c.Options) == 0 {
		return dsn, nil
	}

	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}
	return dsn + separator + strings.Join(c.Options, "&"), nil
}

// DatabaseFile reports the file the database lives at, and whether there is one at all.
func (c DBConfig) DatabaseFile() (string, bool) {
	resource, err := c.resource()
	if err != nil {
		return "", false
	}
	return resource.DatabaseFile()
}

// resource resolves the configured dsn once, so both accessors read the same value the same way.
func (c DBConfig) resource() (config.Resource, error) {
	resource, err := config.NewResource(c.DSN)
	if err != nil {
		return config.Resource{}, fmt.Errorf("db %q: %w", c.Name, err)
	}
	return resource, nil
}

// OpenDB opens the pool for one entry. The parent directory is created because a sqlite database is a file
// the first launch has to be able to write.
func OpenDB(cfg DBConfig) (*sqlx.DB, error) {
	if file, found := cfg.DatabaseFile(); found {
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			return nil, fmt.Errorf("create db directory: %w", err)
		}
	}

	dsn, err := cfg.DSNString()
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Connect(string(cfg.Driver), dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", cfg.Name, err)
	}
	return db, nil
}

// Migrate applies every migration entry in order against the open pool.
//
// An entry naming a directory applies the ordered files inside it; an entry naming a single file applies only
// that file. Both forms live under the embedded tree, so a shipped migration never needs the filesystem.
func Migrate(dsn string, driver engineStore.Driver, migrations []string, assets fs.FS) error {
	for _, entry := range migrations {
		if entry == "" {
			continue
		}
		if err := migrateEntry(dsn, driver, entry, assets); err != nil {
			return fmt.Errorf("migrate %s: %w", entry, err)
		}
	}
	return nil
}

// migrateEntry mounts one location as a migration source and applies it. A directory is mounted as is; a
// single file is mounted through a view that exposes only that file, so both take the same runner.
func migrateEntry(dsn string, driver engineStore.Driver, entry string, assets fs.FS) error {
	relative, ok := migrationPath(entry)
	if !ok {
		return fmt.Errorf("migration must live under embed://, got %q", entry)
	}

	root, err := migrationRoot(assets, relative)
	if err != nil {
		return fmt.Errorf("open migration tree: %w", err)
	}
	if root == nil {
		// The directory carries no migration files, so there is nothing to apply.
		return nil
	}

	source, err := iofs.New(root, ".")
	if err != nil {
		return fmt.Errorf("read migration files: %w", err)
	}
	defer source.Close()

	instance, err := migrate.NewWithSourceInstance("iofs", source, databaseURL(dsn, driver))
	if err != nil {
		return fmt.Errorf("prepare migrator: %w", err)
	}
	defer instance.Close()

	if err := instance.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// migrationRoot returns the filesystem a migration entry points at. A directory is returned as a subtree, a
// single file is wrapped in a view that lists just that file: iofs reads a directory listing, so a bare file
// cannot be handed over directly.
func migrationRoot(assets fs.FS, relative string) (fs.FS, error) {
	info, err := fs.Stat(assets, relative)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("migration entry must be a directory, got %q", relative)
	}

	// An empty directory is a valid state: the schema is shipped but no migration has landed yet. Skipping it
	// keeps the runner from comparing an empty source against the version a previous build already recorded.
	entries, err := fs.ReadDir(assets, relative)
	if err != nil {
		return nil, err
	}
	hasMigration := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sql") {
			hasMigration = true
			break
		}
	}
	if !hasMigration {
		return nil, nil
	}
	return fs.Sub(assets, relative)
}

// databaseURL composes the URL the migrate driver expects: the scheme names the driver and the rest is the DSN the
// connection already uses, so the runner needs no driver specific introspection.
//
// The concatenation stays here rather than going through the adapter: the adapter trims the leading slash a document
// value folds away, and a sqlite DSN starts with one, so a round trip would turn "sqlite3:///tmp/x" into
// "sqlite3://tmp/x" and name a host instead of a file.
func databaseURL(dsn string, driver engineStore.Driver) string {
	return string(driver) + "://" + dsn
}

// migrationPath resolves a migration entry to the path inside the embedded tree.
//
// The value is read the same way as every other resource, so this file carries no scheme of its own: a target that is
// not embedded is refused, and an embedded one yields the relative path fs.Sub expects.
func migrationPath(entry string) (string, bool) {
	resource, err := config.NewResource(entry)
	if err != nil || !resource.IsEmbed() {
		return "", false
	}
	relative, err := resource.Path()
	if err != nil {
		return "", false
	}
	return relative, true
}
