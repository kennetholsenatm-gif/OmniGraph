#!/usr/bin/env python3
"""Idempotent patch: BitNet setup_env.py uses Clang by default; GCC + LLAMA_BUILD_SERVER is reliable on Alma."""
from __future__ import annotations

import sys
from pathlib import Path


def main() -> int:
    path = Path(sys.argv[1] if len(sys.argv) > 1 else "setup_env.py")
    text = path.read_text(encoding="utf-8")
    old = (
        'run_command(["cmake", "-B", "build", *COMPILER_EXTRA_ARGS[arch], '
        '*OS_EXTRA_ARGS.get(platform.system(), []), "-DCMAKE_C_COMPILER=clang", '
        '"-DCMAKE_CXX_COMPILER=clang++"], log_step="generate_build_files")'
    )
    new = (
        'run_command(["cmake", "-B", "build", *COMPILER_EXTRA_ARGS[arch], '
        '*OS_EXTRA_ARGS.get(platform.system(), []), "-DLLAMA_BUILD_SERVER=ON", '
        '"-DCMAKE_C_COMPILER=gcc", "-DCMAKE_CXX_COMPILER=g++"], '
        'log_step="generate_build_files")'
    )
    if new in text:
        print(f"already patched: {path}")
        return 0
    if old not in text:
        print(f"warn: expected cmake stanza not found; skip patch: {path}", file=sys.stderr)
        return 0
    path.write_text(text.replace(old, new, 1), encoding="utf-8")
    print(f"patched: {path}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
