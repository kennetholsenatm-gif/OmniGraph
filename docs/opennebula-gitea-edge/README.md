# OpenNebula + Gitea Edge Migration



Documentation for moving **Gitea** off a Windows workstation (`C:\GiTeaRepos`) and running the **full DevSecOps pipeline** (IAM, messaging, tooling, ChatOps, gateway — not only Git) on OpenNebula/Linux with **`100.64.x`** networking. **Prefer native Alma + Podman**; existing **Docker Compose** remains a documented transitional path — see **[REDUCE-DOCKER.md](REDUCE-DOCKER.md)**.

1. **[REDUCE-DOCKER.md](REDUCE-DOCKER.md)** — **minimize Docker**: native Alma + **Podman** tiers; compose as legacy reference.  
2. **[LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md)** — **AlmaLinux 10 LXC** per stack on OpenNebula (native/Podman preferred).  
3. **[EDGE-MINI-PC-VYOS-PACKETFENCE.md](EDGE-MINI-PC-VYOS-PACKETFENCE.md)** — **mini PC edge**: VyOS (Incus LXC), PacketFence (VM), RatTrap **900/901**, PBR, **WAN VLAN 99**, handoff to **`100.64.x`** / ISR / OpenNebula (aligns with [VLAN_MATRIX](../../deployments/opennebula-kvm/VLAN_MATRIX.md)). **Build path:** [deployments/mini-pc-firewall/README.md](../../deployments/mini-pc-firewall/README.md) (Packer + Semaphore).  
4. **[CONTAINER-LIFT-TO-OPENNEBULA.md](CONTAINER-LIFT-TO-OPENNEBULA.md)** — volume moves for **remaining** OCI workloads.  
5. [WHOLE-REPO-MIGRATION-SCOPE.md](WHOLE-REPO-MIGRATION-SCOPE.md) — repo-wide checklist + secrets.  
6. [01-node-roles-and-placement.md](01-node-roles-and-placement.md) — optional hardware placement.
7. [LEAN_LOCAL_CONTROL_PLANE.md](LEAN_LOCAL_CONTROL_PLANE.md) — lean laptop control plane (Semaphore + tooling), full runtime on OpenNebula.



## Repo mapping (canonical infrastructure)



This folder **aligns** with the DevSecOps pipeline’s existing addressing and OpenNebula conventions—do not invent a parallel `10.10.x.x` scheme unless you are in a non-repo lab.



| Source | Role |

|--------|------|

| [deployments/opennebula-kvm/VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md) | OpenNebula VNET **names**, **VLAN IDs** (`2000 + third octet` for `100.64.x.0/24`), **Virtual Router** and **ISR** rules (single PAT on WAN, no VR SNAT for Internet) |

| [docs/NETWORK_DESIGN.md](../NETWORK_DESIGN.md) | Same **`100.64.x.0/24`** segments as Docker `*_net` names (`gitea_net` ↔ `100.64.1.0/24`, `gateway_net` ↔ `100.64.5.0/24`, …) |

| [docs/DOCSIFY_GITEA.md](../DOCSIFY_GITEA.md) | Gitea-adjacent **Docsify** sync, **webhooks** (`X-Gitea-Signature`), Traefik **`/docs`** routing |

| [ansible/playbooks/network-c3560cx.yml](../../ansible/playbooks/network-c3560cx.yml) | Catalyst **WS-C3560CX** automation (trunks/VLANs as implemented in repo) |

| [ansible/playbooks/network-isr.yml](../../ansible/playbooks/network-isr.yml) | **ISR** edge playbooks (companion to matrix) |

| [EDGE-MINI-PC-VYOS-PACKETFENCE.md](EDGE-MINI-PC-VYOS-PACKETFENCE.md) | **Mini PC** VyOS + PacketFence + RatTrap; PAT on VyOS WAN; **OSPF** transit vs **900/901**; integrations to `100.64.20/51/54/2/50` |
| [deployments/mini-pc-firewall/README.md](../../deployments/mini-pc-firewall/README.md) | **Packer** QCOW2 + **Ansible** Incus host + [Semaphore templates](../../deployments/mini-pc-firewall/semaphore/TEMPLATE-EXAMPLE.md) |

| [ansible/roles/cisco_isr_platform/tasks/ucse.yml](../../ansible/roles/cisco_isr_platform/tasks/ucse.yml) | **UCS-E** / `ucse` interface patterns referenced by the matrix platform carve |



**Quick reference:** Gitea VM traffic in OpenNebula uses VNET **`devsecops-gitea`** · VLAN **2001** · **`100.64.1.0/24`** (matches `gitea_net`). Single-pane / Traefik ingress aligns with **`devsecops-gateway`** · VLAN **2005** · **`100.64.5.0/24`**. Edge LAN is **`devsecops-edge`** · VLAN **86** · **`192.168.86.0/24`**.

