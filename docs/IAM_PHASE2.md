# Phase 2: IAM Mesh — Keycloak OIDC for Gitea, n8n, Zammad

Everything that supports OIDC/SAML authenticates via Keycloak. This document defines the **realm strategy**, **Keycloak OIDC clients**, **Docker environment variables**, and **per-application configuration** so Gitea, n8n, and Zammad use Keycloak as the IdP.

## Prerequisites

- Phase 1 complete: Traefik routes `/gitea`, `/n8n`, `/zammad`, `/keycloak`.
- Keycloak reachable at **http://localhost/keycloak** (or `https://${GATEWAY_DOMAIN}/keycloak` with TLS).
- Realm: **master** (default) or a dedicated realm such as **devsecops**. Use `KEYCLOAK_REALM` consistently.
- **Automation:** Scripts (e.g. `configure-keycloak-oidc-clients.ps1`) use a **service-account client** (tokens), not admin password. See [IAM_LDAP_AND_AUTOMATION.md](IAM_LDAP_AND_AUTOMATION.md) for LDAP admin identity and `KEYCLOAK_AUTOMATION_CLIENT_ID` / `KEYCLOAK_AUTOMATION_CLIENT_SECRET`.

## 1. Keycloak realm and URL base

- **Realm:** Use `master` or create realm `devsecops`. All three applications use the same realm.
- **Keycloak public URL (issuer):** Must be the URL users see in the browser. Behind Traefik at path `/keycloak`:
  - **HTTP:** `http://localhost/keycloak` (or `http://${GATEWAY_DOMAIN}/keycloak`)
  - **HTTPS:** `https://localhost/keycloak` (or `https://${GATEWAY_DOMAIN}/keycloak`)
- Keycloak is already configured with `KC_HTTP_RELATIVE_PATH=/keycloak` and `KC_HOSTNAME=localhost` (or `GATEWAY_DOMAIN`). Ensure realm **SSL required** is set to **none** for local HTTP (see DEPLOYMENT.md).

## 2. Keycloak OIDC clients (strategy)

Create one OIDC client per application in the chosen realm. Redirect URIs must match the gateway base URL and path.

| Application | Client ID   | Client authentication | Standard flow | Redirect URIs (HTTP example) |
|-------------|-------------|------------------------|---------------|------------------------------|
| Gitea       | `gitea`     | On (confidential)      | Yes           | `http://localhost/gitea/user/oauth2/Keycloak/callback` |
| n8n         | `n8n`       | On (confidential)      | Yes           | `http://localhost/n8n/rest/sso/oidc/callback` |
| Zammad      | `zammad`    | Off (public) or On     | Yes           | `http://localhost/zammad/auth/openid_connect/callback` |

- **Valid post logout redirect URIs (Zammad):** `http://localhost/zammad/*`
- **Web origins:** For Keycloak 23+, set **Web origins** to `+` or the exact origin (e.g. `http://localhost`) so CORS works.
- **Scopes:** Default scopes `openid`, `email`, `profile`. Add `offline_access` if refresh tokens are needed (e.g. Gitea).
- **Gitea:** In Keycloak, set client scopes so `offline_access` is in the **default** scopes for the Gitea client.

Run **scripts/configure-keycloak-oidc-clients.ps1** (or the equivalent API/UI steps) to create/update these clients. Store each client secret in Vault at `secret/devsecops` as `GITEA_OIDC_CLIENT_SECRET`, `N8N_OIDC_CLIENT_SECRET`, `ZAMMAD_OIDC_CLIENT_SECRET`, and inject into the tooling stack at runtime.

## 3. Docker environment variables

### 3.1 Gateway / Keycloak (single-pane or IAM)

Used so Keycloak and the gateway agree on the public base URL:

| Variable | Purpose | Example |
|----------|----------|---------|
| `GATEWAY_DOMAIN` | Hostname users use for the gateway | `localhost` |
| `KEYCLOAK_URL` | Keycloak internal URL (Traefik → keycloak-proxy) | `http://keycloak:8080` |
| `KEYCLOAK_REALM` | Realm used for OIDC | `master` |

Keycloak container (IAM stack) already uses `KC_HTTP_RELATIVE_PATH`, `KC_HOSTNAME`, `KC_PROXY_HEADERS`. For a custom domain, set `KC_HOSTNAME` (or `KC_HOSTNAME_URL`) to the public host (e.g. `GATEWAY_DOMAIN`).

