# Phase 3: Gitea Push → SBOM/Trivy Scan → Zammad

This document describes the automation pipeline that triggers on Gitea push events, runs Syft (SBOM) and Trivy (vulnerability scan), and creates a Zammad ticket when CRITICAL vulnerabilities are found.

## Workflow

- **Trigger:** Gitea webhook on **push** events.
- **Webhook URL:** `http://n8n:5678/webhook/gitea-push` (when n8n is reached at `http://n8n:5678`; use your Traefik/public URL if ingress is different).
- **Flow:** Parse payload → Execute Command (clone repo, Syft, Trivy) → Parse Trivy JSON → If any CRITICAL → Create Zammad ticket → Respond 202; else Respond 200.

## Gitea webhook setup

1. In Gitea, open the repository → **Settings** → **Webhooks** → **Add Webhook** → **Gitea**.
2. **Target URL:** `http://n8n:5678/webhook/gitea-push` (from inside the Docker network) or the equivalent URL that reaches n8n (e.g. via Traefik: `https://your-domain/n8n/webhook/gitea-push`).
3. **HTTP Method:** POST.
4. **Trigger:** Choose **Push events**.
5. **Secret (optional):** Set a secret and configure the same value in n8n if you use HMAC verification for the webhook.

## n8n configuration

- **Credentials:** Ensure the **Zammad API** credential exists (Header Auth with token). The workflow references it by name `zammad_api`.
- **Environment:** `ZAMMAD_URL` (default `http://zammad`), `GITEA_URL` (for display), and `GITEA_API_TOKEN` (for cloning private repos; injected into clone URL when set).

## Execute Command requirement

The workflow uses the **Execute Command** node to run, inside the n8n runtime:

1. `git clone` of the repository (branch from the push payload).
2. `syft . -o json` (SBOM; output is not used by the workflow but runs for completeness).
3. `trivy fs . --severity CRITICAL -f json` (output is parsed to detect CRITICAL vulns).

**Implications:**

- Commands run **inside the n8n container** (or the host, if n8n is not in Docker). The stock `n8nio/n8n` image does **not** include `git`, `syft`, or `trivy`.
- **Option A (boilerplate):** Use a **custom n8n image** that installs `git`, [Syft](https://github.com/anchore/syft), and [Trivy](https://github.com/aquasecurity/trivy), and ensure the Execute Command node is **enabled** (it is disabled by default in n8n 2.0+ for security).
- **Option B (production):** Run a separate **scanner service** (e.g. a small container with git, Syft, and Trivy) that exposes an HTTP endpoint accepting `clone_url` and `ref`; the workflow then uses an **HTTP Request** node instead of Execute Command to call that service and parse the returned Trivy JSON. This avoids enabling Execute Command and keeps the n8n image unchanged.

## Zammad ticket shape

When CRITICAL vulnerabilities are found, the workflow creates a ticket via `POST /api/v1/tickets` with:

- **Title:** `CRITICAL vulnerabilities: <repo full_name> (<branch>)`
- **Group:** `Users`
- **Customer:** pusher email from the webhook payload
- **Article:** subject and body containing repository, branch, commit SHA, pusher, and a list of critical CVEs (VulnerabilityID, PkgName, Title).

## Artifacts

| Item | Location |
|------|----------|
| Workflow JSON | `n8n-workflows/gitea-push-sbom-scan.json` |
| Deployment reference | [DEPLOYMENT.md](DEPLOYMENT.md) (step 5, Artifacts Summary) |
