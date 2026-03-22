# Contributing to OmniGraph

## Prerequisites

- **Go** 1.22+ (matches [CI](.github/workflows/ci.yml))
- **Node.js** 20 LTS and npm (for `web/`)

## Clone and remotes

```bash
git clone https://github.com/kennetholsenatm-gif/OmniGraph.git
cd OmniGraph
```

## Control plane (Go)

From the repository root:

```bash
go vet ./...
go test ./...
go build -o bin/omnigraph ./cmd/omnigraph
./bin/omnigraph -version
```

On Windows PowerShell you can use `.\bin\omnigraph.exe -version` after `go build -o bin/omnigraph.exe ./cmd/omnigraph`.

Optional: if you have `make` installed, `make vet`, `make test`, and `make build` run the same steps.

## Web app

```bash
cd web
npm ci
npm run dev
```

CI parity:

```bash
cd web
npm ci
npm run lint
npm run build
```

## Secrets and sensitive data

Do **not** commit real credentials, `.env` files with secrets, or Terraform/OpenTofu state. See [docs/adr/003-memory-only-secrets.md](docs/adr/003-memory-only-secrets.md).

## Pull requests

- Prefer focused changes with tests and docs updates as needed.
- Use the PR template checklist before requesting review.

## Architecture

Start with [docs/architecture.md](docs/architecture.md) and the [ADRs](docs/adr/).
