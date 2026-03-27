#!/usr/bin/env bash
# Point Hermes at local BitNet llama-server (OpenAI-compatible). Idempotent-ish; backs up config once.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

if [[ "${HERMES_BITNET_CONFIG_SKIP:-0}" == "1" ]]; then
  log "HERMES_BITNET_CONFIG_SKIP=1 — skipping Hermes BitNet wiring."
  exit 0
fi

HERMES_RUN_AGENT="${HERMES_RUN_AGENT:-$HOME/.hermes/hermes-agent/run_agent.py}"
if [[ -f "$HERMES_RUN_AGENT" ]]; then
  python3 "$ROOT_DIR/lib/patch_hermes_run_agent_no_tools.py" "$HERMES_RUN_AGENT"
else
  log "warn: $HERMES_RUN_AGENT missing — install Hermes (02-hermes.sh) before BitNet ACP chat."
fi

append_env_kv() {
  local f="$1"
  local key="$2"
  local val="$3"
  [[ -f "$f" ]] || return 0
  if grep -q "^${key}=" "$f" 2>/dev/null; then
    return 0
  fi
  printf '\n# BitNet llama-server OpenAI compat (no tools param); see 08 + README\n%s=%s\n' "$key" "$val" >>"$f"
}

# Hermes → BitNet: vendor server rejects ``tools`` in /v1/chat/completions (no OpenAI function-calling on this build).
# Enables Hermes chat/ACP against BitNet; full coding-agent tool loop needs a llama-server that accepts ``tools`` (see README).
append_env_kv "$HOME/.hermes/.env" HERMES_CHAT_COMPLETIONS_NO_TOOLS 1

GGUF="${BITNET_GGUF:-$BITNET_GGUF_DEFAULT}"
if [[ ! -f "$GGUF" ]]; then
  log "BitNet GGUF not found yet at $GGUF — skip Hermes config.yaml + OPENAI_* append (run 04-bitnet-build.sh first)."
  log "Patch + HERMES_CHAT_COMPLETIONS_NO_TOOLS=.env already applied when possible."
  log "Re-run 08-hermes-bitnet-config.sh after weights exist."
  exit 0
fi

CFG="$HOME/.hermes/config.yaml"
if [[ ! -f "$CFG" ]]; then
  log "warn: $CFG missing — install Hermes (02-hermes.sh) first."
  exit 0
fi

BAK="$CFG.bak-bitnet-bootstrap"
if [[ ! -f "$BAK" ]]; then
  cp -a "$CFG" "$BAK"
  log "Backed up Hermes config to $BAK"
fi

export HERMES_BITNET_GGUF="$GGUF"
export HERMES_BITNET_PORT="$BITNET_PORT"
python3 - <<'PY'
import os
import re
from pathlib import Path

cfg = Path.home() / ".hermes/config.yaml"
text = cfg.read_text(encoding="utf-8")
gguf = os.environ["HERMES_BITNET_GGUF"]
port = os.environ["HERMES_BITNET_PORT"]


def sub_once(pattern: str, repl: str, s: str) -> str:
    out, n = re.subn(pattern, repl, s, count=1, flags=re.MULTILINE)
    if n == 0:
        print("warn: pattern not found, leaving text unchanged:", pattern)
        return s
    return out

# Hermes stock config indents model keys with two spaces.
text = sub_once(r"^  default:\s*.*$", f'  default: "{gguf}"', text)
text = sub_once(r"^  base_url:\s*.*$", f'  base_url: "http://127.0.0.1:{port}/v1"', text)
cfg.write_text(text, encoding="utf-8")
print("updated", cfg)
PY

ENV_FILE="$HOME/.hermes/.env"
if [[ -f "$ENV_FILE" ]]; then
  append() {
    local key="$1"
    local val="$2"
    if grep -q "^${key}=" "$ENV_FILE" 2>/dev/null; then
      return 0
    fi
    printf '\n# Local BitNet llama-server (bootstrap 08-hermes-bitnet-config.sh)\n%s=%s\n' "$key" "$val" >>"$ENV_FILE"
  }
  append OPENAI_BASE_URL "http://127.0.0.1:${BITNET_PORT}/v1"
  append OPENAI_API_BASE "http://127.0.0.1:${BITNET_PORT}/v1"
  append OPENAI_API_KEY "dummy"
  log "Appended OPENAI_* entries to $ENV_FILE when missing."
fi

log "Hermes should use model id: $GGUF"
log "Run: hermes doctor"
log "08-hermes-bitnet-config: done"
