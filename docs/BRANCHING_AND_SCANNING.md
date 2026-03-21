# Branching, path filters, and scanning

This repo mixes **Ansible**, **OpenTofu/Terraform**, **LXC/Incus** playbooks, **Docker Compose**, and small **Python** utilities. You can organize CI in two ways.

## Option A — Single `main` (simplest)

Keep one default branch and use **path filters** in Gitea Actions (or job-level `paths:` in workflows) so a change under `ansible/` does not always run `tflint`, and vice versa.

- **Pros:** One merge queue, one Semaphore manifest, straightforward GitOps.
- **Cons:** Large clones; every contributor needs the full tree (or sparse checkout locally).

## Option B — Stack branches (your “scan LXC like Docker” idea)

Use **long-lived integration branches** (or strict naming) so each area gets the right scanners without noise:

| Branch pattern | Typical contents | Scan focus |
|----------------|------------------|------------|
| `main` | Release / integration | Full pipeline on merge (optional) |
| `stack/ansible` (or `integration/ansible`) | `ansible/` only or majority | `ansible-lint`, Semaphore sync, Ansible-only gates |
| `stack/opentofu` (or `integration/opentofu`) | `opentofu/` | `tflint`, `trivy` misconfig on HCL |
| `stack/lxc` (or `integration/lxc`) | `deployments/local-lxc/`, `ansible/playbooks/deploy-*.yml`, Incus/LXD | **Trivy** `fs` on YAML/Ansible + cloud-init; treat artifacts like IaC, not container images unless you build images in-branch |

**LXC vs Docker:** Trivy’s **image** scan applies to OCI images. For LXC/Incus you usually run **`trivy fs`** on:

- Playbooks and `group_vars` / `host_vars`
- Cloud-init / kickstart snippets
- Any **Dockerfile** or **Containerfile** used to build sidecar images

If a branch builds a **golden image** (Packer/QCOW2), add a CI job that runs `trivy image` on that artifact the same way you would for Docker.

**Merge strategy:** Open PRs from `stack/*` → `main` after checks pass; enable **branch protection** on `main` (required status checks).

## Gitea branch protection (recommended)

In **Repository → Settings → Branches → Branch protection** for `main` (and optionally `stack/*`):

- Require pull request before merging
- **Require status checks** to pass (CI workflow, lint jobs)
- Disallow force-push and deletion

Exact labels depend on your Act runner and workflow names in `.gitea/workflows/`.

## Pre-commit vs CI

| Tool | Pre-commit | CI (`ci.yml`) |
|------|------------|----------------|
| Black | `commit` | `black-check` job |
| Gitleaks | `commit` | `gitleaks` job |
| Trivy `fs` | **`pre-push`** (needs Docker) | `trivy-fs` job |
| ansible-lint / yamllint / tflint | `commit` | `lint-and-ansible` / `tflint` jobs |

Install **pre-push** hooks after cloning:

```bash
pre-commit install --hook-type pre-push
```

(`scripts/setup-lean-local-control.ps1` does this when `pre-commit` is available.)

To skip a hook once: `SKIP=trivy-fs git push`.

## Related

- [CI_CD.md](CI_CD.md) — Gitea Actions, secrets, Semaphore
- [.pre-commit-config.yaml](../.pre-commit-config.yaml) — hook list and versions
