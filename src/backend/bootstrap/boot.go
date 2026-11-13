package bootstrap

import (
	"errors"
	"io/fs"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wensboy/lightpc/backend/api/ui"
	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	domainRuntime "github.com/wensboy/lightpc/backend/domain/runtime"
	"github.com/wensboy/lightpc/backend/internal/cast"
)

func Boot(bootCtx domainContext.BootContext) error {
	switch bootCtx.Mode {
	case domainContext.ModeTui:
		return RunTui(bootCtx.AppContext)
	case domainContext.ModeWeb:
		return RunWeb(bootCtx.AppContext)
	case domainContext.ModeGui:
		return RunGui(bootCtx.AppContext)
	}
	return errors.New("无效启动模式, 执行终止")
}

func RunTui(appCtx domainContext.AppContext) error {
	return nil
}

func RunWeb(appCtx domainContext.AppContext) error {
	return nil
}

func RunGui(appCtx domainContext.AppContext) error {
	cfg := appCtx.ConfigContext
	appCtx.ErrorContext = SetupErrors(cfg)

	appRuntime := domainRuntime.NewAppRuntime().WithAppContext(appCtx)
	engineRuntime := domainRuntime.NewEngineRuntime().WithAppRuntime(appRuntime)
	engineRuntime.WithEventFn(func(eec domainEngine.EngineEventContext) {
		type engineEvent struct {
			Group string `json:"group"`
			Name  string `json:"name"`
			State string `json:"state"`
		}
		eec.Ebus.On(domainBus.EventEngineError, func(payload any) {
			source, ok := cast.To[map[string]any](payload)
			if !ok {
				return
			}
			event := engineEvent{}
			cast.ToStruct(source, &event, "json")
			engine, ok := eec.Registry[event.Group+"."+event.Name]
			if !ok {
				return
			}
			_ = engine.Restart()
		})
	})
	serviceRuntime := domainRuntime.NewServiceRuntime().WithAppRuntime(appRuntime)

	opt := ui.RegisterService(serviceRuntime)
	if dist, ok := appCtx.EmbedContext["/frontend/dist"]; ok && dist != nil {
		if assets, err := fs.Sub(dist, "frontend/dist"); err == nil {
			opt.Assets = application.AssetOptions{Handler: application.AssetFileServerFS(assets)}
		}
	}
	app := SetupApp(cfg, opt)

	appRuntime.WithApp(app)
	appRuntime.WithEngineRuntime(engineRuntime)
	appRuntime.WithServiceRuntime(serviceRuntime)

	appRuntime.Run()
	return app.Run()
}
