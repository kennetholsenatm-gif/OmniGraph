# Populate Semaphore (remote Git + repo mirror)

This guide matches **native Semaphore** (`semaphore server --no-config` in Incus) and the **[Semaphore REST API](https://docs.semaphoreui.com/administration-guide/api/)** (tested against the same major version as [`semaphore_version`](../ansible/roles/semaphore_native/defaults/main.yml), currently **2.17.x**).

## Base URL and port

- **Browser / API from the host:** `http://localhost:3001` (Incus proxy) — use the **same** host you use for [`SEMAPHORE_WEB_ROOT`](SEMAPHORE_INCUS_TROUBLESHOOTING.md).
- **Inside the LXC:** Semaphore listens on **`:3000`**; API paths are still **`/api/...`** (same as public docs).

## Prerequisites

1. Semaphore is **running** and you can log in (see [SEMAPHORE_INCUS_TROUBLESHOOTING.md](SEMAPHORE_INCUS_TROUBLESHOOTING.md) if not).
2. **API token** with permission to manage projects:
   - **UI:** User menu → API tokens → create (recommended).
   - **HTTP:** `POST /api/auth/login` then `POST /api/user/tokens` (see [API docs](https://docs.semaphoreui.com/administration-guide/api/)).
3. **Remote Git** URL and branch your repo lives on (e.g. Gitea/GitHub HTTPS or SSH).
4. **Git credentials in Semaphore:**
   - **Public HTTPS clone:** use the **None** access key (created automatically with each new project).
   - **Private repo:** create a **Key** in the project:
     - **SSH:** type `ssh`, paste the **private** deploy key (read-only) or user key.
     - **HTTPS:** type `login_password` or use a token via Semaphore’s key types as supported by your version — prefer **SSH deploy keys** for automation.

## Semaphore object model (what we mirror)

| Object | Purpose |
|--------|---------|
| **Project** | Top-level bucket (e.g. `devsecops-pipeline`). |
| **Key** | SSH key / none / password material for **Git** and/or **Ansible SSH** (referenced by `ssh_key_id`). |
| **Repository** | Git URL + branch + `ssh_key_id` (required; use **None** for public HTTPS). |
| **Inventory** | `static` (INI), `static-yaml`, or `file` (path inside cloned repo). Phase A uses **static** INI via `semaphore_inventory_body`. |
| **Environment** | Optional JSON env vars; new projects get an **Empty** environment by default. |
| **Template** | Ansible **app** + playbook path + **repository_id** + **inventory_id** (+ optional `environment_id`). |

See also: [Semaphore User Guide — Inventory](https://docs.semaphoreui.com/user-guide/inventory/).

## Phase A (recommended): one repo + one smoke template

Goal: **clone** from Git, run **one** playbook that only touches **localhost**.

1. **Commit** [`ansible/playbooks/semaphore-smoke.yml`](../ansible/playbooks/semaphore-smoke.yml) on your branch (it is safe to run in CI).
2. Set variables (see [Example inventory](#example-inventory)) and run:

```bash
cd ansible
export ANSIBLE_CONFIG="$(pwd)/ansible.cfg"
ansible-playbook -i inventory/semaphore-populate.example.yml playbooks/populate-semaphore.yml \
  -e semaphore_api_token="YOUR_TOKEN" \
  -e semaphore_git_url="https://git.example/your/devsecops-pipeline.git" \
  -e semaphore_git_branch="main"
```

3. In the UI: **Project → Tasks** or run via API:

```bash
curl -sS -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"template_id": TEMPLATE_ID}' \
  "http://localhost:3001/api/project/PROJECT_ID/tasks"
```

(`template_id` is returned when the playbook runs; use **GET** `/api/project/{id}/templates` if needed.)

4. **Troubleshooting task failures**
   - **Git clone fails:** DNS/firewall from the LXC to your forge; wrong URL; private repo without key.
   - **Ansible fails:** `ansible.cfg` in repo — Semaphore runs from the cloned tree; ensure `roles_path` / `inventory` are correct relative to CWD Semaphore uses (often repo root).
   - **`become_password` prompts:** [`ansible.cfg`](../ansible/ansible.cfg) has `become_ask_pass = True` — Semaphore does not handle interactive prompts. Use passwordless sudo on targets or a different strategy for scheduled runs.

## Phase B: mirror more playbooks

- Add **Templates** per playbook under `ansible/playbooks/` (group by domain: Incus, network, etc.).
- Add **Inventories** (static YAML or `file` type pointing at `ansible/inventory/...`).
- Document **network path** from the LXC to **targets** (SSH) and to **Git**.

## Phase C: hardening

- Rotate admin password; **expire** unused API tokens (`DELETE /api/user/tokens/{token_id}`).
- Prefer **read-only** deploy keys for Git; **never** commit tokens or private keys to this repo — use Vault / Ansible Vault for `semaphore_api_token` and key material.

## Example inventory

Copy [`ansible/inventory/semaphore-populate.example.yml`](../ansible/inventory/semaphore-populate.example.yml) to a local path (or use `-e` / `@vault.yml`) and set:

- **`semaphore_api_token`** — never commit this file with a real token.
- **`semaphore_git_url`** / **`semaphore_git_branch`** — your remote.
- Optional **`semaphore_git_ssh_private_key`** — for private Git over SSH (also set `semaphore_git_url` to an `ssh://` or `git@` URL).

## Automation in this repo

### GitOps at scale (recommended): manifest + sync

Semaphore is **not** Terraform: it does not auto-discover every playbook in Git. The scalable pattern is **declare desired Semaphore objects in this repo** and **apply** them with Ansible (or CI).

1. **First run:** run [`sync-semaphore-from-manifest.yml`](../ansible/playbooks/sync-semaphore-from-manifest.yml) once — if `ansible/files/semaphore-manifest.yml` is missing, the playbook **creates it from** `semaphore-manifest.example.yml` (no manual copy).  
   **Optional manual copy (Windows PowerShell from repo root):**  
   `Copy-Item -LiteralPath ansible\files\semaphore-manifest.example.yml -Destination ansible\files\semaphore-manifest.yml`  
   (`semaphore-manifest.yml` is **gitignored** — put URLs/tokens via `-e` / Vault / CI secrets, not committed files.)
2. Edit the manifest (after it exists): add rows under `inventories` and `ansible_templates` (each template = one playbook path).
3. Apply:

```bash
cd ansible
export ANSIBLE_CONFIG="$(pwd)/ansible.cfg"
ansible-playbook -i localhost, -c local playbooks/sync-semaphore-from-manifest.yml \
  -e semaphore_api_token="$TOKEN" \
  -e semaphore_git_url="https://your.forge/you/devsecops-pipeline.git"
```

Optional: run that playbook from **CI** on merge to `main` so Semaphore tracks Git without anyone clicking in the UI. Setup (secrets, runner networking): [CI_CD.md](CI_CD.md).

- **Playbook:** [`ansible/playbooks/sync-semaphore-from-manifest.yml`](../ansible/playbooks/sync-semaphore-from-manifest.yml)
- **Example manifest:** [`ansible/files/semaphore-manifest.example.yml`](../ansible/files/semaphore-manifest.example.yml)
- **Tasks:** [`ansible/tasks/semaphore_sync_from_manifest.yml`](../ansible/tasks/semaphore_sync_from_manifest.yml) (and `semaphore_sync_one_inventory.yml`, `semaphore_sync_one_template.yml`)

**Terraform / OpenTofu** in this repo are **not** wired by the current manifest (Ansible templates only). Extend the sync tasks / manifest schema when you add Semaphore Terraform/Tofu templates.

### Phase A only (single smoke template)

- **Playbook:** [`ansible/playbooks/populate-semaphore.yml`](../ansible/playbooks/populate-semaphore.yml) — idempotent create (Project, Repository, optional SSH key, one Inventory, one Template).
- **Example vars:** [`ansible/inventory/semaphore-populate.example.yml`](../ansible/inventory/semaphore-populate.example.yml).

## Verify with curl

```bash
BASE="http://localhost:3001"
TOKEN="YOUR_TOKEN"
curl -sS -H "Authorization: Bearer $TOKEN" "$BASE/api/projects" | jq .
curl -sS -H "Authorization: Bearer $TOKEN" "$BASE/api/project/PROJECT_ID/repositories" | jq .
curl -sS -H "Authorization: Bearer $TOKEN" "$BASE/api/project/PROJECT_ID/templates" | jq .
```

## Related docs

- [SEMAPHORE_INCUS_TROUBLESHOOTING.md](SEMAPHORE_INCUS_TROUBLESHOOTING.md) — install, `WEB_ROOT`, DB, admin bootstrap.
- [DEPLOYMENT.md](DEPLOYMENT.md) — local control plane overview.
