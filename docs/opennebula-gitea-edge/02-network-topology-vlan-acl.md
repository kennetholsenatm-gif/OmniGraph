# Network Topology, VLANs, and ACL Strategy (Repo-Aligned)



**Single source of truth:** [deployments/opennebula-kvm/VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md) and [docs/NETWORK_DESIGN.md](../NETWORK_DESIGN.md).



This document replaces generic `10.10.x.x` examples with the pipeline’s **`100.64.x.0/24`** OpenNebula VNETs and **VLAN `2000 + third octet`** rule. For a disconnected lab-only design, you may use other RFC1918 space; label it explicitly as non-canonical.



## Design goals



- Match **segmentation** already defined for Docker tooling: Gitea **`100.64.1.0/24`**, gateway (Traefik) **`100.64.5.0/24`**, etc.

- Use **802.1Q trunks** from ISRs and hypervisors to **WS-C3560CX** (and/or Catalyst 3650 per site diagram in the matrix).

- **Virtual Router (VR):** on each workload `100.64.x.0/24`, reserve **`.1`** for the VR default gateway; VR default route → **`192.168.86.1`** (ISR on **VLAN 86**). **Do not** NAT on the VR for Internet if the ISR already performs **single PAT** on the WAN—see matrix.

- **ISR:** static route(s) for **`100.64.0.0/10`** (or tighter prefixes) toward the VR next-hop on the segment where the VR is homed; **single NAT** on WAN per matrix.

- Avoid L2 loops: validate STP root, BPDU guard on access, documented EtherChannel only where dual-linked.

- Apply **default-deny** between zones; align lateral rules with **NETWORK_DESIGN.md** (e.g. no direct Gitea ↔ n8n ↔ Zammad bridges).



## Matrix-aligned segments (Gitea migration focus)



| VLAN | OpenNebula VNET `NAME` | IPv4 segment | Use for this migration |

|------|------------------------|----------------|-------------------------|

| **2001** | `devsecops-gitea` | **100.64.1.0/24** | Gitea VM / K3s workloads hosting Gitea (same as `gitea_net`) |

| **2005** | `devsecops-gateway` | **100.64.5.0/24** | Traefik / single-pane ingress VIPs; path routes like `/docs` |

| **2250** | `devsecops-ceph` | **100.64.250.0/24** | Ceph public (and optionally cluster) front-end—see matrix row **devsecops-ceph** |

| **86** | `devsecops-edge` | **192.168.86.0/24** | Edge / home LAN; ISR SVI; trunks to hypervisors |



Other VNETs (n8n, IAM, messaging, …) remain as in the full matrix; attach K3s worker NICs only if you intentionally mirror the full compose topology on VMs.



## OpenNebula `*.one` snippets (from matrix pattern)



Set **`PHYDEV`** to your trunk NIC or bond on each KVM host.



### Gitea VNET



```text

NAME   = "devsecops-gitea"

DESCRIPTION = "Gitea — repo NETWORK_DESIGN segment 100.64.1.0/24"

VN_MAD = "bridge"

BRIDGE = "onebr-devsecops-gitea"

PHYDEV = "eth0"

VLAN_ID = "2001"

AR = [ TYPE = "IP4", IP = "100.64.1.2", SIZE = 100, LEASES = "YES" ]

```



Reserve **100.64.1.1** for the **Virtual Router** on this segment (see matrix “Virtual Router”). Adjust `IP` / `SIZE` for your lease pool.



### Gateway / Traefik VNET



```text

NAME   = "devsecops-gateway"

DESCRIPTION = "Traefik / single pane — 100.64.5.0/24"

VN_MAD = "bridge"

BRIDGE = "onebr-devsecops-gateway"

PHYDEV = "eth0"

VLAN_ID = "2005"

AR = [ TYPE = "IP4", IP = "100.64.5.2", SIZE = 100, LEASES = "YES" ]

```



### Ceph storage VNET (if using dedicated segment)



```text

NAME   = "devsecops-ceph"

DESCRIPTION = "Ceph MON/OSD client/public — 100.64.250.0/24"

VN_MAD = "bridge"

BRIDGE = "onebr-devsecops-ceph"

PHYDEV = "eth0"

VLAN_ID = "2250"

AR = [ TYPE = "IP4", IP = "100.64.250.2", SIZE = 100, LEASES = "YES" ]

```



### Edge VNET (excerpt from matrix)



```text

NAME   = "devsecops-edge"

DESCRIPTION = "Edge compute / home LAN 192.168.86.0/24 — ISR SVI VLAN 86"

VN_MAD = "bridge"

BRIDGE = "onebr-devsecops-edge"

PHYDEV = "eth0"

VLAN_ID = "86"

AR = [ TYPE = "IP4", IP = "192.168.86.3", SIZE = 100, LEASES = "YES" ]

```



*(Matrix reserves **192.168.86.1** ISR, **192.168.86.2** VR example.)*



