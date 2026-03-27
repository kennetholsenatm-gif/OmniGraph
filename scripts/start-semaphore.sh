#!/usr/bin/env bash
# Provision Semaphore on an Incus LXC (Alma + Postgres + systemd). No Docker.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${REPO_ROOT}/ansible"

if command -v ansible-galaxy >/dev/null 2>&1; then
  ansible-galaxy collection install -r collections/requirements.yml
fi

ansible-playbook -i inventory/lxc.example.yml playbooks/deploy-semaphore-incus.yml \
  -e lxd_become=false \
  -e lxd_manage_daemon=false \
  -e lxd_ensure_idmap=false \
  -e lxd_incus_socket=/run/incus/unix.socket \
  -e 'lxd_apply_names=["devsecops-semaphore"]' \
  "$@"

echo "Open Semaphore UI (default): http://127.0.0.1:3001"
