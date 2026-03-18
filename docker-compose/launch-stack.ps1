# Launch greenfield DevSecOps stack (messaging -> IAM -> tooling).
# With Varlock: export secrets from Vault into the environment, then run this script (no .env needed).
# Without Vault: copy .env.example to ..\.env and set required values.
$ErrorActionPreference = "Stop"
$envFile = "..\.env"
$envFileArg = @()
if (Test-Path $envFile) {
    $envFileArg = @("--env-file", $envFile)
} else {
    Write-Host "No ..\.env found; using current environment (export from Vault per VARLOCK_USAGE.md if using Varlock)."
}
Write-Host "Starting messaging backbone..."
docker compose -f docker-compose.messaging.yml @envFileArg up -d
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "Starting IAM (Keycloak)..."
docker compose -f docker-compose.iam.yml @envFileArg up -d
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "Starting tooling (Gitea, n8n, Zammad)..."
docker compose -f docker-compose.tooling.yml @envFileArg up -d
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "Stack started. Check: docker ps"
