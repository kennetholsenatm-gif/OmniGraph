# Mini PC edge: VyOS + PacketFence + RatTrap (FOSS firewall / NAC)

This document is the **architecture source of truth** for a dedicated **mini PC** acting as the **L3 edge** in front of the home/lab network: **VyOS** (Incus system container), **PacketFence** (Incus **VM**), and **RatTrap** (transparent filtering appliance) on **dedicated VLANs**. It **aligns** with [deployments/opennebula-kvm/VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md) and [docs/NETWORK_DESIGN.md](../NETWORK_DESIGN.md) — **do not** place RatTrap transit or VyOS↔RatTrap **/30** links inside **`100.64.0.0/10`** workload space unless you intentionally merge routing domains.

## Roles

| Component | Deployment | Role |
|-----------|------------|------|
| **VyOS** | Incus **LXC** (privileged as needed for routing/tun) | Default gateway for **selected** LAN VLANs on **eth1** trunk; **WAN** on **eth0** (ISP DHCP); **PBR** to send only chosen traffic through RatTrap; **OSPF** with downstream **ISR**; **VxLAN VTEP** / tunnel toward **`100.64.0.0/10`** services |
| **PacketFence** | Incus **VM** | 802.1X / MAB, RADIUS, captive portal; **not** LXC (DB + kernel/iptables expectations) |
| **RatTrap** | Physical appliance | **Hairpin**: “inside” on **VLAN 900**, “outside” on **VLAN 901** — **no** L3 SVI on the switch if VyOS owns both VIFs |
| **Alma host** | Bare metal on mini PC | **Incus**; **br-lan** (example name) bridges **VyOS LAN** NIC and **PacketFence** NICs |

## Physical mapping (WS-C3560CX-1)

| Port role | Switch | VLAN | Notes |
|-----------|--------|------|--------|
| Cable modem | CX-1 access | **99** (example WAN VLAN) | ISP handoff |
| Mini PC **eth0** | CX-1 access | **99** | Same VLAN as modem — OOB capture, no direct modem–NIC cable required |
| Mini PC **eth1** | CX-1 **802.1Q trunk** | Allowed: internal VLANs + **900** + **901** + NAC + **transit** (see matrix) | VyOS **router-on-a-stick** + RatTrap pair |
| RatTrap “local” | CX-1 access | **900** | Clean side toward appliance |
| RatTrap “internet” | CX-1 access | **901** | Dirty side toward VyOS **VIF 901** |

**Trunk to OpenNebula / ISR path:** CX-1 ↔ **WS-C3560CX-2** carries only the VLANs needed for compute, storage, and **VyOS↔ISR transit** (see [VLAN_MATRIX](../../deployments/opennebula-kvm/VLAN_MATRIX.md) edge table).

## Logical VyOS (RatTrap hairpin)

- **eth0** (in VyOS LXC): **WAN** — DHCP client toward ISP on **VLAN 99**.
- **eth1** trunk subinterfaces:
  - **vif 900** — transit toward RatTrap **inside** (e.g. `10.99.0.1/30`).
  - **vif 901** — transit from RatTrap **outside** (e.g. `10.99.0.5/30`).
- **RatTrap** addressing is **out-of-band** from **`100.64.0.0/10`** — use a dedicated **RFC1918 /30** (example **`10.99.0.0/30`**) documented in your IPAM; do **not** overlap [platform carve](../../deployments/opennebula-kvm/VLAN_MATRIX.md) **`100.64.240–254`**.

## Policy-based routing (PBR)

- **Default:** Internet-bound traffic uses the **main** table → **WAN** (bypass RatTrap).
- **Policy:** For sources in **IOT_SUBNET** (or other groups), install a **custom table** (e.g. **100**) whose default route points to RatTrap **inside** on **VIF 900**.
- **Ingress:** Apply policy on **LAN SVI / vif** where IoT or Google Home VLANs attach.

**Traffic path (policy hit):** Client → VyOS gateway → **PBR** → **VIF 900** → RatTrap → **VIF 901** → VyOS → **NAT** → **WAN**.

**Ops bypass:** Disable or reorder the PBR rule if RatTrap is offline (documented runbook).

## Gateway, NAT, and OSPF (single decision)

**Adopted model for this refactor (document consistently in playbooks):**

1. **Single PAT toward the cable modem** lives on **VyOS WAN** (mini PC **eth0** / VLAN **99**). The ISR **does not** terminate ISP DHCP for the same public path when VyOS is in production as edge.
2. **ISR** remains the **L3 hub** for **OpenNebula UCS** segments and **`100.64.240–254`** platform carve per the matrix, but **northbound** Internet for those flows goes **via static/OSPF route to VyOS** (or **east–west** only inside `100.64` behind **VR** as today).
3. **OSPF:** **VyOS** and **ISR** form an adjacency on a **dedicated transit VLAN** (see matrix — **not** 900/901). **Do not** run OSPF on RatTrap VLANs.

**Google Home / `100.64.244.0/24`:**

