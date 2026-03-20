# Docsify architecture docs (GitOps, zero image rebuild)

This stack serves a **Docsify** site from a **Gitea** repository without building a custom image: a one-shot **`docs-sync`** container clones or pulls into a named volume, and **`docs`** (Nginx) serves it on **`docs_net`**. Traefik exposes the site at **`/docs`** (see `single-pane-of-glass/traefik/dynamic/routes.yml`).

OpenNebula migration cutover: after Gitea moves to **100.64.1.0/24**, run [opennebula-gitea-edge/DOCSIFY-POST-MIGRATION-CHECKLIST.md](opennebula-gitea-edge/DOCSIFY-POST-MIGRATION-CHECKLIST.md).

## Networks and DNS

- **Clone URL** must reach Gitea using the **Compose service hostname** `gitea` (e.g. `http://gitea:3000/Org/repo.git`), not `devsecops-gitea` (`container_name` is not the default DNS name).
- **`docs-sync`** is on **`gitea_net` only** (reaches Gitea).
- **`docs`** (Nginx) is on **`docs_net` only**; Traefik joins **`docs_net`** to route `/docs`.

## Environment variables

| Variable | Where | Purpose |
|----------|--------|---------|
| `DOCS_GIT_REPO` | Tooling compose / `.env` / Ansible | Clone URL (HTTP(S) to Gitea) |
| `DOCS_GIT_BRANCH` | Same | Branch (default `main`) |
| `DOCS_SYNC_ENABLED` | Gateway `.env` | `true` to enable `POST /webhook/docs-sync` |
| `GITEA_DOCS_WEBHOOK_SECRET` | Gateway `.env` (secret) | Must match Gitea webhook “Secret” |
| `DOCS_SYNC_CONTAINER_NAME` | Gateway `.env` | Default `devsecops-docs-sync` |
| `DOCS_SYNC_REPO_FULL_NAME` | Gateway `.env` | Optional filter, e.g. `DevSecOps/architecture-docs` |

See also `devsecops.env.schema` (`@env-spec: DOCSIFY`) and `single-pane-of-glass/gateway.env.schema` (`DOCS_SYNC_WEBHOOK`).

## Private repositories

Embed credentials in the clone URL **only via secrets** (Vault / Varlock / Ansible), never in git:

- Example: `http://token:${GITEA_DOCS_CLONE_TOKEN}@gitea:3000/Org/private-docs.git`
- Or use a Gitea deploy token / PAT allowed for `git clone` over HTTP.

`DOCS_GIT_REPO` is interpolated by Compose from the environment.

## Gitea repository layout

At minimum:

- `index.html` — Docsify bootstrap (copy from [snippets/architecture-docs-index.html](snippets/architecture-docs-index.html) and adjust title/repo).
- `_sidebar.md` — if using `loadSidebar: true` in Docsify config.

## After stack start

1. Create external network **`docs_net`** if missing: `scripts/create-networks.ps1` or OpenTofu / Ansible mesh playbook.
2. Start core tooling + gateway (e.g. `docker-compose/launch-stack.ps1` and Single Pane of Glass compose).
3. Open **`http://localhost/docs/`** (or your gateway host).

## Refreshing docs without rebuilding images

**Manual:** `docker start devsecops-docs-sync` (container exits after pull; Nginx immediately serves updated files from the volume).

**Automated (recommended):** In Gitea, add a **Webhook** (push events):

- **URL:** `http://<gateway-host>/webhook/docs-sync` (or `https://.../webhook/docs-sync` behind TLS).
- **Secret:** same value as `GITEA_DOCS_WEBHOOK_SECRET` on the gateway.
- Set **`DOCS_SYNC_ENABLED=true`** for `webhook-listener`.

The listener verifies **`X-Gitea-Signature`** (HMAC-SHA256 of the raw body) and calls the **Docker Engine HTTP API** on **`/var/run/docker.sock`** to `POST /v1.41/containers/<name>/start`. No `docker` CLI is required inside the container.

### Security note

Mounting the Docker socket grants **significant host capability** to `webhook-listener`. Keep **`DOCS_SYNC_ENABLED=false`** when not using this feature, or remove the socket volume from `single-pane-of-glass/docker-compose.yml` if you only use manual sync.

### Optional n8n relay

Prefer **Gitea → gateway** directly so the raw body and signature are unchanged. If you must receive the webhook in n8n first, see [n8n-workflows/README.md](../n8n-workflows/README.md) (**Gitea docs sync → gateway**); you may need to forward the **raw body** and **`X-Gitea-Signature`** header unchanged for verification to succeed.

## Optional: dedicated hostname

This repo’s default is **path-based** routing (`/docs`). For a **Host** rule (e.g. `docs.gateway.local`), add another router in Traefik dynamic config and ensure DNS resolves to the gateway; see `traefik/dynamic/optional-routes.example.yml` for patterns.

## Troubleshooting

| Issue | Check |
|--------|--------|
| 502 / empty site | Repo missing `index.html` or clone failed (logs: `docker logs devsecops-docs-sync`) |
| Clone auth failure | `DOCS_GIT_REPO` credentials; Gitea must allow HTTP git from `gitea_net` |
| Second `docker start` must work | Sync uses **clone-or-fetch+reset**, not plain `clone` only |
| Webhook 401 | Secret mismatch; header must be Gitea’s `X-Gitea-Signature` |
| Webhook 502 | Socket mount, container name, or Docker API permissions |
