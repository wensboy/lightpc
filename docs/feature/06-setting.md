# 06 - setting

> 核心目的: 设置抽象.
>
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

- `domain/engine/setting.Setting`: 分层设置中心的契约, 提供 `Layers()` / `Layered(key)` / `Resolve(key)` / `Set(layer, key, value)` / `Load(layer, values)` / `Keys()` / `Flush()`.
- `domain/engine/setting.Layer`: 单层描述, 由 `Name` / `Mode`(`ModeReadOnly` / `ModeWriteOnly` / `ModeReadWrite`) / `Path` 组成; `Path` 使用 `proto://` 资源约定, 内存层留空.
- `domain/engine/setting.Entry[T]`: 一份贡献, 记录 `Layer` / `Mode` / `Value`; 泛型只在解析侧有意义, 引擎按 `any` 传递.
- `domain/service.SettingPayload`: 交付前端的负载, 由 `Schema` / `Values` / `Defaults` 三部分组成.
- `domain/service.SettingClient`: 前端的调用面, 仅 `Payload()` 与 `Save(key, value)` 两个动作.
- `engine/settingimpl.Setting`: 契约实现, 内部以 `map[string][]any` 存放每个 key 的分层贡献.
- `engine/settingimpl.LayersFrom(cfg)`: 从 `default.runtime.setting.layers` 解析层列表, 供引导层注册.
- `service/setting.SettingService`: 对外服务, 组装 `Payload`, 承接 `Save`, 并经 `store.fs` 读取 schema 资源.

## 逻辑

1. **初始化**: `Init` 先对每个文件层调用 `ensureLayer` 补齐资源(`embed://` 只校验, `local://` 建目录并按可写性写入空文档), 再逐层 `loadLayer` 并启动 watch; `runtime` 层若未注册则追加到顶层.
2. **读取**: `Resolve` 从最高优先级层向下找第一个贡献; `Keys` 汇总所有有贡献的 key 并升序返回.
3. **写入**: `Set` 把值写入指定层, 目标为 `runtime` 时置脏标记; 同 key 的覆盖关系完全由层序保证, 引擎不做额外比较.
4. **下沉**: `Flush` 只把 `runtime` 层的贡献合并进 `user` 层, 用 `nest` 还原成嵌套文档写盘, 随后清空 `runtime` 并清除脏标记.
5. **持久化触发**: 手动(`Save` 后异步下沉, 不阻塞 UI)、间隔(`watchFlush` 每 30s 检查脏标记)、退出(`Stop` 前兜底一次).
6. **外部变更**: watch 发现 mtime 变化后 `loadLayer` 只替换该层的贡献, 并在锁外逐 key `notify`, 以 `setting::changed` 事件广播.
7. **对外服务**: `Payload` 返回 `schema`(取自 `default.runtime.setting.schema` 指向的 `embed://` 资源) + `values`(引擎聚合, 已嵌套还原) + `defaults`(`default` 层快照); `Save(key, value)` 接受点路径, 写入 `runtime` 后异步下沉.

## 注意点

- 层优先级由数组顺序决定, 索引越大越高, 配置按升序书写, 不要依赖层名字母序.
- `runtime` 层不写进配置: 它是"未保存编辑"的载体, 由引擎在 `Init` 阶段自建.
- 文件是嵌套结构而引擎按点路径寻址, 两者靠 `nest` / `flatten` 互转; 改动任一侧都要同步另一侧, 否则读写不对称.
- `Deps()` 只声明 `store.fs`. 依赖可选引擎(如 `log.direct`)会在该引擎缺席时让整条管线的依赖解析失败, 后续引擎全部跳过 Init.
- `Flush` 只搬 `runtime` 的 key, 用户层里未被编辑的键不会被重写; 这也是并发编辑与文件重载可以各自独立的原因.
- `Payload` 不做清洗: 用户层的非法值原样送达前端, 由渲染层回退到 `defaults` 并把它提升回 `runtime`, 形成闭环.

## 细节

