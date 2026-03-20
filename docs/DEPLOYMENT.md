# Autonomous Zero Trust DevSecOps Pipeline — Deployment Guide

## Overview

This directory contains all artifacts to deploy the pipeline on 100.64.0.0/10 with a messaging backbone (Solace mTLS, NiFi, RabbitMQ, Kafka), isolated tooling (Gitea, n8n, Zammad, Bitwarden), IAM (Keycloak), and agent mesh (SAM). All configuration is FOSS-only.

## Prerequisites

- Docker and Docker Compose (or Podman).
- OpenTofu 1.6+ (for Solace broker and optional Docker networks).
- Ansible (for host hardening, mTLS, and mesh deployment).
- Base host: AlmaLinux 9/10 (or RHEL) with FIPS mode desired; or Debian/Ubuntu for UFW-based playbooks.

## Local LXC-first runtime (before OpenNebula)

Run the same **Docker Compose** stacks inside **LXD** containers on **WSL2** (or an AlmaLinux VM) as a stepping stone to **OpenNebula** VMs.

| Doc / path | Purpose |
|------------|---------|
| [deployments/local-lxc/README.md](../deployments/local-lxc/README.md) | Canonical local target: LXC-first runtime, optional nested Docker |
| [deployments/wsl2-lxc/README.md](../deployments/wsl2-lxc/README.md) | WSL2-specific caveats and compatibility notes |
| [scripts/create-networks.sh](../scripts/create-networks.sh) | All **18** Docker bridge networks (run **inside** each LXC or once per Docker host) |
| [scripts/secrets-bootstrap.sh](../scripts/secrets-bootstrap.sh) | Bash twin of **`secrets-bootstrap.ps1`** (Vault KV + optional core compose); break-glass **bw** still via PowerShell |
| [scripts/start-local-lxc.sh](../scripts/start-local-lxc.sh), [scripts/start-local-lxc.ps1](../scripts/start-local-lxc.ps1) | Convenience wrappers to invoke local LXC provisioning playbook |
| [ansible/playbooks/deploy-devsecops-lxc.yml](../ansible/playbooks/deploy-devsecops-lxc.yml) | Provision instances + sync compose trees |
| [WSL2_LXC_MIGRATION.md](WSL2_LXC_MIGRATION.md) | **IAM → messaging → …** order |
| [WSL2_LXC_GATEWAY.md](WSL2_LXC_GATEWAY.md) | **Traefik** across multiple LXCs |
| [opennebula-kvm/VLAN_MATRIX.md](../deployments/opennebula-kvm/VLAN_MATRIX.md) | Hardware addressing when you lift off WSL |
| [opennebula-gitea-edge/REDUCE-DOCKER.md](opennebula-gitea-edge/REDUCE-DOCKER.md) | **Minimize Docker** — native Alma / Podman tiers; compose as legacy |
| [opennebula-gitea-edge/LXC-ALMA10-OPENNEBULA.md](opennebula-gitea-edge/LXC-ALMA10-OPENNEBULA.md) | **AlmaLinux 10 LXC** per stack on OpenNebula (LXD host VM + `deploy-devsecops-lxc.yml`) |
| [opennebula-gitea-edge/CONTAINER-LIFT-TO-OPENNEBULA.md](opennebula-gitea-edge/CONTAINER-LIFT-TO-OPENNEBULA.md) | **Move the Docker stacks** — volumes + networks + Compose (inside LXC or flat VM) |
| [opennebula-gitea-edge/WHOLE-REPO-MIGRATION-SCOPE.md](opennebula-gitea-edge/WHOLE-REPO-MIGRATION-SCOPE.md) | **All repo artifacts** (compose, Vault, n8n, gateway) to update when Gitea leaves Windows |

## Execution modes (supported)

1. **Flat Docker host (legacy/lab):** `ansible/playbooks/site.yml` + `ansible/roles/devsecops_containers`.
2. **OpenNebula LXC (hybrid core):** `ansible/playbooks/deploy-devsecops-lxc.yml` (or `opennebula-hybrid-site.yml`) + `ansible/roles/lxd_devsecops_stack`.
3. **K3s slice (Gitea via Helm):** `ansible/playbooks/opennebula-hybrid-site.yml` second play + `ansible/roles/opennebula_k3s_gitea`.

Reference inventory for hybrid mode: `ansible/inventory/opennebula-hybrid.example.yml`.

## Lean local control plane (recommended for laptop)

Run only control-plane tools on laptop; run full runtime services on OpenNebula:

