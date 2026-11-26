---
name: '引擎链路'
description: '新增引擎处理链路.'
---

> ai agent 能够将补充的示例追加到 Example 段下, 无明确编辑其他段时不允许编辑.

## Step

1. 在 `backend/domain/engine/<id>/shared.go` 下写出该组配置的**完整映射**: 文档里出现的每个字段都有对应结构。运行时形状会变的字段用 `json.RawMessage`。
2. 在 `backend/domain/engine/<id>/<id>.go` 下声明接口, 在 `shared.go` 下声明错误 key、`Errors` 与 dto。
3. 在 `backend/engine/<id>/` 下实现引擎: 按实例名分文件, 一个文件一个引擎, 辅助职责单独成文件。
4. 在 `backend/domain/engine/key.go` 注册组名与实例名常量, 并在 `var` 段拼出 `EngineDep` 快捷键。
5. 在 `backend/bootstrap/setup.go` 增加 `Setup<Name>`, 负责把文档解成配置结构、转成 Option、交给 `New<Name>`。由 `bootstrap/boot.go` 在依赖就绪后调用。
6. 若引擎需要把变化送达界面, 把事件名加入 `domain/bus/event.go` 的 `UiEvents` 白名单; 转发由 bootstrap 的桥接完成, 引擎本身不转发。

## Hint

- 一个中断点需要用户 rv 输出和纠正:
  + 引擎的职责边界。引擎只做「底层能力」, 面向界面的记录、展示与删除归属对应的 service。以 notice 为例: 引擎负责缓冲、发送与历史, 通知列表的展示与单条删除属于后续的 notice service。
- package 名称规范:
  + `backend/engine/<id>/` → `engine<Id>Impl`
  + `backend/domain/engine/<id>/` → `engine<Id>`
  + 引入时通过 alias 显式标识
- 引擎内错误定义到 `backend/domain/engine/<id>/shared.go` 中, key 的三段式为 `error::engine.<id>::<situation>`。
- 引擎内共享处理定义到 `backend/engine/<id>/shared.go` 中, 只放非导出常量与无状态工具函数。
- **引擎不读配置**。它不导入 `domain/context`, 不知道配置 key, 也不持有 `ConfigContext`。所有值经构造参数进入, 见「三段职责」。
- **引擎不设默认值**。缺少必需参数时不兜底, 而是让 `Init` 或构造失败, 使「初始状态不正确」在启动时暴露, 而不是在运行中以一个看似正常的默认值掩盖。
- 配置文件里的值是**兜底来源**, 不是引擎的默认值: `default.json` 承担原本会硬编码在代码里的常量, 引擎本身不认识它。
- 引擎以 Option 模式接收构造参数:
  + 结构格式 `type <X>Option func(*<X>)`, 由 `With<X><Field>` 生成。
  + 构造格式 `New<X>(opts ...<X>Option) *<X>`。
  + Option 只校验自己的参数是否可用, 不做补全: 一个无效值直接忽略, 让 `Init` 里的有效性判断决定后续。
- 引擎必须同时实现 `domainEngine.Engine` 与 `domain/engine/<id>/` 下的特化接口:
  + 前者让调度器管理生命周期。
  + 后者让其他引擎与 service 通过 `cast.As` 取得能力, 不必认识具体类型。
- 引擎之间只经接口互相依赖, 依赖键一律由 `EngineDep` 拼出, 不写字面量。
- 实现层文件按**实例名**命名, 不写 `logic.go`; 辅助职责单独成文件并按职责命名。

### 三段职责

配置从文档到引擎要经过三步, 每一步有固定归属。混在一起是这类代码最常见的失控方式。

| 段 | 职责 | 位置 | 输入 → 输出 |
|---|---|---|---|
| 结构映射 | 文档的每个字段都有对应结构 | `domain/engine/<id>/shared.go` | 文档形状 → `Config` |
| 结构解析 | 用 `Lookup` 取出子树并解码成 `Config` | `backend/bootstrap/setup.go` | `ConfigContext` → `Config` |
| 结构转换 | 把 `Config` 变成 Option 能接受的参数 | `domain/engine/<id>/shared.go` | `Config` → `[]Entry` / `[]Layer` |
| 构造 | 把参数装进引擎 | `backend/engine/<id>/` | Option → 引擎字段 |

转换若只对某个实现有意义(例如把字符串 mode 变成 `fs.FileMode`), 直接定义在 `backend/engine/<id>/` 下, 由 `With` 接收。

## Example

**当前 log engine 的结构完美符合规范, 后续所有引擎参照类似结构实现**

