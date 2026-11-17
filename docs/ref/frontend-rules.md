# Frontend Rules

> Core purpose: Define the rules that frontend code and design abstractions must follow.

## Tech

- Core Stack
  - Core framework: React 19.3+
  - Language: TypeScript 5.x (Strict Mode)
  - Build tool: Vite 8.x
  - Styling: Tailwind CSS v4.x
  - UI components: shadcn/ui
  - State management: Zustand v5 + React Context
  - Forms: React Hook Form v7 + Zod v4
  - Code style: Biome

- Toolchain & CI
  - Package manager: pnpm 10.x+ (content-addressable storage and strict dependency isolation)
  - CI install mode: `--frozen-lockfile` (detected automatically in CI)
  - Dependency governance: `pnpm-workspace.yaml` (overrides, peerDependencyRules)
  - Lifecycle safety: `pnpm approve-builds` (dependency build scripts are blocked by default)

- Mock System
  - Mock strategy: dependency injection (DI) with interception at the service layer
  - Mock tooling: Vitest (unit-test mocks) / MSW (for occasional HTTP requests) / a custom wrapper for Wails bridge mocks
  - Data validation (Mock): Zod (reuse the schemas of real business models)
  - Environment control: Vite environment variables (`import.meta.env.MODE` / `.env.development`)

- Text & i18n
  - Approach (Text): a lightweight, strongly typed dictionary with room for partial i18n extension
  - Implementation (Text): TypeScript constant enums/dictionaries plus a custom hook (`useText`)
  - Data validation (Text): Zod (validate the types of dynamic interpolation arguments)
  - Interpolation engine: a lightweight interpolation library (such as `interpolate` or plain template-string replacement)

## Must

1. Enable `strict: true` in `tsconfig.json` and strictly forbid the `any` type.
2. Follow feature-based directory partitioning: core business code must live under `src/features/`, split by feature module (for example `src/features/auth/`). Never pile business logic into globally shared directories.
3. Use Zustand for client state shared across modules, and split stores by module (for example `useAuthStore.ts`); never create one enormous global store.
4. Validate every complex data structure crossing the Wails bridge (Go -> TS) with Zod at runtime so that types are absolutely safe.
5. Use the CSS variables provided by shadcn/ui (such as `bg-background` and `text-primary`) to control theming centrally, so dark and light modes switch seamlessly.
6. Offload complex computations or I/O that take longer than 50ms to the Go backend; the frontend is responsible only for presenting the result.
7. Pin an exact pnpm version in the root `package.json` via `"packageManager": "pnpm@x.y.z"`. CI locks the version strictly from this field, preventing `pnpm-lock.yaml` incompatibilities or install behaviour drift caused by version changes.
8. Commit `pnpm-lock.yaml` to version control and enforce `pnpm install --frozen-lockfile` (or `pnpm ci` on pnpm 11+) in the CI pipeline. In CI, an inconsistency between the lock file and `package.json` must fail the installation immediately; CI must never rewrite the lock file.
9. In Wails dev mode (`wails dev`), configure the Vite dev server port through an environment variable such as `WAILS_VITE_PORT`, so pnpm scripts and Wails port detection cooperate correctly and a hard-coded port cannot break the connection.
10. Respect the dependency installation trust boundary: by default pnpm does **not** run lifecycle scripts such as `postinstall` from dependencies. Packages that must run build scripts (for example those compiling native binaries) require explicit review and approval through `pnpm approve-builds`; never enable `dangerouslyAllowAllBuilds` globally.
11. Keep all mock code, fixture data and mock service setup strictly under `src/mocks/`.
12. Add a service layer (such as `src/services/user.ts`) between business code and the native Wails bindings (`wailsjs/go/...`). Business components may call only the service layer; calling generated Go bindings directly from a component is strictly forbidden.
13. Use an environment variable (such as `VITE_ENABLE_MOCK=true`) as the global mock switch. The service layer loads mock implementations from `src/mocks/` only when the flag is explicitly enabled in a development or test environment.
14. Ensure the shape of mock data, including field types and nesting, passes the real Zod schema. Adding or removing mock fields must be accompanied by the corresponding schema update.
15. When removing a feature, delete its mock files under `src/mocks/` in the same change and state this in the PR description.
16. Centralise all static text in the system — including but not limited to page titles, button labels, form labels, toast messages, empty-state hints and confirmation dialogs — under `src/locales/` or `src/constants/texts/`.
17. Split the text dictionary by domain or module (such as `auth.ts`, `dashboard.ts`, `common.ts`) and export it through a single entry file (such as `index.ts`); never maintain one giant dictionary with hundreds of keys in a single file.
18. Make dictionary keys highly readable with the `module.component.meaning` namespace convention (for example `auth.login.submitBtn`).
19. In React components, inject and read text through a shared hook (such as `const t = useText()`) or a strongly typed helper so TypeScript can provide completion and compile-time checks.
20. For text with dynamic variables (such as `Welcome back, {name}`), use a functional dictionary with parameter interpolation and constrain the argument types strictly with Zod or TS generics.
21. Handle the mapping of `Error Code` values returned by the Go backend in one place. The backend returns only a code; the front end translates it into a readable localised message through the dictionary.

