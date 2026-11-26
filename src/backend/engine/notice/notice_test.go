package engineNoticeImpl

import (
	"testing"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineNotice "github.com/wensboy/lightpc/backend/domain/engine/notice"
	"github.com/wensboy/lightpc/backend/internal/ebus"
)

// newEngine builds an engine with the given ring length and a bus a test can observe.
func newEngine(t *testing.T, buf int) (*Notice, domainBus.Bus) {
	t.Helper()

	bus := ebus.NewEventBus()
	engine := NewNotice(WithNoticeSize(buf))
	engine.ectx = domainEngine.EngineContext{Ebus: bus}
	engine.ring = make([]engineNotice.Record, 0, buf)
	engine.state = domainEngine.EngineReady
	return engine, bus
}

func TestNotifyStoresNewestFirst(t *testing.T) {
	engine, _ := newEngine(t, 8)

	first, err := engine.Notify(engineNotice.Input{Title: "first", Source: "test"})
	if err != nil {
		t.Fatalf("notify: %v", err)
	}
	if first.ID == 0 || first.Timestamp == 0 {
		t.Fatalf("notice = %+v", first)
	}
	// An empty level defaults to info rather than being rejected: a caller raising a plain notice should not have to
	// spell the severity out.
	if first.Level != engineNotice.LevelInfo {
		t.Fatalf("level = %q", first.Level)
	}

	if _, err := engine.Notify(engineNotice.Input{Title: "second", Source: "test"}); err != nil {
		t.Fatalf("notify: %v", err)
	}

	history := engine.History()
	if len(history) != 2 {
		t.Fatalf("history = %d", len(history))
	}
	if history[0].Title != "second" || history[1].Title != "first" {
		t.Fatalf("history = %+v", history)
	}
}

func TestNotifyDropsTheOldestAtAFullRing(t *testing.T) {
	engine, _ := newEngine(t, 3)

	for _, title := range []string{"one", "two", "three", "four"} {
		if _, err := engine.Notify(engineNotice.Input{Title: title}); err != nil {
			t.Fatalf("%s: %v", title, err)
		}
	}

	history := engine.History()
	if len(history) != 3 {
		t.Fatalf("history = %d", len(history))
	}
	// The ring is fixed: the oldest entry gives way, so "one" is gone and the order of the rest is unchanged.
	if history[0].Title != "four" || history[2].Title != "two" {
		t.Fatalf("history = %+v", history)
	}
}

func TestNotifyAnnouncesOnTheBus(t *testing.T) {
	engine, bus := newEngine(t, 4)

	var received []uint64
	bus.On(domainBus.EventNoticeSend, func(payload any) {
		notice, ok := payload.(engineNotice.Record)
		if !ok {
			t.Fatalf("payload = %T", payload)
		}
		received = append(received, notice.ID)
	})

	sent, err := engine.Notify(engineNotice.Input{Title: "announce", Level: engineNotice.LevelWarning})
	if err != nil {
		t.Fatalf("notify: %v", err)
	}

	if len(received) != 1 || received[0] != sent.ID {
		t.Fatalf("received = %v, sent = %d", received, sent.ID)
	}
}

func TestNotifyRejectsIncompleteInput(t *testing.T) {
	engine, _ := newEngine(t, 4)

	if _, err := engine.Notify(engineNotice.Input{Content: "no title"}); err == nil {
		t.Fatal("a notice without a title must be rejected")
	}
	if _, err := engine.Notify(engineNotice.Input{Title: "x", Level: "fatal"}); err == nil {
		t.Fatal("an unknown level must be rejected")
	}
}

func TestRemoveAndClear(t *testing.T) {
	engine, _ := newEngine(t, 4)

	kept, _ := engine.Notify(engineNotice.Input{Title: "kept"})
	dropped, _ := engine.Notify(engineNotice.Input{Title: "dropped"})

	if !engine.Remove(dropped.ID) {
		t.Fatal("remove must report the notice it dropped")
	}
	if engine.Remove(dropped.ID) {
		t.Fatal("removing the same id twice must report false")
	}
	history := engine.History()
	if len(history) != 1 || history[0].ID != kept.ID {
		t.Fatalf("history = %+v", history)
	}

	if removed := engine.Clear(); removed != 1 {
		t.Fatalf("clear removed %d", removed)
	}
	if history := engine.History(); len(history) != 0 {
		t.Fatalf("history = %d", len(history))
	}
}
