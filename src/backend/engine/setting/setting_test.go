package engineSettingImpl

import (
	"errors"
	"testing"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineSetting "github.com/wensboy/lightpc/backend/domain/engine/setting"
	"github.com/wensboy/lightpc/backend/internal/cast"
	"github.com/wensboy/lightpc/backend/internal/ebus"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// entryOf casts one untyped contribution back into the shape the engine stores.
func entryOf(t *testing.T, raw any) engineSetting.Entry[any] {
	t.Helper()
	entry, ok := cast.As[engineSetting.Entry[any]](raw)
	if !ok {
		t.Fatalf("contribution must be a setting entry, got %T", raw)
	}
	return entry
}

func newTestSetting(t *testing.T) *Setting {
	t.Helper()
	s := NewSetting(
		WithLayer(engineSetting.Layer{Name: "base", Mode: engineSetting.ModeReadOnly}),
		WithLayer(engineSetting.Layer{Name: "user", Mode: engineSetting.ModeReadWrite}),
	)
	if err := s.Init(domainEngine.EngineContext{}); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSettingLayerPriority(t *testing.T) {
	s := newTestSetting(t)

	if err := s.Load("base", map[string]any{"theme": "light"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("user", "theme", "dark"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("user", "language", "en"); err != nil {
		t.Fatal(err)
	}

	raw, ok := s.Resolve("theme")
	if !ok {
		t.Fatal("theme must resolve")
	}
	entry := entryOf(t, raw)
	if entry.Value != "dark" || entry.Layer != "user" {
		t.Fatalf("entry = %v", entry)
	}

	layered := s.Layered("theme")
	if len(layered) != 2 {
		t.Fatalf("layered = %v", layered)
	}
	if entryOf(t, layered[0]).Layer != "base" || entryOf(t, layered[1]).Layer != "user" {
		t.Fatalf("layered = %v", layered)
	}

	if _, ok := s.Resolve("missing"); ok {
		t.Fatal("an unknown key must not resolve")
	}
}

func TestSettingRejectsWritesToReadOnlyLayer(t *testing.T) {
	s := newTestSetting(t)

	err := s.Set("base", "theme", "light")
	if !errors.Is(err, interError.Retrieve(engineSetting.ErrKeyLayerReadOnly)) {
		t.Fatalf("err = %v", err)
	}
}

func TestSettingRejectsUnknownLayer(t *testing.T) {
	s := newTestSetting(t)

	if err := s.Set("ghost", "theme", "light"); !errors.Is(err, interError.Retrieve(engineSetting.ErrKeyLayerUnknown)) {
		t.Fatalf("set err = %v", err)
	}
	if err := s.Load("ghost", map[string]any{"theme": "light"}); !errors.Is(err, interError.Retrieve(engineSetting.ErrKeyLayerUnknown)) {
		t.Fatalf("load err = %v", err)
	}
}

func TestSettingLoadReplacesLayer(t *testing.T) {
	s := newTestSetting(t)

	if err := s.Set("user", "theme", "dark"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set("user", "language", "en"); err != nil {
		t.Fatal(err)
	}
	if err := s.Load("user", map[string]any{"theme": "light"}); err != nil {
		t.Fatal(err)
	}

	if _, ok := s.Resolve("language"); ok {
		t.Fatal("load must replace the whole layer")
	}
	raw, ok := s.Resolve("theme")
	if !ok || entryOf(t, raw).Value != "light" {
		t.Fatalf("entry = %v, %v", raw, ok)
	}
}

func TestSettingNotifiesChanges(t *testing.T) {
	bus := ebus.NewEventBus()
	changed := make(chan any, 4)
	bus.On(domainBus.EventSettingChanged, func(payload any) { changed <- payload })

	s := NewSetting(WithLayer(engineSetting.Layer{Name: "user", Mode: engineSetting.ModeReadWrite}))
	if err := s.Init(domainEngine.EngineContext{Ebus: bus}); err != nil {
		t.Fatal(err)
	}
	// Set writes and says nothing: whether a change should reach a bus depends on the kind of setting, which only the
	// caller knows. SetAndNotify is the entry point that also announces.
	if err := s.Set("user", "theme", "dark"); err != nil {
		t.Fatal(err)
	}
	select {
	case payload := <-changed:
		t.Fatalf("Set must not publish, got %v", payload)
	default:
	}

	if err := s.SetAndNotify("user", "theme", "dark"); err != nil {
		t.Fatal(err)
	}
	select {
	case payload := <-changed:
		change, ok := cast.As[engineSetting.SettingChangedPayload](payload)
		if !ok {
			t.Fatalf("payload = %v", payload)
		}
		if change.Key != "theme" || change.Value != "dark" {
			t.Fatalf("payload = %+v", change)
		}
	default:
		t.Fatal("SetAndNotify must publish")
	}
}

func TestSettingLifecycle(t *testing.T) {
	s := newTestSetting(t)

	if s.Group() != domainEngine.GroupSetting || s.Name() != domainEngine.NameSetting {
		t.Fatalf("identity = %s.%s", s.Group(), s.Name())
	}
	if s.State() != domainEngine.EngineReady {
		t.Fatalf("state = %v", s.State())
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	if s.State() != domainEngine.EngineRunning {
		t.Fatalf("state = %v", s.State())
	}
	if err := s.Stop(); err != nil {
		t.Fatal(err)
	}
	if s.State() != domainEngine.EngineStopped {
		t.Fatalf("state = %v", s.State())
	}
	if err := s.Restart(); err != nil {
		t.Fatal(err)
	}
	if s.State() != domainEngine.EngineReady {
		t.Fatalf("state = %v", s.State())
	}
}
