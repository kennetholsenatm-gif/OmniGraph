# cisco_isr_platform

Rudimentary **Cisco IOS-XE ISR** automation: inside **802.1Q** subinterfaces, **static routes**, optional **PAT** (disabled by default).

## Requirements

```bash
cd ansible && ansible-galaxy collection install -r collections/requirements.yml
```

## Variables

See `defaults/main.yml` and `group_vars/network_isr.yml.example`.

**VRRP/HSRP** is not templated here (pairs differ by priority/group) — add `ios_config` tasks or a second role after validating CLI in lab.

## Safety

- Run against lab ISR first.
- `cisco_isr_configure_nat_overload` defaults to **false**; define extended ACL `NAT_INSIDE` (or your name) on the device before enabling.
