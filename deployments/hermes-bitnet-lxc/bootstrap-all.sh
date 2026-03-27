#!/usr/bin/env bash
# Run all bootstrap steps in order inside AlmaLinux (or RHEL-family) LXC.
# Usage:
#   ./bootstrap-all.sh              # full run (BitNet model download can be huge)
#   BITNET_SKIP_MODEL_DOWNLOAD=1 ./bootstrap-all.sh   # deps + Hermes + editors + BitNet build only
#   HERMES_BITNET_RUN_INCUS_PREFLIGHT=0 ./bootstrap-all.sh   # skip 00 on noisy hosts
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"
export HERMES_BITNET_SRC_ROOT="${HERMES_BITNET_SRC_ROOT:-$HOME/src}"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

if [[ "${HERMES_BITNET_RUN_INCUS_PREFLIGHT:-1}" != "0" ]]; then
  bash "$ROOT_DIR/00-incus-preflight.sh" || true
fi

bash "$ROOT_DIR/01-dnf-prereqs.sh"
bash "$ROOT_DIR/02-hermes.sh"
bash "$ROOT_DIR/03-code-server.sh"
bash "$ROOT_DIR/07-openvscode-server.sh"
bash "$ROOT_DIR/04-bitnet-build.sh"
bash "$ROOT_DIR/05-qminiwasm.sh"

if [[ "${HERMES_BITNET_RUN_PLAYWRIGHT:-0}" == "1" ]]; then
  bash "$ROOT_DIR/06-playwright-chromium.sh"
fi

bash "$ROOT_DIR/08-hermes-bitnet-config.sh"
bash "$ROOT_DIR/09-openvscode-hermes-acp.sh"

mkdir -p "$HOME/.local/bin"
ln -sf "$ROOT_DIR/run-bitnet-server.sh" "$HOME/.local/bin/run-bitnet-server.sh"
log "Symlinked run-bitnet-server.sh -> ~/.local/bin/run-bitnet-server.sh"

echo ""
echo "=== Bootstrap finished ==="
echo "1) Model weights: if skipped, run 04-bitnet-build.sh with BITNET_SKIP_MODEL_DOWNLOAD unset."
echo "2) BitNet API:     $ROOT_DIR/run-bitnet-server.sh  (systemd: systemd/bitnet-llama-server.service.example)"
echo "3) Hermes:         export PATH=\"\$HOME/.local/bin:\$PATH\"  # then: hermes doctor"
echo "                   (08 patches run_agent + .env for BitNet: HERMES_CHAT_COMPLETIONS_NO_TOOLS=1)"
echo "                   (08-hermes-bitnet-config.sh wires ~/.hermes when GGUF exists)"
echo "4) OpenVSCode:     token: ~/.config/hermes-bitnet-lxc/openvscode.token"
echo "                   systemd: systemd/openvscode-server.service.example"
echo "                   Hermes ACP: OpenRouter coding agent + BitNet chat (09-openvscode-hermes-acp.sh)"
echo "5) Coder VS Code:  code-server --bind-addr 127.0.0.1:${CODE_SERVER_PORT:-8443} \"$QMINI_DIR\"  (unless CODE_SERVER_SKIP=1)"
echo "    (Windows path: C:\\GitHub\\LLM_Pract\\qminiwasm-core → WSL/bind-mount: $QMINI_LOCAL_DEFAULT)"
echo "6) qminiwasm venv: source \"$QMINI_VENV/bin/activate\""
echo "See README.md and incus/README.md for Incus / WSL / RAM notes."
