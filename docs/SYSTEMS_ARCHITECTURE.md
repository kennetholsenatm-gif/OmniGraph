# Systems Architecture Document — DevSecOps Pipeline

This document describes the components and deployment order for the Autonomous Zero Trust DevSecOps pipeline (100.64.0.0/10, Varlock, Solace, Keycloak, n8n, Gitea, Zammad).

## Component Overview

| Component | Purpose | Network | Compose / Location |
|-----------|---------|---------|--------------------|
| **HashiCorp Vault** | Secrets store for Varlock (Keycloak admin, DB passwords, API tokens, admin SSH key). KV store at `secret/devsecops`. | 100.64.20.0/24 (iam_net) | docker-compose.iam.yml |
| **Keycloak** | IAM, OIDC for Teleport, admin console. | 100.64.20.0/24 (iam_net) | docker-compose.iam.yml |
| **PostgreSQL (Keycloak)** | Keycloak persistence. | 100.64.20.0/24 (iam_net) | docker-compose.iam.yml |
| **Solace PubSub+** | Messaging backbone (A2A, mTLS). | 100.64.10.0/24 (msg_backbone_net) | docker-compose.messaging.yml |
| **Solace Agent Mesh (SAM)** | Agent mesh over Solace. | msg_backbone_net + agent_mesh_net | docker-compose.messaging.yml |
| **RabbitMQ, Kafka, NiFi, Postgres** | Messaging and data pipeline. | 100.64.10.0/24 | docker-compose.messaging.yml |
| **Gitea, n8n, Zammad** | Tooling (Git, orchestration, ITSM). | 100.64.1/2/3.0/24 | docker-compose.tooling.yml |

## Install and Configure HashiCorp Vault

Vault is part of the IAM stack and provides the secrets backend for Varlock (no secrets in `.env`).

### 1. Install (Docker)

Vault runs as a container in `docker-compose.iam.yml`:

```bash
cd devsecops-pipeline/docker-compose
docker compose -f docker-compose.iam.yml --env-file ../.env up -d
```

This starts Vault (dev mode), Keycloak DB, and Keycloak. Vault listens on **http://localhost:8200**.

### 2. Configure (one-time)

- **Dev mode token:** Default root token is `devsecops-dev-root` (set via `VAULT_DEV_ROOT_TOKEN_ID` in the compose or env). Use it for local/dev only.
- **Enable KV v2 and UI:** Run the setup script so the Vault UI at **http://localhost:8200/ui** shows the secrets engine:

  ```powershell
  $env:VAULT_ADDR = "http://localhost:8200"
  $env:VAULT_TOKEN = "devsecops-dev-root"
  .\scripts\setup-vault-ui.ps1
  ```

  Or with the Vault CLI:

  ```bash
  export VAULT_ADDR=http://localhost:8200
  export VAULT_TOKEN=devsecops-dev-root
  vault secrets enable -path=secret kv-v2
  ```

  If the path is already in use (e.g. KV v1), the script will report it and you can still use the UI.
- **Log in to the Vault UI:** Open **http://localhost:8200/ui**, choose **Token** auth, and enter the root token (`devsecops-dev-root`). Then go to **Secrets** → **secret** to view or create secrets (e.g. `secret/devsecops`).

- **Store pipeline secrets:** Run the kbolsen registration script so Vault has Keycloak admin and admin SSH public key:

  ```powershell
  $env:VAULT_ADDR = "http://localhost:8200"
  $env:VAULT_TOKEN = "devsecops-dev-root"
  .\scripts\register-kbolsen-in-vault.ps1
  ```

### 3. Production

For production, do **not** use dev mode. Use a proper Vault deployment (HA, storage backend, audit, and restrict root token). Point `VAULT_ADDR` and `VAULT_TOKEN` (or other auth) at that cluster and use the same `VAULT_SECRET_PATH` (e.g. `secret/devsecops`) and key names from `devsecops.env.schema`.

## Execution Order

See [DEPLOYMENT.md](DEPLOYMENT.md): networks → OpenTofu/Ansible (optional) → Docker Compose (messaging → IAM → tooling). IAM stack includes **Vault** and Keycloak; start it before running the kbolsen script or any workflow that reads from Vault.

## References

- [DEPLOYMENT.md](DEPLOYMENT.md) — Full deployment steps.
- [VARLOCK_USAGE.md](VARLOCK_USAGE.md) — Where credentials live (Vault) and how to inject into Compose.
- [scripts/README-kbolsen-vault.md](../scripts/README-kbolsen-vault.md) — Register admin kbolsen and SSH key in Vault.
