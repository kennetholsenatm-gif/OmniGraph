# devsecops_containers

Starts DevSecOps Docker Compose stacks with **env injected by Ansible**. No manual `.env` file; all variables (including secrets) are passed into containers at start time from Ansible vars or from HashiCorp Vault.

## How env gets into containers

- Ansible builds a single `container_env` dict from:
  - **Non-secret defaults** (`role defaults/main.yml`)
  - **Secrets** from `devsecops_secrets` (see below)
- Ansible runs `community.docker.docker_compose_v2` with `env: "{{ container_env }}"`, so Compose substitutes `${VAR}` in the compose files and containers receive env from Ansible only.

## Providing devsecops_secrets

**Option A — HashiCorp Vault (recommended)**  
Run the playbook that fetches from Vault and passes to this role:

```bash
export VAULT_ADDR=http://vault:8200
export VAULT_TOKEN=your-token
ansible-playbook -i inventory.yml playbooks/start-containers-with-vault.yml
```

Secrets must be stored at `secret/data/devsecops` (KV2). Keys match `devsecops.env.schema`. Prefer `KEYCLOAK_AUTOMATION_CLIENT_ID` and `KEYCLOAK_AUTOMATION_CLIENT_SECRET` for Keycloak automation (see docs/IAM_LDAP_AND_AUTOMATION.md); `KEYCLOAK_ADMIN_PASSWORD` is bootstrap-only. Other keys: `GITEA_API_TOKEN`, `ZAMMAD_POSTGRES_PASSWORD`, etc.

**Option B — Ansible Vault (encrypted group_vars)**  
1. Create `group_vars/ai_mesh_nodes/devsecops_secrets.yml` (or all.yml) with a single dict, e.g.:

   ```yaml
   devsecops_secrets:
     GITEA_API_TOKEN: "your-token"
     KEYCLOAK_ADMIN: "admin"
     KEYCLOAK_ADMIN_PASSWORD: "secret"
     # ... all keys from devsecops.env.schema that are @sensitive
   ```

2. Encrypt it: `ansible-vault encrypt group_vars/ai_mesh_nodes/devsecops_secrets.yml`
3. Run the role from a playbook that includes this role (and passes `devsecops_secrets` if not in group_vars), or ensure the play sets `devsecops_secrets` from group_vars.

**Option C — Extra vars file (gitignored)**  
`ansible-playbook -e @secrets.yml playbooks/site-containers.yml` where `secrets.yml` contains `devsecops_secrets: { ... }`. Do not commit `secrets.yml`.

## Role vars

| Var | Default | Description |
|-----|---------|-------------|
| `devsecops_secrets` | (required) | Dict of secret env vars; from Vault or vault-encrypted vars |
| `devsecops_env_defaults` | (role defaults) | Non-secret defaults |
| `start_messaging` | true | Start messaging backbone compose |
| `start_iam` | true | Start IAM compose |
| `start_tooling` | true | Start tooling compose |
| `start_chatops` | true | Start ChatOps / Zulip (`docker-compose.chatops.yml`) |
| `start_discovery` | false | Start NetBox / Dep-Track / Netdisco (`docker-compose.discovery.yml`) |
| `start_llm` | false | Start BitNet LLM stack (`docker-compose.llm.yml`; requires `llm_net`) |
| `start_ai_orchestration` | false | Start n8n-mcp (`docker-compose.ai-orchestration.yml`) |
| `start_sdn_telemetry` | false | Start `docker-compose.network.yml` then `docker-compose.telemetry.yml` (Linux SDN host; requires `sdn_lab_net`, `telemetry_net`) |
| `start_gateway` | true | Start Single Pane of Glass compose |

## Stacks started

1. Messaging (`docker-compose.messaging.yml`)
2. IAM (`docker-compose.iam.yml`)
3. Tooling (`docker-compose.tooling.yml`)
4. ChatOps (`docker-compose.chatops.yml`)
5. Optional: Discovery, LLM, AI orchestration when the corresponding `start_*` flags are true
6. Optional: SDN + telemetry (`docker-compose.network.yml`, `docker-compose.telemetry.yml`) when `start_sdn_telemetry: true`
7. Gateway (`single-pane-of-glass/docker-compose.yml`)

**Compose manifest:** PowerShell (`docker-compose/launch-stack.ps1`, `scripts/secrets-bootstrap.ps1`) uses a **single merged** `docker compose` run for the core stack files listed in `docker-compose/stack-manifest.json`. This role uses **separate** `project_name` values per stack so projects can be managed independently; service names must remain unique across files. Run `scripts/verify-stack-manifest.ps1` after changing the manifest or these tasks.

Networks must exist (create via `deploy-devsecops-mesh.yml` or OpenTofu).
