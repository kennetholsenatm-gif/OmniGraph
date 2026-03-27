#!/usr/bin/env bash
# Run ISR playbooks from WSL/Linux with /mnt/c (world-writable) without losing ansible.cfg.
#
# Usage (from repo ansible/):
#   chmod +x playbooks/run-network-isr.sh
#   ./playbooks/run-network-isr.sh --check --ask-pass
#   ./playbooks/run-network-isr.sh --diff
#
# Until ansible_password is set in inventory (or Vault), pass --ask-pass every run or Ansible
# reports "No authentication methods available".
set -eu
PLAYBOOKS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
ANSIBLE_DIR="$(cd "$PLAYBOOKS_DIR/.." && pwd)"
export ANSIBLE_CONFIG="$ANSIBLE_DIR/ansible.cfg"
export ANSIBLE_ROLES_PATH="$ANSIBLE_DIR/roles"
# /mnt/c may ignore ansible.cfg; network_cli+paramiko still needs explicit host-key policy.
export ANSIBLE_HOST_KEY_CHECKING="${ANSIBLE_HOST_KEY_CHECKING:-false}"
export ANSIBLE_HOST_KEY_AUTO_ADD="${ANSIBLE_HOST_KEY_AUTO_ADD:-true}"
# Network devices: no sudo on controller; avoid stray sudo password prompt from ansible.cfg.
export ANSIBLE_BECOME_ASK_PASS="${ANSIBLE_BECOME_ASK_PASS:-false}"
INV="$ANSIBLE_DIR/inventory/network.yml"
PB="$PLAYBOOKS_DIR/network-isr.yml"
# Run from a non-world-writable CWD so Ansible still loads ANSIBLE_CONFIG (avoids "ignoring ansible.cfg").
cd "${HOME:-/tmp}"
exec ansible-playbook -i "$INV" "$PB" "$@"
