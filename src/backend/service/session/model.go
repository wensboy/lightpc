package serviceSessionImpl

import "database/sql"

// The three dao structs below mirror the tables column for column.
//
// Every column of a row is present, including the ones the current flow does not read: the struct is the single
// description of a table, so a column added by a migration has exactly one place to land and a reader never has
// to cross reference the SQL to know what a row holds.
//
// A nullable column is held as sql.Null* rather than a pointer: the value comes from database/sql, and keeping
// the wrapper means the scan needs no conversion step.

// UserDao maps a row of users.
type UserDao struct {
	ID        int64         `json:"id" db:"id"`
	UID       string        `json:"uid" db:"uid"`
	Uname     string        `json:"uname" db:"uname"`
	Account   string        `json:"account" db:"account"`
	Password  string        `json:"-" db:"password"`
	CreatedAt int64         `json:"createdAt" db:"created_at"`
	UpdatedAt int64         `json:"updatedAt" db:"updated_at"`
	DeletedAt sql.NullInt64 `json:"deletedAt" db:"deleted_at"`
}

// UserProfileDao maps a row of user_profiles. UID is a logical reference to users.uid; there is no foreign key.
//
// The text columns fall back to an empty string, so a projection that only knows the uid can insert a row and
// let the rest stay at their default.
type UserProfileDao struct {
	ID        int64         `json:"id" db:"id"`
	UID       string        `json:"uid" db:"uid"`
	Email     string        `json:"email" db:"email"`
	Phone     string        `json:"phone" db:"phone"`
	Avatar    string        `json:"avatar" db:"avatar"`
	Nickname  string        `json:"nickname" db:"nickname"`
	Gender    string        `json:"gender" db:"gender"`
	Signature string        `json:"signature" db:"signature"`
	CreatedAt int64         `json:"createdAt" db:"created_at"`
	UpdatedAt int64         `json:"updatedAt" db:"updated_at"`
	DeletedAt sql.NullInt64 `json:"deletedAt" db:"deleted_at"`
}

// UserSessionDao maps a row of user_sessions. UID and Account are logical references to users.
//
// Expiration is the length of the window in seconds and is only meaningful when the action is sign_in; zero
// means the sign in never expires.
type UserSessionDao struct {
	ID         int64  `json:"id" db:"id"`
	UID        string `json:"uid" db:"uid"`
	Account    string `json:"account" db:"account"`
	Action     Action `json:"action" db:"action"`
	Expiration int64  `json:"expiration" db:"expiration"`
	CreatedAt  int64  `json:"createdAt" db:"created_at"`
}
