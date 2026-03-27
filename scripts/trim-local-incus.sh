#!/usr/bin/env bash
set -euo pipefail

KEEP_INSTANCES=("${@:-devsecops-gitea}")

if command -v incus >/dev/null 2>&1; then
  CLI="incus"
elif command -v lxc >/dev/null 2>&1; then
  CLI="lxc"
else
  echo "ERROR: neither incus nor lxc found in PATH." >&2
  exit 1
fi

mapfile -t INSTANCES < <($CLI list -c n --format csv | sed '/^$/d')

contains_keep() {
  local needle="$1"
  for i in "${KEEP_INSTANCES[@]}"; do
    [[ "$i" == "$needle" ]] && return 0
  done
  return 1
}

for name in "${INSTANCES[@]}"; do
  if contains_keep "$name"; then
    echo "KEEP $name"
  else
    echo "STOP $name"
    $CLI stop "$name" --force >/dev/null || true
  fi
done

echo "Trim complete. Kept: ${KEEP_INSTANCES[*]}"
