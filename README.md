# Autonomous Zero Trust DevSecOps Pipeline

Deployment artifacts for the pipeline: 100.64.0.0/10 segmented network, Solace mTLS A2A, n8n macro-orchestrator, Gitea, Zammad, Keycloak, **HashiCorp Vault**, Teleport JIT, and Varlock schema-driven secrets.

## Quick Start

1. Read [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for execution order.
2. Create networks (Ansible or OpenTofu from `opentofu/`).
3. Start Docker Compose stacks from `docker-compose/` using `devsecops.env.schema` for required variables.
4. Import n8n workflows from `n8n-workflows/` and configure credentials.

## Layout

- **docs/** — SYSTEMS_ARCHITECTURE.md (incl. install and configure Vault), NETWORK_DESIGN.md, DEPLOYMENT.md, A2A_PAYLOAD_SCHEMA.md, TELEPORT_JIT.md, VARLOCK_USAGE.md
- **opentofu/** — Docker network definitions (100.64.x.x)
- **ansible/** — Playbooks and roles (mesh deploy, mTLS, FIPS/hardening)
- **docker-compose/** — Messaging, tooling (Gitea/n8n/Zammad), IAM (Vault + Keycloak)
- **n8n-workflows/** — DevSecOps Orchestrator, Security Audit, Documentation Generator (JSON)
- **devsecops.env.schema** — Unified Varlock schema (no secrets)

Solace VPN and A2A ACL are in `qminiwasm-automation/infra/opentofu/` (discovery-networks.tf, devsecops-variables.tf).
