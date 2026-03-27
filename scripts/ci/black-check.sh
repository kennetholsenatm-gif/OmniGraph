#!/usr/bin/env bash
# Black --check for Python under scripts/ (expand paths in pyproject.toml as the tree grows).
# Usage: from repo root — bash scripts/ci/black-check.sh
set -euo pipefail

if [[ -n "${1:-}" ]]; then
  ROOT="$1"
else
  if ! ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"; then
    ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
  fi
fi
cd "${ROOT}"

BLACK_VERSION="${BLACK_VERSION:-24.10.0}"

docker run --rm \
  -v "${ROOT}:/workspace" \
  -w /workspace \
  python:3.12-slim-bookworm \
  bash -c "
    set -e
    pip install --no-cache-dir 'black==${BLACK_VERSION}'
    black --check scripts/
  "

echo "black-check.sh: OK"
