package engineStoreImpl

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	"github.com/wensboy/lightpc/backend/internal/cast"
)

// dbConfig is one entry of runtime.store.db.sql. Options are appended to the DSN as query parameters, which
// is how the sqlite driver receives its pragmas.
type dbConfig struct {
	Name       string             `json:"name"`
	Driver     engineStore.Driver `json:"driver"`
	Skipped    bool               `json:"skipped"`
	DSN        string             `json:"dsn"`
	Options    []string           `json:"options"`
	Migrations []string           `json:"migrations"`
}

// DBConfigs reads every database entry the configuration declares.
//
// Skipped keeps an entry in the document without opening it: the name stays visible to the settings view
// while no connection is made and no migration runs.
func DBConfigs(cfg domainContext.ConfigContext) []dbConfig {
	value, found := cfg.Lookup("default.runtime.engines.store.db.sql")
	if !found {
		return nil
	}
	records, ok := cast.To[[]any](value)
	if !ok {
		return nil
	}

	configs := make([]dbConfig, 0, len(records))
	for _, record := range records {
		fields, ok := cast.To[map[string]any](record)
		if !ok {
			continue
		}
		driver, known := engineStore.DriverOf(stringField(fields, "driver"))
		if !known {
			// Naming a driver this build does not link is a configuration mistake, not a fatal one.
			slog.Default().Warn("db driver skipped", "name", stringField(fields, "name"), "driver", stringField(fields, "driver"))
			continue
		}
		config := dbConfig{
			Name:    stringField(fields, "name"),
			Driver:  driver,
			Skipped: boolField(fields, "skipped"),
			DSN:     stringField(fields, "dsn"),
			Options: optionList(fields["options"]),
		}
		if list, ok := cast.To[[]any](fields["migrations"]); ok {
			for _, entry := range list {
				if text, ok := cast.To[string](entry); ok && text != "" {
					config.Migrations = append(config.Migrations, text)
				}
			}
		}
		if config.Skipped || config.Name == "" || config.DSN == "" {
			continue
		}
		configs = append(configs, config)
	}
	return configs
}

// DSNString expands the home directory and appends the options as query parameters.
func (c dbConfig) DSNString() string {
	file := expandHome(c.DSN)
	if len(c.Options) == 0 {
		return file
	}

	separator := "?"
	if strings.Contains(file, "?") {
		separator = "&"
	}
	return file + separator + strings.Join(c.Options, "&")
}

// filePath is the database file the DSN points at, without the query part.
func (c dbConfig) filePath() string {
	file := expandHome(c.DSN)
	if before, _, ok := strings.Cut(file, "?"); ok {
		return before
	}
	return file
}

// OpenDB opens the pool for one entry. The parent directory is created because a sqlite database is a file
// the first launch has to be able to write.
func OpenDB(cfg dbConfig) (*sqlx.DB, error) {
	file := cfg.filePath()
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sqlx.Connect(string(cfg.Driver), cfg.DSNString())
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

// AssetsFrom returns the embedded tree the migrations live under. The entry point mounts the data directory,
// so every migration resource is a path below it.
func AssetsFrom(appCtx *domainContext.AppContext) (fs.FS, bool) {
	if appCtx == nil {
		return nil, false
	}
	tree, found := appCtx.EmbedContext["/data"]
	if !found || tree == nil {
		return nil, false
	}
	return tree, true
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