- Local: Semaphore (**Incus LXC**, native Postgres — run `./scripts/start-semaphore.ps1` / `./scripts/start-semaphore.sh`), Ansible controller, pre-commit, KICS, Inframap, Ansible-CMDB.
- OpenNebula: full DevSecOps runtime stacks.

Bootstrap:

```powershell
./scripts/setup-lean-local-control.ps1
./scripts/start-semaphore.ps1
```

Details: [opennebula-gitea-edge/LEAN_LOCAL_CONTROL_PLANE.md](opennebula-gitea-edge/LEAN_LOCAL_CONTROL_PLANE.md)

After Semaphore is up, keep it aligned with Git: [SEMAPHORE_POPULATE.md](SEMAPHORE_POPULATE.md) — **`ansible/playbooks/sync-semaphore-from-manifest.yml`** (manifest-driven, CI-friendly) or **`ansible/playbooks/populate-semaphore.yml`** (single smoke template). **Gitea Actions** (lint + optional Semaphore sync): [CI_CD.md](CI_CD.md).

```bash
cd ansible
ansible-galaxy collection install -r collections/requirements.yml
ansible-playbook -i inventory/lxc.example.yml playbooks/deploy-devsecops-lxc.yml \
  -e 'lxd_apply_names=["devsecops-iam","devsecops-messaging"]'
```

**Lower CPU (local LXC):** Use `devsecops-dev` with `lxd_apply_names=["devsecops-dev"]` for one LXC running IAM + slim messaging (`docker-compose.messaging.slim.yml`), or set `lxd_messaging_compose_cli='-f docker-compose.messaging.slim.yml'` for the `devsecops-messaging` instance only. See [deployments/local-lxc/README.md](../deployments/local-lxc/README.md).

**WSL + `/mnt/c/`:** `ansible/playbooks` is often **world-writable**; Ansible **ignores** `ansible.cfg` there (see Ansible docs). Use [`ansible/playbooks/run-deploy-devsecops-lxc.sh`](../ansible/playbooks/run-deploy-devsecops-lxc.sh) or export `ANSIBLE_CONFIG` (path to `ansible/ansible.cfg`) and `ANSIBLE_ROLES_PATH` (path to `ansible/roles`) before `ansible-playbook`. Plain `ansible/playbooks/ansible.cfg` is not loaded in that case.

## Execution Order

1. **Network design**  
   Finalize 100.64.x.x subnets and UFW/firewalld rules. See [NETWORK_DESIGN.md](NETWORK_DESIGN.md).

2. **OpenTofu**  
   - **Solace (discovery-networks)**: From your infra repo, apply discovery-networks and devsecops-variables (e.g. `discovery-networks.tf`, `devsecops-variables.tf`) to configure the Solace VPN (mTLS-only) and A2A topics.  
   - **Docker networks (optional)**: From this repo’s `opentofu/`, run `tofu init && tofu apply` to create **18** bridge networks: `gitea_net`, `n8n_net`, `zammad_net`, `bitwarden_net`, `gateway_net`, `portainer_net`, `llm_net`, `chatops_net`, `msg_backbone_net`, `iam_net`, `freeipa_net`, `agent_mesh_net`, `discovery_net`, `sdn_lab_net`, `telemetry_net`, `docs_net`, `sonarqube_net`, `siem_net` (subnets in [NETWORK_DESIGN.md](NETWORK_DESIGN.md)). Docsify: [DOCSIFY_GITEA.md](DOCSIFY_GITEA.md). SonarQube: [SONARQUBE_KEYCLOAK.md](SONARQUBE_KEYCLOAK.md). Wazuh: [WAZUH_SIEM.md](WAZUH_SIEM.md).

3. **Greenfield one-shot (no Ansible):** Use the deployment repo at `C:\GiTeaRepos\Deploy` and run `.\scripts\launch-greenfield.ps1`; enter admin password when prompted. This creates all networks, generates secrets, starts the stack, and writes to Vault. Optionally use `-SaveVaultToken`. Use `-IncludeSdnTelemetry` on **Linux** hosts to add OVS/VyOS + sFlow-RT/Prometheus/Grafana; see [SDN_TELEMETRY.md](SDN_TELEMETRY.md). See [GREENFIELD_ONE_SHOT.md](GREENFIELD_ONE_SHOT.md) (moved pointer).

