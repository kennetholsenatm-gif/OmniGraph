# Local Gitea on LXC runbook

This runbook restores local Gitea in Incus/LXD while keeping local runtime lean.

## Preconditions

- WSL/Alma controller with `incus` or `lxc`
- Ansible available in controller environment
- Inventory prepared (`ansible/inventory/lxc.example.yml` or your local file)
- **Incus bridge DNS/NAT:** if containers cannot resolve public names, tune the managed bridge first — [INCUS_NETWORK_DNS.md](INCUS_NETWORK_DNS.md) (or run `./scripts/setup-incus-network-dns.sh` in WSL).

## 1) Trim local Incus to Gitea only

```powershell
./scripts/trim-local-incus.ps1 -KeepInstances devsecops-gitea
```

## 2) Restore/provision Gitea LXC

```powershell
./scripts/restore-gitea-lxc.ps1 -Inventory inventory/lxc.example.yml -ComposeUp
```

Linux equivalent:

```bash
COMPOSE_UP=true ./scripts/restore-gitea-lxc.sh inventory/lxc.example.yml
```

This script is tuned for non-interactive local use (`lxd_become=false`, `lxd_manage_daemon=false`, `lxd_ensure_idmap=false`) so it should not require a sudo/become password prompt when your user already has Incus socket access.

## 3) Validate

- `incus list` (or `lxc list`) shows `devsecops-gitea` running.
- Gitea UI reachable on `http://127.0.0.1:3000` (or your forwarded endpoint).
- Git SSH reachable on `127.0.0.1:2222`.

## 4) Persistence checks

- Verify compose volumes in `docker-compose.gitea.yml`:
  - `gitea_data`
  - `gitea_db_data`
- Confirm data survives container/service restart.

## 5) Restore `kbolsen/devsecops-pipeline` on local Gitea

This creates the `kbolsen` user (if missing), an **empty** `devsecops-pipeline` repo, and prints `git remote` / `git push` hints.

**WSL / Linux (from repo root):**

```bash
export GITEA_ADMIN_PASSWORD='your-admin-password'
./scripts/setup-kbolsen-devsecops-repo.sh
```

Optional: `KBOLSEN_PASSWORD='...'` to set the owner password explicitly; otherwise a random password is printed once.

**Windows (PowerShell):**

```powershell
$env:GITEA_ADMIN_PASSWORD = 'your-admin-password'
wsl -e bash -lc "cd /mnt/c/GiTeaRepos/devsecops-pipeline && ./scripts/setup-kbolsen-devsecops-repo.sh"
```

Then push your working tree:

```bash
cd /mnt/c/GiTeaRepos/devsecops-pipeline
git remote remove gitea 2>/dev/null || true
git remote add gitea http://kbolsen@127.0.0.1:3000/kbolsen/devsecops-pipeline.git
git push -u gitea HEAD:main   # or HEAD:master if that is your default branch
```

Open: `http://127.0.0.1:3000/kbolsen/devsecops-pipeline`

**Manual alternative:** sign in as admin → **Site Administration** → **Users** → create `kbolsen` → **New Repository** under that user → add the `git remote` above and push.
