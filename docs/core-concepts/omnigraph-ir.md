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

## Related docs

- [Overview](../overview.md)
- [Journeys](../journeys.md)
- [IR contracts reference](../schemas/ir-contracts.md)
- [Security posture](../security/posture.md)
