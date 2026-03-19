# keycloak_iam

Ansible role to manage Keycloak IAM as code: **automation client** (service account) and **OIDC clients** (gitea, n8n, zammad). Uses `community.general.keycloak_client`.

## Requirements

- Keycloak running and reachable from the target host (e.g. IAM stack up, proxy at 8180 or gateway at `/keycloak`).
- `community.general` collection: `ansible-galaxy collection install community.general`.
- Auth: bootstrap admin credentials (for first run) from `devsecops_secrets`: `KEYCLOAK_ADMIN`, `KEYCLOAK_ADMIN_PASSWORD`.

## Role variables

Passed by playbook or group_vars; see `defaults/main.yml` for defaults.

| Variable | Description |
|----------|-------------|
| `keycloak_iam_url` | Keycloak base URL (e.g. `http://127.0.0.1:8180` or `http://localhost/keycloak`). |
| `keycloak_iam_realm` | Realm (default `master`). |
| `keycloak_iam_username` | Admin username (bootstrap). |
| `keycloak_iam_password` | Admin password (bootstrap). |
| `keycloak_iam_gateway_base_url` | Base URL for redirect URIs (e.g. `http://localhost`). |
| `keycloak_iam_automation_client_id` | Client ID for service account (default `devsecops-automation`). |

## Usage

Run the dedicated playbook after the IAM stack is up:

```bash
cd ansible
ansible-playbook -i inventory.yml playbooks/keycloak-iam.yml -e @group_vars/ai_mesh_nodes/devsecops_secrets.yml
```

Or include the role in a play that has `devsecops_secrets` and set `keycloak_iam_username` / `keycloak_iam_password` from it.

## After first run

1. In Keycloak Admin: **Clients → devsecops-automation → Service account roles** → assign **realm-management** → **realm-admin** (or needed roles).
2. **Clients → devsecops-automation → Credentials** → copy the secret to Vault as `KEYCLOAK_AUTOMATION_CLIENT_SECRET`.
3. From then on, scripts and automation use client credentials; no bootstrap password needed for day-to-day.

See **docs/IAM_IAC.md** and **docs/IAM_LDAP_AND_AUTOMATION.md**.
