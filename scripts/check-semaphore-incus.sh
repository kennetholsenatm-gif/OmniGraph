#!/usr/bin/env bash
# Show Semaphore status inside the Incus LXC (not on the host).
# Usage: ./scripts/check-semaphore-incus.sh [instance-name]
set -euo pipefail
NAME="${1:-devsecops-semaphore}"

if ! command -v incus >/dev/null 2>&1; then
  echo "incus not found in PATH" >&2
  exit 1
fi

if ! incus list --format csv -c n | grep -qx "${NAME}"; then
  echo "Instance '${NAME}' not found. Run: scripts/start-semaphore.sh (or deploy-semaphore-incus.yml)" >&2
  exit 1
fi

echo "=== incus config device (HTTP proxies) ==="
incus config device show "${NAME}" | grep -E '^(semaphore_http|semaphore_http6):' -A2 || true

echo ""
echo "=== systemctl status semaphore (inside ${NAME}) ==="
incus exec "${NAME}" -- systemctl status semaphore --no-pager || true

echo ""
echo "=== listen :3000 (inside ${NAME}) ==="
incus exec "${NAME}" -- ss -tlnp 2>/dev/null | grep 3000 || true

echo ""
echo "Open UI (IPv4): http://127.0.0.1:3001"
