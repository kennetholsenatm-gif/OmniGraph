# Greenfield One-Shot Launch

Clone the repo, run one command, enter your admin password once, walk away. When you come back, the infrastructure is fully stood up.

## Target experience

1. **Clone** the repo to any system (Docker and PowerShell available).
2. From the **repo root** run: `.\scripts\launch-greenfield.ps1`
3. When prompted, enter your **break-glass admin username** and **password** (used for Keycloak and Bitwarden).
4. The script creates all Docker networks, generates secrets, starts the merged core stack (IAM, messaging, tooling, ChatOps — `docker-compose/stack-manifest.json`), waits for Vault and Keycloak, writes secrets to Vault.
5. **Walk away.** When you return, the stack is up.

No other mandatory steps. No static `.env` file; secrets exist in memory and in Vault.

## Prerequisites

- **Docker** (or Podman) and **Docker Compose**
- **PowerShell** (Windows or PowerShell Core)
- Repo cloned (e.g. `C:\GiTeaRepos\devsecops-pipeline` or your path)

## How to run

From the **repo root**:

```powershell
.\scripts\launch-greenfield.ps1
```

Optional:

- **Save Vault token for later runs:**  
  `.\scripts\launch-greenfield.ps1 -SaveVaultToken`  
  Then next time you can run `.\scripts\start-from-vault.ps1` instead of re-running the full bootstrap.
- **Non-interactive (CI):**  
  Set `BREAK_GLASS_USER` and `BREAK_GLASS_PASSWORD` in the environment, then run:  
  `.\scripts\launch-greenfield.ps1 -NonInteractive`

## What the script does

1. **Networks** — Runs `scripts\create-networks.ps1` so all **16** Docker networks exist: `gitea_net`, `n8n_net`, `zammad_net`, `bitwarden_net`, `gateway_net`, `portainer_net`, `llm_net`, `chatops_net`, `msg_backbone_net`, `iam_net`, `freeipa_net`, `agent_mesh_net`, `discovery_net`, `sdn_lab_net`, `telemetry_net`, **`docs_net`** (Docsify / Traefik `/docs`; required by `docker-compose.tooling.yml`). See [NETWORK_DESIGN.md](NETWORK_DESIGN.md). If you pulled a newer repo and see `docs_net ... could not be found`, re-run `.\scripts\create-networks.ps1` once.
2. **Bootstrap** — Runs `scripts\secrets-bootstrap.ps1`: generates strong random secrets, sets them in the process environment, starts the merged Docker Compose core stack (IAM, messaging, tooling, ChatOps), waits for Vault, writes secrets to Vault at `secret/devsecops`, creates the break-glass user in Keycloak and injects secrets into Bitwarden. By default **no `docker-compose/.env`** is written; use `-WriteEnvFile` if you need one. On a Linux SDN host, add `-IncludeSdnTelemetry` to `launch-greenfield.ps1` to merge in `docker-compose.network.yml` and `docker-compose.telemetry.yml` (see [SDN_TELEMETRY.md](SDN_TELEMETRY.md)).
3. **Optional** — If you used `-SaveVaultToken`, runs `scripts\save-vault-token-to-keystore.ps1` so the Vault token is stored in Windows Credential Manager or PowerShell SecretStore (no file on disk).

## After the first run

- **Keycloak:** http://127.0.0.1:8180/admin — Log in with your break-glass username and the **generated** admin password (stored in Vault at `secret/devsecops`, key `KEYCLOAK_ADMIN_PASSWORD`).
- **Later runs:** If you used `-SaveVaultToken`, run `.\scripts\start-from-vault.ps1` to load secrets from Vault and start the stack. Otherwise set `$env:VAULT_TOKEN` and run `.\scripts\start-from-vault.ps1`, or re-run `.\scripts\launch-greenfield.ps1` (which will generate new secrets and overwrite Vault).

## Alternatives for networks

- **Ansible:** Run `ansible-playbook -i inventory.yml playbooks/site.yml` (or `deploy-devsecops-mesh.yml`); the playbook creates all Docker networks. Then run `launch-greenfield.ps1` with `-StartStack:$true` or use Ansible to start containers with env from Vault.
- **OpenTofu:** From `opentofu/`, run `tofu init && tofu apply` to create the same networks. Then run `launch-greenfield.ps1` to bootstrap and start the stack.

## See also

- [GREENFIELD_REGISTRATION.md](GREENFIELD_REGISTRATION.md) — How initial values and the Vault token are managed without static files.
- [DEPLOYMENT.md](DEPLOYMENT.md) — Full deployment order and troubleshooting.
- [VARLOCK_USAGE.md](VARLOCK_USAGE.md) — Where secrets live (Vault) and how they are injected.
