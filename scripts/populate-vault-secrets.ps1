# Populate Vault with devsecops secrets from .dev/vault-payload-kbolsen.json (no CLI required).
# Run after Vault container is up. Uses VAULT_ADDR and VAULT_TOKEN (default devsecops-dev-root).
# Then open http://localhost:8200/ui -> Secrets -> secret -> devsecops.

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir
$payloadPath = Join-Path (Join-Path $pipelineRoot ".dev") "vault-payload-kbolsen.json"

$vaultAddr = $env:VAULT_ADDR
if (-not $vaultAddr) { $vaultAddr = "http://localhost:8200" }
$vaultToken = $env:VAULT_TOKEN
if (-not $vaultToken) { $vaultToken = "devsecops-dev-root" }

if (-not (Test-Path $payloadPath)) {
    Write-Host "Run register-kbolsen-in-vault.ps1 first to create $payloadPath"
    exit 1
}

$payload = Get-Content $payloadPath -Raw | ConvertFrom-Json
$data = @{}
$payload.PSObject.Properties | ForEach-Object { $data[$_.Name] = $_.Value }

$headers = @{ "X-Vault-Token" = $vaultToken; "Content-Type" = "application/json" }
$mount = "secret"
$secretPath = "devsecops"

# Try KV v2 first: POST /v1/secret/data/devsecops
$uriV2 = "$vaultAddr/v1/$mount/data/$secretPath"
$bodyV2 = @{ data = $data } | ConvertTo-Json -Depth 10
try {
    Invoke-RestMethod -Uri $uriV2 -Method Post -Headers $headers -Body $bodyV2 | Out-Null
    Write-Host "Secrets written to Vault (KV v2) at $mount/$secretPath"
    Write-Host "Open $vaultAddr/ui -> Secrets -> secret -> devsecops"
    exit 0
} catch {
    $err = $_.Exception.Message
    if ($_.ErrorDetails.Message) { $err += " " + $_.ErrorDetails.Message }
}
# Fallback: KV v1
$uriV1 = "$vaultAddr/v1/$mount/$secretPath"
$bodyV1 = $data | ConvertTo-Json -Depth 10
try {
    Invoke-RestMethod -Uri $uriV1 -Method Post -Headers $headers -Body $bodyV1 | Out-Null
    Write-Host "Secrets written to Vault (KV v1) at $mount/$secretPath"
    Write-Host "Open $vaultAddr/ui -> Secrets -> secret -> devsecops"
    exit 0
} catch {
    Write-Host "KV v2 error: $err"
    Write-Host "KV v1 error: $($_.Exception.Message)"
    Write-Host "Ensure Vault is running (docker ps) and VAULT_TOKEN=$vaultToken"
    exit 1
}
