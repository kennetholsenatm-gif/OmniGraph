# Contributing to OmniGraph



Thank you for your interest in contributing to OmniGraph. The product is **graph-forward**: the React workspace under **`packages/web`** is how most people **see** infrastructure intent and posture. The Go **workspace server**, **`go test`**, and orchestration libraries **validate and emit** the JSON that feeds that UI and CI—they are co-critical but not the whole story. Read [docs/product-philosophy.md](docs/product-philosophy.md) if you are unsure where a change should land.



For the **architectural story** (Go workspaces, Emitter Engine, Wasm hardening, E2E), read [docs/development/platform-architecture.md](docs/development/platform-architecture.md).



This guide covers local setup for **both** the web app and the control plane.

**All copy-paste shell sequences** (clone, `go work sync`, npm, builds, E2E): **[docs/development/contributor-commands.md](docs/development/contributor-commands.md)**.



## Development Prerequisites



### Required Software



- **Go 1.22+** - Matches [CI workflow](.github/workflows/ci.yml)

- **Node.js 20 LTS** and npm - For the web UI (`packages/web`)

- **Git**



### Optional Software



- **Make** - Convenience build targets (not required)

- **OpenTofu/Terraform** - For testing orchestration workflows

- **Ansible** - For testing configuration management integration

- **Docker/Podman** - For containerized execution testing



## Development Setup



### 1. Clone and Configure



Clone the repository and open its root. Exact `git` commands: **[docs/development/contributor-commands.md](docs/development/contributor-commands.md)**.



### 2. Go workspace (control plane + Wasm modules)



From the repository root, ensure the **workspace** is synced so all listed modules resolve together. See **[docs/development/contributor-commands.md](docs/development/contributor-commands.md)** (`go work sync`).

Then build and test the Go control plane using the same document. The root **`go.work`** file groups the main module, **`wasm/*`** toolchains, and any shared **`pkg/`** libraries so backend work stays **decoupled** from the frontend’s npm graph.



### 3. Web workspace (primary UI)



From **`packages/web`**, install dependencies and start the dev server—see **[docs/development/contributor-commands.md](docs/development/contributor-commands.md)**.

CI parity (`npm run lint`, `npm run build`) is in the same document.



#### Wasm Integration



The browser UI uses WebAssembly for HCL diagnostics. Rebuild commands (`make wasm-hcldiag`) and optional spike flags are in **[docs/development/contributor-commands.md](docs/development/contributor-commands.md)**.

Wasm artifacts are consumed from the web package’s static directory (for example `public/wasm/` under the UI root).



End-user-oriented tab reference: [docs/using-the-web.md](docs/using-the-web.md).



### 4. Control plane (Go)



Build and test the workspace server and libraries: **[docs/development/contributor-commands.md](docs/development/contributor-commands.md)** (Go vet, test, build, Windows variants, optional Make).



**Fuzzing (Wasm / parser hardening):** from the relevant module (for example `wasm/hcldiag`), run `go test -fuzz` targets as documented in [ADR 008](docs/core-concepts/adr/008-wasm-bridge-hardening.md).



### 5. End-to-end (E2E) suite



Full-pipeline tests with **simulated Ansible endpoints** and **fixture state** live under **`e2e/`**. Run them from the repo root as in **[docs/development/contributor-commands.md](docs/development/contributor-commands.md)**.

See [docs/development/e2e-testing.md](docs/development/e2e-testing.md).



### 6. Orchestration and IaC runtime



Exercise **OpenTofu/Terraform** and **Ansible** directly against your lab roots. Library code under **`internal/orchestrate`** remains available for integrations; the product surface is the **web workspace**, not a multi-command terminal tool. See [docs/core-concepts/execution-matrix.md](docs/core-concepts/execution-matrix.md).



## Project Structure



```

├── go.work                 # Go workspace: main module + wasm/* (+ pkg/* when present)

├── packages/web/           # Isolated React workspace (npm)

├── e2e/                    # Full-pipeline tests, mock Ansible, failure injection

├── wasm/                   # WebAssembly modules for the UI

├── pkg/

│   └── emitter/            # Emitter Engine: IR → emitted artifacts (public package)

├── cmd/omnigraph/          # Workspace server entry (HTTP API + optional static UI)

├── internal/

│   ├── coerce/             # Schema coercion engine

│   ├── graph/              # Dependency graph generation

│   ├── inventory/          # Dynamic inventory generation

│   ├── orchestrate/        # Pipeline orchestration

│   ├── policy/             # Policy-as-Code (OPA/Rego)

│   ├── reconcile/          # Manifest reconciliation (apply loop)

│   ├── runner/             # Execution runners (exec, container)

│   ├── schema/             # Schema validation

│   ├── security/           # Security scanning

│   ├── serve/              # HTTP API server

│   └── state/              # State management

├── schemas/                # JSON Schema definitions

├── docs/                   # Canonical documentation

├── wiki/                   # Wiki navigation; GitHub Wiki sync (see wiki/SYNC.md)

└── testdata/               # Test fixtures

```



## Code Standards



### Go



- Follow [Effective Go](https://go.dev/doc/effective-go)

- Run `go vet ./...` and `go test ./...` before committing (commands: [Contributor commands](docs/development/contributor-commands.md))

- Use table-driven tests

- Document exported functions and types



### TypeScript/React



- Use TypeScript strict mode

- Follow ESLint configuration (`packages/web/eslint.config.js`)

- Run `npm run lint` before committing ([Contributor commands](docs/development/contributor-commands.md))

- Use functional components with hooks



## Pull Request Process



1. **Fork** the repository

2. **Create a feature branch** from `main`

3. **Make your changes** with tests and documentation

4. **Run the test suite** to ensure CI parity—use the **PR / CI verification checklist** in **[docs/development/contributor-commands.md](docs/development/contributor-commands.md)**

5. **Submit a pull request** with a clear description



### PR Checklist



- [ ] Tests pass locally

- [ ] Documentation updated (if applicable)

- [ ] Commit messages are clear and descriptive

- [ ] No secrets or sensitive data committed



## Security



- **Never commit credentials** or `.env` files with secrets

- Use memory-only secret injection ([ADR 003](docs/core-concepts/adr/003-memory-only-secrets.md))

- Report security vulnerabilities privately

- Wasm bridge discipline: [ADR 008](docs/core-concepts/adr/008-wasm-bridge-hardening.md)



## Architecture



Start with these resources:



- [Platform architecture for contributors](docs/development/platform-architecture.md)

- [Product philosophy](docs/product-philosophy.md)

- [Using the web workspace](docs/using-the-web.md)

- [Architecture Overview](docs/core-concepts/architecture.md)

- [Emitter Engine](docs/core-concepts/emitter-engine.md)

- [Architecture Decision Records](docs/core-concepts/adr/)

- [Execution Matrix](docs/core-concepts/execution-matrix.md)

- [Integrations](docs/core-concepts/integrations.md)



## Getting Help



- Open an issue in your upstream repository

- Check discussions in your upstream repository

- Review the canonical docs under `docs/`. Short wiki navigation lives in `wiki/`; to update the GitHub **Wiki** tab, follow [`wiki/SYNC.md`](wiki/SYNC.md) after enabling Wikis in repo settings (manual git steps: [Contributor commands](docs/development/contributor-commands.md)).
