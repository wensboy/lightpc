package bootstrap

import (
	"log/slog"
	"strings"

	"fmt"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
	domainApp "github.com/wensboy/lightpc/backend/domain/app"
	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	engineLog "github.com/wensboy/lightpc/backend/domain/engine/log"
	engineNotice "github.com/wensboy/lightpc/backend/domain/engine/notice"
	domainEngineSetting "github.com/wensboy/lightpc/backend/domain/engine/setting"
	domainEngineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	engineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	domainRuntime "github.com/wensboy/lightpc/backend/domain/runtime"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	engineLogImpl "github.com/wensboy/lightpc/backend/engine/log"
	engineNoticeImpl "github.com/wensboy/lightpc/backend/engine/notice"
	engineSettingImpl "github.com/wensboy/lightpc/backend/engine/setting"
	engineStoreImpl "github.com/wensboy/lightpc/backend/engine/store"
	"github.com/wensboy/lightpc/backend/internal/cast"
	"github.com/wensboy/lightpc/backend/internal/config"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

func SetupConfig(eCtx domainContext.EmbedContext) *config.Config {
	fileSource := config.NewFileSource()
	fileSource.Parse(domainContext.FsContext{
		Paths: []string{
			"data/conf/default.json",
			"data/conf/errors.json",
		},
		Fs: eCtx["/data"],
	})
	envSource := config.NewEnvSource().WithPrefix("LIGHTPC_")
	envSource.Export(domainContext.FsContext{
		Paths: []string{
			"data/.env",
		},
		Fs: eCtx["/data"],
	})
	flagSource := config.NewFlagSource()
	err := flagSource.ParseCommand(domainContext.FsContext{
		Paths: []string{
			"data/conf/command.json",
		},
		Fs: eCtx["/data"],
	})
	if err != nil {
		panic(err)
	}
	return config.NewConfig().WithSources(flagSource, envSource, fileSource)
}

func SetupApp(cfg domainContext.ConfigContext, opt application.Options) *application.App {
	if value, ok := cfg.Lookup("default.application.name"); ok {
		if name, ok := cast.As[string](value); ok {
			opt.Name = name
		}
	}
	if value, ok := cfg.Lookup("default.application.description"); ok {
		if description, ok := cast.As[string](value); ok {
			opt.Description = description
		}
	}
	var enableDevTools, disableDefaultCtxMenu bool
	if value, ok := cfg.Lookup("default.application.enableDevTools"); ok {
		if v, ok := cast.As[bool](value); ok {
			enableDevTools = v
		}
	}
	if value, ok := cfg.Lookup("default.application.disableDefaultContextMenu"); ok {
		if v, ok := cast.As[bool](value); ok {
			disableDefaultCtxMenu = v
		}
	}
	app := application.New(opt)
	windowOptions := application.WebviewWindowOptions{
		Title: opt.Name,
		// A release build disables both by default, which leaves no way to inspect the interface.
		DevToolsEnabled:            enableDevTools,
		DefaultContextMenuDisabled: disableDefaultCtxMenu,
	}
	if value, ok := cfg.Lookup("default.application.window.width"); ok {
		if width, ok := cast.As[int](value); ok {
			windowOptions.Width = width
		}
	}
	if value, ok := cfg.Lookup("default.application.window.height"); ok {
		if height, ok := cast.As[int](value); ok {
			windowOptions.Height = height
		}
	}
	app.Window.NewWithOptions(windowOptions)
	return app
}

func SetupErrors(cfg domainContext.ConfigContext) domainContext.ErrorContext {
	errorContext := make(domainContext.ErrorContext)
	value, ok := cfg.Lookup("errors")
	if !ok {
		return errorContext
	}
	groups, ok := cast.To[map[string]any](value)
	if !ok {
		return errorContext
	}
	for groupName, group := range groups {
		definition, ok := cast.To[map[string]any](group)
		if !ok {
			continue
		}
		items, ok := cast.To[[]any](definition["definition"])
		if !ok {
			continue
		}
		for _, item := range items {
			record, ok := cast.To[map[string]any](item)
			if !ok {
				continue
			}
			code, codeOK := cast.To[string](record["code"])
			kind, kindOK := cast.To[string](record["kind"])
			if !codeOK || !kindOK {
				continue
			}
			errorContext[code] = kind
			if groupName == "root" {
				interError.Register(
					interError.KeyInternalPrefix+strings.ToLower(strings.Join(strings.Fields(kind), "_")),
					interError.New(interError.Option{Code: code, Kind: kind, Msg: kind}),
				)
			}
		}
	}
	return errorContext
}

// SetupStore registers the file store.
//
// Preparing the root directory and checking that it is usable are bootstrap decisions, not engine phases:
// they run in the pre start hook so the engine only has to report the state it was initialised with.
func SetupStore(cfg domainContext.ConfigContext, er *domainRuntime.EngineRuntime) {
	settings := domainContext.Decode[domainEngineStore.Config](cfg, "default.runtime.engines.store")

	fsOptions := make([]engineStoreImpl.FSOption, 0, 3)
	if settings.Root != "" {
		if resource, err := config.NewResource(settings.Root); err == nil && !resource.IsEmbed() {
			if resolved, err := resource.Path(); err == nil {
				fsOptions = append(fsOptions, engineStoreImpl.WithRoot(resolved))
			}
		}
	}
	if mode := engineStore.ParseMode(settings.Fs.FileMode); mode != 0 {
		fsOptions = append(fsOptions, engineStoreImpl.WithFileMode(mode))
	}
	if mode := engineStore.ParseMode(settings.Fs.DirMode); mode != 0 {
		fsOptions = append(fsOptions, engineStoreImpl.WithDirMode(mode))
	}
	fsStore := engineStoreImpl.NewFS(fsOptions...)
	er.Register(&domainEngine.EngineHook{
		Engine: fsStore,
		PreStart: func(e domainEngine.Engine) error {
			store, ok := cast.As[*engineStoreImpl.FS](e)
			if !ok {
				return interError.Retrieve(domainEngineStore.ErrKeyInvalid).
					Wrap(fmt.Errorf("store hook: unexpected engine %T", e))
			}
			ctx := store.HookContext()
			if ctx.Root != "" && ctx.DirMode != 0 {
				if err := os.MkdirAll(ctx.Root, ctx.DirMode); err != nil {
					return interError.Retrieve(domainEngineStore.ErrKeyInvalid).
						Wrap(fmt.Errorf("create root: %w", err))
				}
			}
			return store.Validate()
		},
	})

	// Databases are opened and migrated from the same hook. One engine owns every connection, keyed by
	// "driver.name", so a caller resolves a database without knowing which driver backs it.
	sqlStore := engineStoreImpl.NewSQL()
	type pending struct {
		dsn        string
		driver     domainEngineStore.Driver
		migrations []string
	}
	queue := make([]pending, 0, 2)

	for _, config := range engineStoreImpl.DBConfigs(settings.Db.SQL) {
		key := domainEngineStore.DBKey(config.Driver, config.Name)
		// Checked before opening: for a file backed driver, opening is what creates the file, so a
		// duplicate entry must be rejected while it still has no side effect.
		if sqlStore.HasKey(key) {
			slog.Default().Warn("db duplicate key skipped", "key", key)
			continue
		}

		database, err := engineStoreImpl.OpenDB(config)
		if err != nil {
			slog.Default().Warn("db open failed", "name", config.Name, "err", err)
			continue
		}
		if !sqlStore.Register(key, config.Driver, database) {
			// A concurrent registration would land here; close what was opened so it cannot leak.
			slog.Default().Warn("db duplicate key skipped", "key", key)
			_ = database.Close()
			continue
		}
		dsn, err := config.DSNString()
		if err != nil {
			slog.Default().Warn("db migrations skipped", "name", config.Name, "err", err)
			continue
		}
		queue = append(queue, pending{dsn: dsn, driver: config.Driver, migrations: config.Migrations})
	}

	er.Register(&domainEngine.EngineHook{
		Engine: sqlStore,
		PreStart: func(e domainEngine.Engine) error {
			store, ok := cast.As[*engineStoreImpl.SQL](e)
			if !ok {
				return interError.Retrieve(domainEngineStore.ErrKeyInvalid).
					Wrap(fmt.Errorf("db hook: unexpected engine %T", e))
			}
			// The entry point mounts the data directory, so every migration resource is a path below it.
			appCtx := store.HookContext().EngineContext.AppCtx
			if appCtx == nil || appCtx.EmbedContext["/data"] == nil {
				slog.Default().Warn("db migrations skipped", "reason", "no embedded tree")
				return nil
			}
			assets := appCtx.EmbedContext["/data"]
			for index, item := range queue {
				if err := engineStoreImpl.Migrate(item.dsn, item.driver, item.migrations, assets); err != nil {
					return fmt.Errorf("migrate database %d: %w", index, err)
				}
			}
			return nil
		},
	})
}

func SetupLog(cfg domainContext.ConfigContext, er *domainRuntime.EngineRuntime) {
	// hookLog wraps a log engine so its output directories are created before it starts. Where a log file lives
	// is a bootstrap decision, so it stays here instead of inside the engine phases.
	hookLog := func(engine domainEngine.Engine, entries []engineLog.LogEntry) domainEngine.Engine {
		return &domainEngine.EngineHook{
			Engine: engine,
			PreStart: func(_ domainEngine.Engine) error {
				return engineLogImpl.PrepareOutputs(entries)
			},
		}
	}
	engines := make([]domainEngine.Engine, 0, 3)

	// Each sub-group names one engine. The contract package reads the records, so this function only decides which
	// groups to start and hands the entries over.
	logSettings := domainContext.Decode[engineLog.Config](cfg, engineLog.EntriesKey)
	if entries := logSettings.Entries(engineLog.SubDirect); len(entries) > 0 {
		options := make([]engineLogImpl.DirectOption, 0, len(entries))
		for _, entry := range entries {
			options = append(options, engineLogImpl.WithDirectEntry(entry))
		}
		engines = append(engines, hookLog(engineLogImpl.NewDirectLog(options...), entries))
	}

	if entries := logSettings.Entries(engineLog.SubFanout); len(entries) > 0 {
		options := make([]engineLogImpl.FanoutOption, 0, len(entries))
		for _, entry := range entries {
			options = append(options, engineLogImpl.WithFanoutEntry(entry))
		}
		engines = append(engines, hookLog(engineLogImpl.NewFanoutLog(options...), entries))
	}

	if entries := logSettings.Entries(engineLog.SubPattern); len(entries) > 0 {
		options := make([]engineLogImpl.PatternOption, 0, len(entries))
		for _, entry := range entries {
			options = append(options, engineLogImpl.WithPatternEntry(entry))
		}
		engines = append(engines, hookLog(engineLogImpl.NewPatternLog(options...), entries))
	}

	er.Register(engines...)
}

func SetupSetting(cfg domainContext.ConfigContext, er *domainRuntime.EngineRuntime) {
	settingSettings := domainContext.Decode[domainEngineSetting.Config](cfg, "default.runtime.engines.setting")
	layers := settingSettings.GenLayers()
	options := make([]engineSettingImpl.Option, 0, len(layers))
	for _, layer := range layers {
		options = append(options, engineSettingImpl.WithLayer(layer))
	}

	settingEngine := engineSettingImpl.NewSetting(options...)
	er.Register(&domainEngine.EngineHook{
		Engine: settingEngine,
		PreStart: func(e domainEngine.Engine) error {
			instance, ok := cast.As[*engineSettingImpl.Setting](e)
			if !ok {
				return interError.Retrieve(domainEngineSetting.ErrKeyLayerUnknown).
					Wrap(fmt.Errorf("setting hook: unexpected engine %T", e))
			}
			// Creating a layer file and reading it are bootstrap decisions, so they happen here instead of
			// inside Init.
			return instance.Prepare()
		},
	})
}

// SetupEventHandle defines the event wiring of the three runtimes.
//
// ar owns the events that make the interface render again and forwards them to the webview. er keeps
// engine traffic inside the engine bus and restarts a failed engine. sr keeps service traffic inside the
// service bus. All three share one bus, so an event published by one layer is visible to the others.
func SetupEventHandle(ar *domainRuntime.AppRuntime, er *domainRuntime.EngineRuntime, sr *domainRuntime.ServiceRuntime) {
	ar.WithEventFn(func(aec domainApp.AppEventContext) {
		if aec.App == nil {
			return
		}
		for _, name := range domainBus.UiEvents {
			aec.Ebus.On(name, func(payload any) {
				aec.App.Event.Emit(name, payload)
			})
		}
	})

	er.WithEventFn(func(eec domainEngine.EngineEventContext) {
		for _, name := range domainBus.EngineOpaqueEvents {
			eec.Ebus.On(name, func(payload any) {
				if eec.AppEbus == nil {
					return
				}
				eec.AppEbus.Emit(name, payload)
			})
		}
		eec.Ebus.On(domainBus.EventEngineError, func(payload any) {
			// One assertion, the same shape the setting handlers use: the emitter sends a named payload, so a listener
			// casts to it instead of decoding a map field by field.
			event, ok := cast.As[domainEngine.EngineStatePayload](payload)
			if !ok {
				return
			}
			engine, ok := eec.Engines[domainEngine.EngineDep(event.Group, event.Name)]
			if !ok {
				return
			}
			_ = engine.Restart()
		})
	})

	sr.WithEventFn(func(sec domainService.ServiceEventContext) {
		// Services already act on their own events, but an expiry has to reach the interface. The service bus
		// is a different bus from the app one, so the event is re-published there; the app runtime forwards
		// every name in UiEvents to the webview, which is where the layout listens.
		for _, name := range domainBus.ServiceOpaqueEvents {
			sec.Ebus.On(name, func(payload any) {
				if sec.AppEbus == nil {
					return
				}
				sec.AppEbus.Emit(name, payload)
			})
		}

	})
}

// SetupNotice registers the notice engine.
//
// The ring length is read here, not inside the engine: the document is the bootstrap's business, and the engine only
// receives the number. A value that is absent or unusable is left out, so the engine falls back to its own default.
func SetupNotice(cfg domainContext.ConfigContext, er *domainRuntime.EngineRuntime) {
	noticeSettings := domainContext.Decode[engineNotice.Config](cfg, "default.runtime.engines.notice")
	options := make([]engineNoticeImpl.NoticeOption, 0, 1)
	if noticeSettings.Buf > 0 {
		options = append(options, engineNoticeImpl.WithNoticeSize(noticeSettings.Buf))
	}
	er.Register(engineNoticeImpl.NewNotice(options...))
}
