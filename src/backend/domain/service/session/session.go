// Package serviceSession defines the session service contract: the account lifecycle and the state a caller
// reads to decide whether the current visitor is signed in.
package serviceSession

// Client is the surface the front end depends on.
//
// The three calls that act on an existing session take the pair the front end stored locally. It is an
// identifier, not a credential: the backend decides from the history table whether the pair is still valid.
type Client interface {
	// Register creates an account together with its profile rows and records the action.
	Register(input RegisterInputDto) (StatusDto, error)
	// Login checks the credentials, records the action and marks the pair as live.
	Login(input LoginInputDto) (StatusDto, error)
	// Logout drops the live mark and records the action.
	Logout(auth AuthMetaDto) (StatusDto, error)
	// Unregister removes the account, its profile rows and records the action.
	Unregister(auth AuthMetaDto) (StatusDto, error)
	// Current reports the signed in state, restoring it from the history when the process restarted.
	Current(auth AuthMetaDto) (StatusDto, error)
}
