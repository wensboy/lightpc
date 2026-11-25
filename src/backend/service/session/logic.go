// Package session implements the account lifecycle: registration, sign in, sign out and removal, plus the
// state a caller reads to decide whether the current visitor is signed in.
package serviceSessionImpl

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	serviceSession "github.com/wensboy/lightpc/backend/domain/service/session"
	"github.com/wensboy/lightpc/backend/internal/cast"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// SessionService owns the accounts and the live pairs.
//
// A pair is the uid and the account the front end keeps locally. There is no token: the pair identifies the rows,
// the history says whether the window is still open, and the cache only answers "already checked" for the pairs
// this process signed in. A restart drops the cache and the history takes over.
type SessionService struct {
	mu    sync.RWMutex
	store engineStore.DBStore
	bus   domainBus.Bus
	// pair is the one session this process signed in. A single client means a single live pair, so no lookup
	// structure is needed: the pair either matches or it does not.
	pair serviceSession.AuthMetaDto
	// expiresAt is the deadline of that pair, zero when the window never closes.
	expiresAt  time.Time
	timer      *time.Timer
	expiration time.Duration
	softDelete bool
	state      domainEngine.EngineState
}

// NewSessionService builds the session service. Init resolves the database and the configuration.
func NewSessionService() *SessionService {
	return &SessionService{state: domainEngine.EnginePending}
}

// Group implements domainService.Service.
func (s *SessionService) Group() string { return domainService.GroupSession }

// Deps implements domainService.Service.
func (s *SessionService) Deps() []string { return []string{domainEngine.StoreDb} }

// State implements domainService.Service.
func (s *SessionService) State() domainEngine.EngineState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

// Init resolves the database, reads the configuration and prepares the pair cache.
//
// The cache lifetime is read from the document, so a deployment can change the window without a rebuild. A zero
// window means a sign in never expires, which the cache expresses as no TTL.
func (s *SessionService) Init(sCtx domainService.ServiceContext) error {
	if sCtx.AppCtx != nil {
		interError.RegisterAll(sCtx.AppCtx.ErrorContext, serviceSession.Errors...)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	engine, found := sCtx.Engines[domainEngine.StoreDb]
	if !found {
		s.state = domainEngine.EngineError
		return interError.Retrieve(serviceSession.ErrKeyDatabaseMissing)
	}

	registry, ok := cast.As[engineStore.DBRegistry](engine)
	if !ok {
		s.state = domainEngine.EngineError
		return interError.Retrieve(serviceSession.ErrKeyDatabaseMissing)
	}

	key := ""
	if sCtx.AppCtx != nil && sCtx.AppCtx.ConfigContext != nil {
		key = stringValue(sCtx.AppCtx.ConfigContext, databaseKey)
		s.softDelete = boolValue(sCtx.AppCtx.ConfigContext, softDeleteKey, true)
		s.expiration = durationValue(sCtx.AppCtx.ConfigContext, expirationKey, 2*time.Hour)
	}
	if key == "" {
		keys := registry.DBKeys()
		if len(keys) == 0 {
			s.state = domainEngine.EngineError
			return interError.Retrieve(serviceSession.ErrKeyDatabaseMissing)
		}
		key = keys[0]
	}

	store, found := registry.DB(key)
	if !found {
		s.state = domainEngine.EngineError
		return interError.Retrieve(serviceSession.ErrKeyDatabaseMissing).Wrap(fmt.Errorf("database %q is not registered", key))
	}
	s.store = store

	s.bus = sCtx.Ebus

	s.state = domainEngine.EngineReady
	slog.Default().Warn("session service ready", "database", key, "expiration", s.expiration.String())
	return nil
}

// Start implements domainService.Service.
func (s *SessionService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = domainEngine.EngineRunning
	return nil
}

// Stop implements domainService.Service by releasing the pending countdown.
func (s *SessionService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopTimerLocked()
	s.pair = serviceSession.AuthMetaDto{}
	s.expiresAt = time.Time{}
	s.state = domainEngine.EngineStopped
	return nil
}

// Deinit implements domainService.Service.
func (s *SessionService) Deinit() error {
	return s.Stop()
}

// Register creates the account, its profile row and the sign up history entry, then signs it in.
func (s *SessionService) Register(input serviceSession.RegisterInputDto) (serviceSession.StatusDto, error) {
	if input.Account == "" || input.Password == "" {
		return serviceSession.StatusDto{}, interError.Retrieve(interError.ErrCodeInvalidParam)
	}

	store := s.db()
	if store == nil {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyDatabaseMissing)
	}

	existing, err := s.findAccount(store, input.Account)
	if err != nil {
		return serviceSession.StatusDto{}, err
	}
	if existing != nil {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyAccountExists)
	}

	uid, err := identifier()
	if err != nil {
		return serviceSession.StatusDto{}, err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return serviceSession.StatusDto{}, err
	}

	user := UserDao{UID: uid, Uname: input.Uname, Account: input.Account, Password: string(hashed)}
	if err := s.insertAccount(store, user, int64(s.expiration.Seconds())); err != nil {
		return serviceSession.StatusDto{}, err
	}

	// Registration creates the account and records it, but it does not sign in: the person asked for an account, not
	// for a session, so the interface sends them to the sign in form and they enter the password themselves.
	return serviceSession.StatusDto{
		Authenticated: false,
		UID:           uid,
		Account:       input.Account,
		Uname:         input.Uname,
	}, nil
}

