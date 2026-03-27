# OpenNebula Virtual Network templates (`*.one`)

These fragments follow [VLAN_MATRIX.md](../VLAN_MATRIX.md) and [docs/opennebula-gitea-edge/02-network-topology-vlan-acl.md](../../docs/opennebula-gitea-edge/02-network-topology-vlan-acl.md).

## Usage

1. Replace **`PHYDEV`** with your KVM host trunk interface or bond (e.g. `eth0`, `ens192`).
2. Replace bridge names if your host uses different `onebr-*` naming; they must match Linux bridges with correct VLAN tagging.
3. Import via **Sunstone** / **onevnet create** / **onevnet update**, or merge into your Infra-as-Code pipeline.
4. Attach a **Virtual Router (VR)** to each `100.64.x.0/24`: reserve **`.1`** for the VR (see matrix). **`192.168.86.0/24`**: reserve **`.1`** ISR, **`.2`** VR per matrix examples.

## Files

| File | VNET |
|------|------|
| [edge-vnet.one](edge-vnet.one) | `devsecops-edge` VLAN 86 |
| [gitea-vnet.one](gitea-vnet.one) | `devsecops-gitea` VLAN 2001 |
| [gateway-vnet.one](gateway-vnet.one) | `devsecops-gateway` VLAN 2005 |
| [ceph-vnet.one](ceph-vnet.one) | `devsecops-ceph` VLAN 2250 |

## Related

- Refined execution phases: [docs/opennebula-gitea-edge/REFINED-EXECUTION.md](../../docs/opennebula-gitea-edge/REFINED-EXECUTION.md)
