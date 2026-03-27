# Compose overrides (LXC / multi-host)

Place **local-only** `docker-compose.override.yml` files here as templates, then copy into the target LXC (e.g. `/opt/devsecops-pipeline/single-pane-of-glass/`) for **Traefik `extra_hosts`**, published ports, or TLS.

**Do not commit** secrets. Use `.gitignore` patterns if you keep examples with placeholders only.

See [docs/WSL2_LXC_GATEWAY.md](../../docs/WSL2_LXC_GATEWAY.md).
