# Semaphore templates — Mini PC firewall (Packer + Ansible)

Use this as a **blueprint** in [Semaphore UI](https://semaphoreui.com/) (v2). Adjust paths if your project checks out the repo somewhere other than `/workspace`.

## Prerequisites

- Semaphore runner is **Linux with KVM** (`/dev/kvm`) for Packer+QEMU builds **or** use a remote runner / SSH inventory target that has KVM.
- `packer`, `qemu-system-x86_64`, `ansible-playbook` on the runner `PATH`.
- Optional: bind-mount this repo into the Semaphore LXC at `/workspace` (`semaphore_host_repo_path` in `deploy-semaphore-incus.yml`).

## Project: `mini-pc-firewall`

| Setting | Example |
|--------|---------|
| Repository URL | Your Gitea/Git remote for `devsecops-pipeline` |
| Branch | `main` |

## Template 1: `packer-build-alma10-incus-host`

**Purpose:** Produce `deployments/mini-pc-firewall/packer/output-alma10-incus-host/` QCOW2.

| Field | Value |
|-------|--------|
| Type | **Shell** |
| Playbook / script | *(use “Task” below)* |

**Environment variables (optional)**

| Name | Value |
|------|--------|
| `PACKER_LOG` | `1` |

**Task (shell)**

```bash
set -euo pipefail
cd /workspace/deployments/mini-pc-firewall/packer
packer init .
packer validate .
packer build -force .
```

**Artifacts:** archive `output-alma10-incus-host/` or copy `*.qcow2` to your image library (S3, Proxmox, etc.).

## Template 2: `configure-mini-pc-firewall-host`

**Purpose:** Install/configure **Incus** on the live mini PC (SSH).

| Field | Value |
|-------|--------|
| Type | **Ansible** |
| Inventory | Paste or attach `ansible/inventory/mini-pc-firewall.yml` (from example) |
| Playbook | `ansible/playbooks/mini-pc-firewall-host.yml` |
| Vault / limits | As needed |

**Extra CLI args**

```text
-e mini_pc_incus_copr=neelc/incus
```

(Verify the COPR exists for your EL release; override or set `mini_pc_incus_install_method=none` and install Incus manually.)

## Template 3: `packer-build-via-ansible` (wrapper)

**Purpose:** Run the repo playbook so Semaphore only needs Ansible.

| Field | Value |
|-------|--------|
| Type | **Ansible** |
| Inventory | `localhost,` (static) |
| Playbook | `ansible/playbooks/packer-build-mini-pc-incus-host.yml` |

**Requirements:** same KVM host as the runner; set `ansible_connection=local` in inventory.

## Suggested order

1. **Template 1** or **3** — golden image  
2. Flash / import QCOW2 to the mini PC  
3. **Template 2** — Incus + host tuning  
4. Follow [EDGE-MINI-PC-VYOS-PACKETFENCE.md](../../../docs/opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md) for VyOS / PacketFence / RatTrap

## See also

- [deployments/mini-pc-firewall/README.md](../README.md)
- [LEAN_LOCAL_CONTROL_PLANE.md](../../../docs/opennebula-gitea-edge/LEAN_LOCAL_CONTROL_PLANE.md)