// Login verifies the credentials, records the action and marks the pair as live.
func (s *SessionService) Login(input serviceSession.LoginInputDto) (serviceSession.StatusDto, error) {
	store := s.db()
	if store == nil {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyDatabaseMissing)
	}

	account, err := s.findAccount(store, input.Account)
	if err != nil {
		return serviceSession.StatusDto{}, err
	}
	if account == nil {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyCredentials)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(input.Password)); err != nil {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyCredentials)
	}

	return s.signIn(account.UID, account.Account, account.Uname)
}

// Logout drops the live mark and records the action.
func (s *SessionService) Logout(auth serviceSession.AuthMetaDto) (serviceSession.StatusDto, error) {
	store := s.db()
	if store == nil {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyNotAuthenticated)
	}
	if auth.UID == "" || auth.Account == "" {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyNotAuthenticated)
	}

	s.unmark()
	if err := s.record(store, ActionSignOut, auth, 0); err != nil {
		return serviceSession.StatusDto{}, err
	}
	return serviceSession.StatusDto{}, nil
}

// Unregister removes the account and its profile rows, then records the action.
//
// A soft delete keeps the rows and stamps deleted_at so the history stays readable; a hard one removes them.
// The session history itself is never removed: it is the record of what happened.
func (s *SessionService) Unregister(auth serviceSession.AuthMetaDto) (serviceSession.StatusDto, error) {
	store := s.db()
	if store == nil {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyNotAuthenticated)
	}
	if auth.UID == "" || auth.Account == "" {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyNotAuthenticated)
	}

	s.mu.RLock()
	softDelete := s.softDelete
	s.mu.RUnlock()

	s.unmark()

	var err error
	if softDelete {
		err = s.removeAccount(store, auth.UID, now())
	} else {
		err = s.deleteAccount(store, auth.UID)
	}
	if err != nil {
		return serviceSession.StatusDto{}, err
	}
	if err := s.record(store, ActionSignDown, auth, 0); err != nil {
		return serviceSession.StatusDto{}, err
	}
	return serviceSession.StatusDto{}, nil
}

// Current reports the signed in state of the given pair.
//
// The order matters: a live cache entry answers on its own, and when it is absent the history is read so a
// restart inside the window keeps the caller signed in. Both halves must agree before the state is reported as
// signed in, which is what keeps a stale pair from passing on its own.
func (s *SessionService) Current(auth serviceSession.AuthMetaDto) (serviceSession.StatusDto, error) {
	store := s.db()
	if store == nil {
		return serviceSession.StatusDto{}, interError.Retrieve(serviceSession.ErrKeyDatabaseMissing)
	}
	if auth.UID == "" || auth.Account == "" {
		return serviceSession.StatusDto{}, nil
	}

	if ok, deadline := s.live(auth); ok {
		account, err := s.findAccountByUID(store, auth.UID, auth.Account)
		if err != nil {
			return serviceSession.StatusDto{}, err
		}
		if account == nil {
			return serviceSession.StatusDto{}, nil
		}

		expiresAt := int64(0)
		if !deadline.IsZero() {
			expiresAt = deadline.Unix()
		}
		return serviceSession.StatusDto{
			Authenticated: true,
			UID:           auth.UID,
			Account:       auth.Account,
			Uname:         account.Uname,
			ExpiresAt:     expiresAt,
		}, nil
	}

	return s.restore(store, auth)
}

// restore reads the latest sign in row of the pair and accepts it while the window still allows it.
//
// The stored time is the authority, so restarting twice cannot renew a window. An expiration of zero means the
// sign in never expires, which is why it short circuits the comparison instead of producing a zero deadline.
func (s *SessionService) restore(store engineStore.DBStore, auth serviceSession.AuthMetaDto) (serviceSession.StatusDto, error) {
	row, err := s.latestSignIn(store, auth)
	if err != nil || row == nil {
		return serviceSession.StatusDto{}, err
	}

	expiresAt := int64(0)
	if row.Expiration > 0 {
		limit := time.Unix(row.CreatedAt, 0).Add(time.Duration(row.Expiration) * time.Second)
		if time.Now().After(limit) {
			return serviceSession.StatusDto{}, nil
		}
		expiresAt = limit.Unix()
	}

	account, err := s.findAccountByUID(store, auth.UID, auth.Account)
	if err != nil {
		return serviceSession.StatusDto{}, err
	}
	if account == nil {
		return serviceSession.StatusDto{}, nil
	}

	// The pair passed both checks. A stored window arms the countdown with the time it has left, so a restart
	// never extends it; a zero expiration arms nothing and the pair stays live.
	if expiresAt > 0 {
		s.markUntil(auth, time.Unix(expiresAt, 0))
	} else {
		s.markUntil(auth, time.Time{})
	}

	return serviceSession.StatusDto{
		Authenticated: true,
		UID:           auth.UID,
		Account:       auth.Account,
		Uname:         account.Uname,
		ExpiresAt:     expiresAt,
	}, nil
}
