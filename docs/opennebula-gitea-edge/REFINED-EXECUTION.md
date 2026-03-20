# Refined execution: Windows Gitea → OpenNebula (repo infrastructure)

Operator checklist for the **refined plan** using canonical [**VLAN_MATRIX.md**](../../deployments/opennebula-kvm/VLAN_MATRIX.md), [**NETWORK_DESIGN.md**](../NETWORK_DESIGN.md), and this folder. Do not treat **`10.10.x`** examples from legacy drafts as authoritative.

## Phase 0 — Whole repository and integration inventory

Before touching hypervisors, enumerate **everything in this repo** that references Gitea, `C:\GiTeaRepos`, or Docker **`gitea_net`**. Work through [WHOLE-REPO-MIGRATION-SCOPE.md](WHOLE-REPO-MIGRATION-SCOPE.md) (compose, Ansible, Vault/Varlock, `n8n-workflows/`, `single-pane-of-glass/`, schemas). Update **`GITEA_*`** and webhook secrets in Vault; grep for `GiTeaRepos` and `gitea:3000` after cutover.

## Phase A — Cisco + OpenNebula + storage

| Step | Action | References |
|------|--------|------------|
| A1 | Reconcile **C3560CX** trunk `allowed_vlans` with **86, 2001, 2005, 2250** (+ platform **240–244** as needed) | `ansible/group_vars/network_c3560cx.yml.example`, [ANSIBLE_NETWORK_DEVICES.md](../ANSIBLE_NETWORK_DEVICES.md) |
| A2 | Optional **ISR** subinterfaces / static **100.64.0.0/10** route via VR | `ansible/group_vars/network_isr.yml.example` |
| A3 | Create Linux **bridges** + OpenNebula **VNETs** | [deployments/opennebula-kvm/onevnet/](../../deployments/opennebula-kvm/onevnet/), [02-network-topology-vlan-acl.md](02-network-topology-vlan-acl.md) |
| A4 | Configure **Virtual Router** (**.1** per `100.64.x`/24; edge **.2**) | [onevnet/VR-NOTES.md](../../deployments/opennebula-kvm/onevnet/VR-NOTES.md), VLAN_MATRIX |
| A5 | **Local** system datastores + **Ceph** on **devsecops-ceph** | [03-storage-ceph-datastores.md](03-storage-ceph-datastores.md) |

## Phase B — Ceph + K3s (optional if Gitea-on-K8s)

| Step | Action | References |
|------|--------|------------|
| B1 | Deploy Ceph; RBD pools; **100.64.250.0/24** front-end | [BRINGUP.md](../../deployments/opennebula-gitea/BRINGUP.md) |
| B2 | K3s VMs on **devsecops-gitea**; **ceph-csi**; **StorageClass** | [BRINGUP.md](../../deployments/opennebula-gitea/BRINGUP.md), [storageclass example](../../deployments/opennebula-gitea/k8s/storageclass-ceph-rbd.example.yaml) |

## Phase B2 — Docker Compose stacks (mandatory for full pipeline)

**Vault, Keycloak, Solace/Kafka, n8n, Zammad, Gitea (compose), Traefik gateway** are **containers**.

**Preferred:** **AlmaLinux 10 LXC** per stack (**Docker-in-LXC**) on **one OpenNebula KVM guest** running LXD/Incus — [LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md), [`deploy-devsecops-lxc.yml`](../../ansible/playbooks/deploy-devsecops-lxc.yml), inventory [`opennebula-lxd.example.yml`](../../ansible/inventory/opennebula-lxd.example.yml).

| Step                                                            | Action                                                                                                                                                                                                                                                                   | References                                                                                                                                                   |
| --------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| B2a                                                             | Inventory stacks; backup volumes; plan **DEPLOYMENT.md** order                                                                                                                                                                                                         | [CONTAINER-LIFT-TO-OPENNEBULA.md](CONTAINER-LIFT-TO-OPENNEBULA.md), [stack-manifest.json](../../docker-compose/stack-manifest.json)                          |
| B2b                                                             | **LXC path:** provision `devsecops-iam`, `devsecops-messaging`, `devsecops-tooling`, `devsecops-gateway`, …; restore volumes **into each LXC’s** Docker; compose up                                                                                                       | [LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md), [lxd_devsecops_stack defaults](../../ansible/roles/lxd_devsecops_stack/defaults/main.yml)                |
| B2c                                                             | **Flat Docker path:** OpenNebula VM(s) with Docker only; `create-networks` / OpenTofu on target; restore volumes; `docker compose up` or [devsecops_containers](../../ansible/roles/devsecops_containers/README.md)                                                      | [CONTAINER-LIFT-TO-OPENNEBULA.md](CONTAINER-LIFT-TO-OPENNEBULA.md)                                                                                           |

*Hybrid:* Gitea on K3s (Phase B) **plus** remaining stacks in **LXCs** or flat Docker is valid.

## Phase C — Helm Gitea + data migration (if using K3s for Gitea)

| Step | Action | References |
|------|--------|------------|
| C1 | **MetalLB** / ingress VIPs prefer **100.64.5.0/24** (**devsecops-gateway**) | [04-gitea-k3s-ha.md](04-gitea-k3s-ha.md) |
| C2 | **cert-manager**, **PostgreSQL**, **Gitea** Helm; `ROOT_URL` + TLS | [helm/gitea-values.example.yaml](../../deployments/opennebula-gitea/helm/gitea-values.example.yaml) |
| C3 | **Offline cutover** from Windows: freeze, `gitea dump`, transfer, restore | [05-migration-runbook.md](05-migration-runbook.md) |

## Phase D — DNS + integrations

| Step | Action | References |
|------|--------|------------|
| D1 | DNS **`git.<domain>`** → ingress VIP (**100.64.5.x** or NAT target) | [02-network-topology-vlan-acl.md](02-network-topology-vlan-acl.md) |
| D2 | **Docsify / webhooks** validation | [DOCSIFY-POST-MIGRATION-CHECKLIST.md](DOCSIFY-POST-MIGRATION-CHECKLIST.md) |
| D3 | Risks / rollback understood | [06-risks-mitigations-rollback.md](06-risks-mitigations-rollback.md) |

## Hardware reminder

| Physical | Role |
|----------|------|
| Mini-PC-1 | OpenNebula FE primary; optional Ceph |
| Mini-PC-2 | Backup / monitoring / optional Ceph |
| 4× UCS-E | KVM; underlay **100.64.245–247** carve per matrix |
| ISR + C3560CX | Trunks / NAT / ACLs |
