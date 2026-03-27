# Mini PC firewall — Packer image + Semaphore / Ansible deploy

This directory supports building a **golden QCOW2** for the **AlmaLinux 10 + Incus** host that runs **VyOS (LXC)**, **PacketFence (VM)**, and **RatTrap hairpin** per [EDGE-MINI-PC-VYOS-PACKETFENCE.md](../../docs/opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md).

## What gets built

| Stage | Tool | Output |
|-------|------|--------|
| **Image** | [HashiCorp Packer](https://www.packer.io/) + QEMU/KVM | `packer/output-alma10-incus-host/` → `alma10-incus-host.qcow2` |
| **Host config** | Ansible | Incus, sysctl, kernel modules, optional bridge tuning |
| **Orchestration** | [Semaphore](https://semaphoreui.com/) (optional) | CI-style `packer build` + `ansible-playbook` against the mini PC |

**Packer** provisions a **generic cloud** Alma 10 disk with:

- `qemu-guest-agent`, `cloud-init`, `firewalld`, `openssh-server`
- `net.ipv4.ip_forward=1` (router host)
- Base packages useful for automation (`git`, `curl`, `jq`, …)

**Incus** is **not** baked into the image by default (COPR/repo URLs drift). Install it with **`ansible/playbooks/mini-pc-firewall-host.yml`** (or enable the optional Packer shell step after you confirm a supported repo).

## Prerequisites (build host)

- **Linux** with KVM: `/dev/kvm`, `qemu-system-x86_64`, `qemu-img`
- [Packer](https://developer.hashicorp.com/packer/install) **≥ 1.9**
- Packer **QEMU plugin** (`packer init` installs it)

```bash
# Debian/Ubuntu example
sudo apt-get install -y qemu-system-x86 qemu-utils packer
```

**WSL2:** Packer+QEMU on WSL can work with nested virt; bare-metal Linux or a dedicated VM is more reliable. On Windows, run Packer inside **WSL2** or a **Linux VM** with KVM.

## Build the QCOW2

```bash
cd deployments/mini-pc-firewall/packer
packer init .
packer validate .
packer build .
```

Override image URL / checksum if Alma rotates releases:

```bash
packer build \
  -var='cloud_image_url=https://repo.almalinux.org/almalinux/10/cloud/x86_64/images/AlmaLinux-10-GenericCloud-x86_64-latest.x86_64.qcow2' \
  -var='cloud_image_checksum=file:https://repo.almalinux.org/almalinux/10/cloud/x86_64/images/CHECKSUM' \
  .
```

Flash or attach the artifact to your **mini PC** (Proxmox import, `dd` to disk, or virtio in a lab).

## Configure the live mini PC (Incus + host tuning)

From the repo **`ansible/`** directory (see [ansible/ansible.cfg](../../ansible/ansible.cfg)):

```bash
ansible-galaxy collection install -r collections/requirements.yml
ansible-playbook -i inventory/mini-pc-firewall.example.yml playbooks/mini-pc-firewall-host.yml -K
```

Tune variables in inventory or `-e`:

- `mini_pc_incus_install_method`: `copr` (default), `none`, or `package`
- `mini_pc_incus_copr`: COPR to enable (verify on [Fedora Copr](https://copr.fedoraproject.org/) for your EL version)

## Semaphore

Use **[semaphore/TEMPLATE-EXAMPLE.md](semaphore/TEMPLATE-EXAMPLE.md)** to wire:

1. **Build** — `packer init` + `packer build` (controller with KVM)
2. **Configure** — `ansible-playbook … mini-pc-firewall-host.yml` (SSH to mini PC)

Local Semaphore on Incus: [LEAN_LOCAL_CONTROL_PLANE.md](../../docs/opennebula-gitea-edge/LEAN_LOCAL_CONTROL_PLANE.md), `scripts/start-semaphore.sh`.

## Security notes

- Packer uses a **temporary** `packer` / `packer` user + password only during the build; change or delete before production use, or rely on your site **cloud-init** / **kickstart** for real deployments.
- Lock down **SSH**, **firewalld**, and **Incus** ACLs before exposing the host.

## See also

- [EDGE-MINI-PC-VYOS-PACKETFENCE.md](../../docs/opennebula-gitea-edge/EDGE-MINI-PC-VYOS-PACKETFENCE.md)
- [deployments/opennebula-kvm/VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md)