## Must Not

1. Treat `utils/` or `helpers/` as a junk drawer, or put code with business semantics into `shared/`.
2. Mix design-system components with business components in `components/`.
3. Rely on a single global store, or hand-write reducers to cache server responses.
4. Hard-code sensitive API keys, tokens or passwords in frontend code. Because the WebView is easy to reverse engineer or debug, secrets must stay in the Go backend.
5. Call external HTTP endpoints directly with `fetch` or `axios` (unless it is an exceptional purely frontend public service). All business data must go through Wails bindings and be forwarded or proxied by the Go backend.
6. Perform frequent, high-frequency IPC calls (for example invoking Go methods directly inside `mousemove` or `scroll` handlers). Apply debouncing or throttling on the frontend side.
7. Pull large and uncontrolled external dependencies into the WebView (such as the full moment.js or a full lodash import). Keep bundle size and memory usage strictly under control.
8. Use heavy CSS-in-JS solutions other than Tailwind CSS (such as styled-components), which add runtime overhead to the WebView rendering engine.
9. Bypass the `cn()` helper from shadcn/ui and concatenate Tailwind class names directly, which leads to confusing style overrides and conflicts.
10. Mix `npm`, `yarn` or `bun` into CI scripts or local development. The `packageManager` field together with `.npmrc`'s `engine-strict` must ensure the team and CI use pnpm only, avoiding lock file format conflicts between package managers.
11. Hand-edit `pnpm-lock.yaml`. Resolve lock conflicts by regenerating the file with `pnpm install`; drive lock updates through `package.json` or the `catalog:` entries instead of tampering with the lock file itself.
12. Abuse root-level `overrides` or `peerDependencyRules` to silence all peer warnings. For React 19 peer issues, first confirm whether upstream packages have adapted, then use the `:peer` selector for precise overrides only when necessary. Never force peer dependencies onto incompatible major versions.
13. Hard-code `if (isMock)` branches inside business components (such as `src/features/...`). Mock switching must be fully hidden inside the service layer.
14. Bundle mock dependencies (such as MSW or faker.js) or mock configuration into production. A `wails build` artifact must never contain any mock code.
15. Hard-code real, sensitive user information or secrets in mock data, even in test environments. Use generators such as faker or sanitised static fixtures instead.
16. Use mocks with serious side effects, such as writing to a real database or triggering a paid external API.
17. Hard-code user-facing strings in `.tsx` components or any UI logic (`<button>Submit</button>` and `placeholder="Enter your username"` are both violations).
18. Inline complex conditional text assembly in components (such as `isVip ? 'VIP' : 'Regular' + ' member'`); move it into a computed property of the dictionary or a dedicated formatter.
19. Rely on string literals as conditions for business logic (such as `if (text === 'Success')`); business logic must use a dedicated enum or status code, while text is display only.

## SHOULD

