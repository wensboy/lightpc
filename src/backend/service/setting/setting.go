// Package setting exposes the settings document to the front end.
package setting

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	"github.com/wensboy/lightpc/backend/domain/engine/setting"
	"github.com/wensboy/lightpc/backend/domain/engine/store"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	"github.com/wensboy/lightpc/backend/internal/cast"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// defaultSchemaPath is used when the configuration does not name a schema resource.
const defaultSchemaPath = "embed://data/conf/schema/setting.schema.json"

// schemaKey is the configuration path that names the schema resource.
const schemaKey = "default.runtime.setting.schema"

// runtimeLayerName is the in-memory layer UI edits land in before they are flushed to the user layer.
const runtimeLayerName = "runtime"

// defaultLayerName is the read only layer shipped with the application.
const defaultLayerName = "default"

// builtinFontKey addresses the bundled font list, which is filtered against the shipped files.
const builtinFontKey = "appearence.font.builtin"

// themePalettesKey addresses the palette list shared by every theme picker.
const themePalettesKey = "appearence.theme.palettes"

// SettingService hands the settings schema and its aggregated values to the front end.
type SettingService struct {
	fileStore  store.FileStore
	setting    setting.Setting
	schemaPath string
	appCtx     *domainContext.AppContext
}

// NewSettingService builds an uninitialised service.
func NewSettingService() *SettingService {
	return &SettingService{schemaPath: defaultSchemaPath}
}

// Group implements domainService.Service.
func (s *SettingService) Group() string { return domainService.GroupSetting }

// Init implements domainService.Service: it resolves the file store and the settings engine.
func (s *SettingService) Init(sCtx domainService.ServiceContext) error {
	storeEngine, found := sCtx.Engines[domainEngine.EngineDep(store.GroupStore, store.NameFs)]
	if !found {
		return interError.Retrieve(store.ErrKeyInvalid)
	}
	fileStore, ok := cast.As[store.FileStore](storeEngine)
	if !ok {
		return interError.Retrieve(store.ErrKeyInvalid)
	}
	s.fileStore = fileStore

	settingKey := domainEngine.EngineDep(setting.GroupSetting, setting.NameSetting)
	settingEngine, found := sCtx.Engines[settingKey]
	if !found {
		return interError.Retrieve(setting.ErrKeyLayerUnknown).Wrap(fmt.Errorf("engine %q is not ready", settingKey))
	}
	instance, ok := cast.As[setting.Setting](settingEngine)
	if !ok {
		return interError.Retrieve(setting.ErrKeyLayerUnknown).Wrap(fmt.Errorf("engine %q is not a setting engine", settingKey))
	}
	s.setting = instance

	s.appCtx = sCtx.AppCtx

	// The schema is a plain document, so its location comes from the configuration rather than a literal.
	if sCtx.AppCtx != nil && sCtx.AppCtx.ConfigContext != nil {
		if value, found := sCtx.AppCtx.ConfigContext.Lookup(schemaKey); found {
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
func (s *SettingService) Payload() (domainService.SettingPayload, error) {
	if s.fileStore == nil || s.setting == nil {
		return domainService.SettingPayload{}, interError.Retrieve(store.ErrKeyInvalid)
	}

	data, err := s.fileStore.Get(context.Background(), s.schemaPath)
	if err != nil {
		return domainService.SettingPayload{}, err
	}
	schema := map[string]any{}
	if err := json.Unmarshal(data, &schema); err != nil {
		return domainService.SettingPayload{}, interError.Retrieve(store.ErrKeyInvalid).Wrap(err)
	}

	flat := map[string]any{}
	for _, key := range s.setting.Keys() {
		raw, found := s.setting.Resolve(key)
		if !found {
			continue
		}
		entry, ok := cast.As[setting.Entry[any]](raw)
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
	return domainService.SettingPayload{
		Schema:      schema,
		Values:      nestValues(flat),
		Defaults:    nestValues(defaults),
		SystemFonts: SystemFontFamilies(),
		Themes:      buildThemes(flat),
	}, nil
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
	for key, value := range base {
		result[key] = value
	}

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

// defaults collects what the default layer contributes, so the renderer has an authored value to fall back
// to when a hand edited user value breaks the schema.
func (s *SettingService) defaults() map[string]any {
	flat := map[string]any{}
	for _, key := range s.setting.Keys() {
		for _, raw := range s.setting.Layered(key) {
			entry, ok := cast.As[setting.Entry[any]](raw)
			if !ok || entry.Layer != defaultLayerName {
				continue
			}
			flat[key] = entry.Value
		}
	}
	return flat
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

// Save writes one value into the runtime layer and flushes it in the background, so the UI update is never
// blocked by disk IO. A failing flush is retried by the engine's periodic flush.
func (s *SettingService) Save(key string, value any) error {
	if s.setting == nil {
		return interError.Retrieve(setting.ErrKeyLayerUnknown)
	}
	if err := s.setting.Set(runtimeLayerName, key, value); err != nil {
		return err
	}
	instance := s.setting
	go func() {
		if err := instance.Flush(); err != nil {
			slog.Default().Warn("setting flush failed", "key", key, "err", err)
		}
	}()
	return nil
}