以下共性抽取自 `log`, `store`, `setting`, `notice` 四个已完成实现, 是对上述 Step 与 Hint 的落实方式。

### 文件职责

```
backend/domain/engine/<id>/
├── <id>.go      引擎接口: <Id>Client 特化契约(仅此处)
└── shared.go    错误 key + Errors + 配置结构(Config) + 转换方法 + 其余必要的处理结构

backend/engine/<id>/
├── <instance>.go    <Instance> 结构 + Option + 生命周期 + 全部能力方法
└── shared.go        非导出常量 + 无状态工具函数
```

| 引擎 | domain 文件 | impl 文件 |
|---|---|---|
| `log` | `direct.go` `fanout.go` `pattern.go` `shared.go` | `direct.go` `fanout.go` `pattern.go` `handler.go` `shared.go` |
| `store` | `db.go` `fs.go` `shared.go` | `db.go` `fs.go` `migrate.go` `signal.go` |
| `setting` | `setting.go` `shared.go` | `setting.go` `shared.go` |
| `notice` | `notice.go` `shared.go` | `notice.go` `shared.go` |

`log` 组的三个引擎各有 `direct.go` / `fanout.go` / `pattern.go`; 辅助职责单独成文件——`handler.go` 是处理器构建、`migrate.go` 是迁移执行、`signal.go` 是停止信号监听。

### 契约层: 完整配置映射

`shared.go` 必须让文档里的每个字段都有落点。判断标准很直接: 打开 `default.json` 对应那一段, 逐个字段能在结构里找到同名 `json` 标签。

```go
// Config maps runtime.engines.log completely: one list per sub-group, each entry describing where its records go.
type Config struct {
	Direct  []LogEntry `json:"direct"`
	Fanout  []LogEntry `json:"fanout"`
	Pattern []LogEntry `json:"pattern"`
}
```

`store` 的映射是嵌套的, 也在一个 `Config` 里展开:

```go
// Config maps runtime.engines.store completely: the root the file store is chrooted to, the modes it applies, and
// every connection the database store opens.
type Config struct {
	Root string   `json:"root"`
	Fs   FsConfig `json:"fs"`
	Db   DbConfig `json:"db"`
}

type FsConfig struct {
	FileMode string `json:"fileMode"`
	DirMode  string `json:"dirMode"`
}

type DbConfig struct {
	SQL []ConnConfig `json:"sql"`
}
```

### 运行时形状会变的字段用 RawMessage

同一个字段在不同条目里长得不一样时, 不拆成多个结构, 而是保留原始字节, 由使用点决定解成什么:

```go
// LogEntry maps one entry of a log sub-group.
type LogEntry struct {
	Name    string `json:"name"`
	Level   string `json:"level"`
	Handler string `json:"handler"`
	// The shape follows the handler: file carries a mode, console carries a stream name, database carries a pool key.
	// Decoding it here would mean a struct that has every field of every handler and is wrong for all of them.
	Output json.RawMessage `json:"output"`
}
```

具体形态在实现层展开, 用「先试数组再退回对象」覆盖 direct 与 fanout / pattern 的差异:

```go
// The declared shape follows the sub-group: direct writes "output": {...} while fanout and pattern write
// "output": [{...}]. Trying the array first and falling back to the object handles both without the caller having
// to say which group it read the entry from.
func outputsOf(item engineLog.LogEntry) []logOutput
```

转换后的形态是引擎自己的结构, 定义在 `engine/<id>/shared.go`:

```go
// logEntry is what the engine actually works with: the configured entry with its level parsed and its outputs
// unfolded into one list.
type logEntry struct {
	name    string
	level   slog.Level
	handler string
	outputs []logOutput
}
```

### 解析在 setup

一个泛型解码器覆盖所有组, 结构本身由契约包定义:

```go
// groupConfig decodes one engine group into the structure the contract declares.
//
// Reading the document is the bootstrap's job: an engine receives values, never a config context. A group that is
// missing or malformed yields the zero value, and each setup function decides what that means for it.
func groupConfig[T any](cfg domainContext.ConfigContext, key string) T {
	var config T
	if cfg == nil {
		return config
	}
	value, found := cfg.Lookup(key)
	if !found {
		return config
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return config
	}
	if err := json.Unmarshal(encoded, &config); err != nil {
		slog.Default().Warn("engine config rejected", "key", key, "err", err)
		return config
	}
	return config
}
```

每个 `Setup<Name>` 自己决定零值意味着什么:

