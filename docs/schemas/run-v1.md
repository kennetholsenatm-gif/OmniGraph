# Run Artifact Schema (`omnigraph/run/v1`)

The run artifact schema records ordered execution steps for one pipeline run.

Normative schema: `schemas/run.v1.schema.json`.

## Producer guidance

Any CI platform can produce this artifact as long as it can run OmniGraph commands
and publish JSON artifacts.

## Typical fields

- Run metadata (`runId`, repository context, commit context)
- Ordered step list with status and timing
- Artifact references for plan/graph/security outputs
