# Autonomous Zero Trust DevSecOps Pipeline — Network Design

## Address Space: 100.64.0.0/10

This document defines the segmented network topology for the pipeline. All tooling runs in dedicated Docker bridge networks with strict isolation; lateral movement between tool subnets is explicitly denied.

## Segment Definition

| Segment            | Subnet            | Docker Network       | Purpose |
|--------------------|-------------------|----------------------|--------|
| Tooling (Gitea)    | 100.64.1.0/24     | `gitea_net`          | Gitea VCS only |
| Tooling (n8n)      | 100.64.2.0/24     | `n8n_net`             | n8n macro-orchestrator only |
| Tooling (Zammad)   | 100.64.3.0/24     | `zammad_net`         | Zammad ITSM only |
| Messaging backbone | 100.64.10.0/24    | `msg_backbone_net`   | Solace PubSub+, NiFi, RabbitMQ, Kafka |
| IAM / edge         | 100.64.20.0/24    | `iam_net`            | Keycloak, Teleport |
| Agent mesh         | 100.64.30.0/24    | `agent_mesh_net`     | SAM, Cline, Critique, Doc agents |

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
   - Teleport: 3023 (Auth), 3024 (Proxy), 3025 (SSH).
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
- **Rich rules**: Allow source 100.64.0.0/10 to ports 55555, 8008, 8080, 3023, 3024, 3025, 5678, 3000, 2222, 80, 443 as needed.
- **Block**: Reject/drop traffic between 100.64.1.0/24, 100.64.2.0/24, 100.64.3.0/24 (no direct tool-to-tool).

## Docker Network Creation

Networks are created with static subnets so that IPs are predictable and UFW/firewalld rules can reference them. Example (Ansible/OpenTofu):

- `gitea_net`: driver bridge, subnet 100.64.1.0/24
- `n8n_net`: driver bridge, subnet 100.64.2.0/24
- `zammad_net`: driver bridge, subnet 100.64.3.0/24
- `msg_backbone_net`: driver bridge, subnet 100.64.10.0/24
- `iam_net`: driver bridge, subnet 100.64.20.0/24
- `agent_mesh_net`: driver bridge, subnet 100.64.30.0/24

Containers that need to reach Solace (e.g. n8n, SAM) are attached to both their tool/agent network and `msg_backbone_net` so they can resolve `solace-pubsub`, `rabbitmq`, `kafka`, etc., without exposing tool networks to each other.
