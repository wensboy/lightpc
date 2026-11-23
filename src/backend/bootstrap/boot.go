package bootstrap

import (
	"io/fs"
	"net/http"

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
	SetupStore(cfg, engineRuntime)
	SetupLog(cfg, engineRuntime)
	SetupSetting(cfg, engineRuntime)

	serviceRuntime := domainRuntime.NewServiceRuntime().WithAppRuntime(appRuntime)

	opt := ui.RegisterService(serviceRuntime)
	if dist, ok := appCtx.EmbedContext["/frontend/dist"]; ok && dist != nil {
		if assets, err := fs.Sub(dist, "frontend/dist"); err == nil {
			opt.Assets = application.AssetOptions{Handler: spaHandler(assets)}
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

// spaHandler serves the embedded bundle and answers index.html for a path that is not a file.
//
// The asset server maps a request onto a file name, so a reload on a client route asks for a document that does
// not exist. Returning the shell instead lets the router resolve the path after the page has loaded, which is
// what makes a deep link or a reload survive.
func spaHandler(fsys fs.FS) http.Handler {
	files := application.AssetFileServerFS(fsys)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		files.ServeHTTP(recorder, r)
		if recorder.status != http.StatusNotFound {
			return
		}

		index, err := fs.ReadFile(fsys, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(index)
	})
}

// statusRecorder remembers the status the wrapped handler wrote, so the fallback can tell a missing file from a
// served one.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader records the status before passing it on.
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
