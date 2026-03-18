# Varlock Usage for DevSecOps Pipeline

## What Is Varlock

Varlock is the project's **schema-driven environment variable** convention. The schema file (`devsecops.env.schema`) defines variable names, types, sensitivity, and defaults. **No secret values** are stored in the schema or in `.env`; values are injected at runtime from a **Vault** (or other secret store).

## Where Credentials Live (Vault)

With Varlock, **all sensitive values are stored in Vault**, not in `.env`. Do not put secrets in `.env`; use Vault as the source of truth.

| What you need | Where it lives |
|---------------|----------------|
| Keycloak admin user / password | Vault path `VAULT_SECRET_PATH` (default `secret/devsecops`), keys `KEYCLOAK_ADMIN`, `KEYCLOAK_ADMIN_PASSWORD` |
| Admin SSH public key (e.g. kbolsen) | Same path; key `admin_ssh_public_key` (used for Gitea/Teleport and other admin access). See `scripts/README-kbolsen-vault.md` to generate and register. |
| Admin MFA email | Same path; key `admin_mfa_email` (e.g. `kenneth.olsen.atm@gmail.com`) for OTP/recovery when MFA is enabled. |
| DB passwords, API tokens, etc. | Same path; keys match `devsecops.env.schema` (e.g. `POSTGRES_PASSWORD`, `KEYCLOAK_DB_PASSWORD`, `SOLACE_PASSWORD`, `GITEA_API_TOKEN`, `ZAMMAD_API_TOKEN`) |

- **Vault address**: `VAULT_ADDR` (default `http://vault:8200`).
- **Secret path**: `VAULT_SECRET_PATH` (default `secret/devsecops`).

Example (Vault KV v2): read the secret at `secret/data/devsecops` and use the keys defined in the schema. Your Keycloak admin login is whatever is stored there for `KEYCLOAK_ADMIN` and `KEYCLOAK_ADMIN_PASSWORD`.

## Inject at Runtime (Vault → Compose)

1. **Docker Compose**: Do **not** put secrets in `.env`. Export variables from Vault into the environment (or a temporary env file that is gitignored and never committed), then run compose so containers receive them.
   - Example (PowerShell, Vault CLI): fetch from Vault, set env vars, then run `.\launch-stack.ps1` (or the individual `docker compose ... up -d` commands). See **Run Compose with Vault** below.
   - Compose files expect variable names from `devsecops.env.schema` (e.g. `POSTGRES_PASSWORD`, `KEYCLOAK_ADMIN`, `KEYCLOAK_ADMIN_PASSWORD`, `SOLACE_PASSWORD`, `GITEA_API_TOKEN`).
2. **n8n**: Store secrets in n8n **Credentials** (e.g. HTTP Header Auth, MQTT password), populated from Vault or by reference. Reference them in workflows as `{{ $credentials.<name> }}`. Do not paste raw secrets into workflow JSON.
3. **Agents and LLMs**: Only expose environment variables that are populated from this schema (e.g. via Vault agent or entrypoint script). Never put raw secrets in AI context or in MCP payloads.

## Secrets creator (greenfield, no static storage)

Use **`scripts/secrets-bootstrap.ps1`** to generate strong random secrets, inject them into Vault, and start the stack **without ever writing secrets to disk**:

```powershell
cd devsecops-pipeline\scripts
.\secrets-bootstrap.ps1
```

- Generates cryptographically random values for all pipeline secrets (Keycloak DB/admin, Vault root token, Zammad Postgres, Gitea/n8n/Zammad/Solace/Teleport API tokens, webhook HMAC, etc.).
- Sets them in the **current process environment** and starts the Docker Compose stacks (IAM → messaging → tooling) so containers receive them via env.
- Writes the same secrets to **Vault** at `secret/devsecops` for Varlock and other consumers.
- **No `.env` or other file** is created; secrets exist only in memory and in Vault.

**Getting initial values without static storage:** All values are **generated** at first run. The only credential you might keep for the next run is the **Vault token**. To start the stack on later runs without re-running bootstrap: (1) After first bootstrap, run `.\save-vault-token-to-keystore.ps1` to store the token in Windows Credential Manager or PowerShell SecretStore (no file). (2) On next boot, run `.\start-from-vault.ps1`; it reads the token from env or keystore, fetches all secrets from Vault, exports to env, and runs the stack. See [GREENFIELD_REGISTRATION.md](GREENFIELD_REGISTRATION.md).

Options: `-StartStack:$false` to only push (generated or existing env) secrets to Vault; `-KeycloakAdminUsername admin` (default); `-VaultAddr http://127.0.0.1:8200`.

## Run Compose with Vault (existing secrets)

If you already have secrets in Vault, ensure `VAULT_ADDR` and `VAULT_TOKEN` are set, export the pipeline variables from Vault into the environment, then start the stack so no secrets live in `.env`:

**PowerShell (example):**

```powershell
$env:VAULT_ADDR = "http://127.0.0.1:8200"
$env:VAULT_TOKEN = "<your-token>"
# Export keys from Vault (e.g. vault kv get -format=json secret/devsecops) into $env:KEYCLOAK_DB_PASSWORD, etc.

cd devsecops-pipeline\docker-compose
.\launch-stack.ps1
```

If your launcher or CI injects env from Vault (e.g. Vault Agent, or a script that runs `vault kv get` and exports), run that first so `KEYCLOAK_ADMIN`, `KEYCLOAK_ADMIN_PASSWORD`, and other required vars are set before `launch-stack.ps1` or `docker compose`.

## Validate Schema

If you have a `varlock` CLI:

```bash
varlock validate --schema devsecops-pipeline/devsecops.env.schema
```

Otherwise, treat the schema as documentation and ensure Vault (or your secret store) provides every `@required: true` variable; no secrets in `.env`.

## Schema Location

- **Unified schema**: `devsecops-pipeline/devsecops.env.schema`
- **Sensitive fields**: Marked with `@sensitive: true`; values only in Vault (or secret manager), never in repo or `.env`.
