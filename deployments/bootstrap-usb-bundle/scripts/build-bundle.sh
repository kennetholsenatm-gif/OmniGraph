#!/usr/bin/env bash
# Build an offline bootstrap bundle: repo snapshot + vendored Ansible collections.
# Run on a connected Linux machine with: git, rsync, ansible-galaxy in PATH.
# Usage: from this directory — ./build-bundle.sh [optional-output-parent]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUNDLE_ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${BUNDLE_ROOT_DIR}/../.." && pwd)"

OUT_PARENT="${1:-${BUNDLE_ROOT_DIR}/out}"
STAMP="$(date +%Y%m%d)"
OUT="${OUT_PARENT}/devsecops-bootstrap-${STAMP}"

if [[ ! -d "${REPO_ROOT}/ansible" ]] || [[ ! -f "${REPO_ROOT}/ansible/collections/requirements.yml" ]]; then
  echo "ERROR: Could not find repo root (expected ansible/collections/requirements.yml). REPO_ROOT=${REPO_ROOT}" >&2
  exit 1
fi

mkdir -p "${OUT}/collections" "${OUT}/images"

echo "==> Syncing repo -> ${OUT}/repo (excluding heavy/ephemeral paths)"
rsync -a --delete \
  --exclude '.git/' \
  --exclude '.cursor/' \
  --exclude '.dev/' \
  --exclude 'out/' \
  --exclude '**/packer/output-*/' \
  --exclude '**/.terraform/' \
  --exclude '**/*.qcow2' \
  --exclude 'debug-*.log' \
  "${REPO_ROOT}/" "${OUT}/repo/"

echo "==> Installing Ansible collections into ${OUT}/collections (offline-ready tree)"
ansible-galaxy collection install \
  -r "${OUT}/repo/ansible/collections/requirements.yml" \
  -p "${OUT}/collections" \
  --force

cp "${SCRIPT_DIR}/bootstrap-on-target.sh" "${OUT}/bootstrap-on-target.sh"
chmod +x "${OUT}/bootstrap-on-target.sh"

{
  echo "built_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "repo_root_source=${REPO_ROOT}"
  if command -v git >/dev/null 2>&1 && git -C "${REPO_ROOT}" rev-parse HEAD >/dev/null 2>&1; then
    echo "git_sha=$(git -C "${REPO_ROOT}" rev-parse HEAD)"
  else
    echo "git_sha=unknown"
  fi
  echo ""
  echo "Optional: place docker save tarballs under images/ and run bootstrap-on-target.sh on the target."
} > "${OUT}/MANIFEST.txt"

echo "==> Done: ${OUT}"
echo "    Copy to USB or: tar -C $(dirname "${OUT}") -czf devsecops-bootstrap-${STAMP}.tgz $(basename "${OUT}")"
