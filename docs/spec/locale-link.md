---
name: '多语言链路'
description: '新增语言与让静态文本跟随设置切换的固定处理链路'
---

> ai agent 能够将补充的示例追加到 Example 段下, 无明确编辑其他段时不允许编辑.

## Step

1. 在 `src/frontend/src/locales/i18n/<locale>/` 下按模块建立字典, 文件名与字段结构必须与 `en-US` 完全一致.
2. 在 `src/frontend/src/locales/index.ts` 的 `dictionaries` 中注册该 locale, 并在 `localeNames` 中登记设置文件使用的显示名.
3. 在 `src/data/conf/setting.default.json` 的 `general.availableLanguages` 增加 `{ id, name }`, `id` 为 locale 目录名, `name` 为界面显示名; `general.language` 存的是 `name`.
4. 页面通过 `useText()` 取字典, 组件因此订阅切换; 非组件位置用 `currentText()`, 且必须在调用点读取, 不得提到模块顶层.

## Hint

- `src/frontend/src/locales/i18n/en-US` 是参考实现: 其他 locale 缺字段会在类型层面报错, 不会静默回退.
- 字典只放静态文本, 不放业务数据与逻辑; 需要拼接的句子写成函数.
- `general.language` 属于 ui 立即刷新型, 由 `features/settings/effects.ts` 的 `applyLocale` 处理, 不需要后端事件.
- 文案在模块顶层求值会冻结首次加载的语言, zod schema 尤其容易踩到: 把 schema 写成函数, 每次调用重新构建.

## Example