### 3.2 Gitea (tooling)

| Variable | Purpose | Example |
|----------|----------|---------|
| `GITEA__server__ROOT_URL` | Public URL of Gitea (Phase 1) | `http://localhost/gitea/` |
| `KEYCLOAK_URL` | Keycloak base URL reachable from browser (for OAuth2 redirect) | `http://localhost/keycloak` |
| `KEYCLOAK_REALM` | Realm for OIDC | `master` |
| `GITEA_OIDC_CLIENT_ID` | Keycloak client ID for Gitea | `gitea` |
| `GITEA_OIDC_CLIENT_SECRET` | Keycloak client secret (from Vault) | *(secret)* |

Gitea configures the OAuth2 authentication source via **Admin → Authentication → Add OAuth2 Source** (type OpenID Connect). Discovery URL: `http://localhost/keycloak/realms/master/.well-known/openid-configuration`. Client ID and Client Key (secret) from the table above. Callback path is built from the source name (e.g. `Keycloak`) as `/user/oauth2/Keycloak/callback`; ensure the Keycloak client redirect URI matches.

### 3.3 n8n (tooling)

| Variable | Purpose | Example |
|----------|----------|---------|
| `N8N_PATH` | Base path behind Traefik (Phase 1) | `/n8n` |
| `N8N_EDITOR_BASE_URL` | Public URL of n8n (for OIDC redirect) | `http://localhost/n8n` |
| `KEYCLOAK_URL` | Keycloak base URL | `http://localhost/keycloak` |
| `KEYCLOAK_REALM` | Realm for OIDC | `master` |
| `N8N_OIDC_CLIENT_ID` | Keycloak client ID for n8n | `n8n` |
| `N8N_OIDC_CLIENT_SECRET` | Keycloak client secret (from Vault) | *(secret)* |

n8n OIDC is configured in **Settings → SSO → OIDC**: Discovery Endpoint `http://localhost/keycloak/realms/master/.well-known/openid-configuration`, Client ID, Client Secret. Redirect URL shown in n8n must match the Keycloak client (e.g. `http://localhost/n8n/rest/sso/oidc/callback`). OIDC via env-only is limited on community edition; use UI or inject credentials from Vault.

### 3.4 Zammad (tooling)

| Variable | Purpose | Example |
|----------|----------|---------|
| `ZAMMAD_BASE_URL` or app config | Public URL of Zammad (for OIDC redirect) | `http://localhost/zammad` |

Zammad has no env vars for OIDC; configure in **Admin → Settings → Security → Third Party Applications → Authentication via OpenID Connect**: Identifier = Keycloak client ID (`zammad`), Issuer = `http://localhost/keycloak/realms/master`, and configure the callback URL so it matches Keycloak (`http://localhost/zammad/auth/openid_connect/callback`). Store the client secret in Zammad’s UI (or use a public client with PKCE).

## 4. Application configuration summary

1. **Keycloak:** Create realm (if not using `master`), create clients `gitea`, `n8n`, `zammad` with redirect URIs above; set client secrets and default scopes. Use **scripts/configure-keycloak-oidc-clients.ps1** for automation.
2. **Gitea:** Admin → Authentication → Add OAuth2 Source (OpenID Connect), Discovery URL = `{KEYCLOAK_URL}/realms/{KEYCLOAK_REALM}/.well-known/openid-configuration`, Client ID and Secret from Keycloak. Name the source (e.g. Keycloak) and ensure Keycloak redirect URI matches `/gitea/user/oauth2/<name>/callback`.
3. **n8n:** Settings → SSO → OIDC: Discovery Endpoint, Client ID, Client Secret; ensure N8N_EDITOR_BASE_URL is set so the redirect URL matches Keycloak.
4. **Zammad:** Admin → Settings → Security → OpenID Connect: Issuer, Identifier (client ID), callback and logout URLs; optionally store client secret if using a confidential client.

## 5. FreeIPA federation (optional)

Keycloak can federate to FreeIPA (LDAP) so that LDAP/Kerberos users appear in Keycloak and can use OIDC to log into Gitea, n8n, and Zammad. Configure Keycloak **User federation → Add provider → ldap** and map LDAP attributes to the realm. Phase 2 does not require FreeIPA; add when LDAP is required.
