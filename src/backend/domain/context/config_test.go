package context

import (
	"strconv"
	"testing"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
)

type stubSource struct {
	values map[string]any
}

func (s stubSource) Kind() uint8 { return domainConfig.SourceConfig }

func (s stubSource) Lookup(key string) (any, bool) {
	value, ok := s.values[key]
	return value, ok
}

func parseMode(text string) (uint32, bool) {
	mode, err := strconv.ParseUint(text, 0, 32)
	if err != nil {
		return 0, false
	}
	return uint32(mode), true
}

func TestAssignIfZeroAssignsParsedValue(t *testing.T) {
	cfg := ConfigContext(stubSource{values: map[string]any{"mode": "0o600"}})
	mode := uint32(0)

	AssignIfZero(cfg, "mode", &mode, uint32(0), parseMode)

	if mode != 0o600 {
		t.Fatalf("mode = %v", mode)
	}
}

func TestAssignIfZeroKeepsExistingValue(t *testing.T) {
	cfg := ConfigContext(stubSource{values: map[string]any{"mode": "0o600"}})
	mode := uint32(0o700)

	AssignIfZero(cfg, "mode", &mode, uint32(0), parseMode)

	if mode != 0o700 {
		t.Fatalf("mode = %v", mode)
	}
}

func TestAssignIfZeroIgnoresMissingKey(t *testing.T) {
	cfg := ConfigContext(stubSource{values: map[string]any{}})
	mode := uint32(0)

	AssignIfZero(cfg, "mode", &mode, uint32(0), parseMode)

	if mode != 0 {
		t.Fatalf("mode = %v", mode)
	}
}

func TestAssignIfZeroIgnoresNonTextValue(t *testing.T) {
	cfg := ConfigContext(stubSource{values: map[string]any{"mode": 420}})
	mode := uint32(0)

	AssignIfZero(cfg, "mode", &mode, uint32(0), parseMode)

	if mode != 0 {
		t.Fatalf("mode = %v", mode)
	}
}

func TestAssignIfZeroIgnoresParseFailure(t *testing.T) {
	cfg := ConfigContext(stubSource{values: map[string]any{"mode": "invalid"}})
	mode := uint32(0)

	AssignIfZero(cfg, "mode", &mode, uint32(0), parseMode)

	if mode != 0 {
		t.Fatalf("mode = %v", mode)
	}
}

func TestAssignIfZeroIgnoresNilConfig(t *testing.T) {
	mode := uint32(0)

	AssignIfZero(nil, "mode", &mode, uint32(0), parseMode)

	if mode != 0 {
		t.Fatalf("mode = %v", mode)
	}
}
