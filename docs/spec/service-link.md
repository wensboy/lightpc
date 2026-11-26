---
name: '服务链路'
description: '新增服务处理链路.'
---

> ai agent 能够将补充的示例追加到 Example 段下, 无明确编辑其他段时不允许编辑.

## Step

1. 根据用户指令完善指定服务所需的数据库相关设计, 将表定义和约束以 migration 的形式记录到指定数据库迁移目录下. 例如: `data/migrate/../*sql`.
2. 根据用户指令在指定的服务实现目录下实现所有的服务功能(严格按照用户定义的实现, 严格1:1). 例如: `backend/service/<service-name>/, backend/domain/service/<service-name>/`.
3. 在 `backend/api/ui/` 下暴露服务, 形式与其他服务保持一致.
4. 在 `frontend/src/service/` 下与其他服务实现类似进行前端服务封装.
5. 在必要的 features 下调用服务.

## Hint

- 几个中断点需要用户rv输出和纠正:
  + 服务数据库表和约束定义(数据库迁移文件), 可能进行一定程度的定义微调, 后续任务以最新的定义为准进行调整.
  + 后端服务实现和暴露, 校验功能是否 1:1 实现.
- package 名称规范:
  + backend/service/<service_id>/ -> `service<Service_id>Impl`
  + backend/domain/service/<service_id>/ -> `service<Service_id>`
  + 引入时通过 alias 显式标识
- 服务内错误定义到 `backend/domain/service/<service_id>/shared.go` 中.
- 服务内共享处理定义到 `backend/service/<service_id>/shared.go` 中. 共享处理包含:
  + 非导出/导出的共享常量, 变量, 类型及其绑定的方法, 接口
  + 非导出/导出的无状态工具函数
- 服务内数据访问规范:
  + 所有的 dto, dao, entity 都要至少映射 json 标签, dao 通常还需要映射 db 相关标签.
  + dto 用于统一所有的前端通过 binding 访问后端服务接口的数据传输结构(包含请求结构和返回结构), 一定存在 dto 结构并必须定义在 `backend/domain/service/<service_id>/shared.go` 下.
  + entity 为后端 service 处理过程中的结构转换, 可能存在并必须定义在 `backend/service/<service_id>/model.go` 下.
  + dao 为数据库表和结构的orm映射, 只有需要调用 db 时存在并必须定义在 `backend/service/<service_id>/model.go` 下.
  + 所有数据库操作和 orm 处理定义到 `backend/service/<service_id>/repo.go` 下. 所有数据库操作 + orm 映射为非导出 service 结构 methods. **只有需要 db 调用时才存在该文件.**
  + 不分 handler, service, repo 三层, 压缩为: logic, repo. logic 实际上等同于 handler + service.
- 服务内数据定义规范:
  + service 结构格式: xxxService. 该结构总是需要实现 domainService.Service 接口. -> 例如: HelpService, SessionService
  + dto, dao 结构格式: xxx<Dto|Dao>. -> 例如: UserDto, UserProfileDto, UserDao, UserProfileDao
  + entity 结构格式: xxx. -> 例如: User, UserProfile
  + **永远不要定义不同的2个结构的转换作为一个单独函数, 直接在对应的函数体内显示转换, 字段全为导出定义.**

## Example

**当前 help service 的结构完美符合规范, 后续所有服务参照类似结构实现. 以下示例同时取自 `session` `setting` `log` `notice`, 凡涉及数据库与事件的地方以它们为准**

以下共性抽取自 `help`, `session`, `setting` 三个已完成的服务实现, 是对上述 Step 与 Hint 的落实方式.

### 文件职责

```
backend/domain/service/<id>/
├── <id>.go      服务契约: <Id>Client 接口(仅此处), 无实现
└── shared.go    错误 key + Errors + 全部 dto 结构

backend/service/<id>/
├── logic.go     <Id>Service 结构 + 生命周期 + 全部业务方法
├── model.go     entity / dao 结构(与解码/映射绑定), 无业务方法
├── repo.go      数据库操作; 仅当服务直接访问 db 时存在
└── shared.go    常量(配置 key) + 无状态工具函数 + 共享类型
```

