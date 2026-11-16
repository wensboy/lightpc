package logimpl

import (
	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	"github.com/wensboy/lightpc/backend/internal/cast"
)

// Setup builds the log engines described by runtime.log.<group>[].
//
// Each group yields at most one engine and a missing or empty group yields none. Only configuration is
// read here: handlers are opened during Init, once the store engine has created the root directory.
// The fanout engine publishes itself as the slog default so a plain slog call reaches every sink.
func Setup(cfg domainContext.ConfigContext) []domainEngine.Engine {
	engines := make([]domainEngine.Engine, 0, 3)
	value, ok := cfg.Lookup("default.runtime.log")
	if !ok {
		return engines
	}
	groups, ok := cast.To[map[string]any](value)
	if !ok {
		return engines
	}

	if items := parseItems(groups["direct"]); len(items) > 0 {
		options := make([]DirectOption, 0, len(items))
		for _, item := range items {
			options = append(options, WithDirectItem(item))
		}
		engines = append(engines, NewDirectLog(options...))
	}

	if items := parseItems(groups["fanout"]); len(items) > 0 {
		options := make([]FanoutOption, 0, len(items))
		for _, item := range items {
			options = append(options, WithFanoutItem(item))
		}
		engines = append(engines, NewFanoutLog(options...))
	}

	if items := parseItems(groups["pattern"]); len(items) > 0 {
		options := make([]PatternOption, 0, len(items))
		for _, item := range items {
			options = append(options, WithPatternItem(item))
		}
		engines = append(engines, NewPatternLog(options...))
	}

	return engines
}

// parseItems reads every configured item of one group, skipping the entries without a name.
func parseItems(raw any) []domainConfig.LogItem {
	records, ok := cast.To[[]any](raw)
	if !ok {
		return nil
	}
	items := make([]domainConfig.LogItem, 0, len(records))
	for _, record := range records {
		if item, ok := parseItem(record); ok {
			items = append(items, item)
		}
	}
	return items
}
