#!/usr/bin/env bash
# Create/update Incus profile "hermes-bitnet" (run on Incus host).
set -euo pipefail

PROFILE_NAME="${INCUS_PROFILE_NAME:-hermes-bitnet}"
LIMITS_MEMORY="${INCUS_LIMITS_MEMORY:-56GiB}"
LIMITS_CPU="${INCUS_LIMITS_CPU:-4}"
NESTING="${INCUS_SECURITY_NESTING:-true}"

if ! command -v incus >/dev/null 2>&1; then
  echo "error: incus not found" >&2
  exit 1
fi

if ! incus profile show "$PROFILE_NAME" >/dev/null 2>&1; then
  incus profile create "$PROFILE_NAME"
fi

incus profile set "$PROFILE_NAME" limits.memory="$LIMITS_MEMORY"
incus profile set "$PROFILE_NAME" limits.cpu="$LIMITS_CPU"
incus profile set "$PROFILE_NAME" security.nesting="$NESTING"

echo "Profile $PROFILE_NAME updated:"
incus profile show "$PROFILE_NAME"
