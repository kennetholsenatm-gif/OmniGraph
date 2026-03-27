# Reset Keycloak so the admin user is created from Vault (or .dev password file).
# Keycloak only applies KEYCLOAK_ADMIN / KEYCLOAK_ADMIN_PASSWORD on first DB init.
# This script: stops IAM stack, removes Keycloak DB volume, sets env from file, starts IAM again.
# Run from repo root (e.g. C:\GiTeaRepos\devsecops-pipeline). Requires: .dev\kbolsen_keycloak_password.txt

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir
$devDir = Join-Path $pipelineRoot ".dev"
$passwordFile = Join-Path $devDir "kbolsen_keycloak_password.txt"
$composeDir = Join-Path $pipelineRoot "docker-compose"
$envFile = Join-Path $pipelineRoot ".env"

if (-not (Test-Path $passwordFile)) {
    Write-Error "Missing $passwordFile. Run register-kbolsen-in-vault.ps1 first."
    exit 1
}

$password = (Get-Content $passwordFile -Raw).Trim()
if ([string]::IsNullOrEmpty($password)) {
    Write-Error "Password file is empty."
    exit 1
}

Write-Host "Stopping IAM stack and removing Keycloak DB volume..."
Set-Location $composeDir
docker compose -f docker-compose.iam.yml --env-file $envFile down -v
$vols = docker volume ls -q -f name=keycloak_db_data 2>$null
foreach ($v in $vols) {
    docker volume rm $v 2>$null
    if ($LASTEXITCODE -eq 0) { Write-Host "Removed volume: $v" }
}

Write-Host "Starting IAM stack with KEYCLOAK_ADMIN=kbolsen and password from file..."
$env:KEYCLOAK_ADMIN = "kbolsen"
$env:KEYCLOAK_ADMIN_PASSWORD = $password
docker compose -f docker-compose.iam.yml --env-file $envFile up -d

Write-Host "Waiting for Keycloak to create master realm (90 s)..."
Start-Sleep -Seconds 90

Write-Host "Setting master realm SSL to NONE (avoids HTTP 400 when client IP is Docker bridge)..."
docker exec devsecops-keycloak-db psql -U keycloak -d keycloak -c "UPDATE realm SET ssl_required = 'NONE' WHERE name = 'master';" 2>$null
if ($LASTEXITCODE -ne 0) { Write-Host "If you still get 400, run: docker exec devsecops-keycloak-db psql -U keycloak -d keycloak -c \"UPDATE realm SET ssl_required = 'NONE' WHERE name = 'master';\"" }

Write-Host ""
Write-Host "Keycloak is ready. Log in at http://localhost:8180/admin with:"
Write-Host "  Username: kbolsen"
Write-Host "  Password: (from $passwordFile)"
Write-Host ""
Write-Host "(Grumble: in production, inject from Vault at runtime and avoid the plain-text file.)"
