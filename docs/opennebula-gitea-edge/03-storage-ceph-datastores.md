# Storage Architecture: OpenNebula Datastores and Ceph (Repo-Aligned)



**Network alignment:** Ceph **client/public** (and optional **cluster**) traffic should use the dedicated segment **`devsecops-ceph`** · **VLAN 2250** · **`100.64.250.0/24`** defined in [deployments/opennebula-kvm/VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md). Do not use a non-canonical `10.10.20.0/24` unless documenting an isolated lab.



## Constraints



- No dedicated SAN/NAS.

- Mix of **Mini-PC disks** and **UCS-E local SSD/HDD** (and platform carve **100.64.240–254** per matrix).

- Must support **K3s PersistentVolumes** for Gitea (repos, attachments, LFS) and preferably **shared** VM images across KVM nodes.



## Recommended architecture



### Tier 1: Per-host system datastore (OpenNebula)



| Datastore | Type | Backing | Use |

|-----------|------|---------|-----|

| `ds-system-local-*` | FS or LVM on each KVM host | Local SSD on UCS-E | Fast VM disks, ephemeral scratch |



**Purpose:** Lowest latency for running VMs; avoids pushing all IOPS through Ceph for every workload.



### Tier 2: Shared image / persistent datastore (Ceph)



| Component | Placement |

|-----------|-----------|

| **MON** | 3 instances: Mini-PC-1, Mini-PC-2, one UCS-E node (typical) |

| **MGR** | Co-locate with MON (active/passive) |

| **OSD** | One or more disks per participating node (SSD strongly preferred) |



**Network:** Bind **public_network** (and **cluster_network** if split) to interfaces on **`100.64.250.0/24`** (VLAN **2250**). Keep Ceph off the Internet; ISR ACLs should deny **`100.64.250.0/24`** → WAN—see [02-network-topology-vlan-acl.md](02-network-topology-vlan-acl.md).



**Ceph pool for RBD (example names):**



- `opennebula-rbd` — OpenNebula Ceph datastore (optional if using COPY scripts).

- `k8s-rbd` — Kubernetes `storageClass` via **ceph-csi-rbd** (example: [deployments/opennebula-gitea/k8s/storageclass-ceph-rbd.example.yaml](../../deployments/opennebula-gitea/k8s/storageclass-ceph-rbd.example.yaml)).



#### Replication for small clusters



For **three OSD hosts**, a common compromise:



- `size=2`, `min_size=1` — survives one full host outage but **not** ideal (risk of data loss if two OSDs with copies fail). Document this explicitly.

- Preferred when fourth OSD host is available: `size=3`, `min_size=2`.



Tune `pg_num` / `pgp_num` with the autoscaler after OSD count stabilizes.



### OpenNebula integration



1. Install Ceph client packages on all KVM nodes.

2. Configure OpenNebula **Ceph datastore** pointing at pool `opennebula-rbd` with appropriate `ceph_user` keyring.

3. For VM disks that must move between hosts, place them on Ceph; for fixed workers, local + backup may suffice.



### Kubernetes integration



1. Deploy **ceph-csi** (`rbd` plugin) in K3s.

2. Create `StorageClass` with provisioner appropriate to your CSI release (e.g. `rbd.csi.ceph.com`).

3. Use `ReadWriteOnce` volumes for Gitea and PostgreSQL unless using a clustered DB with shared storage (not default).



## Backup strategy



| Data | Frequency | Target | Method |

|------|-----------|--------|--------|

| Gitea DB + repos | Daily + pre-cutover | Mini-PC-2 + offsite | `gitea dump` or volume snapshots + `pg_dump` |

| OpenNebula DB | Daily | Mini-PC-2 | DB dump + `onedb` backup procedures |

| Ceph config | On change | Git repo | Export `ceph.conf`, keyrings, crush maps |



### Gitea application backup (conceptual)



- **Preferred:** native `gitea dump` including **repositories**, **database**, and **custom**/`data` where applicable.

- **Database-only:** `pg_dump` / `mysqldump` if DB is external and repos on PVC — still snapshot file-backed git data.



Restore always uses a **matching or newer** Gitea version; read upstream release notes before major jumps.



## Fallback: no Ceph



If operational cost prohibits Ceph:



1. Use **local** OpenNebula datastores on each KVM host.

2. Store schedules pinned per host OR replicate VM disks with **offline** `rsync`/`qemu-img convert` jobs.

3. Use **single-VM Gitea** with scheduled `gitea dump` to Mini-PC-2.



Trade-off: **no** shared PVC HA across nodes; K3s node loss may require manual reschedule with restored volumes.



## Sizing checklist



- [ ] Measure current repo total size + DB size + LFS.

- [ ] Plan Ceph raw capacity ≥ **3×** logical data for `size=3`, or **2×** for `size=2` plus headroom.

- [ ] Leave **20–30%** free on SSD pools for recovery and compaction.

- [ ] Confirm **`devsecops-ceph`** row exists in [VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md) and trunks include VLAN **2250**.


