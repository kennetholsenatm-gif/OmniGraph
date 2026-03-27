# IAM as code: OpenTofu, Ansible, Foreman, Pulumi, Packer

Keycloak realm, clients, LDAP federation, and automation accounts belong in **infrastructure-as-code**, not one-off scripts or manual UI. This doc maps that work to the right tools.

## Tool mapping

| Work | Primary tool | Notes |
|------|--------------|--------|
| **Keycloak realm, clients, identity providers** | **Ansible** (`community.general.keycloak_*`) or **OpenTofu** (Keycloak provider) | Realm, OIDC clients (gitea, n8n, zammad), automation client, LDAP user federation. Run after Keycloak is up (bootstrap or container stack). |
| **Docker networks, host networking** | **OpenTofu** (Docker provider) or **Ansible** (`community.docker.docker_network`) | Already in repo: `opentofu/` for networks; `deploy-devsecops-mesh.yml` for networks + UFW. |
| **Container stack (compose)** | **Ansible** (`community.docker.docker_compose_v2`) | Already in repo: `devsecops_containers` role. |
| **Host hardening, firewalld, FIPS** | **Ansible** | Already in repo: `os_hardening_fips`, `deploy-devsecops-mesh.yml`. |
| **Image build (golden images, agents)** | **Packer** | Build AMIs/VM images that include agents, Docker, or baseline config; Ansible can be the Packer provisioner. |
| **Provisioning (VMs, hosts)** | **Foreman** or **OpenTofu** (cloud/VM provider) | Foreman can kick off Ansible for config; OpenTofu can create instances and output inventory. |
| **Alternative IaC (Keycloak, cloud)** | **Pulumi** | If you prefer Pulumi over OpenTofu/Ansible for Keycloak or cloud resources, same concepts apply: Keycloak provider, realm/client/IdP as code. |

## In this repo

- **Ansible**
  - **`ansible/roles/keycloak_iam`** — Keycloak clients (automation + OIDC) and optional LDAP IdP. Uses `community.general.keycloak_client`. Auth: bootstrap admin (from Vault) or automation client token once it exists. Run after IAM stack is up.
  - **`ansible/playbooks/site.yml`** — Hardening, mesh, containers. Add an optional play or role include for `keycloak_iam` when Keycloak is available.
- **OpenTofu**
  - **`opentofu/`** — Docker networks only (no Keycloak provider yet). To manage Keycloak with OpenTofu, add a separate stack (e.g. `opentofu/keycloak/` or `opentofu/iam/`) with the [Keycloak provider](https://registry.terraform.io/providers/mrparkers/keycloak/latest) and define realm, clients, and identity providers as resources. Requires Keycloak URL and admin auth (bootstrap or automation client).

## Recommended flow

1. **Bootstrap Keycloak** — Start IAM stack (compose or Ansible) so Keycloak is running; one-time admin user created from env (bootstrap).
2. **Run IaC** — Ansible role `keycloak_iam` (or OpenTofu Keycloak stack) creates automation client, OIDC clients (gitea, n8n, zammad), and optionally LDAP user federation. Uses bootstrap admin for first run.
3. **Store automation secret** — After automation client is created, copy its client secret to Vault (`KEYCLOAK_AUTOMATION_CLIENT_SECRET`); from then on scripts and Ansible can use client credentials.
4. **Ongoing** — All changes to realm, clients, and IdPs go through Ansible or OpenTofu; no manual UI or one-off scripts for config.

## Foreman / Pulumi / Packer

- **Foreman:** Use as the orchestrator for provisioning and “run Ansible when host is ready.” Point Foreman at this repo’s playbooks and roles (e.g. `site.yml`, `keycloak_iam`).
- **Pulumi:** Use the Keycloak provider in your chosen language (TypeScript, Python, etc.) to define realm, clients, and IdPs; keep client secrets out of code (Vault or Pulumi secrets).
- **Packer:** Use Ansible (or shell) as the provisioner inside Packer templates to harden and preconfigure images; IAM config (Keycloak) stays in Ansible/OpenTofu and runs at deploy time, not image build time.
