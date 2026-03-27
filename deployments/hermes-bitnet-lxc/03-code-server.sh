#!/usr/bin/env bash
# Install code-server (VS Code in the browser) to ~/.local/bin by default.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

require_cmd curl

if [[ "${CODE_SERVER_SKIP:-0}" == "1" ]]; then
  log "CODE_SERVER_SKIP=1 — skipping Coder code-server install (use OpenVSCode via 07-openvscode-server.sh)."
  exit 0
fi

CODE_SERVER_INSTALL_SH="${CODE_SERVER_INSTALL_SH:-https://code-server.dev/install.sh}"
# Alma/RHEL map to "fedora" in code-server's installer → it runs `sudo rpm -U`, which fails
# without passwordless sudo. Standalone unpacks to ~/.local (no root), same idea as 07-openvscode-server.
CODE_SERVER_INSTALL_METHOD="${CODE_SERVER_INSTALL_METHOD:-standalone}"

INSTALL_ARGS=(--method "$CODE_SERVER_INSTALL_METHOD")
if [[ -n "${CODE_SERVER_INSTALL_VERSION:-}" ]]; then
  INSTALL_ARGS+=(--version "$CODE_SERVER_INSTALL_VERSION")
fi

log "Installing code-server from $CODE_SERVER_INSTALL_SH (${INSTALL_ARGS[*]}) ..."
if ! curl -fsSL "$CODE_SERVER_INSTALL_SH" | sh -s -- "${INSTALL_ARGS[@]}"; then
  log "Install failed. Tips:"
  log "  - Default is --method standalone (no sudo). Re-run or set CODE_SERVER_INSTALL_METHOD=standalone"
  log "  - For system RPM: sudo rpm -U \"\$HOME/.cache/code-server\"/code-server-<ver>-amd64.rpm (see cache dir for exact name)"
  die "code-server install failed"
fi

if [[ -x "$HOME/.local/bin/code-server" ]]; then
  :
elif command -v code-server >/dev/null 2>&1; then
  :
else
  die "code-server not found after install (expected ~/.local/bin/code-server with standalone method)"
fi

if ! command -v code-server >/dev/null 2>&1; then
  log "Add to PATH if needed: export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

ensure_dir "$SRC_ROOT"

# Prefer local qminiwasm-core when present so the editor/agent opens that tree by default.
CODE_SERVER_WORKSPACE="${CODE_SERVER_WORKSPACE:-}"
if [[ -z "$CODE_SERVER_WORKSPACE" ]]; then
  if [[ -e "${QMINI_DIR:-}/.git" ]]; then
    CODE_SERVER_WORKSPACE="$QMINI_DIR"
  else
    CODE_SERVER_WORKSPACE="$SRC_ROOT"
  fi
fi

log "code-server: bind to 127.0.0.1; default workspace $CODE_SERVER_WORKSPACE (qminiwasm-core when local clone exists)"
log "  code-server --bind-addr 127.0.0.1:${CODE_SERVER_PORT} \"$CODE_SERVER_WORKSPACE\""
log "  Add BitNet/Hermes from $SRC_ROOT via File → Open Folder if needed."
log "SSH tunnel from your workstation: ssh -L ${CODE_SERVER_PORT}:127.0.0.1:${CODE_SERVER_PORT} user@container"
log "03-code-server: done"
