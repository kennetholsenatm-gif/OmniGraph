# OmniGraph system architecture

## High-level overview

OmniGraph is a state-aware, multi-paradigm infrastructure orchestration platform. Unlike traditional CI/CD pipelines that treat infrastructure deployments as isolated scripts, OmniGraph acts as a continuous state engine. It bridges the gap between declarative provisioning (OpenTofu/Terraform), imperative configuration (Ansible), and real-time environment telemetry (NetBox/Zabbix) through a unified GitOps workflow.

## Core architectural layers

### Layer 1: Presentation and shift-left IDE (client-side)

The frontend is designed to run almost entirely in the browser, reducing server-side compute and providing fast feedback to developers.

- **Framework:** React + Tailwind CSS (see `web/`).
- **Visualizer engine:** **React Flow** with **Dagre** layout renders `omnigraph/graph/v1` (`schemas/graph.v1.schema.json`, `web/src/graph/`). Custom node components provide SVG-accented styling; a parallel D3 canvas remains optional if product needs layouts React Flow cannot express.
- **Wasm execution context:** **HCL parse diagnostics** ship as Wasm (`wasm/hcldiag`, `web/public/wasm/hcldiag.wasm`, `web/src/hclWasm.ts`). The same binary exports **structure heuristics** (`omnigraphHclStructureLint`) and a **plaintext pattern scan** (`omnigraphTfPatternLint`, rules in `wasm/tfpattern`) as spikes toward tflint/checkov-style coverage. Full **tflint / checkov / ansible-lint** engines in the browser remain roadmap (see `wasm/README.md`, [ADR 001](adr/001-wasm-linters.md)).
- **Real-time contract validation:** `.omnigraph.schema` validated locally in the browser as the user types (JSON Schema in `schemas/`).
- **Repository-wide view:** The Inventory tab can scan an entire checkout (File System Access API in Chromium) by classifying IaC paths the same way as `omnigraph repo scan` in Go (`internal/repo`), then aggregating hosts from every discovered state and inventory file—not a single hand-picked artifact.

### Layer 2: Control plane (the brain)

A lightweight Go binary (`cmd/omnigraph`) that can run as a GitHub Action, a GitLab CI job, or a standalone container. It orchestrates tools rather than replacing them.

- **Schema coercion engine:** Reads `.omnigraph.schema` and produces in-memory `terraform.tfvars.json`, `group_vars/all.yml`, and container `.env` representations (`internal/coerce`, `omnigraph coerce`).
- **State interceptor:** Parses OpenTofu/Terraform JSON state after apply; extracts outputs for downstream tools (`internal/state`, `omnigraph state …`).
- **Dynamic inventory generator:** Builds an ephemeral Ansible inventory from state (`internal/inventory`), used by `internal/orchestrate`.
- **Orchestrated pipeline:** Validate → coerce → plan → inventory/check → apply → Ansible (`omnigraph orchestrate`, `internal/orchestrate`).
- **Native infrastructure IR:** Versioned **`omnigraph/ir/v1`** intent documents (`schemas/ir.v1.schema.json`, `internal/ir`, `omnigraph ir validate|formats`) describe targets, abstract components, and relations; **multi-format backends** emit OpenTofu, Ansible, Helm, Kubernetes, Packer, Compose, CloudFormation, Puppet, and Pulumi-family artifacts incrementally ([ADR 005](adr/005-native-ir-enterprise-identity.md), [omnigraph-ir.md](omnigraph-ir.md)). **Enterprise identity:** `internal/identity` defines RBAC permissions and mapping hooks for **Keycloak (OIDC)** and **FreeIPA (LDAP)**—serve and APIs migrate from shared bearer tokens to IdP-backed `Authorizer` over time.
- **Repository discovery:** `omnigraph repo scan --path <dir>` walks a working tree (skipping `.git`, `node_modules`, `.terraform`, etc.) and emits JSON listing Terraform state, HCL, Ansible, and `.omnigraph.schema` paths—machine input for tooling and parity with the web scanner.

### Layer 3: Execution matrix (the runners)

Pluggable execution: host `os/exec` or ephemeral containers (**Docker/Podman**) via `ContainerRunner`. See [execution-matrix.md](execution-matrix.md).

- **Implemented:** `ExecRunner`, `ContainerRunner`, `omnigraph orchestrate --runner=exec|container`.
- **Roadmap:** Firecracker microVMs; additional engines (e.g. Pulumi) behind explicit `--iac-engine` hooks where not yet implemented.

### Layer 4: Data, telemetry, and security (integrations)

- **Zero-disk secret engine:** Vault, AWS Secrets Manager, or SOPS; fetch via OIDC/JWT; inject into runner memory; **stdout/stderr redaction** uses env-derived secret values in runners ([ADR 003](adr/003-memory-only-secrets.md)).
- **CMDB ingestion:** NetBox (versioned webhook payload in `internal/netbox`, [integrations.md](integrations.md)), Zabbix/Prometheus — **telemetry nodes** (`kind: telemetry`, `state: gray`) merge into graph documents (`internal/telemetry`, `omnigraph graph emit --telemetry-file`). **Triangulated inventory** snapshots and merge rules: [inventory-sources.md](inventory-sources.md), schema `omnigraph/inventory-source/v1` in [`schemas/inventory-source.v1.schema.json`](../schemas/inventory-source.v1.schema.json).
- **Posture scans:** First-party passive checks emit **omnigraph/security/v1** (`schemas/security.v1.schema.json`, `internal/security`, `omnigraph security scan`). Summaries can merge into graph host nodes as `attributes.securityPosture` via `omnigraph graph emit --security-file` (`internal/graph.MergeSecurity`). Optional **serve** APIs (`--enable-security-scan`, `--enable-host-ops`, `--enable-inventory-api`) run only with a **Bearer token** and default to loopback; host-ops uses SSH for systemd/journal (and optional restarts when `--host-ops-allow-writes` is set); inventory API exposes aggregated state hosts (`GET /api/v1/inventory`); privileged calls are recorded in an in-memory audit ring (`GET /api/v1/audit`).

## Lifecycle flow (deployment handoff)

1. **Trigger:** User opens a PR; control plane runs in CI or locally.
2. **Phase 1 — Validation:** Parse `.omnigraph.schema`; fail on type errors.
3. **Phase 2 — Plan:** `tofu plan -out=tfplan`; parse plan for projected resources; `ansible-playbook --check` against projected inventory.
4. **Phase 3 — Visualization:** Merge plan/check results (and optional telemetry fixtures) into a JSON graph for PR/UI.
5. **Phase 4 — Apply and handoff:** On approval, `tofu apply`; intercept new `.tfstate`; map outputs into Ansible context; run `ansible-playbook` against live targets.
6. **Phase 5 — Sync:** Webhook to NetBox (and similar) using **omnigraph/netbox-sync/v1** payload shape (see [integrations.md](integrations.md)).

## Architecture decision records

| ADR | Topic |
|-----|--------|
| [001](adr/001-wasm-linters.md) | WebAssembly for linters in the IDE |
| [002](adr/002-schema-first-config.md) | Schema-first `.omnigraph.schema` |
| [003](adr/003-memory-only-secrets.md) | Memory-only secret injection |
| [004](adr/004-unified-state-locking.md) | Unified state locking (`omnigraph/lock/v1`) |
| [005](adr/005-native-ir-enterprise-identity.md) | Native IR, multi-format backends, Keycloak/FreeIPA RBAC |
