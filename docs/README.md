# OmniGraph documentation

OmniGraph is a **graph-forward web workspace** backed by schema-first contracts (`.omnigraph.schema`, `omnigraph/*/v1`). The React UI in [`web/`](../web/) is the primary place people **see** intent, topology, pipeline context, and posture. The Go **`omnigraph`** CLI validates, orchestrates, scans, and **emits** the JSON that feeds that view and CI—it is essential automation, not the elevator pitch. See [product-philosophy.md](product-philosophy.md).

**Canonical copy lives here.** The GitHub wiki under `wiki/` links back to these paths.

## 15-minute reading order

1. [Product philosophy](product-philosophy.md) — visualization-first positioning; what the CLI is for.
2. [Using the web workspace](using-the-web.md) — run the UI; what each tab does.
3. [Overview](overview.md) — who / what / where, diagrams, artifacts.
4. [Architecture](core-concepts/architecture.md) — layers (presentation first).
5. [Execution matrix](core-concepts/execution-matrix.md) — runners; how orchestration feeds artifacts the UI consumes.
6. [Security posture](security/posture.md) — policy, scans, `serve` hardening.

Then: [CLI and CI](cli-and-ci.md) (headless commands and `testdata/`), [IR model](core-concepts/omnigraph-ir.md), [schemas](schemas/), [reference architectures](reference-architectures/overview.md) (non-normative).

## Find documentation by intent

| Intent | Start here |
|--------|------------|
| Why graph/UI leads; anti-“generic CLI” | [product-philosophy.md](product-philosophy.md) |
| Use the graph and workspace in the browser | [using-the-web.md](using-the-web.md) |
| What is this and for whom? | [overview.md](overview.md) |
| Automation, CI, terminal workflows | [cli-and-ci.md](cli-and-ci.md) |
| How is the system structured? | [core-concepts/architecture.md](core-concepts/architecture.md) |
| How runs produce graph/run artifacts | [core-concepts/execution-matrix.md](core-concepts/execution-matrix.md) |
| Schema and API contracts | [schemas/](schemas/), [core-concepts/omnigraph-ir.md](core-concepts/omnigraph-ir.md) |
| Policy, scans, serve safety | [security/posture.md](security/posture.md) |
| Example deployment patterns (illustrative only) | [reference-architectures/](reference-architectures/) |
| Build, test, contribute | [development/local-dev.md](development/local-dev.md), [CONTRIBUTING.md](../CONTRIBUTING.md) |

## Section map

- **core-concepts/** — architecture, IR, state, integrations, execution, ADRs, inventory.
- **development/** — local builds, web frontend, contributing pointers.
- **reference-architectures/** — example topologies; adapt to your standards.
- **schemas/** — contract references for IR, run, and related formats.
- **security/** — policy, scans, operational hardening for `serve`.
- **meta/** — changelog-style notes.

## See also

- [State management](core-concepts/state-management.md)
- [ADR index](core-concepts/adr/) (individual decision records)