4. **Ansible (optional)**  
   Run host hardening and mesh deployment:
   ```bash
   cd ansible
   ansible-playbook -i inventory.yml playbooks/site.yml
   ```
   This creates the Docker networks (if not using OpenTofu), UFW rules, mTLS cert distribution, and deploys Gitea and n8n in isolated nets. Place Root CA and client certs in `ansible/files/crypto/` (root_ca.crt, node_client.key, node_client.crt).

5. **Start containers (Ansible injects env; no .env file)**  
   Ensure networks exist (step 2 or 3), then start all stacks with **Ansible** so env is injected into containers from Vault or vault-encrypted vars:
   ```bash
   cd ansible
   # Option A: Secrets from HashiCorp Vault (Vault CLI on controller)
   VAULT_ADDR=http://vault:8200 VAULT_TOKEN=your-token ansible-playbook -i inventory.yml playbooks/start-containers-with-vault.yml
   # Option B: Secrets in Ansible Vault (group_vars). Create devsecops_secrets.yml from devsecops_secrets.yml.example, encrypt with ansible-vault encrypt, then:
   ansible-playbook -i inventory.yml playbooks/site.yml
   ```
   The **devsecops_containers** role starts messaging, IAM, tooling, **ChatOps (Zulip)**, optional discovery/LLM/AI stacks (off by default), optional SDN + telemetry, and Single Pane of Glass; all env is passed by Ansible at container start. Do **not** maintain a manual `.env` file. See [VARLOCK_USAGE.md](VARLOCK_USAGE.md) and `ansible/roles/devsecops_containers/README.md`. Canonical compose file sets for PowerShell are in `docker-compose/stack-manifest.json` (run `scripts/verify-stack-manifest.ps1` after edits).
   The **IAM stack** includes **HashiCorp Vault** (port 8200), **Keycloak**, and **keycloak-proxy** (port 8180). Use **http://127.0.0.1:8180** for the admin console so redirects stay on the same URL (proxy forwards your Host header; avoid `localhost` if you see redirects to port 80). Install and configure Vault per [SYSTEMS_ARCHITECTURE.md](SYSTEMS_ARCHITECTURE.md#install-and-configure-hashicorp-vault) (enable KV v2, then run `scripts/register-kbolsen-in-vault.ps1` with `VAULT_ADDR=http://localhost:8200` and `VAULT_TOKEN=devsecops-dev-root`).

   **Credentials (Varlock):** Do not put secrets in `.env`. With Varlock, all sensitive values are in **Vault** at `VAULT_SECRET_PATH` (default `secret/devsecops`). For **greenfield with no static secrets**, run the secrets creator once (or on each host startup):

   ```powershell
   cd C:\GiTeaRepos\Deploy\scripts
   .\secrets-bootstrap.ps1
   ```

   This generates strong random passwords and tokens, sets them in the process environment, starts the **merged core stack** (IAM, messaging, tooling, ChatOps — see `docker-compose/stack-manifest.json`), then writes the same secrets to Vault. By default **no `docker-compose/.env` file** is written (zero-disk); use `-WriteEnvFile` only if you need a local env file for ad-hoc `docker compose` from another shell. Keycloak admin login: use the username passed (default `admin`) and the generated password (in Vault at `secret/devsecops` for this session). To only inject into an already-running Vault: `.\secrets-bootstrap.ps1 -StartStack:$false` (env vars must already be set or script will generate and set them, then push to Vault).

   **Alternative (no Ansible):** Export variables from Vault into the environment, then run `.\launch-stack.ps1` from `docker-compose\` (same core stack as bootstrap; optional LLM by default, optional discovery/AI/identity via switches or `DEVSECOPS_INCLUDE_*` env — see script help). Prefer Ansible so env is injected inside the containers by the playbook. See [VARLOCK_USAGE.md](VARLOCK_USAGE.md).

   **Verify stack:** Run `docker ps` (and `docker ps -a` to see exited/restarting). All services should show `Up` and `(healthy)` where healthchecks exist. If any service is missing or restarting, see Troubleshooting below.

6. **n8n**  
   Import the workflow JSONs from `n8n-workflows/` (DevSecOps Orchestrator, Security Audit, Documentation Generator, **Certificate Registration Request**, **Certificate Registration Approved**, **Gitea Push → SBOM/Trivy Scan → Zammad**, **sFlow anomaly → Zulip**). In n8n, create credentials for Gitea API, Zammad API, Teleport API, and Solace MQTT; reference them in the workflows. Configure webhooks in Zammad and Gitea to point to your n8n instance (e.g. `http://n8n:5678/webhook/zammad-ticket`, `http://n8n:5678/webhook/cert-registration-approved`, `http://n8n:5678/webhook/pull-request`, `http://n8n:5678/webhook/gitea-push`). For **SDN / sFlow-RT anomaly → Zulip**, use `http://n8n:5678/webhook/sflow-anomaly` (or via Traefik: `http://<gateway>/n8n/webhook/sflow-anomaly`); see [SDN_TELEMETRY.md](SDN_TELEMETRY.md) (Phase 5 — n8n webhook) and [n8n-workflows/README.md](../n8n-workflows/README.md). For the **Gitea push** webhook (code push → SBOM/Trivy scan → Zammad ticket on CRITICAL vulns), use URL `http://n8n:5678/webhook/gitea-push` and optionally set a secret in Gitea for HMAC verification. See [AUTOMATION_PHASE3.md](AUTOMATION_PHASE3.md) for webhook setup and Execute Command requirements. For **certificate-based self-registration**, see [CERT_SELF_REGISTRATION.md](CERT_SELF_REGISTRATION.md): users POST to `webhook/cert-registration-request` with cert metadata; a Zammad ticket is created and **no privileges are granted** until a human approves the ticket; on approval, Zammad triggers `cert-registration-approved` and n8n grants Keycloak user and role.

