// Package setting exposes the settings document to the front end.
package setting

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

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

// SettingService hands the settings schema and its aggregated values to the front end.
type SettingService struct {
	fileStore  store.FileStore
	setting    setting.Setting
	schemaPath string
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
	return domainService.SettingPayload{
		Schema:   schema,
		Values:   nestValues(flat),
		Defaults: nestValues(s.defaults()),
	}, nil
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
