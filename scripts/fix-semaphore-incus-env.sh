#!/usr/bin/env bash
# Patch an existing devsecops-semaphore LXC: DB sslmode + optional base64 secret regen.
# Run on the Incus host (WSL/Linux).
#
# Usage:
#   ./scripts/fix-semaphore-incus-env.sh
#   REGEN_SECRETS=1 ./scripts/fix-semaphore-incus-env.sh
# Optional: set public URL Semaphore should advertise (only if you need to change SEMAPHORE_WEB_ROOT):
#   SEMAPHORE_WEB_ROOT=http://example:3001 ./scripts/fix-semaphore-incus-env.sh
set -euo pipefail

INSTANCE="${INSTANCE:-devsecops-semaphore}"
REGEN_SECRETS="${REGEN_SECRETS:-0}"

if ! command -v incus >/dev/null 2>&1; then
  echo "incus not found in PATH" >&2
  exit 1
fi

incus exec "${INSTANCE}" -- env \
  REGEN_SECRETS="${REGEN_SECRETS}" \
  SEMAPHORE_DEBUG="${SEMAPHORE_DEBUG:-0}" \
  SEMAPHORE_WEB_ROOT="${SEMAPHORE_WEB_ROOT:-}" \
  bash -s <<'EOS'
set -euo pipefail
ENV=/etc/semaphore/semaphore.env
test -f "${ENV}"

if [[ -n "${SEMAPHORE_WEB_ROOT:-}" ]]; then
  echo "=== SEMAPHORE_WEB_ROOT (must match browser URL: scheme + host + port) ==="
  sed -i '/^SEMAPHORE_WEB_ROOT=/d' "${ENV}"
  echo "SEMAPHORE_WEB_ROOT=${SEMAPHORE_WEB_ROOT}" >> "${ENV}"
  chmod 0600 "${ENV}"
fi

echo "=== Ensuring PostgreSQL pg_hba allows password auth on 127.0.0.1 (first match wins) ==="
PG_HBA=/var/lib/pgsql/data/pg_hba.conf
if [[ -f "${PG_HBA}" ]]; then
  PG_HBA_MARKER="BEGIN semaphore-install: tcp localhost password auth"
  if ! grep -q "${PG_HBA_MARKER}" "${PG_HBA}" 2>/dev/null; then
    tmp="$(mktemp)"
    {
      echo "# ${PG_HBA_MARKER}"
      echo "# Prepend so these lines take effect before any host ... ident rules."
      echo "host all all 127.0.0.1/32 scram-sha-256"
      echo "host all all ::1/128 scram-sha-256"
      echo "# END semaphore-install"
      cat "${PG_HBA}"
    } > "${tmp}"
    mv -f "${tmp}" "${PG_HBA}"
    chown postgres:postgres "${PG_HBA}"
    chmod 0600 "${PG_HBA}"
    systemctl restart postgresql
  fi
fi

echo "=== Patching ${ENV} (sslmode=disable) ==="
sed -i '/^SEMAPHORE_DB_OPTIONS=/d' "${ENV}"
echo 'SEMAPHORE_DB_OPTIONS={"sslmode":"disable"}' >> "${ENV}"
if ! grep -q '^SEMAPHORE_LOG_LEVEL=' "${ENV}"; then
  echo 'SEMAPHORE_LOG_LEVEL=info' >> "${ENV}"
fi

if [[ "${REGEN_SECRETS}" == "1" ]]; then
  echo "=== Regenerating cookie/access encryption (base64) ==="
  for k in SEMAPHORE_ACCESS_KEY_ENCRYPTION SEMAPHORE_COOKIE_HASH SEMAPHORE_COOKIE_ENCRYPTION; do
    val="$(openssl rand -base64 32 | tr -d '\n\r')"
    sed -i "/^${k}=/d" "${ENV}"
    echo "${k}=${val}" >> "${ENV}"
  done
fi

chmod 0600 "${ENV}"

if [[ "${SEMAPHORE_DEBUG}" == "1" ]]; then
  echo "=== SEMAPHORE_LOG_LEVEL=debug (SEMAPHORE_DEBUG=1) ==="
  sed -i '/^SEMAPHORE_LOG_LEVEL=/d' "${ENV}"
  echo 'SEMAPHORE_LOG_LEVEL=debug' >> "${ENV}"
  chmod 0600 "${ENV}"
fi

echo "=== systemd: relax start limits (override LXC drop-ins) ==="
mkdir -p /etc/systemd/system/semaphore.service.d
cat >/etc/systemd/system/semaphore.service.d/10-startlimits.conf <<'DROPIN'
[Service]
StartLimitIntervalSec=0
RestartSec=5
DROPIN

