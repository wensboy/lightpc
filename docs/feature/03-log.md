# 03 - log

> 核心目的: 日志抽象.
> 
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

## 逻辑

## 注意点

## 细节

- `slog.Handler` 的输出 writer 在构造时绑定, 无法在 `Handle` 中重定向; 因此 `FileHandler` 让 serializer 直接写 `out`, `WithAttrs`/`WithGroup` 通过委托派生新的 serializer 保持 attrs 与 group 语义.
- `NewFileHandler` 打开 `FileOption.Path` 失败时回退到标准输出, `Mode` 为 0 时使用 `0644`; 并发写入由内部互斥锁串行化.
- 级别过滤只由外层 `FileHandler.Enabled` 负责, 内层序列化器创建时固定 `slog.LevelDebug`, 保证 `Handle` 不会被内层再次拦截.