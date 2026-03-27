# Virtual Router (VR) checklist — VLAN_MATRIX

Per [VLAN_MATRIX.md](../VLAN_MATRIX.md):

- On each **100.64._x_.0/24** workload VNET, reserve **100.64._x_.1** for the **VR** (default gateway for leases unless you use a different first-host convention—stay consistent).
- On **devsecops-edge** (**192.168.86.0/24**), typical reservation: **192.168.86.1** = ISR SVI, **192.168.86.2** = VR.
- **VR default route:** `0.0.0.0/0` → **192.168.86.1** (ISR).
- **ISR:** static route **100.64.0.0/10** (or per-prefix summaries) toward the **VR next-hop** on the segment where the VR attaches (often **192.168.86.2**).
- **Do not** enable NAT on the VR for Internet-bound traffic if the ISR already performs **single PAT** toward the WAN (matrix policy).

Create the VR object in OpenNebula and attach NICs to **`devsecops-edge`**, **`devsecops-gitea`**, **`devsecops-gateway`**, **`devsecops-ceph`** as required by your east-west routing design.
