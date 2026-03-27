# Network (100.64)

Summary of `docs/NETWORK_DESIGN.md`. **Canonical count: 17** Docker bridge networks.

## Segment table (abbreviated)

| Segment | Subnet | Docker network |
|---------|--------|----------------|
| Gitea | 100.64.1.0/24 | `gitea_net` |
| n8n | 100.64.2.0/24 | `n8n_net` |
| Zammad | 100.64.3.0/24 | `zammad_net` |
| Bitwarden | 100.64.4.0/24 | `bitwarden_net` |
| Gateway (Traefik) | 100.64.5.0/24 | `gateway_net` |
| Portainer | 100.64.6.0/24 | `portainer_net` |
| LLM | 100.64.7.0/24 | `llm_net` |
| ChatOps | 100.64.8.0/24 | `chatops_net` |
| Messaging | 100.64.10.0/24 | `msg_backbone_net` |
| IAM (+ optional FreeIPA) | 100.64.20.0/24 | `iam_net` |
| Agent mesh | 100.64.30.0/24 | `agent_mesh_net` |
| Discovery | 100.64.40.0/24 | `discovery_net` |
| SDN lab | 100.64.50.0/24 | `sdn_lab_net` |
| Telemetry | 100.64.51.0/24 | `telemetry_net` |
| Docs | 100.64.52.0/24 | `docs_net` |
| SonarQube | 100.64.53.0/24 | `sonarqube_net` |
| SIEM (Wazuh) | 100.64.54.0/24 | `siem_net` |

## Identity plane

Vault, Keycloak, Teleport, and optional **FreeIPA** share **`iam_net`** (no separate `freeipa_net`). For firewall policy, treat IdM / NAC as one **control-plane zone** when useful: `docs/NETWORK_COLLAPSED_IDENTITY_PLANE.md`.

## OpenNebula VLANs

`deployments/opennebula-kvm/VLAN_MATRIX.md` — VLAN `2000 + third octet` rule for `100.64.x.0/24`.

## Full detail

See **`docs/NETWORK_DESIGN.md`** and **`docs/SDN_TELEMETRY.md`**.
