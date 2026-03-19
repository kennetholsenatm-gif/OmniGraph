# Single Pane of Glass

Unified gateway: **Traefik** reverse proxy, **dashboard** (wiki + LMNotebook embed), and **webhook listener** for doc-push refresh. All env can be injected by Ansible; optional `.env` for local overrides.

## WUIs behind the gateway

| Path        | App       | Stack   | Notes |
|------------|-----------|---------|--------|
| `/`        | Dashboard | gateway | Wiki + LMNotebook |
| `/gitea/`  | Gitea     | tooling | Set `GITEA__server__ROOT_URL=http://localhost/gitea/` |
| `/n8n/`    | n8n       | tooling | Set `N8N_PATH=/n8n` so UI and webhooks use the path |
| `/zammad/` | Zammad    | tooling | Set base URL in Zammad config if redirects break |
| `/keycloak/` | Keycloak | IAM     | Set `KC_HTTP_RELATIVE_PATH=/keycloak` in IAM stack |
| `/vault/`  | Vault     | IAM     | Path stripped by Traefik |
| `/bitwarden/` | Bitwarden (Vaultwarden) | tooling | Set `DOMAIN=http://localhost/bitwarden` |
| `/freeipa/` | FreeIPA | identity | LDAP/Kerberos; path stripped. First-run init required. |
| `/portainer/` | Portainer | tooling | Container management; path stripped. On Windows, Docker socket path may need adjustment. |
| `/llm/` | BitNet LLM | llm | OpenAI API compatible; path stripped. See [docs/LLM_BITNET.md](../docs/LLM_BITNET.md). |
| `/docs/` | Docsify (architecture-docs) | tooling | Git-synced static site on `docs_net`. See [docs/DOCSIFY_GITEA.md](../docs/DOCSIFY_GITEA.md). |
| `/sonarqube/` | SonarQube | tooling | SAST; JDBC to messaging Postgres. See [docs/SONARQUBE_KEYCLOAK.md](../docs/SONARQUBE_KEYCLOAK.md). |
| `/wazuh/` | Wazuh dashboard | SIEM (optional) | OpenSearch Dashboards + Wazuh plugin on `siem_net`. See [docs/WAZUH_SIEM.md](../docs/WAZUH_SIEM.md). |

## Quick start (local)

1. **Create networks** (once). Gateway needs all backend networks so Traefik can reach each WUI.
   - **Option A (recommended):** From repo root run `tofu init && tofu apply` in `opentofu/` to create all networks with correct subnets. See [docs/NETWORKS_PHASE1.md](../docs/NETWORKS_PHASE1.md).
   - **Option B:** Run the Phase 1 script from repo root: `.\scripts\create-networks.ps1` (Windows) or `./scripts/create-networks.sh` (Linux/macOS). This creates `gateway_net`, `gitea_net`, `n8n_net`, `zammad_net`, `iam_net`, `llm_net` with subnets from [NETWORK_DESIGN.md](../docs/NETWORK_DESIGN.md). For Bitwarden, FreeIPA, Portainer also run OpenTofu or create those networks manually.
   If you already have these from the tooling/IAM stacks, leave them as-is.

2. **Start backends** so the gateway can proxy to them:
   ```powershell
   cd C:\GiTeaRepos\devsecops-pipeline\docker-compose
   docker compose -f docker-compose.tooling.yml up -d   # Gitea, n8n, Zammad, Bitwarden, Portainer
   docker compose -f docker-compose.iam.yml up -d        # Vault, Keycloak
   docker compose -f docker-compose.identity.yml up -d   # FreeIPA (optional; init required on first run)
   ```

3. **Start the gateway** from the single-pane-of-glass folder:
   ```powershell
   cd C:\GiTeaRepos\devsecops-pipeline\single-pane-of-glass
   docker compose up -d
   ```

4. Open **http://localhost** (port 80). You should see:
   - **/** — Dashboard (Knowledge Base + LMNotebook placeholder)
   - **/gitea/** — Gitea
   - **/n8n/** — n8n
   - **/zammad/** — Zammad
   - **/keycloak/** — Keycloak
   - **/vault/** — Vault
   - **/bitwarden/** — Bitwarden
   - **/freeipa/** — FreeIPA
   - **/portainer/** — Portainer
   - **/llm/** — BitNet LLM (OpenAI API; see [LLM_BITNET.md](../docs/LLM_BITNET.md))
   - **/docs/** — Docsify architecture docs (requires `docs_net` + tooling `docs` / `docs-sync`)
   - **/sonarqube/** — SonarQube (requires `sonarqube_net` + messaging Postgres + `SONAR_JDBC_PASSWORD`)
   - **/wazuh/** — Wazuh dashboard (optional; `siem_net` + `docker-compose/siem/wazuh-config`)

   With TLS certs mounted (see **TLS / HTTPS** below), the same paths are also served on **https://localhost** (port 443).

If a WUI shows 502 Bad Gateway, that backend is not running or not on the expected network. If it loads but redirects break (404 or wrong path), set the base URL / path env for that app as in the table above.

## TLS / HTTPS

Copy `traefik/dynamic/tls.yml.example` to `traefik/dynamic/tls.yml`, then mount the server certificate (and optional client CA for mTLS) into the gateway volume `gateway_tls` so that Traefik can serve HTTPS. Place `tls.crt` and `tls.key` in that volume (container path `/etc/traefik/tls/`). For production, enable HTTP→HTTPS redirect in `traefik/traefik.yml` (uncomment the `http.redirections.entryPoint` block under `entryPoints.web`). See [docs/NETWORKS_PHASE1.md](../docs/NETWORKS_PHASE1.md).

## Hardening: single ingress (production)

For production, make Traefik the sole ingress: remove or comment out the host `ports` for Gitea (3000, 2222), n8n (5678), Zammad (8080) in `docker-compose/docker-compose.tooling.yml`, and the keycloak-proxy port (127.0.0.1:8180) in `docker-compose/docker-compose.iam.yml`. Access all WUIs only via the gateway (e.g. `http://localhost/gitea/`). If you need Git over SSH, keep `2222:22` or use HTTPS-only clone via Traefik.

## Optional env

Copy `.env.example` to `.env` and set if needed:

- `GITEA_API_TOKEN` — for private wiki/repos; create in Gitea → Settings → Applications.
- `GITEA_WIKI_OWNER` / `GITEA_WIKI_REPO` — e.g. `kbolsen` and `devsecops-pipeline` so the Knowledge Base tab can list wiki pages.
- `LMNOTEBOOK_URL` — URL to embed (e.g. your LMNotebook/OpenCode instance).

## Event-driven refresh

When the SAM Doc Agent (or n8n) pushes new docs to Gitea, call:

```http
POST http://localhost/webhook/doc-push
```

(with optional `X-Hub-Signature-256` if `WEBHOOK_HMAC_SECRET` is set). The dashboard will refresh the wiki list over SSE.

**Docsify volume refresh:** With `DOCS_SYNC_ENABLED=true` and `GITEA_DOCS_WEBHOOK_SECRET` set, Gitea can `POST` to `/webhook/docs-sync` (same secret as the Gitea webhook) to start the `devsecops-docs-sync` container. See [docs/DOCSIFY_GITEA.md](../docs/DOCSIFY_GITEA.md).

## Layout

- `traefik/` — static config and dynamic routes (dashboard, gitea, n8n, zammad, keycloak, vault, bitwarden, freeipa, portainer, docs, webhooks).
- `dashboard/` — Node server and static UI (wiki from Gitea API, LMNotebook iframe).
- `webhook-listener/` — receives doc-push, triggers dashboard refresh.
