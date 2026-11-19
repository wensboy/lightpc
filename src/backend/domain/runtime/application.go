package runtime

import (
	"log/slog"

	"github.com/wailsapp/wails/v3/pkg/application"
	domainApp "github.com/wensboy/lightpc/backend/domain/app"
	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	"github.com/wensboy/lightpc/backend/internal/ebus"
)

type Runtime interface {
	Run()  // 资源获取并管理
	Exit() // 资源释放
}

// AppRuntime 为一个标准应用运行时.
type AppRuntime struct {
	// context
	appCtx domainContext.AppContext

	// runtime
	er *EngineRuntime
	sr *ServiceRuntime

	// event bus in app runtime
	ebus domainBus.Bus
	eFn  func(domainApp.AppEventContext)
	// window instance
	app *application.App
}

func NewAppRuntime() *AppRuntime {
	return &AppRuntime{
		ebus: ebus.NewEventBus(),
	}
}

func (ar *AppRuntime) WithAppContext(ac domainContext.AppContext) *AppRuntime {
	ar.appCtx = ac
	return ar
}

func (ar *AppRuntime) WithEngineRuntime(er *EngineRuntime) *AppRuntime {
	ar.er = er
	return ar
}

func (ar *AppRuntime) WithServiceRuntime(sr *ServiceRuntime) *AppRuntime {
	ar.sr = sr
	return ar
}

func (ar *AppRuntime) WithApp(app *application.App) *AppRuntime {
	ar.app = app
	return ar
}

func (ar *AppRuntime) WithEventFn(efn func(domainApp.AppEventContext)) *AppRuntime {
	ar.eFn = efn
	return ar
}

// Run starts the engine and service runtimes and then invokes efn with the event context, so the caller
// can subscribe to runtime events knowing every layer below it is already alive.
func (ar *AppRuntime) Run() {
	if ar.er != nil {
		ar.er.WithAppBus(ar.ebus)
		if err := ar.er.Run(); err != nil {
			slog.Default().Warn("engine runtime failed", "err", err)
		}
	}
	if ar.sr != nil {
		ar.sr.WithAppBus(ar.ebus)
		ar.sr.Run()
	}
	if ar.eFn != nil {
		ar.eFn(domainApp.AppEventContext{Ebus: ar.ebus, App: ar.app})
	}
}

func (ar *AppRuntime) Exit() {
	if ar.sr != nil {
		ar.sr.Exit()
	}
	if ar.er != nil {
		_ = ar.er.Exit()
	}
}
