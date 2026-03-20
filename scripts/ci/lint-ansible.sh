#!/usr/bin/env bash
# Run the same Ansible/YAML checks as Gitea Actions CI (AlmaLinux 9 + Docker).
# Usage: from repo root — bash scripts/ci/lint-ansible.sh
# Requires: Docker, git (optional; falls back to script location for repo root).
set -euo pipefail

ROOT="${1:-}"
if [[ -z "${ROOT}" ]]; then
  if ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"; then
    :
  else
    ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
  fi
fi
cd "${ROOT}"

ALMALINUX_CI_IMAGE="${ALMALINUX_CI_IMAGE:-almalinux:9.7-20260129}"

docker run --rm \
  -v "${ROOT}:/workspace" \
  -w /workspace \
  -e PIP_NO_CACHE_DIR=1 \
  -e ANSIBLE_CONFIG=/workspace/ansible/ansible.cfg \
  "${ALMALINUX_CI_IMAGE}" \
  bash -c '
set -e
dnf install -y python3.11 python3.11-pip git-core 2>/dev/null || true
python3.11 -m venv /tmp/ci-venv
# shellcheck disable=SC1091
. /tmp/ci-venv/bin/activate
pip install --upgrade pip
# Align with .pre-commit-config.yaml (ansible-lint v24.12.2, yamllint v1.37.x)
pip install "ansible-core>=2.15,<2.18" "ansible-lint==24.12.2" "yamllint>=1.37.1,<2"
ansible-galaxy collection install -r ansible/collections/requirements.yml
ansible-lint ansible/
YAMLLINT_DIRS=(ansible)
if [[ -d opentofu ]]; then
  YAMLLINT_DIRS+=(opentofu)
fi
yamllint -d "{extends: default, rules: {line-length: {max: 180}}}" "${YAMLLINT_DIRS[@]}"
echo "lint-ansible.sh: OK"
'
