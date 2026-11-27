# 前端规则

> 核心目的：定义前端代码与设计抽象必须遵守的规则。

## 技术栈

- 核心技术栈
  - 核心框架：React 19.3+
  - 语言：TypeScript 5.x（严格模式）
  - 构建工具：Vite 8.x
  - 样式：Tailwind CSS v4.x
  - UI 组件：shadcn/ui
  - 状态管理：Zustand v5 + React Context
  - 表单：React Hook Form v7 + Zod v4
  - 代码风格：Biome

- 工具链与 CI
  - 包管理器：pnpm 10.x+（内容寻址存储与严格依赖隔离）
  - CI 安装模式：`--frozen-lockfile`（在 CI 中自动启用）
  - 依赖治理：`pnpm-workspace.yaml`（overrides、peerDependencyRules）
  - 生命周期安全：`pnpm approve-builds`（依赖的构建脚本默认被阻止）

- Mock 系统
  - Mock 策略：依赖注入（DI），在服务层进行拦截
  - Mock 工具：Vitest（单元测试 mock）/ MSW（偶尔的 HTTP 请求）/ 针对 Wails 桥接 mock 的自定义封装
  - 数据校验（Mock）：Zod（复用真实业务模型的 schema）
  - 环境控制：Vite 环境变量（`import.meta.env.MODE` / `.env.development`）

- 文本与国际化
  - 方案（文本）：轻量、强类型的字典，为部分国际化扩展预留空间
  - 实现（文本）：TypeScript 常量枚举/字典，配合自定义 hook（`useText`）
  - 数据校验（文本）：Zod（校验动态插值参数的类型）
  - 插值引擎：轻量插值库（如 `interpolate`，或直接用模板字符串替换）

## 必须

1. 在 `tsconfig.json` 中开启 `strict: true`，并严格禁止 `any` 类型。
2. 遵循按功能划分的目录结构：核心业务代码必须位于 `src/features/` 下，按功能模块拆分（例如 `src/features/auth/`）。绝不把业务逻辑堆积到全局共享目录中。
3. 跨模块共享的客户端状态使用 Zustand，并按模块拆分 store（例如 `useAuthStore.ts`）；绝不创建单一庞大全局 store。
4. 所有跨越 Wails 桥接（Go -> TS）的复杂数据结构，都必须在运行时用 Zod 校验，以保证类型绝对安全。
5. 使用 shadcn/ui 提供的 CSS 变量（如 `bg-background`、`text-primary`）集中控制主题，使深色与浅色模式无缝切换。
6. 超过 50ms 的复杂计算或 I/O 交给 Go 后端处理；前端只负责展示结果。
7. 在根目录 `package.json` 中通过 `"packageManager": "pnpm@x.y.z"` 固定确切的 pnpm 版本。CI 严格从该字段锁定版本，避免 `pnpm-lock.yaml` 不兼容或版本变更导致安装行为漂移。
8. 将 `pnpm-lock.yaml` 提交到版本控制，并在 CI 流程中强制使用 `pnpm install --frozen-lockfile`（pnpm 11+ 可用 `pnpm ci`）。在 CI 中，锁文件与 `package.json` 不一致必须立即导致安装失败；CI 绝不能重写锁文件。
9. 在 Wails 开发模式（`wails dev`）下，通过 `WAILS_VITE_PORT` 之类的环境变量配置 Vite 开发服务器端口，使 pnpm 脚本与 Wails 的端口探测正确协作，避免硬编码端口导致连接失败。
10. 尊重依赖安装信任边界：pnpm 默认**不**执行依赖的 `postinstall` 等生命周期脚本。必须运行构建脚本的包（例如编译原生二进制的包）需要通过 `pnpm approve-builds` 显式审查和批准；绝不能全局启用 `dangerouslyAllowAllBuilds`。
11. 所有 mock 代码、fixture 数据与 mock 服务配置严格放在 `src/mocks/` 下。
12. 在业务代码与原生 Wails 绑定（`wailsjs/go/...`）之间加一层服务层（如 `src/services/user.ts`）。业务组件只能调用服务层；严禁在组件中直接调用生成的 Go 绑定。
13. 使用环境变量（如 `VITE_ENABLE_MOCK=true`）作为全局 mock 开关。只有在开发或测试环境中显式开启该标志时，服务层才从 `src/mocks/` 加载 mock 实现。
14. 确保 mock 数据的结构（包括字段类型与嵌套）能通过真实的 Zod schema。增删 mock 字段必须同步更新对应的 schema。
15. 移除某个功能时，在同一次变更中删除它在 `src/mocks/` 下的 mock 文件，并在 PR 描述中说明。
16. 系统中所有静态文本集中管理——包括但不限于页面标题、按钮标签、表单标签、toast 消息、空状态提示与确认弹窗，统一放在 `src/locales/` 或 `src/constants/texts/` 下。
17. 文本字典按领域或模块拆分（如 `auth.ts`、`dashboard.ts`、`common.ts`），并通过单一入口文件（如 `index.ts`）导出；绝不在单个文件中维护包含数百个 key 的巨型字典。
18. 使用 `module.component.meaning` 的命名空间约定让字典 key 具有高度可读性（例如 `auth.login.submitBtn`）。
19. 在 React 组件中，通过共享 hook（如 `const t = useText()`）或强类型辅助函数注入和读取文本，使 TypeScript 能够提供补全与编译期检查。
20. 对于带动态变量的文本（如 `Welcome back, {name}`），使用带参数插值的函数式字典，并用 Zod 或 TS 泛型严格约束参数类型。
21. 集中处理 Go 后端返回的 `Error Code` 映射。后端只返回码，前端通过字典将其翻译为可读的本地化消息。

