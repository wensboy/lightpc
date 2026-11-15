package bootstrap

import (
	"io/fs"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wensboy/lightpc/backend/api/ui"
	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
	domainRuntime "github.com/wensboy/lightpc/backend/domain/runtime"
	storeimpl "github.com/wensboy/lightpc/backend/engine/store"
	"github.com/wensboy/lightpc/backend/internal/cast"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// ErrKeyInvalidMode reports a boot mode that the bootstrap does not dispatch.
const ErrKeyInvalidMode = "error::bootstrap::invalid_mode"

// BootstrapErrors lists every error published when the bootstrap dispatches a mode.
var BootstrapErrors = []interError.Option{
	{Key: ErrKeyInvalidMode, Code: interError.ErrCodeInvalidParam, Msg: "bootstrap: invalid boot mode"},
}

func Boot(bootCtx domainContext.BootContext) error {
	interError.RegisterAll(&bootCtx.AppContext, BootstrapErrors...)
	switch bootCtx.Mode {
	case domainContext.ModeTui:
		return RunTui(bootCtx.AppContext)
	case domainContext.ModeWeb:
		return RunWeb(bootCtx.AppContext)
	case domainContext.ModeGui:
		return RunGui(bootCtx.AppContext)
	}
	return interError.Retrieve(ErrKeyInvalidMode)
}

func RunTui(appCtx domainContext.AppContext) error {
	return nil
}

func RunWeb(appCtx domainContext.AppContext) error {
	return nil
}

func RunGui(appCtx domainContext.AppContext) error {
	cfg := appCtx.ConfigContext

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
	engineRuntime.Register(
		storeimpl.NewFS(),
	)
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
