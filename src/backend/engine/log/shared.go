package logimpl

import (
	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	"github.com/wensboy/lightpc/backend/internal/cast"
)

// logGroupKey is the configuration path that holds every log group.
const logGroupKey = "default.runtime.log"

// ItemsFor reads the configured items of one log group.
//
// Parsing lives here so the bootstrap only decides which groups to start, never how a configuration
// record is shaped. A missing group and an empty group both yield no items.
func ItemsFor(cfg domainContext.ConfigContext, group string) []domainConfig.LogItem {
	value, ok := cfg.Lookup(logGroupKey)
	if !ok {
		return nil
	}
	groups, ok := cast.To[map[string]any](value)
	if !ok {
		return nil
	}
	return parseItems(groups[group])
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
