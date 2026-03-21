#!/usr/bin/env python3
"""Idempotent patch: ACP loads ~/.hermes/.env without overriding existing exports.

OpenVSCode agent wrappers set OPENAI_BASE_URL (e.g. OpenRouter) before starting
Hermes. Upstream ``load_hermes_dotenv`` uses override=True for user .env, which
clobbers those variables. ACP entry only ever loads user .env (no project_env),
so we load it with override=False here.
"""
from __future__ import annotations

import sys
from pathlib import Path

MARKER = "hermes-bitnet-lxc: ACP .env load without override"

OLD = '''def _load_env() -> None:
    """Load .env from HERMES_HOME (default ``~/.hermes``)."""
    from hermes_cli.env_loader import load_hermes_dotenv

    hermes_home = Path(os.getenv("HERMES_HOME", Path.home() / ".hermes"))
    loaded = load_hermes_dotenv(hermes_home=hermes_home)
    if loaded:
        for env_file in loaded:
            logging.getLogger(__name__).info("Loaded env from %s", env_file)
    else:
        logging.getLogger(__name__).info(
            "No .env found at %s, using system env", hermes_home / ".env"
        )
'''

NEW = f'''def _load_env() -> None:
    """Load .env from HERMES_HOME (default ``~/.hermes``)."""
    # {MARKER}: do not override OPENAI_BASE_URL etc. set by agent wrappers (OpenRouter).
    from dotenv import load_dotenv

    hermes_home = Path(os.getenv("HERMES_HOME", Path.home() / ".hermes"))
    user_env = hermes_home / ".env"
    loaded: list[Path] = []
    if user_env.exists():
        try:
            load_dotenv(dotenv_path=user_env, override=False, encoding="utf-8")
        except UnicodeDecodeError:
            load_dotenv(dotenv_path=user_env, override=False, encoding="latin-1")
        loaded.append(user_env)
    if loaded:
        for env_file in loaded:
            logging.getLogger(__name__).info("Loaded env from %s", env_file)
    else:
        logging.getLogger(__name__).info(
            "No .env found at %s, using system env", hermes_home / ".env"
        )
'''


def main() -> int:
    path = Path(
        sys.argv[1]
        if len(sys.argv) > 1
        else str(Path.home() / ".hermes/hermes-agent/acp_adapter/entry.py")
    )
    if not path.is_file():
        print(f"warn: entry.py not found, skip: {path}", file=sys.stderr)
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