1. Take full advantage of the React Compiler in React 19 to reduce manual `useMemo` and `useCallback` usage and keep code simple.
2. Use a virtual list (such as `@tanstack/react-virtual`) for long lists with dozens of items or more, so an excessive DOM node count does not cause scrolling jank in the WebView.
3. Separate purely presentational components (Dumb Components) from stateful container components (Smart Components) to keep UI components highly reusable.
4. Wrap events pushed by the Go backend (Wails Events) in custom hooks so React components share one place for subscription and lifecycle cleanup.
5. Code-split frontend routes (with TanStack Router or React Router) and load each page's JavaScript on demand.
6. Define complex form validation schemas once with Zod and combine them with React Hook Form through `@hookform/resolvers/zod`, keeping a single source of truth.
7. Use pnpm's filtered execution (`--filter`). Avoid running `pnpm build` or `pnpm lint` across the whole repository every time; trigger builds and checks through `pnpm --filter <package>...` based on the change scope (for example Git changed files) to shorten CI feedback time.
8. Cache both the pnpm store and the metadata cache in CI (paths from `pnpm cache path`). Caching the store reuses dependency payloads; caching metadata skips lock file policy checks. Together they maximise CI install speed.
9. Use the `workspace:*` protocol for internal dependencies in a monorepo. This lets pnpm's workspace mechanism fully control internal version resolution and avoids build failures caused by local paths or version mismatches.
10. Periodically run `pnpm audit --fix` from `pnpm-workspace.yaml` or the `pnpm` field in `package.json` to generate targeted `overrides` that fix vulnerabilities. After a fix, confirm whether the `overrides` are still needed and remove stale entries promptly.
11. Centralise all mock registration and interception in the entry file of `src/mocks/` (such as `index.ts`). Load mock modules on demand with Vite's dynamic import (`import.meta.glob`) so development startup stays fast.
12. Add artificial delay to service-layer mocks. Simulate real IPC and backend processing time (for example `await sleep(300)`) so missing loading states or race conditions surface during development.
13. Provide a dedicated `MockEventService` for Wails' event system (backend-to-WebView push). It lets the frontend trigger event streams without the Go backend, so UI response logic can be tested end to end.
14. For external HTTP requests that do not cross the Wails bridge, intercept them at the network layer with MSW (Mock Service Worker) to keep the request path consistent.
15. Split mock files by feature (such as `auth.mock.ts` and `dashboard.mock.ts`) to mirror the `src/features/` layout.
16. Use a dictionary structure that combines flat and nested entries. For example `Common.Confirm` for global actions and `Pages.Auth.Login.Title` for a specific page.
17. Reserve an interface for future i18n expansion. Even if only Chinese is supported today, design the top level as `{ 'zh-CN': {...}, 'en-US': {...} }` so `react-i18next` can be plugged in later.
18. Provide dedicated static text for mocks. Fixture data under `src/mocks/` (names, emails, addresses, and so on) should also come from a unified mock dictionary to keep the front/end contract consistent.
19. Lazy load long text such as terms of service or a privacy policy so the main bundle stays small.

## SHOULD NOT

1. Rely too heavily on frontend persistence (such as storing large amounts of business data in `localStorage`); prefer the Go backend's SQLite or configuration files.
2. Repeat binding call logic for the same Go endpoint across multiple components; wrap it in a unified service layer (for example `src/services/`).
3. Implement time formatting, complex monetary calculations, or encryption and decryption in the frontend; leave these entirely to the Go backend.
4. Write complex native DOM operations or manipulate the `window` object directly inside React components; keep the data-driven React mindset.
5. Use absolute positioning (`position: absolute`) with hard-coded pixel values for core layouts. Prefer Flexbox/Grid and responsive utility classes so the layout adapts to different window sizes.
6. Rely on phantom dependencies. Because pnpm uses a strict `node_modules` layout by default, code importing a transitive dependency without declaring it in `package.json` fails in a clean CI environment. Catch these issues during development with ESLint rules such as `import/no-extraneous-dependencies`.
7. Keep Vite dev server settings or HMR dependencies in the Wails production build (`wails build`). Production builds must output Vite assets to the directory Wails expects (such as `dist/`) and embed them correctly into the Go binary with `//go:embed`.
8. Declare tightly locked tooling dependencies (`react`, `react-dom`, `typescript`, `vite`) with `^` or `~` ranges. Pin exact versions for core frameworks and build tools (manage them centrally in `pnpm-workspace.yaml`'s `catalog:`) so minor upgrades do not create unrelated lock file diff noise during collaboration.
9. Abuse mocks to hide the fact that a real backend endpoint is missing or poorly designed. Mocks must strictly follow the agreed frontend/backend API contract.
10. Write multiple conflicting mocks for the same endpoint in components or services. For complex scenarios, use a parameterised mock factory instead.
11. Introduce a heavy dependency purely to generate mock data, such as importing all locale packs from `@faker-js/faker`. Import selectively or use a lightweight generator.
12. Rely on manually changing global state in the browser console to switch mock data. Switch through controlled code configuration or a development-only UI panel.
13. Pull in a heavy full i18n runtime just for simple text replacement. In a client-only Wails app, prefer a lightweight plain-TS dictionary.
14. Mix text keys from different domains in one file, which causes meaningless Git conflicts (for example changing a Dashboard copy produces a diff in the Auth module).
15. Execute unescaped HTML inside text interpolation. If rich text must be rendered, use `dangerouslySetInnerHTML` explicitly and sanitise it with DOMPurify; never insert user input directly into an interpolation template.
