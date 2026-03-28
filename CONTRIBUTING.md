# Contributing to OmniGraph



Thank you for your interest in contributing to OmniGraph. The product is **graph-forward**: the React workspace under **`packages/web`** is how most people **see** infrastructure intent and posture. The Go CLI and orchestration paths **validate, run, and emit** the JSON that feeds that UI and CI—they are co-critical but not the whole story. Read [docs/product-philosophy.md](docs/product-philosophy.md) if you are unsure where a change should land.



For the **architectural story** (Go workspaces, Reconciler Engine, Wasm hardening, E2E), read [docs/development/platform-architecture.md](docs/development/platform-architecture.md).



This guide covers local setup for **both** the web app and the control plane.



## Development Prerequisites



### Required Software



- **Go 1.22+** - Matches [CI workflow](.github/workflows/ci.yml)

- **Node.js 20 LTS** and npm - For the web UI (`packages/web`)

- **Git** - Version control



### Optional Software



- **Make** - Convenience build targets (not required)

- **OpenTofu/Terraform** - For testing orchestration workflows

- **Ansible** - For testing configuration management integration

- **Docker/Podman** - For containerized execution testing



## Development Setup



### 1. Clone and Configure



```bash

git clone https://github.com/<ORG_OR_USER>/<REPOSITORY>.git

cd <REPOSITORY>

```



### 2. Go workspace (control plane + Wasm modules)



From the repository root, ensure the **workspace** is synced so all listed modules resolve together:



```bash

go work sync

```



Then build and test the CLI (see below). The root **`go.work`** file groups the main module, **`wasm/*`** toolchains, and any shared **`pkg/`** libraries so backend work stays **decoupled** from the frontend’s npm graph.



### 3. Web workspace (primary UI)



```bash

cd packages/web

npm ci

npm run dev

```



CI parity commands:



```bash

cd packages/web

npm ci

npm run lint

npm run build

```



#### Wasm Integration



The browser UI uses WebAssembly for HCL diagnostics. Build it with:



```bash

make wasm-hcldiag

```



Wasm artifacts are consumed from the web package’s static directory (for example `public/wasm/` under the UI root).



Optional Wasm spike test:



```bash

cd packages/web

# Bash/Linux/macOS

VITE_ENABLE_WASM_SPIKE=true npm run dev



# PowerShell

$env:VITE_ENABLE_WASM_SPIKE = "true"

npm run dev

```



End-user-oriented tab reference: [docs/using-the-web.md](docs/using-the-web.md).



### 4. Control plane (Go)



Build and test the CLI:



```bash

# Run tests

go vet ./...

go test ./...



# Build binary

go build -o bin/omnigraph ./cmd/omnigraph



# Verify installation

./bin/omnigraph --version

./bin/omnigraph validate testdata/sample.omnigraph.schema

./bin/omnigraph coerce --format=tfvars testdata/sample.omnigraph.schema

./bin/omnigraph graph emit testdata/sample.omnigraph.schema \

  --plan-json internal/plan/testdata/minimal-plan.json \

  --tfstate internal/state/testdata/minimal.state.json

```



**Windows PowerShell:**

```powershell

go build -o bin\omnigraph.exe .\cmd\omnigraph

.\bin\omnigraph.exe --version

```



**Using Make (optional):**

```bash

make vet

make test

make build

```



**Fuzzing (Wasm / parser hardening):** from the relevant module (for example `wasm/hcldiag`), run `go test -fuzz` targets as documented in [ADR 008](docs/core-concepts/adr/008-wasm-bridge-hardening.md).



### 5. End-to-end (E2E) suite



Full-pipeline tests with **simulated Ansible endpoints** and **fixture state** live under **`e2e/`**:



```bash

go test ./e2e/...

```



See [docs/development/e2e-testing.md](docs/development/e2e-testing.md).



### 6. Orchestration Testing



Test the full orchestration pipeline (requires OpenTofu workspace and Ansible playbook):



```bash

go build -o bin/omnigraph ./cmd/omnigraph



# Dry run (skip Ansible)

./bin/omnigraph orchestrate --workdir /path/to/tf/root --playbook site.yml --auto-approve --skip-ansible



# Full orchestration with containerized runners

./bin/omnigraph orchestrate \

  --workdir /path/to/tf/root \

  --playbook site.yml \

  --auto-approve \

  --runner=container \

  --container-runtime=docker

```



## Project Structure



```

├── go.work                 # Go workspace: main module + wasm/* (+ pkg/* when present)

├── packages/web/           # Isolated React workspace (npm)

├── e2e/                    # Full-pipeline tests, mock Ansible, failure injection

├── wasm/                   # WebAssembly modules for the UI

├── pkg/

│   └── reconciler/         # Reconciler Engine: IR → emitted artifacts (public package)

├── cmd/omnigraph/          # CLI entry point

├── internal/

│   ├── cli/                # Command implementations

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

- Run `go vet ./...` and `go test ./...` before committing

- Use table-driven tests

- Document exported functions and types



### TypeScript/React



- Use TypeScript strict mode

- Follow ESLint configuration (`packages/web/eslint.config.js`)

- Run `npm run lint` before committing

- Use functional components with hooks



## Pull Request Process



1. **Fork** the repository

2. **Create a feature branch** from `main`

3. **Make your changes** with tests and documentation

4. **Run the test suite** to ensure CI parity:

   ```bash

   go work sync

   go vet ./...

   go test ./...

   go test ./e2e/...

   cd packages/web && npm run lint

   ```

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

- [Reconciler Engine](docs/core-concepts/reconciler-engine.md)

- [Architecture Decision Records](docs/core-concepts/adr/)

- [Execution Matrix](docs/core-concepts/execution-matrix.md)

- [Integrations](docs/core-concepts/integrations.md)



## Getting Help



- Open an issue in your upstream repository

- Check discussions in your upstream repository

- Review the canonical docs under `docs/`. Short wiki navigation lives in `wiki/`; to update the GitHub **Wiki** tab, follow [`wiki/SYNC.md`](wiki/SYNC.md) after enabling Wikis in repo settings.

