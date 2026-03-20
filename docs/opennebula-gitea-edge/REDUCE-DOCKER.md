# Reduce Docker (OpenNebula / AlmaLinux 10 LXC)

**Goal:** Use **as little Docker as possible**. Treat **AlmaLinux 10 LXC** as the main unit of isolation; run **systemd-managed native packages or upstream binaries** inside each LXC. Reserve **containers only** where vendors ship no supported bare-metal path or migration cost is unacceptable.

This repo’s **default** automation still centers on **Docker Compose** for reproducibility. **OpenNebula production** should converge on the tiers below; compose trees remain useful as **reference** for versions, env vars, and ports until Ansible/systemd roles replace them.

## Decision tiers

| Tier | Pattern | When |
|------|---------|------|
| **A — Native** | `dnf install` + **systemd units**; data under `/var/lib/...` | Service has official Alma/RHEL packages or a single static binary (Gitea, Vault, Traefik, Postgres) |
| **B — Podman** | **Rootless or rootful Podman**; `podman compose` / Quadlet / Kubernetes YAML | You want OCI images **without** the Docker daemon; better fit for LXC than DinD |
| **C — Docker (legacy)** | Docker CE + `docker compose` inside LXC | **Transitional** only; shrink scope over time |
| **D — Must stay OCI** | Vendor appliance image (e.g. some brokers, Confluent-style stacks) | Document exception; consider **separate small LXC** with **Podman** before Docker |

**Nested Docker-in-LXC:** avoid for new work — it doubles cgroup/network complexity and complicates OpenNebula backups. Prefer **one logical stack per LXC** with **no** inner container runtime when possible.

## Suggested native mapping (high level)

Use [NETWORK_DESIGN.md](../NETWORK_DESIGN.md) addresses on **host or LXC interfaces** (veth/macvlan to `100.64.*` VNETs as you mature networking).

| Current compose slice | Reduce Docker approach |
|----------------------|-------------------------|
| **Gitea** | Official **Gitea binary** + **systemd**; Postgres on same LXC or separate native DB LXC (not `postgres` container) |
| **PostgreSQL** (tooling, messaging DBs) | `postgresql` / `postgresql-server` + `initdb` + systemd |
| **HashiCorp Vault** | Official `vault` RPM/binary + systemd; **not** dev `server -dev` in prod |
| **Keycloak** | **Keycloak** Quarkus distribution ZIP or RPM where available; systemd `keycloak.service` (Java) — heavier than Gitea but **no** Keycloak container required |
| **Traefik / gateway** | `traefik` binary or **Caddy**/`haproxy` from Alma repos; static + dynamic configs from git |
| **n8n** | **Node LTS** from nodesource/module + `n8n` npm global or packaged; systemd **user** or system unit (supported install path is often Node-based outside Docker) |
| **Zammad** | **Hardest** — historically Rails+Elasticsearch stack; keep **one** Podman/stack exception short term or dedicated VM |
| **Nginx (docs)** | `nginx` package; clone/sync via **cron** or **git-hook** instead of `docs-sync` container where feasible |
| **Prometheus / Grafana** | `prometheus`, `alertmanager`, `grafana` packages or upstream tarballs + systemd |
| **Solace / Kafka / NiFi** | Often **OCI-first**; try **Podman** first; **last** resort Docker |

## Ansible / automation

- **`lxd_devsecops_stack`:** set **`lxd_install_docker_in_instance: false`** when the instance is **native-only** or **Podman-only** — see [role README](../../ansible/roles/lxd_devsecops_stack/README.md).
- **Future:** add `ansible/roles/devsecops_native_*` (or extend existing roles) to template systemd units from `devsecops.env.schema` keys; **do not** duplicate secrets in git.
- **Today:** use compose files as a **bill of materials** (ports, env, health checks) when writing unit files.

## OpenNebula-specific

- **One LXC = one systemd “stack”** (e.g. `devsecops-tooling-native` runs Gitea + local Postgres + optional nginx), or **split** DB for HA.
- **Snapshots:** `lxc snapshot` on the LXC is cleaner than coordinating Docker volume exports.
- **K3s:** For services already charted (e.g. Gitea), **Kubernetes can replace Docker** on that slice — see [04-gitea-k3s-ha.md](04-gitea-k3s-ha.md) — but that trades Docker for **kubelet**, not “no orchestrator.”

## Migration sequence

1. Stand up **native** or **Podman** LXC for **greenfield** services first (gateway, Gitea, Vault).
2. Retire matching services from `docker-compose.*.yml` **only after** parity tests (OIDC, webhooks, backups).
3. Keep **one** compose-based LXC only for **stragglers** (document which in this file).

## Related

- LXC topology: [LXC-ALMA10-OPENNEBULA.md](LXC-ALMA10-OPENNEBULA.md) — update mental model: **LXC first, Docker optional**.
- Volume moves: [CONTAINER-LIFT-TO-OPENNEBULA.md](CONTAINER-LIFT-TO-OPENNEBULA.md) — for remaining Docker/Podman volumes only.
