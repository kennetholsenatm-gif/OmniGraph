#!/usr/bin/env bash
# Shared helpers for Hermes + BitNet + qminiwasm LXC bootstrap.
set -euo pipefail

die() { echo "error: $*" >&2; exit 1; }
log() { echo "[hermes-bitnet-lxc] $*"; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

ensure_dir() {
  mkdir -p "$1"
}

SRC_ROOT="${HERMES_BITNET_SRC_ROOT:-$HOME/src}"
BITNET_DIR="${BITNET_DIR:-$SRC_ROOT/BitNet}"
# Canonical local clone (Windows: C:\GitHub\LLM_Pract\qminiwasm-core → WSL / Git Bash / Incus bind-mount).
QMINI_LOCAL_DEFAULT="/mnt/c/GitHub/LLM_Pract/qminiwasm-core"
QMINI_REPO="${QMINI_REPO:-https://github.com/kennetholsenatm-gif/qminiwasm-core.git}"
if [[ -z "${QMINI_DIR:-}" ]]; then
  # .git may be a file (worktree) or directory; use -e
  if [[ -e "$QMINI_LOCAL_DEFAULT/.git" ]]; then
    QMINI_DIR="$QMINI_LOCAL_DEFAULT"
  else
    QMINI_DIR="$SRC_ROOT/qminiwasm-core"
  fi
fi
BITNET_VENV="${BITNET_VENV:-$BITNET_DIR/.venv}"
QMINI_VENV="${QMINI_VENV:-$QMINI_DIR/.venv}"
BITNET_PORT="${BITNET_PORT:-8080}"
CODE_SERVER_PORT="${CODE_SERVER_PORT:-8443}"
# Largest model in BitNet setup_env.py --hf-repo list (fallback: 2B GGUF path in docs)
BITNET_HF_REPO="${BITNET_HF_REPO:-tiiuae/Falcon3-10B-Instruct-1.58bit}"
BITNET_QUANT="${BITNET_QUANT:-i2_s}"
# Directory under BitNet repo for weights (created by setup_env / huggingface)
BITNET_MODEL_DIR="${BITNET_MODEL_DIR:-models/falcon3-10b-1_58}"
# BitNet setup_env maps Falcon/Llama3-class models to this codegen layout on x86_64 (see setup_env.py).
BITNET_CODEGEN_MODEL="${BITNET_CODEGEN_MODEL:-Llama3-8B-1.58-100B-tokens}"
# GCC builds cleanly on Alma 10; Clang 20 hit const errors in ggml-bitnet-mad.cpp in tested BitNet revision.
BITNET_CC="${BITNET_CC:-gcc}"
BITNET_CXX="${BITNET_CXX:-g++}"
QMINI_BRANCH="${QMINI_BRANCH:-white-paper-integration}"
