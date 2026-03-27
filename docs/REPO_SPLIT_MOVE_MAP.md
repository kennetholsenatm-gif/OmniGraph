# Repo split move map (`devsecops-pipeline` -> `deploy`)

This file defines which greenfield/bootstrap assets move to `C:\GiTeaRepos\Deploy`.

## Scope policy

- Keep in `devsecops-pipeline`: control-plane automation, OpenNebula runtime orchestration, network/device automation, local Incus + Gitea runbooks.
- Move to `deploy`: greenfield day-0 bootstrap assets and one-shot environment bring-up wrappers/docs.

## Move map

| Source (`devsecops-pipeline`) | Destination (`Deploy`) | Reason |
|---|---|---|
| `docs/GREENFIELD_ONE_SHOT.md` | `docs/GREENFIELD_ONE_SHOT.md` | Greenfield one-shot process belongs to deployment repo |
| `docs/GREENFIELD_REGISTRATION.md` | `docs/GREENFIELD_REGISTRATION.md` | Greenfield registration workflow |
| `scripts/launch-greenfield.ps1` | `scripts/launch-greenfield.ps1` | One-shot greenfield launcher |
| `scripts/secrets-bootstrap.ps1` | `scripts/secrets-bootstrap.ps1` | Greenfield secret bootstrap |
| `scripts/secrets-bootstrap.sh` | `scripts/secrets-bootstrap.sh` | Linux greenfield secret bootstrap |
| `scripts/populate-vault-secrets.ps1` | `scripts/populate-vault-secrets.ps1` | Initial secret population for deployment bring-up |

## Keep map (explicit)

| Path | Keep in `devsecops-pipeline` |
|---|---|
| `ansible/playbooks/opennebula-hybrid-site.yml` | Core control-plane orchestrator for OpenNebula |
| `ansible/playbooks/deploy-devsecops-lxc.yml` | Core LXC runtime provisioning |
| `scripts/setup-lean-local-control.ps1` / `.sh` | Local control-plane bootstrap |
| `deployments/local-control/semaphore/` | Local control-plane service only |
| `docs/opennebula-gitea-edge/` | OpenNebula runtime and migration runbooks |

## Compatibility notes

- After move, `devsecops-pipeline` keeps pointer docs/scripts that reference `Deploy` for day-0 greenfield actions.
- No secrets are moved into git history as plain text; only scripts/docs are moved.
