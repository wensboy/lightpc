package runtime

import (
	"github.com/wailsapp/wails/v3/pkg/application"
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

func (ar *AppRuntime) Run() {
	if ar.er != nil {
		ar.er.WithAppBus(ar.ebus)
		_ = ar.er.Run()
	}
	if ar.sr != nil {
		ar.sr.WithAppBus(ar.ebus)
		ar.sr.Run()
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
