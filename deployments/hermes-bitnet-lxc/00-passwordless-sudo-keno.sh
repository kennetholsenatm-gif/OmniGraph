#!/usr/bin/env bash
# Apply passwordless sudo for user keno (dev WSL / personal LXC only).
# Must run as root:  sudo bash 00-passwordless-sudo-keno.sh
set -euo pipefail
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
[[ "$(id -u)" -eq 0 ]] || { echo "Run as root: sudo bash $0" >&2; exit 1; }
install -m 440 -o root -g root "$ROOT_DIR/sudoers.d/keno-nopasswd" /etc/sudoers.d/keno
visudo -c -f /etc/sudoers.d/keno
echo "Installed /etc/sudoers.d/keno (NOPASSWD for keno)."
