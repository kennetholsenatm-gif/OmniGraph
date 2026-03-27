# Lean Local Control Plane

Use your laptop as a **control plane only** and run the full runtime stack on OpenNebula.

## Intent

- Keep local resource usage low.
- Keep local workflow cognitively simple.
- Preserve strict production behavior on OpenNebula.

## Mode split

| Mode | Runs where | Purpose | Secrets strictness |
|---|---|---|---|
| Lean local control | Laptop | Semaphore UI, Ansible, lint/security/reporting tools | Relaxed dev defaults allowed |
| OpenNebula runtime | OpenNebula LXCs/K3s | Full DevSecOps runtime stacks | Strict Vault/env posture |

## Local components (recommended)

- Semaphore + Postgres on **Incus** (native systemd; no Docker sidecar)
- Ansible controller + repo playbooks
- Pre-commit hooks: `tflint`, `ansible-lint`, `yamllint`
- KICS scan scripts
- Inframap topology export
- Ansible-CMDB report generation
- Optional Digger workflow for IaC planning

## Quick start (local)

1. Start local control services (pre-commit / collections if available; Semaphore is **Incus-based**, not Docker):

```powershell
./scripts/setup-lean-local-control.ps1
```

2. Provision **Semaphore on Incus** (native Postgres + systemd — no Docker):

```powershell
.\scripts\start-semaphore.ps1
```

Linux/WSL:

```bash
./scripts/start-semaphore.sh
```

3. Run quality/security checks:

```powershell
pre-commit run --all-files
./scripts/scan-kics.ps1
```

4. Generate visibility artifacts:

```powershell
./scripts/generate-inframap.ps1
./scripts/generate-ansible-cmdb.ps1
```

5. Deploy to OpenNebula:

```powershell
cd ansible
ansible-playbook -i inventory/opennebula-hybrid.yml playbooks/opennebula-hybrid-site.yml -K
```

## Local Incus cleanup (control-plane + Gitea only)

Stop non-essential local LXCs and keep only the Gitea instance:

```powershell
./scripts/trim-local-incus.ps1 -KeepInstances devsecops-gitea
```

## Restore Gitea on LXC

Provision/refresh only the Gitea LXC target:

```powershell
./scripts/restore-gitea-lxc.ps1 -Inventory inventory/lxc.example.yml -ComposeUp
```

Runbook: [../LOCAL_GITEA_LXC_RUNBOOK.md](../LOCAL_GITEA_LXC_RUNBOOK.md)

## Semaphore mapping (control-only)

**UI issues (e.g. `ERR_EMPTY_RESPONSE` on `localhost:3001`):** see [SEMAPHORE_INCUS_TROUBLESHOOTING.md](../SEMAPHORE_INCUS_TROUBLESHOOTING.md) — try **`http://127.0.0.1:3001`** or re-run deploy so the IPv6 proxy is present.

Provision Semaphore with `./scripts/start-semaphore.sh` (Ansible + Incus). Create Semaphore templates/projects targeting:

- `ansible/playbooks/deploy-devsecops-lxc.yml`
- `ansible/playbooks/opennebula-hybrid-site.yml`
- `ansible/playbooks/keycloak-iam.yml`
- **Mini PC firewall image / host:** [deployments/mini-pc-firewall/semaphore/TEMPLATE-EXAMPLE.md](../../deployments/mini-pc-firewall/semaphore/TEMPLATE-EXAMPLE.md) (`packer-build-mini-pc-incus-host.yml`, `mini-pc-firewall-host.yml`)

Do not schedule local full-stack compose runs from Semaphore on the laptop.

## Inframap (primary topology)

Use Inframap against OpenTofu/Terraform plan/state artifacts when present.
Scripts write outputs under `docs/artifacts/inframap/`.

## Ansible-CMDB

Generate static host/group reports from your selected inventory.
Scripts write outputs under `docs/artifacts/ansible-cmdb/`.

## Digger (optional)

Digger remains optional in this repo:

- Use for IaC planning/review workflow where needed.
- Keep execution surface minimal; do not couple Digger to runtime services on laptop.

## Notes

- This model reduces local CPU/RAM pressure by removing runtime stacks from laptop.
- It also keeps production strictness untouched in OpenNebula deployment paths.

## Repo split policy

- Greenfield bootstrap content moved to `C:\GiTeaRepos\Deploy`.
- Runtime/control-plane operations remain in `C:\GiTeaRepos\devsecops-pipeline`.

For greenfield entrypoints, use:

- `C:\GiTeaRepos\Deploy\scripts\launch-greenfield.ps1`
- `C:\GiTeaRepos\Deploy\scripts\secrets-bootstrap.ps1`
