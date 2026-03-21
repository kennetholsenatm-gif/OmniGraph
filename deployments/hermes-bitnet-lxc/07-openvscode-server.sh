#!/usr/bin/env bash
# Install Gitpod OpenVSCode Server from official GitHub release tarball + systemd user unit hints.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

require_cmd curl
require_cmd tar
require_cmd python3

if [[ "${OPENVS_CODE_SKIP:-0}" == "1" ]]; then
  log "OPENVS_CODE_SKIP=1 — skipping OpenVSCode Server install."
  exit 0
fi

if ! command -v jq >/dev/null 2>&1; then
  die "jq required (install via 01-dnf-prereqs.sh)"
fi

TAG="${OPENVS_CODE_VERSION:-}"
if [[ -z "$TAG" ]]; then
  TAG="$(curl -fsSL https://api.github.com/repos/gitpod-io/openvscode-server/releases/latest | jq -r .tag_name)"
fi
[[ -n "$TAG" && "$TAG" != "null" ]] || die "could not resolve OpenVSCode Server release tag"

ASSET_URL="$(
  curl -fsSL "https://api.github.com/repos/gitpod-io/openvscode-server/releases/tags/${TAG}" \
    | jq -r '.assets[] | select(.name | test("linux-x64.tar.gz$")) | .browser_download_url' \
    | head -1
)"
[[ -n "$ASSET_URL" && "$ASSET_URL" != "null" ]] || die "no linux-x64.tar.gz asset for $TAG"

ensure_dir "$OPENVS_CODE_HOME"
ensure_dir "$(dirname "$OPENVS_CODE_HOME")"
TOKEN_DIR="$HOME/.config/hermes-bitnet-lxc"
ensure_dir "$TOKEN_DIR"
TOKEN_FILE="${OPENVS_CODE_TOKEN_FILE:-$TOKEN_DIR/openvscode.token}"
if [[ ! -f "$TOKEN_FILE" ]]; then
  python3 -c "import secrets; print(secrets.token_hex(24))" >"$TOKEN_FILE"
  chmod 600 "$TOKEN_FILE"
  log "Wrote connection token to $TOKEN_FILE (keep secret)."
fi

CACHE_DIR="${OPENVS_CODE_CACHE:-$HOME/.cache/openvscode-server}"
ensure_dir "$CACHE_DIR"
TAR_PATH="$CACHE_DIR/$(basename "$ASSET_URL")"

if [[ ! -f "$TAR_PATH" ]]; then
  log "Downloading $ASSET_URL ..."
  curl -fL --retry 3 --retry-delay 2 -o "$TAR_PATH" "$ASSET_URL"
fi

log "Extracting to $OPENVS_CODE_HOME ..."
ensure_dir "$OPENVS_CODE_HOME"
if [[ -d "$OPENVS_CODE_HOME" ]]; then
  find "$OPENVS_CODE_HOME" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
fi
tar -xzf "$TAR_PATH" -C "$OPENVS_CODE_HOME" --strip-components=1

BIN=""
for c in "$OPENVS_CODE_HOME/bin/openvscode-server" "$OPENVS_CODE_HOME/openvscode-server"; do
  if [[ -x "$c" ]]; then
    BIN="$c"
    break
  fi
done
if [[ -z "$BIN" ]]; then
  BIN="$(find "$OPENVS_CODE_HOME" -maxdepth 3 -type f -name 'openvscode-server' -perm -111 2>/dev/null | head -1 || true)"
fi
[[ -n "$BIN" ]] || die "openvscode-server binary not found under $OPENVS_CODE_HOME"

ensure_dir "$HOME/.local/bin"
ln -sf "$BIN" "$HOME/.local/bin/openvscode-server"
log "Symlinked openvscode-server -> $HOME/.local/bin/openvscode-server"

CODE_SERVER_WORKSPACE="${OPENVS_CODE_WORKSPACE:-}"
if [[ -z "$CODE_SERVER_WORKSPACE" ]]; then
  if [[ -e "${QMINI_DIR:-}/.git" ]]; then
    CODE_SERVER_WORKSPACE="$QMINI_DIR"
  else
    CODE_SERVER_WORKSPACE="$SRC_ROOT"
  fi
fi

log "systemd: copy $ROOT_DIR/systemd/openvscode-server.service.example to ~/.config/systemd/user/openvscode-server.service"
log "  Set OPENVS_CODE_HOME / token file paths if non-default."
log "  systemctl --user daemon-reload && systemctl --user enable --now openvscode-server"
log "Browser: append token query param (root URL alone returns Forbidden):"
log "  http://127.0.0.1:$OPENVS_CODE_PORT/?tkn=<contents of $TOKEN_FILE>"
log "Manual run:"
log "  $BIN --host 0.0.0.0 --port $OPENVS_CODE_PORT --connection-token-file \"$TOKEN_FILE\" \"$CODE_SERVER_WORKSPACE\""
log "07-openvscode-server: done"
