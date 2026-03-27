# lxd_devsecops_stack

Creates **LXD/Incus** instances (`images:almalinux/10`), configures the **`docker-nesting`** profile **inline** when you still run **Docker/Podman** in a guest, pushes `docker-compose/` + `scripts/` into `/opt/devsecops-pipeline/`, runs **`create-networks.sh`** (Docker-centric today), optionally **`docker compose up`**.

**Native control plane (no Docker):** the optional **`devsecops-semaphore`** instance sets **`lxd_native_only: true`** and installs **Semaphore + PostgreSQL** via **`roles/semaphore_native`** (systemd). Use **`ansible/playbooks/deploy-semaphore-incus.yml`** or `./scripts/start-semaphore.sh`.

**Reduce Docker:** For OpenNebula / Alma **native** stacks, set **`lxd_install_docker_in_instance: false`**, skip `devsecops_lxc_compose_up`, and manage **systemd + dnf/Podman** per [docs/opennebula-gitea-edge/REDUCE-DOCKER.md](../../../docs/opennebula-gitea-edge/REDUCE-DOCKER.md). The role remains useful to **provision empty LXCs** and sync the repo tree even without `docker-ce`.

Requires **`lxc` (LXD)** or **`incus` (Incus)** on the target **lxd_host**; use `lxd_cli` / `lxd_incus_socket` if needed. Optional: `deployments/local-lxc/scripts/bootstrap-lxd-profile.sh` for manual profile work.

## Requirements

- Control node: `lxc` CLI (LXD/Incus), `ansible-galaxy collection install -r collections/requirements.yml`
- `community.general` (LXD module), `community.docker` optional for future use

## Variables

| Variable | Default | Meaning |
|----------|---------|---------|
| `lxd_apply_names` | `[]` | If non-empty, only these instance names (e.g. `["devsecops-iam"]`) |
| `devsecops_lxc_compose_up` | `false` | Set `true` to run compose after sync (you must inject env first) |
| `lxd_install_docker_in_instance` | `true` | Set `false` to skip `dnf` Docker install (e.g. fix DNS first, install manually) |
| `lxd_sync_host_resolv_conf` | `false` | Set `true` to `incus file push` controller `/etc/resolv.conf` into each instance before dnf (WSL/Incus DNS) |
| `lxd_dns_fallback_enabled` | `true` | If DNS still fails, write static resolver list in each instance and re-test before dnf |
| `lxd_dns_fallback_servers` | `["1.1.1.1","8.8.8.8"]` | Nameservers used to render `/etc/resolv.conf` fallback inside instances |
| `lxd_image` | `images:almalinux/10` | CLI-style `images:…` shorthand; role strips `images:` for the API |
| `lxd_image_simplestreams_server` | `https://images.linuxcontainers.org` | Passed to `lxd_container` `source.server` |
| `lxd_image_alias` | `almalinux/10` | Used when `lxd_image` does not start with `images:` |
| `lxd_ensure_default_storage` | `true` | Create pool `lxd_storage_pool_name` (dir) and `default` profile `root` disk if missing |
| `lxd_storage_pool_name` | `default` | Storage pool used for the `default` profile root disk |
| `lxd_ensure_idmap` | `true` | Ensure `newuidmap`/`newgidmap`, `root:` lines in `/etc/sub{u,g}id`, restart Incus/LXD if fixed |
| `lxd_idmap_root_line` | `root:100000:65536` | Appended only when no `root:` line exists (needs ≥65536 IDs per Incus) |
| `lxd_docker_profile_kernel_modules` | `""` | Comma-separated modules for profile `linux.kernel_modules`; empty skips (WSL: Incus can’t find `modprobe`). Set full list on bare metal + `kmod` |
| `lxd_docker_profile_raw_lxc` | `""` | Multiline profile `raw.lxc`; empty skips/unsets (avoids liblxc temp config failures on WSL/CRLF). See `defaults/main.yml` for tun/AppArmor example |
| `devsecops_pipeline_root` | from playbook | Repo root |
| `lxd_bridge_tune_dns` | `true` | After creating the managed bridge, run `network set` for `dns.mode`, `ipv4.nat`, `ipv4.dhcp`, `dns.nameservers` |
| `lxd_bridge_dns_nameservers` | `["1.1.1.1","8.8.8.8"]` | Upstream resolvers advertised on the bridge (helps fix “no DNS in container”) |
| `lxd_ensure_managed_bridge` | `true` | Create `incusbr0` / `lxdbr0` if missing and attach `default` profile NIC |

See [docs/INCUS_NETWORK_DNS.md](../../../docs/INCUS_NETWORK_DNS.md) for manual verification and troubleshooting.

## Example

```bash
ansible-playbook -i inventory/lxc.example.yml playbooks/deploy-devsecops-lxc.yml \
  -e 'lxd_apply_names=["devsecops-iam","devsecops-messaging"]'
```

## Traefik / gateway

Gateway instance needs **`single-pane-of-glass`** pushed (done when `devsecops-gateway` is in the filter). See `docs/WSL2_LXC_GATEWAY.md`.