7. **Teleport JIT**  
   Implement the JIT flow (Teleport Access Request API or custom sidecar). See [TELEPORT_JIT.md](TELEPORT_JIT.md). Point `TELEPORT_JIT_REQUEST_URL` and n8n’s “Request JIT Access” node to the chosen endpoint.

8. **Varlock**  
   Validate and use the unified schema: [VARLOCK_USAGE.md](VARLOCK_USAGE.md). Ensure all services and agents read only from environment variables populated from `devsecops.env.schema`; never expose raw secrets to AI context or workflow JSON.

9. **Phase 4 Data Fabric (optional)**  
   To log pipeline events from NiFi and Kafka into Postgres (pgvector), see [DATA_FABRIC_PHASE4.md](DATA_FABRIC_PHASE4.md). Ensure `pipeline.pipeline_events` exists (run [docker-compose/init-scripts/init-pipeline-events.sql](docker-compose/init-scripts/init-pipeline-events.sql) once if Postgres was created before the init mount). NiFi receives `POSTGRES_*` env from the messaging compose; register the JDBC Sink connector for Kafka via the Kafka Connect REST API if using `kafka-connect`.

## Artifacts Summary

| Artifact | Location |
|----------|----------|
| **Greenfield one-shot launch** (clone, one command, one password) | `C:\GiTeaRepos\Deploy\docs\GREENFIELD_ONE_SHOT.md`, `C:\GiTeaRepos\Deploy\scripts\launch-greenfield.ps1` |
| **Systems Architecture** (incl. install and configure HashiCorp Vault) | `docs/SYSTEMS_ARCHITECTURE.md` |
| Network design | `docs/NETWORK_DESIGN.md` |
| Unified env schema | `devsecops.env.schema` |
| OpenTofu (Solace + vars) | Your infra repo (discovery-networks.tf, devsecops-variables.tf) |
| OpenTofu (Docker networks) | `opentofu/` |
| Discovery / BOM (NetBox, Dep-Track) | `docs/ADR_FULLSTACK_DISCOVERY.md`, `docker-compose/docker-compose.discovery.yml`, `ansible/playbooks/deploy-fullstack-discovery.yml` |
| NetBox → Termius | `scripts/sync_netbox_to_termius.py`, `docs/NETBOX_TERMIUS_SYNC.md`, `n8n-workflows/netbox-to-termius-sync.json` |
| Solace discovery naming | `docs/SOLACE_DISCOVERY_QUEUES.md`, `docs/snippets/solace-discovery-queues.tf` |
| Wiki-oriented BOM chunks | Generate under `docs/WIKI_EXPORT/` (directory gitignored; keep exports local) |
| Ansible playbooks | `ansible/playbooks/` (site.yml, deploy-devsecops-mesh.yml) |
| Ansible roles | `ansible/roles/` (mtls_distribution, os_hardening_fips, devsecops_containers) |
| Docker Compose | `docker-compose/` (messaging, tooling, IAM, ChatOps; **stack-manifest.json** for merged core; optional discovery, LLM, AI) |
| n8n workflow JSON | `n8n-workflows/*.json` |
| A2A payload schema | `docs/A2A_PAYLOAD_SCHEMA.md` |
| Teleport JIT | `docs/TELEPORT_JIT.md` |
| Cert-based self-registration (human approval via Zammad) | `docs/CERT_SELF_REGISTRATION.md`, `docs/CERT_REGISTRATION_PAYLOAD_SCHEMA.md`, `n8n-workflows/cert-registration-request.json`, `n8n-workflows/cert-registration-approved.json` |
| Automation Phase 3 (Gitea push → SBOM/Trivy → Zammad) | `docs/AUTOMATION_PHASE3.md`, `n8n-workflows/gitea-push-sbom-scan.json` |
| ChatOps: Dependency-Track → Zulip | `n8n-workflows/dependency-track-to-zulip.json`; webhook `http://n8n:5678/webhook/dependency-track-alert`; create Zulip API credential (HTTP Basic Auth: email + API key) and stream (e.g. `security`) |
| AI Orchestration (code-server, MCP, n8n-mcp, Solace config) | `ai-orchestration/` (Phase 1–5 scripts, Solace topic routing and agent cards); optional `docker-compose/docker-compose.ai-orchestration.yml` for n8n-mcp |
| Data Fabric Phase 4 (NiFi/Kafka → Postgres pgvector) | `docs/DATA_FABRIC_PHASE4.md`, `docker-compose/init-scripts/init-pipeline-events.sql`, `docker-compose/kafka-connect/jdbc-sink-pipeline-events.json` |
| Varlock usage | `docs/VARLOCK_USAGE.md` |
| Identity & privilege (greenfield, VARLOCK + LDAP-friendly) | `docs/IDENTITIES_AND_PRIVILEGES.md`, `devsecops.identities.schema`, `privilege_levels.json`, `identities.example.yaml`; `scripts/sync-identities-to-keycloak.ps1`, `scripts/export-identities-to-ldif.ps1` |
| Secrets bootstrap (greenfield, no static storage) | `scripts/secrets-bootstrap.ps1`, `scripts/secrets-bootstrap.sh` (Linux/LXC) |
| Local LXC-first runtime (+ nested Docker where required) | `deployments/local-lxc/`, `ansible/playbooks/deploy-devsecops-lxc.yml`, `docs/WSL2_LXC_MIGRATION.md` |
| Compose stack manifest (merged core file list) | `docker-compose/stack-manifest.json`, `scripts/verify-stack-manifest.ps1`, `scripts/verify-lxd-manifest-parity.ps1` |
| Lean local control plane | `docs/opennebula-gitea-edge/LEAN_LOCAL_CONTROL_PLANE.md`, `scripts/setup-lean-local-control.ps1`, `deployments/local-control/semaphore/docker-compose.yml` |
| Repo ownership split policy | `docs/REPO_SCOPE.md` (`devsecops-pipeline`) and `C:\GiTeaRepos\Deploy\docs\REPO_SCOPE.md` (`deploy`) |
| KICS scan (IaC security) | `scripts/scan-kics.ps1`, `scripts/scan-kics.sh` |
| Inframap topology export | `scripts/generate-inframap.ps1`, `scripts/generate-inframap.sh` |
| Ansible-CMDB report | `scripts/generate-ansible-cmdb.ps1`, `scripts/generate-ansible-cmdb.sh` |
| Pre-commit IaC gates | `.pre-commit-config.yaml` |
| Greenfield registration (initial values without static files) | `docs/GREENFIELD_REGISTRATION.md`; `scripts/start-from-vault.ps1`, `scripts/save-vault-token-to-keystore.ps1` |
| KMS for license/activation keys (e.g. n8n) | `docs/LICENSE_KEYS_KMS.md`; `scripts/store-license-keys-in-vault.ps1`; keys in Vault at `secret/devsecops`, injected at runtime |

