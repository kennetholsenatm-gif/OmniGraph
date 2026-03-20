# Moving the Docker containers to OpenNebula (non-negotiable scope)

The **DevSecOps pipeline is mostly Docker Compose**. When you leave a Windows Gitea box, you are not done until **the same containerized stacks** run on your new Linux estate.

**Recommended on OpenNebula:** **AlmaLinux 10 LXC** per stack with **native systemd services or Podman** first; **Docker** only where no better option — **[REDUCE-DOCKER.md](REDUCE-DOCKER.md)**. Layout and instance names: **[LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md)**.

**Transitional:** Docker Compose inside LXC (repo default automation) until each service is ported off `docker-ce`.

**Gitea data** is one of many state directories (DB files, object stores, Vault raft, etc.).

**Sources of truth for what runs:**

- [docker-compose/stack-manifest.json](../../docker-compose/stack-manifest.json) — **core** + **optional** merge lists
- [ansible/roles/devsecops_containers/README.md](../../ansible/roles/devsecops_containers/README.md) — stacks the **Ansible role** starts (IAM, messaging, tooling, ChatOps, gateway, optionals)
- [docs/DEPLOYMENT.md](../DEPLOYMENT.md) — order of operations

## What “the containers” are (core)

| Stack | Compose file(s) | Docker networks (examples) |
|-------|-------------------|----------------------------|
| **IAM** | `docker-compose/docker-compose.iam.yml` | `iam_net` (100.64.20.0/24) — Vault, Keycloak, proxy |
| **Messaging** | `docker-compose/docker-compose.messaging.yml` | `msg_backbone_net` (100.64.10.0/24); Solace, Kafka, NiFi, Postgres, … |
| **Tooling** | `docker-compose/docker-compose.tooling.yml` | **`gitea_net`** (100.64.1.0/24), `n8n_net`, `zammad_net`, `bitwarden_net`, `portainer_net` |
| **ChatOps** | `docker-compose/docker-compose.chatops.yml` | `chatops_net` (100.64.8.0/24) |
| **Single pane / gateway** | `single-pane-of-glass/docker-compose.yml` | **`gateway_net`** (100.64.5.0/24); Traefik, webhook listener, dashboard |

**Optional stacks** (same repo; toggle via manifest / Ansible / env): discovery, LLM, AI orchestration, FreeIPA, SIEM, SDN telemetry — each has additional compose files and networks per [NETWORK_DESIGN.md](../NETWORK_DESIGN.md).

On OpenNebula, each `*_net` maps to a **VNET** row in [VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md). Your **Docker host VM** needs vNICs (or routed paths via VR) so containers keep the **same subnets** and isolation rules.

## What must physically move

| Asset | What to do |
|-----|------------|
| **Container images** | Re-pull on target (`docker compose pull`) or save/load `docker save` if air-gapped |
| **Named Docker volumes** | **Backup and restore** per service (Postgres, Gitea data, n8n, Zammad, Solace, etc.) — see below |
| **Bind mounts** | Any host paths in compose (e.g. old `C:\` Gitea) → **Linux paths** on VM or drop in favor of named volumes |
| **Bridge networks** | Recreate on each Docker host: [scripts/create-networks.ps1](../../scripts/create-networks.ps1) / [.sh](../../scripts/create-networks.sh) or OpenTofu `opentofu/` |
| **Runtime env** | **No committed `.env`** — reproduce via **Vault** + [scripts/secrets-bootstrap.ps1](../../scripts/secrets-bootstrap.ps1) / Ansible `devsecops_containers` |
| **Compose definitions** | Already in **git** — clone repo on Ansible controller / VM; pin same revision as source |

## Target topology (recommended)

1. **OpenNebula KVM guest** running **AlmaLinux 10** + **LXD/Incus**; provision **`devsecops-*`** LXCs via [`deploy-devsecops-lxc.yml`](../../ansible/playbooks/deploy-devsecops-lxc.yml) — details in **[LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md)**.
2. Attach the **LXD host** VM vNICs to OpenNebula VNETs matching [**VLAN_MATRIX**](../../deployments/opennebula-kvm/VLAN_MATRIX.md) / [**NETWORK_DESIGN**](../NETWORK_DESIGN.md) (edge + routed `100.64` paths via VR/ISR as designed).
3. **Fallback:** **One or more Linux VMs** with only **Docker CE** + **Compose v2** (no LXC) — same volume/network migration steps, weaker isolation.
4. **Split KVM VMs** (IAM VM + messaging VM + …) only if you refuse LXD on a single guest — heavier operationally.

**K3s:** If Gitea runs **in Kubernetes**, you still need a plan for **n8n, Zammad, Vault, Keycloak, Solace…** — either they stay **Docker on VMs** (hybrid) or you port them to charts (out of scope unless you commit to full K8s).

## Volume migration (pattern)

**Offline maintenance window** (parallel to Gitea freeze):

1. On **source** Docker host:  
   `docker compose -f <file> stop` (per stack or whole merge) to quiesce writers.
2. For each **named volume** backing stateful services:  
   - Option A: `docker run --rm -v VOL:/v -v $(pwd):/backup alpine tar czf /backup/VOL.tgz -C /v .`  
   - Option B: DB-native dump (Postgres `pg_dump`, etc.) where compose already documents it.
3. Transfer archives to **target** host (rsync/scp).
4. On **target**: create networks → create empty volumes if needed → extract tar into volume **or** `docker compose up` once and replace data with restore.
5. `docker compose up -d` in **DEPLOYMENT.md** order: networks → IAM (Vault) → messaging → tooling → chatops → gateway — or use **Ansible** to do the same with new inventory.

**Gitea in tooling compose:** same volume backup as any other container; **or** use `gitea dump` from [05-migration-runbook.md](05-migration-runbook.md) if that is your source of truth.

## Ansible / Vault replay

After VMs exist:

- Point **inventory** `ansible_host` at new Docker VM(s).
- Ensure **Vault** reachable from controller; restore Vault **Raft/snapshot** if Vault moved, or re-bootstrap secrets and rotate.
- Run [playbooks that start containers](../../ansible/playbooks/) (e.g. `start-containers-with-vault.yml`, site playbooks per [devsecops_containers README](../../ansible/roles/devsecops_containers/README.md)).

## Checklist — containers actually moved

- [ ] **IAM** stack up; **Vault** unsealed; **Keycloak** reachable
- [ ] **Messaging** stack up; brokers sane (Solace/Kafka as you use)
- [ ] **Tooling** up; **Gitea**/`n8n`/`Zammad`/etc. pass smoke tests
- [ ] **ChatOps** up if you use it
- [ ] **single-pane-of-glass** up; Traefik routes + **`/webhook/docs-sync`** tested
- [ ] All **`100.64.*` docker networks** exist on host (`docker network ls`) with correct subnets
- [ ] **No service** still pointed at old Windows host IP for **inter-container DNS** (update secrets / dependent URLs)

## Related

- Full artifact list + greps: [WHOLE-REPO-MIGRATION-SCOPE.md](WHOLE-REPO-MIGRATION-SCOPE.md)
- Network numbering: [NETWORK_DESIGN.md](../NETWORK_DESIGN.md), [VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md)
- Gitea-only data path: [05-migration-runbook.md](05-migration-runbook.md)
