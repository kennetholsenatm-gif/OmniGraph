# Canonical deployment vision

This document states **site-level non-negotiables** for how this repository is meant to be deployed: physical roles, routing ownership, and what must exist before automation is trustworthy. It complements [NETWORK_DESIGN.md](NETWORK_DESIGN.md) (logical `100.64.x` segments) and [deployments/opennebula-kvm/VLAN_MATRIX.md](../deployments/opennebula-kvm/VLAN_MATRIX.md) (VLANs and platform carve).

For **phased execution order**, see [ROADMAP.md](ROADMAP.md).

**Network simplification (optional):** If you do not want **Vault, Keycloak, LDAP, RADIUS, and PacketFence** isolated from *each other*, treat them as one **identity / NAC control-plane zone** (routing and firewall policy), including **PacketFence on Mini-PC-Network** talking to IdM on **Mini-PC-IAM** without over-segmenting east–west. See [NETWORK_COLLAPSED_IDENTITY_PLANE.md](NETWORK_COLLAPSED_IDENTITY_PLANE.md).

## Physical roles (canonical)

| Role | Placement | Purpose |
|------|-----------|---------|
| **Edge firewall / router** | Dedicated **mini PC** **outside** OpenNebula | **VyOS** (Incus LXC), optional PacketFence/RatTrap per [opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md](opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md). **Single PAT** to the ISP on this path when this profile is active. |
| **IAM + secrets** | **Second** dedicated **mini PC** (not the edge box) | **HashiCorp Vault**, **Keycloak**, and related IAM compose services on logical **`iam_net` (`100.64.20.0/24`)** so **Semaphore, Ansible, n8n, and CI** can use **Varlock / OIDC** before the full OpenNebula lift is complete. |
| **OpenNebula + workloads** | UCS / hypervisors **downstream** of lab routing | Runtime for **LXC + Docker** stacks per [opennebula-gitea-edge/LXC-ALMA10-OPENNEBULA.md](opennebula-gitea-edge/LXC-ALMA10-OPENNEBULA.md); **installing** the OpenNebula control plane itself is **not** automated in this repo (see [REPO_SCOPE.md](REPO_SCOPE.md)). |
| **ISR** | Lab **L3 hub** for UCS / platform carve | Peers with **VyOS** on a **dedicated transit VLAN** (not RatTrap 900/901). Does **not** own the **Google Home** routing domain in the **canonical** profile (see below). |

**Important:** The **OpenNebula front-end mini PC** (if used) is **not** the same device as the **edge VyOS mini PC**. Do not conflate them in runbooks; see [opennebula-gitea-edge/REFINED-EXECUTION.md](opennebula-gitea-edge/REFINED-EXECUTION.md).

## Google Home / `100.64.244.0/24` (canonical profile)

**Canonical (this site):**

- **Attachment:** Google Home (Wi‑Fi / mesh) uses a **dedicated VLAN** terminated on the **edge VyOS** (firewall mini PC), **not** on the ISR.
- **Default gateway:** **VyOS** is the **default gateway** for that VLAN.
- **Reachability:** VyOS **advertises** the **`100.64.244.0/24`** prefix to the rest of the site (e.g. **OSPF** on the **VyOS↔ISR transit**), so ISR and OpenNebula paths learn the prefix without an ISR SVI on that segment.
- **NAT:** Only **one** device performs **PAT** to the ISP for a given flow — with VyOS as edge, **PAT is on VyOS WAN**; ISR **routes** toward VyOS and does **not** duplicate NAT for the same Internet path.

**Legacy / transitional:** An older layout kept **Google Home SVI + NAT on the ISR** for minimal churn; that remains documented as **Profile A/B** in [VLAN_MATRIX.md](../deployments/opennebula-kvm/VLAN_MATRIX.md) for brownfield comparison only.

Rules that never change: **no OSPF on RatTrap VLANs**; **do not** stack two PAT hops to the same ISP path without explicit policy. See [EDGE-MINI-PC-VYOS-PACKETFENCE.md](opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md).

## Automation prerequisite: Vault + IAM on the second mini PC

**Before** relying on repo automation (Ansible pulling secrets, n8n with Keycloak, Gitea OIDC, etc.), **Vault and Keycloak must be reachable** from the controllers that run jobs.

