# Frontend

> Architecture · Frontend

React 19 with Vite, TanStack Router and shadcn/ui on Base UI. The code is organised by feature, not by file
kind: a feature owns its components, state and types.

## Structure

```
src/
├── components/
│   ├── ui/            ← shadcn primitives, lowercase kebab files
│   └── *.tsx          ← shell pieces (nav, switcher)
├── features/<name>/   ← one folder per feature, PascalCase components
├── hooks/             ← shared hooks, kebab file names
├── lib/               ← small helpers with no React state
├── locales/           ← dictionaries, one file per feature
├── services/          ← typed wrappers over the wails bindings
└── stores/            ← cross feature state (dialogs)
```

## Rules

| Rule | Reason |
| --- | --- |
| text comes from `useText()` | one dictionary per locale, no literals in components |
| payloads are validated with Zod at the service boundary | the backend is a process boundary, not a type checker |
| `components/ui` is never edited for feature needs | regenerating it would drop the change |
| a flex child that holds wide content needs `min-w-0` | otherwise it refuses to shrink and widens the frame |
| feature components use PascalCase file names | `ui` keeps kebab for shadcn parity |

## State

Local state stays in the component. Anything two features share goes to `stores/` (zustand). Server data is
fetched through `services/` and validated there, so a component never parses a raw payload.
