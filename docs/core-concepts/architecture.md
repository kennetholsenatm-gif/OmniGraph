# Architecture Overview

OmniGraph separates infrastructure intent, orchestration, and runtime execution into
clear layers so teams can integrate their own providers and delivery workflows.

## Layers

1. Presentation layer: web UI and developer-facing validation feedback
2. Control plane: CLI and orchestration logic in Go
3. Execution layer: host and container runners for external tools
4. Integration layer: inventory, telemetry, identity, and policy adapters

## Key Design Principles

- Schema-first contracts before imperative execution
- Tool-agnostic orchestration rather than tool replacement
- Versioned data formats (`omnigraph/*/v1`) for compatibility
- Explicit boundaries between core behavior and environment-specific examples

## Related Docs

- `docs/core-concepts/omnigraph-ir.md`
- `docs/core-concepts/state-management.md`
- `docs/core-concepts/execution-matrix.md`
- `docs/reference-architectures/overview.md`
