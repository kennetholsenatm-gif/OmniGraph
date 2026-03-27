<#
.SYNOPSIS
  Start the pipeline stack using secrets from Vault only (no .env). Reads VAULT_TOKEN from env or OS keystore.
.DESCRIPTION
  Gets the Vault token from environment (VAULT_TOKEN or VAULT_DEV_ROOT_TOKEN_ID), or from Windows Credential Manager
  (target "devsecops-vault-token") or PowerShell SecretStore ("devsecops-vault-token") if available. Fetches
  secret/devsecops from Vault, exports all keys to the current process environment, then runs launch-stack.ps1.
  No static file; the only "initial" value is the token (env or keystore).
.PARAMETER VaultAddr
  Vault address (default http://127.0.0.1:8200 or env VAULT_ADDR).
.PARAMETER IncludeSdnTelemetry
  Pass to docker-compose/launch-stack.ps1: prepend SDN + telemetry compose files to the merged core stack.
  Same effect as env DEVSECOPS_INCLUDE_SDN_TELEMETRY=1.
.PARAMETER SkipLlm
  Pass to launch-stack.ps1 (default is to start LLM compose unless DEVSECOPS_INCLUDE_LLM=0).
.PARAMETER IncludeDiscovery / IncludeAiOrchestration / IncludeIdentity / IncludeSiem
  Pass through to launch-stack.ps1. Same effect as env DEVSECOPS_INCLUDE_DISCOVERY=1, DEVSECOPS_INCLUDE_AI_ORCHESTRATION=1, DEVSECOPS_INCLUDE_IDENTITY=1, DEVSECOPS_INCLUDE_SIEM=1.
#>
param(
    [string]$VaultAddr = $env:VAULT_ADDR,
    [switch]$IncludeSdnTelemetry = $false,
    [switch]$SkipLlm = $false,
    [switch]$IncludeDiscovery = $false,
    [switch]$IncludeAiOrchestration = $false,
    [switch]$IncludeIdentity = $false,
    [switch]$IncludeSiem = $false
)
if (-not $VaultAddr) { $VaultAddr = "http://127.0.0.1:8200" }

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir
$composeDir = Join-Path $pipelineRoot "docker-compose"

# --- Get Vault token: env first, then OS keystore ---
$token = $env:VAULT_TOKEN
if (-not $token) { $token = $env:VAULT_DEV_ROOT_TOKEN_ID }

if (-not $token) {
    try {
        $cm = Get-Module -ListAvailable -Name CredentialManager
        if ($cm) {
            $cred = Get-StoredCredential -Target "devsecops-vault-token" -ErrorAction SilentlyContinue
            if ($cred) { $token = $cred.GetNetworkCredential().Password }
        }
    } catch {}
}
if (-not $token) {
    try {
        $secret = Get-Secret -Name "devsecops-vault-token" -Vault "SecretStore" -ErrorAction SilentlyContinue
        if ($secret -is [SecureString]) { $token = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto([System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secret)) }
        elseif ($secret) { $token = [string]$secret }
    } catch {}
}

if (-not $token) {
    Write-Host "VAULT_TOKEN (or VAULT_DEV_ROOT_TOKEN_ID) not set and not found in keystore."
    Write-Host "After first run of secrets-bootstrap.ps1, run: .\save-vault-token-to-keystore.ps1"
    Write-Host "Or set env: `$env:VAULT_TOKEN = '<root-token>'"
    exit 1
}

# --- Fetch secret/devsecops from Vault ---
$headers = @{ "X-Vault-Token" = $token }
$uri = "$VaultAddr/v1/secret/data/devsecops"
try {
    $r = Invoke-RestMethod -Uri $uri -Headers $headers -Method Get -ErrorAction Stop
} catch {
    try {
        $uriV1 = "$VaultAddr/v1/secret/devsecops"
        $r = Invoke-RestMethod -Uri $uriV1 -Headers $headers -Method Get -ErrorAction Stop
        $data = $r
    } catch {
        Write-Error "Failed to read from Vault: $_"
        exit 1
    }
}
if (-not $data) { $data = $r.data.data }

foreach ($k in $data.PSObject.Properties.Name) {
    $v = $data.$k
    if ($null -ne $v -and $v -isnot [System.Collections.IDictionary]) {
        Set-Item -Path "Env:$k" -Value $v
    }
}

Write-Host "Exported pipeline secrets from Vault to environment. Starting stack..."
$launch = Join-Path $composeDir "launch-stack.ps1"
$launchArgs = @{}
if ($IncludeSdnTelemetry -or ($env:DEVSECOPS_INCLUDE_SDN_TELEMETRY -eq "1")) { $launchArgs["IncludeSdnTelemetry"] = $true }
if ($SkipLlm) { $launchArgs["SkipLlm"] = $true }
if ($IncludeDiscovery) { $launchArgs["IncludeDiscovery"] = $true }
if ($IncludeAiOrchestration) { $launchArgs["IncludeAiOrchestration"] = $true }
if ($IncludeIdentity) { $launchArgs["IncludeIdentity"] = $true }
if ($IncludeSiem -or ($env:DEVSECOPS_INCLUDE_SIEM -eq "1")) { $launchArgs["IncludeSiem"] = $true }
Push-Location $composeDir
try {
    & $launch @launchArgs
} finally {
    Pop-Location
}
