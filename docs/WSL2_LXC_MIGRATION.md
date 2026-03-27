# Local LXC migration order (before OpenNebula)

Follow this **order** so dependencies and Docker networks come up predictably. Each step assumes **one LXD instance** per row, **Docker-in-LXC**, **100.64.x** bridges created with [`scripts/create-networks.sh`](../scripts/create-networks.sh).

Canonical target reference: [`deployments/local-lxc/README.md`](../deployments/local-lxc/README.md).  
WSL2-specific caveats remain in [`deployments/wsl2-lxc/README.md`](../deployments/wsl2-lxc/README.md).

**Hypervisor on AlmaLinux 10:** `dnf install incus` from BaseOS/CRB will **not** work; use COPR **`neelc/incus`** (RHEL+EPEL 10 repo file) as documented in [`deployments/wsl2-lxc/README.md`](../deployments/wsl2-lxc/README.md). Then use **`incus`** wherever examples show **`lxc`** (same subcommands).

**Ansible on WSL + `/mnt/c/`:** World-writable dirs cause Ansible to **ignore** CWD `ansible.cfg`. Use `ansible/playbooks/run-deploy-devsecops-lxc.sh` with the same args you would pass to `ansible-playbook`, or `cd ansible` and run `ansible-playbook -i inventory/lxc.example.yml playbooks/deploy-devsecops-lxc.yml ...` from a path where config is honored, or set `ANSIBLE_CONFIG` + `ANSIBLE_ROLES_PATH` explicitly.

## Phase 1 — IAM

1. `cd ansible && ansible-playbook -i inventory/lxc.example.yml playbooks/deploy-devsecops-lxc.yml -e 'lxd_apply_names=["devsecops-iam"]'`
2. After Vault is up:  
   `lxc exec devsecops-iam -- bash -lc 'export DEVSECOPS_REPO_ROOT=/opt/devsecops-pipeline; /opt/devsecops-pipeline/scripts/secrets-bootstrap.sh --no-start'`
3. Start IAM stack only (inside LXC):
   ```bash
   lxc exec devsecops-iam -- bash -lc 'cd /opt/devsecops-pipeline/docker-compose && docker compose -f docker-compose.iam.yml up -d'
   ```
4. When Vault is healthy, re-push or refresh secrets with `VAULT_TOKEN` set (same command as step 2 with `--no-start`).

**Validate:** `curl -s http://127.0.0.1:8200/v1/sys/health` from inside LXC; Keycloak via published port or `lxc exec` curl.

## Phase 2 — Messaging

1. `ansible-playbook ... -e 'lxd_apply_names=["devsecops-messaging"]'`
2. Ensure **routing/DNS** from messaging LXC to IAM LXC for Keycloak/Vault if automation needs it (add **host records** on `lxdbr0` IPs or **extra_hosts** in compose overrides — future enhancement).
3. `lxc exec devsecops-messaging -- bash -lc 'cd /opt/devsecops-pipeline/docker-compose && docker compose -f docker-compose.messaging.yml up -d'`

**Validate:** `docker exec devsecops-postgres pg_isready` inside messaging LXC.

## Phase 3 — Tooling, ChatOps, Telemetry, Gateway

See [WSL2_LXC_GATEWAY.md](WSL2_LXC_GATEWAY.md) for **Traefik** placement.

Use `-e 'lxd_apply_names=[...]'` per instance; re-run **secrets-bootstrap.sh** or Ansible Vault injection so env matches.

## Cross-LXC networking checklist

- [ ] Each LXC has **unique `lxdbr0` address**; note IPs with `lxc list`.
- [ ] For services that used **Docker DNS** on a single host, add **`extra_hosts`** or **routed** access to peer LXC bridge IP (document peer IPs in NetBox or `deployments/wsl2-lxc/README.md` appendix).
- [ ] Publish **80/443/8200** from Windows → WSL → LXC as needed (`netsh portproxy` or WSL mirrored networking).

## OpenNebula next

Each LXC maps cleanly to a **VM**; keep the same compose trees and [VLAN_MATRIX](../deployments/opennebula-kvm/VLAN_MATRIX.md) for hardware.
