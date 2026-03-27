# freeradius_alma

FOSS **FreeRADIUS** on AlmaLinux/RHEL-family (`radiusd`): installs packages, templated `clients.d` for NAS secrets, optional **LDAP** module stub for FreeIPA/AD follow-up.

**Does not** replace Keycloak; use for **802.1X / VPN / network device** AAA alongside IAM. See [docs/IAM_LDAP_AND_AUTOMATION.md](../../../docs/IAM_LDAP_AND_AUTOMATION.md) for LDAP directory alignment.

## Variables

- `freeradius_clients` — list of `{ name, ipaddr | cidr, secret }` (vault the secrets).
- `freeradius_ldap_enabled` — install `freeradius-ldap` and drop `mods-available/ldap.devsecops` (wire into `sites-enabled` per your policy).

## Usage

```bash
cd ansible
ansible-playbook -i inventory/opennebula-hybrid.yml playbooks/deploy-freeradius-native.yml -e @group_vars/freeradius/vault.yml --limit devsecops-radius-01
```

Place FreeRADIUS on a dedicated LXC/VM (e.g. `devsecops-radius`) with a stable `100.64.x` address from [docs/NETWORK_DESIGN.md](../../../docs/NETWORK_DESIGN.md).
