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
# Canonical local tree (your GitHub checkout): Windows C:\GitHub\LLM_Pract\qminiwasm-core → WSL /mnt/c/GitHub/LLM_Pract/qminiwasm-core
# Override with QMINI_LOCAL_DEFAULT or set QMINI_DIR explicitly (Incus bind-mount, etc.).
QMINI_LOCAL_DEFAULT="${QMINI_LOCAL_DEFAULT:-/mnt/c/GitHub/LLM_Pract/qminiwasm-core}"
# Remote used only when cloning into QMINI_DIR (no .git yet). Point at your fork or mirror; local editing stays on LLM_Pract path when detected.
QMINI_REPO="${QMINI_REPO_URL:-${QMINI_REPO:-https://github.com/kennetholsenatm-gif/qminiwasm-core.git}}"
if [[ -z "${QMINI_DIR:-}" ]]; then
  QMINI_DIR="$SRC_ROOT/qminiwasm-core"
  # Prefer first existing local clone among candidates (.git may be file for worktrees).
  for _q in "$QMINI_LOCAL_DEFAULT" "${QMINI_LOCAL_ALT:-}"; do
    [[ -n "$_q" && -e "$_q/.git" ]] || continue
    QMINI_DIR="$_q"
    break
  done
fi
BITNET_VENV="${BITNET_VENV:-$BITNET_DIR/.venv}"
# Keep the venv on the Linux filesystem (ext4) — not under /mnt/c — so pip/torch unpack is fast.
QMINI_VENV="${QMINI_VENV:-$HOME/venvs/qminiwasm-core}"
BITNET_PORT="${BITNET_PORT:-8080}"
CODE_SERVER_PORT="${CODE_SERVER_PORT:-8443}"
# Default 3010: many Gitea installs use 3000 on the same host/WSL — override with OPENVS_CODE_PORT.
OPENVS_CODE_PORT="${OPENVS_CODE_PORT:-3010}"
OPENVS_CODE_HOME="${OPENVS_CODE_HOME:-$HOME/openvscode-server}"
# 0 = no URL token (--without-connection-token); bind OPENVS_CODE_BIND (default 127.0.0.1). 1 = --connection-token-file + ?tkn= for LAN.
OPENVS_CODE_REQUIRE_TOKEN="${OPENVS_CODE_REQUIRE_TOKEN:-0}"
OPENVS_CODE_BIND="${OPENVS_CODE_BIND:-127.0.0.1}"
# Largest model in BitNet setup_env.py --hf-repo list (fallback: 2B GGUF path in docs)
BITNET_HF_REPO="${BITNET_HF_REPO:-tiiuae/Falcon3-10B-Instruct-1.58bit}"
BITNET_QUANT="${BITNET_QUANT:-i2_s}"
# Directory under BitNet repo for -md (parent of HF model_name folder); see BitNet setup_env.py.
BITNET_MODEL_DIR="${BITNET_MODEL_DIR:-models/falcon3-10b-instruct}"
# Subfolder name under BITNET_MODEL_DIR for Falcon3-10B-Instruct-1.58bit weights + gguf
BITNET_HF_MODEL_NAME="${BITNET_HF_MODEL_NAME:-Falcon3-10B-Instruct-1.58bit}"
# Expected GGUF after setup_env (set BITNET_GGUF to override for run-bitnet / Hermes)
BITNET_GGUF_DEFAULT="$BITNET_DIR/$BITNET_MODEL_DIR/$BITNET_HF_MODEL_NAME/ggml-model-${BITNET_QUANT}.gguf"
# BitNet setup_env maps Falcon/Llama3-class models to this codegen layout on x86_64 (see setup_env.py).
BITNET_CODEGEN_MODEL="${BITNET_CODEGEN_MODEL:-Llama3-8B-1.58-100B-tokens}"
# GCC builds cleanly on Alma 10; Clang 20 hit const errors in ggml-bitnet-mad.cpp in tested BitNet revision.
BITNET_CC="${BITNET_CC:-gcc}"
BITNET_CXX="${BITNET_CXX:-g++}"
QMINI_BRANCH="${QMINI_BRANCH:-white-paper-integration}"
