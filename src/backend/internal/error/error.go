package error

import (
	"fmt"
	"sync"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
)

const (
	ErrorRoot = '1' // Code[0]

	// KeyInternalPrefix is the ekey prefix reserved for the built-in root errors.
	KeyInternalPrefix = "error::internal::"
	// DefaultKey is returned by Retrieve whenever the requested ekey is unknown.
	DefaultKey = KeyInternalPrefix + "not_exist"
)

var (
	// registry holds every global error. key format: error::<domain>::<detail>
	// eg: error::engine.store::invalid_target
	registry = make(map[string]*Terror)
	// registryMu protects registry: errors are registered during initialisation and read at runtime.
	registryMu sync.RWMutex
)

// RegisterAll publishes every option, resolving Kind from the error context by Code.
//
// The parameter is the map rather than the application context: this function only needs to look a code up, so
// taking the whole context would let it reach fields it has no business reading. A nil map is fine and yields an
// empty kind, which is what a caller with no context to offer gets.
//
// The first registration wins, so repeated calls during initialisation are safe.
func RegisterAll(errCtx domainContext.ErrorContext, options ...Option) {
	for _, option := range options {
		option.Kind = errCtx[option.Code]
		Register(option.Key, New(option))
	}
}

// Register publishes e under ekey. The first registration wins so every caller
// observes the same instance for a given ekey. A nil error is ignored.
func Register(ekey string, e *Terror) {
	if e == nil {
		return
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, found := registry[ekey]; found {
		return
	}
	e.Key = ekey
	registry[ekey] = e
}

// Retrieve returns the error registered under ekey, the built-in root cause, or an unregistered placeholder.
func Retrieve(ekey string) *Terror {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if e, found := registry[ekey]; found {
		return e
	}
	if e, found := registry[DefaultKey]; found {
		return e
	}
	return &Terror{Key: ekey, Code: ErrCodeUnknown, Msg: "error: unregistered key"}
}

// Terror is a traceable error structure that implements the standard error and Unwrap interfaces.
type Terror struct {
	Key    string
	Code   string
	Kind   string
	Msg    string
	source error
}

type Option struct {
	Key  string
	Code string
	Kind string
	Msg  string
}

func New(opt Option) *Terror {
	return &Terror{
		Key:  opt.Key,
		Code: opt.Code,
		Kind: opt.Kind,
		Msg:  opt.Msg,
	}
}

// Wrap returns a copy carrying err as the root cause, keeping registered instances immutable.
func (e *Terror) Wrap(err error) *Terror {
	clone := *e
	clone.source = err
	return &clone
}

func (e Terror) Error() string {
	out := fmt.Sprintf("[%s|%s] %s", e.Code, e.Kind, e.Msg)
	if e.Code == "" || e.Code[0] == ErrorRoot || e.source == nil {
		return out
	}
	return fmt.Sprintf("%s -> %s", out, e.source.Error())
}

func (e Terror) Unwrap() error {
	return e.source
}

// Is matches by ekey first and falls back to the error code, so wrapped copies keep comparing equal.
func (e Terror) Is(target error) bool {
	other, ok := target.(*Terror)
	if !ok {
		return false
	}
	if e.Key != "" || other.Key != "" {
		return e.Key == other.Key
	}
	return e.Code == other.Code
}