```go
func SetupStore(cfg domainContext.ConfigContext, er *domainRuntime.EngineRuntime) {
	settings := groupConfig[domainEngineStore.Config](cfg, "default.runtime.engines.store")

	fsOptions := make([]engineStoreImpl.FSOption, 0, 3)
	if settings.Root != "" {
		if resource, err := config.NewResource(settings.Root); err == nil && !resource.IsEmbed() {
			if resolved, err := resource.Path(); err == nil {
				fsOptions = append(fsOptions, engineStoreImpl.WithRoot(resolved))
			}
		}
	}
	if mode := engineStoreImpl.ParseMode(settings.Fs.FileMode); mode != 0 {
		fsOptions = append(fsOptions, engineStoreImpl.WithFileMode(mode))
	}
	if mode := engineStoreImpl.ParseMode(settings.Fs.DirMode); mode != 0 {
		fsOptions = append(fsOptions, engineStoreImpl.WithDirMode(mode))
	}
	fsStore := engineStoreImpl.NewFS(fsOptions...)
}
```

三个细节值得注意:

- `Lookup` 的路径指向**叶子**时返回标量, 指向**子树**时返回 `map`。`groupConfig` 解的是子树, 所以用 json 往返; 若某个 key 直接指向数字, 必须按标量读, 且用 `cast.As` 而非 `cast.To`(前者做可转换翻译, 后者只是类型断言, `float64` 到 `int` 只有前者成立)。
- Option 的参数无效时**不追加**, 而不是追加一个会被忽略的值。这样 `New<X>` 收到的列表里没有噪声。
- 解析失败时只告警, 让 `Init` 在做有效性判断时暴露后果。

### 转换在契约包

`Config` 到 Option 参数的映射属于契约包, 因为它只依赖契约自己的类型:

```go
// Entries returns the entries of one sub-group, dropping the ones a caller could not address.
//
// It is a conversion, not a read: the config already holds the decoded lists, and this only picks one and removes the
// entries without a name. A missing sub-group yields an empty slice rather than nil, so a caller can range over it.
func (c Config) Entries(sub string) []LogEntry
```

```go
// Runtimes returns the layers the engine keeps, in ascending priority, dropping the ones without a name.
//
// It is a conversion: the config already holds the decoded list, and this only maps the declared mode onto Mode.
func (c Config) Runtimes() []Layer
```

「是否可用」这种判断也是转换, 而不是读取:

```go
// Usable reports whether a connection can be served at all: it needs a name, a DSN and a driver this build links.
//
// The check is a conversion of the config into a decision, not a read: nothing here touches the document, and the
// caller still decides what to do with an entry that fails.
func (c ConnConfig) Usable() bool
```

### 实现层: 只接收值

结构体命名 `<Instance>`, 由 `New<Instance>` 构造. 字段分两组, 中间空行:

```go
type DirectLog struct {
	handlers map[string]slog.Handler
	entries  []engineLog.LogEntry

	mu     sync.RWMutex
	ectx   domainEngine.EngineContext
	state  domainEngine.EngineState
	closed bool
}
```

Option 只校验自己的参数, 不补全, 也不读外部状态:

```go
// WithNoticeSize sets the ring length. It is the caller's job to read that value from the document: the engine never
// touches the configuration, so it can be constructed from a literal in a test just as well as from a file.
func WithNoticeSize(size int) NoticeOption {
	return func(n *Notice) {
		if size <= 0 {
			return
		}
		n.size = size
	}
}
```

总有一对编译期断言, 证明两个契约都实现:

```go
var (
	_ domainEngine.Engine    = (*Notice)(nil)
	_ engineNotice.Client    = (*Notice)(nil)
)
```

生命周期方法固定这一组, 顺序与文件内一致:

```go
func (x *Xxx) Group() string    { return domainEngine.GroupXxx }
func (x *Xxx) Name() string     { return domainEngine.NameXxx }
func (x *Xxx) Deps() []string   { return []string{domainEngine.GroupStore} }
func (x *Xxx) State() domainEngine.EngineState
func (x *Xxx) Init(eCtx domainEngine.EngineContext) error
func (x *Xxx) Start() error
func (x *Xxx) Restart() error
func (x *Xxx) Stop() error
```

`domainEngine.Engine` 没有 `Deinit`——释放统一走 `Stop`。

| 方法 | `DirectLog` | `FS` | `Setting` | `Notice` |
|---|---|---|---|---|
| `Group` `Name` | ✓ | ✓ | ✓ | ✓ |
| `Deps` | `store` | — | `store.fs` + `log` | — |
| `State` | ✓ | ✓ | ✓ | ✓ |
| `Init` `Start` `Restart` `Stop` | ✓ | ✓ | ✓ | ✓ |

