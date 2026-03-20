# Docsify + webhooks — post-migration checklist

Use after Gitea moves to OpenNebula (**100.64.1.0/24**). Full context: [DOCSIFY_GITEA.md](../DOCSIFY_GITEA.md).

## Clone / sync

- [ ] **`DOCS_GIT_REPO`** resolves to the **new** Gitea URL or in-compose `http://gitea:3000/...` (Docker) / VM DNS name for OpenNebula deployment.
- [ ] **Deploy tokens** or credentials in secrets still valid after user DB restore.

## Webhooks

- [ ] Gitea webhook URL reaches gateway **`/webhook/docs-sync`** (or HTTPS URL behind Traefik on **100.64.5.0/24**).
- [ ] **`GITEA_DOCS_WEBHOOK_SECRET`** matches Gitea webhook **Secret**.
- [ ] **`X-Gitea-Signature`** (HMAC-SHA256) verifies — body must be **unchanged** end-to-end (avoid n8n rewriting; see [n8n-workflows/README.md](../../n8n-workflows/README.md) if relaying).

## CI / repo hygiene

- [ ] No **`C:\GiTeaRepos`** or other Windows paths in workflows, hooks, or docs build scripts.
- [ ] **`ROOT_URL`** and public clone URLs match DNS/TLS after cutover.

## Smoke test

1. Push a commit to the architecture-docs repo.
2. Confirm **`docs-sync`** runs (container start or job log).
3. Load **`/docs`** (or host rule) and verify new content.
