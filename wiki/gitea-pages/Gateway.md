# Single pane of glass (gateway)

From `single-pane-of-glass/README.md`. Traefik + dashboard + webhooks.

## Paths (via gateway)

| Path | Backend |
|------|---------|
| `/` | Dashboard (wiki list + LMNotebook embed) |
| `/gitea/` | Gitea |
| `/n8n/` | n8n |
| `/zammad/` | Zammad |
| `/keycloak/` | Keycloak |
| `/vault/` | Vault |
| `/bitwarden/` | Bitwarden |
| `/freeipa/` | FreeIPA (optional) |
| `/portainer/` | Portainer |
| `/llm/` | BitNet / LLM |
| `/docs/` | Docsify |
| `/sonarqube/` | SonarQube |
| `/wazuh/` | Wazuh |

Set each app’s **base URL / path** env so redirects work behind Traefik (see README table).

## Wiki / Knowledge Base env

- `GITEA_URL` — e.g. `http://gitea:3000` from inside Compose
- `GITEA_API_TOKEN` — private repos or rate limits
- `GITEA_WIKI_OWNER` / `GITEA_WIKI_REPO` — e.g. `kbolsen` / `devsecops-pipeline`

Dashboard calls Gitea API: `GET .../wiki/pages` and `GET .../wiki/page/{slug}`.

## Doc push (refresh wiki list)

```http
POST http://localhost/webhook/doc-push
```

Optional HMAC: `X-Hub-Signature-256` if `WEBHOOK_HMAC_SECRET` is set.

## Docsify refresh

`DOCS_SYNC_ENABLED=true` and Gitea webhook to `/webhook/docs-sync` — see `docs/DOCSIFY_GITEA.md`.
