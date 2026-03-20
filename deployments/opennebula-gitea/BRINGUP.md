# Ceph + K3s bring-up (repo segments)

Anchored to [docs/opennebula-gitea-edge/03-storage-ceph-datastores.md](../../docs/opennebula-gitea-edge/03-storage-ceph-datastores.md) and [04-gitea-k3s-ha.md](../../docs/opennebula-gitea-edge/04-gitea-k3s-ha.md).

## Prerequisites

- Linux bridges / OpenNebula VNETs **devsecops-ceph** (**100.64.250.0/24**, VLAN **2250**) and **devsecops-gitea** (**100.64.1.0/24**, VLAN **2001**) deployed ([onevnet/](../opennebula-kvm/onevnet/) templates).
- Hypervisor NICs can reach Ceph ports on **100.64.250.0/24** (firewall + ISR ACLs deny WAN).

## Ceph (summary)

1. Install Ceph (cephadm or Ansible) on **three** nodes: Mini-PC-1, Mini-PC-2, one UCS-E (see storage doc).
2. Bind **`public_network`** to **100.64.250.0/24**; optionally split **`cluster_network`**.
3. Create RBD pool(s): `opennebula-rbd` / `k8s-rbd` per [03-storage-ceph-datastores.md](../../docs/opennebula-gitea-edge/03-storage-ceph-datastores.md).
4. Distribute `ceph.conf` and client keyrings to **all** KVM nodes.

## K3s

1. Provision **3+ OpenNebula VMs** (example: 1 server, 2 agents) with primary dataplane NIC on **devsecops-gitea**.
2. Attach a second NIC to VMs that must speak Ceph on **devsecops-ceph**, **or** route **100.64.250.0/24** from worker subnets via VR/ISR (prefer direct storage VLAN for latency).
3. Install **K3s**; install **ceph-csi** RBD driver.
4. Apply [k8s/storageclass-ceph-rbd.example.yaml](k8s/storageclass-ceph-rbd.example.yaml) (edit `clusterID`, secrets, pool).

## Validate

```bash
kubectl get csinode
kubectl get sc
```

## Next

- Helm: [README.md](README.md), [helm/gitea-values.example.yaml](helm/gitea-values.example.yaml)
- Cutover: [docs/opennebula-gitea-edge/05-migration-runbook.md](../../docs/opennebula-gitea-edge/05-migration-runbook.md)