## Virtual Router and ISR (summary)



Per [VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md):



- On each **100.64._x_.0/24** VNET, reserve **100.64._x_.1** for the VR.

- VR **default route:** `0.0.0.0/0` → **192.168.86.1** (ISR).

- ISR **static routes** toward **100.64.0.0/10** via the VR next-hop on the attachment segment (e.g. **192.168.86.2** on VLAN 86 when VR is homed there).



## Switching (WS-C3560CX / Catalyst)



1. Create VLANs **2001, 2005, 2250, 86** (and other matrix IDs used on the host).

2. Trunk to ISR and UCS-E / Mini-PC hypervisors with **allowed VLAN** list matching the matrix (least privilege).

3. STP: consistent **RPVST+** or **MST**; root primary/secondary documented; **PortFast + BPDU Guard** on host-access ports only.



### Example trunk (conceptual)



```text

interface GigabitEthernet1/0/1

 description TRUNK-to-ISR-A

 switchport trunk encapsulation dot1q

 switchport mode trunk

 switchport trunk allowed vlan 86,2001,2005,2250

```



## KVM host bridging



Map VLANs to Linux bridges consumed by OpenNebula (names align with `BRIDGE = onebr-*` above):



| Bridge (example) | VLAN | Segment |

|------------------|------|---------|

| `onebr-devsecops-edge` | 86 | 192.168.86.0/24 |

| `onebr-devsecops-gitea` | 2001 | 100.64.1.0/24 |

| `onebr-devsecops-gateway` | 2005 | 100.64.5.0/24 |

| `onebr-devsecops-ceph` | 2250 | 100.64.250.0/24 |



### Illustrative Linux VLAN → bridge



```bash

# Trunk on eno1 — example VLANs only

ip link set eno1 up

for vid in 86 2001 2005 2250; do

  ip link add link eno1 name eno1.${vid} type vlan id ${vid}

  ip link set eno1.${vid} up

done

# Attach each to a bridge matching OpenNebula BRIDGE names (onebr-devsecops-*)

```



Persist with **Netplan** or **NetworkManager** on each hypervisor.



## ACL matrix (high level, repo segments)



Align with [NETWORK_DESIGN.md](../NETWORK_DESIGN.md) **Isolation Rules**: tooling segments should not reach each other directly unless a documented exception (e.g. gateway joins multiple segments for Traefik).



| Flow | Action | Notes |

|------|--------|------|

| WAN → **ingress VIP** on **100.64.5.0/24** | PERMIT **443** (HTTPS), optional **SSH** if Git SSH terminates on LB | Match Traefik / MetalLB pool on gateway VNET |

| WAN → Gitea **100.64.1.0/24** | Usually **DENY** direct; prefer **publish via gateway** | Exceptions: dedicated DMZ IP for Git SSH |

| **100.64.250.0/24** (Ceph) ↔ Internet | **DENY** | Storage stays internal |

| **100.64.250.0/24** ↔ Ceph nodes only | **PERMIT** MON/OSD ports | Split public/cluster networks if required |

| Admin → **192.168.86.0/24** / hypervisor mgmt | **PERMIT** SSH / Sunstone | Via jump or VPN |

| **100.64.1.0/24** → **100.64.2.0/24** (n8n) | **DENY** default | Same lateral discipline as Docker tooling |



### Example ISR extended ACL fragment (illustrative)



Use real VIPs from MetalLB or OpenNebula on **`100.64.5.0/24`**:



```cisco

ip access-list extended FROM-WAN-TO-GATEWAY

 permit tcp any host 100.64.5.20 eq 443

 permit tcp any host 100.64.5.21 eq 2222

 deny ip any any log

```



Replace hosts with your **Traefik** and **Git SSH** VIPs.



## DNS and certificates



- Internal: corp DNS for **`git.<domain>`** resolving to **ingress** on **100.64.5.0/24** (or public DMZ if NAT’d).

- **cert-manager** on K3s: SANs must match `ROOT_URL` and any **webhook** URLs documented in [DOCSIFY_GITEA.md](../DOCSIFY_GITEA.md).



## Validation checklist



- [ ] Trunk **allowed VLAN** lists match [VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md) for this site.

- [ ] VR **.1** and default route; ISR **100.64.0.0/10** route and **single PAT** policy per matrix.

- [ ] From a test VM on **100.64.1.0/24**, reach gateway VIP on **100.64.5.0/24** only on allowed ports.

- [ ] Ceph nodes on **100.64.250.0/24** isolated from WAN; OSD peering OK.

- [ ] Gitea HTTP(S) + SSH clone paths validated end-to-end after cutover.



## Non-repo / lab-only note



A parallel scheme using **`10.10.x.x`** VLANs **10, 20, 30, 40** is **not** the canonical pipeline design; if used temporarily, document divergence from **VLAN_MATRIX.md** to avoid segment collisions when merging with compose-based **100.64** tooling.


