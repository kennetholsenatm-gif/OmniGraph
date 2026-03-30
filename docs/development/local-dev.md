# Local Development

## Prerequisites

- Go `1.23+` (see [CI workflow](../../.github/workflows/ci.yml) for the exact pin)
- Node.js `20+`
- Git

## Go workspace

From the **repository root**, sync the workspace so all modules listed in **`go.work`** resolve together:

```bash
go work sync
```

This keeps the **control plane** and **Wasm toolchains** independent of the frontend npm graph.

## Build and test (Go)

```bash
go vet ./...
go test ./...
go build -o bin/omnigraph ./cmd/omnigraph
./bin/omnigraph -h
```

See [CI and contributor automation](../ci-and-contributor-automation.md) for smoke tests and HTTP routes.

## Run web frontend

```bash
cd packages/web
npm ci
npm run dev
```

## Suggested verification

```bash
cd packages/web
npm run lint
npm run build
```

## E2E suite

```bash
go test ./e2e/...
```

See [E2E testing](e2e-testing.md).

## Further reading

- [Platform architecture](platform-architecture.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
