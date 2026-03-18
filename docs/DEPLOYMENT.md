# Autonomous Zero Trust DevSecOps Pipeline — Deployment Guide

## Overview

This directory contains all artifacts to deploy the pipeline on 100.64.0.0/10 with a messaging backbone (Solace mTLS, NiFi, RabbitMQ, Kafka), isolated tooling (Gitea, n8n, Zammad), IAM (Keycloak), and agent mesh (SAM). All configuration is FOSS-only.

## Prerequisites

- Docker and Docker Compose (or Podman).
- OpenTofu 1.6+ (for Solace broker and optional Docker networks).
- Ansible (for host hardening, mTLS, and mesh deployment).
- Base host: AlmaLinux 9/10 (or RHEL) with FIPS mode desired; or Debian/Ubuntu for UFW-based playbooks.

## Execution Order

1. **Network design**  
   Finalize 100.64.x.x subnets and UFW/firewalld rules. See [NETWORK_DESIGN.md](NETWORK_DESIGN.md).

2. **OpenTofu**  
   - **Solace (discovery-networks)**: From `qminiwasm-automation/infra/opentofu`, apply `discovery-networks.tf` and `devsecops-variables.tf` to configure the Solace VPN (mTLS-only) and A2A topics.  
   - **Docker networks (optional)**: From `devsecops-pipeline/opentofu`, run `tofu init && tofu apply` to create the six bridge networks (gitea_net, n8n_net, zammad_net, msg_backbone_net, iam_net, agent_mesh_net).

3. **Ansible**  
   Run host hardening and mesh deployment:
   ```bash
   cd devsecops-pipeline/ansible
   ansible-playbook -i inventory.yml playbooks/site.yml
   ```
   This creates the Docker networks (if not using OpenTofu), UFW rules, mTLS cert distribution, and deploys Gitea and n8n in isolated nets. Place Root CA and client certs in `ansible/files/crypto/` (root_ca.crt, node_client.key, node_client.crt).

4. **Docker Compose**  
   Create networks if not already present, then start stacks in order:
   ```bash
   cd devsecops-pipeline/docker-compose
   # Ensure networks exist (or use Ansible/OpenTofu). On Windows PowerShell use 2>$null instead of 2>/dev/null.
   docker network create --driver bridge --subnet 100.64.1.0/24 gitea_net 2>/dev/null || true
   docker network create --driver bridge --subnet 100.64.2.0/24 n8n_net 2>/dev/null || true
   docker network create --driver bridge --subnet 100.64.3.0/24 zammad_net 2>/dev/null || true
   docker network create --driver bridge --subnet 100.64.10.0/24 msg_backbone_net 2>/dev/null || true
   docker network create --driver bridge --subnet 100.64.20.0/24 iam_net 2>/dev/null || true
   docker network create --driver bridge --subnet 100.64.30.0/24 agent_mesh_net 2>/dev/null || true
   docker compose -f docker-compose.messaging.yml --env-file ../.env up -d
   docker compose -f docker-compose.iam.yml --env-file ../.env up -d
   docker compose -f docker-compose.tooling.yml --env-file ../.env up -d
   ```
   The **IAM stack** includes **HashiCorp Vault** (port 8200), **Keycloak**, and **keycloak-proxy** (port 8180). Use **http://127.0.0.1:8180** for the admin console so redirects stay on the same URL (proxy forwards your Host header; avoid `localhost` if you see redirects to port 80). Install and configure Vault per [SYSTEMS_ARCHITECTURE.md](SYSTEMS_ARCHITECTURE.md#install-and-configure-hashicorp-vault) (enable KV v2, then run `scripts/register-kbolsen-in-vault.ps1` with `VAULT_ADDR=http://localhost:8200` and `VAULT_TOKEN=devsecops-dev-root`).

   **Credentials (Varlock):** Do not put secrets in `.env`. With Varlock, all sensitive values are in **Vault** at `VAULT_SECRET_PATH` (default `secret/devsecops`). For **greenfield with no static secrets**, run the secrets creator once (or on each host startup):

   ```powershell
   cd devsecops-pipeline\scripts
   .\secrets-bootstrap.ps1
   ```

   This generates strong random passwords and tokens, sets them in the process environment, starts the full stack (IAM → messaging → tooling), then writes the same secrets to Vault. **Nothing is written to disk**; secrets exist only in memory and in Vault. Keycloak admin login: use the username passed (default `admin`) and the generated password (in Vault at `secret/devsecops` for this session). To only inject into an already-running Vault: `.\secrets-bootstrap.ps1 -StartStack:$false` (env vars must already be set or script will generate and set them, then push to Vault).

   **Alternative:** Export variables from Vault into the environment, then run `.\launch-stack.ps1` from `docker-compose\`. See [VARLOCK_USAGE.md](VARLOCK_USAGE.md).

   **Verify stack:** Run `docker ps` (and `docker ps -a` to see exited/restarting). All services should show `Up` and `(healthy)` where healthchecks exist. If any service is missing or restarting, see Troubleshooting below.

5. **n8n**  
   Import the workflow JSONs from `devsecops-pipeline/n8n-workflows/` (DevSecOps Orchestrator, Security Audit, Documentation Generator, **Certificate Registration Request**, **Certificate Registration Approved**). In n8n, create credentials for Gitea API, Zammad API, Teleport API, and Solace MQTT; reference them in the workflows. Configure webhooks in Zammad and Gitea to point to your n8n instance (e.g. `http://n8n:5678/webhook/zammad-ticket`, `http://n8n:5678/webhook/cert-registration-approved`, `http://n8n:5678/webhook/pull-request`). For **certificate-based self-registration**, see [CERT_SELF_REGISTRATION.md](CERT_SELF_REGISTRATION.md): users POST to `webhook/cert-registration-request` with cert metadata; a Zammad ticket is created and **no privileges are granted** until a human approves the ticket; on approval, Zammad triggers `cert-registration-approved` and n8n grants Keycloak user and role.

