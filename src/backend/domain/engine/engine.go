package engine

import (
	domainBus "github.com/wensboy/lightpc/backend/domain/bus"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
)

const (
	EnginePending EngineState = "pending" // 等待
	EngineReady   EngineState = "ready"   // 就绪
	EngineRunning EngineState = "running" // 运行
	EngineStopped EngineState = "stopped" // 停止
	EngineError   EngineState = "error"   // 错误
)

type EngineState string

// EngineContext 作为引擎初始化传递上下文传递必要的信息.
type EngineContext struct {
	AppCtx  *domainContext.AppContext
	Ebus    domainBus.Bus
	Engines map[string]Engine // 依赖过滤暴露引擎依赖的引擎.
}

// EngineEventContext 用于外部定义引擎事件处理能力.
type EngineEventContext struct {
	Registry map[string]Engine
	Ebus     domainBus.Bus
}

// Engine 定义引擎实现的接口, 暴露底层能力, 为 service 编排逻辑提供接口.
type Engine interface {
	Group() string // 引擎组
	Name() string  // 实例名称
	Deps() []string

	State() EngineState // state 状态机, 用于判断并处理 engine

	Init(EngineContext) error
	Start() error
	Restart() error
	Stop() error
}
