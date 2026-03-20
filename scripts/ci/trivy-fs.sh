#!/usr/bin/env bash
# Repo-wide Trivy filesystem scan (vuln + misconfig + secret). Uses Docker for a pinned binary.
# Usage: from repo root — bash scripts/ci/trivy-fs.sh
# Env: TRIVY_VERSION (default 0.57.1), TRIVY_SCAN_ARGS (extra CLI args)
set -euo pipefail

if [[ -n "${1:-}" ]]; then
  ROOT="$1"
else
  if ! ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"; then
    ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
  fi
fi
cd "${ROOT}"

TRIVY_VERSION="${TRIVY_VERSION:-0.57.1}"
IMAGE="aquasec/trivy:${TRIVY_VERSION}"

# Reduce noise in large IaC trees until findings are triaged; tighten over time.
SEVERITY="${TRIVY_SEVERITY:-HIGH,CRITICAL}"
export TRIVY_DISABLE_VEX_NOTICE="${TRIVY_DISABLE_VEX_NOTICE:-1}"

# shellcheck disable=SC2086
docker run --rm \
  -v "${ROOT}:/work" \
  -w /work \
  -e TRIVY_DISABLE_VEX_NOTICE \
  "${IMAGE}" \
  fs --exit-code 1 \
  --severity "${SEVERITY}" \
  --skip-dirs .git \
  --skip-dirs .venv \
  --skip-dirs .terraform \
  --skip-dirs node_modules \
  --skip-dirs .cursor \
  --skip-dirs .trivy-cache \
  ${TRIVY_SCAN_ARGS:-} \
  .

echo "trivy-fs.sh: OK"
