package ui

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	domainRuntime "github.com/wensboy/lightpc/backend/domain/runtime"
	serviceGreet "github.com/wensboy/lightpc/backend/service/greet"
	serviceHelp "github.com/wensboy/lightpc/backend/service/help"
	serviceSetting "github.com/wensboy/lightpc/backend/service/setting"
)

// RegisterService builds the services the UI talks to and exposes them to the front end.
func RegisterService(sr *domainRuntime.ServiceRuntime) application.Options {
	greetService := serviceGreet.NewGreetService()
	settingService := serviceSetting.NewSettingService()
	helpService := serviceHelp.NewService()
	sr.Register(greetService, settingService, helpService)
	return application.Options{
		Services: []application.Service{
			application.NewService(greetService),
			application.NewService(settingService),
			application.NewService(helpService),
		},
	}
}
