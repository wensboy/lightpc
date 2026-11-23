package serviceSessionImpl

import (
	"context"
	"database/sql"
	"time"

	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	serviceSession "github.com/wensboy/lightpc/backend/domain/service/session"
)

// Every method below is an unexported SessionService method: SQL never leaves this file, and a caller asks for
// an intent ("find the live account", "record a sign in") rather than for a statement.

// findAccount reads the live row of account, reporting nil when none exists.
//
// The soft deleted rows are filtered out here, so a removed account can register the same address again while
// its history stays intact.
func (s *SessionService) findAccount(store engineStore.DBStore, account string) (*UserDao, error) {
	ctx := context.Background()
	response, err := store.Dispatch(ctx, engineStore.QueryCommand{
		SQL: `SELECT id, uid, uname, account, password, created_at, updated_at, deleted_at
			FROM users WHERE account = ? AND deleted_at IS NULL LIMIT 1`,
		Args: []any{account},
	})
	if err != nil {
		return nil, err
	}

	query, ok := response.(engineStore.QueryResponse)
	if !ok || query.Rows == nil {
		return nil, nil
	}
	defer func() { _ = query.Rows.Close() }()

	// The statement carries LIMIT 1, so the cursor holds either one row or none. Reading it as "at most one"
	// keeps the exit paths explicit instead of hiding them inside a loop that only runs once.
	if !query.Rows.Next() {
		return nil, query.Rows.Err()
	}

	var row UserDao
	if err := query.Rows.Scan(&row.ID, &row.UID, &row.Uname, &row.Account, &row.Password,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt); err != nil {
		return nil, err
	}
	return &row, nil
}

// latestSignIn reads the most recent sign in row of one pair, which is what the restore path replays.
//
// The account is part of the condition: a pair that never signed in has no window, even when the uid is known.
func (s *SessionService) latestSignIn(store engineStore.DBStore, auth serviceSession.AuthMetaDto) (*UserSessionDao, error) {
	ctx := context.Background()
	response, err := store.Dispatch(ctx, engineStore.QueryCommand{
		SQL: `SELECT id, uid, account, action, expiration, created_at
			FROM user_sessions WHERE action = ? AND uid = ? AND account = ?
			ORDER BY created_at DESC, id DESC LIMIT 1`,
		Args: []any{string(ActionSignIn), auth.UID, auth.Account},
	})
	if err != nil {
		return nil, err
	}

	query, ok := response.(engineStore.QueryResponse)
	if !ok || query.Rows == nil {
		return nil, nil
	}
	defer func() { _ = query.Rows.Close() }()

	if !query.Rows.Next() {
		return nil, query.Rows.Err()
	}

	var row UserSessionDao
	var action sql.NullString
	if err := query.Rows.Scan(&row.ID, &row.UID, &row.Account, &action, &row.Expiration, &row.CreatedAt); err != nil {
		return nil, err
	}
	// A row without an action cannot be interpreted; the column is constrained, so this only guards a manual
	// edit of the file.
	if !action.Valid {
		return nil, nil
	}
	row.Action = Action(action.String)
	return &row, nil
}

// findAccountByUID reads the live row of uid, used when only the stored pair is available.
func (s *SessionService) findAccountByUID(store engineStore.DBStore, uid, account string) (*UserDao, error) {
	ctx := context.Background()
	response, err := store.Dispatch(ctx, engineStore.QueryCommand{
		SQL: `SELECT id, uid, uname, account, password, created_at, updated_at, deleted_at
			FROM users WHERE uid = ? AND account = ? AND deleted_at IS NULL LIMIT 1`,
		Args: []any{uid, account},
	})
	if err != nil {
		return nil, err
	}

	query, ok := response.(engineStore.QueryResponse)
	if !ok || query.Rows == nil {
		return nil, nil
	}
	defer func() { _ = query.Rows.Close() }()

	if !query.Rows.Next() {
		return nil, query.Rows.Err()
	}

	var row UserDao
	if err := query.Rows.Scan(&row.ID, &row.UID, &row.Uname, &row.Account, &row.Password,
		&row.CreatedAt, &row.UpdatedAt, &row.DeletedAt); err != nil {
		return nil, err
	}
	return &row, nil
}

// insertAccount writes the account, its profile row and the sign up history entry as one transaction.
//
// Only the link is filled in: the profile row carries the uid and the two timestamps, and every other column
// keeps its default until a feature needs it.
func (s *SessionService) insertAccount(store engineStore.DBStore, user UserDao, expiration int64) error {
	ctx := context.Background()
	now := time.Now().Unix()

	tx, err := store.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	statements := []engineStore.ExecCommand{
		{
			SQL: `INSERT INTO users (uid, uname, account, password, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?)`,
			Args: []any{user.UID, user.Uname, user.Account, user.Password, now, now},
		},
		{
			SQL:  `INSERT INTO user_profiles (uid, created_at, updated_at) VALUES (?, ?, ?)`,
			Args: []any{user.UID, now, now},
		},
		{
			SQL:  `INSERT INTO user_sessions (uid, account, action, expiration, created_at) VALUES (?, ?, ?, ?, ?)`,
			Args: []any{user.UID, user.Account, string(ActionSignUp), 0, now},
		},
	}
	for _, statement := range statements {
		if _, err := tx.Dispatch(ctx, statement); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// record appends one entry to the session history.
//
// The expiration is only meaningful for a sign in, and the table constrains it to a non negative integer, so the
// other actions store zero.
func (s *SessionService) record(store engineStore.DBStore, action Action, auth serviceSession.AuthMetaDto, expiration int64) error {
	if action != ActionSignIn {
		expiration = 0
	}

	ctx := context.Background()
	_, err := store.Dispatch(ctx, engineStore.ExecCommand{
		SQL:  `INSERT INTO user_sessions (uid, account, action, expiration, created_at) VALUES (?, ?, ?, ?, ?)`,
		Args: []any{auth.UID, auth.Account, string(action), expiration, now()},
	})
	return err
}

// removeAccount hides the account and its profile rows behind deleted_at, keeping every row readable.
func (s *SessionService) removeAccount(store engineStore.DBStore, uid string, at int64) error {
	ctx := context.Background()
	statements := []engineStore.ExecCommand{
		{SQL: `UPDATE users SET deleted_at = ?, updated_at = ? WHERE uid = ? AND deleted_at IS NULL`, Args: []any{at, at, uid}},
		{SQL: `UPDATE user_profiles SET deleted_at = ?, updated_at = ? WHERE uid = ? AND deleted_at IS NULL`, Args: []any{at, at, uid}},
	}
	for _, statement := range statements {
		if _, err := store.Dispatch(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

// deleteAccount removes the account and its profile rows for good.
//
// The two statements run in one transaction: a profile row left behind would keep referencing a uid that no
// longer exists, and there is no foreign key to clean it up.
func (s *SessionService) deleteAccount(store engineStore.DBStore, uid string) error {
	ctx := context.Background()

	tx, err := store.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	statements := []engineStore.ExecCommand{
		{SQL: `DELETE FROM user_profiles WHERE uid = ?`, Args: []any{uid}},
		{SQL: `DELETE FROM users WHERE uid = ?`, Args: []any{uid}},
	}
	for _, statement := range statements {
		if _, err := tx.Dispatch(ctx, statement); err != nil {
			return err
		}
	}
	return tx.Commit()
}
