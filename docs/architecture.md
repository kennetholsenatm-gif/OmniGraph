# OmniGraph system architecture

## High-level overview

OmniGraph is a state-aware, multi-paradigm DevSecOps orchestration platform. Unlike traditional CI/CD pipelines that treat infrastructure deployments as isolated scripts, OmniGraph acts as a continuous state engine. It bridges the gap between declarative provisioning (OpenTofu/Terraform), imperative configuration (Ansible), and real-time environment telemetry (NetBox/Zabbix) through a unified GitOps workflow.

## Core architectural layers

### Layer 1: Presentation and shift-left IDE (client-side)

The frontend is designed to run almost entirely in the browser, reducing server-side compute and providing fast feedback to developers.

- **Framework:** React + Tailwind CSS (see `web/`).
- **Visualizer engine:** Custom SVG/D3.js rendering for the dynamic dependency graph (planned).
- **Wasm execution context:** Industry-standard tools compiled to WebAssembly for local linting and scanning without round-trips to a backend (roadmap; see `wasm/README.md` and [ADR 001](adr/001-wasm-linters.md)).
- **Real-time contract validation:** `.omnigraph.schema` validated locally in the browser as the user types (JSON Schema in `schemas/`).

### Layer 2: Control plane (the brain)

A lightweight Go binary (`cmd/omnigraph`) that can run as a GitHub Action, a GitLab CI job, or a standalone container. It orchestrates tools rather than replacing them.

- **Schema coercion engine:** Reads `.omnigraph.schema` and generates `terraform.tfvars.json`, `group_vars/all.yml`, and container `.env` **representations** in memory (implementation planned).
- **State interceptor:** Parses OpenTofu/Terraform `.tfstate` after apply; extracts outputs (IPs, keys, etc.) (planned).
- **Dynamic inventory generator:** Builds an ephemeral Ansible inventory from intercepted state (planned).

### Layer 3: Execution matrix (the runners)

Pluggable execution: ephemeral sandboxes (Docker/Podman or Firecracker microVMs) for specific toolchains. See [execution-matrix.md](execution-matrix.md).

### Layer 4: Data, telemetry, and security (integrations)

- **Zero-disk secret engine:** Vault, AWS Secrets Manager, or SOPS; fetch via OIDC/JWT; inject into runner memory; mask in logs. See [ADR 003](adr/003-memory-only-secrets.md).
- **CMDB ingestion:** NetBox (IPAM, intended state), Zabbix/Prometheus (health/metrics) — populates contextual nodes in the visualizer. See [integrations.md](integrations.md).

## Lifecycle flow (deployment handoff)

1. **Trigger:** User opens a PR; control plane runs in CI or locally.
2. **Phase 1 — Validation:** Parse `.omnigraph.schema`; fail on type errors.
3. **Phase 2 — Plan:** `tofu plan -out=tfplan`; parse plan for projected resources; `ansible-playbook --check` against projected inventory.
4. **Phase 3 — Visualization:** Merge plan/check results into a JSON graph for PR/UI.
5. **Phase 4 — Apply and handoff:** On approval, `tofu apply`; **intercept** new `.tfstate`; map outputs (e.g. `aws_instance.web.public_ip`) into Ansible context; run `ansible-playbook` against live targets.
6. **Phase 5 — Sync:** Webhook to NetBox (and similar), e.g. `{"action": "create", "ip": "10.0.5.21", "role": "web-server"}`.

## Architecture decision records

| ADR | Topic |
|-----|--------|
| [001](adr/001-wasm-linters.md) | WebAssembly for linters in the IDE |
| [002](adr/002-schema-first-config.md) | Schema-first `.omnigraph.schema` |
| [003](adr/003-memory-only-secrets.md) | Memory-only secret injection |
