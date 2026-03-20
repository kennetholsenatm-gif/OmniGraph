# cisco_isr_platform

**IOS‑XE ISR** (e.g. **ISR4351** with **3× GE** front panel: `GigabitEthernet0/0/0` … `/2`) — baseline **security**, **SSH + local user**, **banner**, **logging sync**, **inside 802.1Q**, **VRRP**, optional **multicast/IGMPv3**, optional **OSPF/BGP** lines, **NAT/static** helpers.

## Design / routing

See **[docs/CISCO_ISR_EDGE.md](../../../docs/CISCO_ISR_EDGE.md)** for **DMVPN vs VXLAN**, **OSPF vs BGP to VyOS**, and **why Google Home doesn’t run an IGP.**

## Requirements

```bash
cd ansible
ansible-galaxy collection install -r collections/requirements.yml
# Controller SSH (network_cli): same Python as ansible-playbook needs paramiko or ansible-pylibssh
python3 -m pip install -r requirements-controller.txt
```

## SSH and local users (`iac` + `kbolsen`)

- **Ansible SSH user:** inventory **`ansible_user`** should be **`iac`** (same as **`cisco_isr_admin_username`**). Do **not** use **`kbolsen`** as `ansible_user` — `kbolsen` is a separate human admin account on the box.
- **Password (IOS user `iac`):** set **`cisco_isr_admin_password`** in `group_vars/network_isr.yml` (copy from `group_vars/network_isr.yml.example`). HashiCorp Vault is **not** required for lab; use Ansible Vault encrypt later if you want.
- **SSH from Ansible to the ISR:** set **`ansible_password`** in inventory `vars` or use **`ansible-playbook … --ask-pass`** (same password is fine for bootstrap).
- **Automation public key (`iac`):** add **one line** to **`files/ssh/iac.pub`** (copy from `files/ssh/iac.pub.example`; **`iac.pub`** is **gitignored**), **or** set **`cisco_isr_admin_ssh_pubkey`** / **`cisco_isr_admin_ssh_pubkey_file: "ssh/iac.pub"`**.
- **Human admin (`kbolsen`):** configured as **`nopassword` + SSH pubkey** (key-only interactive login). Default **`files/ssh/kbolsen.pub`** is tracked in-repo (same material as **`.dev/kbolsen_admin.pub`**). Override with **`cisco_isr_human_admin_ssh_pubkey`** or **`cisco_isr_human_admin_ssh_pubkey_file`**. Disable with **`cisco_isr_human_admin_enabled: false`**.

The role generates an RSA **host** key when `cisco_isr_generate_ssh_rsa_key` is true (best‑effort; lab first).

## Hostname scheme

| Variable | Example | Result |
|----------|---------|--------|
| `cisco_isr_hostname` | `isr01` | IOS `hostname isr01` |
| `cisco_isr_domain_name` | `edge.lab` | `ip domain name edge.lab` |
| (empty hostname) | — | Uses `inventory_hostname` (sanitized, max 63 chars) |

Use **`inventory/host_vars/isr-primary.yml`** (copy from `*.example.yml`) for per‑appliance names and **VRRP priorities** when your inventory file lives in `inventory/`.

## UCS-E Service Module (IMC + `ucse` link)

For **ISR4k** blades (**UCS-E140S**, etc.) in **SM** slots **1/0**, **2/0**, … enable **`cisco_isr_ucse_enabled`** and define **`cisco_isr_ucse_modules`** (see `defaults/main.yml`). This applies:

- **`interface ucse<s>/<ss>/0`** — L3 (default) or **802.1Q trunk** into the module.
- **`ucse subslot <s>/<ss>`** with **`imc ip address … default-gateway …`** — **IMC** (CIMC-style management IP).

Default **`imc_access_port`**: `shared-lom ge2` (override per Cisco UCS-E doc if you use **ge3** or **console** path).

Addresses should follow **`deployments/opennebula-kvm/VLAN_MATRIX.md`** (e.g. **100.64.245.0/24** UCS-A, **100.64.246.0/24** UCS-B): reserve **.1** for the ISR on `ucse`, **.10** for IMC (example), hypervisors **.20+**.

## VRRP

Put **`vrrp`** under each inside subinterface dict:

