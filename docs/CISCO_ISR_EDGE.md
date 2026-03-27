# Cisco ISR edge — routing overlays (design notes)

## What you asked vs what maps where

| Concept | What it is | On your ISR4351 lab |
|---------|------------|---------------------|
| **VXLAN** | L2 overlay with UDP encapsulation; control plane often **BGP EVPN** (type-2/3/5 routes) in DC fabrics | Not ISR bread-and-butter; ISR is good at **WAN + services + mVPN**, not full EVPN leaf. |
| **DMVPN** | **mGRE + NHRP** (dynamic spoke resolution) + **IPsec**; hub/spoke or phase-3 partial mesh | Valid ISR strength for many WANs. **Not VXLAN.** |
| **BGP + VXLAN** | Colloquial short‑hand for **EVPN** (MP‑BGP carrying EVPN NLRI) | If you later run **VyOS/FFR as a Linux VTEP**, BGP EVPN there is normal; the ISR would typically be **eBGP handoff / default route** into that island, not the EVPN RR. |
| **OSPF / ISIS** | Classic **IGP** between Cisco and **VyOS/FRR** | **OSPFv2** is the lowest‑friction IGP between IOS‑XE and VyOS/FRR (`ospfd`/`zebra`). Use one area (0) to start; add `passive-interface default` and un‑pass only transit VLANs. |
| **Google Home / Nest Wi‑Fi** | Consumer gateway | **No IGP.** Treat as **L3 hop**: RFC1918 or carve `100.64.243.0/24` **/30** toward it (see `deployments/opennebula-kvm/VLAN_MATRIX.md`), **static default or statics** from ISR. DHCP from the mesh for Wi‑Fi clients stays on Google. |

## Recommended lab stack (this repo’s direction)

1. **Underlay between ISR and VyOS:** **OSPF area 0** on the segments you want automated (e.g. `100.64.240.0/24`, `100.64.244.0/24`, future `100.64.50.0/24` SDN leg).  
2. **Google segment:** **No OSPF** — static summary or host routes from ISR; **NAT policy** from docs (`VLAN_MATRIX` / `cisco_isr_platform` NAT block when you enable PAT).  
3. **DMVPN:** add only if you need dynamic spoke VPN — separate playbook/role chunk (`tunnel`, `nhrp`, IKEv2/IPsec profile); don’t conflate with VXLAN.  
4. **VXLAN/EVPN:** park on **OpenNebula + Linux/VyOS VTEPs** first; ISR stays **northbound** (WAN + firewall + NAT + BGP/OSPF peering), not the VXLAN control plane.

## Automation

- Role: [`ansible/roles/cisco_isr_platform/`](../ansible/roles/cisco_isr_platform/README.md)  
- Inventory: `network_isr` (see `ansible/inventory/network.example.yml`)  
- Hostname scheme: `cisco_isr_hostname` + `cisco_isr_domain_name` → operator FQDN convention (e.g. `isr01.edge.lab`).

## Native VLAN

**Do not use VLAN 1 as the native VLAN** on the Catalyst trunk toward the ISR. Use an **unused** id (e.g. **3999**) as native on the switch; ISR keeps **802.1Q subinterfaces** for real VLANs (`240`, `244`, `86`, …). If you ever place the ISR port in **switchport trunk** mode (unusual on 4351 GE), the role can set `switchport trunk native vlan {{ cisco_isr_inside_native_vlan_id }}`.

## IGMPv3

`ip igmp version 3` is applied per-interface via `cisco_isr_igmp_interfaces` after `ip multicast-routing distributed` (toggle `cisco_isr_multicast_routing`). Design **RP/BSR** before turning on **PIM sparse** everywhere.

## UCS-E (ISR4k service module) — **IMC** + internal `ucse` link

Your **ISR4351** may host **UCS-E140S-M2** (or similar) in **SM subslot `1/0`** and **`2/0`** (see `show inventory`).

| Function | IOS-XE pattern | Example (VLAN_MATRIX) |
|----------|----------------|------------------------|
| **IMC** (web UI to manage the blade) | `ucse subslot 1/0` → `imc ip address … default-gateway …` | IMC **100.64.245.10/24**, gateway **100.64.245.1** |
| **Router ↔ blade** (traffic through IOS) | `interface ucse1/0/0` L3 or trunk | L3 **100.64.245.1/24** on `ucse1/0/0`; second module **100.64.246.1/24** on `ucse2/0/0` |

**“ICM”** in operator shorthand is the same family as **IMC / CIMC** configuration in the UCS-E Getting Started guide: use **`imc access-port`** (`shared-lom ge2`, `shared-lom ge3`, `shared-lom console`, …) to match how you physically share the management port.

Automation: set `cisco_isr_ucse_enabled: true` and `cisco_isr_ucse_modules` in `group_vars/network_isr.yml` (see `network_isr.yml.example`). Hypervisor and VM addressing stay in **100.64.245.0/24** and **100.64.246.0/24** per your OpenNebula plan; avoid duplicating the same /24 on a **Gi** subinterface and **`ucse`** unless you explicitly design HA.
