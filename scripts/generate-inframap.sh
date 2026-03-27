#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUT_DIR="${REPO_ROOT}/docs/artifacts/inframap"
mkdir -p "${OUT_DIR}"

if ! command -v inframap >/dev/null 2>&1; then
  echo "ERROR: inframap binary not found in PATH." >&2
  exit 1
fi

if [[ -f "${REPO_ROOT}/opentofu/terraform.tfstate" ]]; then
  inframap generate "${REPO_ROOT}/opentofu/terraform.tfstate" > "${OUT_DIR}/inframap.dot"
elif [[ -d "${REPO_ROOT}/opentofu" ]]; then
  (cd "${REPO_ROOT}/opentofu" && inframap generate . > "${OUT_DIR}/inframap.dot")
else
  echo "ERROR: no opentofu/terraform path found for inframap input." >&2
  exit 1
fi

echo "Inframap output: ${OUT_DIR}/inframap.dot"
