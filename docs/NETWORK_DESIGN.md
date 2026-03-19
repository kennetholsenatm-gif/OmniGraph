# Autonomous Zero Trust DevSecOps Pipeline — Network Design

## Address Space: 100.64.0.0/10

This document defines the segmented network topology for the pipeline. All tooling runs in dedicated Docker bridge networks with strict isolation; lateral movement between tool subnets is explicitly denied.

## Segment Definition

| Segment            | Subnet            | Docker Network       | Purpose |
|--------------------|-------------------|----------------------|--------|
| Tooling (Gitea)    | 100.64.1.0/24     | `gitea_net`          | Gitea VCS only |
| Tooling (n8n)      | 100.64.2.0/24     | `n8n_net`            | n8n macro-orchestrator only |
| Tooling (Zammad)   | 100.64.3.0/24     | `zammad_net`         | Zammad ITSM only |
| Tooling (Bitwarden)| 100.64.4.0/24     | `bitwarden_net`      | Vaultwarden only |
| Gateway (Traefik)  | 100.64.5.0/24     | `gateway_net`        | Single pane of glass ingress |
| Tooling (Portainer)| 100.64.6.0/24     | `portainer_net`      | Portainer UI |
| LLM / inference    | 100.64.7.0/24     | `llm_net`            | BitNet / optional LLM gateway |
| ChatOps            | 100.64.8.0/24     | `chatops_net`        | Zulip (and related ChatOps) |
| Messaging backbone | 100.64.10.0/24    | `msg_backbone_net`   | Solace PubSub+, NiFi, RabbitMQ, Kafka |
| IAM / edge         | 100.64.20.0/24    | `iam_net`            | Vault, Keycloak, Teleport |
| FreeIPA (optional) | 100.64.21.0/24    | `freeipa_net`        | LDAP/Kerberos when not using Keycloak-only |
| Agent mesh         | 100.64.30.0/24    | `agent_mesh_net`     | SAM, Cline, Critique, Doc agents |
| Discovery / BOM    | 100.64.40.0/24    | `discovery_net`      | NetBox, NetDISCO, Dependency-Track, scanners |
| SDN lab            | 100.64.50.0/24    | `sdn_lab_net`        | VyOS lab leg, sFlow exporters, **n8n** second attachment |
| Telemetry          | 100.64.51.0/24    | `telemetry_net`      | sFlow-RT, Prometheus, Grafana; Traefik joins for `/grafana`, `/sflow-rt` |
| Docs (Docsify)     | 100.64.52.0/24    | `docs_net`           | Nginx serving cloned architecture-docs; Traefik joins for `/docs` |
| SonarQube          | 100.64.53.0/24    | `sonarqube_net`      | SAST; Traefik joins for `/sonarqube`; also attaches to `msg_backbone_net` for JDBC |
| SIEM (Wazuh)       | 100.64.54.0/24    | `siem_net`           | Wazuh manager/indexer/dashboard; Traefik joins for `/wazuh` |

**Canonical count:** **18** Docker bridge networks (tooling, gateway, messaging, IAM, FreeIPA, agent mesh, discovery, SDN, telemetry, docs, SonarQube, SIEM). Create them with `scripts/create-networks.ps1`, OpenTofu (`opentofu/`), or `ansible/playbooks/deploy-devsecops-mesh.yml` — all use the same subnets above. SDN/telemetry details: [SDN_TELEMETRY.md](SDN_TELEMETRY.md).

## Isolation Rules

- **No shared bridge** between Gitea, n8n, and Zammad. Each service runs in its own Docker network.
- **Communication paths**:
  - n8n receives webhooks (via reverse proxy or host mapping) from Zammad and Gitea.
  - n8n and agents publish/subscribe via **Solace** (and optionally RabbitMQ/Kafka) on the messaging backbone.
  - JIT and SSO go through **Teleport** and **Keycloak** on the IAM network.
- **Lateral traversal**: No direct IP access between 100.64.1.0/24, 100.64.2.0/24, and 100.64.3.0/24. Traffic between tooling and messaging/IAM is allowed only on approved ports (see UFW below).

## UFW Rules (Host-Level)

When the host uses UFW (e.g. Debian/Ubuntu):

1. **Default policy**: `ufw default deny incoming`, `ufw default allow outgoing`.
2. **Allow from 100.64.0.0/10** (internal) only for:
   - Solace: 55555 (SMF), 8008 (WebSocket), 8080 (SEMP) — or 8883 if TLS termination is on host.
   - Teleport: 3023 (SSH proxy), 3024 (tunnel), 3025 (auth), **3080** (web UI).
   - Discovery (from `discovery_net` or ops hosts): **8000** (NetBox), **8081** (Dependency-Track bundled), **161/udp** (SNMP for NetDISCO).
   - Keycloak: 8080 (HTTP) or 8443 (HTTPS).
   - n8n webhook ingress: 5678 (or the port exposed to the reverse proxy).
   - Gitea: 3000, 2222 (SSH).
   - Zammad: 80, 443 (or exposed port).
3. **Explicit deny** (optional but recommended): Drop traffic from 100.64.1.0/24 to 100.64.2.0/24 and 100.64.3.0/24; from 100.64.2.0/24 to 100.64.1.0/24 and 100.64.3.0/24; from 100.64.3.0/24 to 100.64.1.0/24 and 100.64.2.0/24. (Tooling segments cannot talk directly to each other.)
4. **Allow** messaging backbone (100.64.10.0/24) to be reached from n8n_net and agent_mesh_net for Solace/RabbitMQ/Kafka client connections.
5. **Allow** IAM (100.64.20.0/24) to be reached from n8n_net and agent_mesh_net for Teleport and Keycloak.

## Firewalld (RHEL/AlmaLinux)

On RHEL/AlmaLinux use firewalld instead of UFW. Equivalent rules:

- **Default zone**: public; default target: DROP for incoming.
- **Rich rules**: Allow source 100.64.0.0/10 to ports 55555, 8008, 8080, 3023, 3024, 3025, 3080, 8000, 8081, 5678, 3000, 2222, 80, 443 as needed (plus **161/udp** from `discovery_net` if using NetDISCO).
- **Block**: Reject/drop traffic between 100.64.1.0/24, 100.64.2.0/24, 100.64.3.0/24 (no direct tool-to-tool).

## Docker Network Creation

Networks are created with static subnets so that IPs are predictable and UFW/firewalld rules can reference them. Example (Ansible/OpenTofu/`create-networks.ps1`):

- `gitea_net`: 100.64.1.0/24 · `n8n_net`: 100.64.2.0/24 · `zammad_net`: 100.64.3.0/24 · `bitwarden_net`: 100.64.4.0/24 · `gateway_net`: 100.64.5.0/24 · `portainer_net`: 100.64.6.0/24 · `llm_net`: 100.64.7.0/24 · `chatops_net`: 100.64.8.0/24 · `msg_backbone_net`: 100.64.10.0/24 · `iam_net`: 100.64.20.0/24 · `freeipa_net`: 100.64.21.0/24 · `agent_mesh_net`: 100.64.30.0/24 · `discovery_net`: 100.64.40.0/24 · `sdn_lab_net`: 100.64.50.0/24 · `telemetry_net`: 100.64.51.0/24 · `docs_net`: 100.64.52.0/24 · `sonarqube_net`: 100.64.53.0/24 · `siem_net`: 100.64.54.0/24

Containers that need to reach Solace (e.g. n8n, SAM) are attached to both their tool/agent network and `msg_backbone_net` so they can resolve `solace-pubsub`, `rabbitmq`, `kafka`, etc., without exposing tool networks to each other.