| 服务 | domain 文件 | impl 文件 |
|---|---|---|
| `help` | `help.go` `shared.go` | `logic.go` `shared.go` |
| `session` | `session.go` `shared.go` | `logic.go` `model.go` `repo.go` `shared.go` |
| `setting` | `setting.go` `shared.go` | `logic.go` `shared.go` |
| `log` | `log.go` `shared.go` | `logic.go` `model.go` `repo.go` `shared.go` |
| `notice` | `notice.go` `shared.go` | `logic.go` `shared.go` |

`model.go` 与 `repo.go` 成对出现, 因为两者都是「服务自己碰数据库」的产物: `session` 与 `log` 各有 dao 和 SQL, 其余三个服务既不映射表也不查询, 因此只有 `logic.go` 与 `shared.go`. 一个服务若只读一个已解析好的引擎状态(`notice` 读引擎的环), 同样不需要 `model.go`.

### 契约层(domain)

契约包只放声明, 不含任何逻辑:

| 声明 | 位置 | 命名 |
|---|---|---|
| 服务接口 | `<id>.go` | `<Id>Client` |
| 错误 key | `shared.go` | `ErrKey<Situation>` |
| 错误列表 | `shared.go` | `Errors` |
| dto | `shared.go` | `<X>Dto` |

接口方法以动词开头, 每个方法一行注释说明返回什么:

```go
type HelpClient interface {
	// Payload returns every documentation set, ordered by version.
	Payload() (DocsPayloadDto, error)
	// Read returns the markdown of one document of the given version.
	Read(version, path string) (string, error)
}
```

五个服务的接口与所在文件:

| 服务 | 接口 | 文件 | 方法 |
|---|---|---|---|
| `help` | `HelpClient` | `help/help.go` | `Payload` `Read` |
| `session` | `SessionClient` | `session/session.go` | `Register` `Login` `Logout` `Unregister` `Current` |
| `setting` | `SettingClient` | `setting/setting.go` | `Payload` `Save` |
| `log` | `LogClient` | `log/log.go` | `Listeners` `Records` `Clear` `SetFilter` |
| `notice` | `NoticeClient` | `notice/notice.go` | `History` `Send` `Remove` |

错误 key 的三段式命名, 中间是 service id:

```go
const (
	ErrKeyDatabaseMissing  = "error::service.session::database_missing"
	ErrKeyAccountExists    = "error::service.session::account_exists"
	ErrKeyCredentials      = "error::service.session::credentials"
	ErrKeyNotAuthenticated = "error::service.session::not_authenticated"
)

var Errors = []interError.Option{
	{Key: ErrKeyDatabaseMissing, Code: interError.ErrCodeNotExist, Msg: "session: database unavailable"},
	{Key: ErrKeyCredentials, Code: interError.ErrCodeInvalidParam, Msg: "session: invalid credentials"},
}
```

`Errors` 的注释写明为何放在契约层:

```go
// The list lives with the contract rather than with the implementation: a caller that compares against a key
// should not have to import the implementation to spell it.
```

dto 一律带 `json` 标签, 且**返回切片时不用 nil**:

| 服务 | dto |
|---|---|
| `help` | `DocsPayloadDto` `DocsVersionDto` `DocsGroupDto` `DocsItemDto` |
| `session` | `StatusDto` `AuthMetaDto` `RegisterInputDto` `LoginInputDto` |
| `setting` | `SettingPayloadDto` `SettingThemeDto` |
| `log` | `LogRecordDto` `LogRecordsDto` `LogQueryDto` `FilterParamsDto` `ListenerDto` |
| `notice` | `NoticeItemDto` `NoticeHistoryDto` `NoticeSendInputDto` `NoticeRemoveInputDto` |

请求结构以 `InputDto` 结尾, 应答结构以 `Dto` 结尾, 一个方法一个结构而不是把两者合并: `LogQueryDto` 只出现在入参位置, `LogRecordsDto` 只出现在返回位置, 读代码时方向一目了然.

### 实现层(service)

结构体命名 `<Id>Service`, 由 `New<Id>Service()` 构造. 字段分两组, 中间空行:

```go
// SessionService owns the accounts and the one session this process signed in.
type SessionService struct {
	mu    sync.RWMutex
	store engineStore.DBStore
	bus   domainBus.Bus
	// pair is the one session this process signed in. A single client means a single live pair, so no lookup
	// structure is needed: the pair either matches or it does not.
	pair       serviceSession.AuthMetaDto
	// expiresAt is the deadline of that pair, zero when the window never closes.
	expiresAt  time.Time
	timer      *time.Timer
	expiration time.Duration
	softDelete bool
	state      domainEngine.EngineState
}
```

字段顺序是「先依赖与业务状态, 空行, 再同步原语」的反面——这里把 `mu` 放在最前, 因为它是所有字段的守卫, 读结构体时先看到锁, 再看它保护什么.

生命周期方法固定这一组, 顺序与文件内一致:

```go
func (s *XxxService) Group() string            { return domainService.GroupXxx }
func (s *XxxService) Deps() []string           { return []string{domainEngine.StoreFs} }
func (s *XxxService) State() domainEngine.EngineState   // 需要状态机时才有
func (s *XxxService) Init(sCtx domainService.ServiceContext) error
func (s *XxxService) Start() error             // 需要缓存/连接时才有
func (s *XxxService) Stop() error
func (s *XxxService) Deinit() error            { return nil }   // 或转调 Stop
```

五个服务各自实现了哪些:

| 方法 | `HelpService` | `SessionService` | `SettingService` | `LogService` | `NoticeService` |
|---|---|---|---|---|---|
| `Group` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `Deps` | `store.fs` | `store.db` | — | `store.db` | `notice` + `setting` |
| `State` | — | ✓ | — | ✓ | — |
| `Init` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `Start` | — | ✓ | — | ✓ | — |
| `Stop` | — | ✓ | — | ✓ | — |
| `Deinit` | ✓ | ✓ | ✓ | ✓ | ✓ |

`State` / `Start` / `Stop` 三者同进同出: 一个服务需要状态机, 就说明它有需要启动与释放的运行期资源(会话的计时器、日志的边界表). 只做一次性解析的服务三者皆无, `Deinit` 写成 `return nil`.

`Deps` 缺失的 `SettingService` 在 `Init` 内直接取引擎, 靠 `Init` 的错误返回暴露依赖问题; 新服务应声明 `Deps`, 让调度器在注册阶段就校验. `NoticeService` 是唯一依赖两个引擎的例子, 缺任一个都按各自的 `ErrKey` 报错, 而不是带着半个能力启动.

### Init 的固定五步

```go
func (s *XxxService) Init(sCtx domainService.ServiceContext) error {
	// 1. 注册错误, 必须传 AppCtx
	if sCtx.AppCtx != nil {
		interError.RegisterAll(sCtx.AppCtx, serviceXxx.Errors...)
	}

	// 2. 取引擎: 键由 EngineDep 拼出, 不写字面量
	key := domainEngine.EngineDep(domainEngine.GroupStore, domainEngine.NameFs)
	engine, found := sCtx.Engines[key]
	if !found {
		return interError.Retrieve(engineStore.ErrKeyInvalid).
			Wrap(fmt.Errorf("engine %q is not ready", key))
	}

	// 3. 断言具体能力, 失败同样带上键名
	fileStore, ok := cast.As[engineStore.FileStore](engine)
	if !ok {
		return interError.Retrieve(engineStore.ErrKeyInvalid).
			Wrap(fmt.Errorf("engine %q is not a file store", key))
	}
	s.fileStore = fileStore

	// 4. 读配置: 缺失只告警, 不阻止启动
	if sCtx.AppCtx == nil || sCtx.AppCtx.ConfigContext == nil {
		slog.Default().Warn("xxx service without configuration")
		return nil
	}

	// 5. 收尾一条就绪日志, 带上可供排查的键值
	slog.Default().Warn("xxx service ready", "roots", s.roots, "versions", len(s.versions))
	return nil
}
```

`session` 在此基础上把「何时过期」做成了一个倒计时器, 而不是一个带 TTL 的缓存:

```go
// arm schedules the expiry of the current pair. A zero window means the session never closes, so no timer is armed
// and the deadline stays zero.
func (s *SessionService) arm() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	if s.expiration <= 0 {
		s.expiresAt = time.Time{}
		return
	}

	s.expiresAt = time.Now().Add(s.expiration)
	account := s.pair.Account
	s.timer = time.AfterFunc(s.expiration, func() {
		s.expire(account)
	})
}
```

三个理由, 都来自实际踩过的坑:

| 项 | 缓存 | 计时器 |
|---|---|---|
| 到期时刻 | 由清理协程决定, 与实际 TTL 有偏差 | `AfterFunc` 精确在截止点触发 |
| 事件 | 只能事后从驱逐回调得知 | 立即执行回调, 可同步发总线事件 |
| 状态 | 缓存里是否有条目 | `pair` + `expiresAt`, 一个进程一份 |

到期时向总线发事件, 让界面跳到登录页:

```go
// Notify the interface before dropping the pair: a listener that reloads the view asks Current, and by then the
// answer must already be "signed out".
s.bus.Emit(domainBus.EventSessionExpired, account)
```

顺序与日志服务的「先存后发」相同——先让状态成立, 再通知别人.

### 依赖声明与取用

`Deps()` 与 `Init` 里取引擎的键必须一致, 都走 `EngineDep`:

```go
func (s *HelpService) Deps() []string { return []string{domainEngine.StoreFs} }

// Init 内
engineKey := domainEngine.EngineDep(domainEngine.GroupStore, domainEngine.NameFs)
```

多个依赖时各自断言, 断言失败的错误带上引擎键与期望能力:

```go
settingKey := domainEngine.EngineDep(domainEngine.GroupSetting, domainEngine.NameSetting)
settingEngine, found := sCtx.Engines[settingKey]
instance, ok := cast.As[engineSetting.Setting](settingEngine)
```

### 配置读取

配置 key 定义为 `shared.go` 的常量, 不写在方法体内:

| 服务 | 常量 | 值 |
|---|---|---|
| `help` | `rootsKey` | `default.runtime.services.docs` |
| `session` | `expirationKey` | `default.runtime.services.session.expiration` |
| `session` | `databaseKey` | `default.runtime.services.session.database` |
| `session` | `softDeleteKey` | `default.runtime.services.softDelete` |
| `log` | `databaseKey` | `default.runtime.services.log.database` |
| `log` | `limitKey` | `default.runtime.services.log.limit` |
| `notice` | `historyKey` | `general.notification.history` |

`notice` 是唯一读**设置文档**而不是**引擎文档**的服务: 条数是一个用户偏好, 可能被任意一层覆盖, 所以经设置引擎的 `Resolve` 取值, 而不是直接 `Lookup`.

带单位的字段一律经 `internal/config` 解析, 缺省回退到硬编码的默认值而非零值:

```go
// durationValue reads a value written as "<number> <unit>", falling back when it is absent or malformed.
func durationValue(cfg domainContext.ConfigContext, key string, fallback time.Duration) time.Duration {
	text := stringValue(cfg, key)
	if text == "" {
		return fallback
	}
	parsed, err := config.ParseDuration(text)
	if err != nil {
		return fallback
	}
	return parsed
}
```

数值字段的形态要按来源区分: 引擎文档解出来是 `float64`, 设置引擎的 `Resolve` 则可能给出任意层里的原始类型, 所以收敛函数接受三种:

```go
// intValue narrows a resolved setting onto an int, reporting false when the value is of another shape.
func intValue(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}
```

`cast.To` 是纯类型断言, `cast.As` 才做可转换翻译. 从 `Lookup` 取标量时用后者, 否则 `float64` 永远匹配不上 `int`.

### 返回值与错误

| 约定 | 示例 |
|---|---|
| 集合永不为 nil | `serviceSetting.SettingPayloadDto{Versions: []DocsVersionDto{}}` |
| 业务错误用契约 key | `interError.Retrieve(serviceHelp.ErrKeyVersionUnknown)` |
| 需要原因时 `Wrap` | `interError.Retrieve(engineStore.ErrKeyInvalid).Wrap(fmt.Errorf("create root: %w", err))` |
| 底层驱动的错误直接上抛 | `return "", err` |