### Init 的固定四步

`Init` 不再读配置, 只在第一步注册错误、第二步接上下文、第三步做有效性判断、第四步打一条就绪日志:

```go
func (n *Notice) Init(eCtx domainEngine.EngineContext) error {
	// 1. 注册错误, 必须传 AppCtx
	if eCtx.AppCtx != nil {
		interError.RegisterAll(eCtx.AppCtx, engineNotice.Errors...)
	}

	// 2. 接上下文与状态
	n.mu.Lock()
	n.ectx = eCtx
	// 3. 判断必需参数是否可用, 不做补全
	if n.size <= 0 {
		n.size = defaultBuf
	}
	n.ring = make([]engineNotice.Record, 0, n.size)
	n.next = 0
	n.closed = false
	n.state = domainEngine.EngineReady
	size := n.size
	n.mu.Unlock()

	// 4. 收尾一条就绪日志, 带上可供排查的键值
	slog.Default().Warn("notice engine ready", "buf", size)
	return nil
}
```

`log` 的第三步是打开处理器, 单条不可用时只掉那一条:

```go
// handlersFor opens every output of one entry and returns the bound handlers.
//
// It runs during Init rather than at construction, because the store engine has created the root directory by
// then. Level filtering belongs to the outer handler, so the serializer keeps its own default.
```

### 无默认不兜底

「兜底值」只存在于 `default.json`, 引擎侧只判断参数是否可用:

| 参数 | 缺失时 | 位置 |
|---|---|---|
| `notice.buf` | `Init` 用包内默认并照常启动 | `engine/notice` |
| `store.root` | 不追加 `WithRoot`, `Validate` 报错 | `SetupStore` |
| `store.fs.*Mode` | 不追加对应 Option, 由 `NewFileHandler` 一类使用点各自决定 | `SetupStore` |
| `log.<sub>` | 该子组不启动 | `SetupLog` |
| `setting.layers` | 只有运行时层, 无持久层 | `SetupSetting` |

判断写在使用点, 不写成「没有就给一个看起来合理的值」:

```go
// Option 内
if size <= 0 {
	return
}

// Setup 内
if noticeSettings.Buf > 0 {
	options = append(options, engineNoticeImpl.WithNoticeSize(noticeSettings.Buf))
}
```

### 生命周期中的状态

`State` 三态推进 `Pending → Ready → Running`, `Stop` 落 `Stopped`:

```go
// Start implements domainEngine.Engine. The engine serves calls synchronously, so it only moves the state.
func (n *Notice) Start() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.state = domainEngine.EngineRunning
	return nil
}
```

`Restart` 复用 `Init`, 取上下文时不持锁调用:

```go
func (n *Notice) Restart() error {
	n.mu.Lock()
	eCtx := n.ectx
	n.mu.Unlock()
	return n.Init(eCtx)
}
```

`Stop` 负责释放持有的资源与状态, 不只是改标记:

```go
// Stop implements domainEngine.Engine. The ring is process state, so stopping drops what is in it.
func (n *Notice) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.closed = true
	n.ring = n.ring[:0]
	n.state = domainEngine.EngineStopped
	return nil
}
```

### 依赖其他引擎

取引擎的键由 `EngineDep` 拼出, 断言具体能力, 失败带上键名:

```go
engineKey := domainEngine.EngineDep(domainEngine.GroupStore, domainEngine.NameDb)
engine, found := sCtx.Engines[engineKey]
if !found {
	return interError.Retrieve(engineStore.ErrKeyInvalid).
		Wrap(fmt.Errorf("engine %q is not ready", engineKey))
}

registry, ok := cast.As[engineStore.DBRegistry](engine)
if !ok {
	return interError.Retrieve(engineStore.ErrKeyInvalid).
		Wrap(fmt.Errorf("engine %q is not a database registry", engineKey))
}
```

只需一个能力时声明窄接口, 不 import 具体实现:

```go
// poolSource is the one capability the database handler needs from the store: the pool behind a key. Declaring it
// here keeps this file free of the concrete store type.
type poolSource interface {
	RawDB(key string) (*sqlx.DB, bool)
}
```

契约本身也是「一个能力一个接口」: `DirectClient` 只读、`DirectServer` 只写、`DirectLog` 合成两者。调用方取它需要的那一半。

### 向界面送达变化

引擎不直接调用界面。它把事件交给 engine bus, 由 bootstrap 的桥接上抛到 app bus, 再由 `UiEvents` 白名单决定是否到达 webview:

