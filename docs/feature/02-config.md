# 02 - config

> 核心目的: 配置处理细节.
> 
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

- 

## 逻辑

## 注意点

- config 核心聚焦: comptime(编译时配置, 程序启动时加载一次, 静态配置, 不允许修改); runtime(运行时配置, 动态配置, 允许修改).
- config 所有 key 按照 . 格式访问, 例如所有key都是: `key1.key2,key3` 格式.
- config 应当将所有 key 在适当时机展开平铺, 避免复杂 key 访问开销. 在 key 层级不深的情况下通过 key 路径递归即可.
- config 默认情况下不直接看到 value 结构, 通过结构强映射在查找到 key=value 后解析, 强映射定义在 `src/backend/domain/config/map.go` 下, tag 强对应.
-  

## 细节

- embed.FS 的路径不带前导 `/`, 而 `EmbedContext` 的 key 形如 `/data`, 因此 `FileSource.Parse` 与 `EnvSource.Export` 读取前需要去掉前导 `/`.
- `FileSource` 的顶层 key 使用去扩展名的配置文件名(`default` 而非 `default.json`), 否则 `.` 分隔的 key 路径无法按段逐层迭代.
- `EventBus.Emit` 必须复制处理器列表后再释放读锁调用, 否则处理器内部调用 `On` 会因 `RWMutex` 不可重入而死锁.
- `FileSource` 当前只解析 `.json`(标准库), `.yaml` 依赖三方库, 未纳入实现.
- `FileSource.Parse` 按传入的文件路径解析(`Paths` 为配置文件路径), 与 `entry.go` 的调用方式一致.
- urfave/cli v3 的 `FlagBase.Deprecated` 是 string 而 domain `Flag.Deprecated` 是 bool, 转换时需构造废弃提示文本; `FlagBase.Get()` 在未 `Apply` 时返回声明的默认值, 可直接用于 `FlagSource.Lookup`.
- `Flag.Value.Type` 与 Go 类型名保持一致(`string`/`bool`/`int`/`int64`/`uint`/`float64`), 未设置或未知类型按 `string` 处理; JSON number 统一解析为 `float64`, 必须依靠 `type` 才能区分 int 与 float.
- flag 默认值通过 `cast.As[T]` 从 `Default` 做直接类型转换, 转换失败时取零值.