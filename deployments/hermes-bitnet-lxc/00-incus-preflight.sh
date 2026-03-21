#!/usr/bin/env bash
# Preflight for Incus on the *host* (WSL Alma, bare metal, or VM). Safe to run inside an LXC too (no-op + hints).
# Exits 0 always; prints actionable warnings.
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "$ROOT_DIR/lib/common.sh"

log "Incus / LXC preflight"

if ! command -v incus >/dev/null 2>&1; then
  log "incus CLI not found — if you are on the container guest, run this script on the Incus *host* instead."
  log "Fallback: run bootstrap-all.sh directly on Alma WSL or bare metal without Incus."
  exit 0
fi

if incus info >/dev/null 2>&1; then
  log "incus info: OK"
  incus info 2>/dev/null | head -20 || true
else
  log "warn: incus info failed — start incusd (systemctl --user or systemctl) or check permissions (incus admin init)."
fi

if [[ -f /proc/1/cgroup ]] && grep -qE 'lxc|docker|incus' /proc/1/cgroup 2>/dev/null; then
  log "This shell appears to be inside a container cgroup — nested Incus from here is unusual; prefer host-side incus launch."
fi

if [[ -f /proc/sys/kernel/osrelease ]] && grep -qi microsoft /proc/version 2>/dev/null; then
  log "WSL detected: nested Incus + LXC is fragile. Ensure WSL systemd, adequate .wslconfig memory, and test: incus launch images:almalinux/10/cloud t-incus-test -c limits.memory=4GiB"
fi

log "See: $ROOT_DIR/incus/README.md for profile + launch examples."
log "00-incus-preflight: done (exit 0)"
