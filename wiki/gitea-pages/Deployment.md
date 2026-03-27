# Deployment checklist

Condensed from `docs/DEPLOYMENT.md`. **No secrets here** — use Vault / Varlock.

## Order of operations

1. **Network design** — Finalize `100.64.x.x` and host firewall rules (`docs/NETWORK_DESIGN.md`).
2. **Docker networks** — `scripts/create-networks.ps1` / `create-networks.sh`, `opentofu/`, or `ansible/playbooks/deploy-devsecops-mesh.yml` (**17** bridges; IAM includes optional FreeIPA on `iam_net`).
3. **Secrets** — `scripts/secrets-bootstrap.ps1` (Deploy repo) or Ansible; see `docs/VARLOCK_USAGE.md`. Do **not** rely on a long-lived `.env` for production.
4. **Stacks** — Core compose set in `docker-compose/stack-manifest.json`; verify with `scripts/verify-stack-manifest.ps1`.
5. **IAM** — Vault, Keycloak, keycloak-proxy (`http://127.0.0.1:8180` for admin UI per docs).
6. **Gateway (optional)** — `single-pane-of-glass/` Traefik + dashboard; see [[Gateway]].

## Pointers

| Topic | Doc path |
|-------|-----------|
| Greenfield one-shot | `docs/GREENFIELD_ONE_SHOT.md` |
| OpenNebula / edge | `docs/opennebula-gitea-edge/` |
| USB bootstrap | `docs/BOOTSTRAP_USB_BUNDLE.md`, `deployments/bootstrap-usb-bundle/` |
| CI / branching | `docs/CI_CD.md`, `docs/BRANCHING_AND_SCANNING.md` |

## Brownfield note

If you previously used **`freeipa_net`**, remove it after stopping FreeIPA; FreeIPA uses **`iam_net`** only. See `docs/DEPLOYMENT.md` (OpenTofu state note).
