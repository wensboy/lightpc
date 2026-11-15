package store

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
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
