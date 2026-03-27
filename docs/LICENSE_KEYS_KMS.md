# KMS for License and Activation Keys

## Purpose

Product **activation keys** and **license keys** (e.g. n8n, commercial tools) must not live in repo or in static `.env` files. They are managed like other secrets: stored in **Vault** and injected at runtime so the pipeline acts as a **key management** layer (KMS-style).

## How it works

- **Vault** is the store for all secrets, including license/activation keys.
- Keys are written to Vault at **`secret/devsecops`** (or a dedicated path like **`secret/licenses`**) and read at startup by `start-from-vault.ps1` or `secrets-bootstrap.ps1` (which exports them to the environment).
- **Docker Compose** passes the keys into containers via environment variables (e.g. `N8N_LICENSE_ACTIVATION_KEY`). No key is ever committed or kept in a file.

## Keys to manage

| Key | Vault key / env var | Consumed by |
|-----|----------------------|-------------|
| n8n license activation | `N8N_LICENSE_ACTIVATION_KEY` | n8n container (env) |
| (Future) Other product keys | e.g. `GITEA_LICENSE_KEY` | Respective service |

Schema: `devsecops.env.schema` defines these under the relevant `@env-spec` (e.g. N8N). All are `@sensitive: true` and optional unless required by the product.

## Storing keys without static files

You **never** put the key in a file in the repo. Options:

1. **Vault UI**  
   Open `http://127.0.0.1:8200` → Secrets → `secret` → `devsecops` (or `licenses`). Add or edit the key `N8N_LICENSE_ACTIVATION_KEY` with your activation key.

2. **Script (recommended)**  
   Run **`scripts/store-license-keys-in-vault.ps1`**. It prompts for the n8n activation key (or reads from env if already set) and writes it to Vault at `secret/devsecops`. Nothing is written to disk.

3. **One-time env + bootstrap**  
   Set `$env:N8N_LICENSE_ACTIVATION_KEY = "<your-key>"` in the same session, then run `.\secrets-bootstrap.ps1 -StartStack:$false`. The script will push current env (including the key) to Vault. Clear the env after if desired.

4. **Vault CLI**  
   `vault kv patch secret/devsecops N8N_LICENSE_ACTIVATION_KEY="<key>"` (with `VAULT_ADDR` and `VAULT_TOKEN` set).

After the key is in Vault, **start-from-vault.ps1** (or any flow that exports `secret/devsecops` to env) will provide it to the stack; the n8n container receives `N8N_LICENSE_ACTIVATION_KEY` and n8n uses it for activation.

## Optional: dedicated licenses path

If you prefer to separate licenses from other pipeline secrets, use a second path (e.g. `secret/licenses`) and store keys there. Then either:

- Merge that path into the env export in `start-from-vault.ps1` (read both `secret/devsecops` and `secret/licenses` and set env from both), or  
- Have a small wrapper that reads `secret/licenses` and exports only license-related vars before starting compose.

Schema and scripts can be extended to support `VAULT_LICENSES_PATH` (default `secret/licenses`) and document the same “no static file” flow.

## Summary

- **KMS** = Vault; license/activation keys are secrets stored in Vault.
- **No static storage**: use Vault UI, `store-license-keys-in-vault.ps1`, or one-time env + bootstrap.
- **Runtime**: keys are injected into containers via env from Vault at startup.
