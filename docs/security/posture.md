# Security posture

Policy checks and passive scans ultimately **feed the same graph model** the web workspace uses: merged **`securityPosture`** on nodes after `graph emit --security-file`, and **Posture** tab JSON you can paste or generate from **`omnigraph security`**. This page maps how those concerns surface in the **CLI** and in **`serve`**, and points to ADRs—not a full organizational compliance checklist.

## Policy-as-code

- **`omnigraph validate --policy-dir …`** evaluates embedded Rego in `omnigraph/policy/v1` policy sets alongside JSON Schema validation. Add **`--enforce`** to fail the process when denied violations are reported.
- **`omnigraph policy`** exposes `check`, `dry-run`, `report`, `list`, and `validate` for working with policy directories and inputs (see `omnigraph policy --help`).

Example policy sets live under [`testdata/policies/`](../../testdata/policies/).

**Why the split matters:** schema validation answers “is the document well-formed for OmniGraph?”, while policy answers “does this intent violate our rules?”.

## Passive security scans

- **`omnigraph security`** runs read-only posture modules and emits **`omnigraph/security/v1`** JSON.
- **`omnigraph graph emit --security-file …`** merges that document so host nodes carry **`securityPosture`** in the graph.

The CLI describes these scans as **authorized validation only**—use them on systems you own or are explicitly permitted to assess (`omnigraph security --help`).

## HTTP API hardening (`serve`)

Read `omnigraph serve --long` before exposing anything beyond your laptop.

| Concern | Behavior |
|--------|----------|
| Bind address | Default `127.0.0.1:38671` (loopback). Widen only with intent. |
| Static UI | Optional `--web-dist` (e.g. `packages/web/dist` after `npm run build`). Without it, only `/api/v1/*` exists. |
| Experimental APIs | `POST /api/v1/security/scan`, host-ops routes, `GET /api/v1/inventory`, etc. register only when the matching `--enable-*` flags are set **and** authentication is configured. |
| Bearer token | `--auth-token` or `OMNIGRAPH_SERVE_TOKEN` for gated endpoints. |
| Metrics | `--enable-metrics` exposes Prometheus data at `/metrics`. |

Core routes documented in the `serve` long help include health, repo scan, and workspace summary; experimental routes are called out there as well.

## WebAssembly and client-side execution

The web workspace loads **Go-built Wasm** for HCL diagnostics and related checks (see [ADR 001](../core-concepts/adr/001-wasm-linters.md)). That boundary accepts **untrusted text** from the user’s session. We treat it as a **robustness surface**: Go handlers avoid **panics** on user-controlled input, return **valid JSON** envelopes, and are exercised with **`go test -fuzz`** in the Wasm modules; TypeScript callers **catch** bridge failures and show **panel-level errors** instead of crashing the tab. Full rationale: [ADR 008 — Wasm bridge hardening](../core-concepts/adr/008-wasm-bridge-hardening.md). Operational build notes: [`wasm/README.md`](../../wasm/README.md).

## Design rationales (ADRs)

These records explain **why** core behavior is shaped the way it is:

- [ADR 001 — WASM linters](../core-concepts/adr/001-wasm-linters.md)
- [ADR 008 — Wasm bridge hardening](../core-concepts/adr/008-wasm-bridge-hardening.md)
- [ADR 002 — Schema-first configuration](../core-concepts/adr/002-schema-first-config.md)
- [ADR 003 — Memory-only secrets](../core-concepts/adr/003-memory-only-secrets.md)
- [ADR 004 — Unified state locking](../core-concepts/adr/004-unified-state-locking.md)
- [ADR 005 — Native IR enterprise identity](../core-concepts/adr/005-native-ir-enterprise-identity.md)

## Related reading

- [Using the web workspace](../using-the-web.md) — Posture tab and graph context
- [CLI and CI](../cli-and-ci.md) — validate with policy, security scan, serve
- [Overview](../overview.md) — system context
- [State management](../core-concepts/state-management.md) — locks and run metadata
