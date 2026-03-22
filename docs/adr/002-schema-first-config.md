# ADR 002: Schema-first configuration (`.omnigraph.schema`)

## Status

Accepted

## Context

Terraform/OpenTofu and Ansible each have their own variable formats (`tfvars`, `group_vars`, extra vars). Duplication and drift between them cause failed applies and configuration bugs.

## Decision

Abstract user-facing variables into a **single** JSON or YAML document governed by `.omnigraph.schema` (JSON Schema in `schemas/omnigraph.schema.json`). The control plane **coerces** this document into tool-specific artifacts **in memory** during a run.

## Consequences

- **Positive:** One source of truth; type safety before execution; easier IDE validation.
- **Negative:** Requires maintaining mapping logic from schema to each toolchain.
- **Mitigation:** Version the schema; generate stubs and docs from the schema where possible.
