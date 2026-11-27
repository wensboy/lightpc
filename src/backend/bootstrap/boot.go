package bootstrap

import (
	"io/fs"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wensboy/lightpc/backend/api/ui"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainRuntime "github.com/wensboy/lightpc/backend/domain/runtime"
	interError "github.com/wensboy/lightpc/backend/internal/error"
)

// ErrKeyInvalidMode reports a boot mode that the bootstrap does not dispatch.
const ErrKeyInvalidMode = "error::bootstrap::invalid_mode"

// BootstrapErrors lists every error published when the bootstrap dispatches a mode.
var BootstrapErrors = []interError.Option{
	{Key: ErrKeyInvalidMode, Code: interError.ErrCodeInvalidParam, Msg: "bootstrap: invalid boot mode"},
}

func Boot(bootCtx domainContext.BootContext) error {
	interError.RegisterAll(bootCtx.AppContext.ErrorContext, BootstrapErrors...)
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
	SetupStore(cfg, engineRuntime)
	SetupLog(cfg, engineRuntime)
	SetupSetting(cfg, engineRuntime)
	SetupNotice(cfg, engineRuntime)

	serviceRuntime := domainRuntime.NewServiceRuntime().WithAppRuntime(appRuntime)

	opt := ui.RegisterService(serviceRuntime)
	if dist, ok := appCtx.EmbedContext["/frontend/dist"]; ok && dist != nil {
		if assets, err := fs.Sub(dist, "frontend/dist"); err == nil {
			opt.Assets = application.AssetOptions{
				Handler:    application.AssetFileServerFS(assets),
				Middleware: ui.SpaFallback(assets),
			}
		}
	}
	app := SetupApp(cfg, opt)

	SetupEventHandle(appRuntime, engineRuntime, serviceRuntime)

	appRuntime.WithApp(app)
	appRuntime.WithEngineRuntime(engineRuntime)
	appRuntime.WithServiceRuntime(serviceRuntime)

	appRuntime.Run()
	return app.Run()
}
