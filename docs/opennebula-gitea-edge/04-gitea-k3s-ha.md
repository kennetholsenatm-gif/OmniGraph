# Gitea on K3s — Reference Architecture (Repo-Aligned)



**Segments:** Gitea workloads mirror [docs/NETWORK_DESIGN.md](../NETWORK_DESIGN.md): **`gitea_net`** = **`100.64.1.0/24`** OpenNebula VNET **`devsecops-gitea`** (VLAN **2001**). Ingress / Traefik aligns with **`gateway_net`** = **`100.64.5.0/24`**, VNET **`devsecops-gateway`** (VLAN **2005**). Full matrix: [deployments/opennebula-kvm/VLAN_MATRIX.md](../../deployments/opennebula-kvm/VLAN_MATRIX.md).



## Stack overview



| Layer | Technology |

|-------|------------|

| Hypervisor / IaaS | OpenNebula + KVM on UCS-E |

| Kubernetes | K3s (3 VMs minimum: 1× server + 2× agents; expand to 3 servers for CP HA) |

| Ingress | Traefik (K3s default) or NGINX Ingress Controller — publish on **100.64.5.0/24** when following single-pane pattern |

| TLS | cert-manager + ACME (DNS-01 recommended behind corporate DNS) |

| Git HTTPS | Ingress → `gitea-http` service |

| Git SSH | Kubernetes `Service` `type: LoadBalancer` (MetalLB pool on **100.64.5.0/24** or **100.64.1.0/24** per policy) OR NodePort with ISR DNAT |

| Persistence | Ceph RBD via `StorageClass` (e.g. `ceph-rbd`); Ceph NICs on **100.64.250.0/24** (**devsecops-ceph**) |

| Database | PostgreSQL in-cluster (Bitnami subchart or separate release) or dedicated VM |



## Namespace layout



- `gitea-prod` — Gitea server, PVCs, secrets

- `gitea-db` — PostgreSQL (if split for blast radius)

- `ingress-nginx` or `kube-system` extensions as per installer



## VM networking (OpenNebula)



- Attach Gitea / K3s worker VMs to **`devsecops-gitea`** (**100.64.1.0/24**) for Git and app traffic consistent with compose **`gitea_net`**.

- Attach ingress controllers (or MetalLB speaker NICs) to **`devsecops-gateway`** (**100.64.5.0/24**) so VIPs match [NETWORK_DESIGN.md](../NETWORK_DESIGN.md) **Gateway (Traefik)** segment.

- Optional: second vNIC on workers for **`devsecops-ceph`** if host networking is required for CSI (often Ceph is reached routably from worker IPs on the storage VLAN).



## High availability (realistic for this footprint)



| Component | HA mechanism |

|-----------|----------------|

| Gitea pods | `replicas: 2+` with **ReadWriteOnce** PVC — **RWO cannot mount on two nodes simultaneously**. For true active-active app HA you need **shared RWX** storage (e.g. CephFS/NFS) **or** single replica with fast restore. |

| Practical choice | **1 replica** Gitea with **Ceph snapshots** + quick reschedule, **or** **2 replicas** only if using **ReadWriteMany** backed volumes. |

| PostgreSQL | Use **Bitnami PostgreSQL HA** (Patroni) or external managed DB for quorum-based failover. |

| Ingress | 2+ ingress controller replicas + stable VIP on **100.64.5.0/24** (MetalLB / kube-vip) |

| K3s control plane | Single server acceptable short-term; add 2 more server nodes when resources allow |



**Recommendation:** Start with **single Gitea replica + RWO on Ceph** and **HA PostgreSQL** (or robust external DB VM). Add CephFS later if you require multi-replica Gitea without downtime.



## After migration: Docsify and webhooks



If you use the repo’s **Docsify** + **single-pane** stack ([DOCSIFY_GITEA.md](../DOCSIFY_GITEA.md)) after moving Gitea to OpenNebula:



- **`DOCS_GIT_REPO`** must resolve Gitea using the **service hostname** reachable from `docs-sync` (e.g. `http://gitea:3000/...` on Docker; on OpenNebula use the **stable DNS name or IP** for Gitea on **100.64.1.0/24**).

- **Webhooks** to **`/webhook/docs-sync`** must preserve **`X-Gitea-Signature`** (HMAC-SHA256). Prefer **Gitea → gateway** direct; if relaying through n8n, see [n8n-workflows README](../../n8n-workflows/README.md) notes.

- Traefik routes **`/docs`** expect the gateway on **`gateway_net`** (**100.64.5.0/24**); keep that correlation when placing ingress VMs.



## Deployment order



1. Install K3s on VMs (fixed hostnames, NTP, firewall allowing pod CIDR).

2. Install **Ceph CSI** and verify `StorageClass`.

3. Install **cert-manager**.

4. Install **MetalLB** (or equivalent) with IP pool on **`100.64.5.0/24`** for ingress and Git SSH VIPs if that matches your ISR DNAT plan.

5. Install **PostgreSQL** (if not external).

6. Install **Gitea** (Helm) with values pointing to DB and storage class; **`ROOT_URL`** must match public DNS and TLS cert.

7. Configure SMTP, OAuth/LDAP, and integrations; re-validate **webhooks** per DOCSIFY_GITEA.



## Helm chart sources



- Official: `https://gitea-charts.gitea.io` (chart `gitea`)

- Ensure chart `appVersion` matches tested Gitea version for migration.



See example values in:



- [deployments/opennebula-gitea/helm/gitea-values.example.yaml](../../deployments/opennebula-gitea/helm/gitea-values.example.yaml)

- [deployments/opennebula-gitea/helm/postgresql-values.example.yaml](../../deployments/opennebula-gitea/helm/postgresql-values.example.yaml)



## Git over SSH options



| Option | Pros | Cons |

|--------|------|------|

| `LoadBalancer` + MetalLB VIP on **100.64.5.0/24** | Aligns with gateway segment; stable ISR DNAT target | Requires LB pool on that VNET |

| `LoadBalancer` + IP on **100.64.1.0/24** | Co-located with Gitea segment | ISR ACLs must allow WAN → that VIP |

| `NodePort` + static node pinning | Simple | Fragile |

| Separate small VM with `sshd` + git shell | Isolated | Extra hop / sync complexity |



## Observability



- Metrics: Prometheus scrape Gitea if enabled, plus ingress and DB metrics (telemetry segment **100.64.51.0/24** per matrix if mirrored).

- Logs: Loki or ELK; trace webhook delivery failures post-migration.



## Security hardening



- TLS 1.2+ only; HSTS on public ingress.

- Rotate join tokens and kubeconfigs; restrict `kubectl` to mgmt VPN.

- Store `INTERNAL_TOKEN`, `SECRET_KEY`, DB passwords in Kubernetes **Secrets** (or external vault).


