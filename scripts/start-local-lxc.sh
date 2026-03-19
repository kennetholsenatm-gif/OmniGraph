#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ANSIBLE_DIR="$REPO_ROOT/ansible"
INVENTORY="${1:-inventory/lxc.example.yml}"
NAMES_JSON="${2:-[\"devsecops-iam\",\"devsecops-messaging\"]}"

echo "Starting local LXC target via Ansible..."
echo "Inventory: $INVENTORY"
echo "Instances: $NAMES_JSON"

cd "$ANSIBLE_DIR"
ansible-galaxy collection install -r collections/requirements.yml

# Always prompt for become password in WSL/mnt-c contexts for deterministic behavior.
export ANSIBLE_BECOME_ASK_PASS=true
ansible-playbook -K -i "$INVENTORY" playbooks/deploy-devsecops-lxc.yml -e "lxd_apply_names=$NAMES_JSON"
