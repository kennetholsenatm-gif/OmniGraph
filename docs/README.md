# OmniGraph documentation

**Wayfinding:** You are in the **canonical documentation tree** (`docs/` on `main`). Pages under **`guides/`** are task-oriented; **`core-concepts/`** holds architecture and contracts; **`development/`** is for contributors building the product. The GitHub **Wiki** (`wiki/`) mirrors short navigation only—long-form content stays here.

OmniGraph is a **graph-forward web workspace** backed by schema-first contracts (`.omnigraph.schema` in **TOML** recommended, or YAML/JSON; `omnigraph/*/v1` artifacts). The React UI lives in the isolated npm package [`packages/web`](../packages/web). The Go **workspace server** and **`go test`** paths validate and **emit** the JSON that feeds that view and CI—essential for contributors, not the elevator pitch. See [product-philosophy.md](product-philosophy.md).

**Canonical copy lives here.** The GitHub wiki under `wiki/` links back to these paths.

## 15-minute reading order

1. [Getting started (workspace only)](getting-started.md) — first session in the UI, no terminal.
2. [Product philosophy](product-philosophy.md) — visualization-first positioning; local server and tests.
3. [UX architecture](core-concepts/ux-architecture.md) — progressive disclosure, backend truth (SSE), contextual debugging.
4. [Understanding the UI modes](guides/ui-modes.md) — Topology, Reconciliation, Posture; mapping to sidebar tabs.
5. [Graph dependencies and blast radius](guides/graph-dependencies-and-blast-radius.md) — `dependencyRole` on graph edges; Go/TS blast helpers.
6. [Data handoff](core-concepts/data-handoff.md) — provider artifacts, SSE, Inventory, Topology (technical bridge).
7. [NOC / SRE workflow](guides/workflows-noc-sre.md) · [SOC / SecOps workflow](guides/workflows-soc-secops.md) — persona-shaped paths through the workspace.
8. [Platform architecture for contributors](development/platform-architecture.md) — narrative: workspaces, Emitter Engine, Wasm safety, E2E (start here if you ship code).
9. [Using the web workspace](using-the-web.md) — run the UI; what each tab does.
10. [Overview](overview.md) — who / what / where, diagrams, artifacts.
11. [Architecture](core-concepts/architecture.md) — layers (presentation first).
12. [Execution matrix](core-concepts/execution-matrix.md) — runners; how orchestration feeds artifacts the UI consumes.
13. [Security posture](security/posture.md) — policy, scans, `serve` hardening, Wasm boundary.

Then: [CI and contributor automation](ci-and-contributor-automation.md) (`go test`, workspace server, `testdata/`), **[Contributor commands](development/contributor-commands.md)** (single copy-paste shell reference), [Cognitive design standards](development/cognitive-design-standards.md), [Cognitive validation gates](development/cognitive-validation-gates.md), [IR model](core-concepts/omnigraph-ir.md), [Emitter Engine](core-concepts/emitter-engine.md), [E2E testing](development/e2e-testing.md), [schemas](schemas/), [reference architectures](reference-architectures/overview.md) (non-normative).

## Find documentation by intent

| Intent | Start here |
|--------|------------|
| Terms and acronyms (graph, schema shapes, WASM-backed HCL) | [GLOSSARY.md](GLOSSARY.md) |
| Why graph/UI leads; anti-“generic CLI” | [product-philosophy.md](product-philosophy.md) |
| Why graph/UI leads; workspace-first product | [product-philosophy.md](product-philosophy.md) |
| First session in the browser only | [getting-started.md](getting-started.md) |
| Dependency roles + blast radius on graph edges | [guides/graph-dependencies-and-blast-radius.md](guides/graph-dependencies-and-blast-radius.md) |
| Provider outputs to UI (SSE, Inventory, graph emit) | [core-concepts/data-handoff.md](core-concepts/data-handoff.md) |
| NOC/SRE or SOC/SecOps shaped tours | [guides/workflows-noc-sre.md](guides/workflows-noc-sre.md), [guides/workflows-soc-secops.md](guides/workflows-soc-secops.md) |
| Use the graph and workspace in the browser | [using-the-web.md](using-the-web.md) |
| What is this and for whom? | [overview.md](overview.md) |
| Automation, CI, contributor verification | [ci-and-contributor-automation.md](ci-and-contributor-automation.md), [Contributor commands](development/contributor-commands.md), [examples/quickstart/README.md](../examples/quickstart/README.md) |
| How is the system structured? | [core-concepts/architecture.md](core-concepts/architecture.md) |
| How runs produce graph/run artifacts | [core-concepts/execution-matrix.md](core-concepts/execution-matrix.md) |
| Schema and API contracts | [schemas/](schemas/), [core-concepts/omnigraph-ir.md](core-concepts/omnigraph-ir.md) |
| Policy, scans, serve safety | [security/posture.md](security/posture.md) |
| Example deployment patterns (illustrative only) | [reference-architectures/](reference-architectures/) |
| Build, test, contribute | [development/local-dev.md](development/local-dev.md), [Contributor commands](development/contributor-commands.md), [Cognitive design standards](development/cognitive-design-standards.md), [CONTRIBUTING.md](../CONTRIBUTING.md) |

## Section map

- **`examples/quickstart/`** (repo root) — minimal `.tfstate.json` + `.omnigraph.schema` for graph emit tests and Topology fixtures.
- **guides/** — task-oriented guides (UI modes, graph dependencies / blast radius, persona workflows).
- **core-concepts/** — architecture, UX architecture, IR, state, integrations, execution, ADRs, inventory.
- **development/** — platform architecture narrative, local builds, web frontend, E2E, **[Contributor commands](development/contributor-commands.md)** (shell reference), [Cognitive standards](development/cognitive-design-standards.md), [Validation gates](development/cognitive-validation-gates.md), [WASI parser plugins](development/wasm-plugins.md).
- **reference-architectures/** — example topologies; adapt to your standards.
- **schemas/** — contract references for IR, run, and related formats.
- **security/** — policy, scans, operational hardening for `serve`.
- **meta/** — changelog-style notes.

## See also

- [State management](core-concepts/state-management.md)
- [ADR index](core-concepts/adr/) (individual decision records)
