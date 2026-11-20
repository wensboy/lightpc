# Backend

> Architecture · Backend

The Go side is a small set of layers with one direction of dependency: `api` → `service` → `domain`, with
`engine` implementing the domain contracts.

## Layers

| Layer | Directory | Holds |
| --- | --- | --- |
| api | `backend/api` | command and ui entry points, binds services to the webview |
| service | `backend/service` | use cases, one package per service |
| domain | `backend/domain` | contracts only: `engine`, `service`, `bus`, `context`, `config` |
| engine | `backend/engine` | implementations: `store`, `log`, `setting` |
| internal | `backend/internal` | helpers with no domain meaning |

## Startup

```
AppRuntime.Run
  ├─ EngineRuntime.Run   // topological order by Deps(), Init then Start
  ├─ ServiceRuntime.Run  // same order, gets the engine registry injected
  └─ efn(AppEventContext)
```

An engine declares `Deps()` by string key (`store.fs`, `log.direct`). A missing dependency aborts the whole
pipeline, so an engine must only depend on what the configuration always creates.

## Adding a service

1. declare the contract and payload types in `backend/domain/service`;
2. implement `Group` / `Init` / `Deinit` plus the methods in `backend/service/<name>`;
3. register it in `backend/api/ui/entry.go`;
4. regenerate the bindings so the frontend can call it.

Services receive `ServiceContext` with the engine registry and the app context, which is where
configuration and embedded resources come from.
