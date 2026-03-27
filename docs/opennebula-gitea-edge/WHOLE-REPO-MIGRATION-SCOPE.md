# Whole-repo migration scope (Gitea + DevSecOps pipeline)

This document describes **everything in the `devsecops-pipeline` repository and its runtime** that must move or be **re-pointed** when **Gitea** leaves a Windows workstation (`C:\GiTeaRepos`) for an OpenNebula-backed Linux instance.

**The Docker containers are the product.** IAM, messaging, tooling (Gitea/n8n/Zammad/…), ChatOps, and the single-pane gateway are **Compose stacks** on Linux. **Preferred on OpenNebula:** **AlmaLinux 10 LXC** per stack (**Docker-in-LXC**) — [LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md). **Volume/network/runtime migration:** [CONTAINER-LIFT-TO-OPENNEBULA.md](CONTAINER-LIFT-TO-OPENNEBULA.md). Hardware: [01-node-roles-and-placement.md](01-node-roles-and-placement.md).

**Canonical clone/working tree:** [README.md](../../README.md) — **`C:\GiTeaRepos\devsecops-pipeline`** (or your Linux equivalent after migration).

## What “moves” vs what you “redeploy”

| Category | Moves (data/images/config) | On new OpenNebula/Linux infrastructure |
|----------|----------------------------|----------------------------------------|
| **All Compose workloads** | **Named volumes**, DB files, broker state, TLS assets, bind-mount data | **LXC path:** one **`devsecops-*`** Alma 10 LXC per stack, nested Docker, same compose files — [LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md). **Flat path:** one Docker VM. Re-inject env from **Vault/Ansible** — [CONTAINER-LIFT-TO-OPENNEBULA.md](CONTAINER-LIFT-TO-OPENNEBULA.md) |
| **Git history & Gitea DB** | `gitea dump` and/or Gitea **Docker volume** backup | Linux paths, `ROOT_URL`, TLS; may be **tooling** container **or** K3s Helm |
| **Container images** | Optional `docker save` / registries | `docker compose pull` on target |
| **This git repo (clone)** | Already in Gitea | Developer **git remote** + CI |
| **Vault** | **Raft snapshot** / operational backup of KV | Restore or re-bootstrap; [ansible/roles/devsecops_containers](../../ansible/roles/devsecops_containers/README.md) |
| **Ansible inventory** | In repo | New VM IPs; `ansible_host` |
| **n8n workflows JSON** | In repo; **execution DB** in volume | Import + restore **n8n** volume |
| **OpenTofu state** | Backup if used | Re-apply Docker networks on new hosts |

## Repository artifact map (verify after cutover)

Use this as a checklist that **no integration still assumes Windows-only Gitea**.

### Core docs and schemas

| Path | Why it matters |
|------|----------------|
| [docs/NETWORK_DESIGN.md](../NETWORK_DESIGN.md) | **`100.64.x` segments** — VMs/containers must match after lift |
| [docs/DEPLOYMENT.md](../DEPLOYMENT.md) | Order: networks → Vault → compose → Ansible |
| [devsecops.env.schema](../../devsecops.env.schema) | **`GITEA_*`**, **`DOCS_*`** — defaults still mention `C:/GiTeaRepos`; override in Vault/env on Linux |
| [identities.example.yaml](../../identities.example.yaml), [identities.yaml](../../identities.yaml) | Service accounts tied to IAM; not Gitea disk but SSO may use Gitea OAuth |
| [docker-compose/stack-manifest.json](../../docker-compose/stack-manifest.json) | Core + optional stacks touching Gitea webhooks |

### Docker Compose / runtime

| Area | Gitea-related touchpoints |
|------|---------------------------|
| `docker-compose/docker-compose.tooling.yml` | Gitea container, `gitea_net` |
| `docker-compose/*` | n8n, Zammad, agents — **webhooks** to Gitea |
| `single-pane-of-glass/` | Traefik, **`/webhook/docs-sync`**, [DOCSIFY_GITEA.md](../DOCSIFY_GITEA.md) |
| `docker-compose/docker-compose.chatops.yml` | ChatOps may reference repo URLs |

