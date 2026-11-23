package serviceSession

import (
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

const (
	// ErrKeyDatabaseMissing reports that no database is reachable or the configured key is unknown.
	ErrKeyDatabaseMissing = "error::service.session::database_missing"
	// ErrKeyAccountExists reports that the account is already taken by a live row.
	ErrKeyAccountExists = "error::service.session::account_exists"
	// ErrKeyCredentials reports that the pair account and password does not match.
	ErrKeyCredentials = "error::service.session::credentials"
	// ErrKeyNotAuthenticated reports a call that needs a session but none is live.
	ErrKeyNotAuthenticated = "error::service.session::not_authenticated"
)

// Errors lists every error this service publishes.
//
// The list lives with the contract rather than with the implementation: a caller that compares against a key
// should not have to import the implementation to spell it.
var Errors = []interError.Option{
	{Key: ErrKeyDatabaseMissing, Code: interError.ErrCodeNotExist, Msg: "session: database unavailable"},
	{Key: ErrKeyAccountExists, Code: interError.ErrCodeInvalidParam, Msg: "session: account already registered"},
	{Key: ErrKeyCredentials, Code: interError.ErrCodeInvalidParam, Msg: "session: invalid credentials"},
	{Key: ErrKeyNotAuthenticated, Code: interError.ErrCodeNotExist, Msg: "session: not signed in"},
}

// AuthMetaDto is what the front end sends back on every call that acts on the current session.
//
// It carries no secret: the pair identifies which rows to look at, and the backend decides from the stored
// history whether that pair is still inside its window. A caller that loses the pair simply signs in again.
type AuthMetaDto struct {
	UID     string `json:"uid"`
	Account string `json:"account"`
}

// StatusDto is the signed in state of the current visitor.
//
// Authenticated is only true when both the cache and the stored session row agree, so a pair left behind by a
// restart can never report a signed in state on its own. ExpiresAt is a unix second; zero means the session has
// no deadline.
type StatusDto struct {
	Authenticated bool   `json:"authenticated"`
	UID           string `json:"uid"`
	Account       string `json:"account"`
	Uname         string `json:"uname"`
	ExpiresAt     int64  `json:"expiresAt"`
}

// RegisterInputDto carries what a new account needs.
type RegisterInputDto struct {
	Uname    string `json:"uname"`
	Account  string `json:"account"`
	Password string `json:"password"`
}

// LoginInputDto carries what a sign in needs.
type LoginInputDto struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}
