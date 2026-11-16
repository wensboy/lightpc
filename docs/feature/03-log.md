# 03 - log

> 核心目的: 日志抽象.
>
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

| 类型 | 位置 | 说明 |
|---|---|---|
| `FileHandler` | `internal/log` | 实现 `slog.Handler`, 按级别过滤后写入文件; 打开失败则该 handler 静默 |
| `FileOption` | `internal/log` | `SerializerOption` + `Serializer` + `Path` + `Mode` |
| `ConsoleHandler` | `internal/log` | 与文件版同构, 输出到 `stdout` / `stderr`, 不拥有流 |
| `ConsoleOption` | `internal/log` | `SerializerOption` + `Serializer` + `Out` |
| `DirectLog` | `engine/log` | 1 对 1 分发, 一条 record 只由一个 handler 消费 |
| `FanoutLog` | `engine/log` | 1 对多广播 |
| `PatternLog` | `engine/log` | 按模式路由 |

三个引擎都按客户端 / 服务端分开暴露, 方法签名一致, 匹配语义不同:

| 引擎 | 客户端 | 服务端 | 匹配 | 记录去向 |
|---|---|---|---|---|
| direct | `DirectClient.Logger(name) (*slog.Logger, bool)` | `DirectServer.Register(name, handler)` | 精确 | 该 name 的 1 个 handler |
| fanout | `FanoutClient.Logger(name) (*slog.Logger, bool)` | `FanoutServer.Register(name, handler)` | 精确 | 该 name 的全部 handler |
| pattern | `PatternClient.Logger(name) (*slog.Logger, bool)` | `PatternServer.Register(pattern, handler)` | glob(`path.Match`) | 命中全部 pattern 的 handler |

```go
handler := log.NewFileHandler(log.FileOption{Path: "app.log", Mode: 0o644}).
	WithLevel(slog.LevelInfo)
logger := slog.New(handler)
```

- `FileHandler` 可链式: `WithSerializer` / `WithOut` / `WithLevel` 均返回自身.
- engine 侧常量: `GroupLog = "log"`, `NameDirect` / `NameFanout` / `NamePattern`, 三者均实现 `domainEngine.Engine`.

## 逻辑

```
logger.Info
  └─ FileHandler.Enabled   // 只比较 fh.level
        └─ FileHandler.Handle
              ├─ mu.Lock
              ├─ serializer.Handle(record)   // 内层 level 固定 Debug, 不再过滤
              └─ mu.Unlock
```

- `WithAttrs` / `WithGroup`: 派生新的 `FileHandler`, 并通过 serializer 委托保留 attrs 与 group 语义.
- `NewFileHandler`: 打开 `Path` 失败回退标准输出; `Mode` 为 0 时用 `0644`.
- 三个引擎生命周期一致: `Init` 保存 `EngineContext` 并置 `Ready`, `Start` 只切状态到 `Running`, `Stop` 只置 `Stopped` 与 `closed`, 保留 handler 表以便重启后无需重新绑定.
- direct / fanout 的 `Logger(name)` 精确匹配; pattern 的 `Logger(name)` 遍历已注册 glob, 把 `path.Match` 命中的 handler 全部收集.
- 三者统一用 `slog.NewMultiHandler` 组合 handler, 即使只有一个命中, 客户端拿到的仍是 MultiHandler, 行为一致.

## 注意点

- 级别过滤职责单一: 外层 `Enabled` 决定写不写, 内层 serializer 不参与判断.
- 日志引擎按"作用范围"区分: direct 一对一、fanout 广播、pattern 按规则路由, 三者共享 `log` 组, 注册 key 为 `log.direct` / `log.fanout` / `log.pattern`.
- 写入必须串行化, `Handle` 全程持锁, 避免并发写出交错.

## 细节

