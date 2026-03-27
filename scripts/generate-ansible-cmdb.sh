#!/usr/bin/env bash
set -euo pipefail

INVENTORY="${1:-ansible/inventory/opennebula-hybrid.example.yml}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUT_DIR="${REPO_ROOT}/docs/artifacts/ansible-cmdb"
mkdir -p "${OUT_DIR}"

if ! command -v ansible-cmdb >/dev/null 2>&1; then
  echo "ERROR: ansible-cmdb not found in PATH." >&2
  exit 1
fi

if [[ ! -f "${REPO_ROOT}/${INVENTORY}" ]]; then
  echo "ERROR: Inventory not found: ${REPO_ROOT}/${INVENTORY}" >&2
  exit 1
fi

ansible-cmdb -i "${REPO_ROOT}/${INVENTORY}" --format html_fancy --output-dir "${OUT_DIR}" >/dev/null
echo "Ansible-CMDB report generated in ${OUT_DIR}"
