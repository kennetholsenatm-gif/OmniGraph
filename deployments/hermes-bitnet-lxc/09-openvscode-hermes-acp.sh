#!/usr/bin/env bash
# Hermes Agent inside OpenVSCode Server via Agent Client Protocol (ACP).
# - Installs ACP Python deps into the Hermes installer venv (Hermes docs: agent-client-protocol).
# - Installs the Open-VSX "ACP Client" extension (formulahendry.acp-client) for Gitpod OpenVSCode.
# - Patches acp_adapter/session.py so HERMES_ACP_MODEL overrides config (OpenRouter model id vs BitNet GGUF path).
# - Patches acp_adapter/entry.py so ~/.hermes/.env loads with override=False (wrapper OPENAI_BASE_URL kept).
# - Patches run_agent.py so HERMES_ACP_PRESERVE_OPENROUTER_ENV restores OPENAI_BASE_URL + NO_TOOLS after dotenv.
# - Installs ~/.local/bin/hermes-acp-coding-agent (OpenRouter, tools on).
# - Merges two agents into acp.agents (User + Machine):
#     * Hermes (BitNet - chat) — local llama-server, HERMES_CHAT_COMPLETIONS_NO_TOOLS=1
#     * Hermes (OpenRouter - tools) — tool-capable coding agent (needs API key in ~/.hermes/.env)
#
# Hermes upstream ACP doc mentions anysphere.acp-client + acpClient.agents; OpenVSCode uses Open VS X,
# where the maintained client is formulahendry.acp-client and settings key is acp.agents (command/args/env).
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

if [[ "${OPENVS_CODE_SKIP:-0}" == "1" ]]; then
  log "OPENVS_CODE_SKIP=1 — skipping OpenVSCode Hermes ACP integration."
  exit 0
fi

if [[ "${OPENVS_CODE_HERMES_ACP_SKIP:-0}" == "1" ]]; then
  log "OPENVS_CODE_HERMES_ACP_SKIP=1 — skipping Hermes ACP integration."
  exit 0
fi

HERMES_AGENT_DIR="${HERMES_AGENT_DIR:-$HOME/.hermes/hermes-agent}"
HERMES_VENV_PY="${HERMES_VENV_PY:-$HERMES_AGENT_DIR/venv/bin/python3}"
ACP_EXT_ID="${OPENVS_CODE_ACP_EXTENSION:-formulahendry.acp-client}"
# Match openvscode-server --user-data-dir layout (default: ~/.openvscode-server).
OVS_USER_DATA="${OPENVS_CODE_USER_DATA:-$HOME/.openvscode-server}"
OVS_USER_SETTINGS="${OPENVS_CODE_USER_SETTINGS:-$OVS_USER_DATA/data/User/settings.json}"
OVS_MACHINE_SETTINGS="${OPENVS_CODE_MACHINE_SETTINGS:-$OVS_USER_DATA/data/Machine/settings.json}"
HERMES_BIN="${HERMES_BIN:-$HOME/.local/bin/hermes}"
CODING_AGENT_WRAPPER="${HERMES_ACP_CODING_AGENT_BIN:-$HOME/.local/bin/hermes-acp-coding-agent}"

if [[ ! -x "$HERMES_BIN" ]]; then
  if command -v hermes >/dev/null 2>&1; then
    HERMES_BIN="$(command -v hermes)"
  else
    die "hermes not found (install 02-hermes.sh first). Set HERMES_BIN= if nonstandard."
  fi
fi

[[ -d "$HERMES_AGENT_DIR" ]] || die "Hermes agent tree missing: $HERMES_AGENT_DIR"
[[ -x "$HERMES_VENV_PY" ]] || die "Hermes venv python missing: $HERMES_VENV_PY (re-run Hermes install)"

BIN_OVS=""
for c in "$OPENVS_CODE_HOME/bin/openvscode-server" "$OPENVS_CODE_HOME/openvscode-server"; do
  if [[ -x "$c" ]]; then
    BIN_OVS="$c"
    break
  fi
done
if [[ -z "$BIN_OVS" ]]; then
  BIN_OVS="$(find "$OPENVS_CODE_HOME" -maxdepth 3 -type f -name 'openvscode-server' -perm -111 2>/dev/null | head -1 || true)"
fi
[[ -n "$BIN_OVS" ]] || die "openvscode-server binary not found under $OPENVS_CODE_HOME (run 07-openvscode-server.sh)"

log "ACP: installing agent-client-protocol into Hermes venv (required for hermes acp)"
if command -v uv >/dev/null 2>&1; then
  (cd "$HERMES_AGENT_DIR" && uv pip install 'agent-client-protocol>=0.8.1,<1.0' --python "$HERMES_VENV_PY")
