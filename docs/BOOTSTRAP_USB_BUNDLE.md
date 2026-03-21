# Bootstrap USB bundle (offline-first Mini PC bring-up)

**Practical approach:** Before investing in a **custom live ISO**, ship a **data partition on USB** (or a tarball you `rsync` to the box) that contains a **frozen copy of this repo**, **vendored Ansible collections**, optional **`docker save` image tarballs**, and **one entrypoint script**. That gives you Ansible (and optionally OpenTofu) **without** depending on lab routing, Gitea, or Semaphore being up.

**Primary use case:** Bring up **Mini-PC-IAM** (or any Alma host) when **WAN or internal routing is not ready** yet. After Vault exists and KV is populated, switch to normal Varlock / controller flows.

See also: [deployments/bootstrap-usb-bundle/README.md](../deployments/bootstrap-usb-bundle/README.md) (build/run quick reference).

## What goes on the USB (layout)

Suggested top-level directory (e.g. mount at `/mnt/usb` or copy to `/opt/bootstrap-bundle`):

| Path | Contents |
|------|----------|
| `repo/` | Snapshot of `devsecops-pipeline` (git archive or `rsync`; pin commit SHA in a manifest). |
| `collections/` | Output of `ansible-galaxy collection install -r ansible/collections/requirements.yml -p collections` (run on a **connected** build machine; produces `ansible_collections/` layout for `ANSIBLE_COLLECTIONS_PATH`). |
| `images/` | Optional: `docker save` → `*.tar` for IAM (and deps) if the target has **no** registry access at bootstrap time. |
| `secrets/` | **Optional second medium only:** Ansible Vault file, or empty — **do not** put production secrets on a public ISO. |
| `bootstrap-on-target.sh` | Copy from [deployments/bootstrap-usb-bundle/scripts/bootstrap-on-target.sh](../deployments/bootstrap-usb-bundle/scripts/bootstrap-on-target.sh). |

**OpenTofu:** For Mini-PC-IAM you can create Docker bridges with [`scripts/create-networks.sh`](../scripts/create-networks.sh) or [`opentofu/`](../opentofu/) from `repo/`. Prefer **one path** in your runbook to limit moving parts; the bundle script defaults to **shell `create-networks.sh`** when Docker is available.

## Build the bundle (connected machine)

On a Linux host with **Ansible** / **ansible-galaxy** (same major as target, if possible), **git**, **rsync**, and (optional) **Docker**:

```bash
cd deployments/bootstrap-usb-bundle/scripts
chmod +x build-bundle.sh
./build-bundle.sh                    # writes ../out/devsecops-bootstrap-YYYYMMDD/
```

Copy `out/devsecops-bootstrap-*` to the USB filesystem or tarball it:

```bash
tar -C deployments/bootstrap-usb-bundle/out -czf devsecops-bootstrap.tgz devsecops-bootstrap-YYYYMMDD
```

**Pre-pull images (optional):** On the build machine, with the same compose tags as production:

```bash
cd repo/docker-compose
docker compose -f docker-compose.iam.yml pull
docker save $(docker images --format '{{.Repository}}:{{.Tag}}' | grep -E 'vault|keycloak|teleport|postgres' ) -o ../bundle/images/iam-core.tar
```

Adjust image list to match your `docker-compose.iam.yml` pins; split into multiple `.tar` files if needed.

## Run on the target (Mini-PC-IAM)

**Assumptions:** AlmaLinux (or RHEL-family) is **installed on disk**, **Docker** and **Docker Compose plugin** are installed (or install them once by hand / kickstart). **Incus/LXD** is required if you use [`deploy-devsecops-lxc.yml`](../ansible/playbooks/deploy-devsecops-lxc.yml) for `devsecops-iam` LXC; otherwise use bare-metal mesh + IAM compose per your design.

1. Mount USB: `sudo mount /dev/sdX1 /mnt/usb` (adjust device).
2. `cd /mnt/usb` (or wherever the bundle root is).
3. `chmod +x bootstrap-on-target.sh`
4. Run with explicit paths:

```bash
sudo BUNDLE_ROOT=/mnt/usb REPO_SUBDIR=repo ./bootstrap-on-target.sh
```

**Optional one-shot Ansible (LXC `devsecops-iam` only):** install **`ansible-core`** on the target (or use a venv), place an inventory file (copy from [`ansible/inventory/mini-pc-iam.example.yml`](../ansible/inventory/mini-pc-iam.example.yml)), then:

```bash
sudo RUN_ANSIBLE_BOOTSTRAP=1 \
  BOOTSTRAP_INVENTORY=/root/mini-pc-iam.yml \
  ANSIBLE_PLAYBOOK_EXTRA_ARGS='-K' \
  BUNDLE_ROOT=/mnt/usb \
  ./bootstrap-on-target.sh
```

That runs [`ansible/playbooks/bootstrap-mini-pc-iam.yml`](../ansible/playbooks/bootstrap-mini-pc-iam.yml), which imports `deploy-devsecops-lxc.yml` with **`lxd_apply_names: [devsecops-iam]`** by default. Use `ANSIBLE_PLAYBOOK_EXTRA_ARGS` for `--vault-password-file`, `-e` overrides, etc.

Environment variables (see script header):

- `BUNDLE_ROOT` — directory containing `repo/`, `collections/`, optional `images/`.
- `SKIP_DOCKER_LOAD` — set to `1` if no `images/*.tar` or you already pulled.
- `SKIP_CREATE_NETWORKS` — set to `1` if bridges already exist.
- `RUN_ANSIBLE_BOOTSTRAP` — set to `1` to run `bootstrap-mini-pc-iam.yml` after Docker steps.
- `BOOTSTRAP_INVENTORY` — required when `RUN_ANSIBLE_BOOTSTRAP=1`; path to inventory with **`lxd_hosts`**.
- `ANSIBLE_PLAYBOOK_EXTRA_ARGS` — extra arguments for `ansible-playbook` (quoted string; often `-K` for sudo password).

## Secrets policy

- **Public / lab ISO or shared USB:** no plaintext Vault tokens, no `devsecops_secrets` except inside **Ansible Vault** with a **passphrase not stored on the stick**.
- **Split bundle:** “wide” USB (repo + collections + images) + **private** USB or TPM-backed unlock for vault password file.
- After first bootstrapping Vault, rotate away any bootstrap tokens per [VARLOCK_USAGE.md](VARLOCK_USAGE.md).

## When to graduate to a custom live ISO

Freeze the **bundle layout** and `bootstrap-on-target.sh` behavior first. A **live ISO** then only adds: automated **kickstart**, **persistent overlay**, and **auto-run** on boot — same payloads as the bundle.

## References

- [CANONICAL_DEPLOYMENT_VISION.md](CANONICAL_DEPLOYMENT_VISION.md) — IAM placement and ordering.
- [ROADMAP.md](ROADMAP.md) — P2 IAM / physical prerequisites.
- [ansible/playbooks/bootstrap-mini-pc-iam.yml](../ansible/playbooks/bootstrap-mini-pc-iam.yml) — LXC IAM-only wrapper.
- [ansible/collections/requirements.yml](../ansible/collections/requirements.yml) — collection pins to vendor.
- [scripts/create-networks.sh](../scripts/create-networks.sh) — Docker bridges without OpenTofu.
