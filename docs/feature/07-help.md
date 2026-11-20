# 07 - help

> 核心目的: 文档浏览链路.
>
> ai agent 只能同步追加到末尾的 `## 细节` 段下.

## 对象与接口

| 类型 | 位置 | 说明 |
|---|---|---|
| `HelpClient` | `domain/service` | 服务契约: `Payload()` 列出文档集, `Read(version, path)` 取正文 |
| `DocsVersion` | `domain/service` | 一个文档集: `Version` `Description` `Author` `Date` `Deprecated` `Tags` `Toc` |
| `DocsGroup` | `domain/service` | 一级目录: `Title` + `Items` |
| `DocsItem` | `domain/service` | 二级条目: `Title` + 相对 `Path` |
| `DocsPayload` | `domain/service` | 响应体: `Versions` |
| `Service` | `service/help` | 契约实现, 读 `runtime.docs` 声明的目录 |
| `render.ts` | `features/help` | markdown 解析、净化、分段与排版规则 |
| `HelpMarkdown` | `features/help/components` | 把分段映射成组件或容器 |

```go
type HelpClient interface {
	Payload() (HelpPayload, error)
	Read(version, path string) (string, error)
}
```

- 注册 key 为 `service::help`; 依赖只有 `store.fs`.
- 文档目录由 `default.runtime.docs` 声明, 当前是 `embed://data/docs/beta`.

## 逻辑

```
Init
  ├─ 注册 Errors
  ├─ 取 store.fs 引擎, 失败即返回
  ├─ 读 default.runtime.docs
  └─ 记录每个目录到 roots

Payload
  └─ 逐个 root: 读 <root>/meta.json -> decodeMeta -> 追加 DocsVersion

Read(version, path)
  ├─ rootOf(version): 按 meta.json 的 version 反查目录
  ├─ 拒绝含 ".." 或以 "/" 开头的 path
  └─ 读 <root>/<path> 返回原文

前端 HelpMarkdown
  ├─ renderMarkdown: marked 解析 (gfm)
  ├─ sanitizeHtml: DOMPurify 白名单 + 外链 target/rel
  ├─ splitBlocks: 按 <pre> 切分 html 段与代码段
  └─ 渲染: 代码段 -> MarkdownCodeBlock, 其余 -> dangerouslySetInnerHTML
```

- 目录顺序由 meta.json 的书写顺序决定: `toc` 是对象, 用 `json.Decoder` 逐 token 解析才不丢序.
- 版本下拉列出 `Payload` 的 `Versions`, 切换版本时清空正文并回到版本摘要.
- 未选中文档时, header 显示版本, body 显示描述 / 作者 / 日期 / 标签.

## 注意点

- `joinResource` 必须走 `internal/path.Join`: `path.Join` 会 `Clean`, 把 `embed://` 折成 `embed:/`, 解析随即失败且日志无痕.
- 界面空白先看 `help service ready roots=[...]`: 无此行说明配置没读到, 有此行但列表空说明 meta.json 读不出.
- 文档路径来自 toc, 但仍要挡 `..` 与绝对路径, 否则能读穿目录.
- 弹窗宽度约束是一条链: `DialogContent` -> `SidebarProvider` -> `main` -> 滚动区 -> 内容根, 任一层缺 `min-w-0` 都会被宽内容撑开.
- 代码块用官方 `ui/code-block`, markdown 侧只在 `components/markdown/CodeBlock.tsx` 组合, 不改 ui.
- `shiki` 语法按需加载, 首次渲染先出纯文本再着色, 这是官方 hook 的既定行为.

## 细节

- `meta.json` 的结构由 `data/conf/schema/docs.meta.schema.json` 描述, `version` / `author` / `date` / `description` / `toc` 为必填, `toc` 只允许两层.
- `toc` 的键是一级目录标题, 值是相对 `meta.json` 的 markdown 路径; 前端把一级标题作为分组, 二级作为可搜索条目.
- `Read` 的 `version` 参数取自 meta.json 的 `version` 字段, 不是目录名: 目录改名后前端无需调整.
- `Payload` 遇到读不出或解析失败的 meta.json 时跳过该集合并记 `Warn`, 不让一个坏目录拖垮整个列表.
- markdown 渲染分段而不是整块注入: 代码段需要 React 组件承载标题栏与动作, 其余内容仍是净化后的 html.
- `splitBlocks` 用 `<template>` 解析 html, 模板内容是惰性文档片段, 不会触发图片加载.
- 代码块标题取自围栏的语言标注, 空标注显示 `text`; 头部动作只有换行切换与复制, 与官方示例对齐.
- `useHighlightedCode` 的语言表按需 `import`, 新增语言只需在 `lib/highlighter.tsx` 的 `languages` 里加一行, 别名加在 `aliases`.
- 排版规则集中在 `render.ts` 的 `prose`: 表格 `w-full` + 单元格 `break-words`, 长代码行由 `CodeBlockContent` 自身滚动.
