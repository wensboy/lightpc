# 04 - engine & runtime

> 核心目的: 引擎和运行时抽象设计.
> 
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

## 逻辑

## 注意点

## 细节

- 引擎注册 key 为 `group.name`, 对应 backend-rules SHOULD 5 的"value 有效用 `.` 分隔"; 查询时 `group` 为空返回空列表, `name` 为空返回组内全部引擎.
- `Run` 的拓扑排序为确定性的 Kahn 迭代(key 排序后逐轮推进), 未注册的依赖视为已满足, 某一轮无法推进即判定为依赖环.
- `Run` 与 `Exit` emit 的 payload 统一为 `map[string]any{group,name,state}`; `Run` 中在同函数内定义同 tag 的 struct, on 侧用 `cast.ToStruct` 解析.
- `Exit` 先取走并清空 `execStack` 再逆序 `Stop`, 避免重复退出.
- `EngineRuntime.Run` 在启动前调用 `WithEventFn` 注入的回调, 把 registry 快照与总线交给外部注册监听; 未注入时跳过.
- `ServiceContext.Engines` 取自 `EngineRuntime.registry` 的快照(同包加读锁复制), `ServiceContext.AppCtx` 取自 `AppRuntime.appCtx`.
- `AppRuntime` 先启动 engine 再启动 service, 退出时逆序: 先 `ServiceRuntime.Exit` 再 `EngineRuntime.Exit`.