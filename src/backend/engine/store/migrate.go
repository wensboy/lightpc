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
	"github.com/wensboy/lightpc/backend/internal/cast"
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
func DBConfigs(entries []engineStore.DbConfig) []DBConfig {
	configs := make([]DBConfig, 0, len(entries))
	for _, entry := range entries {
		driver, known := engineStore.DriverOf(entry.Driver)
		if !known || entry.Skipped {
			continue
		}
		configs = append(configs, DBConfig{
			Name:       entry.Name,
			Driver:     driver,
			Skipped:    entry.Skipped,
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
	resource, err := config.NewResource(c.DSN)
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
	resource, err := config.NewResource(c.DSN)
	if err != nil {
		return "", false
	}
	return resource.DatabaseFile()
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
	relative, ok := embeddedRelative(entry)
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

// databaseURL composes the URL the migrate driver expects: the scheme names the driver and the rest is the
// DSN the connection already uses, so the runner needs no driver specific introspection.
func databaseURL(dsn string, driver engineStore.Driver) string {
	return string(driver) + "://" + dsn
}

// embeddedRelative turns an embedded resource into a path inside the embedded tree.
func embeddedRelative(resource string) (string, bool) {
	const prefix = "embed://"
	if !strings.HasPrefix(resource, prefix) {
		return "", false
	}
	return strings.TrimPrefix(resource, prefix), true
}

// expandHome resolves a leading tilde, which the document uses to point into the user directory.
func expandHome(value string) string {
	if !strings.HasPrefix(value, "~") {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return value
	}
	return filepath.Join(home, strings.TrimPrefix(value, "~"))
}

// stringField reads a string member of a decoded record.
func stringField(fields map[string]any, key string) string {
	value, _ := cast.To[string](fields[key])
	return value
}

// boolField reads a boolean member of a decoded record.
func boolField(fields map[string]any, key string) bool {
	value, _ := cast.To[bool](fields[key])
	return value
}

// optionList reads the "key=value" options of a decoded record, keeping the declared order.
func optionList(raw any) []string {
	list, ok := cast.To[[]any](raw)
	if !ok {
		return nil
	}
	options := make([]string, 0, len(list))
	for _, entry := range list {
		if text, ok := cast.To[string](entry); ok && text != "" {
			options = append(options, text)
		}
	}
	return options
}
