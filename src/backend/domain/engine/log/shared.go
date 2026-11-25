package engineLog

import (
	"encoding/json"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// EntriesKey is the configuration path every log engine reads its entries from.
const EntriesKey = "default.runtime.engines.log"

// Sub-groups of EntriesKey. Each names one engine, and the entry list lives directly under it.
const (
	SubDirect  = "direct"
	SubFanout  = "fanout"
	SubPattern = "pattern"
)

// Error keys published by the log engines during Init.
const (
	// ErrKeyInvalidItem reports a configured item without a name or without any usable output.
	ErrKeyInvalidItem = "error::engine.log::invalid_item"
	// ErrKeyHandlerFailed reports an output whose handler could not be created.
	ErrKeyHandlerFailed = "error::engine.log::handler_failed"
)

// Errors lists every log error published when a log engine initialises.
// Kind is resolved from the application error context at registration time.
var Errors = []interError.Option{
	{Key: ErrKeyInvalidItem, Code: interError.ErrCodeInvalidParam, Msg: "log: invalid item"},
	{Key: ErrKeyHandlerFailed, Code: interError.ErrCodeOpFailed, Msg: "log: handler creation failed"},
}

type LogClient struct {
	Direct  DirectLog
	Fanout  FanoutLog
	Pattern PatternLog
}

// LogHookContext is what a log hook may inspect between the engine phases. All three log engines share the
// same shape: they differ in how a record is routed, not in what they need to prepare.
type LogHookContext struct {
	// LogEntries lists the configured entries of the group, in configuration order.
	LogEntries []LogEntry
	// EngineContext is the context the engine was initialised with.
	EngineContext domainEngine.EngineContext
}

// DirectHookContext is the direct log engine's hook context.
type DirectHookContext = LogHookContext

// FanoutHookContext is the fanout log engine's hook context.
type FanoutHookContext = LogHookContext

// PatternHookContext is the pattern log engine's hook context.
type PatternHookContext = LogHookContext

// 下面结构完全映射 data/conf/default.json#runtime.engine.log
type LogSerializer struct {
	Type         string `json:"type"`
	EnableSource bool   `json:"enableSource"`
}

// 映射 data/conf/default.json#runtime.engine.log.<sub>.[i]结构
type LogEntry struct {
	Name    string `json:"name"`
	Level   string `json:"level"`
	Handler string `json:"handler"`
	// 根据当前所属的 sub 和 handler 决定解析结构
	// sub = direct, handler = file -> FileOutput
	// sub = direct, handler = console -> ConsoleOutput
	// sub = direct, hadnler = xxx -> XxxOutput
	// sub = fanout, handler = file -> []FileOutput
	// sub = fanout, handler = console -> []ConsoleOutput
	// sub = fanout, handler = xxx -> []XxxOutput
	// sub = pattern, handler = file -> []FileOutput
	// sub = pattern, handler = console -> []ConsoleOutput
	// sub = pattern, handler = xxx -> []XxxOutput
	Output json.RawMessage `json:"output"`
}

// 终端输出
type ConsoleOutput struct {
	Serializer LogSerializer `json:"serializer"`
	// 语义: 只能为 stdout, stderr
	Path string `json:"path"`
}

// 文件输出
type FileOutput struct {
	Serializer LogSerializer `json:"serializer"`
	// 语义: 有效的本地文件路径.
	Path string `json:"path"`
	Mode string `json:"mode"`
}

// 数据库输出
type DatabaseOutput struct {
	Serializer LogSerializer `json:"serializer"`
	// 语义: 已注册的 db 名称, 通过 store.db 获取实例.
	Path string `json:"path"`
}

// EntriesFor reads the entries of one sub-group.
//
// The document is decoded straight into the structures that describe it, so nothing outside this package has to
// know how a record is shaped. The output stays a raw message because its meaning depends on the entry: a direct
// entry declares one object while fanout and pattern declare an array.
func EntriesFor(cfg domainContext.ConfigContext, sub string) []LogEntry {
	value, found := cfg.Lookup(EntriesKey)
	if !found {
		return nil
	}

	groups, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := groups[sub]
	if !ok {
		return nil
	}

	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	entries := make([]LogEntry, 0)
	if err := json.Unmarshal(encoded, &entries); err != nil {
		return nil
	}

	kept := entries[:0]
	for _, entry := range entries {
		if entry.Name == "" {
			continue
		}
		kept = append(kept, entry)
	}
	return kept
}
