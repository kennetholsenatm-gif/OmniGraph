# IR Contracts

This page summarizes key OmniGraph schema contracts used across the web workspace, `go test`, and library consumers.

## Primary contracts

- `omnigraph/ir/v1` -> `schemas/ir.v1.schema.json`
- `omnigraph/run/v1` -> `schemas/run.v1.schema.json`
- `omnigraph/graph/v1` -> `schemas/graph.v1.schema.json`
- `omnigraph/security/v1` -> `schemas/security.v1.schema.json`
- `omnigraph/inventory-source/v1` -> `schemas/inventory-source.v1.schema.json`
- `omnigraph/integration-run/v1` -> `schemas/integration-run.v1.schema.json` (stdin for integration WASM guests)
- `omnigraph/integration-result/v1` -> `schemas/integration-result.v1.schema.json` (stdout from those guests)

## Integration runs vs IR

**IR (`omnigraph/ir/v1`)** remains **declared infrastructure intent**: components, targets, and static relations. It is not silently rewritten by integration plugins.

**Integration runs (`omnigraph/integration-run/v1`)** are a **separate** envelope the host passes on stdin to a WASM micro-container. They may carry an **`irContext`** object (labels, slices of intent, or correlation ids) for the guest to interpret, plus **`credentials`** and **`allowedFetchPrefixes`** supplied by the operator. The guest returns **`omnigraph/integration-result/v1`**, optionally embedding a normalized **`inventory-source/v1`** snapshot.

Validation helpers: [`internal/schema/validate_integration.go`](../../internal/schema/validate_integration.go).

## Versioning policy

- Keep `apiVersion` stable for non-breaking changes
- Introduce new version for breaking changes
- Preserve backward compatibility where feasible for automation consumers

## Authoring: TOML vs JSON-shaped contracts

- **Project / `.omnigraph.schema`** (`omnigraph/v1alpha1`) — **Recommended: TOML** for human-authored intent in repos and the Schema Contract tab. The Go project parser accepts **YAML** and **JSON** as well (YAML-first decode, then TOML fallback) so existing automation keeps working.
- **`omnigraph/graph/v1`** — Remains **JSON** in normal workflows: emitted by the graph emit path (CI/tests/integration), pasted into Topology, and validated by `schemas/graph.v1.schema.json`. Edge **`dependencyRole`** lives here, not in the Project file.
- **Machine artifacts** — OpenTofu/Terraform **state** and **plan JSON**, Ansible inventory, and CI outputs stay in provider-native formats; OmniGraph reads them for reconciliation and emit without asking operators to rewrite them as TOML.

## Graph v1, dependencies, and live telemetry

`omnigraph/graph/v1` ([`schemas/graph.v1.schema.json`](../../schemas/graph.v1.schema.json)) is the **declared topology** contract: nodes, edges, optional **`dependencyRole`** on each edge (**`necessary`** = critical path for blast-radius closure; **`sufficient`** = contextual / non-blocking).

- **Static intent** — The JSON document states what the automation **believes** about wiring. It is the source of truth for structure in CI, PR comments, and the **Topology** tab.
- **Conditional / dynamic modifiers** — Today, extra selectors or conditions can be carried as **edge `attributes`** or node **attributes** while keeping the core schema stable. A future revision may add first-class fields (for example `when` or telemetry selectors); until then, treat such metadata as **conventions** documented next to your graph emit pipeline.
- **Live signals** — Workspace **SSE** streams (for example `GET /api/v1/workspace/stream` from the **local workspace server**) deliver **summaries** of state-shaped inputs. They do **not** replace the graph IR; they **inform** operators and UIs when reality diverges from the last emitted snapshot.
- **Blast radius** — Plan-time orchestration can compute a **downstream closure** along **necessary** edges from Terraform/OpenTofu **`resource_changes`**, then enforce **policy limits** before apply. The web client can compute the same closure client-side for triage when graph JSON is present. **Server-side recomputation from SSE alone** remains **roadmap** unless/until the control plane emits graph deltas on the wire.

## Enclave manifests

Wasm enclave documents (`omnigraph/enclave/v1`) are validated in Go with explicit **`spec.requires`** and **`spec.provides`** lists. Use [`internal/schema.ValidateEnclaveRaw`](../../internal/schema/validate_enclave.go) (and related tests) so peer references and `peer://` / `enclave://` environment values stay inside the declared contract.

## Related docs

- [Backend Wasm plugins](../development/wasm-plugins.md) (parser vs integration plugins)
- [Architecture overview](../core-concepts/architecture.md)
- [Overview](../overview.md)
- [Using the web workspace](../using-the-web.md)
- [CI and contributor automation](../ci-and-contributor-automation.md)
- [OmniGraph IR](../core-concepts/omnigraph-ir.md)
- [Security posture](../security/posture.md)
