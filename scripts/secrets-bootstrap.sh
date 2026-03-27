#!/usr/bin/env bash
set -euo pipefail

DEPLOY_SCRIPT="/mnt/c/GiTeaRepos/Deploy/scripts/secrets-bootstrap.sh"
if [[ ! -f "${DEPLOY_SCRIPT}" ]]; then
  echo "Moved script not found: ${DEPLOY_SCRIPT}" >&2
  exit 1
fi

echo "secrets-bootstrap.sh moved to Deploy repo -> ${DEPLOY_SCRIPT}"
exec "${DEPLOY_SCRIPT}" "$@"
