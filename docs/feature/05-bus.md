# 05 - bus

> 核心目的: 总线抽象.
> 
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

## 逻辑

## 注意点

1. 总线通常作为其他包的依赖, 通常不依赖其他包. 所有内容内聚在: `domain/bus`, `internal/ebus`.
2. 所有事件 key 定义到 `domain/bus/event.go` 中, 格式为: `EventXXX`, 通常为 const.
3. 所有 event key 使用 `::` 分隔语义与词.
4. 事件的 payload 在 emit 传递时通过 `map[string]any` 传递, 在 on 监听时通过 mapstructure 解码为实际的结构 T, 该结构直接定义在监听的 on 回调函数中.

## 细节