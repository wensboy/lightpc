# Develop

> Quick Guide · Develop

Set up a working copy, run the application and iterate on the interface.

## Prerequisites

| Tool | Version | Why |
| --- | --- | --- |
| Go | 1.24+ | Backend and the Wails host |
| Node.js | 22+ | Frontend toolchain |
| pnpm | 10+ | The only supported package manager |

## First run

```bash
pnpm install
pnpm build:dev
go build -o bin/lightpc .
./bin/lightpc --mode 2
```

`pnpm build:dev` enables the mock switches in `.env.development`, so the interface works without a backend
account service.

## Project layout

- `src/backend` holds the Go side, split into domain, engine and service layers
- `src/frontend` holds the React application
- `src/data` holds the configuration, the schemas and the documentation

## Next steps

Read the [Architecture](../architecture/overview.md) section before changing a layer boundary, and the
[Manual](../manual/engine.md) when adding an engine.
