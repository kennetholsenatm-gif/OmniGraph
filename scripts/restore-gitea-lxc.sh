#!/usr/bin/env bash
set -euo pipefail

INVENTORY="${1:-inventory/lxc.example.yml}"
COMPOSE_UP="${COMPOSE_UP:-false}"
INCUS_SOCKET_PATH="${INCUS_SOCKET_PATH:-/run/incus/unix.socket}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if ! command -v ansible-playbook >/dev/null 2>&1; then
  echo "ERROR: ansible-playbook not found in PATH." >&2
  exit 1
fi

cd "${REPO_ROOT}/ansible"
ansible-galaxy collection install -r collections/requirements.yml

EXTRA=(
  -e 'lxd_apply_names=["devsecops-gitea"]'
  -e 'lxd_optional_stack_enable={gitea_lite:true}'
  -e 'lxd_become=false'
  -e 'lxd_manage_daemon=false'
  -e 'lxd_ensure_idmap=false'
  -e "lxd_incus_socket=${INCUS_SOCKET_PATH}"
  -e 'lxd_install_docker_in_instance=true'
)

if [[ "${COMPOSE_UP}" == "true" ]]; then
  EXTRA+=(-e 'devsecops_lxc_compose_up=true')
fi

ansible-playbook -i "${INVENTORY}" playbooks/deploy-devsecops-lxc.yml "${EXTRA[@]}"
echo "Gitea LXC restore command completed."
