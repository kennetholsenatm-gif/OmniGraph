# Local LXC runtime (canonical local target)

This is the canonical local deployment target for the repository.  
Goal: run the DevSecOps stack in LXC containers first, with Docker Compose nested inside LXC only where workloads require it.

The prior `deployments/wsl2-lxc/` path remains as a compatibility alias during migration.

## Why local LXC first

- Keeps local development closer to OpenNebula/KVM operating shape.
- Preserves existing compose files while reducing local-vs-staging drift.
- Supports WSL2, Linux workstation, or VM-hosted LXD/Incus.

## Runtime modes

- **Preferred:** LXC-first with nested Docker only per workload container.
- **Fallback:** Host Docker for temporary troubleshooting only (not canonical path).

### CPU and instance count

- **Default:** One LXC per stack (`devsecops-iam`, `devsecops-messaging`, …) — closest to future VM boundaries.
- **Lower CPU / fewer daemons:** Use **one LXC + one Docker** with instance **`devsecops-dev`** (IAM + [docker-compose.messaging.slim.yml](../../docker-compose/docker-compose.messaging.slim.yml): Postgres + RabbitMQ only; no Kafka, NiFi, Solace).
- **Slim messaging only:** Keep separate IAM LXC but run messaging with slim compose:
  `-e 'lxd_messaging_compose_cli="-f docker-compose.messaging.slim.yml"'` when applying `devsecops-messaging`.

## Directory layout

| Path | Purpose |
|------|---------|
| `profiles/docker-nesting.yaml` | Profile baseline for Docker-in-LXC |
| `cloud-init/docker-and-compose.user-data.yaml` | Optional cloud-init bootstrap |
| `scripts/bootstrap-lxd-profile.sh` | Create/update `docker-nesting` profile |
| `.env.example` | Local target defaults and selectors |

## Quick start

1. Initialize LXD/Incus on your local host.
2. Apply profile:
   ```bash
   ./deployments/local-lxc/scripts/bootstrap-lxd-profile.sh
   ```
3. Run Ansible LXC provisioning:
   ```bash
   cd ansible
   ansible-galaxy collection install -r collections/requirements.yml
   ansible-playbook -i inventory/lxc.example.yml playbooks/deploy-devsecops-lxc.yml \
     -e 'lxd_apply_names=["devsecops-iam","devsecops-messaging"]'
   ```
   **All-in-one dev (single nested Docker):** restrict instances so you do not also provision separate iam/messaging LXCs for the same host:
   ```bash
   ansible-playbook -i inventory/lxc.example.yml playbooks/deploy-devsecops-lxc.yml \
     -e 'lxd_apply_names=["devsecops-dev"]'
   ```
4. Validate IAM/messaging in-container compose startup per:
   - `docs/WSL2_LXC_MIGRATION.md`
   - `docs/WSL2_LXC_GATEWAY.md`

## Host notes

- **WSL2:** systemd must be enabled in distro config.
- **Alma/RHEL-compatible hosts:** use Incus packaging guidance from `deployments/wsl2-lxc/README.md` until this document is fully merged.
- **Line endings:** keep shell scripts as LF.

## OpenNebula alignment

Treat each local LXC workload boundary as a future VM boundary.
The network and VLAN design remains in `deployments/opennebula-kvm/VLAN_MATRIX.md`.
