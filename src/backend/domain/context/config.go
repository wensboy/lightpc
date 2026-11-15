package context

import (
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
