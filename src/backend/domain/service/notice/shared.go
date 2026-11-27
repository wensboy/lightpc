package serviceNotice

import interError "github.com/wensboy/lightpc/backend/internal/error"

// Error keys published by the notice service during Init.
const (
	// ErrKeyNoticeMissing reports that the notice engine is not ready.
	ErrKeyNoticeMissing = "error::service.notice::notice_missing"
	// ErrKeySettingMissing reports that the settings engine is not ready.
	ErrKeySettingMissing = "error::service.notice::setting_missing"
	// ErrKeyInvalidId reports a delete request whose ids are all unusable.
	ErrKeyInvalidId = "error::service.notice::invalid_id"
)

// The list lives with the contract rather than with the implementation: a caller that compares against a key should
// not have to import the implementation to spell it.
var Errors = []interError.Option{
	{Key: ErrKeyNoticeMissing, Code: interError.ErrCodeNotExist, Msg: "notice: engine unavailable"},
	{Key: ErrKeySettingMissing, Code: interError.ErrCodeNotExist, Msg: "notice: settings unavailable"},
	{Key: ErrKeyInvalidId, Code: interError.ErrCodeInvalidParam, Msg: "notice: no usable id"},
}

// Level is the severity a notice carries, mirroring the engine's own values as text.
//
// It is repeated here rather than imported so this contract does not depend on the engine package: the service is
// what the interface talks to, and its vocabulary should stand on its own.
type Level string

const (
	LevelInfo    Level = "info"    // 信息
	LevelSuccess Level = "success" // 成功
	LevelFail    Level = "fail"    // 失败
	LevelError   Level = "error"   // 错误
	LevelWarn    Level = "warn"    // 警告
)

// NoticeItemDto is one notice the sheet shows.
type NoticeItemDto struct {
	ID        uint64 `json:"id"`
	Level     Level  `json:"level"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Source    string `json:"source"`
	Timestamp int64  `json:"timestamp"`
}

// NoticeSummaryDto counts the notices the answer carries, by severity.
//
// Total plus the three severities worth a badge: the footer shows four cells, and computing them here keeps the count
// and the list from disagreeing when a window is capped or a notice is removed between two reads.
type NoticeSummaryDto struct {
	Total int `json:"total"`
	Fail  int `json:"fail"`
	Error int `json:"error"`
	Warn  int `json:"warn"`
}

// NoticeHistoryDto is the answer of one read.
//
// Limit travels with the list because the count comes from the settings document: the interface can show what it is
// currently capped at without asking a second question.
//
// LeftEdge is the id the window starts at, zero when the list is empty. The engine keeps the ring, the service keeps
// the cap, so the interface cannot compute this on its own: once the ring holds more than the cap, the oldest ids fall
// out of the answer, and that boundary is what "capped" means.
type NoticeHistoryDto struct {
	Items    []NoticeItemDto  `json:"items"`
	Limit    int              `json:"limit"`
	LeftEdge uint64           `json:"leftEdge"`
	Summary  NoticeSummaryDto `json:"summary"`
}

// NoticeSendInputDto is what the interface submits to raise a notice.
type NoticeSendInputDto struct {
	Level   Level  `json:"level"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Source  string `json:"source"`
}

// NoticeRemoveInputDto is what the interface submits to drop notices.
//
// It carries a list because the sheet has two operations with one meaning: a single row is removed with one id, the
// whole list with every id. A separate clear method would duplicate the same authority.
type NoticeRemoveInputDto struct {
	IDs []uint64 `json:"ids"`
}
