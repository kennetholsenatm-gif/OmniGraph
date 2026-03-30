# Integrations

Telemetry, inventory, identity, and secrets integrations **enrich what operators see** in graph and workspace contexts (and what automation can emit), not only “backend plumbing.”

## WASM micro-containers for external systems

Connections to **external HTTP APIs** (for example **NetBox** for IPAM/CMDB-style inventory, **Zabbix** for monitoring inventory) are implemented as **Class B** backend WASI plugins—see [Backend Wasm plugins](../development/wasm-plugins.md). The Go **core** does not embed vendor-specific HTTP clients for those systems; it loads `.wasm` artifacts and delegates API paths, query shapes, and response mapping to the guest. Egress is **only** through the host’s **`omnigraph.http_fetch`** import with **per-run URL prefix allowlists**.

This keeps **blast radius** bounded: a defect in one integration cannot silently widen network access beyond the prefixes the operator configured for that run, and integration logic stays **replaceable** without recompiling the workspace server.

## Integration categories

- Secret backends (runtime retrieval, no committed credentials)
- Inventory and CMDB sources (prefer **inventory-source/v1** snapshots produced by WASM integrations)
- Telemetry enrichment for graph and run context
- Identity and authorization providers

## Provider neutrality

Provider names in examples are illustrative. Teams can map OmniGraph workflows to their selected stack (for example GitHub, GitLab, Gitea, or self-hosted CI).

Use placeholders in examples:

- `https://git.example.com/<org>/<repo>`
- `https://id.example.com`
- `https://inventory.example.com/api`

## Operational entrypoints

- **Workspace server (primary):** **`POST /api/v1/integrations/run`** when **`--enable-integration-run-api`** is set (requires auth like other experimental APIs). The body supplies a workspace-relative **`wasmPath`** and a **`run`** object validating as **`omnigraph/integration-run/v1`**.

See also [Inventory sources](inventory-sources.md) and [IR contracts reference](../schemas/ir-contracts.md).
