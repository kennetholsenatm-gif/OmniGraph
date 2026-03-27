# WSL2 + AlmaLinux + LXD/LXC (+ Docker-in-LXC)

> Compatibility path: the canonical local target is now `deployments/local-lxc/`.
> Keep this directory for WSL2-specific notes and legacy links during migration.

**Locked runtime (plan):** **(A) LXD inside AlmaLinux on WSL2** — develop and run the DevSecOps stack in unprivileged/privileged LXC containers with **Docker CE + Compose** inside each workload container.

## Why this path

- Matches the refactor plan before **OpenNebula KVM**: each LXC ≈ a future VM; Compose inside stays the same.
- **Stability caveat:** WSL2 uses Microsoft’s kernel. **LXD + nested Docker** works for many users but is **less supported** than **AlmaLinux on bare metal or Hyper-V VM + LXD**. If you hit cgroup/AppArmor issues, switch to **(B) Alma VM + LXD** using the **same** `profiles/`, Ansible role, and compose paths (no repo change except where you SSH).

## AlmaLinux 10 image sources

| Use case | Image / how to get it |
|----------|------------------------|
| **LXD `lxc launch`** | `images:almalinux/10` or `images:almalinux/10/cloud` (verify `lxc image list images: almalinux` on your host). |
| **WSL distro** | Microsoft Store does not ship Alma; use **official rootfs** import (AlmaLinux WSL wiki / `AlmaLinux-WSL`) or run **Alma in LXC only** and use **Windows Terminal + `wsl -d Ubuntu`** only for editing—**recommended:** primary shell **inside** `lxc exec devsecops-admin -- bash` on Alma. |
| **Docker base** | `almalinux:10` from Docker Hub for inner containers (unchanged). |

Confirm the image alias exists before automation (use **`lxc`** or **`incus`** depending on your install):

```bash
incus image copy images:almalinux/10 local: --copy-aliases --auto-update
# or: lxc image copy images:almalinux/10 local: --copy-aliases --auto-update
```

## Prerequisites (WSL2 host)

