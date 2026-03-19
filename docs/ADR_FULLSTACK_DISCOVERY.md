# ADR: Full-stack discovery and BOM pipeline

## Status

Accepted — dev/greenfield baseline; production hardening is incremental.

## Context

Operators need a **source of truth** for infrastructure (devices, IPs, sites), **L2/L3 discovery** to reduce drift, and **SBOM + vulnerability aggregation** with APIs suitable for n8n and SAM. All services must sit on an isolated Docker segment (**discovery_net**, 100.64.40.0/24) consistent with [NETWORK_DESIGN.md](NETWORK_DESIGN.md).

## Decision

| Layer | Choice | Rationale |
|-------|--------|-----------|
| SoT | **NetBox** | Mature DCIM/IPAM API, webhooks, multi-tenant tags; industry default for automation. |
| Discovery | **NetDISCO** | SNMP/LLDP-based L2/L3; suitable for brownfield. **Note:** upstream container images vary; enable via Compose **profile** `netdisco` or run documented `docker run` jobs if image maintenance blocks CI. |
| SBOM | **Syft** + **Trivy** | Generate and scan locally/CI; push results to Dependency-Track via API. |
| Vuln / BOM hub | **Dependency-Track** | OWASP project; aggregates SBOMs, policy gates, API for ChatOps (e.g. Zulip). |
| Messaging contract | **Solace** (infra repo) | Queue/topic **names** for BOM ingestion are documented in [SOLACE_DISCOVERY_QUEUES.md](SOLACE_DISCOVERY_QUEUES.md); broker objects remain in the infra OpenTofu repo. |

## Non-goals

- Running Solace or production HA databases **inside** this compose file (only clients on `discovery_net` where needed).
- Fully automated CMMC audit evidence — this ADR provides **mapping hooks** and honest scope (documentation + schema), not fake compliance automation.

## CMMC / SA mapping (appendix)

| Practice (illustrative) | How this stack helps |
|-------------------------|----------------------|
| CM-2 / CM-3 (baseline / change) | NetBox as SoT; webhook → n8n for change awareness. |
| CM-8 (inventory) | NetBox devices/VMs; NetDISCO for discovered vs declared reconciliation. |
| SA-10 / SA-11 (dev practices / testing) | Dep-Track + CI SBOM upload; Trivy gates in pipeline (existing n8n/Gitea flows). |
| SA-15 (supply chain) | SBOM in Dep-Track; **CBOM** fields (where supported) align with legacy-crypto visibility — operators still review findings manually. |

## PQC / “legacy crypto” narrative

Dependency-Track and CycloneDX can carry component metadata; **flagging weak algorithms** is a **policy and scanner** concern. This repo supplies **schema keys** and **documentation** so Varlock/Vault can hold scanner API tokens — not automated PQC attestation.

## Taxonomy (NetBox → downstream)

See [NETBOX_TERMIUS_SYNC.md](NETBOX_TERMIUS_SYNC.md) for folder/tag mapping to Termius and pruning rules.

## References

- [docker-compose/docker-compose.discovery.yml](../docker-compose/docker-compose.discovery.yml)
- [ansible/playbooks/deploy-fullstack-discovery.yml](../ansible/playbooks/deploy-fullstack-discovery.yml)
- [devsecops.env.schema](../devsecops.env.schema) — `@env-spec: DISCOVERY_*`, `NETBOX_*`, `DEP_TRACK_*`, `NETDISCO_*`
- [SOLACE_DISCOVERY_QUEUES.md](SOLACE_DISCOVERY_QUEUES.md)
- [docs/snippets/solace-discovery-queues.tf](snippets/solace-discovery-queues.tf)
