#!/usr/bin/env bash
# Run all bootstrap steps in order inside AlmaLinux (or RHEL-family) LXC.
# Usage:
#   ./bootstrap-all.sh              # full run (BitNet model download can be huge)
#   BITNET_SKIP_MODEL_DOWNLOAD=1 ./bootstrap-all.sh   # deps + Hermes + code-server + BitNet build only
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

export HERMES_BITNET_SRC_ROOT="${HERMES_BITNET_SRC_ROOT:-$HOME/src}"

bash "$ROOT_DIR/01-dnf-prereqs.sh"
bash "$ROOT_DIR/02-hermes.sh"
bash "$ROOT_DIR/03-code-server.sh"
bash "$ROOT_DIR/04-bitnet-build.sh"
bash "$ROOT_DIR/05-qminiwasm.sh"

mkdir -p "$HOME/.local/bin"
ln -sf "$ROOT_DIR/run-bitnet-server.sh" "$HOME/.local/bin/run-bitnet-server.sh"
log "Symlinked run-bitnet-server.sh -> ~/.local/bin/run-bitnet-server.sh"

echo ""
echo "=== Bootstrap finished ==="
echo "1) Model weights: if skipped, run 04-bitnet-build.sh with BITNET_SKIP_MODEL_DOWNLOAD unset."
echo "2) BitNet API:     $ROOT_DIR/run-bitnet-server.sh"
echo "3) Hermes:         export PATH=\"\$HOME/.local/bin:\$PATH\""
echo "                   export OPENAI_BASE_URL=http://127.0.0.1:${BITNET_PORT:-8080}/v1"
echo "                   export OPENAI_API_KEY=local"
echo "                   hermes model   # pick custom / OpenAI-compatible"
echo "4) VS Code web:    code-server --bind-addr 127.0.0.1:${CODE_SERVER_PORT:-8443} \"$HERMES_BITNET_SRC_ROOT\""
echo "5) qminiwasm venv: source \"${QMINI_VENV:-$HERMES_BITNET_SRC_ROOT/qminiwasm-core/.venv}/bin/activate\""
