#!/usr/bin/env bash
# Apply Semaphore desired state from ansible/files/semaphore-manifest.yml (bootstraps from example if missing).
# Usage: set env vars, then from repo root — bash scripts/ci/semaphore-sync.sh
#
# Required:
#   SEMAPHORE_API_TOKEN
#   SEMAPHORE_URL          (e.g. http://127.0.0.1:3001 — must match how the runner reaches Semaphore)
#   SEMAPHORE_GIT_URL      (clone URL Semaphore uses for this repo, e.g. https://gitea.example/org/devsecops-pipeline.git)
# Optional:
#   SEMAPHORE_GIT_BRANCH   (default: main)
#
set -euo pipefail

if [[ -n "${1:-}" ]]; then
  ROOT="$1"
else
  if ! ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"; then
    ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
  fi
fi
cd "${ROOT}"

export ANSIBLE_CONFIG="${ROOT}/ansible/ansible.cfg"

: "${SEMAPHORE_API_TOKEN:?Set SEMAPHORE_API_TOKEN}"
: "${SEMAPHORE_URL:?Set SEMAPHORE_URL}"
: "${SEMAPHORE_GIT_URL:?Set SEMAPHORE_GIT_URL}"

GIT_BRANCH="${SEMAPHORE_GIT_BRANCH:-main}"
if [[ -z "${GIT_BRANCH}" ]]; then
  GIT_BRANCH=main
fi

ansible-playbook -i localhost, -c local ansible/playbooks/sync-semaphore-from-manifest.yml \
  -e "semaphore_api_token=${SEMAPHORE_API_TOKEN}" \
  -e "semaphore_url=${SEMAPHORE_URL}" \
  -e "semaphore_git_url=${SEMAPHORE_GIT_URL}" \
  -e "semaphore_git_branch=${GIT_BRANCH}"