- 不同 key 的 value 形态不同且无法在编译期确定, 因此 `Setting` 接口与实现都不带泛型参数, 贡献以 `any` 传递.
- 存储结构为 `map[string][]any`, 每个元素实际是 `setting.Entry[any]`; 调用方用 `cast.As[setting.Entry[any]]` 取回结构, 再按自己的约定解析 `Value`, 引擎本身从不读取 payload.
- `Entry[T]` 保留泛型, 用于需要类型确定性的解析侧; 引擎只负责按层序返回结构, 不承担解码职责.
- 层序由 `Entry.Layer` 反查层索引维护: 插入时按索引排序, `Resolve` 直接取切片末元素即最高优先级.
- `Init` 在读取每个层之前调用 `ensureLayer`: `embed://` 层只校验存在(内嵌内容不可写), `local://` 层在缺失时用 `MkdirAll` 建父目录, 可写层再写入一份 `{}\n` 空文档; 校验失败只记录警告并跳过该层, 不影响引擎启动.
- 层配置的 `path` 与 `default.json` 中其他文件资源一致, 使用 `proto://<source-path>` 约定并由 `internal/path` 统一解析, 引擎不再自行处理 `~` 展开.
- `data/conf/schema/setting.schema.json` 完整描述 `setting.default.json` 的 `appearence` / `general` / `account` / `advanced`, 含 `mode` 枚举、尺寸 `pattern` 与调色板结构.
- 引擎内部以点路径寻址叶子: `loadLayer` 读盘后用 `flatten` 展开成 `appearence.mode` 之类的 key, `writeLayer` 前用 `nest` 还原成嵌套文档. 内存侧因此能按叶子读写, 文件侧保持与 schema 一致的嵌套结构.
- 层分三类: `default`(随包的 `ro` 层)、`user`(唯一负责持久化的 `rw` 层)、`runtime`(无 path 的内存层). `runtime` 恒在最顶层, 由 `Init` 自动追加, 不从配置注册.
- `Set` 写入 `runtime` 层时置脏标记; `Flush` 只把 `runtime` 的 key 合并进 `user` 层并落盘, 随后清空 `runtime`. 同 key 时 `runtime` 索引更高, `Resolve` 从后往前取, 天然高层覆盖; 不同 key 各自留在各自层, 用户文件被监视器重载时只替换 `user` 层的贡献.
- 持久化有三条触发: `SettingService.Save` 在 `Set` 后异步 `Flush`(不阻塞 UI); `watchFlush` 以 30s 间隔检查脏标记作为安全网; `Stop` 在停机前 `Flush` 一次.
- `Keys()` 列出所有有贡献的 key(升序), `Flush()` 是新增的契约方法, 二者都被 `SettingService` 直接使用.
- 层列表由 `LayersFrom` 从 `default.runtime.setting.layers` 解析, 顺序即优先级(索引递增), `mode` 的 `ro` / `wo` / `rw` 映射为 `setting.Mode`, 未知值回退 `ro`; 不再硬编码层名与路径.
- `Deps()` 只声明 `store.fs`. 早期版本额外依赖 `log.direct`, 而 `default.json` 的 `runtime.log.direct` 为空数组时不创建该引擎, 依赖缺失会让引擎管线在中途 `return`, 后续引擎(含 store)全部跳过 Init. 诊断一律走 `slog.Default()`.
- `SettingService.Payload` 返回 `schema` / `values` / `defaults` 三部分: `schema` 经 `store.fs` 读 `embed://` 资源(路径取自 `default.runtime.setting.schema`), `values` 是引擎聚合结果(已嵌套还原), `defaults` 是 `default` 层的快照, 供渲染器在用户值非法时回退.
- `SettingService.Save(key, value)` 把值写入 `runtime` 层, 只接受点路径, 由引擎负责合并与落盘; 前端因此可以按叶子提交, 不必回传整组配置.
- 文件监视发现变更后, `loadLayer` 在锁外逐 key 调用 `notify`, 以 `setting::changed` 事件广播, 使前端不必重启即可感知用户层被外部编辑.