echo "=== systemd: use env-only mode (--no-config) ==="
# Without this, Semaphore exits: "Cannot Find configuration! ... semaphore setup"
SEM_BIN="$(command -v semaphore 2>/dev/null || true)"
if [[ -z "${SEM_BIN}" && -x /usr/local/bin/semaphore ]]; then
  SEM_BIN=/usr/local/bin/semaphore
fi
if [[ -z "${SEM_BIN}" ]]; then
  echo "ERROR: semaphore binary not found in PATH" >&2
  exit 1
fi
cat >/etc/systemd/system/semaphore.service.d/20-no-config.conf <<EOF
[Service]
ExecStart=
ExecStart=${SEM_BIN} server --no-config
EOF

echo "=== Testing psql (PGSSLMODE=disable) ==="
# Do not `source` semaphore.env — systemd format is not guaranteed bash-safe (JSON values).
SEMAPHORE_DB_PASS="$(grep '^SEMAPHORE_DB_PASS=' "${ENV}" | sed 's/^SEMAPHORE_DB_PASS=//')"
SEMAPHORE_DB_USER="$(grep '^SEMAPHORE_DB_USER=' "${ENV}" | sed 's/^SEMAPHORE_DB_USER=//')"
SEMAPHORE_DB="$(grep '^SEMAPHORE_DB=' "${ENV}" | sed 's/^SEMAPHORE_DB=//')"
export PGPASSWORD="${SEMAPHORE_DB_PASS}"
export PGSSLMODE=disable
psql -h 127.0.0.1 -p 5432 -U "${SEMAPHORE_DB_USER}" -d "${SEMAPHORE_DB}" -c "SELECT 1" >/dev/null

systemctl daemon-reload
systemctl reset-failed semaphore 2>/dev/null || true
systemctl restart semaphore || true
sleep 3

if systemctl is-active --quiet semaphore; then
  echo "=== semaphore.service is active ==="
  systemctl --no-pager --full status semaphore || true
else
  echo "=== semaphore.service is NOT active — recent logs ==="
  systemctl --no-pager --full status semaphore || true
  journalctl -u semaphore -b -n 200 --no-pager || true
  echo ""
  echo "=== Short foreground sample (timeout 6s; ignore exit 124) ==="
  if command -v systemd-run >/dev/null 2>&1; then
    timeout 6s systemd-run --wait --pty --uid=semaphore --gid=semaphore \
      --property=EnvironmentFile=/etc/semaphore/semaphore.env \
      --property=WorkingDirectory=/var/lib/semaphore \
      "${SEM_BIN}" server --no-config 2>&1 | tail -n 80 || true
  fi
fi

# Native Semaphore does not auto-create admin from env (unlike Docker). Idempotent bootstrap.
SEM_BIN="${SEM_BIN:-}"
if [[ -z "${SEM_BIN}" ]] && [[ -x /usr/local/bin/semaphore ]]; then
  SEM_BIN=/usr/local/bin/semaphore
fi
if [[ -z "${SEM_BIN}" ]]; then
  SEM_BIN="$(command -v semaphore 2>/dev/null || true)"
fi
ADMIN_LOGIN="${ADMIN_LOGIN:-admin}"
ADMIN_EMAIL="${ADMIN_EMAIL:-admin@local.dev}"
ADMIN_NAME="${ADMIN_NAME:-Local Admin}"
ADMIN_PASS="${ADMIN_PASS:-admin_local_dev}"
if command -v systemd-run >/dev/null 2>&1 && [[ -f /etc/semaphore/semaphore.env ]] && systemctl is-active --quiet semaphore 2>/dev/null && [[ -n "${SEM_BIN}" ]]; then
  if systemd-run --pipe --wait -p User=semaphore -p Group=semaphore -p WorkingDirectory=/var/lib/semaphore \
      -p EnvironmentFile=/etc/semaphore/semaphore.env \
      "${SEM_BIN}" user list --no-config 2>/dev/null | grep -Fxq "${ADMIN_LOGIN}"; then
    echo "=== Semaphore admin user '${ADMIN_LOGIN}' already exists ==="
  else
    echo "=== Creating Semaphore admin user '${ADMIN_LOGIN}' (native CLI; set ADMIN_LOGIN/ADMIN_PASS to override) ==="
    systemd-run --pipe --wait -p User=semaphore -p Group=semaphore -p WorkingDirectory=/var/lib/semaphore \
      -p EnvironmentFile=/etc/semaphore/semaphore.env \
      "${SEM_BIN}" user add --no-config --admin \
      --login "${ADMIN_LOGIN}" --email "${ADMIN_EMAIL}" --name "${ADMIN_NAME}" --password "${ADMIN_PASS}" \
      || echo "WARN: user add failed — user may already exist or DB error; check journal." >&2
  fi
fi
EOS

echo "Tip: re-run with more logging: SEMAPHORE_DEBUG=1 bash fix-semaphore-incus-env.sh"
echo "Or: incus exec ${INSTANCE} -- journalctl -u semaphore -b -n 200 --no-pager"
