# 后端规则

> 核心目的：定义后端代码与设计抽象必须遵守的规则。

## 必须

1. 每一个可以作为选项暴露的结构体字段，都必须使用返回原始引用的 `withxxx` 设置器。
2. `domain/` 目录只能包含与领域强相关的声明：常量、结构体、**接口**和变量。实现放在 `internal/`、`service/`、`engine/` 或同类目录。
3. 文档注释与文件顶部注释必须全部使用英文书写。
4. 实现需求时，优先标准库，其次发布已久且 star 数较高的第三方库，最后才在合适目录自行实现算法。
5. 包名按下表命名，引入时用同名的显式别名：

   | 目录 | 包名 |
   | --- | --- |
   | `backend/service/<id>/` | `service<Id>Impl` |
   | `backend/domain/service/<id>/` | `service<Id>` |
   | `backend/engine/<sub>/` | `engine<Sub>Impl` |
   | `backend/domain/engine/<sub>/` | `engine<Sub>` |

6. 与标准库同名或需要消歧的包，用领域限定的别名；`internal/` 下的包冲突时在别名前加 `inter` 前缀。
7. 每个服务按下表组织文件，缺一不可的只有 `logic.go` 与 `shared.go`：

   | 文件 | 内容 | 存在条件 |
   | --- | --- | --- |
   | `logic.go` | `<Id>Service` 结构、生命周期、全部业务方法 | 总是 |
   | `shared.go` | 常量、共享类型及方法、无状态工具函数 | 总是 |
   | `model.go` | entity 与 dao（含解码/映射方法） | 有内部结构时 |
   | `repo.go` | 全部数据库操作，为非导出 service 方法 | 直接访问 db 时 |

8. 传输结构一律带 `Dto` 后缀并映射 `json` 标签，定义在 `domain/service/<id>/shared.go`；entity 无后缀；dao 带 `Dao` 后缀并同时映射 `json` 与 `db` 标签。
9. 服务自有错误定义在 `domain/service/<id>/shared.go`，key 为 `error::service.<id>::<situation>`，并汇总为 `Errors []interError.Option`。
10. 每个引擎暴露一个 `<Name>HookContext`（定义在对应 `domain/engine/<sub>/` 文件），由 `HookContext()` 返回。
11. 与引擎阶段无关的资源准备（建目录、建文件、校验）一律放在 bootstrap 的 `PreStart` 钩子里，`Init` 只注册错误与保存上下文。
12. 引擎交给消费者之前必须经 `domainEngine.Unwrap` 解包，使调用方能断言具体类型。
13. 返回切片时使用 `make([]T, 0)` 而非 `nil`——`nil` 序列化为 `null`，前端校验会拒绝整份负载。
14. 带单位的配置值写成 `"<number> <unit>"`，统一经 `internal/config` 解析，调用处不自行 `strconv`。
15. 带 `TODO` 注释的代码必须严格按描述实现，实现方式不限，但函数签名不得更改。

## 禁止

1. 添加不记录设计决策的解释性注释。
2. 为确认只调用一次、且无复用可能的逻辑抽取函数，例如仅创建一个结构体实例的一次性构造函数。
3. 为两个结构之间的转换单独定义函数；在发生的函数体内直接构造，字段全部导出。

## 应当

1. 配置文件优先 JSON，结构化文件优先 YAML；每个此类文件都在 `data/conf/schema/` 下有同名 schema，且符合 RFC 定义的 JSON Schema。
2. `data/template` 下的文件使用 `text/template`，扩展名用 `tpl`。
3. 尽可能在 `main.go` 中声明所有 `embed.FS`，使嵌入资源集中管理。
4. `internal/cast` 下 `cast.go` 提供复杂结构转换，`fallback.go` 提供带回调（默认值）的转换。~~转换一般应直接用 `spf13/cast`，包装函数也基于它构建。~~
5. key 指向有效值时用 `.` 分隔；无对应值或值无效时用 `::`；多语义段时段内用 `.`、段间用 `:`。
6. 用于查找值的 key 可以是字面字符串；不用于查找的 key（如事件名）应集中为常量。
7. 每当 `data/conf/*.json` 变更，都要校验配置与 schema 同步并更新 schema。
8. `backend/engine/` 下的结构体用空行分隔基础字段与核心字段：核心字段自上而下，基础字段自下而上。
9. 单个源文件内的声明自上而下按 `[const, var, type, exported_obj, unexported_obj]` 排列；`type` 与其方法相邻，可与 `exported_obj` 交错。
10. 引擎配置声明的所有基于协议的资源都通过 `store.fs` 解析，例如 `embed://<path>`。
11. 数据库按 `driver.name` 注册并暴露，重复 key 的检查必须在打开连接之前——文件型驱动的打开动作本身就会创建文件。
12. 驱动名在 `domain/engine/store/db.go` 用枚举声明；未在代码枚举中的条目跳过并告警，不视为致命错误。
13. `migrations` 只接受目录；目录内无 `.sql` 文件时视为「尚无迁移」直接跳过，而不是报错。
14. 需要到达界面的事件登记在 `domain/bus.UiEvents`；跨 runtime 转发时，`ServiceEventContext` 与 `EngineEventContext` 必须带上 `AppEbus`。
15. 日志用 `slog` 的键值形式；每个可能失败的分支至少留一条，使「界面空白但日志干净」不会发生。
16. 只读资源的解析结果在 `Init` 中一次建好索引，避免每次调用重复解析；单个条目损坏时跳过并告警。
17. 遍历游标时若 SQL 带 `LIMIT 1`，写成 `if !rows.Next()`，不要把 `return` 藏进只跑一次的循环。
18. 测试文件放在外部测试包（`<pkg>_test`），需要内部状态时才用同包。

## 不应

1. 在 `Init` 中做与引擎阶段无关的事：创建目录、写入文件、读取外部资源来校验可用性。
2. 把包装器（如 `EngineHook`）当作实体交给消费者——它只拥有生命周期，断言具体类型前必须先解包。
3. 依赖缓存清理回调来通知「已过期」：清理按自己的节奏运行，短周期会延迟；定时器更贴近截止时间。
4. 在引用自身已明确知晓的场合重复发事件，例如登出时再通知一次「已登出」。
