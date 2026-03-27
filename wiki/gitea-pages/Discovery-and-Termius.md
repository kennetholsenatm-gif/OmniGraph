# Discovery, BOM, and Termius

Pasted from `docs/WIKI_EXPORT/DISCOVERY_AND_TERMIUS_WIKI.md`. **No secrets** — names, flows, tables only.

## Chunk A — Glossary (HBOM / SBOM / services)

| Term | Meaning | Typical tool |
|------|---------|--------------|
| **SBOM** | Software Bill of Materials — components of an application/release | Syft, CycloneDX |
| **HBOM** | Hardware Bill of Materials — physical/logical device makeup | NetBox (custom fields), CMDB |
| **CBOM** | Cryptography Bill of Materials — algorithms/certs in components | CycloneDX fields, scanner policies |
| **VEX** | Vulnerability Exploitability eXchange — exploitability context | Dependency-Track, CSAF |
| **Services BOM** | Runtime service inventory + deps | NetBox services + discovery |

## Chunk B — Discovery stack components

| Component | Network | Port (host dev) | Role |
|-----------|---------|-----------------|------|
| NetBox | discovery_net | 8000 | SoT devices/VMs/IPAM |
| NetBox worker | discovery_net | — | RQ jobs |
| Dependency-Track (bundled) | discovery_net | 8081 | SBOM + vuln hub |
| NetDISCO | discovery_net (optional profile) | varies | L2/L3 SNMP discovery |

## Chunk C — Event flow (high level)

| Step | Source | Channel | Consumer | Outcome |
|------|--------|---------|----------|---------|
| 1 | Gitea | `gitea_webhook` | n8n | Pipeline triggers |
| 2 | CI runner | HTTP API | Dependency-Track | SBOM uploaded |
| 3 | Dependency-Track | Webhook | n8n → Zulip | ChatOps alert |
| 4 | NetBox | Object change | Webhook / Solace topic | Sync jobs (e.g. Termius) |
| 5 | n8n | Schedule | `sync_netbox_to_termius.py` | Termius JSON / Teams API |

Topic/queue **names** (not credentials): `docs/SOLACE_DISCOVERY_QUEUES.md`.

## Chunk D — NetBox → Termius taxonomy

| NetBox field | Termius concept |
|--------------|-----------------|
| Tenant + Site | Folder / group |
| Tags | Tags |
| Device role | Tag `role:<name>` |
| Primary IPv4 | Host address |

Pruning: **no auto-delete** by default; manual decommission in Termius until list APIs are standardized for your tenant.

## Chunk E — Compliance pointers (honest scope)

| Area | Stack contribution |
|------|--------------------|
| Inventory | NetBox + optional NetDISCO |
| Supply chain | Dep-Track + SBOM pipelines |
| Crypto visibility | Scanner + CBOM fields where available — **human review** still required |

## Chunk F — Related repo paths

| Doc / artifact | Path |
|----------------|------|
| ADR | `docs/ADR_FULLSTACK_DISCOVERY.md` |
| Discovery compose | `docker-compose/docker-compose.discovery.yml` |
| Ansible deploy | `ansible/playbooks/deploy-fullstack-discovery.yml` |
| Termius sync | `scripts/sync_netbox_to_termius.py` |
| Env schema | `devsecops.env.schema` |
