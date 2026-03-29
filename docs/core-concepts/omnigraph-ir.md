# OmniGraph IR (`omnigraph/ir/v1`)

The OmniGraph IR is a versioned, engine-neutral infrastructure intent document used
for validation, graphing, and backend emission workflows.

Normative schema: `schemas/ir.v1.schema.json`.

## Purpose

- Represent desired state once and emit to multiple backend formats
- Keep policy and authorization checks at a stable contract boundary
- Allow UI and automation workflows to reason about infrastructure consistently

## Core Fields

- `apiVersion`: must be `omnigraph/ir/v1`
- `kind`: must be `InfrastructureIntent`
- `metadata`: labels and naming metadata
- `spec.targets[]`: deployment targets
- `spec.components[]`: abstract infrastructure components
- `spec.relations[]`: dependencies and topology edges

## Notes

IR describes *intent*, not a specific deployment product stack. Any provider details
outside the IR contract belong in environment-specific configuration or reference
architecture examples.

## Relations to graphs, telemetry, and blast radius

- **`spec.relations[]`** in IR describe **static** dependencies between abstract components. When the emitter builds **`omnigraph/graph/v1`**, those relations become edges; optional **`dependencyRole`** on graph edges refines **hard** (**necessary**) vs **soft** (**sufficient**) dependencies for operational math (downstream blast closure, triage highlighting).
- **Live telemetry** (inventory refreshes, workspace **SSE** summaries, future diagnostic feeds) supplies **signals about runtime state**. It should be interpreted **together** with the last validated graph: telemetry can prompt **re-emit** or **re-validate** the graph when drift is detected, but it does not silently rewrite IR on its own.
- **Diagnostic-style workflows** (incident narrowing, “what breaks if this resource changes?”) combine:
  1. **Declared** edges and roles from graph JSON.
  2. **Plan mutations** from OpenTofu/Terraform **`resource_changes`** (when present) to choose seed vertices.
  3. **Optional** live updates from the browser or control plane to refresh **node state** and attributes shown on the canvas.

Documenting this split keeps expectations honest: the IR and graph schemas remain **versioned contracts**; SSE and similar streams are **observation layers** that trigger human or automated **reconciliation**, not a second hidden source of graph truth.

## Related docs

- [Emitter Engine](emitter-engine.md) — how IR is compiled into execution artifacts
- [Overview](../overview.md)
- [Using the web workspace](../using-the-web.md)
- [CLI and CI](../cli-and-ci.md)
- [IR contracts reference](../schemas/ir-contracts.md)
- [Security posture](../security/posture.md)
- [Platform architecture for contributors](../development/platform-architecture.md)
