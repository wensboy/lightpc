package bootstrap

import (
	"testing"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
)

type stubSource struct {
	values map[string]any
}

func (s *stubSource) Kind() uint8 { return domainConfig.SourceConfig }

func (s *stubSource) Lookup(key string) (any, bool) {
	value, ok := s.values[key]
	return value, ok
}

func TestSetupErrors(t *testing.T) {
	cfg := domainContext.ConfigContext(&stubSource{values: map[string]any{
		"errors": map[string]any{
			"version":     "v1",
			"description": "错误映射定义",
			"root": map[string]any{
				"description": "根因",
				"definition": []any{
					map[string]any{"code": "10010000", "kind": "Fatal"},
					map[string]any{"code": "10010001", "kind": "Unknown"},
					map[string]any{"code": "10010002"},
				},
			},
			"business": map[string]any{
				"definition": []any{
					map[string]any{"code": "20010000", "kind": "Business"},
				},
			},
		},
	}})

	errorContext := SetupErrors(cfg)

	if len(errorContext) != 3 {
		t.Fatalf("errorContext = %v", errorContext)
	}
	if errorContext["10010000"] != "Fatal" || errorContext["10010001"] != "Unknown" || errorContext["20010000"] != "Business" {
		t.Fatalf("errorContext = %v", errorContext)
	}
}

func TestSetupErrorsWithoutConfig(t *testing.T) {
	cfg := domainContext.ConfigContext(&stubSource{values: map[string]any{}})

	errorContext := SetupErrors(cfg)

	if len(errorContext) != 0 {
		t.Fatalf("errorContext = %v", errorContext)
	}
}
