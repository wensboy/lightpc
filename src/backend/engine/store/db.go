package engineStoreImpl

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

var (
	_ engineStore.DBStore = (*SQL)(nil)
	_ domainEngine.Engine = (*SQL)(nil)
)

// SQLOption is a critical construction parameter for SQL.
type SQLOption func(*SQL)

// WithDB sets the underlying sqlx connection pool.
func WithDB(db *sqlx.DB) SQLOption {
	return func(s *SQL) { s.db = db }
}

// SQL implements engineStore.DBStore on top of sqlx.
type SQL struct {
	db           *sqlx.DB
	queryTimeout time.Duration

	mu     sync.RWMutex
	ectx   domainEngine.EngineContext
	state  domainEngine.EngineState
	closed bool
}

// sqlTx adapts *sqlx.Tx to engineStore.Tx.
type sqlTx struct {
	tx *sqlx.Tx
}

// NewSQL builds a SQL instance. Callers must call Validate before concurrent use.
func NewSQL(opts ...SQLOption) *SQL {
	s := &SQL{state: domainEngine.EnginePending}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithQueryTimeout sets the timeout applied to every Dispatch and returns *SQL for chaining.
// It must be called before concurrent use.
func (s *SQL) WithQueryTimeout(d time.Duration) *SQL {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queryTimeout = d
	return s
}

// Group implements domainEngine.Engine.
func (s *SQL) Group() string { return domainEngine.GroupStore }

// Name implements domainEngine.Engine.
func (s *SQL) Name() string { return domainEngine.NameDb }

// Deps implements domainEngine.Engine.
func (s *SQL) Deps() []string { return []string{} }

// State implements domainEngine.Engine.
func (s *SQL) State() domainEngine.EngineState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Init implements domainEngine.Engine: it keeps the engine context, publishes the store errors,
// validates the pool and watches signals.
func (s *SQL) Init(eCtx domainEngine.EngineContext) error {
	s.setState(domainEngine.EnginePending)
	interError.RegisterAll(eCtx.AppCtx, engineStore.Errors...)

	s.mu.Lock()
	s.ectx = eCtx
	s.mu.Unlock()

	if err := s.Validate(); err != nil {
		s.setState(domainEngine.EngineError)
		return err
	}

	watchStopSignal(eCtx.Ebus, domainEngine.NameDb, s.Close)
	s.setState(domainEngine.EngineReady)
	return nil
}

// Start implements domainEngine.Engine. SQL serves requests synchronously, so it only moves the state.
func (s *SQL) Start() error {
	s.setState(domainEngine.EngineRunning)
	return nil
}

// Restart implements domainEngine.Engine by reopening the pool with the retained context and validating it again.
func (s *SQL) Restart() error {
	s.mu.Lock()
	s.closed = false
	eCtx := s.ectx
	s.mu.Unlock()
	return s.Init(eCtx)
}

// Stop implements domainEngine.Engine by closing the engineStore.
func (s *SQL) Stop() error { return s.Close() }

// Validate checks that the connection pool is configured. It must be called before concurrent use.
func (s *SQL) Validate() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(errors.New("db not configured"))
	}
	return nil
}

// Dispatch implements engineStore.Dispatcher.
func (s *SQL) Dispatch(ctx context.Context, cmd engineStore.Command) (engineStore.Response, error) {
	db, timeout, err := s.snapshot()
	if err != nil {
		return nil, err
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	switch c := cmd.(type) {
	case engineStore.QueryCommand:
		rows, err := db.QueryxContext(ctx, c.SQL, c.Args...)
		if err != nil {
			return nil, err
		}
		return engineStore.QueryResponse{Rows: rows}, nil
	case engineStore.ExecCommand:
		res, err := db.ExecContext(ctx, c.SQL, c.Args...)
		if err != nil {
			return nil, err
		}
		return engineStore.ExecResponse{Result: res}, nil
	default:
		return nil, interError.Retrieve(engineStore.ErrKeyUnsupportedCommand).Wrap(fmt.Errorf("unsupported command %T", cmd))
	}
}

// Begin implements engineStore.DBStore.Begin.
func (s *SQL) Begin(ctx context.Context) (engineStore.Tx, error) {
	db, _, err := s.snapshot()
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqlTx{tx: tx}, nil
}

// Close implements engineStore.DBStore.Close and is idempotent.
func (s *SQL) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	db := s.db
	s.mu.Unlock()

	if db == nil {
		s.setState(domainEngine.EngineStopped)
		return nil
	}
	if err := db.Close(); err != nil {
		s.setState(domainEngine.EngineError)
		return err
	}
	s.setState(domainEngine.EngineStopped)
	return nil
}

// Dispatch implements engineStore.Dispatcher with the same semantics as SQL.Dispatch.
func (t *sqlTx) Dispatch(ctx context.Context, cmd engineStore.Command) (engineStore.Response, error) {
	switch c := cmd.(type) {
	case engineStore.QueryCommand:
		rows, err := t.tx.QueryxContext(ctx, c.SQL, c.Args...)
		if err != nil {
			return nil, err
		}
		return engineStore.QueryResponse{Rows: rows}, nil
	case engineStore.ExecCommand:
		res, err := t.tx.ExecContext(ctx, c.SQL, c.Args...)
		if err != nil {
			return nil, err
		}
		return engineStore.ExecResponse{Result: res}, nil
	default:
		return nil, interError.Retrieve(engineStore.ErrKeyUnsupportedCommand).Wrap(fmt.Errorf("unsupported command %T", cmd))
	}
}

// Commit implements engineStore.Tx.Commit.
func (t *sqlTx) Commit() error { return t.tx.Commit() }

// Rollback implements engineStore.Tx.Rollback.
func (t *sqlTx) Rollback() error { return t.tx.Rollback() }

// setState stores the engine state.
func (s *SQL) setState(state domainEngine.EngineState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
}

// snapshot snapshots the pool and the timeout, and reports an error when SQL is invalid or closed.
func (s *SQL) snapshot() (*sqlx.DB, time.Duration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return nil, 0, interError.Retrieve(engineStore.ErrKeyClosed)
	}
	if s.db == nil {
		return nil, 0, interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(errors.New("db not configured"))
	}
	return s.db, s.queryTimeout, nil
}
