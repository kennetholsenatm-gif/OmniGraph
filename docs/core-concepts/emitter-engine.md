# Emitter Engine

The **Emitter Engine** is OmniGraph’s **first-class translation layer**: it reads versioned **declarative intent** (`omnigraph/ir/v1`) and **emits concrete artifacts** that execution tools understand—**Ansible INI inventory** from `spec.targets` today, with additional formats registered over time. It is the architectural heart between “what the graph says” and “what Ansible (and siblings) can run.”

For **manifest reconciliation** (desired vs actual infrastructure via `internal/reconcile` and related libraries), see [Declarative reconciliation (reference architecture)](../reference-architectures/declarative-reconciliation.md). **Orchestration** (OpenTofu/Ansible pipeline stages) chains external tools against a workspace. The Emitter Engine **materializes files and blobs from IR**; orchestration **invokes external programs** with those artifacts and other inputs.

## Why it deserves its own name

Previously, emitters could feel like **hidden implementation detail** buried under `internal/`. Treating them as the **Emitter Engine** makes three things obvious to contributors:

1. **Extension is expected**—new output shapes should land as new backends, not one-off scripts.
2. **The contract is narrow**—inputs are `omnigraph/ir/v1`; outputs are typed **artifacts** with paths and media types.
3. **Failures are ordinary Go errors**—callers decide policy; the engine does not swallow mistakes.

## Implementation home

The engine is implemented in Go as the **public** package [`pkg/emitter`](../../pkg/emitter/). Normative JSON Schema for the input document: [`schemas/ir.v1.schema.json`](../../schemas/ir.v1.schema.json). Conceptual IR fields are summarized in [OmniGraph IR](omnigraph-ir.md).

## Core interfaces

A **backend** implements a single output **format** and knows how to **emit** from an IR document:

- **`Format() string`** — Stable identifier (for example `ansible-inventory-ini`). All known format ids are enumerated in code for docs and test parity.
- **`Emit(ctx context.Context, doc *Document) ([]Artifact, error)`** — Produces zero or more **artifacts** (path, media type, description, raw bytes). A nil document is rejected with an error; partial success is expressed by returning an error, not a panic.

A **registry** holds named backends:

- **`Register(b Backend)`** — Registers a backend; duplicate or invalid registration is a programmer error (panic), not a runtime user error.
- **`Get(format string) Backend`** — Lookup for a single format.
- **`Emit(ctx, format, doc)`** — Runs one backend by format id.

The default registry wires **real emitters** where implemented and **stub backends** that return `ErrNotImplemented` for formats still on the roadmap—so the surface area is visible without pretending unfinished work is complete.

## Extending safely

1. **Define or reuse a format constant** aligned with [`AllFormats`](../../pkg/emitter/format.go) (add a new constant there if you introduce a new id).
2. **Implement `Backend`** with deterministic `Emit` behavior; validate inputs and return wrapped errors with context.
3. **Register** your backend in the default registry constructor (alongside existing implementations).
4. **Add tests** in the style of [`emit_ansible_test.go`](../../pkg/emitter/emit_ansible_test.go): table-driven cases, golden or string-stable expectations, and explicit error paths.

Selection of which backends run for a given workflow may use **`spec.emitHints`** (ordering hints) or caller options in Go; keep registration in one place so discovery stays centralized.

## Related documentation

- [OmniGraph IR](omnigraph-ir.md) — document shape and purpose
- [Platform architecture for contributors](../development/platform-architecture.md) — narrative context and glossary
- [Declarative reconciliation (reference)](../reference-architectures/declarative-reconciliation.md) — desired-vs-actual manifest control (`apply`)
- [Execution matrix](execution-matrix.md) — how runners consume emitted artifacts in broader flows