## 禁止

1. 把 `utils/` 或 `helpers/` 当作杂物抽屉，或者把带有业务语义的代码放进 `shared/`。
2. 在 `components/` 中混放设计系统组件与业务组件。
3. 依赖单一全局 store，或手写 reducer 缓存服务端响应。
4. 在前端代码中硬编码敏感 API key、token 或密码。由于 WebView 易于逆向或调试，密钥必须留在 Go 后端。
5. 直接用 `fetch` 或 `axios` 调用外部 HTTP 接口（纯前端公共服务等例外情况除外）。所有业务数据必须经过 Wails 绑定，由 Go 后端转发或代理。
6. 进行频繁、高频的 IPC 调用（例如在 `mousemove` 或 `scroll` 回调中直接调用 Go 方法）。应在前端做防抖或节流。
7. 向 WebView 引入庞大且不受控的外部依赖（如完整的 moment.js 或整包 lodash）。严格控制包体积与内存占用。
8. 使用除 Tailwind CSS 之外的重量级 CSS-in-JS 方案（如 styled-components），它们会给 WebView 渲染引擎带来运行时开销。
9. 绕过 shadcn/ui 的 `cn()` 辅助函数而直接拼接 Tailwind 类名，这会导致样式覆盖与冲突难以排查。
10. 在 CI 脚本或本地开发中混用 `npm`、`yarn` 或 `bun`。`packageManager` 字段配合 `.npmrc` 的 `engine-strict` 必须确保团队与 CI 只使用 pnpm，避免不同包管理器之间的锁文件格式冲突。
11. 手工编辑 `pnpm-lock.yaml`。应通过 `pnpm install` 重新生成文件来解决锁冲突；通过 `package.json` 或 `catalog:` 条目驱动锁更新，而不是篡改锁文件本身。
12. 滥用根级 `overrides` 或 `peerDependencyRules` 来消除所有 peer 警告。对于 React 19 的 peer 问题，先确认上游包是否已适配，只在必要时用 `:peer` 选择器做精确覆盖。绝不把 peer 依赖强制到不兼容的大版本。
13. 在业务组件（如 `src/features/...`）中硬编码 `if (isMock)` 分支。Mock 切换必须完全隐藏在服务层内部。
14. 将 mock 依赖（如 MSW、faker.js）或 mock 配置打进生产包。`wails build` 产物绝不能包含任何 mock 代码。
15. 即使在测试环境中，也不得在 mock 数据里硬编码真实敏感用户信息或密钥。应使用 faker 之类的生成器或脱敏后的静态 fixture。
16. 使用有严重副作用的 mock，如写入真实数据库或触发付费外部 API。
17. 在 `.tsx` 组件或任何 UI 逻辑中硬编码面向用户的字符串（`<button>Submit</button>` 和 `placeholder="Enter your username"` 都算违规）。
18. 在组件中内联复杂的条件文本拼接（如 `isVip ? 'VIP' : 'Regular' + ' member'`）；应移到字典的计算属性或专用格式化函数中。
19. 依赖字符串字面量作为业务逻辑条件（如 `if (text === 'Success')`）；业务逻辑必须使用专用枚举或状态码，文本只用于展示。
20. 提交生成物。`frontend/bindings/` 由 `task generate:bindings` 写出并保持忽略；后端暴露面变化后重跑该任务，绝不手敲 `wails3 generate bindings` 并改动参数——漏掉 `-ts` 会重写全部文件，把一次构建变成全仓库重命名。

