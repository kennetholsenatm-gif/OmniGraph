#!/usr/bin/env bash
# Run *inside* the Gitea Docker container (paths: curl -> 127.0.0.1:3000).
# Expects env: ADMIN_PW KB_PW OWNER REPO EMAIL

set -euo pipefail

: "${ADMIN_PW:?ADMIN_PW missing}"
: "${KB_PW:?KB_PW missing}"
: "${OWNER:?OWNER missing}"
: "${REPO:?REPO missing}"
: "${EMAIL:?EMAIL missing}"

TOKEN_JSON="$(curl -fsS -u "admin:${ADMIN_PW}" -X POST "127.0.0.1:3000/api/v1/users/admin/tokens" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"setup-kbolsen-repo\",\"scopes\":[\"write:admin\"]}")"
TOKEN="$(printf "%s" "${TOKEN_JSON}" | sed -n "s/.*\"sha1\":\"\\([^\"]*\\)\".*/\\1/p")"
if [[ -z "${TOKEN}" ]]; then
  echo "ERROR: could not create admin token. Response: ${TOKEN_JSON}" >&2
  exit 1
fi

HTTP_USER="$(curl -sS -o /tmp/gu.out -w "%{http_code}" -X POST "127.0.0.1:3000/api/v1/admin/users" \
  -H "Authorization: token ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${OWNER}\",\"email\":\"${EMAIL}\",\"password\":\"${KB_PW}\",\"must_change_password\":false}")"
if [[ "${HTTP_USER}" == "201" ]]; then
  echo "User ${OWNER} created."
else
  if grep -qiE 'already|exist|422|used' /tmp/gu.out 2>/dev/null; then
    echo "User ${OWNER} already exists (OK)."
  else
    echo "WARN: create user HTTP ${HTTP_USER}" >&2
    cat /tmp/gu.out >&2 || true
  fi
fi

HTTP_REPO="$(curl -sS -o /tmp/gr.out -w "%{http_code}" -X POST "127.0.0.1:3000/api/v1/admin/users/${OWNER}/repos" \
  -H "Authorization: token ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"${REPO}\",\"private\":false,\"auto_init\":false,\"default_branch\":\"main\"}")"
if [[ "${HTTP_REPO}" == "201" ]]; then
  echo "Repository ${OWNER}/${REPO} created."
elif [[ "${HTTP_REPO}" == "409" ]]; then
  echo "Repository ${OWNER}/${REPO} already exists (OK)."
else
  echo "ERROR: create repo HTTP ${HTTP_REPO}" >&2
  cat /tmp/gr.out >&2 || true
  exit 1
fi
