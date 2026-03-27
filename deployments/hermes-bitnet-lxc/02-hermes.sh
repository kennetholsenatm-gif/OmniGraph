#!/usr/bin/env bash
# Install Hermes Agent via official installer (interactive TTY recommended for first run).
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

require_cmd curl
require_cmd git

HERMES_INSTALLER_URL="${HERMES_INSTALLER_URL:-https://raw.githubusercontent.com/NousResearch/hermes-agent/main/scripts/install.sh}"
SKIP_SETUP="${HERMES_SKIP_SETUP:-1}"
HERMES_INSTALL_LOG="${HERMES_INSTALL_LOG:-$HOME/log/hermes-install.log}"
ensure_dir "$(dirname "$HERMES_INSTALL_LOG")"

if [[ "$SKIP_SETUP" == "1" ]]; then
  log "Running Hermes installer with --skip-setup (set HERMES_SKIP_SETUP=0 for interactive wizard)."
  log "Logging installer stream to $HERMES_INSTALL_LOG"
  curl -fsSL "$HERMES_INSTALLER_URL" | tee "$HERMES_INSTALL_LOG" | bash -s -- --skip-setup
else
  log "Running Hermes installer (interactive)."
  log "Logging installer stream to $HERMES_INSTALL_LOG"
  curl -fsSL "$HERMES_INSTALLER_URL" | tee "$HERMES_INSTALL_LOG" | bash
fi

if [[ "$(id -u)" -eq 0 ]] && [[ -w /var/log ]]; then
  cp -a "$HERMES_INSTALL_LOG" /var/log/hermes-install.log 2>/dev/null || true
fi

if ! echo "${PATH:-}" | tr ':' '\n' | grep -qx "$HOME/.local/bin"; then
  log "Add to shell rc if needed: export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

if [[ -x "$HOME/.local/bin/hermes" ]]; then
  "$HOME/.local/bin/hermes" version || true
  log "Primary qminiwasm-core tree for tools/agent work: $QMINI_DIR"
  log "(Windows: C:\\GitHub\\LLM_Pract\\qminiwasm-core — use WSL path $QMINI_LOCAL_DEFAULT when available.)"
  log "Next: configure model provider — after BitNet llama-server is up:"
  log "  export OPENAI_BASE_URL=http://127.0.0.1:${BITNET_PORT}/v1"
  log "  export OPENAI_API_KEY=local"
  log "  hermes model   # or hermes setup"
  log "  hermes doctor"
else
  log "warn: ~/.local/bin/hermes not found; check installer output"
fi

log "02-hermes: done"
