# Register admin kbolsen and SSH key in Vault

One-time setup: generate the admin user **kbolsen**, an SSH key on your machine, set **MFA to kenneth.olsen.atm@gmail.com**, and store credentials in Vault so the pipeline uses them (no secrets in `.env`).

## What was created

- **Username:** `kbolsen` (used for Keycloak admin and, where applicable, Gitea/other admin).
- **MFA email:** `kenneth.olsen.atm@gmail.com` — stored in Vault as `admin_mfa_email` for OTP, recovery, or identity binding when you enable MFA in Keycloak/Vault.
- **SSH key pair** (on your laptop, under `.dev/`, gitignored):
  - Private: `.dev/kbolsen_admin` — keep this only on your machine; use it for Git SSH and admin access.
  - Public: `.dev/kbolsen_admin.pub` — this is stored in Vault as `admin_ssh_public_key` for services (e.g. Gitea) to authorize you.
- **Keycloak password:** Generated once and saved to `.dev/kbolsen_keycloak_password.txt` (gitignored). Also written to Vault as `KEYCLOAK_ADMIN_PASSWORD`.

## Run the registration script

From `devsecops-pipeline` (or repo root with script path adjusted):

```powershell
# Set Vault access
$env:VAULT_ADDR = "http://vault:8200"   # or your Vault address
$env:VAULT_TOKEN = "<your-token>"
# Optional: $env:VAULT_SECRET_PATH = "secret/devsecops"

.\scripts\register-kbolsen-in-vault.ps1
```

The script will:

1. Use the existing key at `.dev/kbolsen_admin.pub` (or prompt you to generate it).
2. Use or generate `KEYCLOAK_ADMIN_PASSWORD` and save it once to `.dev/kbolsen_keycloak_password.txt`.
3. Write a payload to `.dev/vault-payload-kbolsen.json` (gitignored).
4. If the Vault CLI is available, run `vault kv put` to store:
   - `KEYCLOAK_ADMIN` = `kbolsen`
   - `KEYCLOAK_ADMIN_PASSWORD` = (generated password)
   - `admin_ssh_public_key` = (contents of `.dev/kbolsen_admin.pub`)
   - `admin_mfa_email` = `kenneth.olsen.atm@gmail.com` (override with `$env:ADMIN_MFA_EMAIL` if needed)

If the Vault CLI is not installed or the put fails, the script prints the exact `vault kv put` command and the paths to the payload and password file so you can put the secrets manually.

## After registration

- **Keycloak:** Log in at `http://localhost:8180/admin` with username **kbolsen** and the password from `.dev/kbolsen_keycloak_password.txt` (or from Vault at `VAULT_SECRET_PATH`).
- **Git / SSH:** Use the private key `.dev/kbolsen_admin` for SSH auth (e.g. add to ssh-agent, or configure Git to use it). Services that read `admin_ssh_public_key` from Vault will accept this key.

## Regenerating the key

If you need a new key:

```powershell
Remove-Item .dev\kbolsen_admin, .dev\kbolsen_admin.pub -ErrorAction SilentlyContinue
ssh-keygen -t ed25519 -C "kbolsen" -f .dev\kbolsen_admin -N '""'
# Then run .\scripts\register-kbolsen-in-vault.ps1 again (optionally remove .dev\kbolsen_keycloak_password.txt to generate a new password).
```
