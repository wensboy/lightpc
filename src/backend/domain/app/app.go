package app

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
)

type AppEventContext struct {
	Ebus domainBus.Bus
	App  *application.App
}
