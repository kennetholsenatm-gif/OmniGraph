# Gitea wiki — setup and publish

Operator-facing wiki for your Gitea repo (e.g. **`kbolsen/devsecops-pipeline`**). Deep reference stays in git under `docs/`; the wiki holds summaries, checklists, and links.

## Prerequisites

1. **Enable wiki** — In Gitea: repository **Settings → Features** (or **Wiki** tab) and create the first page if prompted.
2. **Access token** — User settings → **Applications** → generate a token with permission to edit the repo wiki. Pass it to the scripts with **`--token`** (do not commit it).

## Source pages in this repo

Markdown sources live in **[wiki/gitea-pages/](../wiki/gitea-pages/)** (versioned). Edit there, then publish with the shell script below.

| Wiki title   | Source file                    |
|-------------|---------------------------------|
| Home        | `wiki/gitea-pages/Home.md`      |
| Deployment  | `wiki/gitea-pages/Deployment.md` |
| Network     | `wiki/gitea-pages/Network.md`   |
| Identity    | `wiki/gitea-pages/Identity.md`  |
| Discovery-and-Termius | `wiki/gitea-pages/Discovery-and-Termius.md` |
| Gateway     | `wiki/gitea-pages/Gateway.md`   |
| Runbooks    | `wiki/gitea-pages/Runbooks.md`  |

## Publish (bash)

From repo root (requires **curl** and **python3**):

```bash
chmod +x scripts/publish-gitea-wiki-pages.sh scripts/verify-gitea-wiki.sh
./scripts/publish-gitea-wiki-pages.sh \
  --url http://localhost:3000 \
  --owner kbolsen \
  --repo devsecops-pipeline \
  --token YOUR_PAT
```

Options:

- **`--dry-run`** — List files that would be published; no HTTP calls; **no token**.
- **`--pages-dir PATH`** — Override markdown directory (default: `wiki/gitea-pages`).
- **`--message TEXT`** — Wiki git commit message.

Gitea **1.22+** uses `POST /api/v1/repos/{owner}/{repo}/wiki/new` with JSON `title`, `content_base64` (UTF-8, base64), `message`. Existing pages are updated with `PATCH .../wiki/page/{pageName}`.

## Verify (bash)

```bash
./scripts/verify-gitea-wiki.sh \
  --url http://localhost:3000 \
  --owner kbolsen \
  --repo devsecops-pipeline \
  --token YOUR_PAT
```

## Publish (PowerShell, optional)

```powershell
.\scripts\Publish-GiteaWikiPages.ps1 -Owner kbolsen -Repo devsecops-pipeline -GiteaUrl http://localhost:3000 -Token YOUR_PAT
.\scripts\verify-gitea-wiki.ps1 -Owner kbolsen -Repo devsecops-pipeline -GiteaUrl http://localhost:3000 -Token YOUR_PAT
```

## Gateway Knowledge Base

The dashboard reads the wiki via Gitea’s API. Configure the gateway container with whatever mechanism you already use for secrets (e.g. Vault); see [single-pane-of-glass/README.md](../single-pane-of-glass/README.md). After publishing, you can `POST` **`/webhook/doc-push`** on the gateway to refresh the sidebar (optional HMAC per that README).

## Scope

Do **not** put Vault payloads, live inventory IPs, or secrets in the wiki. See [WIKI_EXPORT/DISCOVERY_AND_TERMIUS_WIKI.md](WIKI_EXPORT/DISCOVERY_AND_TERMIUS_WIKI.md).
