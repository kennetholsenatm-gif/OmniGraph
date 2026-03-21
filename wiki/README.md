# Gitea wiki sources

Operator-facing markdown for the Gitea wiki lives in **`gitea-pages/`** (one `.md` file per wiki page title).

Publish (bash, from repo root):

```bash
./scripts/publish-gitea-wiki-pages.sh \
  --url http://localhost:3000 \
  --owner kbolsen \
  --repo devsecops-pipeline \
  --token YOUR_PAT
```

Documentation: **docs/GITEA_WIKI.md**.
