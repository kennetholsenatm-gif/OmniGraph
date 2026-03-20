# IAM: LDAP for identity, tokens/certificates for automation

No static admin passwords in env or files. **Human admins** log in with LDAP (FreeIPA). **Automation** (scripts, Ansible, CI) uses a Keycloak **service-account client** (client credentials) or client certificates; store only the client secret (or cert) in Vault.

## Model

| Who / what        | Identity              | How to authenticate                          |
|-------------------|-----------------------|-----------------------------------------------|
| Human admin       | LDAP user (FreeIPA)   | Log in to Keycloak Admin UI with LDAP creds   |
| Scripts / Ansible | Service-account client| Client credentials token (or mTLS client cert)|

- **No** `KEYCLOAK_ADMIN` / `KEYCLOAK_ADMIN_PASSWORD` in normal operation. Those are **bootstrap-only** (see below).
- **No** password files (e.g. `.dev/kbolsen_keycloak_password.txt`) for automation.

## 1. Bootstrap (one-time)

Keycloak creates the initial master-realm admin only on **first DB init**. You have to provide one set of credentials once:

- Either: start Keycloak with `KEYCLOAK_ADMIN` and `KEYCLOAK_ADMIN_PASSWORD` (env or Vault), use that **only** to configure LDAP and the automation client, then stop using that account (or remove it).
- Or: use a short-lived bootstrap password from Vault for first start, then rotate it out after LDAP + automation client are in place.

After bootstrap, **do not** rely on that initial admin for day-to-day use.

## 2. LDAP (FreeIPA) as admin identity

1. In FreeIPA, create an LDAP user that will act as Keycloak admin (e.g. `keycloak-admin` or a person’s account).
2. In Keycloak: **User federation → Add provider → ldap**. Configure connection to FreeIPA (bind DN, URL, sync settings). Map LDAP groups/attributes as needed.
3. In Keycloak: assign that LDAP user (or an LDAP group) the **realm-admin** (or equivalent) role so they can use the Admin UI.
4. From then on, admins use **that LDAP account** to log in to Keycloak; no static admin password in env.

## 3. Automation: service-account client (tokens)

1. In Keycloak (master realm): **Clients → Create**:
   - Client ID: `devsecops-automation` (or your choice).
   - Client authentication: **On** (confidential).
   - Save, then open the client.
2. **Capability config:** Enable **Service accounts enabled**.
3. **Credentials tab:** Copy the client secret; store it in Vault at `secret/devsecops` as `KEYCLOAK_AUTOMATION_CLIENT_SECRET`. Store the client ID as `KEYCLOAK_AUTOMATION_CLIENT_ID` (e.g. `devsecops-automation`).
4. **Service account roles:** Open **Service account roles** (or the service user linked to this client). Under **realm-management**, assign at least **realm-admin** (or the roles your scripts need). The access token will carry these roles and the Admin API will accept it.
5. Scripts and Ansible get a token with **client credentials**:
   - `POST {KEYCLOAK_URL}/realms/master/protocol/openid-connect/token`
   - Body: `grant_type=client_credentials&client_id=...&client_secret=...`
   - Use the returned `access_token` as `Authorization: Bearer {token}` for the Admin API.

All automation (e.g. `configure-keycloak-oidc-clients.ps1`, `sync-identities-to-keycloak.ps1`) should use **KEYCLOAK_AUTOMATION_CLIENT_ID** and **KEYCLOAK_AUTOMATION_CLIENT_SECRET** (from env or Vault), not admin username/password.

## 4. Optional: client certificates (mTLS)

For stronger automation auth, use a client that requires a **client certificate** instead of a secret:

- In Keycloak: create a client (or reuse `devsecops-automation`) and enable **X.509 client certificate authentication** (or configure the client to require a certificate).
- Store the client cert (and key) in Vault or a secure store; scripts use the cert when calling the token endpoint and Admin API.
- Then you do not need to store `KEYCLOAK_AUTOMATION_CLIENT_SECRET`; the cert is the credential.

Details depend on your Keycloak and reverse-proxy setup (e.g. Traefik client cert validation and Keycloak’s expectations).

## 5. Vault keys (summary)

| Purpose              | Vault path / keys |
|----------------------|--------------------|
| Automation (scripts) | `secret/devsecops`: `KEYCLOAK_AUTOMATION_CLIENT_ID`, `KEYCLOAK_AUTOMATION_CLIENT_SECRET` |
| Bootstrap (first run)| `KEYCLOAK_ADMIN`, `KEYCLOAK_ADMIN_PASSWORD` — use only once, then rely on LDAP + automation client |
| OIDC app clients     | `GITEA_OIDC_CLIENT_SECRET`, `N8N_OIDC_CLIENT_SECRET`, etc. |

## 6. IAM as code (Ansible / OpenTofu)

This work belongs in **Ansible** or **OpenTofu**, not one-off scripts. See **docs/IAM_IAC.md**.

- **Ansible:** Role `ansible/roles/keycloak_iam` uses `community.general.keycloak_client` to create the automation client and OIDC clients (gitea, n8n, zammad). Run `playbooks/keycloak-iam.yml` after the IAM stack is up (bootstrap admin from Vault or group_vars).
- **OpenTofu:** Add a Keycloak provider and define realm, clients, and LDAP IdP as resources in a separate stack if you prefer declarative IaC.

## 7. Optional FreeRADIUS (FOSS AAA)

For network/device AAA (802.1X, VPN, infrastructure auth) add native FreeRADIUS on Alma using:

- Role: `ansible/roles/freeradius_alma`
- Playbook: `ansible/playbooks/deploy-freeradius-native.yml`

Recommended pattern:

- Keep **Keycloak** for OIDC/OAuth app auth.
- Use **FreeRADIUS** for RADIUS-speaking infrastructure.
- Back FreeRADIUS against **FreeIPA/LDAP** (same identity source) to avoid account drift.

Store `freeradius_clients[*].secret` and LDAP bind secrets in Vault/Ansible Vault, not plaintext vars.

## 8. References

- Keycloak: [Service accounts](https://www.keycloak.org/docs/latest/server_admin/#_service_accounts), [Admin REST API](https://www.keycloak.org/docs-api/latest/rest-api/index.html#_clients_resource).
- FreeIPA / LDAP: `docker-compose.identity.yml`, `docs/IDENTITIES_AND_PRIVILEGES.md`.
