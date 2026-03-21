#!/usr/bin/env python3
"""Idempotent patch: honor HERMES_CHAT_COMPLETIONS_NO_TOOLS for BitNet llama-server.

BitNet's vendored llama.cpp server rejects OpenAI chat payloads that include ``tools``
(see ``examples/server/utils.hpp`` → unsupported_params). Hermes sends ``tools`` for
coding-agent function calling; stripping the key avoids HTTP 500 and yields **chat-only**
against this server — not a full tool loop. Remove the env var once a BitNet-compatible
``llama-server`` accepts ``tools`` / ``tool_choice`` (see bundle README).
"""
from __future__ import annotations

import sys
from pathlib import Path

MARKER = "hermes-bitnet-lxc: BitNet llama-server rejects OpenAI tools param"

INSERT = f'''
        # {MARKER}
        if os.getenv("HERMES_CHAT_COMPLETIONS_NO_TOOLS", "").strip().lower() in {{"1", "true", "yes", "on"}}:
            api_kwargs.pop("tools", None)
'''

ANCHOR_BEFORE = '        if self.max_tokens is not None:\n            api_kwargs.update(self._max_tokens_param(self.max_tokens))'


def main() -> int:
    path = Path(sys.argv[1] if len(sys.argv) > 1 else str(Path.home() / ".hermes/hermes-agent/run_agent.py"))
    if not path.is_file():
        print(f"warn: run_agent.py not found, skip: {path}", file=sys.stderr)
        return 0
    text = path.read_text(encoding="utf-8")
    if MARKER in text:
        print(f"already patched: {path}")
        return 0
    if ANCHOR_BEFORE not in text:
        print(
            f"warn: expected anchor not found (Hermes version drift?); skip patch: {path}",
            file=sys.stderr,
        )
        return 0
    path.write_text(text.replace(ANCHOR_BEFORE, INSERT + "\n" + ANCHOR_BEFORE, 1), encoding="utf-8")
    print(f"patched: {path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
