# Varlock Usage for DevSecOps Pipeline

## What Is Varlock

Varlock is the project's **schema-driven environment variable** convention. The schema file (`devsecops.env.schema`) defines variable names, types, sensitivity, and defaults. **No secret values** are stored in the schema or in `.env`; values are injected at runtime from a **Vault** (or other secret store).

## Flow: Generate → Vault → Retrieve → Inject (no static secrets)

There is **no static secrets file** to maintain. The intended flow:

1. **Generate** — `scripts/secrets-bootstrap.ps1` generates strong random values for all pipeline secrets (Keycloak DB/admin, Vault root token, Zammad Postgres, Gitea/n8n/Zammad/Solace/Teleport tokens, webhook HMAC, **Bitwarden admin token**, gateway refresh secret, etc.) and keeps them only in memory.
2. **Vault** — The same script enables KV v2 at `secret/` and writes all generated (or existing env) secrets to **Vault** at `secret/devsecops`. Secrets are not written to disk.
3. **Retrieve** — On later runs you **connect to Vault** and read `secret/devsecops`:
   - **Script:** `scripts/start-from-vault.ps1` uses the Vault token (from env or Windows Credential Manager / SecretStore) to fetch all keys from Vault and export them into the current process environment.
   - **Ansible:** `playbooks/start-containers-with-vault.yml` fetches from Vault and passes the same keys into the play as `devsecops_secrets`.
4. **Inject** — Containers receive secrets only via **injected env** (no `.env` file):
   - When using the script: the shell that ran `start-from-vault.ps1` has env set; `launch-stack.ps1` runs Docker Compose in that shell, so every container (including **Bitwarden**, Keycloak, Gitea, n8n, Zammad, Zulip/ChatOps) gets e.g. `BITWARDEN_ADMIN_TOKEN`, `KEYCLOAK_ADMIN_PASSWORD`, etc. from that env.
   - When using Ansible: the `devsecops_containers` role passes `devsecops_secrets` (from Vault or from ansible-vault-encrypted group_vars) into the Compose `env:` so every container receives the same keys.

**Bitwarden and Vault:** Bitwarden (Vaultwarden) does not talk to Vault directly. It receives **ADMIN_TOKEN** (and any other config) as environment variables. Those variables are populated from Vault by the orchestrator (script or Ansible) before the container starts. So "Bitwarden connect to Vault" means: **retrieve from Vault → inject into Bitwarden container as env**. Same for every other service; Vault is the single source of truth, and the only thing that connects to Vault is your host/orchestrator (script or Ansible).

## Where Credentials Live (Vault)

With Varlock, **all sensitive values are stored in Vault**, not in `.env`. Do not put secrets in `.env`; use Vault as the source of truth.

| What you need | Where it lives |
|---------------|----------------|
| Keycloak bootstrap (first init) | `KEYCLOAK_ADMIN`, `KEYCLOAK_ADMIN_PASSWORD` — use once; then prefer LDAP + automation client (see docs/IAM_LDAP_AND_AUTOMATION.md) |
| Keycloak automation (scripts)  | `KEYCLOAK_AUTOMATION_CLIENT_ID`, `KEYCLOAK_AUTOMATION_CLIENT_SECRET` at `secret/devsecops` — service-account token, no admin password |
| Admin SSH public key (e.g. kbolsen) | Same path; key `admin_ssh_public_key` (used for Gitea/Teleport and other admin access). See `scripts/README-kbolsen-vault.md` to generate and register. |
| Admin MFA email | Same path; key `admin_mfa_email` (e.g. `kenneth.olsen.atm@gmail.com`) for OTP/recovery when MFA is enabled. |
| DB passwords, API tokens, etc. | Same path; keys match `devsecops.env.schema` (e.g. `POSTGRES_PASSWORD`, `KEYCLOAK_DB_PASSWORD`, `SOLACE_PASSWORD`, `GITEA_API_TOKEN`, `ZAMMAD_API_TOKEN`) |

- **Vault address**: `VAULT_ADDR` (default `http://vault:8200`).
- **Secret path**: `VAULT_SECRET_PATH` (default `secret/devsecops`).

Example (Vault KV v2): read the secret at `secret/data/devsecops` and use the keys defined in the schema. Your Keycloak admin login is whatever is stored there for `KEYCLOAK_ADMIN` and `KEYCLOAK_ADMIN_PASSWORD`.

## Inject at Runtime (Ansible → Containers)

**All env (including secrets) is injected by Ansible into containers; no manual `.env` file.**

1. **Ansible (primary)**: Use the `devsecops_containers` role so that env is passed into containers at start time from Ansible vars or from HashiCorp Vault.
   - **From HashiCorp Vault**: `VAULT_ADDR=... VAULT_TOKEN=... ansible-playbook -i inventory.yml playbooks/start-containers-with-vault.yml`. Secrets are read from `secret/data/devsecops` (KV2) and passed to Docker Compose by Ansible.
   - **From Ansible Vault**: Put `devsecops_secrets` in group_vars (encrypted with `ansible-vault encrypt`). Run `ansible-playbook -i inventory.yml playbooks/site.yml`; the role receives secrets from group_vars and injects them into the containers.
   - See `ansible/roles/devsecops_containers/README.md` and `ansible/group_vars/ai_mesh_nodes/devsecops_secrets.yml.example`.
