#!/usr/bin/env bash
# Bridge ISR serial console ↔ TCP (for Ansible ansible.netcommon.telnet to 127.0.0.1:PORT).
# Requires: socat (apt install socat)
#
# Usage:
#   chmod +x scripts/socat-console-bridge.sh
#   export ISR_CONSOLE_DEVICE=/dev/ttyUSB0   # or /dev/ttyS2 (WSL COM3 mapping)
#   export ISR_CONSOLE_PORT=3322
#   export ISR_CONSOLE_BAUD=9600
#   ./scripts/socat-console-bridge.sh
#
# Windows: use WSL2 + usbipd-win to attach the USB-serial adapter, then set ISR_CONSOLE_DEVICE
# to the device shown in WSL (often /dev/ttyUSB0). See docs/ISR_CONSOLE_SOCAT_ANSIBLE.md

set -euo pipefail

PORT="${ISR_CONSOLE_PORT:-3322}"
DEV="${ISR_CONSOLE_DEVICE:-/dev/ttyUSB0}"
BAUD="${ISR_CONSOLE_BAUD:-9600}"
BIND="${ISR_CONSOLE_BIND:-127.0.0.1}"

if [[ ! -e "$DEV" && ! -c "$DEV" ]]; then
  echo "ERROR: serial device not found: $DEV" >&2
  exit 1
fi

echo "socat: $BIND:$PORT <-> $DEV (${BAUD} 8N1 raw)"
echo "Test:  nc -v $BIND $PORT   (expect Cisco prompt)"
echo "Stop:  Ctrl+C"

exec socat -d -d \
  "TCP-LISTEN:${PORT},bind=${BIND},reuseaddr,fork,end-close" \
  "FILE:${DEV},b${BAUD},cs8,raw,echo=0"