```go
// Payload implements serviceHelp.HelpClient.
func (s *HelpService) Payload() (serviceHelp.DocsPayloadDto, error) {
	if s.fileStore == nil {
		return serviceHelp.DocsPayloadDto{Versions: []serviceHelp.DocsVersionDto{}}, interError.Retrieve(engineStore.ErrKeyInvalid)
	}
	return serviceHelp.DocsPayloadDto{Versions: s.versions}, nil
}
```

### 只读服务的缓存

资源在进程内不会变时, `Init` 末尾一次性建索引, 避免每次调用重复解析:

```go
// The sets are embedded, so they never change while the process runs. Init reads every meta.json once and keeps
// both the ordered list the reader shows and the version to directory lookup that Read needs.
type HelpService struct {
	versions  []serviceHelp.DocsVersionDto
	byVersion map[string]string
}
```

单个条目损坏时跳过并告警, 不中断整体:

```go
parsed, err := decodeMeta(data)
if err != nil {
	slog.Default().Warn("help meta invalid", "path", location, "err", err)
	continue
}
```

### 数据访问(session 与 log 为范例)

`repo.go` 内全部是**非导出**的 service 方法, 一个方法对应一个意图, SQL 不外泄:

```go
// session
func (s *SessionService) findAccount(store engineStore.DBStore, account string) (*UserDao, error)
func (s *SessionService) latestSignIn(store engineStore.DBStore, auth serviceSession.AuthMetaDto) (*UserSessionDao, error)
func (s *SessionService) findAccountByUID(store engineStore.DBStore, uid, account string) (*UserDao, error)
func (s *SessionService) insertAccount(store engineStore.DBStore, user UserDao, expiration int64) error
func (s *SessionService) record(store engineStore.DBStore, action Action, auth serviceSession.AuthMetaDto, expiration int64) error
func (s *SessionService) removeAccount(store engineStore.DBStore, uid string, at int64) error
func (s *SessionService) deleteAccount(store engineStore.DBStore, uid string) error
```

```go
// log
func (s *LogService) latest(store engineStore.DBStore, listener string, since int64, limit int) ([]serviceLog.LogRecordDto, error)
func (s *LogService) listeners(store engineStore.DBStore) ([]string, error)
func (s *LogService) boundaryOf(listener string) int64
```

`store` 作为第一个参数显式传入, 而不是从结构体字段读: 这些方法都要在同一个事务或同一份快照上工作, 让调用方交出句柄比隐式依赖字段更清楚. `boundaryOf` 是例外——它只碰内存里的边界表, 不接触数据库, 放在 `repo.go` 是因为它是 `latest` 的配套.

带 `LIMIT 1` 的查询不写成 `for rows.Next() { return }`, 一律 `if !rows.Next()`:

```go
// The statement carries LIMIT 1, so the cursor holds either one row or none. Reading it as "at most one" keeps
// the exit paths explicit instead of hiding them inside a loop that only runs once.
if !query.Rows.Next() {
	return nil, query.Rows.Err()
}
```

多语句写入放在事务里, 由 repo 方法整体提交, logic 层只表达业务意图:

```go
tx, err := store.Begin(ctx)
defer func() { _ = tx.Rollback() }()
for _, statement := range statements {
	if _, err := tx.Dispatch(ctx, statement); err != nil {
		return err
	}
}
return tx.Commit()
```

软删与硬删由 `repo.go` 提供两个方法, 分支留在 `logic.go`:

```go
var err error
if softDelete {
	err = s.removeAccount(store, session.uid, time.Now().Unix())
} else {
	err = s.deleteAccount(store, session.uid)
}
```

### entity 与 dao

| 类别 | 命名 | 位置 | 标签 | 示例 |
|---|---|---|---|---|
| dto | `<X>Dto` / `<X>InputDto` | `domain/service/<id>/shared.go` | `json` | `StatusDto` `LogQueryDto` |
| entity | `<X>` | `service/<id>/model.go` | 视用途 | `meta` `entry` |
| dao | `<X>Dao` | `service/<id>/model.go` | `json` + `db` | `UserDao` `LogDao` |