6. **Teleport JIT**  
   Implement the JIT flow (Teleport Access Request API or custom sidecar). See [TELEPORT_JIT.md](TELEPORT_JIT.md). Point `TELEPORT_JIT_REQUEST_URL` and n8n’s “Request JIT Access” node to the chosen endpoint.

7. **Varlock**  
   Validate and use the unified schema: [VARLOCK_USAGE.md](VARLOCK_USAGE.md). Ensure all services and agents read only from environment variables populated from `devsecops.env.schema`; never expose raw secrets to AI context or workflow JSON.

## Artifacts Summary

| Artifact | Location |
|----------|----------|
| **Systems Architecture** (incl. install and configure HashiCorp Vault) | `docs/SYSTEMS_ARCHITECTURE.md` |
| Network design | `docs/NETWORK_DESIGN.md` |
| Unified env schema | `devsecops.env.schema` |
| OpenTofu (Solace + vars) | `qminiwasm-automation/infra/opentofu/discovery-networks.tf`, `devsecops-variables.tf` |
| OpenTofu (Docker networks) | `devsecops-pipeline/opentofu/` |
| Ansible playbooks | `devsecops-pipeline/ansible/playbooks/` (site.yml, deploy-devsecops-mesh.yml) |
| Ansible roles | `devsecops-pipeline/ansible/roles/` (mtls_distribution, os_hardening_fips) |
| Docker Compose | `devsecops-pipeline/docker-compose/` (messaging, tooling, IAM — IAM includes Vault + Keycloak) |
| n8n workflow JSON | `devsecops-pipeline/n8n-workflows/*.json` |
| A2A payload schema | `docs/A2A_PAYLOAD_SCHEMA.md` |
| Teleport JIT | `docs/TELEPORT_JIT.md` |
| Cert-based self-registration (human approval via Zammad) | `docs/CERT_SELF_REGISTRATION.md`, `docs/CERT_REGISTRATION_PAYLOAD_SCHEMA.md`, `n8n-workflows/cert-registration-request.json`, `n8n-workflows/cert-registration-approved.json` |
| Varlock usage | `docs/VARLOCK_USAGE.md` |
| Identity & privilege (greenfield, VARLOCK + LDAP-friendly) | `docs/IDENTITIES_AND_PRIVILEGES.md`, `devsecops.identities.schema`, `privilege_levels.json`, `identities.example.yaml`; `scripts/sync-identities-to-keycloak.ps1`, `scripts/export-identities-to-ldif.ps1` |
| Secrets bootstrap (greenfield, no static storage) | `scripts/secrets-bootstrap.ps1` |
| Greenfield registration (initial values without static files) | `docs/GREENFIELD_REGISTRATION.md`; `scripts/start-from-vault.ps1`, `scripts/save-vault-token-to-keystore.ps1` |
| KMS for license/activation keys (e.g. n8n) | `docs/LICENSE_KEYS_KMS.md`; `scripts/store-license-keys-in-vault.ps1`; keys in Vault at `secret/devsecops`, injected at runtime |

