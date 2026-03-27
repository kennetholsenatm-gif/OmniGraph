#!/usr/bin/env bash
# Gitleaks secret scan (matches .gitleaks.toml). Uses official image for a pinned binary.
# Usage: from repo root — bash scripts/ci/gitleaks.sh
set -euo pipefail

if [[ -n "${1:-}" ]]; then
  ROOT="$1"
else
  if ! ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"; then
    ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
  fi
fi
cd "${ROOT}"

# Default image: Docker Hub (override with GITLEAKS_IMAGE)
GITLEAKS_IMAGE="${GITLEAKS_IMAGE:-zricethezav/gitleaks:v8.21.2}"

CONFIG_ARGS=()
if [[ -f "${ROOT}/.gitleaks.toml" ]]; then
  CONFIG_ARGS=(--config /repo/.gitleaks.toml)
fi

docker run --rm \
  -v "${ROOT}:/repo" \
  -w /repo \
  "${GITLEAKS_IMAGE}" \
  detect \
  --source /repo \
  "${CONFIG_ARGS[@]}" \
  --verbose \
  --no-banner

echo "gitleaks.sh: OK"
