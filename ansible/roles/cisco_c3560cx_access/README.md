# cisco_c3560cx_access

Rudimentary **Catalyst C3560-CX** (or other **IOS L2/L3** switches): **VLANs**, **802.1Q trunks**, **access** ports.

## Requirements

```bash
ansible-galaxy collection install -r collections/requirements.yml
```

## Variables

See `defaults/main.yml` and `group_vars/network_c3560cx.yml.example`.

Set `allowed_vlans` on each trunk (e.g. `240-244,2001-2054,86`) to match [VLAN_MATRIX.md](../../../deployments/opennebula-kvm/VLAN_MATRIX.md).

## Notes

- Some images use `switchport trunk allowed vlan` without `add` first — adjust if the switch rejects the line.
- For **layer-3 SVI** on the 3560CX, extend this role or add `ios_config` tasks (not included by default).
