#!/usr/bin/env bash
# Idempotent: create SonarQube DB role and database on messaging Postgres (msg_backbone_net).
set -euo pipefail
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD required}"
if [ -z "${SONAR_JDBC_PASSWORD:-}" ]; then
  echo "SONAR_JDBC_PASSWORD unset; skipping SonarQube DB bootstrap."
  exit 0
fi
: "${PGHOST:=postgres}"
: "${POSTGRES_USER:=qminiwasm}"
USER_NAME="${SONAR_JDBC_USERNAME:-sonar}"
DB_NAME="${SONAR_JDBC_DB:-sonar}"
export PGPASSWORD="${POSTGRES_PASSWORD}"

until pg_isready -h "$PGHOST" -U "$POSTGRES_USER" -d postgres; do
  echo "Waiting for Postgres at $PGHOST..."
  sleep 2
done

esc_sql() { printf "%s" "$1" | sed "s/'/''/g"; }
PASS_ESC=$(esc_sql "$SONAR_JDBC_PASSWORD")

if psql -h "$PGHOST" -U "$POSTGRES_USER" -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='${USER_NAME}'" | grep -q 1; then
  psql -h "$PGHOST" -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 -c "ALTER ROLE \"${USER_NAME}\" WITH LOGIN PASSWORD '${PASS_ESC}';"
else
  psql -h "$PGHOST" -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 -c "CREATE ROLE \"${USER_NAME}\" LOGIN PASSWORD '${PASS_ESC}';"
fi

if psql -h "$PGHOST" -U "$POSTGRES_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" | grep -q 1; then
  psql -h "$PGHOST" -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 -c "ALTER DATABASE \"${DB_NAME}\" OWNER TO \"${USER_NAME}\";" || true
else
  psql -h "$PGHOST" -U "$POSTGRES_USER" -d postgres -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"${DB_NAME}\" OWNER \"${USER_NAME}\";"
fi

echo "SonarQube database '${DB_NAME}' and role '${USER_NAME}' are ready."
