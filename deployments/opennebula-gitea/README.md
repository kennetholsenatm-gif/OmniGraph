# OpenNebula edge — Gitea on K3s deployment assets

Example Helm values and stubs for deploying Gitea after OpenNebula VMs and Ceph CSI are ready.

## Prerequisites

- K3s cluster with working `kubectl`
- `StorageClass` for Ceph RBD (replace name `ceph-rbd` in values)
- cert-manager + ClusterIssuer (e.g. `letsencrypt-dns`)
- Stable LoadBalancer IP pool (MetalLB) on DMZ VLAN for SSH if not using ingress TCP

## Layout

| Path | Description |
|------|-------------|
| [BRINGUP.md](BRINGUP.md) | Ceph + K3s + CSI order of operations (repo segments) |
| [helm/gitea-values.example.yaml](helm/gitea-values.example.yaml) | Gitea chart overrides |
| [helm/postgresql-values.example.yaml](helm/postgresql-values.example.yaml) | Standalone PostgreSQL if not using embedded chart |
| [k8s/storageclass-ceph-rbd.example.yaml](k8s/storageclass-ceph-rbd.example.yaml) | Example `StorageClass` for ceph-csi |
| [kustomize/](kustomize/) | Optional namespaced stubs |

Copy `*.example.yaml` to `*.yaml`, fill secrets via CI or SOPS, and install with Helm.

## Install (illustrative)

```bash
helm repo add gitea-charts https://gitea-charts.gitea.io
helm repo update

# After copying and editing values:
helm upgrade --install gitea gitea-charts/gitea \
  --namespace gitea-prod --create-namespace \
  -f helm/gitea-values.yaml
```