`help` 目前没有 `model.go`: 它读到的 `meta.json` 直接解进 dto, 中间不存在需要命名的结构. 一旦某个解析结果要在多个方法间传递, 就必须提到 `model.go` 并给它一个名字.

dao 映射整张表, 而不是只映射当前流程用到的列:

```go
// UserDao maps a row of users.
//
// Every column of a row is present, including the ones the current flow does not read: the struct is the single
// description of a table, so a column added by a migration has exactly one place to land and a reader never has
// to cross reference the SQL to know what a row holds.
type UserDao struct {
	ID        int64         `json:"id" db:"id"`
	UID       string        `json:"uid" db:"uid"`
	Uname     string        `json:"uname" db:"uname"`
	Account   string        `json:"account" db:"account"`
	Password  string        `json:"-" db:"password"`
	CreatedAt int64         `json:"createdAt" db:"created_at"`
	UpdatedAt int64         `json:"updatedAt" db:"updated_at"`
	DeletedAt sql.NullInt64 `json:"deletedAt" db:"deleted_at"`
}
```

两处细节: 密码列用 `json:"-"` 防止结构体被序列化时泄露哈希; 可空列用 `sql.NullInt64` 而不是指针, 因为值直接来自 `database/sql`, 保留包装类型让扫描不需要转换步骤.

转换不抽函数, 在发生处直接构造:

```go
s.versions = append(s.versions, serviceHelp.DocsVersionDto{
	Version:    parsed.Version,
	Index:      parsed.Index,
	Deprecated: parsed.Deprecated,
	Tags:       parsed.Tags,
	Toc:        parsed.Toc,
})
```

### 日志

| 场景 | 级别 | 键 |
|---|---|---|
| 服务就绪 | `Warn`(与其他启动日志一致可查) | `roots` `versions` / `database` `expiration` / `limit` / `history` |
| 配置缺失 | `Warn` | `key` |
| 单个资源不可用 | `Warn` | `path` `err` |
| 令牌过期 | `Warn` | `account` |

```go
slog.Default().Warn("help service ready", "roots", s.roots, "versions", len(s.versions))
slog.Default().Warn("session service ready", "database", key, "expiration", s.expiration.String())
slog.Default().Warn("log service ready", "database", key, "limit", s.limit)
slog.Default().Warn("notice service ready", "history", s.limit())
```

键值取「排查时会想知道的那个数」: `log` 给数据库键与条数上限, `notice` 给当前生效的历史长度——这两个服务的行为都由配置决定, 一条日志就能看出配置有没有生效.

每个失败点都留一条, 使「界面空白但日志干净」不会发生: `help meta unreadable`, `help meta invalid`, `db driver skipped`, `db duplicate key skipped` 都属此类.

### 暴露与前端封装

| 环节 | 位置 | 形式 |
|---|---|---|
| 注册 | `backend/api/ui/entry.go` | `serviceXxxImpl.NewXxxService()` → `sr.Register(...)` → `application.NewService(...)` |
| 前端封装 | `frontend/src/services/<id>.ts` | Zod 校验 + 方法一一对应 |
| 调用 | `features/<feature>/` | 只经封装层, 不直接碰 binding |

当前已暴露: `greet` `setting` `help` `session` `log` `notice`. 引擎**不**在此列——它经事件总线到达界面, 经 `sCtx.Engines` 到达服务, 没有 binding.

```ts
// services/help.ts
export interface HelpService {
  payload(): Promise<DocsVersion[]>
  read(version: string, path: string): Promise<string>
}

export const helpService: HelpService = {
  async payload() {
    const parsed = helpPayloadSchema.safeParse(await HelpServiceBinding.Payload())
    if (!parsed.success) {
      throw new Error(`help payload rejected: ${parsed.error.message}`)
    }
    return parsed.data.versions
  },
}
```

binding 的导出名由 `wails3 generate bindings` 决定(当前为 `<Id>Service` 命名空间), 封装层用别名导入, 使组件不依赖生成物的具体形状.