else
  log "uv not on PATH — using ensurepip + pip in Hermes venv"
  "$HERMES_VENV_PY" -m ensurepip --upgrade >/dev/null
  "$HERMES_VENV_PY" -m pip install -q 'agent-client-protocol>=0.8.1,<1.0'
fi

log "ACP: installing VS Code extension $ACP_EXT_ID (Open VS X)"
"$BIN_OVS" --install-extension "$ACP_EXT_ID"

if [[ -f "$HERMES_AGENT_DIR/acp_adapter/session.py" ]]; then
  python3 "$ROOT_DIR/lib/patch_hermes_acp_session_model_env.py" "$HERMES_AGENT_DIR/acp_adapter/session.py"
else
  log "warn: $HERMES_AGENT_DIR/acp_adapter/session.py missing — skip HERMES_ACP_MODEL patch"
fi

if [[ -f "$HERMES_AGENT_DIR/acp_adapter/entry.py" ]]; then
  python3 "$ROOT_DIR/lib/patch_hermes_acp_entry_dotenv_no_override.py" "$HERMES_AGENT_DIR/acp_adapter/entry.py"
else
  log "warn: $HERMES_AGENT_DIR/acp_adapter/entry.py missing — skip ACP dotenv patch"
fi

if [[ -f "$HERMES_AGENT_DIR/run_agent.py" ]]; then
  python3 "$ROOT_DIR/lib/patch_hermes_run_agent_preserve_acp_openrouter_env.py" "$HERMES_AGENT_DIR/run_agent.py"
else
  log "warn: $HERMES_AGENT_DIR/run_agent.py missing — skip ACP OpenRouter env preserve patch"
fi

ensure_dir "$HOME/.local/bin"
cp -f "$ROOT_DIR/hermes-acp-coding-agent.sh" "$CODING_AGENT_WRAPPER"
chmod 755 "$CODING_AGENT_WRAPPER"
log "Installed $CODING_AGENT_WRAPPER"

merge_hermes_acp_settings() {
  local target="$1"
  ensure_dir "$(dirname "$target")"
  export _OVS_SETTINGS_PATH="$target"
  export _HERMES_BIN="$HERMES_BIN"
  export _CODING_WRAPPER="$CODING_AGENT_WRAPPER"
  export _BITNET_PORT="${BITNET_PORT:-8080}"
  python3 - <<'PY'
import json
import os
from pathlib import Path

path = Path(os.environ["_OVS_SETTINGS_PATH"])
hermes_bin = os.environ["_HERMES_BIN"]
wrapper = os.environ["_CODING_WRAPPER"]
bitnet_port = os.environ.get("_BITNET_PORT", "8080").strip() or "8080"

data = {}
if path.is_file():
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError:
        data = {}

agents = data.get("acp.agents")
if agents is None:
    agents = {}
elif not isinstance(agents, dict):
    agents = {}

for stale in (
    "Hermes Agent",
    "Hermes (BitNet \u2014 chat)",
    "Hermes (OpenRouter \u2014 tools)",
):
    agents.pop(stale, None)

k_bitnet = "Hermes (BitNet - chat)"
k_open = "Hermes (OpenRouter - tools)"

bitnet_entry = {
    "command": hermes_bin,
    "args": ["acp"],
    "env": {
        "OPENAI_BASE_URL": f"http://127.0.0.1:{bitnet_port}/v1",
        "OPENAI_API_KEY": "dummy",
        "HERMES_CHAT_COMPLETIONS_NO_TOOLS": "1",
    },
}
open_entry = {
    "command": wrapper,
    "args": [],
    "env": {},
}

agents[k_bitnet] = bitnet_entry
agents[k_open] = open_entry
data["acp.agents"] = agents

new_text = json.dumps(data, indent=2, sort_keys=False) + "\n"
old_text = path.read_text(encoding="utf-8") if path.is_file() else ""
if old_text == new_text:
    print(f"[hermes-bitnet-lxc] ACP: unchanged ({path})")
else:
    path.write_text(new_text, encoding="utf-8")
    print(f"[hermes-bitnet-lxc] ACP: wrote {k_bitnet} + {k_open} -> {path}")
PY
}

merge_hermes_acp_settings "$OVS_USER_SETTINGS"
merge_hermes_acp_settings "$OVS_MACHINE_SETTINGS"

log "Done. In OpenVSCode: reload the window (Command Palette -> Developer: Reload Window), ACP -> Agents."
log "  - Hermes (OpenRouter - tools): add OPENROUTER_API_KEY to ~/.hermes/.env for a working coding agent."
log "  - Hermes (BitNet - chat): local BitNet; no OpenAI tools on that server."
log "Upstream ACP doc: $HERMES_AGENT_DIR/docs/acp-setup.md"
log "09-openvscode-hermes-acp: done"
