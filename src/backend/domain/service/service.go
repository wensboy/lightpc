package service

import (
	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	domainEngine "github.com/wensboy/lightpc/backend/domain/engine"
)

type ServiceContext struct {
	AppCtx *domainContext.AppContext
	Ebus   domainBus.Bus
	// 暴露所有有效的 engines
	Engines map[string]domainEngine.Engine
}

type ServiceEventContext struct {
	Registry map[string]Service
	Ebus     domainBus.Bus
}

type Service interface {
	Group() string

	Init(ServiceContext) error
	Deinit() error
}
