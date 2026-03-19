# Wazuh SIEM (optional stack)

## Layout

- Compose: [docker-compose/docker-compose.siem.yml](../docker-compose/docker-compose.siem.yml)
- Network: **`siem_net`** (`100.64.54.0/24`); Traefik joins this net and routes **`/wazuh`** to **`https://wazuh.dashboard:5601`** with **`insecureSkipVerify`** on the backend transport (dashboard uses TLS with generated certs).
- Config + certs: **`docker-compose/siem/wazuh-config/`** — populate with [scripts/bootstrap-wazuh-siem-config.ps1](../scripts/bootstrap-wazuh-siem-config.ps1).

## Host prerequisites (Linux)

```bash
sudo sysctl -w vm.max_map_count=262144
```

Persist in `/etc/sysctl.d/`. Wazuh on Docker Desktop / Windows is unsupported for production parity.

## Passwords and `internal_users.yml`

Upstream demo `internal_users.yml` uses fixed bcrypt hashes (e.g. admin **`SecretPassword`**, kibanaserver **`kibanaserver`**). Compose defaults match those:

| Env | Default (lab) |
|-----|----------------|
| `WAZUH_INDEXER_PASSWORD` | `SecretPassword` |
| `WAZUH_DASHBOARD_KIBANA_PASSWORD` | `kibanaserver` |
| `WAZUH_MANAGER_API_PASSWORD` | `MyS3cr37P450r.*-` |

To use **strong secrets from Vault**, set these in the process environment **and** regenerate bcrypt hashes in **`wazuh-config/wazuh_indexer/internal_users.yml`** (see Wazuh `hash.sh` / password tools for your version). Mismatch prevents indexer/dashboard login.

## Agents (containers / hosts)

- Point agents at **`wazuh.manager`** on **`siem_net`** (e.g. `100.64.54.x`) for ports **1514**, **1515**, **55000** — not through Traefik HTTP.
- From other Docker services, attach the workload to **`siem_net`** or publish manager ports on the host if required.

## Keycloak OIDC (dashboard)

Wazuh Dashboard is OpenSearch Dashboards with the Wazuh plugin. OIDC is configured via **`opensearch_security`** in **`opensearch_dashboards.yml`** and related indexer security config — not a single env var.

1. Create a **confidential** Keycloak client with redirect URI like `https://<gateway>/wazuh/opendistro/security/oauth2/callback` (exact path depends on OpenSearch Security plugin version; verify in Wazuh 4.9 docs).
2. Mount or merge YAML that sets `opensearch_security.auth.type: openid` (and issuer, client ID, client secret, scopes).
3. Store the client secret in Vault; inject via env only if your entrypoint supports templating — often a **mounted `opensearch_dashboards.yml`** fragment is clearer.

**Template fragment** (illustrative — adjust keys for your patch level):

```yaml
opensearch_security.auth.type: "openid"
opensearch_security.openid.connect_url: "http://keycloak-proxy:80/keycloak/realms/master/.well-known/openid-configuration"
opensearch_security.openid.client_id: "wazuh-dashboard"
opensearch_security.openid.client_secret: "${WAZUH_DASHBOARD_OIDC_SECRET}"
opensearch_security.openid.base_redirect_url: "https://your-gateway.example.com/wazuh"
```

Align **`KEYCLOAK_PUBLIC_URL`** / realm with your IAM stack. Use HTTPS in production.

## Zero-disk secrets

- Optional: add **`WAZUH_INDEXER_PASSWORD`**, **`WAZUH_MANAGER_API_PASSWORD`**, **`WAZUH_DASHBOARD_KIBANA_PASSWORD`** to Vault after aligning **`internal_users.yml`**.
- Sonar-related secrets are documented in [SONARQUBE_KEYCLOAK.md](SONARQUBE_KEYCLOAK.md).
