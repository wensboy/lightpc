# 04 - engine & runtime

> 核心目的: 引擎和运行时抽象设计.
>
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

| 类型 | 说明 |
|---|---|
| `EngineState` | `pending` / `ready` / `running` / `stopped` / `error` |
| `Engine` | `Group` `Name` `Deps` `State` + `Init` `Start` `Restart` `Stop` |
| `EngineContext` | `AppCtx *AppContext` + `Ebus Bus`, Init 时由引擎保存 |
| `EngineEventContext` | `Registry map[string]Engine` + `Ebus`, 交给外部注册监听 |
| `EngineRuntime` | 注册 / 查询 / 拓扑启动 / 逆序退出 |
| `ServiceRuntime` | 以 group 为 key 注册 service, 初始化时注入 engines |
| `AppRuntime` | engine -> service 顺序启动, 逆序退出 |

```go
type Engine interface {
	Group() string
	Name() string
	Deps() []string
	State() EngineState
	Init(EngineContext) error
	Start() error
	Restart() error
	Stop() error
}
```

- 注册 key 为 `group.name`; 现有实现: `store.fs` / `store.db` / `log.direct` / `log.fanout` / `log.pattern`.

## 逻辑

```
EngineRuntime.Run
  ├─ 注册本运行时的错误 (依赖 AppCtx.ErrorContext 解析 kind)
  ├─ registry 快照 + key 排序
  ├─ Kahn 迭代求拓扑序          // 某一轮无法推进 => 依赖环
  ├─ execStack = ordered
  ├─ eFn(EngineEventContext)    // 外部注册监听
  └─ 逐个 Init -> start -> ready, 失败即 emit error 并返回

EngineRuntime.Exit: 取走并清空 execStack, 逆序 Stop, 逐个 emit stop/error
```

| 阶段 | FS | SQL |
|---|---|---|
| `New` | 全部字段零值, 不写默认值 | 同 |
| `Init` | 保存 ectx -> 配置补 root/mode -> `MkdirAll` -> `Validate` -> 注册 stop 监听 -> `Ready` | 保存 ectx -> `Validate` 连接池 -> 注册 stop 监听 -> `Ready` |
| `Start` | 只切状态到 `Running` | 同 |
| `Stop` | `Close`, 状态 `Stopped` | 同 |
| `Restart` | 复用保存的 ectx 重新 `Init` | 同 |

- FS 的 root 来自 `default.runtime.store.root`(`~` 经 `os.UserHomeDir` 展开), 权限位来自 `fs.fileMode` / `fs.dirMode`.
- `ServiceRuntime.Run` 先把 `EngineRuntime.registry` 快照为 `ServiceContext.Engines`, 再逐个 `Init`.
- `AppRuntime.Exit` 顺序相反: 先 `ServiceRuntime.Exit` 再 `EngineRuntime.Exit`.

## 注意点

- 引擎必须把 `EngineContext` 保存到结构字段, 否则 `Restart` 无法复用 `AppCtx` 与 `Ebus`.
- 配置驱动的默认值只在字段为零值时生效; `WithRoot` / `WithFileMode` / `WithDirMode` 赋非零值后不会被覆盖.
- `dirMode` 为 0 时跳过 `MkdirAll`, 避免创建出权限为 0 的 root.
- 同步即时处理型引擎(如 store) `Start` 只切状态, 无需拉起协程; 需要常驻协程的引擎才在 `Start` 中启动.

## 细节

- 引擎注册 key 为 `group.name`, 对应 backend-rules SHOULD 5 的"value 有效用 `.` 分隔"; 查询时 `group` 为空返回空列表, `name` 为空返回组内全部引擎.
- `Run` 的拓扑排序为确定性的 Kahn 迭代(key 排序后逐轮推进), 某一轮无法推进即判定为依赖环; 未注册的依赖不再视为已满足, 直接返回 `error::engine.runtime::missing_dependency`.
- `Register` 跳过 `nil` 与 `group`/`name` 为空的引擎, 避免注册出非法 ekey.
- 每个引擎 `Init` 前按自身 `Deps()` 经 `Engine(group, name)` 过滤出依赖实例写入 `EngineContext.Engines`(key 为实例自身的 `group.name`); 无依赖时为空 map, 不会与其它引擎共享.
- 依赖项可以只写组名(`core`), 此时 `Engine(group, "")` 命中该组全部引擎, 必须全部注入而非只取首个; 依赖匹配为空才判定为缺失.
- `Run` 与 `Exit` emit 的 payload 统一为 `map[string]any{group,name,state}`; `Run` 中在同函数内定义同 tag 的 struct, on 侧用 `cast.ToStruct` 解析.
- `Exit` 先取走并清空 `execStack` 再逆序 `Stop`, 避免重复退出.
- `EngineRuntime.Run` 在启动前调用 `WithEventFn` 注入的回调, 把 registry 快照与总线交给外部注册监听; 未注入时跳过.
- `ServiceContext.Engines` 取自 `EngineRuntime.registry` 的快照(同包加读锁复制), `ServiceContext.AppCtx` 取自 `AppRuntime.appCtx`.
- `AppRuntime` 先启动 engine 再启动 service, 退出时逆序: 先 `ServiceRuntime.Exit` 再 `EngineRuntime.Exit`.
- `AppRuntime.Run` 丢弃 engine 启动错误(`_ = ar.er.Run()`), 单个引擎失败不会阻断窗口拉起; 需要可观测性时应在该处记录或上报.
- store 引擎在 `Init` 中保存 `EngineContext` 并注册 `engine::stop` 监听, 只有 payload 的 `group.name` 与自身匹配才执行 `Close`.
- store 的默认值统一由 `context.AssignIfZero` 在 `Init` 内联应用, `New` 阶段不写死任何默认值.
