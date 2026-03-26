#!/usr/bin/env bash
# Apply branch ruleset for main via GitHub API. Requires: gh CLI, repo admin, repo scope.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
JSON_FILE="${SCRIPT_DIR}/github-ruleset-main.json"

if ! command -v gh >/dev/null 2>&1; then
  echo "Install GitHub CLI: https://cli.github.com/" >&2
  exit 1
fi

if [[ ! -f "${JSON_FILE}" ]]; then
  echo "Missing ${JSON_FILE}" >&2
  exit 1
fi

cd "${REPO_ROOT}"
REPO_SLUG="$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || true)"
if [[ -z "${REPO_SLUG}" ]]; then
  echo "Run from a clone with 'gh auth login' and a default repo, or set GH_REPO=owner/name." >&2
  exit 1
fi

echo "Creating ruleset on ${REPO_SLUG} ..."
gh api "repos/${REPO_SLUG}/rulesets" --method POST --input "${JSON_FILE}"
echo "Done. Verify under Settings → Rules → Rulesets."
