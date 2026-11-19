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
	Services map[string]Service
	Ebus     domainBus.Bus
	// Engines mirrors ServiceContext so event handlers can resolve an engine without keeping one alive.
	Engines map[string]domainEngine.Engine
}

type Service interface {
	Group() string

	Init(ServiceContext) error
	Deinit() error
}
