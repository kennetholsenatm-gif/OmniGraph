# Changelog Notes

This file tracks notable documentation and architecture-note transitions.

## Current transition

- **Architecture refactor (merged via PR #24):** React app under `packages/web`; Emitter
  Engine in `pkg/emitter` (public; formerly `pkg/reconciler`); root `go.work` for Wasm modules; HCL Wasm handlers
  hardened (`recover` + TS bridge resilience); `wasm/tfpattern` fuzzing in CI; `e2e/` CLI
  and httptest fixtures. Narrative docs: [platform architecture](../development/platform-architecture.md),
  [Emitter Engine](../core-concepts/emitter-engine.md), [ADR 008](../core-concepts/adr/008-wasm-bridge-hardening.md),
  [E2E testing](../development/e2e-testing.md).
- Documentation hierarchy split into core concepts, development, schema references,
  and reference architectures.
- Local environment assumptions replaced with generic placeholders in canonical docs.
- Wiki content aligned to mirror canonical docs.
