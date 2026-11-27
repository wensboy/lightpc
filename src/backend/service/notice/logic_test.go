package serviceNoticeImpl

import (
	"testing"

	engineNotice "github.com/wensboy/lightpc/backend/domain/engine/notice"
	engineSetting "github.com/wensboy/lightpc/backend/domain/engine/setting"
	serviceNotice "github.com/wensboy/lightpc/backend/domain/service/notice"
)

// stubRing stands in for the engine: it answers with whatever it was handed, newest first.
type stubRing struct {
	engineNotice.Client
	records []engineNotice.Record
	removed []uint64
}

func (s *stubRing) History() []engineNotice.Record {
	history := make([]engineNotice.Record, len(s.records))
	for index, record := range s.records {
		history[len(s.records)-1-index] = record
	}
	return history
}

func (s *stubRing) Remove(id uint64) bool {
	s.removed = append(s.removed, id)
	for index, record := range s.records {
		if record.ID == id {
			s.records = append(s.records[:index], s.records[index+1:]...)
			return true
		}
	}
	return false
}

// resolveOf builds a resolver that answers one key the way the settings engine does: with an Entry, not a bare value.
func resolveOf(value any) func(string) (any, bool) {
	return func(key string) (any, bool) {
		if key != historyKey {
			return nil, false
		}
		return engineSetting.Entry[any]{Layer: "user", Mode: engineSetting.ModeReadWrite, Value: value}, true
	}
}

func newService(ring engineNotice.Client, resolve func(string) (any, bool)) *NoticeService {
	return &NoticeService{notices: ring, resolve: resolve}
}

func TestLimitReadsTheEntryNotTheWrapper(t *testing.T) {
	// The preference travels wrapped; reading the wrapper as the value silently fell back to the default, which is why
	// the setting appeared to have no effect at all.
	if limit := newService(&stubRing{}, resolveOf(float64(5))).limit(); limit != 5 {
		t.Fatalf("limit = %d", limit)
	}
	if limit := newService(&stubRing{}, resolveOf(7)).limit(); limit != 7 {
		t.Fatalf("limit = %d", limit)
	}
}

func TestHistoryKeepsTheNewestWindow(t *testing.T) {
	ring := &stubRing{records: []engineNotice.Record{
		{ID: 1, Level: engineNotice.LevelInfo, Title: "one"},
		{ID: 2, Level: engineNotice.LevelWarn, Title: "two"},
		{ID: 3, Level: engineNotice.LevelFail, Title: "three"},
		{ID: 4, Level: engineNotice.LevelError, Title: "four"},
		{ID: 5, Level: engineNotice.LevelSuccess, Title: "five"},
		{ID: 6, Level: engineNotice.LevelInfo, Title: "six"},
		{ID: 7, Level: engineNotice.LevelWarn, Title: "seven"},
	}}

	history, err := newService(ring, resolveOf(5)).History()
	if err != nil {
		t.Fatal(err)
	}

	// The window holds the newest five; the two before them stay in the ring and simply fall out of view.
	if len(history.Items) != 5 {
		t.Fatalf("items = %d", len(history.Items))
	}
	if history.Items[0].ID != 7 || history.Items[4].ID != 3 {
		t.Fatalf("items = %+v", history.Items)
	}
	if history.LeftEdge != 3 {
		t.Fatalf("left edge = %d", history.LeftEdge)
	}

	// The counts describe the window: ids 3 to 7, which include one warn (7), one fail (3) and one error (4). Ids 1 and
	// 2 also carry a warn, and they are exactly what the cap left out.
	if history.Summary.Total != 5 || history.Summary.Warn != 1 || history.Summary.Fail != 1 || history.Summary.Error != 1 {
		t.Fatalf("summary = %+v", history.Summary)
	}
}

func TestHistoryShrinksAfterARemoval(t *testing.T) {
	ring := &stubRing{records: []engineNotice.Record{}}
	for id := uint64(1); id <= 5; id++ {
		ring.records = append(ring.records, engineNotice.Record{ID: id, Level: engineNotice.LevelInfo})
	}
	service := newService(ring, resolveOf(5))

	// Removing one row in the middle leaves four: the window does not pad itself from what is left in the ring.
	if _, err := service.Remove(serviceNotice.NoticeRemoveInputDto{IDs: []uint64{3}}); err != nil {
		t.Fatal(err)
	}

	history, err := service.History()
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Items) != 4 {
		t.Fatalf("items = %d", len(history.Items))
	}
	for _, item := range history.Items {
		if item.ID == 3 {
			t.Fatalf("removed row is still listed")
		}
	}
}

func TestClearHidesEverythingUpToTheNewestId(t *testing.T) {
	ring := &stubRing{records: []engineNotice.Record{}}
	for id := uint64(1); id <= 7; id++ {
		ring.records = append(ring.records, engineNotice.Record{ID: id, Level: engineNotice.LevelInfo})
	}
	service := newService(ring, resolveOf(5))

	if cleared, err := service.Clear(); err != nil || cleared != 7 {
		t.Fatalf("clear = %d, %v", cleared, err)
	}

	history, err := service.History()
	if err != nil {
		t.Fatal(err)
	}
	// The window used to show ids 3 to 7; after clearing, the older two must not slide back in to fill the gap.
	if len(history.Items) != 0 {
		t.Fatalf("items = %+v", history.Items)
	}

	// A notice raised after the clear is visible, and the ones the mark covered stay hidden however many arrive.
	ring.records = append(ring.records, engineNotice.Record{ID: 8, Level: engineNotice.LevelWarn})
	history, err = service.History()
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Items) != 1 || history.Items[0].ID != 8 {
		t.Fatalf("items = %+v", history.Items)
	}

	for id := uint64(9); id <= 13; id++ {
		ring.records = append(ring.records, engineNotice.Record{ID: id, Level: engineNotice.LevelInfo})
	}
	history, err = service.History()
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Items) != 5 || history.Items[4].ID != 9 {
		t.Fatalf("items = %+v", history.Items)
	}
	for _, item := range history.Items {
		if item.ID <= 7 {
			t.Fatalf("cleared id %d came back", item.ID)
		}
	}
}
