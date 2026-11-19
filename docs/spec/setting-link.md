---
name: '设置链路'
description: '加入设置字段并联调通前后端的固定处理链路'
---

> ai agent 能够将补充的示例追加到 Example 段下, 无明确编辑其他段时不允许编辑.

## Step

1. 在 `src/data/conf/setting.default.json` 下配置指定字段和默认值并同步对应的 schema 到 `src/data/conf/schema/setting.schema.json`.
2. 判断是 ui 立即刷新型还是懒刷新型.
- ui 立即刷新型: 前端注册 ui 处理
- ui 懒刷新型: engine runtime, service runtime emit 设置更新事件, 指定 service init 监听该事件指定设置 key 更新状态. *app runtime window app emit 触发前端重渲染(如果有必要)*.

## Hint

- `src/data/conf/schema/extension.schema.json` 是所有在 `src/data/conf/schema/setting.schema.json` 中拓展字段的定义文件.
- `src/data/conf/schema/setting.schema.json` 决定 ui 展示, `src/data/conf/setting.default.json` 只是提供值显示的最底层, 用户目录下的 setting.json 能够覆盖, 并且原则上能够完全覆盖, 这意味着即使 ui 不显示的(derived)的字段也能够由用户修改覆盖.
- 从是否立即影响ui变化的角度, 设置项只分为: ui立即刷新型(前端刷新, 写回后端) 和 ui懒刷新型(写回后端, 前端自动看到最新值按照预期渲染或只影响了后端无任何ui变化). 后者通常通过全局 store 配合后端事件进行 ui 透明化值更新.

## Example

- 新增内置字体时：把文件放进 public/fonts/、在 index.css 加一条 @font-face、把族名加进 setting.default.json 的 builtin，Payload 会自动校验三者一致。