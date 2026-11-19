package ui

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	domainRuntime "github.com/wensboy/lightpc/backend/domain/runtime"
	serviceGreet "github.com/wensboy/lightpc/backend/service/greet"
	serviceSetting "github.com/wensboy/lightpc/backend/service/setting"
)

// RegisterService builds the services the UI talks to and exposes them to the front end.
func RegisterService(sr *domainRuntime.ServiceRuntime) application.Options {
	greetService := serviceGreet.NewGreetService()
	settingService := serviceSetting.NewSettingService()
	sr.Register(greetService, settingService)
	return application.Options{
		Services: []application.Service{
			application.NewService(greetService),
			application.NewService(settingService),
		},
	}
}
