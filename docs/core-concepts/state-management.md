# State Management

State management in OmniGraph focuses on deterministic orchestration, locking, and
safe hand-off across pipeline phases.

## Principles

- Versioned lock and state contracts
- Explicit ownership of apply phases
- Safe concurrency behavior for CI workflows
- Audit-friendly run metadata

See ADRs under `docs/core-concepts/adr/` for design decisions that define locking,
schema-first configuration, and secret handling.
