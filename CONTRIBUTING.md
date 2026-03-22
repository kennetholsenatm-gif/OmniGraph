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
./bin/omnigraph --version
./bin/omnigraph validate testdata/sample.omnigraph.schema
./bin/omnigraph coerce --format=tfvars testdata/sample.omnigraph.schema
./bin/omnigraph graph emit testdata/sample.omnigraph.schema \
  --plan-json internal/plan/testdata/minimal-plan.json \
  --tfstate internal/state/testdata/minimal.state.json
```

On Windows PowerShell you can use `.\bin\omnigraph.exe --version` after `go build -o bin\omnigraph.exe .\cmd\omnigraph`.

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

Optional Wasm spike (browser `WebAssembly.instantiate` smoke test):

```bash
cd web
set VITE_ENABLE_WASM_SPIKE=true
npm run dev
```

(On PowerShell, use `$env:VITE_ENABLE_WASM_SPIKE = "true"` before `npm run dev`.)

### HCL diagnostics Wasm (ADR 001)

The web app loads `/wasm/hcldiag.wasm` for real-time HCL parse feedback. Build it with Go 1.22+:

```bash
make wasm-hcldiag
```

`web/public/wasm/wasm_exec.js` is checked in (from the Go distribution). CI rebuilds both `wasm_exec.js` and `hcldiag.wasm` before the web build.

### Orchestrate (magic handoff)

From the repo root (with a real OpenTofu workspace and playbook):

```bash
go build -o bin/omnigraph ./cmd/omnigraph
./bin/omnigraph orchestrate --workdir /path/to/tf/root --playbook site.yml --auto-approve --skip-ansible
```

Use `--runner=container` and `--container-runtime=docker` to run OpenTofu/Ansible inside ephemeral containers (see `docs/execution-matrix.md`).

## Secrets and sensitive data

Do **not** commit real credentials, `.env` files with secrets, or Terraform/OpenTofu state. See [docs/adr/003-memory-only-secrets.md](docs/adr/003-memory-only-secrets.md).

## Pull requests

- Prefer focused changes with tests and docs updates as needed.
- Use the PR template checklist before requesting review.

## Architecture

Start with [docs/architecture.md](docs/architecture.md) and the [ADRs](docs/adr/).
