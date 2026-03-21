#!/usr/bin/env bash
# Quick check: canonical Windows→WSL paths and common.sh resolution (run inside WSL).
set -eu
for p in /mnt/c/GitHub/LLM_Pract/qminiwasm-core /mnt/c/GiTeaRepos/devsecops-pipeline; do
  printf '%s: ' "$p"
  if [[ -e "$p/.git" ]]; then echo OK; else echo missing; fi
done
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=common.sh
source "$ROOT/lib/common.sh"
echo "QMINI_DIR=$QMINI_DIR"
echo "OPENVS_CODE_PORT=$OPENVS_CODE_PORT (use this in browser; keep 3000 for Gitea)"
echo "BITNET_PORT=$BITNET_PORT"
