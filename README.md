# OmniGraph

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

OmniGraph is a state-aware, multi-paradigm DevSecOps orchestration platform. It bridges declarative provisioning (OpenTofu/Terraform), imperative configuration (Ansible), and environment telemetry (NetBox, Zabbix) through a unified GitOps workflow.

## Repository layout

| Area | Path |
|------|------|
| Web UI (React, Tailwind; future D3 graph + Wasm linters) | `web/` |
| Control plane CLI (Go) | `cmd/omnigraph/` |
| JSON Schema for `.omnigraph.schema` | `schemas/` |
| Architecture and ADRs | `docs/` |
| Wasm linter roadmap | `wasm/` |

## Quick start

See [CONTRIBUTING.md](CONTRIBUTING.md) for prerequisites, local development, and CI parity commands.

## Documentation

- [Architecture overview](docs/architecture.md)
- [Execution matrix (plugins)](docs/execution-matrix.md)
- [Integrations](docs/integrations.md)
- [Architecture Decision Records](docs/adr/)

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE).
