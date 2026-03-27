# OmniGraph documentation

OmniGraph is a schema-first control plane: you describe project intent in `.omnigraph.schema` (and related contracts), validate it, optionally chain OpenTofu/Terraform-style planning with Ansible handoff, and emit versioned artifacts (`omnigraph/graph/v1`, telemetry, security posture) for the web UI or CI.

**Canonical copy lives here.** The GitHub wiki under `wiki/` only links back to these paths so we do not maintain two divergent narratives.

## 15-minute reading order

1. [Overview](overview.md) — who uses this, what problems it addresses, where things live in the repo.
2. [Journeys](journeys.md) — copy-paste CLI flows against `testdata/`.
3. [Architecture](core-concepts/architecture.md) — layers and design principles (with diagram).
4. [Execution matrix](core-concepts/execution-matrix.md) — runners, orchestration phases, trade-offs.
5. [Security posture](security/posture.md) — policy, scans, `serve` hardening, ADR pointers.

Then go deeper: [IR model](core-concepts/omnigraph-ir.md), [schemas](schemas/), and [reference architectures](reference-architectures/overview.md) (non-normative examples).

## Find documentation by intent

| Intent | Start here |
|--------|------------|
| What is this and for whom? | [overview.md](overview.md) |
| What can I run end-to-end? | [journeys.md](journeys.md) |
| How is the system structured? | [core-concepts/architecture.md](core-concepts/architecture.md) |
| How are external tools orchestrated? | [core-concepts/execution-matrix.md](core-concepts/execution-matrix.md) |
| Schema and API contracts | [schemas/](schemas/), [core-concepts/omnigraph-ir.md](core-concepts/omnigraph-ir.md) |
| Policy, scans, serve safety | [security/posture.md](security/posture.md) |
| Example deployment patterns (illustrative only) | [reference-architectures/](reference-architectures/) |
| Build, test, contribute | [development/local-dev.md](development/local-dev.md), [CONTRIBUTING.md](../CONTRIBUTING.md) |

## Section map

- **core-concepts/** — architecture, IR, state, integrations, execution, ADRs, inventory.
- **development/** — local builds, web frontend, contributing pointers.
- **reference-architectures/** — example topologies; adapt to your standards.
- **schemas/** — contract references for IR, run, and related formats.
- **security/** — consolidated DevSecOps-oriented view of CLI and server behavior.
- **meta/** — changelog-style notes.

## See also

- [State management](core-concepts/state-management.md)
- [ADR index](core-concepts/adr/) (individual decision records)
