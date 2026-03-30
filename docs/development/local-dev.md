# Local Development

## Prerequisites

- Go `1.23+` (see [CI workflow](../../.github/workflows/ci.yml) for the exact pin)
- Node.js `20+`
- Git

## Go workspace

From the **repository root**, sync the workspace so all modules listed in **`go.work`** resolve together. Exact commands: **[Contributor commands](contributor-commands.md)** (`go work sync`).

This keeps the **control plane** and **Wasm toolchains** independent of the frontend npm graph.

## Build and test (Go)

Run `go vet`, `go test`, and build the workspace server binary as documented in **[Contributor commands](contributor-commands.md)**.

See [CI and contributor automation](../ci-and-contributor-automation.md) for smoke tests and HTTP routes.

## Run web frontend

Use **`packages/web`** with `npm ci` and `npm run dev`—see **[Contributor commands](contributor-commands.md)**.

## Suggested verification

Lint and production build for the web package: **[Contributor commands](contributor-commands.md)**.

## E2E suite

From the repository root, run the e2e package as in **[Contributor commands](contributor-commands.md)**.

See [E2E testing](e2e-testing.md).

## Further reading

- [Platform architecture](platform-architecture.md)
- [CONTRIBUTING.md](../../CONTRIBUTING.md)
- [Contributor commands](contributor-commands.md)
