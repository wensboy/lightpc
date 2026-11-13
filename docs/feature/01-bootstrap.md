# 01 - bootstrap

> 核心目的: 程序启动引导结构.
> 
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

- `backend/domain/context`: `xxxContext` 总是传输型上下文用于传输数据, 可以被持有并持续调用获取上下文信息. 上下文可组合或嵌入.
- `backend/domain/runtime`: `xxxRuntime` 总是标记一个有限状态机的所有状态信息, 用于状态驱动变化. 通常为 struct 结构.
- `backend/domain/bus`: `xxxBus` 总线结构, 按照语义决定实现和功能, 总线是实例级别隔离的.

## 逻辑

## 注意点

## 细节

- `SetupApp` 从 `default.application.{name,description}` 与 `default.application.window.{width,height}` 取值; wails v3 的 `application.Options` 没有窗口字段, 标准窗口通过 `app.Window.NewWithOptions` 创建.
- `cmd` 入口通过 `c.Int("mode")` 读取 `mode` flag 并转 `uint8` 写入 `BootContext.Mode`.
- `SetupErrors` 读取 `errors.<group>.definition[]` 中每项的 `code` 与 `kind` 构建 `map[code]kind`; 配置缺失、分组不是对象或缺字段时跳过该条目. `errors.json` 与 `default.json` 一并由 `SetupConfig` 加载.