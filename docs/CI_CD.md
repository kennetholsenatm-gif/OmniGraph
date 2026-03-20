# CI/CD (Gitea Actions)

This repo uses **Gitea Actions** (`.gitea/workflows/`) so lint and Semaphore sync are **declarative**—configure the forge once, not ad-hoc shell from an agent.

## Prerequisites

1. **Gitea:** Enable **Actions** and register at least one [Act runner](https://docs.gitea.com/usage/actions/act-runner) with the label your workflows use (default in this repo: **`ubuntu-latest`**). Change `runs-on` in `.gitea/workflows/*.yml` if your runner uses a different label (e.g. `self-hosted`).
2. **Runner capabilities**
   - **CI (`ci.yml`):** Docker (pulls `almalinux:9.7-20260129` and `ghcr.io/terraform-linters/tflint:v0.54.0`).
   - **Semaphore sync (`semaphore-sync.yml`):** Python 3 + `apt`/`sudo` (Ubuntu-style image) for the Ansible controller venv, **or** adjust the workflow to match your runner OS.

## Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| [`.gitea/workflows/ci.yml`](../.gitea/workflows/ci.yml) | `push` / `pull_request` to `main` or `master`, `workflow_dispatch` | `ansible-lint`, `yamllint` (AlmaLinux container via [scripts/ci/lint-ansible.sh](../scripts/ci/lint-ansible.sh)); `tflint` on `opentofu/`. |
| [`.gitea/workflows/semaphore-sync.yml`](../.gitea/workflows/semaphore-sync.yml) | `workflow_dispatch` (optional `push` to `main` after you uncomment) | Runs [scripts/ci/semaphore-sync.sh](../scripts/ci/semaphore-sync.sh) → `ansible/playbooks/sync-semaphore-from-manifest.yml`. |

Local parity (no Actions): from repo root run `bash scripts/ci/lint-ansible.sh`.

## Semaphore sync — repository secrets

Add these in **Gitea → Repository → Settings → Secrets** (or organization secrets):

| Secret | Required | Description |
|--------|----------|-------------|
| `SEMAPHORE_API_TOKEN` | Yes | API token with rights to manage projects/templates (see [SEMAPHORE_POPULATE.md](SEMAPHORE_POPULATE.md)). |
| `SEMAPHORE_URL` | Yes | Base URL Semaphore serves the API on, e.g. `http://127.0.0.1:3001` or `http://semaphore.example:3001`. **Must be reachable from the Act runner.** |
| `SEMAPHORE_GIT_URL` | Yes | Git clone URL Semaphore should use for this repository (HTTPS or SSH as your Semaphore supports). |
| `SEMAPHORE_GIT_BRANCH` | No | Branch (defaults to `main` if unset or empty). |

**Networking:** If Semaphore listens only on `127.0.0.1` on your laptop, a **remote** runner cannot call it. Run the Act runner **on the same host** as Semaphore, use a **routable IP/hostname**, or **VPN/tunnel** so `SEMAPHORE_URL` is valid from the runner.

**Browser vs API URL:** Semaphore UI and API should share the same logical host as [`SEMAPHORE_WEB_ROOT`](SEMAPHORE_INCUS_TROUBLESHOOTING.md) expectations—avoid mixing `localhost` vs `127.0.0.1` for assets.

Manifest file `ansible/files/semaphore-manifest.yml` is **gitignored**; the sync playbook bootstraps from `semaphore-manifest.example.yml` if missing. Tokens belong in **secrets** or `-e`, not in committed files.

## Operator checklist

1. Enable Actions and register a runner with the correct `runs-on` label.
2. Push a branch and open a PR — confirm **CI** turns green.
3. Add Semaphore secrets (table above).
4. Run workflow **Semaphore sync** manually once (`workflow_dispatch`); confirm project/repo/templates in Semaphore UI.
5. (Optional) Uncomment `push: branches: [main]` in `semaphore-sync.yml` to sync on every merge.

## Git hygiene

- **Default branch:** `main` (also `master` triggers CI for compatibility).
- **Branch protection (recommended):** Require the **CI** workflow to pass before merge; avoid pushing broken Ansible to `main`.

## Related docs

- [DEPLOYMENT.md](DEPLOYMENT.md) — lean control plane and Semaphore entrypoints.
- [SEMAPHORE_POPULATE.md](SEMAPHORE_POPULATE.md) — API, manifest schema, troubleshooting.
