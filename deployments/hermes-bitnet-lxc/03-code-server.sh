#!/usr/bin/env bash
# Install code-server (VS Code in the browser) to ~/.local/bin by default.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

require_cmd curl

CODE_SERVER_INSTALL_SH="${CODE_SERVER_INSTALL_SH:-https://code-server.dev/install.sh}"

log "Installing code-server from $CODE_SERVER_INSTALL_SH ..."
if ! curl -fsSL "$CODE_SERVER_INSTALL_SH" | sh; then
  log "Install script failed (often sudo password on Alma/RHEL). As root run:"
  log "  rpm -U \"\$HOME/.cache/code-server\"/code-server-*-amd64.rpm"
  die "code-server install failed"
fi

if ! command -v code-server >/dev/null 2>&1; then
  die "code-server not on PATH after install"
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