1. **WSL2** with a **systemd-enabled** distro (e.g. Ubuntu 24.04 on WSL with `[boot] systemd=true` in `wsl.conf`).
2. **Container hypervisor on the same Linux where you run Ansible** (the role runs `lxc` or `incus` on the controller):
   - **Ubuntu (WSL):** `sudo snap install lxd` then `sudo lxd init` (CLI often `/snap/bin/lxc`).
   - **AlmaLinux / Rocky / RHEL-compatible 10:** BaseOS + CRB **do not** include `incus`. Use **[COPR `neelc/incus`](https://copr.fedorainfracloud.org/coprs/neelc/incus/)** — for EL10, **download the repo file** (do not rely on `dnf copr enable` alone):

     ```bash
     sudo dnf install -y epel-release
     sudo dnf config-manager --set-enabled crb
     cd /etc/yum.repos.d
     sudo curl -fsSLO https://copr.fedorainfracloud.org/coprs/neelc/incus/repo/rhel+epel-10/neelc-incus-rhel+epel-10.repo
     sudo dnf install -y incus
     ```

     Add to **`/etc/subuid`** and **`/etc/subgid`** (per [COPR instructions](https://copr.fedorainfracloud.org/coprs/neelc/incus/)):

     ```text
     root:1000000:65536
     ```

     Enable and initialize (unit names depend on the RPM; list with `systemctl list-unit-files '*incus*'`):

     ```bash
     sudo systemctl enable --now incus.socket incus.service
     sudo incus admin init
     ```

     Ansible discovers **`/usr/bin/incus`** once installed; use `-e 'lxd_cli=/usr/bin/incus'` if needed.
3. **Storage/network init** — on WSL, **`dir`** storage is simplest; default bridge is usually fine.
4. Profile **`docker-nesting`** — see [`profiles/docker-nesting.yaml`](profiles/docker-nesting.yaml); the **`lxd_devsecops_stack`** role applies an equivalent profile named `docker-nesting` automatically.
5. **Firewall:** allow **forwarding** between **`lxdbr0` / `incusbr0`** and WSL; publish ports from Windows via **mirrored networking** or `netsh interface portproxy`.
6. **LF line endings** for `*.sh` under `/mnt/c/`: CRLF breaks the shebang (`env: 'bash\r': No such file`). The repo uses `.gitattributes` (`*.sh text eol=lf`); after checkout, run `git add --renormalize .` or `sed -i 's/\r$//' path/to/script.sh` if needed.

### `wsl --update` from Linux shows “command not found”

Run **`wsl.exe --update`** from **Windows** (PowerShell or cmd), not from inside the Linux shell. From Linux you can call **`/mnt/c/Windows/System32/wsl.exe --update`**.

### Incus: `dial unix /var/lib/incus/unix.socket` / “failed to connect to local daemon”

1. **Daemon not running:** `sudo systemctl status incus.socket incus.service` — fix errors, then `sudo systemctl enable --now incus.socket incus.service`.
2. **WSL without systemd:** In **`/etc/wsl.conf`** set **`[boot]`** **`systemd=true`**, then **restart WSL from Windows** (`wsl --shutdown`, reopen distro). Run **`systemctl is-system-running`** — it should not say *offline*.
3. **Wrong socket path (common on RHEL/Alma COPR):** the client may default to **`/var/lib/incus/unix.socket`** while the RPM listens on **`/run/incus/unix.socket`**. Check: `ls -l /run/incus/unix.socket /var/lib/incus/unix.socket`. Then either:
   - `export INCUS_SOCKET=/run/incus/unix.socket` and `sudo -E incus admin init`, or
   - pass **`lxd_incus_socket: /run/incus/unix.socket`** to the **`lxd_devsecops_stack`** role / `-e lxd_incus_socket=...`.

The Ansible role waits for either path and sets **`INCUS_SOCKET`** for all `incus` calls when it finds a socket.

## Layout

| Path | Purpose |
|------|---------|
| [`profiles/docker-nesting.yaml`](profiles/docker-nesting.yaml) | LXD profile: nesting, optional kernel modules |
| [`cloud-init/docker-and-compose.user-data.yaml`](cloud-init/docker-and-compose.user-data.yaml) | Optional cloud-init: Docker CE + compose plugin |
| [`scripts/bootstrap-lxd-profile.sh`](scripts/bootstrap-lxd-profile.sh) | Apply profile from YAML via `lxc profile edit` |
| [`../opennebula-kvm/VLAN_MATRIX.md`](../opennebula-kvm/VLAN_MATRIX.md) | North-star IP plan when you move to hardware |

## Stack layout (LXC per compose group)

See [docs/WSL2_LXC_GATEWAY.md](../../docs/WSL2_LXC_GATEWAY.md) for **Traefik** placement.

| Instance (example) | Compose / scope |
|--------------------|-----------------|
| `devsecops-iam` | `docker-compose/docker-compose.iam.yml` |
| `devsecops-messaging` | `docker-compose/docker-compose.messaging.yml` |
| `devsecops-tooling` | `docker-compose/docker-compose.tooling.yml` |
| `devsecops-gateway` | `single-pane-of-glass/docker-compose.yml` |
| `devsecops-chatops` | `docker-compose/docker-compose.chatops.yml` |
| `devsecops-telemetry` | `docker-compose/docker-compose.telemetry.yml` (+ optional SDN) |

Optional stacks: discovery, SIEM, LLM, identity, AI orchestration — same pattern.

## Automation

**Collections:** install from the **`ansible/`** directory, or pass a path relative to your shell CWD (from **`ansible/playbooks/`** use **`-r ../collections/requirements.yml`**).

The role uses **`become: true`** only for **systemd + `incus`/`lxc`** (repo checks run as your user). If sudo **requires a password**, use **`ansible-playbook -K`**, or **`./run-deploy-devsecops-lxc.sh`** (sets **`ANSIBLE_CONFIG`** + **`ANSIBLE_BECOME_ASK_PASS`**), or **`export ANSIBLE_BECOME_ASK_PASS=true`** before `ansible-playbook`. Otherwise Ansible may **time out** waiting for sudo with no TTY (`Timed out waiting for become success`).

```bash
cd ansible
ansible-galaxy collection install -r collections/requirements.yml
ansible-playbook -K -i inventory/lxc.example.yml playbooks/deploy-devsecops-lxc.yml --tags iam
```

See [docs/WSL2_LXC_MIGRATION.md](../../docs/WSL2_LXC_MIGRATION.md) for **IAM → messaging** order and checks.

## Secrets

Use [`scripts/secrets-bootstrap.sh`](../../scripts/secrets-bootstrap.sh) inside the LXC (or host with Docker pointed at LXC—prefer **per-LXC** run). Full break-glass / Bitwarden flows remain in PowerShell unless `bw` + `curl` paths are extended in bash.
