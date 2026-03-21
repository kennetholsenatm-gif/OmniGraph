#!/usr/bin/env bash
# Publish markdown files from wiki/gitea-pages to a Gitea repo wiki (REST API).
# Gitea 1.22+: POST .../wiki/new and PATCH .../wiki/page/{pageName} with JSON
# { "title", "content_base64", "message" }.
#
# Usage:
#   ./scripts/publish-gitea-wiki-pages.sh --url http://localhost:3000 --owner kbolsen \
#     --repo devsecops-pipeline --token YOUR_PAT
#   ./scripts/publish-gitea-wiki-pages.sh --url ... --owner ... --repo ... --dry-run
#
# Requires: curl, python3.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PAGES_DIR="$REPO_ROOT/wiki/gitea-pages"
MESSAGE="Publish wiki pages from devsecops-pipeline repo"
DRY_RUN=0
GITEA_URL=""
OWNER=""
REPO=""
TOKEN=""

usage() {
  echo "Usage: $0 --url <gitea-base-url> --owner <owner> --repo <repo> --token <pat> [options]" >&2
  echo "  --pages-dir <dir>   default: <repo>/wiki/gitea-pages" >&2
  echo "  --message <text>    wiki commit message" >&2
  echo "  --dry-run           list .md files only (no HTTP, no token)" >&2
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --url) GITEA_URL="${2:-}"; shift 2 ;;
    --owner) OWNER="${2:-}"; shift 2 ;;
    --repo) REPO="${2:-}"; shift 2 ;;
    --token) TOKEN="${2:-}"; shift 2 ;;
    --pages-dir) PAGES_DIR="${2:-}"; shift 2 ;;
    --message) MESSAGE="${2:-}"; shift 2 ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage ;;
    *) echo "Unknown option: $1" >&2; usage ;;
  esac
done

if [[ -z "$GITEA_URL" || -z "$OWNER" || -z "$REPO" ]]; then
  echo "Error: --url, --owner, and --repo are required." >&2
  usage
fi
if [[ "$DRY_RUN" -eq 0 && -z "$TOKEN" ]]; then
  echo "Error: --token is required unless --dry-run." >&2
  exit 1
fi

GITEA_URL="${GITEA_URL%/}"
OWNER_ENC="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$OWNER")"
REPO_ENC="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$REPO")"
API_BASE="${GITEA_URL}/api/v1/repos/${OWNER_ENC}/${REPO_ENC}"

if [[ ! -d "$PAGES_DIR" ]]; then
  echo "Error: pages directory not found: $PAGES_DIR" >&2
  exit 1
fi

wiki_list_json() {
  local out code
  out="$(mktemp)"
  code="$(curl -sS -o "$out" -w "%{http_code}" -X GET "${API_BASE}/wiki/pages" \
    -H "Authorization: token ${TOKEN}" \
    -H "Accept: application/json")"
  if [[ "$code" == "200" ]]; then
    cat "$out"
  else
    echo "[]"
    if [[ "$code" == "404" ]]; then
      echo "Warning: wiki list returned 404 (wiki may be uninitialized). New pages will use POST only." >&2
    else
      echo "Warning: wiki list failed (HTTP $code). Continuing with empty list." >&2
    fi
  fi
  rm -f "$out"
}

slug_for_title() {
  local pages_json="$1" title="$2"
  printf '%s' "$pages_json" | python3 -c "
import json, sys
pages = json.load(sys.stdin)
want = sys.argv[1]
for p in pages:
    if p.get('title') == want:
        print(p.get('slug') or p.get('name') or p.get('title') or '')
        raise SystemExit(0)
" "$title"
}

build_body() {
  local title="$1" file="$2"
  python3 -c "
import json, base64, sys
title, path, msg = sys.argv[1], sys.argv[2], sys.argv[3]
with open(path, 'rb') as f:
    raw = f.read()
b64 = base64.b64encode(raw).decode('ascii')
print(json.dumps({'title': title, 'content_base64': b64, 'message': msg}, ensure_ascii=False))
" "$title" "$file" "$MESSAGE"
}

post_page() {
  local body="$1" code out
  out="$(mktemp)"
  code="$(curl -sS -o "$out" -w "%{http_code}" -X POST "${API_BASE}/wiki/new" \
    -H "Authorization: token ${TOKEN}" \
    -H "Content-Type: application/json; charset=utf-8" \
    --data-binary "$body")"
  rm -f "$out"
  [[ "$code" == "201" ]]
}

patch_page() {
  local slug="$1" body="$2" enc
  enc="$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$slug")"
  curl -sS -f -X PATCH "${API_BASE}/wiki/page/${enc}" \
    -H "Authorization: token ${TOKEN}" \
    -H "Content-Type: application/json; charset=utf-8" \
    --data-binary "$body" >/dev/null
}

MD_FILES=()
while IFS= read -r line; do
  [[ -n "$line" ]] && MD_FILES+=("$line")
done < <(find "$PAGES_DIR" -maxdepth 1 -type f -name '*.md' | LC_ALL=C sort)

if [[ ${#MD_FILES[@]} -eq 0 ]]; then
  echo "Warning: no .md files in $PAGES_DIR" >&2
  exit 0
fi

if [[ "$DRY_RUN" -eq 1 ]]; then
  for f in "${MD_FILES[@]}"; do
    base="$(basename "$f")"
    title="${base%.md}"
    echo "[dry-run] would publish: $title ($f)"
  done
  echo "Done. ${#MD_FILES[@]} file(s) under $PAGES_DIR"
  exit 0
fi

EXISTING_JSON="$(wiki_list_json)"

for f in "${MD_FILES[@]}"; do
  base="$(basename "$f")"
  title="${base%.md}"
  body="$(build_body "$title" "$f")"
  slug="$(slug_for_title "$EXISTING_JSON" "$title" || true)"

  if post_page "$body"; then
    echo "Created wiki page: $title"
    EXISTING_JSON="$(wiki_list_json)"
    continue
  fi

  if [[ -z "$slug" ]]; then
    EXISTING_JSON="$(wiki_list_json)"
    slug="$(slug_for_title "$EXISTING_JSON" "$title" || true)"
  fi
  if [[ -z "$slug" ]]; then
    echo "Error: could not create or resolve slug for page: $title" >&2
    exit 1
  fi

  patch_page "$slug" "$body"
  echo "Updated wiki page: $title (slug: $slug)"
  EXISTING_JSON="$(wiki_list_json)"
done

echo "Done. ${#MD_FILES[@]} file(s) processed from $PAGES_DIR"
