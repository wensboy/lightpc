# 05 - bus

> 核心目的: 总线抽象.
>
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

```go
type EventHandler func(any)

type Bus interface {
	On(name string, handler EventHandler)
	Emit(name string, payload any)
}
```

| 事件常量 | 值 | 触发时机 |
|---|---|---|
| `EventEnginePend` | `engine::pend` | 引擎进入等待 |
| `EventEngineInit` | `engine::init` | `Init` 完成 |
| `EventEngineReady` | `engine::ready` | `Start` 完成 |
| `EventEngineStart` | `engine::start` | 启动 |
| `EventEngineError` | `engine::error` | 任一阶段失败 |
| `EventEngineStop` | `engine::stop` | 停止 |

- 接口在 `domain/bus`, 唯一实现在 `internal/ebus`.

## 逻辑

```
Emit(name, payload)
  ├─ RLock
  ├─ 复制该事件的 handler 切片
  ├─ RUnlock
  └─ 逐个调用 handler(payload)
```

- payload 统一为 `map[string]any`, 引擎生命周期事件字段为 `{group, name, state}`.
- on 侧在回调内定义 tag 对齐的 struct, 用 `cast.To` + `cast.ToStruct` 还原.
- `EngineRuntime.Run` 启动前执行 `eFn(EngineEventContext)`, 把 registry 快照与总线交给外部注册监听.
- `storeimpl` 在 `Init` 中监听 `engine::stop`, 命中自身 `group.name` 才执行 `Close`.

## 注意点

1. 总线通常作为其他包的依赖, 通常不依赖其他包. 所有内容内聚在: `domain/bus`, `internal/ebus`.
2. 所有事件 key 定义到 `domain/bus/event.go` 中, 格式为: `EventXXX`, 通常为 const.
3. 所有 event key 使用 `::` 分隔语义与词.
4. 事件的 payload 在 emit 传递时通过 `map[string]any` 传递, 在 on 监听时通过 mapstructure 解码为实际的结构 T, 该结构直接定义在监听的 on 回调函数中.
5. 总线按实例隔离: `AppRuntime` / `EngineRuntime` / `ServiceRuntime` 各持一条总线, 子运行时通过 `WithAppBus` 接收上层总线.

## 细节

- `Emit` 必须复制 handler 列表再释放读锁调用, 否则 handler 内部调用 `On` 会因 `RWMutex` 不可重入而死锁.
- 同一事件可注册多个 handler, 按注册顺序同步调用; 任一 handler 阻塞会阻塞后续 handler 与 `Emit` 调用方.
- handler 收到的是 `any`, 类型断言失败应直接返回, 不要 panic, 避免影响其它监听者.
- 事件 key 在 `engine/` 实现中以常量引用, 不在业务代码里写字面量.
