#!/usr/bin/env bash
# Run BitNet llama-server with OpenAI-compatible API for Hermes (127.0.0.1).
set -euo pipefail
_SCRIPT="${BASH_SOURCE[0]}"
[[ -L "$_SCRIPT" ]] && _SCRIPT="$(readlink -f "$_SCRIPT" 2>/dev/null || readlink "$_SCRIPT")"
ROOT_DIR="$(cd "$(dirname "$_SCRIPT")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

BITNET_BUILD_DIR="${BITNET_BUILD_DIR:-$BITNET_DIR/build}"
LLAMA_SERVER=""
for c in "$BITNET_BUILD_DIR/bin/llama-server" "$BITNET_BUILD_DIR/bin/llama-server.exe"; do
  if [[ -x "$c" ]]; then
    LLAMA_SERVER="$c"
    break
  fi
done
[[ -n "$LLAMA_SERVER" ]] || die "llama-server not found under $BITNET_BUILD_DIR/bin — run 04-bitnet-build.sh first"

if [[ -n "${BITNET_GGUF:-}" ]]; then
  MODEL="$BITNET_GGUF"
else
  MODEL="$(find "$BITNET_DIR/models" -type f \( -name '*.gguf' -o -name '*.GGUF' \) 2>/dev/null | head -1 || true)"
fi
[[ -n "${MODEL:-}" && -f "$MODEL" ]] || die "No GGUF found. Set BITNET_GGUF=/path/to/model.gguf or run 04-bitnet-build.sh without BITNET_SKIP_MODEL_DOWNLOAD=1"

THREADS="${BITNET_THREADS:-$(nproc 2>/dev/null || echo 4)}"
CTX="${BITNET_CTX:-4096}"

log "Starting llama-server: $LLAMA_SERVER"
log "  model=$MODEL"
log "  bind=127.0.0.1 port=$BITNET_PORT threads=$THREADS ctx=$CTX"
log "Hermes: OPENAI_BASE_URL=http://127.0.0.1:${BITNET_PORT}/v1 OPENAI_API_KEY=local"

exec "$LLAMA_SERVER" -m "$MODEL" --host 127.0.0.1 --port "$BITNET_PORT" -t "$THREADS" -c "$CTX" "$@"
