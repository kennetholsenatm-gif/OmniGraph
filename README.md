# OmniGraph

OmniGraph is a **schema-first control plane** for infrastructure work: validate project intent, orchestrate OpenTofu/Terraform-style planning with optional Ansible handoff, and emit **versioned artifacts** (graphs, telemetry, security posture) for automation or the web UI. It coordinates the tools you already use; it does not replace them.

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

## Documentation

**Reading order:** start with the [documentation hub](docs/README.md), then [Overview](docs/overview.md) (who / what / where) and [Journeys](docs/journeys.md) (hands-on CLI scenarios with `testdata/`).

| Section | Path |
|--------|------|
| Core concepts | [docs/core-concepts/](docs/core-concepts/) |
| Security and policy surface | [docs/security/posture.md](docs/security/posture.md) |
| Development | [docs/development/](docs/development/) |
| Reference architectures (examples) | [docs/reference-architectures/](docs/reference-architectures/) |
| Schema references | [docs/schemas/](docs/schemas/) |

## Reference Architectures

Reference architecture documentation describes **example** environments only.
These examples are intentionally non-normative and should be adapted to your
organization's network, identity, and platform standards.
