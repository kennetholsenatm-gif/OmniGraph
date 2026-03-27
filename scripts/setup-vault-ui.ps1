# Use the root token to set up Vault so http://localhost:8200/ui/vault/secrets/ shows the KV store.
# Ensures KV v2 is enabled at path "secret" (for secret/devsecops). Run after Vault container is up.
# Usage: set VAULT_ADDR and VAULT_TOKEN, then .\scripts\setup-vault-ui.ps1

$ErrorActionPreference = "Stop"
$vaultAddr = $env:VAULT_ADDR
if (-not $vaultAddr) { $vaultAddr = "http://localhost:8200" }
$vaultToken = $env:VAULT_TOKEN
if (-not $vaultToken) { $vaultToken = "devsecops-dev-root" }

$headers = @{
    "X-Vault-Token" = $vaultToken
    "Content-Type"  = "application/json"
}
$body = @{ type = "kv"; options = @{ version = "2" } } | ConvertTo-Json

Write-Host "Vault address: $vaultAddr"
Write-Host "Enabling KV v2 at path 'secret'..."

try {
    $response = Invoke-RestMethod -Uri "$vaultAddr/v1/sys/mounts/secret" -Method Post -Headers $headers -Body $body
    Write-Host "KV v2 enabled at secret/."
} catch {
    if ($_.Exception.Response.StatusCode -eq 400 -or $_.ErrorDetails.Message -match "already in use") {
        Write-Host "Mount 'secret' already exists (OK)."
    } else {
        Write-Host "Error: $_"
        exit 1
    }
}

Write-Host ""

# Populate secret/devsecops from payload if it exists (so the UI shows secrets)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir
$payloadPath = Join-Path (Join-Path $pipelineRoot ".dev") "vault-payload-kbolsen.json"
if (Test-Path $payloadPath) {
    Write-Host "Populating secret/devsecops from payload..."
    & (Join-Path $scriptDir "populate-vault-secrets.ps1")
} else {
    Write-Host "No payload at $payloadPath. Run scripts\register-kbolsen-in-vault.ps1 first, then re-run this script or run scripts\populate-vault-secrets.ps1"
}

Write-Host ""
Write-Host "Vault UI:"
Write-Host "  1. Open: $vaultAddr/ui"
Write-Host "  2. Sign in with token: $vaultToken"
Write-Host "  3. Go to: Secrets -> secret -> devsecops (to view KEYCLOAK_ADMIN, admin_ssh_public_key, etc.)"
