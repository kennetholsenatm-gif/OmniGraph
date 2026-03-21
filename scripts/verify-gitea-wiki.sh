#!/usr/bin/env bash
# List Gitea wiki pages for a repository (GET .../wiki/pages).
#
# Usage:
#   ./scripts/verify-gitea-wiki.sh --url http://localhost:3000 --owner kbolsen \
#     --repo devsecops-pipeline --token YOUR_PAT

set -euo pipefail

GITEA_URL=""
OWNER=""
REPO=""
TOKEN=""

usage() {
  echo "Usage: $0 --url <gitea-base-url> --owner <owner> --repo <repo> --token <pat>" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --url) GITEA_URL="${2:-}"; shift 2 ;;
    --owner) OWNER="${2:-}"; shift 2 ;;
    --repo) REPO="${2:-}"; shift 2 ;;
    --token) TOKEN="${2:-}"; shift 2 ;;
    -h|--help) usage ;;
    *) echo "Unknown option: $1" >&2; usage ;;
  esac
done

if [[ -z "$GITEA_URL" || -z "$OWNER" || -z "$REPO" || -z "$TOKEN" ]]; then
  echo "Error: --url, --owner, --repo, and --token are required." >&2
  usage
fi

GITEA_URL="${GITEA_URL%/}"
OWNER_ENC="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$OWNER")"
REPO_ENC="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$REPO")"
URL="${GITEA_URL}/api/v1/repos/${OWNER_ENC}/${REPO_ENC}/wiki/pages"

out="$(mktemp)"
code="$(curl -sS -o "$out" -w "%{http_code}" -X GET "$URL" \
  -H "Authorization: token ${TOKEN}" \
  -H "Accept: application/json")"

if [[ "$code" != "200" ]]; then
  echo "Error: GET wiki/pages failed (HTTP $code)" >&2
  cat "$out" >&2 || true
  rm -f "$out"
  exit 1
fi

count="$(python3 -c "import json,sys; print(len(json.load(open(sys.argv[1]))))" "$out")"
echo "OK: ${count} wiki page(s) at ${OWNER}/${REPO}"
python3 -c "
import json, sys
for p in json.load(open(sys.argv[1])):
    print('  -', p.get('title', ''))
" "$out"
rm -f "$out"
exit 0