## Troubleshooting

| Issue | Fix |
|-------|-----|
| **Kafka** exits with `KAFKA_PROCESS_ROLES is not set` | Compose uses Confluent Kafka 7.5.3 (Zookeeper mode). The `latest` image requires KRaft; pin to `confluentinc/cp-kafka:7.5.3` and `confluentinc/cp-zookeeper:7.5.3` in `docker-compose.messaging.yml`. |
| **Keycloak** exits with “don't use --optimized for first ever server start” | Use `command: ["start"]` (no `--optimized`) for the first run. After the first successful boot, you can switch to `["start", "--optimized"]` for faster restarts. |
| **Keycloak** exits with “Key material not provided to setup HTTPS” / “see the http-enabled option” | Set `KC_HTTP_ENABLED: "true"` in the Keycloak service environment (compose has this for dev; use TLS in production). |
| **Keycloak** shows “We are sorry... HTTPS required”, **HTTP 400**, or redirect to **localhost:80** / **-102** | Use **http://127.0.0.1:8180** (not localhost). The proxy forwards `X-Forwarded-Host` from your request so redirects use the same host:port. Set master realm SSL to “none” once: `docker exec devsecops-keycloak-db psql -U keycloak -d keycloak -c "UPDATE realm SET ssl_required = 'NONE' WHERE name = 'master';"` then restart Keycloak. |
| **Solace** shows `(unhealthy)` | The standard image may not expose `/health-check/direct-active` on port 8080. Compose uses a healthcheck that accepts HTTP 200/401/404 on `http://localhost:8080/`. If still unhealthy, ensure `start_period` is sufficient (e.g. 120s) or use `condition: service_started` for SAM. |
| **SAM** stays `Created` | SAM depends on Solace being healthy. Once Solace is healthy, run `docker compose -f docker-compose.messaging.yml up -d solace-agent-mesh` or restart the messaging stack. |
| **Tooling not running** | Start with `docker compose -f docker-compose.tooling.yml --env-file ../.env up -d`. Ensure all six networks exist (see step 4). |
| **ERR_CONNECTION_REFUSED** on `http://localhost:8180` | Ensure Docker Desktop is running. Try **http://127.0.0.1:8180** (some setups resolve `localhost` differently). Restart the IAM stack: from `docker-compose` run `.\launch-stack.ps1` (with `..\.env` or env vars from Vault set). If the proxy and Keycloak show `Up` in `docker ps` but the host still can’t connect, restart Docker Desktop to restore port forwarding. |

## Optional Next Steps

- **Teleport JIT:** Wire the JIT flow per [TELEPORT_JIT.md](TELEPORT_JIT.md); set `TELEPORT_JIT_REQUEST_URL` and n8n credentials.
- **n8n:** Import workflow JSONs from `n8n-workflows/`, create credentials (Gitea, Zammad, Teleport, Solace), and configure webhooks in Zammad/Gitea to n8n.
- **Secrets:** With Varlock, store all secrets in Vault at `VAULT_SECRET_PATH`; do not store them in `.env`. See [VARLOCK_USAGE.md](VARLOCK_USAGE.md) for injection and credential locations.
- **Packer / AlmaLinux 10 + FIPS:** Add image build with Packer and FIPS hardening if the pipeline plan calls for it; align with `ansible/roles/os_hardening_fips`.

## 1-bit LLM (Optional)

If you have a local 1-bit LLM (e.g. from C:\HF1BitLLM or BitNet), set `LLM_SERVICE_ENDPOINT` (and related vars) in your `.env` and in the messaging stack for SAM. Otherwise leave empty; SAM and n8n agents can use another LLM backend.
