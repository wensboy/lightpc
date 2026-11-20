# IPC

> Architecture · IPC

The frontend never speaks to an engine. It calls a service, either directly through a generated binding or
indirectly through a bus event the app runtime forwards to the webview.

## Calling a service

```
component
  └─ features/<name>/service wrapper      // validates with Zod
       └─ bindings/.../service/<name>     // generated, one function per method
            └─ Go service method
```

Regenerate the bindings after changing a service signature:

```bash
wails3 generate bindings
```

`bindings/` is generated output. It is committed so a fresh checkout compiles without running the generator,
but it must never be edited by hand.

## Events

```
engine      ──emit──▶  bus  ──▶  app runtime ar hook  ──▶  webview event
service     ──emit──▶  bus
```

Only events that make the interface render again cross the seam. Everything else stays inside the backend,
where a service or engine listens on the same bus.

## Errors

A service returns an error key (`error::service.<name>::<reason>`). The runtime marshals it to a shape the
frontend can display, so a component shows a message instead of a stack trace.
