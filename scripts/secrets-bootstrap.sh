#!/usr/bin/env bash
#
# Bash twin of secrets-bootstrap.ps1: generate secrets, export to environment,
# optionally start core Docker Compose (stack-manifest coreStack + optional SDN),
# push to Vault KV v2 at secret/data/devsecops.
# Default: no docker-compose/.env (--write-env-file to opt in).
#
# Break-glass Keycloak user + Bitwarden (bw): use secrets-bootstrap.ps1 or install pwsh in LXC.
#
set -euo pipefail

START_STACK=1
WRITE_ENV=0
INCLUDE_SDN=0
VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"

usage() {
  echo "Usage: $0 [--no-start] [--write-env-file] [--include-sdn-telemetry] [--vault-addr URL]"
  echo "Env: DEVSECOPS_REPO_ROOT (optional path to repo root)"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-start) START_STACK=0 ;;
    --write-env-file) WRITE_ENV=1 ;;
    --include-sdn-telemetry) INCLUDE_SDN=1 ;;
    --vault-addr) VAULT_ADDR="$2"; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
  shift
done

command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }
command -v openssl >/dev/null || { echo "openssl is required" >&2; exit 1; }

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
if [[ -n "${DEVSECOPS_REPO_ROOT:-}" ]]; then
  REPO_ROOT=$(cd "$DEVSECOPS_REPO_ROOT" && pwd)
else
  REPO_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
fi
COMPOSE_DIR="$REPO_ROOT/docker-compose"
MANIFEST="$COMPOSE_DIR/stack-manifest.json"

rand_b64() { openssl rand -base64 48 | tr -d '/+=' | head -c "${1:-32}"; }
rand_pw() { openssl rand -base64 36 | tr -d '/+=' | head -c 24; }

SECRET_KEYS=(
  KEYCLOAK_DB_PASSWORD KEYCLOAK_ADMIN_PASSWORD VAULT_DEV_ROOT_TOKEN_ID
  ZAMMAD_POSTGRES_PASSWORD GITEA_API_TOKEN SOLACE_PASSWORD SOLACE_ADMIN_PASSWORD
  ZAMMAD_API_TOKEN WEBHOOK_HMAC_SECRET N8N_API_TOKEN TELEPORT_API_TOKEN
  POSTGRES_PASSWORD RABBITMQ_DEFAULT_PASS BITWARDEN_ADMIN_TOKEN GATEWAY_REFRESH_SECRET
  ZULIP_POSTGRES_PASSWORD ZULIP_MEMCACHED_PASSWORD ZULIP_RABBITMQ_PASSWORD ZULIP_REDIS_PASSWORD
  ZULIP_SECRET_KEY ZULIP_EMAIL_PASSWORD ZULIP_OIDC_CLIENT_SECRET
  NETBOX_DB_PASSWORD NETBOX_REDIS_PASSWORD NETBOX_SECRET_KEY NETBOX_SUPERUSER_PASSWORD
  NETBOX_API_TOKEN DEP_TRACK_API_KEY TERMIUS_API_TOKEN
  VYOS_USER_PASSWORD VYOS_ENROLL_KEY GRAFANA_ADMIN_PASSWORD GRAFANA_OIDC_CLIENT_SECRET SFLOW_RT_ADMIN_TOKEN
  SONAR_JDBC_PASSWORD SONARQUBE_OIDC_CLIENT_SECRET
)

declare -A SECRETS
for k in "${SECRET_KEYS[@]}"; do
  if [[ -n "${!k:-}" ]]; then
    SECRETS[$k]="${!k}"
  elif [[ "$k" =~ PASSWORD ]]; then
    SECRETS[$k]="$(rand_pw)"
  else
    SECRETS[$k]="$(rand_b64 32)"
  fi
done

KEYCLOAK_ADMIN="${KEYCLOAK_ADMIN:-admin}"
SECRETS[KEYCLOAK_ADMIN]="$KEYCLOAK_ADMIN"
SECRETS[ZULIP_ADMINISTRATOR]="${ZULIP_ADMINISTRATOR:-admin@breakglass.local}"

