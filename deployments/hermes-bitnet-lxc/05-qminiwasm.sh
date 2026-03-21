#!/usr/bin/env bash
# Clone qminiwasm-core (CPU baseline). Optional: QMINI_USE_ARC=1 for pip install -e ".[arc]" experiments.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

require_cmd git
require_cmd python3

ensure_dir "$SRC_ROOT"

if [[ ! -d "$QMINI_DIR/.git" ]]; then
  log "No git repo at $QMINI_DIR — cloning $QMINI_REPO (branch $QMINI_BRANCH) ..."
  ensure_dir "$(dirname "$QMINI_DIR")"
  git clone --branch "$QMINI_BRANCH" --single-branch "$QMINI_REPO" "$QMINI_DIR"
else
  log "Using qminiwasm-core at $QMINI_DIR"
  if [[ "${QMINI_SYNC_BRANCH:-0}" == "1" ]]; then
    log "QMINI_SYNC_BRANCH=1 — fetching and checking out $QMINI_BRANCH ..."
    git -C "$QMINI_DIR" fetch origin
    git -C "$QMINI_DIR" checkout "$QMINI_BRANCH"
    git -C "$QMINI_DIR" pull --ff-only origin "$QMINI_BRANCH" || true
  else
    log "Leaving branch/worktree unchanged (set QMINI_SYNC_BRANCH=1 to align with origin/$QMINI_BRANCH)."
  fi
fi

if [[ ! -d "$QMINI_VENV" ]]; then
  python3 -m venv "$QMINI_VENV"
fi
# shellcheck disable=SC1090
source "$QMINI_VENV/bin/activate"
python -m pip install --upgrade pip wheel setuptools

cd "$QMINI_DIR"
if [[ "${QMINI_USE_ARC:-0}" == "1" ]]; then
  log "Installing with [arc] extra (Intel Arc/XPU-oriented; verify for Iris Xe before relying on it)."
  pip install -e ".[arc]"
else
  log "Installing CPU baseline (pip install -e .)."
  pip install -e "."
fi

python -c "import qminiwasm; print('qminiwasm import OK')" || log "warn: import check failed — see repo README for deps"

log "Activate training venv: source \"$QMINI_VENV/bin/activate\""
log "05-qminiwasm: done"