- **Recommended (minimal churn):** Keep **Google Home** SVI and **NAT context** on the **ISR** as in VLAN_MATRIX; use VyOS **PBR + RatTrap** for **other** IoT or lab VLANs only.
- **Alternative (future):** Move the Google Home SVI to a **VyOS vif**, **redistribute static** into **OSPF Area 1**, and **remove** duplicate NAT — requires a **cutover** window and ISR route cleanup.

## Integrations to the DevSecOps `100.64.x` plane

Targets are the same segments as [NETWORK_DESIGN.md](../NETWORK_DESIGN.md):

| Integration | Segment | Doc |
|-------------|---------|-----|
| **FreeIPA / LDAP** | `100.64.21.0/24` (`freeipa_net`) | [NETWORK_DESIGN](../NETWORK_DESIGN.md) |
| **sFlow-RT / telemetry** | `100.64.51.0/24` (`telemetry_net`) | [SDN_TELEMETRY.md](../SDN_TELEMETRY.md) |
| **Wazuh / SIEM** | `100.64.54.0/24` (`siem_net`) | [WAZUH_SIEM.md](../WAZUH_SIEM.md) |
| **n8n automation** | `100.64.2.0/24` (`n8n_net`) | [NETWORK_DESIGN](../NETWORK_DESIGN.md) |
| **SDN lab** | `100.64.50.0/24` (`sdn_lab_net`) | [SDN_TELEMETRY.md](../SDN_TELEMETRY.md) |

**VyOS** exports **sFlow** toward **sFlow-RT**; **syslog** forward to **Wazuh**; **API** or **automation hooks** align with **n8n** and **PacketFence** webhooks as in your operator runbooks.

## Cisco ISR (downstream)

- Physical and routing relationship to **VyOS** and **OpenNebula** is summarized in [deployments/opennebula-kvm/VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md).
- ISR-focused Ansible and design notes: [docs/CISCO_ISR_EDGE.md](../CISCO_ISR_EDGE.md), [ansible/playbooks/network-isr.yml](../../ansible/playbooks/network-isr.yml).

## Incus notes

- **VyOS LXC** may require **nesting / privileged** options for **TUN/TAP**, **iptables/nft**, and **network namespaces** — align with [ansible/roles/lxd_devsecops_stack](../../ansible/roles/lxd_devsecops_stack) patterns (`docker-nesting` profile, `raw.lxc` where documented).
- **PacketFence VM**: allocate **vCPU/RAM/disk** per vendor guidance; attach NICs to the same **host bridge** as VyOS **LAN** leg.

## Automation (phased)

- **Switches:** Extend [ansible/playbooks/network-c3560cx.yml](../../ansible/playbooks/network-c3560cx.yml) / inventory for **VLAN 99, 900, 901**, trunk **allowed VLAN** lists — **after** VLAN IDs are frozen in VLAN_MATRIX.
- **VyOS / PacketFence:** Version-pinned **config** or **Ansible** roles are **out of scope** until CLI/API contracts are frozen; track in repo issues.

## Packer + Semaphore + Ansible (mini PC host image)

Use the repo **golden image** and **host** playbooks to prepare the Alma **bare metal** before you define VyOS/PacketFence instances:

| Step | Doc / asset |
|------|-------------|
| **Build QCOW2** (Linux + KVM) | [deployments/mini-pc-firewall/README.md](../../deployments/mini-pc-firewall/README.md) — `packer init` / `packer build` in `deployments/mini-pc-firewall/packer/` |
| **Automate Packer** | [ansible/playbooks/packer-build-mini-pc-incus-host.yml](../../ansible/playbooks/packer-build-mini-pc-incus-host.yml) |
| **Configure live host** (Incus, sysctl, COPR) | [ansible/playbooks/mini-pc-firewall-host.yml](../../ansible/playbooks/mini-pc-firewall-host.yml) + [ansible/inventory/mini-pc-firewall.example.yml](../../ansible/inventory/mini-pc-firewall.example.yml) |
| **Semaphore templates** | [deployments/mini-pc-firewall/semaphore/TEMPLATE-EXAMPLE.md](../../deployments/mini-pc-firewall/semaphore/TEMPLATE-EXAMPLE.md) |
| **Local Semaphore (Incus)** | [LEAN_LOCAL_CONTROL_PLANE.md](LEAN_LOCAL_CONTROL_PLANE.md), `scripts/start-semaphore.sh` |

**Note:** The Packer template installs **router host** prerequisites (forwarding, bridges, firewall packages). **Incus** is installed by **Ansible** (`mini_pc_incus_host`) so COPR/repo URLs can be validated per EL release.

## References

- [VLAN_MATRIX](../../deployments/opennebula-kvm/VLAN_MATRIX.md) — workload `2000+` rule + **physical edge** table
- [NETWORK_DESIGN.md](../NETWORK_DESIGN.md) — `100.64.x` ↔ `*_net`
- [REDUCE-DOCKER.md](REDUCE-DOCKER.md) — native vs compose
- [LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md) — LXC on OpenNebula
