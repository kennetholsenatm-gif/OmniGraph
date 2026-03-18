<#
.SYNOPSIS
  Store product license/activation keys in Vault (KMS) without writing them to any file.
.DESCRIPTION
  Prompts for n8n activation key (or uses N8N_LICENSE_ACTIVATION_KEY from env if set), then writes
  or patches secret/devsecops in Vault so that start-from-vault.ps1 and the stack can inject the key
  at runtime. No key is written to disk.
.PARAMETER VaultAddr
  Vault address (default from env VAULT_ADDR or http://127.0.0.1:8200).
.PARAMETER VaultToken
  Vault token (default from env VAULT_TOKEN or VAULT_DEV_ROOT_TOKEN_ID).
#>
param(
    [string]$VaultAddr = $env:VAULT_ADDR,
    [string]$VaultToken = $env:VAULT_TOKEN
)
if (-not $VaultAddr) { $VaultAddr = "http://127.0.0.1:8200" }
if (-not $VaultToken) { $VaultToken = $env:VAULT_DEV_ROOT_TOKEN_ID }

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir

# --- Resolve token from keystore if not in env ---
if (-not $VaultToken) {
    try {
        $cm = Get-Module -ListAvailable -Name CredentialManager
        if ($cm) {
            $cred = Get-StoredCredential -Target "devsecops-vault-token" -ErrorAction SilentlyContinue
            if ($cred) { $VaultToken = $cred.GetNetworkCredential().Password }
        }
    } catch {}
}
if (-not $VaultToken) {
    try {
        $secret = Get-Secret -Name "devsecops-vault-token" -Vault "SecretStore" -ErrorAction SilentlyContinue
        if ($secret -is [SecureString]) { $VaultToken = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto([System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secret)) }
    } catch {}
}

if (-not $VaultToken) {
    Write-Error "Set VAULT_TOKEN or run save-vault-token-to-keystore.ps1 after bootstrap."
    exit 1
}

$headers = @{ "X-Vault-Token" = $VaultToken; "Content-Type" = "application/json" }

# --- Read existing secret so we don't overwrite other keys ---
$uriRead = "$VaultAddr/v1/secret/data/devsecops"
$data = @{}
try {
    $r = Invoke-RestMethod -Uri $uriRead -Headers $headers -Method Get -ErrorAction Stop
    if ($r.data -and $r.data.data) {
        $r.data.data.PSObject.Properties | ForEach-Object { $data[$_.Name] = $_.Value }
    }
} catch {
    try {
        $r = Invoke-RestMethod -Uri "$VaultAddr/v1/secret/devsecops" -Headers $headers -Method Get -ErrorAction Stop
        $r.PSObject.Properties | ForEach-Object { $data[$_.Name] = $_.Value }
    } catch {}
}

# --- n8n activation key ---
$n8nKey = $env:N8N_LICENSE_ACTIVATION_KEY
if (-not $n8nKey) {
    $prompt = Read-Host "Enter n8n license activation key (or leave blank to skip)"
    $n8nKey = $prompt.Trim()
}
if ($n8nKey) {
    $data["N8N_LICENSE_ACTIVATION_KEY"] = $n8nKey
    Write-Host "Will store N8N_LICENSE_ACTIVATION_KEY in Vault (value not echoed)."
}

if ($data.Count -eq 0) {
    Write-Host "No keys to write."
    exit 0
}

# --- Write back (KV v2) ---
$uriWrite = "$VaultAddr/v1/secret/data/devsecops"
$body = @{ data = $data } | ConvertTo-Json -Depth 10
try {
    Invoke-RestMethod -Uri $uriWrite -Method Post -Headers $headers -Body $body -ErrorAction Stop | Out-Null
    Write-Host "License key(s) stored in Vault at secret/devsecops."
    Write-Host "Run start-from-vault.ps1 (or restart stack with env from Vault) so n8n receives N8N_LICENSE_ACTIVATION_KEY."
} catch {
    try {
        $uriV1 = "$VaultAddr/v1/secret/devsecops"
        Invoke-RestMethod -Uri $uriV1 -Method Post -Headers $headers -Body ($data | ConvertTo-Json -Depth 10) -ErrorAction Stop | Out-Null
        Write-Host "License key(s) stored in Vault (KV v1) at secret/devsecops."
    } catch {
        Write-Error "Vault write failed: $_"
        exit 1
    }
}
