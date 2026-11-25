package serviceLogImpl

import (
	"regexp"
	"sync"

	serviceLog "github.com/wensboy/lightpc/backend/domain/service/log"
)

// Configuration paths read by the service.
const (
	// databaseKey names the database entry the log table lives in.
	databaseKey = "default.runtime.services.log.database"
	// limitKey caps how many records one read returns.
	limitKey = "default.runtime.services.log.limit"
)

// defaultLimit is what a read returns when neither the request nor the configuration names a cap.
const defaultLimit = 500

// levelFilter lets through the levels the interface has ticked.
//
// The interface sends a set of levels rather than a floor, because the menu ticks them one by one. An empty set
// disables the filter instead of rejecting everything: a cleared menu means "show all", which is what the toolbar
// shows before the user touches it.
type levelFilter struct {
	mu     sync.RWMutex
	levels map[int64]struct{}
}

// Name implements serviceLog.Filter.
func (f *levelFilter) Name() string { return "level" }

// Set replaces the allowed set. The map is built here, so a reader only ever holds a reference to a finished set.
func (f *levelFilter) Set(levels []int64) {
	next := make(map[int64]struct{}, len(levels))
	for _, level := range levels {
		next[level] = struct{}{}
	}

	f.mu.Lock()
	f.levels = next
	f.mu.Unlock()
}

// Apply implements serviceLog.Filter, keeping the order of the records it keeps.
func (f *levelFilter) Apply(records []serviceLog.LogRecordDto) []serviceLog.LogRecordDto {
	f.mu.RLock()
	allowed := f.levels
	f.mu.RUnlock()

	if len(allowed) == 0 {
		return records
	}

	kept := records[:0]
	for _, record := range records {
		if _, ok := allowed[record.Level]; ok {
			kept = append(kept, record)
		}
	}
	return kept
}

// patternFilter lets through the records whose message or detail matches a regular expression.
//
// An empty pattern disables the filter. A pattern that does not compile is treated the same way, so a half typed
// expression in the toolbar never blanks the list: the reader sees everything until the expression becomes valid.
type patternFilter struct {
	mu      sync.RWMutex
	pattern *regexp.Regexp
	// text keeps the source of the expression, so a caller can tell "off" from "invalid".
	text string
}

// Name implements serviceLog.Filter.
func (f *patternFilter) Name() string { return "pattern" }

// Set replaces the expression. An empty or invalid expression clears it.
func (f *patternFilter) Set(pattern string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.text = pattern
	if pattern == "" {
		f.pattern = nil
		return
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		f.pattern = nil
		return
	}
	f.pattern = compiled
}

// Apply implements serviceLog.Filter, keeping the order of the records it keeps.
func (f *patternFilter) Apply(records []serviceLog.LogRecordDto) []serviceLog.LogRecordDto {
	f.mu.RLock()
	pattern := f.pattern
	f.mu.RUnlock()

	if pattern == nil {
		return records
	}

	kept := records[:0]
	for _, record := range records {
		// Only the two text fields are searched. The level, the instant and the source are metadata: matching them
		// would make a search for "4" return every warning, which is not what someone typing a keyword expects.
		if pattern.MatchString(record.Msg) || (record.Detail != "" && pattern.MatchString(record.Detail)) {
			kept = append(kept, record)
		}
	}
	return kept
}
