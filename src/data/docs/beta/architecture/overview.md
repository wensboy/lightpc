# Overview

> Architecture · Overview

The application is one process with two halves: a Go backend that owns configuration, storage, logging and
services, and a React frontend that owns the interface. They meet at a narrow seam of generated bindings.

## Layers

```
frontend (React)
    │  wails bindings
    ▼
api            ← exposes services to the webview
    │
service        ← use cases, one package per service
    │
domain         ← contracts: engine, service, bus, context
    │
engine         ← implementations: store, log, setting
    │
internal       ← cast, config, error, path, pattern, monitor
```

## Where a change belongs

| You want to | Change |
| --- | --- |
| add a capability the interface calls | a `service/*` package plus a file in `domain/service` |
| change how data is persisted | `engine/store` |
| add a configuration field | `data/conf/default.json` and its schema |
| change the interface | `frontend/src/features/<domain>` |

## Reading order

Start with [Backend](backend.md) to see how a request travels, then [Frontend](frontend.md) for how the
interface consumes it, and [IPC](ipc.md) for the seam between them.
