package bootstrap

import (
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	domainRuntime "github.com/wensboy/lightpc/backend/domain/runtime"
	logimpl "github.com/wensboy/lightpc/backend/engine/log"
	settingimpl "github.com/wensboy/lightpc/backend/engine/setting"
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
	app := application.New(opt)
	windowOptions := application.WebviewWindowOptions{Title: opt.Name}
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

func SetupLog(cfg domainContext.ConfigContext, er *domainRuntime.EngineRuntime) {
	engines := make([]domainEngine.Engine, 0, 3)

	if items := logimpl.ItemsFor(cfg, "direct"); len(items) > 0 {
		options := make([]logimpl.DirectOption, 0, len(items))
		for _, item := range items {
			options = append(options, logimpl.WithDirectItem(item))
		}
		engines = append(engines, logimpl.NewDirectLog(options...))
	}

	if items := logimpl.ItemsFor(cfg, "fanout"); len(items) > 0 {
		options := make([]logimpl.FanoutOption, 0, len(items))
		for _, item := range items {
			options = append(options, logimpl.WithFanoutItem(item))
		}
		engines = append(engines, logimpl.NewFanoutLog(options...))
	}

	if items := logimpl.ItemsFor(cfg, "pattern"); len(items) > 0 {
		options := make([]logimpl.PatternOption, 0, len(items))
		for _, item := range items {
			options = append(options, logimpl.WithPatternItem(item))
		}
		engines = append(engines, logimpl.NewPatternLog(options...))
	}

	er.Register(engines...)
}

func SetupSetting(cfg domainContext.ConfigContext, er *domainRuntime.EngineRuntime) {
	layers := settingimpl.LayersFrom(cfg)
	options := make([]settingimpl.Option, 0, len(layers))
	for _, layer := range layers {
		options = append(options, settingimpl.WithLayer(layer))
	}
	er.Register(settingimpl.NewSetting(options...))
}
