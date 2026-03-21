# Roadmap — physical prerequisites and OpenNebula lift

Phases below align with [CANONICAL_DEPLOYMENT_VISION.md](CANONICAL_DEPLOYMENT_VISION.md). **Numeric** addressing stays authoritative in [NETWORK_DESIGN.md](NETWORK_DESIGN.md) and [deployments/opennebula-kvm/VLAN_MATRIX.md](../deployments/opennebula-kvm/VLAN_MATRIX.md).

## P0 — Freeze IPAM and VLAN IDs

| Goal | Actions / assets |
|------|-------------------|
| No conflicting VLAN or `100.64` assignments | Lock **edge** VLANs (e.g. 99, 86, 900, 901, transit **298**) and **workload** `2000 + third octet` rule in [VLAN_MATRIX.md](../deployments/opennebula-kvm/VLAN_MATRIX.md). |
| Docker logical nets match automation | Confirm **17** bridge names/subnets match [NETWORK_DESIGN.md](NETWORK_DESIGN.md); verify with [`scripts/verify-stack-manifest.ps1`](../scripts/verify-stack-manifest.ps1) after `stack-manifest.json` edits. |

## P1 — Edge mini PC (VyOS firewall, Google Home on VyOS)

| Goal | Actions / assets |
|------|------------------|
| **Firewall/router outside OpenNebula** | [opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md](opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md) — VyOS WAN, trunk, RatTrap hairpin, **OSPF** to ISR. |
| **Google Home on VyOS** | **`100.64.244.0/24`**: VyOS **SVI + default gateway** for that VLAN; **advertise** prefix toward ISR (see **Profile C** in [VLAN_MATRIX.md](../deployments/opennebula-kvm/VLAN_MATRIX.md)). |
| Host image + Incus | [deployments/mini-pc-firewall/README.md](../deployments/mini-pc-firewall/README.md), [`ansible/playbooks/mini-pc-firewall-host.yml`](../ansible/playbooks/mini-pc-firewall-host.yml), [`ansible/inventory/mini-pc-firewall.example.yml`](../ansible/inventory/mini-pc-firewall.example.yml). |
| Switch trunks | [`ansible/playbooks/network-c3560cx.yml`](../ansible/playbooks/network-c3560cx.yml) (extend when VLAN list is frozen). |

**Known gap:** Version-pinned **VyOS / PacketFence** Ansible modules or config export are **not** in repo until CLI/API contracts are frozen — track in a repo issue; use manual cutover checklists until then.

## P2 — Second mini PC: Vault + IAM (`iam_net`)

| Goal | Actions / assets |
|------|------------------|
| **Docker bridges** on IAM host | [`scripts/create-networks.ps1`](../scripts/create-networks.ps1) / [`scripts/create-networks.sh`](../scripts/create-networks.sh), or [`opentofu/`](../opentofu/) `docker_network` resources, or [`ansible/playbooks/deploy-devsecops-mesh.yml`](../ansible/playbooks/deploy-devsecops-mesh.yml). |
| **IAM compose** | [`docker-compose/docker-compose.iam.yml`](../docker-compose/docker-compose.iam.yml). |
| **Automation** | [`ansible/playbooks/start-containers-with-vault.yml`](../ansible/playbooks/start-containers-with-vault.yml) with IAM-only toggles; see [CANONICAL_DEPLOYMENT_VISION.md](CANONICAL_DEPLOYMENT_VISION.md#iac-iam-only-mini-pc) and [`ansible/inventory/mini-pc-iam.example.yml`](../ansible/inventory/mini-pc-iam.example.yml). |
| **Docs** | [SYSTEMS_ARCHITECTURE.md](SYSTEMS_ARCHITECTURE.md), [VARLOCK_USAGE.md](VARLOCK_USAGE.md). |
| **Greenfield bootstrap** | Optional: **`C:\GiTeaRepos\Deploy`** per [REPO_SCOPE.md](REPO_SCOPE.md). |
| **Offline Mini-PC-IAM** | [BOOTSTRAP_USB_BUNDLE.md](BOOTSTRAP_USB_BUNDLE.md), [`deployments/bootstrap-usb-bundle/`](../deployments/bootstrap-usb-bundle/README.md) — build bundle on a connected host; run `bootstrap-on-target.sh` from USB |

## P3+ — OpenNebula runtime and LXC / Compose lift

| Goal | Actions / assets |
|------|------------------|
| Hypervisor guest + Incus | [opennebula-gitea-edge/LXC-ALMA10-OPENNEBULA.md](opennebula-gitea-edge/LXC-ALMA10-OPENNEBULA.md), [`ansible/playbooks/deploy-devsecops-lxc.yml`](../ansible/playbooks/deploy-devsecops-lxc.yml). |
| Hybrid + optional Helm Gitea | [`ansible/playbooks/opennebula-hybrid-site.yml`](../ansible/playbooks/opennebula-hybrid-site.yml), [`ansible/inventory/opennebula-hybrid.example.yml`](../ansible/inventory/opennebula-hybrid.example.yml). |
| Migration checklist | [opennebula-gitea-edge/REFINED-EXECUTION.md](opennebula-gitea-edge/REFINED-EXECUTION.md), [opennebula-gitea-edge/CONTAINER-LIFT-TO-OPENNEBULA.md](opennebula-gitea-edge/CONTAINER-LIFT-TO-OPENNEBULA.md). |

**Out of scope here:** OpenNebula **installation** automation; add under a future epic if you adopt a provider or appliance workflow.

## Quick reference

| Phase | You are done when |
|-------|-------------------|
| **P0** | VLAN_MATRIX + NETWORK_DESIGN committed; no ID collisions with production home LAN. |
| **P1** | VyOS PAT to ISP; Google Home VLAN **DG = VyOS**; ISR learns **`100.64.244.0/24`** without owning SVI (canonical profile). |
| **P2** | `VAULT_ADDR` / Keycloak reachable from controllers; Varlock path populated for automation clients. |
| **P3+** | Workloads run on OpenNebula-backed Linux per LXC/Compose playbooks. |
