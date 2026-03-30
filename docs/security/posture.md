# Security posture

Policy checks and passive scans ultimately **feed the same graph model** the web workspace uses: merged **`securityPosture`** on nodes when graph emit merges a security document, and **Posture** tab JSON you can paste or produce from your own tooling. This page maps how those concerns surface in **`go test`**, the **`internal/policy`** package, the **local workspace server**, and points to ADRs—not a full organizational compliance checklist.

## Policy-as-code

- Contributor and CI gates evaluate embedded Rego in `omnigraph/policy/v1` policy sets alongside JSON Schema validation via the same libraries the product uses (see [CI and contributor automation](../ci-and-contributor-automation.md) and [`internal/policy`](../../internal/policy)).
- Example policy sets live under [`testdata/policies/`](../../testdata/policies/).

**Why the split matters:** schema validation answers “is the document well-formed for OmniGraph?”, while policy answers “does this intent violate our rules?”.

## Passive security scans

- Read-only posture modules can emit **`omnigraph/security/v1`** JSON for merging into graphs or offline review.
- Graph emit in Go can merge a security document so host nodes carry **`securityPosture`**.

Use scans only on systems you own or are explicitly permitted to assess.

## HTTP API hardening (`serve`)

Read the workspace server `-h` output before exposing anything beyond your laptop.

| Concern | Behavior |
|--------|----------|
| Bind address | Default `127.0.0.1:38671` (loopback). Widen only with intent. |
| Static UI | Optional `--web-dist` (e.g. `packages/web/dist` after `npm run build`). Without it, only `/api/v1/*` exists. |
| Experimental APIs | `POST /api/v1/security/scan`, host-ops routes, `GET /api/v1/inventory`, etc. register only when the matching `--enable-*` flags are set **and** authentication is configured. |
| Bearer token | `--auth-token` or `OMNIGRAPH_SERVE_TOKEN` for gated endpoints. |
| Metrics | `--enable-metrics` exposes Prometheus data at `/metrics`. |

Core routes include health, repo scan, and workspace summary; experimental routes are called out in the server flag help.

## WebAssembly and client-side execution

The web workspace loads **Go-built Wasm** for HCL diagnostics and related checks (see [ADR 001](../core-concepts/adr/001-wasm-linters.md)). That boundary accepts **untrusted text** from the user’s session. We treat it as a **robustness surface**: Go handlers avoid **panics** on user-controlled input, return **valid JSON** envelopes, and are exercised with **`go test -fuzz`** in the Wasm modules; TypeScript callers **catch** bridge failures and show **panel-level errors** instead of crashing the tab. Full rationale: [ADR 008 — Wasm bridge hardening](../core-concepts/adr/008-wasm-bridge-hardening.md). Operational build notes: [`wasm/README.md`](../../wasm/README.md).

## Design rationales (ADRs)

See the [ADR index](../core-concepts/adr/) for memory-only secrets, Wasm, and related decisions.

## Related reading

- [CI and contributor automation](../ci-and-contributor-automation.md)
- [Overview](../overview.md)
- [Using the web workspace](../using-the-web.md)
- [IR contracts](../schemas/ir-contracts.md)
