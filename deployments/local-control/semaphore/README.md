# Local Semaphore (Lean Control Plane)

**Supported path:** **Incus (LXC)** — Semaphore + PostgreSQL installed **natively** on AlmaLinux (systemd). **No Docker.**

The legacy `docker-compose.yml` in this directory is **optional / transitional** only. Prefer Ansible + Incus.

## Why this exists

- Offload full runtime services to OpenNebula.
- Keep laptop usage low and workflow simple.
- Provide a local UI to trigger Ansible playbooks that target OpenNebula.

## Start (Incus / WSL)

From repo root:

```bash
./scripts/start-semaphore.sh
```

Or from `ansible/`:

```bash
ansible-playbook -i inventory/lxc.example.yml playbooks/deploy-semaphore-incus.yml \
  -e lxd_become=false -e lxd_manage_daemon=false -e lxd_ensure_idmap=false \
  -e lxd_incus_socket=/run/incus/unix.socket \
  -e 'lxd_apply_names=["devsecops-semaphore"]'
```

**UI (default Incus proxy):** `http://127.0.0.1:3001`

Default credentials are defined in `ansible/roles/semaphore_native/defaults/main.yml` (change if you expose beyond localhost).

## Optional: bind-mount this repo into the LXC at `/workspace`

WSL example:

```bash
./scripts/start-semaphore.sh -e semaphore_host_repo_path=/mnt/c/GiTeaRepos/devsecops-pipeline
```

Then point Semaphore task templates at `/workspace/ansible/playbooks/...`.

## Recommended project templates

- `/workspace/ansible/playbooks/deploy-devsecops-lxc.yml`
- `/workspace/ansible/playbooks/opennebula-hybrid-site.yml`
- `/workspace/ansible/playbooks/keycloak-iam.yml`

## Legacy: Docker Compose (not recommended)

Only if you cannot use Incus:

```bash
cd deployments/local-control/semaphore
docker compose up -d
```

This maps host `3001` → container `3000` and bind-mounts the repo at `/workspace` read-only.

## Scope guardrail

Do not add the full DevSecOps runtime stacks to this compose file.
This directory is intentionally **control-plane-only**.