### Automation

| Path | Action after Gitea move |
|------|-------------------------|
| `ansible/` | **`devsecops_containers`** role: env injection; inventory hostnames; playbooks that curl Gitea API |
| `ansible/inventory/*` | Update any **`ansible_host`** if Gitea or gateway IP changes |
| `scripts/secrets-bootstrap.ps1` / `.sh` | Writes Vault KV; ensure **`GITEA_URL`** / tokens match new endpoint |
| `scripts/create-networks.ps1` / `.sh` | Run on **each** new Docker host or VM set |

### Integrations and workflows

| Path | Action |
|------|--------|
| [n8n-workflows/](../../n8n-workflows/) | Import/repair workflows; Gitea trigger nodes; **base URL** + credentials |
| [ai-orchestration/](../../ai-orchestration/) | Any Gitea API usage in agents |
| `opentofu/` | Network definitions — align IPs with OpenNebula VNETs if bridging |

### Deployment manifests (Kubernetes track)

| Path | Action |
|------|--------|
| [deployments/opennebula-gitea/](../../deployments/opennebula-gitea/) | Helm, CSI, [BRINGUP.md](../../deployments/opennebula-gitea/BRINGUP.md) |
| [deployments/opennebula-kvm/onevnet/](../../deployments/opennebula-kvm/onevnet/) | VNETs for segments mapping to **`gitea_net`** / **`gateway_net`** |

## Logical segments to preserve (from NETWORK_DESIGN)

When **Gitea** is on OpenNebula, keep the same **isolation model** (no ad-hoc bridging of tooling nets):

| Docker network | IPv4 | OpenNebula VNET (matrix) |
|----------------|------|---------------------------|
| `gitea_net` | 100.64.1.0/24 | `devsecops-gitea` (2001) |
| `gateway_net` | 100.64.5.0/24 | `devsecops-gateway` (2005) |
| `n8n_net`, `zammad_net`, … | per [NETWORK_DESIGN.md](../NETWORK_DESIGN.md) | matching rows in [VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md) |

If you only migrate **Gitea** first, **webhooks** from n8n/CI must still **reach** the new Gitea IP/DNS; other stacks may stay on Docker until phase 2.

## Environment and secrets (Varlock / Vault)

Search and update (example patterns):

- `GITEA_URL`, `GITEA_REPO_URL`, `DOCS_GIT_REPO`
- `GITEA_DATA_PATH`, `GITEA_REPOS_ROOT` — **Linux paths** or remove host bind-mounts if Gitea is fully in-cluster PVCs
- `GITEA_DOCS_WEBHOOK_SECRET` and gateway `GITEA_DOCS_WEBHOOK_SECRET` alignment
- Any **API tokens** stored in Vault at `secret/devsecops` (or your mount)

**Project scanners:** `grep -r "gitea:3000"`, `grep -r "GiTeaRepos"`, `grep -r "git\\.example"` from repo root after cutover.

## CI / developer machines

- [ ] All **`git remote`** URLs updated to new **`git.<domain>`**
- [ ] **SSH** `known_hosts` if host keys rotated
- [ ] **Personal access tokens** / deploy keys recreated if DB restore invalidates
- [ ] **No Windows-only paths** in pipeline YAML, hooks, or README copy-paste blocks

## Tied documents

- **Alma 10 LXC on OpenNebula:** [LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md)
- **Compose / volumes / stacks:** [CONTAINER-LIFT-TO-OPENNEBULA.md](CONTAINER-LIFT-TO-OPENNEBULA.md)
- Data migration steps: [05-migration-runbook.md](05-migration-runbook.md)
- Docsify + webhooks: [DOCSIFY-POST-MIGRATION-CHECKLIST.md](DOCSIFY-POST-MIGRATION-CHECKLIST.md)
- Phased execution: [REFINED-EXECUTION.md](REFINED-EXECUTION.md)
