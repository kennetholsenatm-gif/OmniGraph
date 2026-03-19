# Ansible: Cisco ISR & C3560-CX

Rudimentary automation for **IOS-XE ISR** and **Catalyst C3560-CX** switches, aligned with [deployments/opennebula-kvm/VLAN_MATRIX.md](../deployments/opennebula-kvm/VLAN_MATRIX.md).

## Quick start

```powershell
cd ansible
ansible-galaxy collection install -r collections/requirements.yml
copy inventory\network.example.yml inventory\network.yml
copy group_vars\network_isr.yml.example group_vars\network_isr.yml
copy group_vars\network_c3560cx.yml.example group_vars\network_c3560cx.yml
# Edit IPs, VLANs, credentials; use ansible-vault for secrets.
ansible-playbook -i inventory/network.yml playbooks/network-site.yml
```

## Layout

| Path | Purpose |
|------|---------|
| `playbooks/network-isr.yml` | ISR role only |
| `playbooks/network-c3560cx.yml` | Access switches only |
| `playbooks/network-site.yml` | Both (ISR first) |
| `roles/cisco_isr_platform/` | Inside dot1Q subifs, static routes, optional NAT |
| `roles/cisco_c3560cx_access/` | `ios_vlans`, trunks, access ports |

## Not automated yet

- **VRRP/HSRP** (add after you fix priorities in lab)
- **WAN DHCP** interface details (modem path)
- **Leaf–spine** — future NX-OS / EVPN roles when hardware arrives (see VLAN_MATRIX **Roadmap**)

## Inventory

Use `ansible_connection: ansible.netcommon.network_cli` and `ansible_network_os: ios` (see `inventory/network.example.yml`).
