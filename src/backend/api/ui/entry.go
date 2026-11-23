package ui

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	domainRuntime "github.com/wensboy/lightpc/backend/domain/runtime"
	serviceGreetImpl "github.com/wensboy/lightpc/backend/service/greet"
	serviceHelpImpl "github.com/wensboy/lightpc/backend/service/help"
	serviceSessionImpl "github.com/wensboy/lightpc/backend/service/session"
	serviceSettingImpl "github.com/wensboy/lightpc/backend/service/setting"
)

// RegisterService builds the services the UI talks to and exposes them to the front end.
func RegisterService(sr *domainRuntime.ServiceRuntime) application.Options {
	greetService := serviceGreetImpl.NewGreetService()
	settingService := serviceSettingImpl.NewSettingService()
	helpService := serviceHelpImpl.NewHelpService()
	sessionService := serviceSessionImpl.NewSessionService()
	sr.Register(greetService, settingService, helpService, sessionService)
	return application.Options{
		Services: []application.Service{
			application.NewService(greetService),
			application.NewService(settingService),
			application.NewService(helpService),
			application.NewService(sessionService),
		},
	}
}
