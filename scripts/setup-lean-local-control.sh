#!/usr/bin/env bash
set -euo pipefail

SKIP_SEMAPHORE="${SKIP_SEMAPHORE:-false}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${REPO_ROOT}"
echo "Lean local control-plane setup in ${REPO_ROOT}"

if [[ "${SKIP_SEMAPHORE}" != "true" ]]; then
  echo "Semaphore: provision via Incus (Ansible), not Docker — run:"
  echo "  ${REPO_ROOT}/scripts/start-semaphore.sh"
fi

echo "Installing pre-commit hooks (if available)..."
if command -v pre-commit >/dev/null 2>&1; then
  pre-commit install
else
  echo "WARN: pre-commit not found in PATH."
fi

echo "Installing Ansible collections (if available)..."
if command -v ansible-galaxy >/dev/null 2>&1; then
  (cd "${REPO_ROOT}/ansible" && ansible-galaxy collection install -r collections/requirements.yml)
else
  echo "WARN: ansible-galaxy not found in PATH."
fi

echo "Done. Open Semaphore at http://127.0.0.1:3001"
