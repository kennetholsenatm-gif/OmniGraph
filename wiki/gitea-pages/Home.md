# DevSecOps pipeline wiki

Versioned operator surface for **`devsecops-pipeline`**. Canonical clone: **`C:\GiTeaRepos\devsecops-pipeline`** (Gitea).

## Read first (in repo)

- **Site topology & phases:** `docs/CANONICAL_DEPLOYMENT_VISION.md`, `docs/ROADMAP.md`
- **Execution order:** `docs/DEPLOYMENT.md`
- **Network:** `docs/NETWORK_DESIGN.md`
- **Secrets:** `docs/VARLOCK_USAGE.md` — no secrets in this wiki

## Wiki map

| Page | Purpose |
|------|---------|
| [[Deployment]] | Networks → stacks → Vault / Varlock checklist |
| [[Network]] | 100.64 segments, 17 Docker bridges |
| [[Identity]] | LDAP accounts, groups, Keycloak sync pointers |
| [[Discovery-and-Termius]] | BOM / discovery glossary and flows |
| [[Gateway]] | Traefik paths, doc-push webhook |
| [[Runbooks]] | Incident / ops links (stub) |

## Publish these pages from git

Sources live under `wiki/gitea-pages/` in the repo. From repo root:

```bash
./scripts/publish-gitea-wiki-pages.sh --url http://localhost:3000 --owner kbolsen \
  --repo devsecops-pipeline --token YOUR_PAT
```

See `docs/GITEA_WIKI.md`.
