# Graph dependencies and blast radius

OmniGraph’s **topology** is carried in **`omnigraph/graph/v1`** JSON: `spec.nodes` and `spec.edges`.  
**This is not the same file as `.omnigraph.schema`.** The Project document (`.omnigraph.schema`, `omnigraph/v1alpha1`) holds **schema-first project intent** (network, tags, metadata). **Author that Project file in TOML** when you want the least friction for humans; the CLI and UI also accept YAML and JSON. **Dependency logic for incident triage** lives on **graph edges** in the **graph v1 JSON** so the canvas, emitters, and automation share one contract.

## Where to declare dependency roles

On each edge in `spec.edges`, you may set:

| Field (JSON)       | Values                         | Default if omitted |
|--------------------|--------------------------------|--------------------|
| `dependencyRole`   | `necessary`, `sufficient`      | **`necessary`**    |

- **`necessary`** — A **hard** dependency on the **critical path**. If the source fails, anything **downstream** along necessary edges is treated as **in blast radius** for triage overlays and the diagnostic helpers in the Go package `internal/graph`.
- **`sufficient`** — An **optimization**, convenience, or **fallback** path (e.g. extra CMDB context). It **does not** expand the **primary** blast radius when computing “what else breaks if this fails,” so operators are not flooded with optional paths during an incident.

### Small example

```json
{
  "from": "api-gateway",
  "to": "auth-service",
  "kind": "routes",
  "dependencyRole": "necessary"
},
{
  "from": "telemetry-netbox",
  "to": "auth-service",
  "kind": "context",
  "dependencyRole": "sufficient"
}
```

If `auth-service` is the incident node, **necessary** edges drive **downstream** impact; the NetBox link is **context**, not treated as part of the primary blast set.

Canonical JSON Schema: [schemas/graph.v1.schema.json](../../schemas/graph.v1.schema.json).

## Blast radius (diagnostic engine)

The **blast radius** answers: *“If these nodes are failing, what else is on the **necessary** dependency cone?”* without manually tracing every edge on the canvas.

**Go API** (authoritative for CI, emitters, and tests):

- `graph.DownstreamBlast(doc, incidentIDs)` — Transitive closure following **outgoing** edges where `dependencyRole` is **effective necessary** (omitted counts as necessary). Includes the incident nodes. Result IDs are sorted.
- `graph.UpstreamBlast(doc, incidentIDs)` — Same, but **backward** along necessary edges (useful for **root-context** / “what feeds this?” reasoning).

The web workspace uses the **same rules** in TypeScript (`packages/web/src/graph/blastRadius.ts`) for canvas-side blast math when graph JSON is loaded. Shared fixtures live under [testdata/graph/blast_fixtures.json](../../testdata/graph/blast_fixtures.json) and are exercised by Go tests.

### Vocabulary

Use **graph-forward** language: **downstream**, **upstream**, **necessary path**, **fractured** / **unresolved** where drift docs apply. Do **not** use the word **ghost** for missing or hypothetical nodes (see drift reporting elsewhere).

## Relationship to `.omnigraph.schema`

After you edit the **Project** file (TOML recommended, YAML/JSON supported), you still **emit or author** **`omnigraph/graph/v1`** JSON for the topology the workspace renders. Keep **dependency roles** on **graph edges**; do not invent parallel edge semantics on the Project document.

## Lifecycle, auditing, and drift reconciliation

`dependencyRole` is **intent** about the critical path. It goes stale when architecture, failover design, or observability proves the graph wrong.

### When to add or change tags

- **New integration** — A service, datastore, or control plane now participates in user-visible failure. Decide: is it **necessary** (outage there breaks the product path) or **sufficient** (context, cache, optional telemetry)?
- **Architecture PRs** — Any change to data flow, sync, or blast-sensitive edges should update `spec.edges` in the same review as HCL or playbooks.
- **Runbooks** — Encode “if X fails, touch Y and Z” as **necessary** edges so triage overlays match how your team actually fights incidents.

### Review and guardrails

- Treat **`dependencyRole` like code**: diff `omnigraph/graph/v1` in PRs; reject drive-by edits that widen **necessary** cones without justification.
- **Automated checks**: blast-radius fixtures under [testdata/graph/blast_fixtures.json](../../testdata/graph/blast_fixtures.json) and Go tests in `internal/graph` help lock traversal semantics; extend fixtures when you add critical-path shapes.

### Spotting a stale tag

Signals are **judgment-heavy**—use them as prompts, not proof:

- **Surprise during triage** — Blast radius or downstream list does not match what on-call expects from reality.
- **Evidence vs graph** — Inventory, plan, traces, or CMDB show a dependency that matters in incidents but only appears as **sufficient** (or is missing).
- **Over-stated necessity** — Observability shows no critical traffic on an edge still marked **necessary**; consider demoting to **sufficient** after review.

### Reconciliation loop

1. Compare **declared graph** to **evidence** (state, plan, inventory exports, your own diagrams).
2. Edit **`spec.edges`** in graph JSON: set `dependencyRole` to `necessary` or `sufficient` (or omit for default **necessary**).
3. Re-validate graph v1 (schema + optional CI) and re-emit if your pipeline generates topology from tools.

### OmniGraph as an active diagnostic assistant

The workspace **directs attention**: **Topology** + **Triage mode** + selection-scoped panels highlight *where* to look next instead of dumping a flat log wall. That is **assisted** diagnosis—the graph does not replace human judgment about what “critical” means for your product.

### Roadmap: classifying a newly discovered dependency in the UI

**Today**, there is no modal that asks “Is this database link **necessary** or **sufficient**?” You **edit graph JSON** (Topology editor or emitted artifact) and set `dependencyRole` on the edge.

**Intended direction (roadmap)** — When the workspace can reliably surface a **new or unclassified** dependency candidate (for example from telemetry correlation or diff against last emit), it could **prompt** the operator to classify it. Until that ships, use the reconciliation loop above.

## See also

- [Understanding the UI modes](ui-modes.md) — Topology vs Reconciliation vs Posture
- [Getting started](../getting-started.md) — Five-minute observation drill (external trigger + live sync)
- [Using the web workspace](../using-the-web.md)
- [Data handoff](../core-concepts/data-handoff.md) — provider artifacts to UI
