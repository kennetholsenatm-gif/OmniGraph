# OmniGraph

OmniGraph is a unified Infrastructure as Code (IaC) control plane for modeling,
validating, and orchestrating infrastructure intent across multiple execution tools.
It provides a schema-first contract, a Go-based orchestration backend, and a web UI
for graph and run visibility.

## Core Capabilities

- Schema-first infrastructure contracts and validation workflows
- Native infrastructure intent model (`omnigraph/ir/v1`)
- CLI orchestration across planning, apply, inventory, and post-apply steps
- Pluggable execution runners (host and container patterns)
- Web UI for graph visualization and run timeline artifacts
- Optional WebAssembly-based local lint/analysis components

## Generic Quickstart

### 1) Build the CLI

```bash
go build -o bin/omnigraph ./cmd/omnigraph
./bin/omnigraph --help
```

PowerShell:

```powershell
go build -o bin\omnigraph.exe .\cmd\omnigraph
.\bin\omnigraph.exe --help
```

### 2) Run the web UI locally

```bash
cd web
npm ci
npm run dev
```

### 3) Validate a schema artifact (example)

```bash
./bin/omnigraph validate testdata/sample.omnigraph.schema
```

## Repository Layout

- `cmd/` and `internal/`: CLI and control plane implementation
- `schemas/`: versioned schema contracts
- `docs/`: canonical documentation
- `web/`: React frontend
- `wasm/`: WebAssembly modules used by UI/runtime features

## Documentation Map

- Core concepts: `docs/core-concepts/`
- Development workflows: `docs/development/`
- Reference architectures (examples): `docs/reference-architectures/`
- Schema references: `docs/schemas/`

## Reference Architectures

Reference architecture documentation describes **example** environments only.
These examples are intentionally non-normative and should be adapted to your
organization's network, identity, and platform standards.
