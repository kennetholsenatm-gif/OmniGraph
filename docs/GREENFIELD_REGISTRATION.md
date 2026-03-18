# Greenfield Product Registration (No Static Secrets)

## The problem

You want **all pipeline secrets in Vault** and **no static files** (no `.env` with passwords). But something has to bootstrap Vault and provide the first way to read from it. So:

- **Initial values** must exist somehow: either **generated at runtime** or **entered once** by a human/CI.
- After a **reboot**, Vault has the secrets but you need **one credential** (the Vault token) to read them—and that credential must come from somewhere that isn’t a committed file.

## Approach: Generate at runtime, single optional seed

### 1. First run: no static storage

**All initial values are generated**, not read from a file:

1. Run **`scripts/secrets-bootstrap.ps1`**.
2. It **generates** strong random values for every pipeline secret (Keycloak DB, admin password, Vault root token, Zammad, Gitea, Solace, etc.) **in memory**.
3. It **starts** the stack (Vault, Keycloak, messaging, tooling) with those values in the **process environment**.
4. It **writes** the same values to **Vault** at `secret/devsecops` via the API.
5. **Nothing is written to disk.** Secrets exist only in memory and inside Vault.

So the “initial values” are **created at runtime**; there are no static files to protect.

### 2. After reboot: the one credential

After a restart, Vault (and optionally other services) may still be running and already hold the secrets. To **start the stack again** or **run compose** without re-running bootstrap, you need to **read from Vault**. That requires **one** credential: **VAULT_TOKEN** (the root token used in dev, or an app token in production).

You have three ways to provide it **without** a static file:

| Option | How | Use case |
|--------|-----|----------|
| **A. Re-run bootstrap** | Run `secrets-bootstrap.ps1` again. It generates **new** secrets and a **new** Vault root token, starts Vault with that token, and overwrites `secret/devsecops`. | Dev / ephemeral; you accept new secrets (and new token) each time, or you run bootstrap only once and keep the process/session that has env vars. |
| **B. OS credential store** | Save the Vault token **once** in the OS (e.g. Windows Credential Manager). A script reads it from there and uses it to fetch all other secrets from Vault into the environment, then starts the stack. **No file.** | Dev or single-host; you want to restart without re-running bootstrap and without typing the token. |
| **C. Human / CI** | Human stores the token in a password manager; when starting the stack they run `$env:VAULT_TOKEN = "<paste>"` then a script that reads from Vault and exports env, then launch-stack. Or CI stores the token as a secret (e.g. GitHub Actions secret) and the pipeline exports it and reads from Vault. | Any environment; the “initial” value is the token, stored in one place that is not a repo file. |

So the **only** value that might be “stored” somewhere (outside Vault) is the **Vault token**. Everything else stays in Vault and is pulled at runtime.

### 3. Recommended flow (greenfield)

1. **First time (or after “reset”)**  
   - Run: `.\scripts\secrets-bootstrap.ps1`  
   - Optional: run `.\scripts\save-vault-token-to-keystore.ps1` to store the current Vault token in Windows Credential Manager (so you don’t have to type it again).

2. **Later runs (after reboot)**  
   - Run: `.\scripts\start-from-vault.ps1`  
   - That script gets `VAULT_TOKEN` from the environment or from the OS keystore, reads `secret/devsecops` from Vault, exports all keys to the environment, and runs the Docker Compose stack. **No static file**; the only “initial” value is the token (env or keystore).

3. **If you never save the token**  
   - Either re-run `secrets-bootstrap.ps1` each time (option A), or set `$env:VAULT_TOKEN` manually each session (option C).

## Summary

- **Initial values** = **generated** by `secrets-bootstrap.ps1` (no static storage).
- **Vault** = store for all pipeline secrets after first run.
- **Single “registration” secret** = the **Vault token**; optional persistence in OS keystore or CI so you can start from Vault on subsequent runs without a static file.

See **`scripts/start-from-vault.ps1`** and **`scripts/save-vault-token-to-keystore.ps1`** for the concrete steps.

**Linux/macOS:** Use `VAULT_TOKEN` in the environment (e.g. in your shell profile or a small wrapper that sources it from a single file outside the repo), or configure PowerShell SecretStore and use `Set-Secret` / `Get-Secret` for the token so `start-from-vault.ps1` can read it.