```go
// The record is stored first and announced second. An announcement that arrives before the entry exists would leave
// the interface showing a toast it cannot find in the sheet, which is the one ordering that must not happen.
if bus != nil {
	bus.Emit(domainBus.EventNoticeSend, record)
}
```

事件名定义在 `domain/bus/event.go`, 需要送达界面时加入 `UiEvents`:

```go
var UiEvents = []string{EventSettingChanged, EventSessionExpired, EventNoticeSend}
```

引擎不是服务, 因此不做 binding 暴露: 它经事件到达界面, 经 `sCtx.Engines` 到达 service。

### 数据结构命名

| 类别 | 命名 | 位置 | 标签 | 示例 |
|---|---|---|---|---|
| dto | `<X>Dto` | `domain/engine/<id>/shared.go` | `json` | `LogRecordDto` |
| 配置 | `Config` | `domain/engine/<id>/shared.go` | `json` | `FsConfig` `DbConfig` `ConnConfig` |
| 域内值 | `<X>` | `domain/engine/<id>/shared.go` | `json` | `LogEntry` `Record` `Layer` |
| 引擎内部结构 | `<x>` | `engine/<id>/shared.go` | — | `logEntry` `logOutput` |

引擎的 `shared.go` 里**不出现 dto**: dto 是面向前端 binding 的结构, 引擎不参与 binding。同名字段出现在事件载荷里, 靠 `json` 标签序列化即可。

```go
// Record is one notice the engine keeps.
//
// It is a value the engine owns, not a dto: nothing here is designed around a binding. The interface receives the
// same fields over an event, and the notice service that follows reads them through the client.
type Record struct { … }
```

### 日志

| 场景 | 级别 | 键 |
|---|---|---|
| 引擎就绪 | `Warn`(与其他启动日志一致可查) | `buf` / `database` `limit` |
| 单条配置不可用 | `Warn` | `name` `path` `err` |
| 驱动未链接 | `Warn` | `name` `driver` |
| 单个输出打开失败 | `Warn` | `path` `err` |
| 配置解码失败 | `Warn` | `key` `err` |

```go
slog.Default().Warn("notice engine ready", "buf", size)
slog.Default().Warn("engine config rejected", "key", key, "err", err)
slog.Default().Warn("log output skipped", "path", path, "err", interError.Retrieve(engineLog.ErrKeyHandlerFailed).Wrap(err))
```

每个失败点都留一条, 使「功能静默失效但日志干净」不会发生。配置项坏掉时跳过该条并继续, 不让一个坏条目拦住整个引擎。

### 注册与装配

| 环节 | 位置 | 形式 |
|---|---|---|
| 引擎 Setup | `backend/bootstrap/setup.go` | `func Setup<Name>(cfg, ...)` → `groupConfig[...]` → `er.Register(engineXxxImpl.NewXxx(opts...))` |
| 调用顺序 | `backend/bootstrap/boot.go` | 依赖在前: `SetupStore` → `SetupLog` → `SetupSetting` → `SetupNotice` |
| 键常量 | `backend/domain/engine/key.go` | `Group<Id>` / `Name<Instance>` + `var` 段的 `EngineDep` |

```go
// SetupNotice registers the notice engine.
//
// The ring length is read here, not inside the engine: the document is the bootstrap's business, and the engine only
// receives the number. A value that is absent or unusable is left out, so the engine falls back to its own default.
func SetupNotice(cfg domainContext.ConfigContext, er *domainRuntime.EngineRuntime) {
	noticeSettings := groupConfig[engineNotice.Config](cfg, "default.runtime.engines.notice")
	options := make([]engineNoticeImpl.NoticeOption, 0, 1)
	if noticeSettings.Buf > 0 {
		options = append(options, engineNoticeImpl.WithNoticeSize(noticeSettings.Buf))
	}
	er.Register(engineNoticeImpl.NewNotice(options...))
}
```

需要包装钩子时用 `domainEngine.EngineHook`, 把「装配期才知道的事」留在 bootstrap:

```go
// hookLog wraps a log engine so its output directories are created before it starts. Where a log file lives
// is a bootstrap decision, so it stays here instead of inside the engine phases.
hookLog := func(engine domainEngine.Engine, entries []engineLog.LogEntry) domainEngine.Engine
```

同理, 读嵌入树这类「怎么挂载」的决策也留在 bootstrap:

```go
// The entry point mounts the data directory, so every migration resource is a path below it.
appCtx := store.HookContext().EngineContext.AppCtx
if appCtx == nil || appCtx.EmbedContext["/data"] == nil {
	slog.Default().Warn("db migrations skipped", "reason", "no embedded tree")
	return nil
}
assets := appCtx.EmbedContext["/data"]
```
