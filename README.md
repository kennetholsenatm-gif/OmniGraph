# OmniGraph

OmniGraph is a state-aware infrastructure orchestration engine and toolchain.

This repository is the public OmniGraph codebase only. Private lab infrastructure,
site-specific runbooks, and personal environment wiring are intentionally excluded.

## Quick Start

1. Install Go (`1.23+` recommended).
2. Build:
   - `go build ./cmd/omnigraph`
3. Show CLI help:
   - `./omnigraph --help`

## Project Scope

- Core CLI and control-plane logic in `cmd/` and `internal/`
- Public schemas in `schemas/`
- Public docs in `docs/`
- Web client in `web/`

## Out of Scope

- Private infrastructure topology
- Personal/local workstation paths
- Organization-specific deployment runbooks