2. **n8n**: Store secrets in n8n **Credentials**, populated from Vault or by reference. Reference them in workflows as `{{ $credentials.<name> }}`. Do not paste raw secrets into workflow JSON.
3. **Agents and LLMs**: Only expose environment variables that are populated from this schema. Never put raw secrets in AI context or in MCP payloads.

## Secrets creator (greenfield, no static storage)

Use **`scripts/secrets-bootstrap.ps1`** to generate strong random secrets, inject them into Vault, and start the stack **without ever writing secrets to disk**:

```powershell
cd scripts
.\secrets-bootstrap.ps1
```

- Generates cryptographically random values for all pipeline secrets (Keycloak DB/admin, Vault root token, Zammad Postgres, Gitea/n8n/Zammad/Solace/Teleport API tokens, webhook HMAC, etc.).
- Sets them in the **current process environment** and starts the merged Docker Compose core stack (IAM, messaging, tooling, ChatOps per `docker-compose/stack-manifest.json`) so containers receive them via env.
- Writes the same secrets to **Vault** at `secret/devsecops` for Varlock and other consumers.
- **No `.env` or other file** is created; secrets exist only in memory and in Vault.

**Getting initial values without static storage:** All values are **generated** at first run. The only credential you might keep for the next run is the **Vault token**. To start the stack on later runs without re-running bootstrap: (1) After first bootstrap, run `.\save-vault-token-to-keystore.ps1` to store the token in Windows Credential Manager or PowerShell SecretStore (no file). (2) On next boot, run `.\start-from-vault.ps1`; it reads the token from env or keystore, fetches all secrets from Vault, exports to env, and runs the stack. See [GREENFIELD_REGISTRATION.md](GREENFIELD_REGISTRATION.md).

Options: `-StartStack:$false` to only push (generated or existing env) secrets to Vault; `-KeycloakAdminUsername admin` (default); `-VaultAddr http://127.0.0.1:8200`; `-IdentityBackend Keycloak|FreeIPA`; `-SkipBitwardenInject` to skip Bitwarden; `-OnlyBreakGlass` with `-StartStack:$false` to run only break-glass steps (stack and Vaultwarden must already be up).

### Break-glass admin (human-in-the-loop)

At script start, `secrets-bootstrap.ps1` prompts for a **break-glass admin username** and **password** (SecureString). These are used to:

1. **Create a master user** in Keycloak (master realm) or FreeIPA (`-IdentityBackend Keycloak` or `FreeIPA`). The script calls the Keycloak Admin API or `docker exec` against the FreeIPA container so a human administrator can log in with that identity.
2. **Inject all generated pipeline secrets into Bitwarden (Vaultwarden)** so the same human can access them. Because Vaultwarden uses client-side encryption, the script uses the **Bitwarden CLI (`bw`)** only: it points `bw` at your local Vaultwarden instance (`BW_SERVER` / `BITWARDEN_DOMAIN`), authenticates with the break-glass username and password, unlocks the vault, and creates each secret as a Login item (no secrets written to disk).

**Prerequisites:**

- **Bitwarden CLI:** Install `bw` (Bitwarden CLI) on the host and ensure it is on `PATH` if you want secret injection into Bitwarden. Use `-SkipBitwardenInject` if `bw` is not installed.
- **Break-glass Bitwarden account:** The Bitwarden account (used by `bw login`) must already exist on Vaultwarden. If it does not, the script will create the user in Keycloak/FreeIPA but login to Bitwarden will fail with a clear message: sign up once at `BITWARDEN_DOMAIN` (e.g. `http://localhost:8484`) with the same username and password, then re-run the script (e.g. `.\secrets-bootstrap.ps1 -StartStack:$false -OnlyBreakGlass`) to inject secrets.

You can pass credentials as parameters to avoid prompts: `-BreakGlassUsername admin` and `-BreakGlassPassword` (SecureString). No secrets are written to disk; credentials are used only in memory for LDAP/Keycloak user creation and for `bw` login/unlock and item creation.

## Run stacks (Ansible injects env; no .env)

**Preferred:** Use Ansible so env is injected into containers by the playbook (no `.env` on disk).

- **From HashiCorp Vault:**  
  `VAULT_ADDR=http://vault:8200 VAULT_TOKEN=... ansible-playbook -i inventory.yml playbooks/start-containers-with-vault.yml`

- **From Ansible Vault (encrypted group_vars):**  
  Create `group_vars/ai_mesh_nodes/devsecops_secrets.yml` from `devsecops_secrets.yml.example`, fill values, run `ansible-vault encrypt group_vars/ai_mesh_nodes/devsecops_secrets.yml`, then run `ansible-playbook -i inventory.yml playbooks/site.yml`.

**Alternative (script):** If you cannot use Ansible, use `scripts/start-from-vault.ps1` to export from Vault into the environment and then run `launch-stack.ps1` from the same shell. Do not write secrets to `.env`.

## Validate Schema

If you have a `varlock` CLI:

```bash
varlock validate --schema devsecops.env.schema
```

Otherwise, treat the schema as documentation and ensure Vault (or your secret store) provides every `@required: true` variable; no secrets in `.env`.

## Schema Location

- **Unified schema**: `devsecops.env.schema` (repo root)
- **Sensitive fields**: Marked with `@sensitive: true`; values only in Vault (or secret manager), never in repo or `.env`.
