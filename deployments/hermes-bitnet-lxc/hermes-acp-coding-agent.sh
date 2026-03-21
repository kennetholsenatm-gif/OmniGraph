#!/usr/bin/env bash
# Hermes ACP using a tool-capable OpenAI-compatible API (default: OpenRouter).
# BitNet llama-server cannot accept ``tools``; use this entry for coding-agent loops.
#
# Env (typical: ~/.hermes/.env):
#   OPENROUTER_API_KEY or OPENAI_API_KEY — required (can live only in ~/.hermes/.env)
# Optional:
#   HERMES_ACP_OPENROUTER_BASE — default https://openrouter.ai/api/v1
#   HERMES_ACP_MODEL — default openai/gpt-4o-mini
#   HERMES_AGENT_DIR — default ~/.hermes/hermes-agent (venv for python -m acp_adapter)
#
# Uses ``python -m acp_adapter`` so hermes_cli/main.py is not imported (that path loads
# ~/.hermes/.env with override=True and would clobber OPENAI_BASE_URL before ACP starts).
set -euo pipefail

HERMES_AGENT_DIR="${HERMES_AGENT_DIR:-$HOME/.hermes/hermes-agent}"
VENV_PY="${HERMES_VENV_PY:-$HERMES_AGENT_DIR/venv/bin/python3}"

BASE="${HERMES_ACP_OPENROUTER_BASE:-https://openrouter.ai/api/v1}"
MODEL="${HERMES_ACP_MODEL:-openai/gpt-4o-mini}"

export OPENAI_BASE_URL="$BASE"
export OPENROUTER_API_KEY="${OPENROUTER_API_KEY:-${OPENAI_API_KEY:-}}"
export OPENAI_API_KEY="${OPENAI_API_KEY:-$OPENROUTER_API_KEY}"
export HERMES_ACP_MODEL="$MODEL"
# Allow Hermes to send OpenAI ``tools`` (overrides e.g. systemd HERMES_CHAT_COMPLETIONS_NO_TOOLS=1).
export HERMES_CHAT_COMPLETIONS_NO_TOOLS=0
# run_agent.py imports load_hermes_dotenv(override=True); ~/.hermes/.env would clobber the two exports above.
export HERMES_ACP_PRESERVE_OPENROUTER_ENV=1

[[ -x "$VENV_PY" ]] || {
  echo "hermes-acp-coding-agent: missing venv python: $VENV_PY" >&2
  exit 1
}

exec "$VENV_PY" -m acp_adapter "$@"
