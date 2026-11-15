# 01 - bootstrap

> 核心目的: 程序启动引导结构.
>
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

- `backend/domain/context`: `xxxContext` 总是传输型上下文用于传输数据, 可以被持有并持续调用获取上下文信息. 上下文可组合或嵌入.
- `backend/domain/runtime`: `xxxRuntime` 总是标记一个有限状态机的所有状态信息, 用于状态驱动变化. 通常为 struct 结构.
- `backend/domain/bus`: `xxxBus` 总线结构, 按照语义决定实现和功能, 总线是实例级别隔离的.

| 入口 | 签名 | 职责 |
|---|---|---|
| `api/cmd.Execute` | `Execute(EmbedContext) int` | 组装上下文, 交 cli 解析, 返回进程状态码 |
| `bootstrap.SetupConfig` | `SetupConfig(EmbedContext) *config.Config` | 加载 default / errors / command / .env 四类来源 |
| `bootstrap.SetupApp` | `SetupApp(ConfigContext, application.Options) *application.App` | 按配置构建 app 与主窗口 |
| `bootstrap.SetupErrors` | `SetupErrors(ConfigContext) ErrorContext` | 由 `errors.json` 构建 `code -> kind` 并注册 root 组 |
| `bootstrap.Boot` | `Boot(BootContext) error` | 按 `Mode` 分发 TUI / Web / GUI |

## 逻辑

```
Execute
  ├─ SetupConfig ─ FileSource(default.json, errors.json)
  │                 EnvSource(.env)
  │                 FlagSource(command.json)
  ├─ appCtx{Context, EmbedContext, ConfigContext, ErrorContext}
  ├─ setupCommand(rootCmd, appCtx) ─ 闭包注入, 不依赖 ctx 断言
  └─ rootCmd.Run ─ Action ─ Boot(Mode=c.Int("mode"))
        ├─ ModeTui ─ RunTui   (空实现)
        ├─ ModeWeb ─ RunWeb   (空实现)
        └─ ModeGui ─ RunGui
              ├─ AppRuntime + EngineRuntime + ServiceRuntime
              ├─ engineRuntime.Register(storeimpl.NewFS())
              ├─ ui.RegisterService(serviceRuntime)
              ├─ SetupApp ─ application.New + Window.NewWithOptions
              ├─ appRuntime.Run()   // engine -> service
              └─ app.Run()          // wails 事件循环
```

- `RunGui` 把 `EmbedContext["/frontend/dist"]` 经 `fs.Sub` 挂到 `AssetFileServerFS`, 作为窗口静态资源.
- 未识别 `Mode` 时注册并返回 `error::bootstrap::invalid_mode`.

## 注意点

- 上下文在入口一次性构建, 下游只读; `ErrorContext` 必须在 `Boot` 之前就位, 否则 engine 注册错误时拿不到 kind.
- `AppRuntime.Run` 内部是 `_ = ar.er.Run()`, engine 启动失败不会阻断窗口拉起, 也不会进入日志.
- `mode` 在 `command.json` 中为 `required: true`, 拉起时必须显式传参 (`--mode 2`).

## 细节

- `SetupApp` 从 `default.application.{name,description}` 与 `default.application.window.{width,height}` 取值; wails v3 的 `application.Options` 没有窗口字段, 标准窗口通过 `app.Window.NewWithOptions` 创建.
- `cmd` 入口通过 `c.Int("mode")` 读取 `mode` flag 并转 `uint8` 写入 `BootContext.Mode`.
- `SetupErrors` 读取 `errors.<group>.definition[]` 中每项的 `code` 与 `kind` 构建 `map[code]kind`; 配置缺失、分组不是对象或缺字段时跳过该条目. `errors.json` 与 `default.json` 一并由 `SetupConfig` 加载.
- `root` 组按 `error::internal::<kind snake>` 额外注册到全局错误表, `not_exist` 即 `error.Retrieve` 的兜底 key.
- `RunGui` 当前只注册 `storeimpl.NewFS()` 一个 engine, 其 root 来自 `default.runtime.store.root`, 首次拉起会自动创建目录.
