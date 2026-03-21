#!/usr/bin/env python3
"""Idempotent patch: allow HERMES_ACP_MODEL to override default model in ACP sessions.

When using OpenRouter (or any tool-capable API) from OpenVSCode while config.yaml
still points at a local BitNet GGUF path, set HERMES_ACP_MODEL=openai/gpt-4o-mini
in the agent wrapper environment so ACP uses the remote model id.
"""
from __future__ import annotations

import sys
from pathlib import Path

MARKER = "hermes-bitnet-lxc: HERMES_ACP_MODEL override"

OLD = """        elif isinstance(model_cfg, str) and model_cfg.strip():
            default_model = model_cfg.strip()

        kwargs = {
"""

NEW = f"""        elif isinstance(model_cfg, str) and model_cfg.strip():
            default_model = model_cfg.strip()

        # {MARKER}
        import os as _os
        _hermes_acp_model = (_os.getenv("HERMES_ACP_MODEL") or "").strip()
        if _hermes_acp_model:
            default_model = _hermes_acp_model

        kwargs = {{
"""


def main() -> int:
    path = Path(
        sys.argv[1]
        if len(sys.argv) > 1
        else str(Path.home() / ".hermes/hermes-agent/acp_adapter/session.py")
    )
    if not path.is_file():
        print(f"warn: session.py not found, skip: {path}", file=sys.stderr)
        return 0
    text = path.read_text(encoding="utf-8")
    if MARKER in text:
        print(f"already patched: {path}")
        return 0
    if OLD not in text:
        print(f"warn: anchor not found; Hermes version drift? skip: {path}", file=sys.stderr)
        return 0
    path.write_text(text.replace(OLD, NEW, 1), encoding="utf-8")
    print(f"patched: {path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
