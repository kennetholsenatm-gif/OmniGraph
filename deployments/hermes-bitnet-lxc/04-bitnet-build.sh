#!/usr/bin/env bash
# Clone Microsoft BitNet, Python venv, pip deps, cmake build with llama-server, optional model setup.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

require_cmd git
require_cmd python3
require_cmd cmake

ensure_dir "$SRC_ROOT"

if [[ ! -d "$BITNET_DIR/.git" ]]; then
  log "Cloning BitNet into $BITNET_DIR ..."
  git clone --recursive https://github.com/microsoft/BitNet.git "$BITNET_DIR"
else
  log "Updating existing $BITNET_DIR ..."
  git -C "$BITNET_DIR" pull --ff-only || true
  git -C "$BITNET_DIR" submodule update --init --recursive
fi

if command -v git-lfs >/dev/null 2>&1; then
  git lfs install 2>/dev/null || true
fi

if [[ ! -d "$BITNET_VENV" ]]; then
  python3 -m venv "$BITNET_VENV"
fi
# shellcheck disable=SC1090
source "$BITNET_VENV/bin/activate"
python -m pip install --upgrade pip wheel setuptools

if [[ -f "$BITNET_DIR/requirements.txt" ]]; then
  pip install -r "$BITNET_DIR/requirements.txt"
fi

cd "$BITNET_DIR"
log "Installing gguf-py and generating include/bitnet-lut-kernels.h (codegen_tl2, model=$BITNET_CODEGEN_MODEL) ..."
pip install "$BITNET_DIR/3rdparty/llama.cpp/gguf-py"
python utils/codegen_tl2.py \
  --model "$BITNET_CODEGEN_MODEL" \
  --BM 256,128,256,128 \
  --BK 96,96,96,96 \
  --bm 32,32,32,32

BITNET_BUILD_DIR="${BITNET_BUILD_DIR:-$BITNET_DIR/build}"
log "Configuring CMake (LLAMA_BUILD_SERVER=ON, BITNET_X86_TL2=OFF, CC=$BITNET_CC CXX=$BITNET_CXX) ..."
cmake -S "$BITNET_DIR" -B "$BITNET_BUILD_DIR" \
  -DLLAMA_BUILD_SERVER=ON \
  -DBITNET_X86_TL2=OFF \
  -DCMAKE_C_COMPILER="$BITNET_CC" \
  -DCMAKE_CXX_COMPILER="$BITNET_CXX"
cmake --build "$BITNET_BUILD_DIR" -j"$(nproc 2>/dev/null || echo 4)"

LLAMA_SERVER=""
for c in "$BITNET_BUILD_DIR/bin/llama-server" "$BITNET_BUILD_DIR/bin/llama-server.exe"; do
  if [[ -x "$c" ]]; then
    LLAMA_SERVER="$c"
    break
  fi
done
if [[ -z "$LLAMA_SERVER" ]]; then
  log "warn: llama-server binary not found under $BITNET_BUILD_DIR/bin; try: find \"$BITNET_BUILD_DIR\" -name 'llama-server*'"
fi

if [[ "${BITNET_SKIP_MODEL_DOWNLOAD:-0}" == "1" ]]; then
  log "BITNET_SKIP_MODEL_DOWNLOAD=1 — skipping setup_env.py / weights."
  log "04-bitnet-build: done (build only)"
  exit 0
fi

log "Patching setup_env.py to use GCC + LLAMA_BUILD_SERVER=ON (upstream defaults to Clang; see lib/patch_bitnet_setup_env.py)."
python3 "$ROOT_DIR/lib/patch_bitnet_setup_env.py" "$BITNET_DIR/setup_env.py"

log "Downloading / preparing model via setup_env.py (repo=$BITNET_HF_REPO quant=$BITNET_QUANT dir=$BITNET_MODEL_DIR) ..."
log "If OOM or HF errors, set BITNET_HF_REPO to a smaller model (see BitNet README / setup_env.py --help)."
cd "$BITNET_DIR"
ensure_dir "$BITNET_DIR/$(dirname "$BITNET_MODEL_DIR")"
python setup_env.py \
  --hf-repo "$BITNET_HF_REPO" \
  -md "$BITNET_MODEL_DIR" \
  -q "$BITNET_QUANT"

log "Find generated .gguf under $BITNET_DIR/$BITNET_MODEL_DIR (or $BITNET_DIR/models). Run:"
log "  BITNET_GGUF=/path/to/model.gguf $ROOT_DIR/run-bitnet-server.sh"
log "04-bitnet-build: done"
