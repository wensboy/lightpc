package bootstrap

import (
	"log/slog"
	"strings"

	"fmt"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
	domainApp "github.com/wensboy/lightpc/backend/domain/app"
	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	domainEngineSetting "github.com/wensboy/lightpc/backend/domain/engine/setting"
	domainEngineStore "github.com/wensboy/lightpc/backend/domain/engine/store"
	domainRuntime "github.com/wensboy/lightpc/backend/domain/runtime"
	domainService "github.com/wensboy/lightpc/backend/domain/service"
	engineLogImpl "github.com/wensboy/lightpc/backend/engine/log"
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
	fsStore := engineStoreImpl.NewFS()
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

	for _, config := range engineStoreImpl.DBConfigs(cfg) {
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
		queue = append(queue, pending{dsn: config.DSNString(), driver: config.Driver, migrations: config.Migrations})
	}

	er.Register(&domainEngine.EngineHook{
		Engine: sqlStore,
		PreStart: func(e domainEngine.Engine) error {
			store, ok := cast.As[*engineStoreImpl.SQL](e)
			if !ok {
				return interError.Retrieve(domainEngineStore.ErrKeyInvalid).
					Wrap(fmt.Errorf("db hook: unexpected engine %T", e))
			}
			assets, found := engineStoreImpl.AssetsFrom(store.HookContext().EngineContext.AppCtx)
			if !found {
				slog.Default().Warn("db migrations skipped", "reason", "no embedded tree")
				return nil
			}
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
	hookLog := func(engine domainEngine.Engine, items []domainConfig.LogItem) domainEngine.Engine {
		return &domainEngine.EngineHook{
			Engine: engine,
			PreStart: func(_ domainEngine.Engine) error {
				return engineLogImpl.PrepareOutputs(items)
			},
		}
	}
	engines := make([]domainEngine.Engine, 0, 3)

	if items := engineLogImpl.ItemsFor(cfg, "direct"); len(items) > 0 {
		options := make([]engineLogImpl.DirectOption, 0, len(items))
		for _, item := range items {
			options = append(options, engineLogImpl.WithDirectItem(item))
		}
		engines = append(engines, hookLog(engineLogImpl.NewDirectLog(options...), items))
	}

	if items := engineLogImpl.ItemsFor(cfg, "fanout"); len(items) > 0 {
		options := make([]engineLogImpl.FanoutOption, 0, len(items))
		for _, item := range items {
			options = append(options, engineLogImpl.WithFanoutItem(item))
		}
		engines = append(engines, hookLog(engineLogImpl.NewFanoutLog(options...), items))
	}

	if items := engineLogImpl.ItemsFor(cfg, "pattern"); len(items) > 0 {
		options := make([]engineLogImpl.PatternOption, 0, len(items))
		for _, item := range items {
			options = append(options, engineLogImpl.WithPatternItem(item))
		}
		engines = append(engines, hookLog(engineLogImpl.NewPatternLog(options...), items))
	}

	er.Register(engines...)
}

func SetupSetting(cfg domainContext.ConfigContext, er *domainRuntime.EngineRuntime) {
	layers := engineSettingImpl.LayersFrom(cfg)
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
		eec.Ebus.On(domainBus.EventEngineError, func(payload any) {
			type engineEvent struct {
				Group string `json:"group"`
				Name  string `json:"name"`
				State string `json:"state"`
			}
			source, ok := cast.To[map[string]any](payload)
			if !ok {
				return
			}
			event := engineEvent{}
			cast.ToStruct(source, &event, "json")
			engine, ok := eec.Engines[event.Group+"."+event.Name]
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
		sec.Ebus.On(domainBus.EventSessionExpired, func(payload any) {
			if sec.AppEbus == nil {
				return
			}
			sec.AppEbus.Emit(domainBus.EventSessionExpired, payload)
		})
	})
}
