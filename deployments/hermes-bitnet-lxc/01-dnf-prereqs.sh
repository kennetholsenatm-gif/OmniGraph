#!/usr/bin/env bash
# AlmaLinux / RHEL-family: packages Hermes installer does not pull on almalinux ID,
# plus build deps for BitNet (clang, cmake) and common tools.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

if [[ "$(id -u)" -ne 0 ]] && command -v sudo >/dev/null 2>&1; then
  SUDO="sudo"
elif [[ "$(id -u)" -eq 0 ]]; then
  SUDO=""
else
  die "run as root or with sudo available"
fi

log "dnf: installing base packages (git, nodejs, ripgrep, ffmpeg, build chain, cmake, jq)..."

$SUDO dnf install -y \
  git \
  curl \
  wget \
  ca-certificates \
  tar \
  xz \
  which \
  jq \
  ripgrep \
  ffmpeg-free \
  gcc \
  gcc-c++ \
  make \
  cmake \
  ninja-build \
  pkgconf-pkg-config \
  openssl-devel \
  python3 \
  python3-devel \
  python3-pip \
  zlib-devel \
  perl-IPC-Cmd \
  git-lfs \
  nodejs \
  npm \
  || die "dnf install failed"

# Prefer clang for BitNet upstream; Alma 10 AppStream typically provides a recent clang.
if ! command -v clang >/dev/null 2>&1; then
  $SUDO dnf install -y clang llvm || log "warn: clang not installed as a standalone package; check llvm-toolset"
fi

if command -v clang >/dev/null 2>&1; then
  clang --version | head -1
else
  log "warn: clang still missing — BitNet build may fail until you install clang >= 18"
fi

cmake --version | head -1
git --version
rg --version | head -1 || true
log "01-dnf-prereqs: done"
