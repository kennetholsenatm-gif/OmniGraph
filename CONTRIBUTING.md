# Contributing to OmniGraph

Thank you for your interest in contributing to OmniGraph! This guide covers everything you need to set up a development environment and contribute effectively.

## Development Prerequisites

### Required Software

- **Go 1.22+** - Matches [CI workflow](.github/workflows/ci.yml)
- **Node.js 20 LTS** and npm - For the web UI (`web/`)
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

### 2. Control Plane (Go)

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

### 3. Web Application

```bash
cd web
npm ci
npm run dev
```

CI parity commands:

```bash
cd web
npm ci
npm run lint
npm run build
```

#### Wasm Integration

The browser UI uses WebAssembly for HCL diagnostics. Build it with:

```bash
make wasm-hcldiag
```

Optional Wasm spike test:

```bash
cd web
# Bash/Linux/macOS
VITE_ENABLE_WASM_SPIKE=true npm run dev

# PowerShell
$env:VITE_ENABLE_WASM_SPIKE = "true"
npm run dev
```

### 4. Orchestration Testing

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
├── cmd/omnigraph/          # CLI entry point
├── internal/
│   ├── cli/                # Command implementations
│   ├── coerce/             # Schema coercion engine
│   ├── graph/              # Dependency graph generation
│   ├── inventory/          # Dynamic inventory generation
│   ├── ir/                 # Infrastructure Intent Reference
│   ├── orchestrate/        # Pipeline orchestration
│   ├── policy/             # Policy-as-Code (OPA/Rego)
│   ├── runner/             # Execution runners (exec, container)
│   ├── schema/             # Schema validation
│   ├── security/           # Security scanning
│   ├── serve/              # HTTP API server
│   └── state/              # State management
├── schemas/                # JSON Schema definitions
├── web/                    # React frontend
├── wasm/                   # WebAssembly modules
├── docs/                   # Canonical documentation
├── wiki/                   # Wiki navigation + sync instructions (see wiki/SYNC.md)
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
- Follow ESLint configuration (`web/eslint.config.js`)
- Run `npm run lint` before committing
- Use functional components with hooks

## Pull Request Process

1. **Fork** the repository
2. **Create a feature branch** from `main`
3. **Make your changes** with tests and documentation
4. **Run the test suite** to ensure CI parity:
   ```bash
   go vet ./...
   go test ./...
   cd web && npm run lint
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

## Architecture

Start with these resources:

- [Architecture Overview](docs/core-concepts/architecture.md)
- [Architecture Decision Records](docs/core-concepts/adr/)
- [Execution Matrix](docs/core-concepts/execution-matrix.md)
- [Integrations](docs/core-concepts/integrations.md)

## Getting Help

- Open an issue in your upstream repository
- Check discussions in your upstream repository
- Review the canonical docs under `docs/`. Short wiki-style navigation lives in `wiki/`; to update the GitHub **Wiki** tab, follow [`wiki/SYNC.md`](wiki/SYNC.md) after enabling Wikis in repo settings.