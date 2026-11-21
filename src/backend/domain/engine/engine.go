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
	Engines map[string]Engine
	Ebus    domainBus.Bus
	AppEbus domainBus.Bus // 向 app runtime 抛出事件
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

// EngineHook wraps an engine to expose explicit lifecycle hooks. A nil callback is a no-op, so a caller
// only provides the hooks it needs.
type EngineHook struct {
	Engine
	PreStart  func(Engine) error
	PostStart func(Engine) error
	PreStop   func(Engine) error
	PostStop  func(Engine) error
}

// Start implements Engine: the pre hook runs before the engine starts and the post hook after it
// succeeded, so a failing start skips the post hook.
func (eh *EngineHook) Start() error {
	if err := callHook(eh.PreStart, eh.Engine); err != nil {
		return err
	}
	if err := eh.Engine.Start(); err != nil {
		return err
	}
	if err := callHook(eh.PostStart, eh.Engine); err != nil {
		return err
	}
	return nil
}

// Stop implements Engine and mirrors Start: a failing stop skips the post hook.
func (eh *EngineHook) Stop() error {
	if err := callHook(eh.PreStop, eh.Engine); err != nil {
		return err
	}
	if err := eh.Engine.Stop(); err != nil {
		return err
	}
	if err := callHook(eh.PostStop, eh.Engine); err != nil {
		return err
	}
	return nil
}

// Unwrap exposes the engine a hook wraps, so a consumer that needs the concrete type reaches the engine
// instead of the wrapper.
func (eh *EngineHook) Unwrap() Engine {
	return eh.Engine
}

// Unwrap strips every hook wrapper from engine. A consumer that resolves the concrete implementation, such
// as a service casting to a file store, therefore works whether or not the engine is hooked.
func Unwrap(engine Engine) Engine {
	for {
		hook, ok := engine.(interface{ Unwrap() Engine })
		if !ok {
			return engine
		}
		inner := hook.Unwrap()
		if inner == nil {
			return engine
		}
		engine = inner
	}
}

// callHook runs hook with engine when the hook is set.
func callHook(hook func(Engine) error, engine Engine) error {
	if hook != nil {
		return hook(engine)
	}
	return nil
}

// EngineDep composes the dependency key of an engine with the same algorithm the registry uses.
// An empty group yields an empty key, an empty name yields the group, otherwise it is group.name.
func EngineDep(group, name string) string {
	switch {
	case group == "":
		return ""
	case name == "":
		return group
	default:
		return group + "." + name
	}
}
