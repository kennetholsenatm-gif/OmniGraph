# Ansible: Cisco ISR & C3560-CX

Rudimentary automation for **IOS-XE ISR** and **Catalyst C3560-CX** switches, aligned with [deployments/opennebula-kvm/VLAN_MATRIX.md](../deployments/opennebula-kvm/VLAN_MATRIX.md).

## Quick start

**SSH to devices (`network_cli`):** install **`paramiko`** into the same Python as Ansible (or install **`ansible-pylibssh`**). Otherwise tasks fail with `paramiko is not installed`.

```bash
cd ansible
python3 -m pip install -r requirements-controller.txt
# or: python3 -m pip install paramiko
```

```powershell
cd ansible
ansible-galaxy collection install -r collections/requirements.yml
copy inventory\network.example.yml inventory\network.yml
copy group_vars\network_isr.yml.example group_vars\network_isr.yml
copy group_vars\network_c3560cx.yml.example group_vars\network_c3560cx.yml
# Edit IPs, VLANs, credentials (plain group_vars OK if Vault isn't up; don't commit secrets).
ansible-playbook -i inventory/network.yml playbooks/network-site.yml
```

**No HashiCorp Vault:** set `ansible_password` in inventory and `cisco_isr_admin_password` in `group_vars/network_isr.yml`, or use `--ask-pass` for SSH. See `group_vars/network_isr.yml.example`.

**`No authentication methods available`:** `network_cli` has no password or key. Set **`ansible_password`** (plain or Vault) or run **`ansible-playbook … --ask-pass`**. The password must match the IOS user in **`ansible_user`** (often **`iac`**; must match **exact case**, e.g. **`Iac`**).

**Day-0 ISR over USB console:** serial is not SSH. Use **socat** + **`ansible.netcommon.telnet`** — see [ISR_CONSOLE_SOCAT_ANSIBLE.md](ISR_CONSOLE_SOCAT_ANSIBLE.md) (inventory `isr-console-bridge.example.yml`, scripts `socat-console-bridge.ps1` / `.sh`).

## WSL + `/mnt/c` (world-writable)

Ansible **ignores `ansible.cfg` when the current working directory is world-writable** (typical under `/mnt/c`). Fix one of:

- **`export ANSIBLE_CONFIG=/path/to/ansible/ansible.cfg`** before `ansible-playbook`, **and** run from a **non–world-writable** directory (e.g. `cd ~` then use absolute paths to `-i` and the playbook), **or**
- Use **`ansible/playbooks/run-network-isr.sh`** (it sets `ANSIBLE_CONFIG`, `ANSIBLE_HOST_KEY_CHECKING`, and runs from `$HOME`).

Until **`ansible_password`** is in inventory or Vault, pass **`--ask-pass`** or you get **`No authentication methods available`**. **`ansible_user`** must match the IOS username **exactly** (SSH is case-sensitive), e.g. **`Iac`** vs **`iac`**.

**Roles:** `network-isr.yml` and `network-c3560cx.yml` use an **explicit role path** (`{{ playbook_dir }}/../roles/...`), so **`cisco_isr_platform` resolves even when `ansible.cfg` is ignored**.

**Inventory parse errors:** keep `inventory/network.yml` valid YAML — copy from `inventory/network.example.yml` and edit hosts only.

**Optional:** `./playbooks/run-network-isr.sh --check --ask-pass` exports `ANSIBLE_CONFIG` / `ANSIBLE_ROLES_PATH` and runs from `$HOME` so `ansible.cfg` is not skipped for world-writable CWD.

### SSH host keys (paramiko “authenticity can’t be established”)

**`network_cli` uses paramiko**, not OpenSSH — **`ansible_ssh_common_args: -o StrictHostKeyChecking=no` does not disable paramiko host-key checks.**

Use all of:

- Inventory: **`ansible_host_key_checking: false`** (see `inventory/network.example.yml`).
- `ansible.cfg`: **`host_key_checking = False`** and **`[paramiko_connection] host_key_auto_add = True`** (this repo includes both).

If your shell still loads a config that leaves checking on, export:

```bash
export ANSIBLE_HOST_KEY_CHECKING=false
export ANSIBLE_HOST_KEY_AUTO_ADD=true
```

**`playbooks/run-network-isr.sh`** sets **`ANSIBLE_CONFIG`**, runs from **`$HOME`** (so `ansible.cfg` isn’t skipped for world-writable CWD), and exports the env vars above.

Or SSH once manually: `ssh iac@<ISR_IP>` and accept the host key, then re-run Ansible.

### “Unable to connect to port 22”

That is **not** Ansible-specific: nothing is accepting TCP **22** on that IP from your controller (down host, wrong IP, routing/VLAN, Windows firewall, or SSH not enabled on the ISR `line vty` / `transport input ssh`). From the same machine run `nc -vz <IP> 22` or `ssh -v iac@<IP>` and fix L3 reachability / IOS SSH before re-running the playbook.

## Layout

| Path | Purpose |
|------|---------|
| `playbooks/network-isr.yml` | ISR role only |
| `playbooks/network-c3560cx.yml` | Access switches only |
| `playbooks/network-site.yml` | Both (ISR first) |
| `roles/cisco_isr_platform/` | Baseline hardening, SSH+local users (**`iac`** for Ansible + **`kbolsen`** key-only human), banner, logging sync, WAN+inside L3, **UCS-E IMC + `ucse`**, **VRRP**, optional OSPF/BGP lines, IGMPv3, NAT/static |
| `roles/cisco_c3560cx_access/` | `ios_vlans`, trunks, access ports |

## OpenNebula Gitea / Ceph VLANs (refined edge plan)

Align **C3560CX** trunks and **ISR** subinterfaces with [deployments/opennebula-kvm/VLAN_MATRIX.md](../deployments/opennebula-kvm/VLAN_MATRIX.md):

| VLAN | VNET | Segment |
|------|------|---------|
| 86 | `devsecops-edge` | 192.168.86.0/24 |
| 2001 | `devsecops-gitea` | 100.64.1.0/24 |
| 2005 | `devsecops-gateway` | 100.64.5.0/24 |
| 2250 | `devsecops-ceph` | 100.64.250.0/24 |

**`group_vars/network_c3560cx.yml.example`** uses `allowed_vlans: "240-244,2001-2054,2250,86"` so **2250** (outside the 2001–2054 tool range) is explicit. **ISR:** optional `cisco_isr_inside_subinterfaces` and `cisco_isr_static_routes` examples are in **`group_vars/network_isr.yml.example`**. Execution checklist: [docs/opennebula-gitea-edge/REFINED-EXECUTION.md](../docs/opennebula-gitea-edge/REFINED-EXECUTION.md).

## Design references

- **[CISCO_ISR_EDGE.md](CISCO_ISR_EDGE.md)** — DMVPN vs VXLAN/EVPN, OSPF to VyOS/FRR, Google Home (no IGP), native VLAN.
- **Inventory host_vars:** `inventory/host_vars/isr-primary.example.yml` (rename → `isr-primary.yml`) for hostname + **VRRP priority** pairs.

## Not automated yet

- **DMVPN / FlexVPN** tunnel templates (add dedicated task file when hub/spoke addressing is fixed)
- **Full BGP address-family** blocks beyond `router bgp …` line lists (extend or use `ios_config` with multi-parent if needed)
- **Leaf–spine** — future NX-OS / EVPN roles when hardware arrives (see VLAN_MATRIX **Roadmap**)

## Inventory

Use `ansible_connection: ansible.netcommon.network_cli` and `ansible_network_os: ios` (see `inventory/network.example.yml`).
