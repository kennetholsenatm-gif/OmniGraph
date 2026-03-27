# Architecture Decision Records & Implementation Plans

| Document | Description |
|----------|-------------|
| [ADR-001-fullstack-discovery-bom.md](ADR-001-fullstack-discovery-bom.md) | Tool selection for HBOM/SBOM/Services BOM; NetBox, NetDISCO, Syft, Trivy, Dependency-Track; CMMC/PQC/CBOM. |
| [ADR-002-termix-netbox-sync.md](ADR-002-termix-netbox-sync.md) | Termix population from NetBox; ingestion pathway, asset mapping, RBAC, decommission pruning. |
| [IMPLEMENTATION_PLAN_FULLSTACK_DISCOVERY.md](IMPLEMENTATION_PLAN_FULLSTACK_DISCOVERY.md) | Execution order and deliverables for full-stack discovery (OpenTofu, Ansible, Varlock, Doc_Agent). |
| [IMPLEMENTATION_PLAN_TERMIX.md](IMPLEMENTATION_PLAN_TERMIX.md) | Execution order and deliverables for Termix sync (sync_to_termix.py, n8n, Doc_Agent). |

Related artifacts:

- **Varlock schemas:** `fullstack-discovery.env.schema`, `termix.env.schema` (repo root).
- **Ansible:** `ansible/playbooks/deploy-fullstack-discovery.yml`.
- **Script:** `scripts/sync_to_termix.py`.
- **OpenTofu:** `discovery_net` in `opentofu/main.tf`.
