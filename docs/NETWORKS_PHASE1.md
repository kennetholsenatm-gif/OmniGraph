# Phase 1: Network & Gateway — Docker networks

Networks required for the Single Pane of Glass (Traefik) to route to Gitea, Keycloak, n8n, and Zammad. Subnets align with [NETWORK_DESIGN.md](NETWORK_DESIGN.md).

## Option A: OpenTofu (canonical)

From the repo root:

```bash
cd opentofu
tofu init
tofu apply
```

This creates all pipeline networks (including `gateway_net`, `gitea_net`, `n8n_net`, `zammad_net`, `iam_net`) with the defined subnets. See `opentofu/variables.tf` for CIDRs.

## Option B: Scripts (no OpenTofu)

From the repo root:

- **Windows (PowerShell):** `.\scripts\create-networks.ps1`
- **Linux/macOS:** `./scripts/create-networks.sh` (chmod +x if needed)

Creates only the five Phase 1 networks:

| Network      | Subnet          |
|-------------|------------------|
| gateway_net | 100.64.5.0/24   |
| gitea_net   | 100.64.1.0/24   |
| n8n_net     | 100.64.2.0/24   |
| zammad_net  | 100.64.3.0/24   |
| iam_net     | 100.64.20.0/24  |

Start order: create networks first, then start IAM/tooling stacks, then the gateway (`single-pane-of-glass`).

## HTTPS (TLS)

To enable HTTPS: copy `single-pane-of-glass/traefik/dynamic/tls.yml.example` to `tls.yml` in the same directory, then mount your server certificate and key into the gateway volume `gateway_tls` as `tls.crt` and `tls.key` (container path `/etc/traefik/tls/`). Without `tls.yml` or with an empty volume, Traefik starts normally and only HTTP (port 80) is used. See [single-pane-of-glass/README.md](../single-pane-of-glass/README.md#tls--https).
