package bootstrap

import (
	"testing"

	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	"github.com/wensboy/lightpc/backend/internal/config"
	interError "github.com/wensboy/lightpc/backend/internal/error"
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
					map[string]any{"code": "10010002", "kind": "Not Exist"},
					map[string]any{"code": "10010003"},
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

	if len(errorContext) != 4 {
		t.Fatalf("errorContext = %v", errorContext)
	}
	if errorContext["10010000"] != "Fatal" || errorContext["10010001"] != "Unknown" || errorContext["10010002"] != "Not Exist" || errorContext["20010000"] != "Business" {
		t.Fatalf("errorContext = %v", errorContext)
	}
	if notExist := interError.Retrieve("error::internal::not_exist"); notExist == nil || notExist.Code != interError.ErrCodeNotExist {
		t.Fatalf("error::internal::not_exist = %v", notExist)
	}
	if fatal := interError.Retrieve("error::internal::fatal"); fatal == nil || fatal.Code != interError.ErrCodeFatal {
		t.Fatalf("error::internal::fatal = %v", fatal)
	}
}

func TestSetupErrorsWithoutConfig(t *testing.T) {
	cfg := domainContext.ConfigContext(&stubSource{values: map[string]any{}})

	errorContext := SetupErrors(cfg)

	if len(errorContext) != 0 {
		t.Fatalf("errorContext = %v", errorContext)
	}
}

func TestSetupErrorsFromConfigFile(t *testing.T) {
	fileSource := config.NewFileSource()
	fileSource.Parse(domainContext.FsContext{Paths: []string{"../../data/conf/errors.json"}})

	errorContext := SetupErrors(domainContext.ConfigContext(fileSource))

	if len(errorContext) == 0 {
		t.Fatal("errorContext must not be empty")
	}
	if errorContext[interError.ErrCodeEngineInvalidPath] != "Invalid Path" {
		t.Fatalf("engine invalid path = %q", errorContext[interError.ErrCodeEngineInvalidPath])
	}
	invalidParameter := interError.Retrieve("error::internal::invalid_parameter")
	if invalidParameter == nil || invalidParameter.Code != interError.ErrCodeInvalidParam {
		t.Fatalf("error::internal::invalid_parameter = %v", invalidParameter)
	}
}
