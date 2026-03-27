#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
OUT_DIR="${REPO_ROOT}/docs/artifacts/kics"
mkdir -p "${OUT_DIR}"

if ! command -v kics >/dev/null 2>&1; then
  echo "ERROR: kics binary not found in PATH." >&2
  exit 1
fi

TARGETS=()
for p in "${REPO_ROOT}/ansible" "${REPO_ROOT}/deployments" "${REPO_ROOT}/docker-compose" "${REPO_ROOT}/opentofu"; do
  [[ -d "$p" ]] && TARGETS+=("$p")
done

if [[ ${#TARGETS[@]} -eq 0 ]]; then
  echo "ERROR: no scan targets found." >&2
  exit 1
fi

IFS=, read -r -a _ <<< "${TARGETS[*]}"
kics scan -p "${TARGETS[*]// /,}" --report-formats json,sarif --output-path "${OUT_DIR}"
echo "KICS reports written to ${OUT_DIR}"
