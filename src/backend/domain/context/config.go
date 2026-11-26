package context

import (
	"encoding/json"
	"log/slog"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
)

type ConfigContext domainConfig.Source

// AssignIfZero assigns the configuration text stored under key to target when target still equals zero.
// A missing key, a non text value or a failed parse leaves target untouched.
func AssignIfZero[T comparable](cfg ConfigContext, key string, target *T, zero T, parse func(string) (T, bool)) {
	if cfg == nil || target == nil || *target != zero {
		return
	}
	value, found := cfg.Lookup(key)
	if !found {
		return
	}
	text, ok := value.(string)
	if !ok {
		return
	}
	if parsed, ok := parse(text); ok {
		*target = parsed
	}
}

// Decode unmarshals the subtree stored under key into T.
//
// Reading the document is the caller's job; this only turns what a dotted lookup returns into the structure that
// describes it. A missing key, an absent context or a value that does not fit the target all yield the zero value, so
// a caller can treat "not configured" and "configured wrongly" the same way and decide what that means for it.
func Decode[T any](cfg ConfigContext, key string) T {
	var target T
	if cfg == nil {
		return target
	}
	value, found := cfg.Lookup(key)
	if !found {
		return target
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return target
	}
	if err := json.Unmarshal(encoded, &target); err != nil {
		slog.Default().Warn("config decode failed", "key", key, "err", err)
		return target
	}
	return target
}