## 应当

1. 充分利用 React 19 的 React Compiler，减少手工 `useMemo` 与 `useCallback` 的使用，保持代码简洁。
2. 对数十项以上的长列表使用虚拟列表（如 `@tanstack/react-virtual`），避免 DOM 节点过多导致 WebView 滚动卡顿。
3. 将纯展示组件（Dumb Components）与有状态的容器组件（Smart Components）分离，使 UI 组件保持高度可复用。
4. 将 Go 后端推送的事件（Wails Events）封装到自定义 hook 中，让 React 组件共享统一的订阅与生命周期清理位置。
5. 对前端路由做代码分割（使用 TanStack Router 或 React Router），按需加载每个页面的 JavaScript。
6. 用 Zod 一次性定义复杂表单校验 schema，并通过 `@hookform/resolvers/zod` 与 React Hook Form 结合，保持单一数据源。
7. 使用 pnpm 的过滤执行（`--filter`）。避免每次都对整个仓库运行 `pnpm build` 或 `pnpm lint`；根据变更范围（例如 Git 变更文件）通过 `pnpm --filter <package>...` 触发构建与检查，缩短 CI 反馈时间。
8. 在 CI 中同时缓存 pnpm store 与元数据缓存（路径来自 `pnpm cache path`）。缓存 store 可复用依赖内容；缓存元数据可跳过锁文件策略检查。两者结合可最大化 CI 安装速度。
9. monorepo 内部依赖使用 `workspace:*` 协议，让 pnpm 的工作区机制完全控制内部版本解析，避免本地路径或版本不匹配导致构建失败。
10. 定期从 `pnpm-workspace.yaml` 或 `package.json` 的 `pnpm` 字段运行 `pnpm audit --fix`，生成针对性的 `overrides` 修复漏洞。修复后确认这些 `overrides` 是否仍有必要，并及时清理失效条目。
11. 所有 mock 注册与拦截集中在 `src/mocks/` 的入口文件（如 `index.ts`）中。利用 Vite 的动态导入（`import.meta.glob`）按需加载 mock 模块，保持开发启动速度。
12. 在服务层 mock 中加入人为延迟，模拟真实 IPC 与后端处理耗时（如 `await sleep(300)`），使缺失的加载状态或竞态条件在开发阶段暴露。
13. 为 Wails 的事件系统（后端到 WebView 的推送）提供专用的 `MockEventService`，让前端无需 Go 后端即可触发事件流，从而端到端测试 UI 响应逻辑。
14. 对于不跨越 Wails 桥接的外部 HTTP 请求，用 MSW（Mock Service Worker）在网络层拦截，保持请求路径一致。
15. 按功能拆分 mock 文件（如 `auth.mock.ts`、`dashboard.mock.ts`），与 `src/features/` 的布局保持一致。
16. 使用扁平与嵌套结合的字典结构，例如全局动作用 `Common.Confirm`，特定页面用 `Pages.Auth.Login.Title`。
17. 为未来的国际化扩展预留接口。即使目前只支持中文，也把顶层设计为 `{ 'zh-CN': {...}, 'en-US': {...} }`，以便日后接入 `react-i18next`。
18. 为 mock 提供专门的静态文本。`src/mocks/` 下的 fixture 数据（姓名、邮箱、地址等）也应来自统一的 mock 字典，保持前后端契约一致。
19. 对服务条款、隐私政策等长文本做懒加载，使主包保持较小体积。

