<#
.SYNOPSIS
  Greenfield secrets bootstrap: generate strong random secrets, inject into Vault, start stack with env-only (no .env or static files).
.DESCRIPTION
  Generates cryptographically random values for all pipeline secrets, sets them in the current process
  environment, starts the Docker Compose stacks (so containers get them via env), then writes the same
  secrets to Vault at secret/devsecops for Varlock and other consumers. Secrets are never written to
  disk; only in memory and in Vault.
.PARAMETER StartStack
  If set (default), start IAM then messaging then tooling after generating secrets. If not set, only
  inject into Vault (assumes Vault is already running).
.PARAMETER VaultAddr
  Vault address (default http://127.0.0.1:8200).
.PARAMETER KeycloakAdminUsername
  Keycloak admin username (default admin).
.EXAMPLE
  .\secrets-bootstrap.ps1
  Generate secrets, set env, start full stack, then push secrets to Vault.
.EXAMPLE
  .\secrets-bootstrap.ps1 -StartStack:$false
  Only push currently generated or existing env secrets to Vault (Vault must be up).
#>
param(
    [switch]$StartStack = $true,
    [string]$VaultAddr = "http://127.0.0.1:8200",
    [string]$KeycloakAdminUsername = "admin"
)

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir
$composeDir = Join-Path $pipelineRoot "docker-compose"

function New-RandomSecret {
    param([int]$ByteLength = 32, [switch]$Base64Safe)
    $bytes = New-Object byte[] $ByteLength
    [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
    if ($Base64Safe) {
        [Convert]::ToBase64String($bytes) -replace '[+/=]', ''
    } else {
        [Convert]::ToBase64String($bytes)
    }
}

function New-StrongPassword {
    param([int]$Length = 24)
    $raw = New-RandomSecret -ByteLength ([Math]::Max($Length, 24)) -Base64Safe
    $raw.Substring(0, [Math]::Min($Length, $raw.Length))
}

function New-Token {
    param([int]$Length = 32)
    New-RandomSecret -ByteLength ([Math]::Max([int]($Length / 2), 16)) -Base64Safe | ForEach-Object { $_.Substring(0, [Math]::Min($Length, $_.Length)) }
}

# Ensure KV v2 at secret/ (idempotent). Dev mode often has secret/ already; 400 = exists.
function Enable-VaultKv2 {
    param([string]$Addr, [string]$Token)
    $headers = @{ "X-Vault-Token" = $Token; "Content-Type" = "application/json" }
    $body = '{"type":"kv","options":{"version":"2"}}'
    try {
        Invoke-RestMethod -Uri "$Addr/v1/sys/mounts/secret" -Method Post -Headers $headers -Body $body -ErrorAction Stop
        return $true
    } catch {
        $code = $_.Exception.Response.StatusCode.value__
        if ($code -eq 400) { return $true }
        try {
            Invoke-RestMethod -Uri "$Addr/v1/sys/mounts/secret" -Method Get -Headers @{ "X-Vault-Token" = $Token } -ErrorAction Stop
            return $true
        } catch {}
        return $false
    }
}

function Write-VaultSecrets {
    param([hashtable]$Data, [string]$Addr, [string]$Token, [string]$Mount = "secret", [string]$Path = "devsecops")
    $headers = @{ "X-Vault-Token" = $Token; "Content-Type" = "application/json" }
    $uri = "$Addr/v1/$Mount/data/$Path"
    $body = @{ data = $Data } | ConvertTo-Json -Depth 10
    try {
        Invoke-RestMethod -Uri $uri -Method Post -Headers $headers -Body $body | Out-Null
        return $true
    } catch {
        $uriV1 = "$Addr/v1/$Mount/$Path"
        try {
            $bodyV1 = $Data | ConvertTo-Json -Depth 10
            Invoke-RestMethod -Uri $uriV1 -Method Post -Headers $headers -Body $bodyV1 | Out-Null
            return $true
        } catch {
            Write-Warning "Vault write failed: $_"
            return $false
        }
    }
}

# --- Generate or reuse from env (no disk) ---
$secretKeys = @(
    "KEYCLOAK_DB_PASSWORD", "KEYCLOAK_ADMIN_PASSWORD", "VAULT_DEV_ROOT_TOKEN_ID",
    "ZAMMAD_POSTGRES_PASSWORD", "GITEA_API_TOKEN", "SOLACE_PASSWORD", "SOLACE_ADMIN_PASSWORD",
    "ZAMMAD_API_TOKEN", "WEBHOOK_HMAC_SECRET", "N8N_API_TOKEN", "TELEPORT_API_TOKEN",
    "POSTGRES_PASSWORD", "RABBITMQ_DEFAULT_PASS"
)
$secrets = @{}
foreach ($k in $secretKeys) {
    $v = Get-Item -Path "Env:$k" -ErrorAction SilentlyContinue
    if ($v -and $v.Value) {
        $secrets[$k] = $v.Value
    } else {
        if ($k -match "PASSWORD") { $secrets[$k] = New-StrongPassword -Length 24 }
        else { $secrets[$k] = New-Token -Length 32 }
    }
}
$secrets["KEYCLOAK_ADMIN"] = if ($env:KEYCLOAK_ADMIN) { $env:KEYCLOAK_ADMIN } else { $KeycloakAdminUsername }

Write-Host "Secrets prepared (from env or generated); setting in environment."

# Export to environment so compose can substitute (no .env file)
foreach ($k in $secrets.Keys) {
    Set-Item -Path "Env:$k" -Value $secrets[$k]
}

Write-Host "Secrets set in environment (KEYCLOAK_ADMIN=$($secrets['KEYCLOAK_ADMIN']))."

if ($StartStack) {
    if (-not (Test-Path $composeDir)) {
        Write-Error "Compose directory not found: $composeDir"
        exit 1
    }
    Push-Location $composeDir
    try {
        Write-Host "Starting IAM stack (Vault, Keycloak, proxy)..."
        docker compose -f docker-compose.iam.yml up -d
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        Write-Host "Starting messaging stack..."
        docker compose -f docker-compose.messaging.yml up -d
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        Write-Host "Starting tooling stack..."
        docker compose -f docker-compose.tooling.yml up -d
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        Write-Host "Waiting for Vault to be ready (15s)..."
        Start-Sleep -Seconds 15
    } finally {
        Pop-Location
    }
}

# Wait for Vault and inject (use VAULT_TOKEN if set, e.g. for inject-only when Vault already running)
$token = if ($env:VAULT_TOKEN) { $env:VAULT_TOKEN } else { $secrets["VAULT_DEV_ROOT_TOKEN_ID"] }
$maxAttempts = 12
$attempt = 0
$vaultOk = $false
while ($attempt -lt $maxAttempts) {
    try {
        $r = Invoke-RestMethod -Uri "$VaultAddr/v1/sys/health" -Method Get -TimeoutSec 5 -ErrorAction Stop
        $vaultOk = $true
        break
    } catch {
        $attempt++
        if ($attempt -ge $maxAttempts) {
            Write-Warning "Vault not reachable at $VaultAddr after $maxAttempts attempts. Push secrets later with: `$env:VAULT_TOKEN='...'; .\secrets-bootstrap.ps1 -StartStack:`$false"
            exit 0
        }
        Start-Sleep -Seconds 5
    }
}

if (-not $vaultOk) { exit 1 }

if (-not (Enable-VaultKv2 -Addr $VaultAddr -Token $token)) {
    Write-Warning "Could not enable KV v2 at secret/; continuing anyway."
}

$written = Write-VaultSecrets -Data $secrets -Addr $VaultAddr -Token $token
if ($written) {
    Write-Host "Secrets written to Vault at secret/devsecops (no static file)."
    Write-Host "Vault UI: $VaultAddr (token in env VAULT_DEV_ROOT_TOKEN_ID for this session only)."
} else {
    Write-Warning "Vault write failed. Env vars are set for this session; restart stack with same env or re-run this script."
}

Write-Host ""
Write-Host "Keycloak admin: $KeycloakAdminUsername / (password was generated; in Vault at secret/devsecops or in this session env)."
Write-Host "Use http://127.0.0.1:8180/admin to log in."
# Optional: set master realm to allow HTTP (avoids "HTTPS required" when using proxy)
if ($StartStack) {
    Write-Host "Waiting for Keycloak to init (60s) then setting master realm ssl_required=NONE..."
    Start-Sleep -Seconds 60
    docker exec devsecops-keycloak-db psql -U keycloak -d keycloak -c "UPDATE realm SET ssl_required = 'NONE' WHERE name = 'master';" 2>$null
    if ($LASTEXITCODE -eq 0) { Write-Host "Master realm set to allow HTTP." } else { Write-Host "If you see HTTPS required, run: docker exec devsecops-keycloak-db psql -U keycloak -d keycloak -c \"UPDATE realm SET ssl_required = 'NONE' WHERE name = 'master';\"" }
}

Write-Host ""
Write-Host "To avoid storing anything on disk, run this script on each host startup and start the stack from the same shell, or export env from Vault before starting compose."
