#!/usr/bin/env bash
# Run ansible-playbook from this directory without relying on CWD ansible.cfg.
# On WSL + /mnt/c/, the playbooks directory is world-writable; Ansible ignores
# ansible.cfg there — set ANSIBLE_CONFIG + ANSIBLE_ROLES_PATH explicitly.
set -eu
PLAYBOOKS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
ANSIBLE_DIR="$(cd "$PLAYBOOKS_DIR/.." && pwd)"
export ANSIBLE_CONFIG="$ANSIBLE_DIR/ansible.cfg"
export ANSIBLE_ROLES_PATH="$ANSIBLE_DIR/roles"
# Hypervisor tasks use become; without this, sudo waits for a TTY and times out. Override: ANSIBLE_BECOME_ASK_PASS=false for NOPASSWD sudo.
export ANSIBLE_BECOME_ASK_PASS="${ANSIBLE_BECOME_ASK_PASS:-true}"

cd "$PLAYBOOKS_DIR"
exec ansible-playbook "$@"
