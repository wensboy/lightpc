package serviceSessionImpl

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	serviceSession "github.com/wensboy/lightpc/backend/domain/service/session"
	"github.com/wensboy/lightpc/backend/internal/cast"
	"github.com/wensboy/lightpc/backend/internal/config"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// Action is one entry of the session history. The four values are the ones the check constraint on the column
// accepts, so a typo fails at the insert rather than producing a row nothing can interpret.
type Action string

const (
	ActionSignIn   Action = "sign_in"
	ActionSignOut  Action = "sign_out"
	ActionSignUp   Action = "sign_up"
	ActionSignDown Action = "sign_down"
)

// pairKey names a session in a log line. The pair is also what the front end stores locally, so the same string
// appears on both sides and no opaque value has to be handed around.
func pairKey(auth serviceSession.AuthMetaDto) string {
	return auth.UID + ":" + auth.Account
}

// expirationKey names the configured window length.
const expirationKey = "default.runtime.services.session.expiration"

// databaseKey optionally names the database entry the tables live in.
const databaseKey = "default.runtime.services.session.database"

// softDeleteKey controls whether a removal keeps the rows.
const softDeleteKey = "default.runtime.services.softDelete"

// stringValue reads a text setting, returning an empty string when it is absent.
func stringValue(cfg domainContext.ConfigContext, key string) string {
	value, found := cfg.Lookup(key)
	if !found {
		return ""
	}
	text, _ := cast.To[string](value)
	return text
}

// boolValue reads a boolean setting, falling back to the given default.
func boolValue(cfg domainContext.ConfigContext, key string, fallback bool) bool {
	value, found := cfg.Lookup(key)
	if !found {
		return fallback
	}
	flag, ok := cast.To[bool](value)
	if !ok {
		return fallback
	}
	return flag
}

// durationValue reads a value written as "<number> <unit>", falling back when it is absent or malformed.
//
// Parsing goes through internal/config so the document keeps one notation for every sized field, and the unit
// stays in the document rather than in a comment.
func durationValue(cfg domainContext.ConfigContext, key string, fallback time.Duration) time.Duration {
	text := stringValue(cfg, key)
	if text == "" {
		return fallback
	}
	parsed, err := config.ParseDuration(text)
	if err != nil {
		return fallback
	}
	return parsed
}

// identifier builds the uid of a new account. It stays stable for the lifetime of the row, which is why the
// account can be renamed later without breaking the references that point at it.
func identifier() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

// now is the single place the flow reads the clock, so the history stays consistent within one call.
func now() int64 {
	return time.Now().Unix()
}

// db returns the open database, so a caller does not take the lock itself.
func (s *SessionService) db() engineStore.DBStore {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.store
}

// live reports whether the pair is the one this process signed in and whether its window is still open.
//
// There is no cache layer: a single client means a single live pair, so one deadline on the service answers the
// question outright. The second return is the deadline, zero when the window never closes.
func (s *SessionService) live(auth serviceSession.AuthMetaDto) (bool, time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.pair.UID == "" || s.pair != auth {
		return false, time.Time{}
	}
	if s.expiresAt.IsZero() {
		return true, time.Time{}
	}
	if time.Now().After(s.expiresAt) {
		return false, time.Time{}
	}
	return true, s.expiresAt
}

// mark records the pair as live and arms the countdown that ends it.
//
// A timer rather than a cache sweep: the sweep runs on its own schedule, so a short window would be noticed
// late. The timer fires at the deadline and publishes the event, which is what lets the interface sign out
// without waiting for the next request. A zero window arms nothing and reports no deadline.
func (s *SessionService) mark(auth serviceSession.AuthMetaDto) int64 {
	deadline, _ := s.markUntil(auth, time.Time{})
	return deadline
}

// markUntil makes the pair live until deadline, or for the configured window when deadline is zero.
//
// It returns the deadline as a unix second, zero when the window never closes, and the remaining time so a
// caller can tell whether the countdown was armed at all.
func (s *SessionService) markUntil(auth serviceSession.AuthMetaDto, deadline time.Time) (int64, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pair = auth
	s.expiresAt = time.Time{}
	s.stopTimerLocked()

	if deadline.IsZero() {
		if s.expiration <= 0 {
			return 0, 0
		}
		deadline = time.Now().Add(s.expiration)
	}

	remaining := time.Until(deadline)
	if remaining <= 0 {
		return 0, 0
	}

	s.expiresAt = deadline
	s.timer = time.AfterFunc(remaining, func() { s.expire(auth) })
	return deadline.Unix(), remaining
}

// expire ends a session whose countdown ran out and publishes the event.
//
// The pair is compared before anything is cleared: a sign in that happened after this timer was armed has
// replaced the pair, and its own timer owns the next deadline.
func (s *SessionService) expire(auth serviceSession.AuthMetaDto) {
	s.mu.Lock()
	if s.pair.UID == "" || s.pair != auth || s.expiresAt.IsZero() || time.Now().Before(s.expiresAt) {
		s.mu.Unlock()
		return
	}
	s.pair = serviceSession.AuthMetaDto{}
	s.expiresAt = time.Time{}
	s.timer = nil
	bus := s.bus
	s.mu.Unlock()

	key := pairKey(auth)
	slog.Default().Warn("session window closed", "pair", key)
	if bus != nil {
		bus.Emit(domainBus.EventSessionExpired, key)
	}
}

// unmark drops the live pair and stops its countdown. Nothing is published: a sign out is what the caller asked
// for, so it already knows.
func (s *SessionService) unmark() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pair = serviceSession.AuthMetaDto{}
	s.expiresAt = time.Time{}
	s.stopTimerLocked()
}

// stopTimerLocked releases the pending countdown. The caller holds the lock.
func (s *SessionService) stopTimerLocked() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
}

// signIn records the action, marks the pair and answers with the state.
//
// Both writes happen before the answer is returned, so the live pair and the history never disagree about
// whether a session exists.
func (s *SessionService) signIn(uid, account, uname string) (serviceSession.StatusDto, error) {
	store := s.db()
	if store == nil {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyDatabaseMissing)
	}

	auth := serviceSession.AuthMetaDto{UID: uid, Account: account}
	window := int64(s.expiration.Seconds())
	if err := s.record(store, ActionSignIn, auth, window); err != nil {
		return serviceSession.StatusDto{}, err
	}

	return serviceSession.StatusDto{
		Authenticated: true,
		UID:           uid,
		Account:       account,
		Uname:         uname,
		ExpiresAt:     s.mark(auth),
	}, nil
}
