package serviceNoticeImpl

import (
	"fmt"
	"log/slog"

	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineNotice "github.com/wensboy/lightpc/backend/domain/engine/notice"
	engineSetting "github.com/wensboy/lightpc/backend/domain/engine/setting"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	serviceNotice "github.com/wensboy/lightpc/backend/domain/service/notice"
	"github.com/wensboy/lightpc/backend/internal/cast"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

var (
	_ domainService.Service      = (*NoticeService)(nil)
	_ serviceNotice.NoticeClient = (*NoticeService)(nil)
)

// NoticeService serves the notice sheet: it reads the list, raises a notice on behalf of the interface, and removes
// the ones the person dismissed.
//
// It owns no state of its own. The engine holds the ring and the settings engine holds the cap, so every method is a
// translation between the two and the dto shape the interface expects. That is deliberate: a notice is process state,
// and duplicating it here would mean two lists that can disagree.
type NoticeService struct {
	notices engineNotice.Client
	// resolve reads a setting through the settings engine, so a user override is honoured without this service
	// knowing which layer it came from.
	resolve func(string) (any, bool)
}

// NewNoticeService builds the service. Init resolves the two engines.
func NewNoticeService() *NoticeService {
	return &NoticeService{}
}

// Group implements domainService.Service.
func (s *NoticeService) Group() string { return domainService.GroupNotice }

// Deps implements domainService.Service: the ring lives in the notice engine, the cap in the settings engine.
func (s *NoticeService) Deps() []string {
	return []string{domainEngine.GroupNotice, domainEngine.GroupSetting}
}

// Init implements domainService.Service.
func (s *NoticeService) Init(sCtx domainService.ServiceContext) error {
	if sCtx.AppCtx != nil {
		interError.RegisterAll(sCtx.AppCtx.ErrorContext, serviceNotice.Errors...)
	}

	noticeKey := domainEngine.EngineDep(domainEngine.GroupNotice, domainEngine.NameNotice)
	noticeEngine, found := sCtx.Engines[noticeKey]
	if !found {
		return interError.Retrieve(serviceNotice.ErrKeyNoticeMissing).
			Wrap(fmt.Errorf("engine %q is not ready", noticeKey))
	}
	notices, ok := cast.As[engineNotice.Client](noticeEngine)
	if !ok {
		return interError.Retrieve(serviceNotice.ErrKeyNoticeMissing).
			Wrap(fmt.Errorf("engine %q is not a notice engine", noticeKey))
	}
	s.notices = notices

	settingKey := domainEngine.EngineDep(domainEngine.GroupSetting, domainEngine.NameSetting)
	settingEngine, found := sCtx.Engines[settingKey]
	if !found {
		return interError.Retrieve(serviceNotice.ErrKeySettingMissing).
			Wrap(fmt.Errorf("engine %q is not ready", settingKey))
	}
	settings, ok := cast.As[engineSetting.Setting](settingEngine)
	if !ok {
		return interError.Retrieve(serviceNotice.ErrKeySettingMissing).
			Wrap(fmt.Errorf("engine %q is not a settings engine", settingKey))
	}
	s.resolve = settings.Resolve

	slog.Default().Warn("notice service ready", "history", s.limit())
	return nil
}

// Deinit implements domainService.Service.
func (s *NoticeService) Deinit() error { return nil }

// History implements serviceNotice.NoticeClient.
//
// The cap is applied here rather than in the engine: the ring keeps what the process produced, while how much of it
// the sheet shows is a preference. Trimming on the way out means changing the preference takes effect on the next
// read without touching what is stored.
func (s *NoticeService) History() (serviceNotice.NoticeHistoryDto, error) {
	if s.notices == nil {
		return serviceNotice.NoticeHistoryDto{Items: []serviceNotice.NoticeItemDto{}}, interError.Retrieve(serviceNotice.ErrKeyNoticeMissing)
	}

	limit := s.limit()
	records := s.notices.History()
	if len(records) > limit {
		records = records[:limit]
	}

	items := make([]serviceNotice.NoticeItemDto, 0, len(records))
	for _, record := range records {
		items = append(items, serviceNotice.NoticeItemDto{
			ID:        record.ID,
			Level:     serviceNotice.Level(record.Level),
			Title:     record.Title,
			Content:   record.Content,
			Source:    record.Source,
			Timestamp: record.Timestamp,
		})
	}

	// The oldest id in the answer is the left edge. It is zero for an empty window, which lets a caller tell "nothing
	// yet" from "everything shown" without comparing lengths against the cap.
	var leftEdge uint64
	if len(items) > 0 {
		leftEdge = items[len(items)-1].ID
	}
	return serviceNotice.NoticeHistoryDto{Items: items, Limit: limit, LeftEdge: leftEdge}, nil
}

// Send implements serviceNotice.NoticeClient.
//
// The interface raises its notices through the same engine the backend uses, so a toast and a stored notice are always
// the same event: the engine announces it on the bus after recording it.
func (s *NoticeService) Send(input serviceNotice.NoticeSendInputDto) (serviceNotice.NoticeItemDto, error) {
	if s.notices == nil {
		return serviceNotice.NoticeItemDto{}, interError.Retrieve(serviceNotice.ErrKeyNoticeMissing)
	}

	record, err := s.notices.Notify(engineNotice.Input{
		Level:   engineNotice.Level(input.Level),
		Title:   input.Title,
		Content: input.Content,
		Source:  input.Source,
	})
	if err != nil {
		return serviceNotice.NoticeItemDto{}, err
	}

	return serviceNotice.NoticeItemDto{
		ID:        record.ID,
		Level:     serviceNotice.Level(record.Level),
		Title:     record.Title,
		Content:   record.Content,
		Source:    record.Source,
		Timestamp: record.Timestamp,
	}, nil
}

// Remove implements serviceNotice.NoticeClient.
//
// One list serves both gestures: a single row sends one id, the clear button sends every id it holds. Ids that are no
// longer present are not an error, because the caller cannot know what another window removed in between; the count
// reports what actually went.
func (s *NoticeService) Remove(input serviceNotice.NoticeRemoveInputDto) (int, error) {
	if s.notices == nil {
		return 0, interError.Retrieve(serviceNotice.ErrKeyNoticeMissing)
	}

	removed := 0
	for _, id := range input.IDs {
		if id == 0 {
			continue
		}
		if s.notices.Remove(id) {
			removed++
		}
	}
	if removed == 0 && len(input.IDs) > 0 {
		// Every id was unusable, which is a caller mistake rather than an empty list.
		return 0, interError.Retrieve(serviceNotice.ErrKeyInvalidId)
	}
	return removed, nil
}

// limit resolves how many notices the sheet shows, falling back when the preference is unreadable.
func (s *NoticeService) limit() int {
	if s.resolve == nil {
		return defaultHistory
	}
	value, found := s.resolve(historyKey)
	if !found {
		return defaultHistory
	}
	limit, ok := intValue(value)
	if !ok || limit <= 0 {
		return defaultHistory
	}
	return limit
}
