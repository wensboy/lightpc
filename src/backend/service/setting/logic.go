// Package setting exposes the settings document to the front end.
package serviceSettingImpl

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineSetting "github.com/wensboy/lightpc/backend/domain/engine/setting"
	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	serviceSetting "github.com/wensboy/lightpc/backend/domain/service/setting"
	"github.com/wensboy/lightpc/backend/internal/cast"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// NewSettingService builds an uninitialised service.
func NewSettingService() *SettingService {
	return &SettingService{schemaPath: defaultSchemaPath}
}

// SettingService hands the settings schema and its aggregated values to the front end.
type SettingService struct {
	fileStore  engineStore.FileStore
	setting    engineSetting.Setting
	schemaPath string
	appCtx     *domainContext.AppContext
	// bus is the service bus. An ordinary setting is announced here rather than on the engine bus, so a change that
	// only concerns services and the interface does not wake every engine.
	bus domainBus.Bus
}

// Group implements domainService.Service.
func (s *SettingService) Group() string { return domainService.GroupSetting }

// Init implements domainService.Service: it resolves the file store and the settings engine.
func (s *SettingService) Init(sCtx domainService.ServiceContext) error {
	storeEngine, found := sCtx.Engines[domainEngine.EngineDep(domainEngine.GroupStore, domainEngine.NameFs)]
	if !found {
		return interError.Retrieve(engineStore.ErrKeyInvalid)
	}
	fileStore, ok := cast.As[engineStore.FileStore](storeEngine)
	if !ok {
		return interError.Retrieve(engineStore.ErrKeyInvalid)
	}
	s.fileStore = fileStore

	settingKey := domainEngine.EngineDep(domainEngine.GroupSetting, domainEngine.NameSetting)
	settingEngine, found := sCtx.Engines[settingKey]
	if !found {
		return interError.Retrieve(engineSetting.ErrKeyLayerUnknown).Wrap(fmt.Errorf("engine %q is not ready", settingKey))
	}
	instance, ok := cast.As[engineSetting.Setting](settingEngine)
	if !ok {
		return interError.Retrieve(engineSetting.ErrKeyLayerUnknown).Wrap(fmt.Errorf("engine %q is not a setting engine", settingKey))
	}
	s.setting = instance

	s.appCtx = sCtx.AppCtx
	s.bus = sCtx.Ebus

	// The schema is a plain document, so its location comes from the configuration rather than a literal.
	if sCtx.AppCtx != nil && sCtx.AppCtx.ConfigContext != nil {
		if value, found := sCtx.AppCtx.ConfigContext.Lookup("default.runtime.engines.setting.schema"); found {
			if configured, ok := cast.To[string](value); ok && configured != "" {
				s.schemaPath = configured
			}
		}
	}
	return nil
}

// Deinit implements domainService.Service.
func (s *SettingService) Deinit() error { return nil }

// Payload returns the schema and the aggregated values the front end renders into setting cards.
func (s *SettingService) Payload() (serviceSetting.SettingPayloadDto, error) {
	if s.fileStore == nil || s.setting == nil {
		return serviceSetting.SettingPayloadDto{}, interError.Retrieve(engineStore.ErrKeyInvalid)
	}

	data, err := s.fileStore.Get(context.Background(), s.schemaPath)
	if err != nil {
		return serviceSetting.SettingPayloadDto{}, err
	}
	schema := map[string]any{}
	if err := json.Unmarshal(data, &schema); err != nil {
		return serviceSetting.SettingPayloadDto{}, interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(err)
	}

	flat := map[string]any{}
	for _, key := range s.setting.Keys() {
		raw, found := s.setting.Resolve(key)
		if !found {
			continue
		}
		entry, ok := cast.As[engineSetting.Entry[any]](raw)
		if !ok {
			continue
		}
		flat[key] = entry.Value
	}
	defaults := s.defaults()
	s.pruneBuiltinFonts(flat)
	mergeThemePalettes(flat, defaults)
	// The palette list stays in the view because theme.default reads its options from it. It is marked as
	// derived in the schema, so the renderer uses it for the picker without drawing a card of its own; the
	// Themes field carries the same data already resolved for painting.
	return serviceSetting.SettingPayloadDto{
		Schema:      schema,
		Values:      nestValues(flat),
		Defaults:    nestValues(defaults),
		SystemFonts: SystemFontFamilies(),
		Themes:      buildThemes(flat),
	}, nil
}

