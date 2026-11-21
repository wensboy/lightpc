package engineStore

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
)

// Command marks every database command.
//
// Implementations identify commands with a type switch. SQL commands are provided here;
// non SQL stores may declare their own command types while keeping this contract intact.
type Command interface {
	command()
}

// Response marks every command response.
type Response interface {
	response()
}

// Dispatcher is the lowest database capability: it hands command messages to an implementation.
//
// Every data interaction goes through Dispatch, so the contract is unaware of SQL or any query language.
type Dispatcher interface {
	Dispatch(ctx context.Context, cmd Command) (Response, error)
}

// DBStore is the top level database contract.
//
// Lifecycle operations use plain methods while queries and mutations go through Dispatch.
type DBStore interface {
	Dispatcher
	// Begin starts a transaction and returns a Tx handle that is also a Dispatcher.
	Begin(ctx context.Context) (Tx, error)
	// Close releases the connection pool and is idempotent; every call reports ErrClosed afterwards.
	Close() error
}

// Tx is a transaction handle that is itself a Dispatcher.
type Tx interface {
	Dispatcher
	// Commit commits the transaction; every call reports ErrClosed afterwards.
	Commit() error
	// Rollback rolls the transaction back and is idempotent.
	Rollback() error
}

// QueryCommand is a SQL query command answered by QueryResponse.
type QueryCommand struct {
	SQL  string
	Args []any
}

func (QueryCommand) command() {}

// ExecCommand is a SQL execution command answered by ExecResponse.
type ExecCommand struct {
	SQL  string
	Args []any
}

func (ExecCommand) command() {}

// QueryResponse answers QueryCommand.
//
// Rows is a sqlx.Rows cursor: callers iterate with Next, Scan and Err, and must Close it to release resources.
type QueryResponse struct {
	Rows *sqlx.Rows
}

func (QueryResponse) response() {}

// ExecResponse answers ExecCommand.
type ExecResponse struct {
	Result sql.Result
}

func (ExecResponse) response() {}

// DbHookContext is what a database hook may inspect between the engine phases.
type DbHookContext struct {
	// EngineContext is the context the engine was initialised with.
	EngineContext domainEngine.EngineContext
}

// Driver is the database engine a connection speaks to.
//
// The list is wider than the drivers this build links: naming a driver is not the same as importing it, so a
// configuration may declare a database this binary cannot open. Such an entry is skipped at start up instead
// of failing the pipeline, which keeps one document usable across builds that ship different drivers.
type Driver string

const (
	DriverSQLite3    Driver = "sqlite3"
	DriverMySQL      Driver = "mysql"
	DriverPostgres   Driver = "postgres"
	DriverSQLServer  Driver = "sqlserver"
	DriverClickHouse Driver = "clickhouse"
	DriverCockroach  Driver = "cockroach"
	DriverTiDB       Driver = "tidb"
)

// drivers lists every name a configuration may use.
var drivers = map[Driver]struct{}{
	DriverSQLite3:    {},
	DriverMySQL:      {},
	DriverPostgres:   {},
	DriverSQLServer:  {},
	DriverClickHouse: {},
	DriverCockroach:  {},
	DriverTiDB:       {},
}

// Valid reports whether the name is a known driver.
func (d Driver) Valid() bool {
	_, found := drivers[d]
	return found
}

// DriverOf narrows a configured string, reporting false when the name is unknown.
func DriverOf(value string) (Driver, bool) {
	driver := Driver(value)
	return driver, driver.Valid()
}

// DBKey composes the lookup key of one connection. A driver may hold several databases, so the name is part
// of the key and "driver.name" is what callers pass to DB.
func DBKey(driver Driver, name string) string {
	return string(driver) + "." + name
}

// DBRegistry is the database group exposed by the store: connections are looked up by DBKey and no caller
// needs to know which driver a key resolves to.
type DBRegistry interface {
	// DB returns the store registered under key.
	DB(key string) (DBStore, bool)
	// DBKeys lists every registered key, ordered alphabetically.
	DBKeys() []string
}
