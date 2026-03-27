# Autonomous Zero Trust DevSecOps Pipeline

Deployment artifacts for the pipeline: 100.64.0.0/10 segmented network, Solace mTLS A2A, n8n macro-orchestrator, Gitea, Zammad, **Bitwarden (Vaultwarden)**, Keycloak, **HashiCorp Vault**, Teleport JIT, and Varlock schema-driven secrets.

**Read first (site topology and phased delivery):** [docs/CANONICAL_DEPLOYMENT_VISION.md](docs/CANONICAL_DEPLOYMENT_VISION.md) · [docs/ROADMAP.md](docs/ROADMAP.md)

## Quick Start

**Greenfield (one command):** From repo root run `.\scripts\launch-greenfield.ps1`. Enter your admin username and password when prompted; the script creates networks, generates secrets, starts the stack, and writes to Vault. Walk away and come back to a fully stood-up infrastructure. Optionally add `-SaveVaultToken` to persist the Vault token for later runs (then use `.\scripts\start-from-vault.ps1` next time). See [docs/GREENFIELD_ONE_SHOT.md](docs/GREENFIELD_ONE_SHOT.md).

**Alternative (step-by-step):**

1. Read [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for execution order.
2. Create networks (Ansible, OpenTofu from `opentofu/`, or `.\scripts\create-networks.ps1`).
3. **Secrets:** No static `.env`. Generate and push to Vault with `scripts/secrets-bootstrap.ps1`; on later runs retrieve from Vault with `scripts/start-from-vault.ps1` (or use Ansible). See [docs/VARLOCK_USAGE.md](docs/VARLOCK_USAGE.md).
4. Start stacks via **Ansible** or, after `start-from-vault.ps1`, run `docker-compose/launch-stack.ps1` (merged core: IAM, messaging, tooling, **ChatOps**; file list in `docker-compose/stack-manifest.json`). See `ansible/roles/devsecops_containers/README.md` and `scripts/verify-stack-manifest.ps1`.
5. Import n8n workflows from `n8n-workflows/` and configure credentials.

## Layout

- **docs/** — **CANONICAL_DEPLOYMENT_VISION.md** (edge VyOS + IAM mini PC + Google Home profile), **ROADMAP.md** (P0–P3+), **BOOTSTRAP_USB_BUNDLE.md** (offline USB bundle before custom ISO), **NETWORK_COLLAPSED_IDENTITY_PLANE.md** (optional single zone for IdM/RADIUS/NAC), **GREENFIELD_ONE_SHOT.md** (pointer to Deploy repo), **GITEA_WIKI.md** (publish `wiki/gitea-pages` to Gitea wiki API), SYSTEMS_ARCHITECTURE.md, NETWORK_DESIGN.md, NETWORKS_PHASE1.md, **IAM_PHASE2.md**, **IAM_LDAP_AND_AUTOMATION.md**, **IAM_IAC.md** (OpenTofu/Ansible/Foreman/Pulumi/Packer for IAM), DEPLOYMENT.md, **CI_CD.md** (Gitea Actions, Semaphore sync), **BRANCHING_AND_SCANNING.md** (branch layout + scanners), A2A_PAYLOAD_SCHEMA.md, TELEPORT_JIT.md, VARLOCK_USAGE.md; **opennebula-gitea-edge/** — **REDUCE-DOCKER.md** (native/Podman first), **LXC-ALMA10-OPENNEBULA.md**, **CONTAINER-LIFT-TO-OPENNEBULA.md**, **WHOLE-REPO-MIGRATION-SCOPE.md**
- **deployments/bootstrap-usb-bundle/** — `build-bundle.sh` + `bootstrap-on-target.sh` for offline mini PC bring-up
- **opentofu/** — Docker network definitions (100.64.x.x)
- **ansible/** — Playbooks and roles (mesh deploy, mTLS, FIPS/hardening)
- **docker-compose/** — Messaging, tooling (Gitea/n8n/Zammad/Bitwarden), IAM (Vault + Keycloak), ChatOps (Zulip); optional discovery, LLM, AI orchestration. **stack-manifest.json** defines the merged “core” set for PowerShell (`launch-stack.ps1`, `secrets-bootstrap.ps1`).
- **single-pane-of-glass/** — Unified gateway: Traefik, dashboard (wiki + LMNotebook), webhook listener; see [single-pane-of-glass/README.md](single-pane-of-glass/README.md)

**Canonical location:** **C:\GiTeaRepos\devsecops-pipeline**. Clone from Gitea and work there; do not use C:\GitHub\LLM_Pract for this repo.
- **docs/WIKI_EXPORT/** — Optional local wiki/BOM exports (not tracked; see `.gitignore`).
- **n8n-workflows/** — DevSecOps Orchestrator, Security Audit, Documentation Generator (JSON)
- **devsecops.env.schema** — Unified Varlock schema (no secrets)

Solace VPN and A2A ACL are in your infra repo (discovery-networks.tf, devsecops-variables.tf). This repo lives at **C:\GiTeaRepos\devsecops-pipeline** (clone from Gitea).