## Troubleshooting

| Issue | Fix |
|-------|-----|
| **Kafka** exits with `KAFKA_PROCESS_ROLES is not set` | Compose uses Confluent Kafka 7.5.3 (Zookeeper mode). The `latest` image requires KRaft; pin to `confluentinc/cp-kafka:7.5.3` and `confluentinc/cp-zookeeper:7.5.3` in `docker-compose.messaging.yml`. |
| **Keycloak** exits with “don't use --optimized for first ever server start” | Use `command: ["start"]` (no `--optimized`) for the first run. After the first successful boot, you can switch to `["start", "--optimized"]` for faster restarts. |
| **Keycloak** exits with “Key material not provided to setup HTTPS” / “see the http-enabled option” | Set `KC_HTTP_ENABLED: "true"` in the Keycloak service environment (compose has this for dev; use TLS in production). |
| **Keycloak** / Hibernate: `password authentication failed for user "keycloak"` / failed JDBC connection | Postgres stores the `keycloak` user password in volume `keycloak_db_data` **only on first init** (`POSTGRES_PASSWORD` = `KEYCLOAK_DB_PASSWORD` at that time). Changing Vault/env later does not update the DB. **Fix A:** Start compose with the **same** `KEYCLOAK_DB_PASSWORD` as when the volume was created (from `secret/devsecops` via `start-from-vault.ps1` or your saved env). **Fix B (destructive):** stop IAM services, remove the Keycloak DB volume (`docker volume ls` → e.g. `*_keycloak_db_data`), then greenfield/bootstrap again so Postgres and Keycloak share one new password. Avoid re-running `secrets-bootstrap` in a fresh shell without loading Vault if the DB volume already exists—it may regenerate `KEYCLOAK_DB_PASSWORD` and mismatch the DB. |
| **Keycloak** shows “We are sorry... HTTPS required”, **HTTP 400**, or redirect to **localhost:80** / **-102** | Use **http://127.0.0.1:8180** (not localhost). The proxy forwards `X-Forwarded-Host` from your request so redirects use the same host:port. Set master realm SSL to “none” once: `docker exec devsecops-keycloak-db psql -U keycloak -d keycloak -c "UPDATE realm SET ssl_required = 'NONE' WHERE name = 'master';"` then restart Keycloak. |
| **Solace** shows `(unhealthy)` | The standard image may not expose `/health-check/direct-active` on port 8080. Compose uses a healthcheck that accepts HTTP 200/401/404 on `http://localhost:8080/`. If still unhealthy, ensure `start_period` is sufficient (e.g. 120s) or use `condition: service_started` for SAM. |
| **SAM** stays `Created` | SAM depends on Solace being healthy. Once Solace is healthy, run `docker compose -f docker-compose.messaging.yml up -d solace-agent-mesh` or restart the messaging stack. |
| **Tooling not running** | Start with `docker compose -f docker-compose.tooling.yml --env-file ../.env up -d`. Ensure required external networks exist ([NETWORK_DESIGN.md](NETWORK_DESIGN.md); `create-networks.ps1` creates **16**, including **`docs_net`**). |
| **ChatOps (Zulip) missing after Vault restart** | Use an updated `launch-stack.ps1`: core stack is merged from `stack-manifest.json` (includes ChatOps). Run `.\scripts\verify-stack-manifest.ps1` if you forked compose files. |
| **`network docs_net declared as external, but could not be found`** | Tooling and gateway expect a pre-created **`docs_net`** (subnet `100.64.52.0/24`). From repo root run **`.\scripts\create-networks.ps1`** (idempotent; adds any missing nets), or: `docker network create --driver bridge --subnet 100.64.52.0/24 docs_net`. Common after a repo update if networks were created before `docs_net` was added. |
| **ERR_CONNECTION_REFUSED** on `http://localhost:8180` | Ensure Docker Desktop is running. Try **http://127.0.0.1:8180** (some setups resolve `localhost` differently). Restart the IAM stack: from `docker-compose` run `.\launch-stack.ps1` (with repo-root `.env` or env vars from Vault set). If the proxy and Keycloak show `Up` in `docker ps` but the host still can’t connect, restart Docker Desktop to restore port forwarding. |

## Optional Next Steps

- **Teleport JIT:** Wire the JIT flow per [TELEPORT_JIT.md](TELEPORT_JIT.md); set `TELEPORT_JIT_REQUEST_URL` and n8n credentials.
- **n8n:** Import workflow JSONs from `n8n-workflows/`, create credentials (Gitea, Zammad, Teleport, Solace), and configure webhooks in Zammad/Gitea to n8n.
- **Secrets:** With Varlock, store all secrets in Vault at `VAULT_SECRET_PATH`; do not store them in `.env`. See [VARLOCK_USAGE.md](VARLOCK_USAGE.md) for injection and credential locations.
- **Packer / AlmaLinux 10 + FIPS:** Add image build with Packer and FIPS hardening if the pipeline plan calls for it; align with `ansible/roles/os_hardening_fips`.

## 1-bit LLM (Optional)

If you have a local 1-bit LLM (e.g. from C:\HF1BitLLM or BitNet), set `LLM_SERVICE_ENDPOINT` (and related vars) in your `.env` and in the messaging stack for SAM. Otherwise leave empty; SAM and n8n agents can use another LLM backend.
