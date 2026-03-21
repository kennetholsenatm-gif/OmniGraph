# Collapsed identity / NAC control plane (optional)

This document describes a **simpler network posture** than strict per-service Docker isolation for components that **must talk to each other constantly**: **Keycloak**, **Vault**, **LDAP (FreeIPA)**, **RADIUS (FreeRADIUS)**, **PacketFence**, and related admin APIs.

It does **not** replace the canonical segment table in [NETWORK_DESIGN.md](NETWORK_DESIGN.md) for CI/CD tooling (Gitea, n8n, Zammad, messaging, etc.). Those segments still benefit from **lateral-movement** controls because they are not all mutual dependencies.

## Rationale

- **Diminishing returns:** Putting Keycloak, LDAP, RADIUS, and NAC on **separate Docker bridges** adds operational cost (extra attaches, DNS, firewall rules) without a large security gain **if** every automation path already needs **east–west** access among them.
- **Real isolation target:** Segment the **control plane** as a whole from **end-user VLANs**, **guest Wi‑Fi**, and **Internet ingress**, not from itself.
- **Two mini PCs:** **Mini-PC-IAM** and **Mini-PC-Network** can still be **different machines**. “Same network” here means a **single security / routing zone** (one IPv4 prefix or trusted L3 path), not necessarily one Docker bridge spanning two hosts.

## Recommended zone model

| Concept | Suggested approach |
|---------|-------------------|
| **Logical zone** | **Identity / access control plane** — Vault, Keycloak, Teleport (if colocated), FreeIPA, FreeRADIUS, PacketFence management/RADIUS paths. |
| **IPv4** | **Use `100.64.20.0/24` (`iam_net`)** for Vault, Keycloak, Teleport, and optional **FreeIPA** on a single Docker host. On **two mini PCs**, use one **policy zone** (same prefix or trusted L3 path) for IdM ↔ NAC. |
| **Docker (single host)** | **FreeIPA** is attached to **`iam_net`** in [docker-compose.identity.yml](../docker-compose/docker-compose.identity.yml) (no `freeipa_net`). |
| **FreeRADIUS** | Today [deploy-freeradius-native.yml](../ansible/playbooks/deploy-freeradius-native.yml) installs **OS packages** on `freeradius_servers`; place that host/LXC in the **same zone** (same site prefix or routed `/24` with permissive **intra-zone** ACL). |
| **PacketFence** | Runs on **Mini-PC-Network** per [EDGE-MINI-PC-VYOS-PACKETFENCE.md](opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md). Ensure **L3 reachability** from PacketFence to LDAP/RADIUS/Keycloak (static route or shared VLAN toward **Mini-PC-IAM**). Treat **PF ↔ IdM** as **same trust zone** in firewall policy even if subnets differ. |
| **What stays split** | **Tooling** (`gitea_net`, `n8n_net`, …), **messaging**, **discovery**, **SIEM**: keep documented separation unless you have a reason to collapse further. |

## Cross-host “same network”

Docker **cannot** span two bare-metal hosts without an **overlay** (VXLAN, etc.). For **two mini PCs**, “same network” practically means:

1. **Routing:** Both hosts have interfaces or routes so **Mini-PC-Network** reaches **Mini-PC-IAM**’s `100.64.20.0/24` (or your chosen zone prefix) and **reverse** for callbacks if needed.
2. **Policy:** Firewall rules allow **that zone** to flow **LDAP (389/636), RADIUS (1812/1813), HTTPS** between PF, FreeRADIUS, Keycloak, Vault, and IPA — **without** exposing those ports to untrusted VLANs.

## Repo status

- **Compose:** [docker-compose.iam.yml](../docker-compose/docker-compose.iam.yml) and [docker-compose.identity.yml](../docker-compose/docker-compose.identity.yml) both use **`iam_net`** for IAM-plane containers (Vault, Keycloak, Teleport, optional FreeIPA). **`freeipa_net` / `100.64.21.0/24`** are **retired** in this repo—see [NETWORK_DESIGN.md](NETWORK_DESIGN.md) migration note.
- **PacketFence** remains **outside** these compose files; align **routing + firewall** using this doc and [VLAN_MATRIX.md](../deployments/opennebula-kvm/VLAN_MATRIX.md).

## Related

- [NETWORK_DESIGN.md](NETWORK_DESIGN.md) — canonical `100.64` table and isolation rules for the full pipeline.
- [CANONICAL_DEPLOYMENT_VISION.md](CANONICAL_DEPLOYMENT_VISION.md) — Mini-PC-IAM vs Mini-PC-Network roles.
- [EDGE-MINI-PC-VYOS-PACKETFENCE.md](opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md) — PacketFence placement.
