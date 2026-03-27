# Node Roles and Placement

## Hardware inventory (reference)

| Role | Qty | Model / notes |
|------|-----|----------------|
| Edge routers | 2 | Cisco ISR4351 |
| Switches | 2 | Cisco WS-C3560CX-8TCS |
| Compute blades | 4 | Cisco UCS-E140S-M2/K9 (2 per ISR; KVM-capable bare metal) |
| Mini-PCs | 2 | ~16 GB RAM each |

## Repo alignment: UCS / ISR underlay

OpenNebula **KVM node** management IPs and **per-chassis** underlay are described in [deployments/opennebula-kvm/VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md) **Platform carve (100.64.240–254)**:

| Third octet | Typical use |
|-------------|-------------|
| **245** | UCS-A OpenNebula host underlay (IMC, ISR `ucse` leg) |
| **246** | UCS-B OpenNebula host underlay |
| **247** | UCS-C (add rows per chassis) |

**Prefix length is variable** (`/29`–`/24`) per matrix; size each subnet for hypervisor count + growth. IOS-XE UCS-E patterns (`ucse subslot`, `interface ucse1/0/0`) are implemented in Ansible: [ansible/roles/cisco_isr_platform/tasks/ucse.yml](../../ansible/roles/cisco_isr_platform/tasks/ucse.yml).

**Rule:** one **ISR-visible** L3 network per UCS domain—do not collapse two UCS installs into one ISR LAN if that violates your addressing model (matrix explains why).

## Recommended assignment

### OpenNebula control plane

| Node | Primary function | Notes |
|------|------------------|-------|
| **Mini-PC-1** | **OpenNebula Front End (primary)** | Sunstone/FireEdge (if used), `oned`, scheduler, MariaDB/MySQL for OpenNebula DB (or remote DB on same host). Keeps management off UCS-E. |
| **Mini-PC-2** | **Standby / utility** | Backup target for Gitea/OpenNebula dumps, monitoring stack (Prometheus/Grafana/Loki optional), scripted FE rebuild materials, optional second Ceph MON/OSD. |

### KVM hypervisors

| Node | Primary function |
|------|------------------|
| **UCS-E blade 1–4** | **OpenNebula KVM nodes** (`KVM` or `qemu-kvm` hosts). Run all tenant VMs including K3s and optional third Ceph OSD node. |

### Ceph (recommended 3-node minimum for quorum)

| Node | Ceph daemons (typical) |
|------|-------------------------|
| Mini-PC-1 | `mon`, `mgr`, `osd` (dedicated disk or partition) |
| Mini-PC-2 | `mon`, `mgr`, `osd` |
| **One UCS-E host** (e.g. blade with most spare disk) | `mon` (optional if only 3 nodes total use 3 mons on these three), `osd` |

If a fourth physical disk set is available on a second UCS-E node, add OSDs for more resilience and rebalance PGs.

### K3s (Gitea workload)

Deploy **as VMs on OpenNebula**, not directly on bare metal Mini-PCs, unless capacity forces a hybrid model.

| VM | vCPU / RAM (starting point) | Role |
|----|------------------------------|------|
| `k3s-cp-1` | 2–4 / 4–8 GB | K3s server (control plane) |
| `k3s-worker-1` | 4–8 / 8–16 GB | Agent; schedules Gitea + ingress |
| `k3s-worker-2` | 4–8 / 8–16 GB | Agent; schedules Gitea + DB or separate DB VM |

**HA note:** True K3s control-plane HA needs an odd number of server nodes (typically 3) or an external datastore. On this footprint, options are:

1. **Pragmatic:** 1 control-plane VM + etcd embedded; accept CP single point of failure; mitigate with fast rebuild from IaC and backups.
2. **Stronger:** Add `k3s-cp-2` and `k3s-cp-3` on two additional UCS-E hosts when resources allow.

### Gitea components (logical)

- **HTTP(S) / Git over HTTPS:** Ingress (Traefik or NGINX) + TLS (cert-manager).
- **SSH:** Either NodePort/LoadBalancer to `gitea` SSH port, or separate VM/LB with TCP passthrough; document chosen VIP.
- **Database:** PostgreSQL (recommended) or MariaDB; run in-cluster (Helm) or as a dedicated OpenNebula VM for simpler ops.
- **Object storage (LFS/attachments):** Local PVC on Ceph RBD or S3-compatible later.

## Traffic domains

- **North-south:** Client → ISR → **192.168.86.0/24** (VLAN **86**, `devsecops-edge`) and/or NAT to workload segments per [VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md); Git HTTPS often fronted via **`devsecops-gateway`** (**100.64.5.0/24**).
- **East-west:** K3s pod network; Gitea user-facing NIC on **`devsecops-gitea`** (**100.64.1.0/24**, same octet as `gitea_net` in [NETWORK_DESIGN.md](../NETWORK_DESIGN.md)); Ceph on **`devsecops-ceph`** (**100.64.250.0/24**, VLAN **2250**).
- **Management:** OpenNebula admins → edge / platform carve → Sunstone / SSH to hosts on per-UCS underlay (**245–247** carve) as applicable.

## Rationale summary

- Mini-PCs host **orchestration and durability services** (OpenNebula FE, backups, monitoring, Ceph quorum).
- UCS-E blades provide **dense KVM capacity** for K3s and future tenants without competing with `oned` on constrained flash/RAM profiles.
- Gitea runs in **K3s** for rolling updates, ingress TLS, and volume integration via Ceph CSI.