**Operator checklist (refined plan):** [REFINED-EXECUTION.md](REFINED-EXECUTION.md) · OpenNebula **`*.one`** templates: [deployments/opennebula-kvm/onevnet/](../../deployments/opennebula-kvm/onevnet/)

### Phased automation (mini PC edge)

- **Phase A — Switch:** extend [ansible/playbooks/network-c3560cx.yml](../../ansible/playbooks/network-c3560cx.yml) for **VLAN 99, 900, 901** and **trunk allowed VLANs** to the mini PC after IDs are frozen in [VLAN_MATRIX](../../deployments/opennebula-kvm/VLAN_MATRIX.md).
- **Phase B/C — VyOS / PacketFence:** defer Ansible roles until **VyOS CLI** and **Incus** profiles are pinned — see [EDGE-MINI-PC-VYOS-PACKETFENCE.md](EDGE-MINI-PC-VYOS-PACKETFENCE.md) § Automation (phased).



## Contents



| Doc | Purpose |

|-----|---------|

| [REDUCE-DOCKER.md](REDUCE-DOCKER.md) | **Native / Podman / Docker** tiers; service mapping off compose |

| [LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md) | **Alma 10 LXC** (`devsecops-iam`, …) on OpenNebula; **prefer no Docker** inside |

| [EDGE-MINI-PC-VYOS-PACKETFENCE.md](EDGE-MINI-PC-VYOS-PACKETFENCE.md) | **Edge mini PC** — VyOS LXC, PacketFence VM, RatTrap **900/901**, PBR, WAN **99**, `100.64` integrations |

| [deployments/mini-pc-firewall/README.md](../../deployments/mini-pc-firewall/README.md) | **Packer + Ansible + Semaphore** — golden QCOW2, Incus host role, template examples |

| [CONTAINER-LIFT-TO-OPENNEBULA.md](CONTAINER-LIFT-TO-OPENNEBULA.md) | **Compose stacks** — volumes, networks, migration order (flat Docker or inside LXC) |

| [WHOLE-REPO-MIGRATION-SCOPE.md](WHOLE-REPO-MIGRATION-SCOPE.md) | **Entire `devsecops-pipeline` artifact map** — compose, Ansible, Vault, n8n, webhooks, segments to re-point |

| [01-node-roles-and-placement.md](01-node-roles-and-placement.md) | Hardware → role mapping (FE, KVM, K3s, Ceph) + UCS / matrix carve |

| [02-network-topology-vlan-acl.md](02-network-topology-vlan-acl.md) | Matrix-aligned VLANs, VR/ISR, `.one` snippets, ACLs |

| [03-storage-ceph-datastores.md](03-storage-ceph-datastores.md) | Local + Ceph datastores; **`devsecops-ceph`** segment |

| [04-gitea-k3s-ha.md](04-gitea-k3s-ha.md) | K3s stack, `100.64.1` / `100.64.5`, Docsify webhooks |

| [05-migration-runbook.md](05-migration-runbook.md) | Phased migration + post-cutover Docsify/webhook checklist |

| [06-risks-mitigations-rollback.md](06-risks-mitigations-rollback.md) | Risks, mitigations, rollback |

| [REFINED-EXECUTION.md](REFINED-EXECUTION.md) | Phased checklist (Ansible → VNETs → Ceph/K3s → Helm → DNS) |

| [DOCSIFY-POST-MIGRATION-CHECKLIST.md](DOCSIFY-POST-MIGRATION-CHECKLIST.md) | Docsify + webhooks + CI paths after cutover |

| [../../ansible/playbooks/opennebula-hybrid-site.yml](../../ansible/playbooks/opennebula-hybrid-site.yml) | Hybrid orchestrator: LXC stack + optional K3s Helm Gitea |
| [LEAN_LOCAL_CONTROL_PLANE.md](LEAN_LOCAL_CONTROL_PLANE.md) | Local control-only mode (Semaphore, lint/security/reporting tools) |



## Deployment assets



- [Helm values (Gitea)](../../deployments/opennebula-gitea/helm/gitea-values.example.yaml)

- [Helm values (PostgreSQL)](../../deployments/opennebula-gitea/helm/postgresql-values.example.yaml)

- [Example Ceph RBD StorageClass](../../deployments/opennebula-gitea/k8s/storageclass-ceph-rbd.example.yaml)

- [Kustomize overlay stub](../../deployments/opennebula-gitea/kustomize/README.md)

- [Deployment README](../../deployments/opennebula-gitea/README.md)

- [Ceph + K3s bring-up](../../deployments/opennebula-gitea/BRINGUP.md)

- [OpenNebula VNET templates (`*.one`)](../../deployments/opennebula-kvm/onevnet/README.md)


