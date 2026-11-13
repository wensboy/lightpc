package ui

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	domainRuntime "github.com/wensboy/lightpc/backend/domain/runtime"
	serviceGreet "github.com/wensboy/lightpc/backend/service/greet"
)

func RegisterService(sr *domainRuntime.ServiceRuntime) application.Options {
	greetService := serviceGreet.NewGreetService()
	sr.Register(greetService)
	return application.Options{
		Services: []application.Service{
			application.NewService(greetService),
		},
	}
}
