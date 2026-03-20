# devsecops_env

Merges `devsecops_env_defaults` (from `devsecops_containers` role defaults) with `devsecops_secrets` into:

- `devsecops_combined_env` — generic name for reuse (e.g. LXC env-file injection).
- `container_env` — alias expected by `devsecops_containers` after `include_role`.

**Requires:** `devsecops_secrets` (Ansible Vault group_vars or extra vars).

**Used by:**

- `devsecops_containers`
- `lxd_devsecops_stack` when `lxd_inject_ansible_env: true`