- `slog.Handler` 的输出 writer 在构造时绑定, 无法在 `Handle` 中重定向; 因此 `FileHandler` 让 serializer 直接写 `out`, `WithAttrs`/`WithGroup` 通过委托派生新的 serializer 保持 attrs 与 group 语义.
- `NewFileHandler` 不再降级: 打开 `FileOption.Path` 失败时 out 与 serializer 均保持为空, handler 静默; `Mode` 为 0 时使用 `0644`; 并发写入由内部互斥锁串行化.
- `ConsoleHandler` 与 `FileHandler` 同构, 输出目标由 `ConsoleOption.Out` 决定(默认 `os.Stdout`, 传 `os.Stderr` 即切错误流); 它不拥有流, `Close` 只停止后续写入.
- 级别过滤只由外层 `FileHandler.Enabled` 负责, 内层序列化器创建时固定 `slog.LevelDebug`, 保证 `Handle` 不会被内层再次拦截.
- `DirectLog` 的 handler 表由读写锁保护, `Logger` 与 `Register` 均可并发调用; `Stop` 不清空 handler 表, `Restart` 后仍可直接取回.
- 客户端与服务端接口分开暴露, 其它 engine 通过 `cast.As[DirectClient]` / `cast.As[DirectServer]` 把 `Engine` 实例转成对应能力使用.
- `Logger` 按名称精确匹配, 不做前缀或模糊匹配; 未注册返回 `(nil, false)`, 调用方必须先判 `ok`.
- 三个引擎的构造 Option 分别命名为 `WithDirectHandler` / `WithFanoutHandler` / `WithPatternHandler`, 均与 `Register` 一样对空名称与 `nil` handler 静默忽略, 与 store 引擎注册时跳过空标识的约定一致.
- 服务端语义差异: `DirectServer.Register` 为**覆盖**(同名只保留最后一次), `FanoutServer` 与 `PatternServer` 的 `Register` 为**追加**(同名多次注册叠加).
- `FanoutLog.Logger` 在锁内 `slices.Clone` 后再组合, 避免后续 `append` 修改底层数组影响已经返回的 `MultiHandler`.
- `PatternLog.Logger` 在锁内完成匹配与收集, `path.Match` 对非法 glob 返回错误, 按不匹配处理而非 panic.
- `NewFileHandler` 在构造时即打开文件, 由 `closer` 字段记录资源所有权; `Close` 幂等, 输出目标不归 handler 所有(标准输出或经 `WithOut` 替换)时不关闭任何资源.
- `WithAttrs` / `WithGroup` 派生的 handler 不继承 `closer`, 其 `Close` 为空操作, 资源生命周期始终由原始 handler 掌握.
- `Close` 之后 `Handle` 静默丢弃记录并返回 `nil`, 日志写入不会因关闭失败而上抛.
- 过滤集中在 `Enabled`: 级别下限、输出可用性(out 与 serializer 非空)、`closed` 三者一次判定, `Handle` 不再承担过滤职责.
- `closed` 使用 `atomic.Bool`, `Enabled` 可无锁读取; `mu` 只用于串行化 `Handle` 的写入.
- 装配逻辑位于 `engine/log` 的 `Setup`, 读取 `default.runtime.log.<group>[]`, 三组各构建一个实例: direct 每个 item 对应一个具名 logger, fanout 与 pattern 把同一 item 的全部 output 追加到同一名称下; `bootstrap.SetupLog` 只做转发.
- 分组缺失或列表为空时不创建实例, 因此返回值长度在 0~3 之间.
- `output` 在 direct 组是对象, 在 fanout / pattern 组是数组; `logimpl.setupItem` 统一转成列表再逐个构建 handler, 两种形态共用一条解析路径.
- 结构定义在 `domain/config/map.go`(`LogItem` / `LogOutput` / `LogSerializer`), 装配与解析在 `engine/log/setup.go`, 引导层不再持有配置结构与处理函数.
- `internal/log` 提供映射: `ParseLevel`(文本 -> `slog.Level`, 未知回退 info)、`ParseMode`(如 `0o644` -> `FileMode`)、`NewSerializer`(text / json, 未知回退 text).
- 文件输出的 `path` 支持 `~` 前缀, 父目录在构建时自动创建; `handler` 为 `file` 走 `FileHandler`, 其余走 `ConsoleHandler`(`out` 可取 `stdout` / `stderr`).
- `slog` 全局默认 logger 固定取自 fanout 引擎的空名句柄: `fanout.Logger("")` 聚合全部已注册 handler, 一次普通 `slog` 调用即同时落到 stdout 与配置的 file.
- fanout 的 `Logger(name)` 有双语义: name 非空只广播该名称下的 handler, name 为空广播全部 handler; direct / pattern 仍按各自规则匹配.
