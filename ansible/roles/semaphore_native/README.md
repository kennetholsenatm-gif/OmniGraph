# semaphore_native

Installs **Semaphore UI** + **PostgreSQL** on AlmaLinux using **systemd** (no Docker).

Used by `lxd_devsecops_stack` when provisioning the `devsecops-semaphore` LXC with `lxd_native_only: true`.

## Defaults

See `defaults/main.yml` (version, DB password, admin password, `semaphore_web_host`, Incus proxy port).

## Optional bind mount

Set `semaphore_host_repo_path` (host path visible to Incus) to expose the control-plane repo at `/workspace` inside the LXC for Semaphore task templates.

## Firewall / exposure

The role adds an Incus **proxy** device mapping host `semaphore_incus_host_port` (default `3001`) to Semaphore listening on `127.0.0.1:semaphore_listen_port` inside the container.

## Troubleshooting installs

- Install log inside the LXC: `/tmp/semaphore-install.log` (the script tees stdout/stderr there).
- If `systemctl start semaphore` fails: `journalctl -u semaphore -n 100 --no-pager` inside the LXC.
- The script prefers the **GitHub `.rpm`** (`dnf install --nogpgcheck`) and falls back to the **`.tar.gz`** binary under `/usr/local/bin`.