- **Logical network:** Same as everywhere else: **`iam_net` = `100.64.20.0/24`** ([NETWORK_DESIGN.md](NETWORK_DESIGN.md)).
- **Physical host:** A **second mini PC** running **Docker** + **Docker Compose**, with bridge networks created to match [NETWORK_DESIGN.md](NETWORK_DESIGN.md) — use [`scripts/create-networks.ps1`](../scripts/create-networks.ps1) or [`scripts/create-networks.sh`](../scripts/create-networks.sh), [`opentofu/`](../opentofu/) (Docker networks), or [`ansible/playbooks/deploy-devsecops-mesh.yml`](../ansible/playbooks/deploy-devsecops-mesh.yml) on that host (see [IaC: IAM-only host](#iac-iam-only-mini-pc) below).
- **Bootstrap:** Day-0 secrets can come from the **Deploy** repo greenfield flow ([REPO_SCOPE.md](REPO_SCOPE.md): `C:\GiTeaRepos\Deploy`) or **Ansible Vault**–encrypted `devsecops_secrets` until KV is populated; see [VARLOCK_USAGE.md](VARLOCK_USAGE.md) and [`ansible/roles/devsecops_containers/README.md`](../ansible/roles/devsecops_containers/README.md).

**Routing:** Controllers (laptop Semaphore, n8n host, etc.) need **L3 routes** (static or OSPF) to reach the IAM mini PC’s **`100.64.20.0/24`** (or the host’s published address if you NAT port-forward in lab — document whichever you choose in your IPAM).

**Offline / no controller yet:** Use a **USB data-partition bundle** (repo snapshot + vendored Ansible collections + optional `docker save` tarballs + `bootstrap-on-target.sh`) so you can run automation **before** lab Git or full routing exists. See [BOOTSTRAP_USB_BUNDLE.md](BOOTSTRAP_USB_BUNDLE.md) and [deployments/bootstrap-usb-bundle/README.md](../deployments/bootstrap-usb-bundle/README.md). Graduate to a custom live ISO only after this layout is stable.

## Repo scope and deferrals

| In scope (this repo) | Out of scope (plan / other repos) |
|----------------------|-----------------------------------|
| Ansible playbooks and roles for **mesh**, **containers**, **Cisco**, **mini PC Incus host** | **Installing** OpenNebula (front-end, schedulers, node join) as turnkey IaC |
| Docker Compose definitions and **stack-manifest** parity | **VyOS / PacketFence config-as-code** until CLI contracts are frozen (track as a follow-up issue) |
| **100.64** + VLAN matrix as **numeric** source of truth | ISP / home LAN numbering outside documented carve |

- **Greenfield one-shot** (clone, bootstrap, Vault seed): primary docs and scripts under **`C:\GiTeaRepos\Deploy`** per [REPO_SCOPE.md](REPO_SCOPE.md).
- **Runtime operations and OpenNebula-oriented orchestration**: **`C:\GiTeaRepos\devsecops-pipeline`** (this repo).

## IaC: Mini-PC-IAM (LXC-first, recommended)

Provision **only** the `devsecops-iam` LXC (Docker-in-LXC: Vault, Keycloak, Teleport) on an Alma + Incus host:

```bash
cd ansible
ansible-playbook -i inventory/mini-pc-iam.yml playbooks/bootstrap-mini-pc-iam.yml -K
```

Inventory must define **`lxd_hosts`** (see [`ansible/inventory/mini-pc-iam.example.yml`](../ansible/inventory/mini-pc-iam.example.yml)). USB bundle: set `RUN_ANSIBLE_BOOTSTRAP=1` and `BOOTSTRAP_INVENTORY` — [BOOTSTRAP_USB_BUNDLE.md](BOOTSTRAP_USB_BUNDLE.md).

## IaC: IAM-only mini PC (bare-metal Docker)

Use the same **`ai_mesh_nodes`** group the mesh and container playbooks expect, but limit the **devsecops_containers** role to **IAM** only with **extra vars** (after networks exist on that host).

**1. Networks on the IAM mini PC** (once per host):

```bash
cd ansible
ansible-playbook -i inventory/mini-pc-iam.yml playbooks/deploy-devsecops-mesh.yml -K
```

Copy [`ansible/inventory/mini-pc-iam.example.yml`](../ansible/inventory/mini-pc-iam.example.yml) → **`ansible/inventory/mini-pc-iam.yml`** with your real `ansible_host` and user (that filename is **gitignored** in `.gitignore`).

**2. IAM stack only** (secrets from Vault on controller, or pre-set `devsecops_secrets` via Ansible Vault):

```bash
ansible-playbook -i inventory/mini-pc-iam.yml playbooks/start-containers-with-vault.yml -K \
  -e 'start_messaging=false' \
  -e 'start_tooling=false' \
  -e 'start_chatops=false' \
  -e 'start_gateway=false' \
  -e 'start_iam=true'
```

If you use **`playbooks/site.yml`**, it always runs **hardening** and **full mesh** then **all default stacks** unless you override — for a **minimal IAM-only** pass, prefer **`deploy-devsecops-mesh.yml`** plus **`start-containers-with-vault.yml`** (or a small wrapper play) with the toggles above. Role toggle reference: [`ansible/roles/devsecops_containers/README.md`](../ansible/roles/devsecops_containers/README.md).

**3. Clone / sync** the repo (or a deployment checkout) onto the IAM mini PC so `docker-compose/` paths in the role resolve (`devsecops_compose_base` is relative to playbooks).

## Related documents

- [ROADMAP.md](ROADMAP.md) — phased delivery
- [DEPLOYMENT.md](DEPLOYMENT.md) — full execution order
- [SYSTEMS_ARCHITECTURE.md](SYSTEMS_ARCHITECTURE.md) — components and Vault
- [opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md](opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md) — edge mini PC architecture
- [deployments/mini-pc-firewall/README.md](../deployments/mini-pc-firewall/README.md) — Packer + Ansible for edge host
