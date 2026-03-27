# IR Contracts

This page summarizes key OmniGraph schema contracts used across CLI and UI flows.

## Primary contracts

- `omnigraph/ir/v1` -> `schemas/ir.v1.schema.json`
- `omnigraph/run/v1` -> `schemas/run.v1.schema.json`
- `omnigraph/graph/v1` -> `schemas/graph.v1.schema.json`
- `omnigraph/security/v1` -> `schemas/security.v1.schema.json`

## Versioning policy

- Keep `apiVersion` stable for non-breaking changes
- Introduce new version for breaking changes
- Preserve backward compatibility where feasible for automation consumers

## Related docs

- [Overview](../overview.md)
- [Using the web workspace](../using-the-web.md)
- [CLI and CI](../cli-and-ci.md)
- [OmniGraph IR](../core-concepts/omnigraph-ir.md)
- [Security posture](../security/posture.md)