export KEYCLOAK_PUBLIC_URL="${KEYCLOAK_PUBLIC_URL:-http://127.0.0.1:8180/keycloak}"
for k in "${!SECRETS[@]}"; do
  export "$k"="${SECRETS[$k]}"
done

echo "Secrets prepared; exported to environment (zero-disk)."

if [[ "$WRITE_ENV" -eq 1 ]]; then
  ENV_FILE="$COMPOSE_DIR/.env"
  : >"$ENV_FILE"
  for k in "${!SECRETS[@]}"; do
    v="${SECRETS[$k]}"
    if [[ "$v" =~ [[:space:]\"'#] ]]; then
      printf '%s=%q\n' "$k" "$v" >>"$ENV_FILE"
    else
      printf '%s=%s\n' "$k" "$v" >>"$ENV_FILE"
    fi
  done
  echo "Also wrote $ENV_FILE (opt-in)."
fi

compose_core_files() {
  [[ -f "$MANIFEST" ]] || { echo "Missing $MANIFEST" >&2; return 1; }
  local args=()
  if [[ "$INCLUDE_SDN" -eq 1 ]]; then
    while IFS= read -r f; do args+=(-f "$COMPOSE_DIR/$f"); done < <(jq -r '.sdnTelemetry.files[]' "$MANIFEST")
  fi
  while IFS= read -r f; do args+=(-f "$COMPOSE_DIR/$f"); done < <(jq -r '.coreStack.files[]' "$MANIFEST")
  ( cd "$COMPOSE_DIR" && docker compose "${args[@]}" up -d --remove-orphans )
}

if [[ "$START_STACK" -eq 1 ]]; then
  echo "Starting core stack (docker compose)..."
  compose_core_files
  echo "Waiting 60s before Vault check..."
  sleep 60
fi

TOKEN="${VAULT_TOKEN:-${SECRETS[VAULT_DEV_ROOT_TOKEN_ID]}}"
wait_vault() {
  local i=0 max=24
  while [[ $i -lt $max ]]; do
    if curl -sS -m 5 "${VAULT_ADDR}/v1/sys/health" >/dev/null 2>&1; then
      echo "Vault is ready."
      return 0
    fi
    i=$((i + 1))
    echo "Waiting for Vault... $i/$max"
    sleep 5
  done
  return 1
}

if ! wait_vault; then
  echo "Vault not reachable at $VAULT_ADDR" >&2
  exit 1
fi

curl -sS -X POST -H "X-Vault-Token: $TOKEN" -H "Content-Type: application/json" \
  -d '{"type":"kv","options":{"version":"2"}}' "${VAULT_ADDR}/v1/sys/mounts/secret" >/dev/null 2>&1 || true

DATA_JSON="{}"
for k in "${!SECRETS[@]}"; do
  DATA_JSON=$(jq -n --argjson d "$DATA_JSON" --arg k "$k" --arg v "${SECRETS[$k]}" '$d + {($k): $v}')
done

if curl -sS -f -X POST -H "X-Vault-Token: $TOKEN" -H "Content-Type: application/json" \
  -d "$(jq -n --argjson d "$DATA_JSON" '{data: $d}')" "${VAULT_ADDR}/v1/secret/data/devsecops" >/dev/null; then
  echo "Secrets written to Vault KV v2 at secret/data/devsecops"
else
  if curl -sS -f -X POST -H "X-Vault-Token: $TOKEN" -H "Content-Type: application/json" \
    -d "$DATA_JSON" "${VAULT_ADDR}/v1/secret/devsecops" >/dev/null; then
    echo "Secrets written to Vault KV v1 at secret/devsecops"
  else
    echo "Vault write failed; secrets remain in this shell env only." >&2
    exit 1
  fi
fi

echo ""
echo "Keycloak admin: $KEYCLOAK_ADMIN (password in env / Vault KEYCLOAK_ADMIN_PASSWORD)."
echo "For break-glass + Bitwarden: run scripts/secrets-bootstrap.ps1 or install PowerShell in LXC."
echo "Inject-only: export VAULT_TOKEN; $0 --no-start"