## 不应

1. 过度依赖前端持久化（如在 `localStorage` 中存储大量业务数据）；优先使用 Go 后端的 SQLite 或配置文件。
2. 在多个组件中重复同一个 Go 端点的绑定调用逻辑；应封装到统一的服务层（如 `src/services/`）中。
3. 在前端实现时间格式化、复杂货币计算或加解密；这些完全交给 Go 后端。
4. 在 React 组件中编写复杂的原生 DOM 操作，或直接操作 `window` 对象；保持 React 的数据驱动思维。
5. 对核心布局使用硬编码像素值的绝对定位（`position: absolute`）；优先使用 Flexbox/Grid 与响应式工具类，使布局适配不同窗口尺寸。
6. 依赖幻影依赖（phantom dependencies）。由于 pnpm 默认使用严格的 `node_modules` 布局，导入未声明在 `package.json` 中的间接依赖会在干净的 CI 环境中失败。应在开发阶段通过 `import/no-extraneous-dependencies` 等 ESLint 规则捕获这类问题。
7. 在 Wails 生产构建（`wails build`）中保留 Vite 开发服务器配置或 HMR 依赖。生产构建必须将 Vite 产物输出到 Wails 期望的目录（如 `dist/`），并通过 `//go:embed` 正确嵌入 Go 二进制。
8. 用 `^` 或 `~` 范围声明强锁定的工具依赖（`react`、`react-dom`、`typescript`、`vite`）。核心框架与构建工具应固定确切版本（在 `pnpm-workspace.yaml` 的 `catalog:` 中集中管理），避免小版本升级在协作时产生无关的锁文件 diff 噪音。
9. 滥用 mock 来掩盖真实后端端点缺失或设计不良的事实。Mock 必须严格遵循约定的前后端 API 契约。
10. 在组件或服务中为同一端点编写多个相互冲突的 mock。复杂场景应改用参数化的 mock 工厂。
11. 仅仅为了生成 mock 数据而引入重量级依赖，例如从 `@faker-js/faker` 导入全部语言包。应按需导入或使用轻量生成器。
12. 依赖在浏览器控制台手动修改全局状态来切换 mock 数据。应通过受控的代码配置或仅开发环境可见的 UI 面板切换。
13. 为了简单的文本替换引入重量级的完整 i18n 运行时。在纯客户端的 Wails 应用中，优先使用轻量的纯 TypeScript 字典。
14. 在一个文件中混放不同领域的文本 key，这会导致无意义的 Git 冲突（例如修改 Dashboard 文案却在 Auth 模块产生 diff）。
15. 在文本插值中执行未转义的 HTML。如果必须渲染富文本，应显式使用 `dangerouslySetInnerHTML` 并用 DOMPurify 净化；绝不把用户输入直接插入插值模板。
16. 让空状态挂在共用父级。一个导航入口有真实面板后，空状态归属该面板；未实现的入口共用一份 `pending` 文案，已实现的面板各自描述自己的状态。