// Save writes one value into the runtime layer and flushes it in the background, so the UI update is never
// blocked by disk IO. A failing flush is retried by the engine's periodic flush.
func (s *SettingService) Save(key string, value any, kind serviceSetting.Kind) error {
	if s.setting == nil {
		return interError.Retrieve(engineSetting.ErrKeyLayerUnknown)
	}
	if !kind.Known() {
		return interError.Retrieve(engineSetting.ErrKeyLayerUnknown).
			Wrap(fmt.Errorf("save %q: unknown kind %q", key, kind))
	}

	// The engine is asked with the entry point that matches the destination: an engine kind is announced on the engine
	// bus by the engine itself, a service kind is announced here, after the write succeeded. The engine never learns
	// which bus a change belongs to.
	var err error
	if kind == serviceSetting.KindEngine {
		err = s.setting.SetAndNotify(runtimeLayerName, key, value)
	} else {
		err = s.setting.Set(runtimeLayerName, key, value)
	}
	if err != nil {
		return err
	}
	if kind == serviceSetting.KindService {
		s.announce(key, value)
	}
	instance := s.setting
	go func() {
		if err := instance.Flush(); err != nil {
			slog.Default().Warn("setting flush failed", "key", key, "err", err)
		}
	}()
	return nil
}

// pruneBuiltinFonts drops any declared family that has no matching file in the bundle. The picker can
// therefore never offer a font that would fail to load, and a stale entry is reported instead of hidden.
func (s *SettingService) pruneBuiltinFonts(flat map[string]any) {
	raw, found := flat[builtinFontKey]
	if !found {
		return
	}
	list, ok := cast.To[[]any](raw)
	if !ok {
		return
	}
	bundled := s.bundledFamilies()
	kept := make([]any, 0, len(list))
	for _, entry := range list {
		family, ok := cast.To[string](entry)
		if !ok {
			continue
		}
		if _, found := bundled[normaliseFamily(family)]; !found {
			slog.Default().Warn("setting builtin font has no file", "family", family)
			continue
		}
		kept = append(kept, family)
	}
	flat[builtinFontKey] = kept
}

// defaults collects what the default layer contributes, so the renderer has an authored value to fall back
// to when a hand edited user value breaks the schema.
func (s *SettingService) defaults() map[string]any {
	flat := map[string]any{}
	for _, key := range s.setting.Keys() {
		for _, raw := range s.setting.Layered(key) {
			entry, ok := cast.As[engineSetting.Entry[any]](raw)
			if !ok || entry.Layer != defaultLayerName {
				continue
			}
			flat[key] = entry.Value
		}
	}
	return flat
}

// announce publishes an ordinary setting change on the service bus.
//
// The payload is the service's own type, not the engine's, even though the fields match: the two buses are isolated, so
// an alias would let a change made for one side leak into the other. A listener casts to the type of the bus it
// subscribed to, and that is the whole contract between them.
func (s *SettingService) announce(key string, value any) {
	if s.bus == nil {
		return
	}
	s.bus.Emit(domainBus.EventSettingChanged, serviceSetting.SettingChangedPayload{Key: key, Value: value})
}

// mergeThemePalettes combines the shipped palettes with the user ones. A user palette with a new name is
// appended; an existing name is merged key by key, so a user file can retune one colour without restating
// the whole palette.
func mergeThemePalettes(flat, defaults map[string]any) {
	shipped, ok := cast.To[[]any](defaults[themePalettesKey])
	if !ok {
		return
	}

	merged := slices.Clone(shipped)
	position := make(map[string]int, len(shipped))
	for index, entry := range merged {
		item, ok := cast.To[map[string]any](entry)
		if !ok {
			continue
		}
		if name, ok := cast.To[string](item["name"]); ok && name != "" {
			position[name] = index
		}
	}

	custom, ok := cast.To[[]any](flat[themePalettesKey])
	if !ok {
		flat[themePalettesKey] = merged
		return
	}
	for _, entry := range custom {
		item, ok := cast.To[map[string]any](entry)
		if !ok {
			continue
		}
		name, ok := cast.To[string](item["name"])
		if !ok || name == "" {
			continue
		}
		if index, found := position[name]; found {
			base, _ := cast.To[map[string]any](merged[index])
			merged[index] = mergeTheme(base, item)
			continue
		}
		position[name] = len(merged)
		merged = append(merged, item)
	}
	flat[themePalettesKey] = merged
}

// mergeTheme folds a user palette into a shipped one, one colour mode and one key at a time. Keys the user
// did not mention keep their shipped value.
func mergeTheme(base, override map[string]any) map[string]any {
	result := make(map[string]any, len(base))
	maps.Copy(result, base)

	for _, mode := range []string{"light", "dark"} {
		basePalette, _ := cast.To[map[string]any](base[mode])
		overridePalette, _ := cast.To[map[string]any](override[mode])
		if len(overridePalette) == 0 {
			continue
		}
		folded := make(map[string]any, len(basePalette)+len(overridePalette))
		maps.Copy(folded, basePalette)
		maps.Copy(folded, overridePalette)
		result[mode] = folded
	}
	return result
}

// nestValues rebuilds the nested document the renderer expects from the engine's dotted keys.
func nestValues(flat map[string]any) map[string]any {
	document := map[string]any{}
	for key, value := range flat {
		segments := strings.Split(key, ".")
		cursor := document
		for _, segment := range segments[:len(segments)-1] {
			child, ok := cursor[segment].(map[string]any)
			if !ok {
				child = map[string]any{}
				cursor[segment] = child
			}
			cursor = child
		}
		cursor[segments[len(segments)-1]] = value
	}
	return document
}
