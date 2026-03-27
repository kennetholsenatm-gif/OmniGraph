#!/usr/bin/env bash
# Create Gitea user kbolsen (if missing), empty repo devsecops-pipeline, then print git commands.
# Run on the WSL/Alma host that has incus + the devsecops-gitea LXC.
#
# Usage:
#   export GITEA_ADMIN_PASSWORD='your-admin-password'
#   ./scripts/setup-kbolsen-devsecops-repo.sh
#
# Optional env:
#   GITEA_LXC=devsecops-gitea GITEA_CONTAINER=devsecops-gitea GITEA_HTTP=http://127.0.0.1:3000
#   KBOLSEN_PASSWORD='...'   # if unset, a random password is generated and printed once

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INNER="${SCRIPT_DIR}/gitea-inner-create-kbolsen-repo.sh"

GITEA_LXC="${GITEA_LXC:-devsecops-gitea}"
GITEA_CONTAINER="${GITEA_CONTAINER:-devsecops-gitea}"
GITEA_HTTP="${GITEA_HTTP:-http://127.0.0.1:3000}"
OWNER="${OWNER:-kbolsen}"
REPO="${REPO:-devsecops-pipeline}"
EMAIL="${KBOLSEN_EMAIL:-${OWNER}@local.dev}"

GITEA_ADMIN_PASSWORD="${1:-${GITEA_ADMIN_PASSWORD:-}}"
if [[ -z "${GITEA_ADMIN_PASSWORD}" ]]; then
  echo "ERROR: Set GITEA_ADMIN_PASSWORD or pass the admin password as the first argument." >&2
  exit 1
fi

if [[ ! -f "${INNER}" ]]; then
  echo "ERROR: missing ${INNER}" >&2
  exit 1
fi

KBOLSEN_PASSWORD="${KBOLSEN_PASSWORD:-}"
if [[ -z "${KBOLSEN_PASSWORD}" ]]; then
  KBOLSEN_PASSWORD="$(tr -dc 'A-Za-z0-9@#%+=._-' </dev/urandom | head -c 24)"
  echo "Generated KBOLSEN_PASSWORD (save this): ${KBOLSEN_PASSWORD}"
fi

cat "${INNER}" | incus exec "${GITEA_LXC}" -- \
  env \
    ADMIN_PW="${GITEA_ADMIN_PASSWORD}" \
    KB_PW="${KBOLSEN_PASSWORD}" \
    OWNER="${OWNER}" \
    REPO="${REPO}" \
    EMAIL="${EMAIL}" \
  docker exec -i \
    -e ADMIN_PW -e KB_PW -e OWNER -e REPO -e EMAIL \
    --user git "${GITEA_CONTAINER}" \
    bash -s

cat <<EOF

Open (after push): ${GITEA_HTTP}/${OWNER}/${REPO}

From your devsecops-pipeline clone (example path):

  cd /mnt/c/GiTeaRepos/devsecops-pipeline
  git remote remove gitea 2>/dev/null || true
  git remote add gitea http://${OWNER}@127.0.0.1:3000/${OWNER}/${REPO}.git
  git push -u gitea HEAD:main

When prompted, use the ${OWNER} password (KBOLSEN_PASSWORD printed above).
If your default branch is master:  git push -u gitea HEAD:master

SSH (optional): git remote add gitea ssh://git@127.0.0.1:2222/${OWNER}/${REPO}.git
EOF