```yaml
cisco_isr_inside_subinterfaces:
  - vlan: 240
    ipv4: 100.64.240.2
    mask: 255.255.255.0
    description: PLATFORM_CORE
    vrrp:
      group: 240
      vip: 100.64.240.1
      priority: 110
      preempt: true
```

Secondary ISR: same VIP/group, **lower priority** (e.g. `90`).

## Native VLAN ≠ 1

On the **Catalyst** trunk toward the ISR, set **`switchport trunk native vlan 3999`** (or another unused ID). The ISR defaults to **L3 subinterfaces**; native VLAN is primarily a **switch** concern unless you enable `cisco_isr_inside_switchport_trunk`.

## UCS-E — keep data interfaces out of `shutdown`

When **`cisco_isr_ucse_enabled: true`**, the role issues **`no shutdown`** on **`interface ucse<s>/<ss>/0`** for every subslot listed in **`cisco_isr_ucse_modules`** plus optional **`cisco_isr_ucse_extra_subslots`** (e.g. `["1/0","2/0"]`) so blades aren’t left administratively down before L3/IMC is fully modeled. Toggle with **`cisco_isr_ucse_ensure_no_shutdown`**.

## Control-plane & data-plane security (optional)

- **Data plane (WAN ingress ACL):** dict **`cisco_isr_wan_infrastructure_acl`** — set **`enabled: true`** only after you validate **`lines`** for your WAN (**DHCP** vs static) and **transit** traffic. Applied as **`ip access-group <name> in`** on **`cisco_isr_wan_interface`**.
- **Control plane (CoPP-style):** **`cisco_isr_copp_enabled: true`** pushes a **`policy-map type control-plane`** with **`police rate percent`** on **`class class-default`** and attaches it under **`control-plane host`**. Disable if your IOS-XE build rejects the syntax.
- **Management to CPU:** combine with **`cisco_isr_vty_access_class`** (standard ACL name) on VTY lines to restrict **SSH sources**.

## Variables (summary)

| Area | Variables |
|------|-----------|
| Identity | `cisco_isr_hostname`, `cisco_isr_domain_name` |
| Admin | `cisco_isr_admin_username` (default `iac`), `cisco_isr_admin_password`, `cisco_isr_admin_ssh_pubkey` / `…_pubkey_file` |
| Human admin | `cisco_isr_human_admin_username` (default `kbolsen`), `cisco_isr_human_admin_ssh_pubkey` / `…_pubkey_file`, `nopassword` key-only |
| WAN / inside | `cisco_isr_wan_interface`, `cisco_isr_inside_interface`, `cisco_isr_inside_subinterfaces`, optional `cisco_isr_loopback_config` |
| UCS-E | `cisco_isr_ucse_enabled`, `cisco_isr_ucse_modules`, `cisco_isr_ucse_extra_subslots`, `cisco_isr_ucse_ensure_no_shutdown` |
| Data-plane ACL | `cisco_isr_wan_infrastructure_acl.{enabled,name,lines}`, `cisco_isr_wan_infrastructure_acl_lines_template` |
| Control-plane | `cisco_isr_copp_enabled`, `cisco_isr_copp_policy_map`, `cisco_isr_copp_police_rate_percent` |
| Banner / lines | `cisco_isr_banner_login`, `cisco_isr_line_*`, `cisco_isr_vty_access_class` |
| Logging | `cisco_isr_logging_buffered_size`, `cisco_isr_logging_console`, `cisco_isr_logging_sync_level` |
| Multicast | `cisco_isr_multicast_routing`, `cisco_isr_igmp_interfaces` |
| IGP / BGP | `cisco_isr_ospf.{enabled,process_id,lines}`, `cisco_isr_bgp.{enabled,as,lines}` |
| NAT | `cisco_isr_configure_nat_overload`, ACL name vars (define ACL on box before enabling) |

Full defaults: **`defaults/main.yml`**.

## Safety

- Run **lab ISR** first; review diffs (`ansible-playbook … --check` where supported).
- NAT remains **off** until you place the extended ACL referenced by `cisco_isr_nat_acl_name`.
- **WAN ingress ACL** can cut **management**, **DHCP**, or **transit** if `lines` are wrong — enable **`cisco_isr_wan_infrastructure_acl.enabled`** only with a reviewed ACL.

## Playbook

```bash
ansible-playbook -i inventory/network.yml playbooks/network-isr.yml
```
