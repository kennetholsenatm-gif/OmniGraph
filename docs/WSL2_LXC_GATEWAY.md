# Traefik / gateway strategy (multi-LXC)

On a **single Docker host**, Traefik joins **many** bridge networks ([`single-pane-of-glass/docker-compose.yml`](../single-pane-of-glass/docker-compose.yml)). Splitting stacks across **multiple LXC** breaks that unless you redesign connectivity.

## Decision (locked for Phase 1)

| Option | Description | Use when |
|--------|-------------|----------|
| **A (recommended)** | Run **`devsecops-gateway`** LXC **plus** every backend Traefik must reach **on the same Docker daemon** | Accept a **larger** gateway LXC (merge tooling + gateway compose into one instance) — **not implemented as default** in repo; see Option B for current layout. |
| **B (current automation default)** | **Separate LXC per stack**; Traefik in **`devsecops-gateway`** LXC uses **published ports** or **extra_hosts** pointing at **other LXC `lxdbr0` IPs** | Matches [`lxd_devsecops_stack`](../ansible/roles/lxd_devsecops_stack) defaults. Requires **compose overrides** to replace `http://gitea:3000` style names with **`http://<peer-ip>:3000`** or run a **DNS forwarder**. |
| **C** | **socat / reverse proxy** on LXD host forwarding to per-LXC ports | Quick lab; not for production parity. |

## Implemented default: Option B

- Ansible creates **`devsecops-gateway`** and pushes **`single-pane-of-glass/`** only to that instance.
- **You must** either:
  1. Add **`docker-compose.override.yml`** (gitignored) on the gateway LXC with **`services.traefik.extra_hosts`** / **environment** overrides for each backend IP, **or**
  2. Consolidate backends that Traefik needs onto the **same** LXC as Traefik (fat gateway instance).

## Example override pattern (conceptual)

```yaml
# /opt/devsecops-pipeline/single-pane-of-glass/docker-compose.override.yml (do not commit secrets)
services:
  traefik:
    extra_hosts:
      - "gitea:10.x.x.x"   # IP of tooling LXC on lxdbr0 + docker gitea_net mapping is non-trivial; prefer same-LXC Docker DNS
```

Because **Docker DNS does not span LXCs**, **Option A** (fat LXC) is the least painful for Traefik until you add **external DNS** or **service mesh**.

## Recommendation

For **home lab**: use **one** `devsecops-app` LXC running **core compose files + gateway** together (single `docker compose` with multiple `-f` files), matching the original **single-host** model. Split **IAM** only for isolation experiments.

Document your choice in runbooks; OpenNebula phase can **re-split** by VM with routed IPs.
