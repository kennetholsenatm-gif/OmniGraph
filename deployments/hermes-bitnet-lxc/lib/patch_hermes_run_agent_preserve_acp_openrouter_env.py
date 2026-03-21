#!/usr/bin/env python3
"""Idempotent patch: preserve OpenRouter wrapper env across run_agent dotenv load.

``hermes-acp-coding-agent`` exports ``OPENAI_BASE_URL`` (OpenRouter) and
``HERMES_CHAT_COMPLETIONS_NO_TOOLS=0`` before ``python -m acp_adapter``. At import
time ``run_agent`` calls ``load_hermes_dotenv(..., override=True)`` for
``~/.hermes/.env``, which overwrites those with BitNet + ``NO_TOOLS=1``. This block
snapshots them when ``HERMES_ACP_PRESERVE_OPENROUTER_ENV=1`` and restores after load.
"""
from __future__ import annotations

import sys
from pathlib import Path

MARKER = "hermes-bitnet-lxc: preserve ACP OpenRouter env across load_hermes_dotenv"

OLD = """_hermes_home = Path(os.getenv("HERMES_HOME", Path.home() / ".hermes"))
_project_env = Path(__file__).parent / '.env'
_loaded_env_paths = load_hermes_dotenv(hermes_home=_hermes_home, project_env=_project_env)
"""

NEW = f"""_hermes_home = Path(os.getenv("HERMES_HOME", Path.home() / ".hermes"))
_project_env = Path(__file__).parent / '.env'
# {MARKER}
_preserve_acp_or = os.getenv("HERMES_ACP_PRESERVE_OPENROUTER_ENV", "").strip().lower() in {{"1", "true", "yes", "on"}}
_preserve_openai_base = os.environ.get("OPENAI_BASE_URL") if _preserve_acp_or else None
_preserve_no_tools = os.environ.get("HERMES_CHAT_COMPLETIONS_NO_TOOLS") if _preserve_acp_or else None
_loaded_env_paths = load_hermes_dotenv(hermes_home=_hermes_home, project_env=_project_env)
if _preserve_acp_or:
    if _preserve_openai_base:
        os.environ["OPENAI_BASE_URL"] = _preserve_openai_base
    if _preserve_no_tools is not None:
        os.environ["HERMES_CHAT_COMPLETIONS_NO_TOOLS"] = _preserve_no_tools
"""


def main() -> int:
    path = Path(
        sys.argv[1]
        if len(sys.argv) > 1
        else str(Path.home() / ".hermes/hermes-agent/run_agent.py")
    )
    if not path.is_file():
        print(f"warn: run_agent.py not found, skip: {path}", file=sys.stderr)
        return 0
    text = path.read_text(encoding="utf-8")
    if MARKER in text:
        print(f"already patched: {path}")
        return 0
    if OLD not in text:
        print(
            f"warn: anchor not found; Hermes version drift? skip: {path}",
            file=sys.stderr,
        )
        return 0
    path.write_text(text.replace(OLD, NEW, 1), encoding="utf-8")
    print(f"patched: {path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
