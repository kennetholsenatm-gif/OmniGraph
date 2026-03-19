<#
.SYNOPSIS
  One-command greenfield launch: create networks, run secrets bootstrap, optionally save Vault token.
.DESCRIPTION
  Ensures all Docker networks exist (create-networks.ps1), then runs secrets-bootstrap.ps1 to generate
  secrets, start the stack (IAM, messaging, tooling), and push to Vault. Optionally saves the Vault
  token to the OS keystore so start-from-vault.ps1 can start the stack on later runs without re-running bootstrap.
  Target: clone repo, run this script, enter admin password once, walk away; infrastructure is up when you return.
.PARAMETER SaveVaultToken
  After bootstrap, run save-vault-token-to-keystore.ps1 so the next run can use start-from-vault.ps1.
.PARAMETER NonInteractive
  Use BREAK_GLASS_USER and BREAK_GLASS_PASSWORD from the environment instead of prompting (for CI).
.PARAMETER StartStack
  Pass-through to secrets-bootstrap.ps1. If set (default), start the stack after generating secrets.
.PARAMETER VaultAddr
  Pass-through to secrets-bootstrap.ps1.
.PARAMETER KeycloakAdminUsername
  Pass-through to secrets-bootstrap.ps1.
.PARAMETER BreakGlassUsername
  Pass-through to secrets-bootstrap.ps1. When -NonInteractive, falls back to $env:BREAK_GLASS_USER.
.PARAMETER BreakGlassPassword
  Optional pass-through to secrets-bootstrap.ps1 (SecureString). If omitted, bootstrap generates a random password and displays it once at the end.
.PARAMETER IdentityBackend
  Pass-through to secrets-bootstrap.ps1.
.PARAMETER SkipBitwardenInject
  Pass-through to secrets-bootstrap.ps1.
.PARAMETER OnlyBreakGlass
  Pass-through to secrets-bootstrap.ps1.
.PARAMETER WriteEnvFile
  Pass-through to secrets-bootstrap.ps1 (emit docker-compose/.env; default off).
.PARAMETER IncludeSdnTelemetry
  Pass-through to secrets-bootstrap.ps1 (add network + telemetry compose on stack start).
.EXAMPLE
  .\scripts\launch-greenfield.ps1
  Create networks, run bootstrap (prompt for admin username/password), start stack. One command, one prompt.
.EXAMPLE
  .\scripts\launch-greenfield.ps1 -SaveVaultToken
  Same as above; also save Vault token so next time run .\scripts\start-from-vault.ps1.
.EXAMPLE
  $env:BREAK_GLASS_USER = "admin"; $env:BREAK_GLASS_PASSWORD = "secret"; .\scripts\launch-greenfield.ps1 -NonInteractive
  CI/non-interactive: no prompts; use env for break-glass credentials.
#>
param(
    [switch]$SaveVaultToken = $false,
    [switch]$NonInteractive = $false,
    [switch]$StartStack = $true,
    [string]$VaultAddr = "http://127.0.0.1:8200",
    [string]$KeycloakAdminUsername = "admin",
    [string]$BreakGlassUsername = "",
    [SecureString]$BreakGlassPassword = $null,
    [ValidateSet("Keycloak", "FreeIPA")]
    [string]$IdentityBackend = "Keycloak",
    [switch]$SkipBitwardenInject = $false,
    [switch]$OnlyBreakGlass = $false,
    [switch]$WriteEnvFile = $false,
    [switch]$IncludeSdnTelemetry = $false
)

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir
$scriptsDir = Join-Path $pipelineRoot "scripts"

# --- Ensure networks exist ---
Write-Host "Ensuring Docker networks..."
& (Join-Path $scriptsDir "create-networks.ps1")
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

# --- NonInteractive: fill break-glass from env when not provided ---
if ($NonInteractive) {
    if ([string]::IsNullOrWhiteSpace($BreakGlassUsername)) {
        $BreakGlassUsername = $env:BREAK_GLASS_USER
    }
    if (-not $BreakGlassPassword -and -not [string]::IsNullOrWhiteSpace($env:BREAK_GLASS_PASSWORD)) {
        $BreakGlassPassword = ConvertTo-SecureString -String $env:BREAK_GLASS_PASSWORD -AsPlainText -Force
    }
}

# --- Build bootstrap arguments (hashtable splat = explicit param names, no positional mix-up) ---
$bootstrapParams = @{
    StartStack           = $StartStack
    VaultAddr            = $VaultAddr
    KeycloakAdminUsername = $KeycloakAdminUsername
    IdentityBackend      = $IdentityBackend
    SkipBitwardenInject  = $SkipBitwardenInject
    OnlyBreakGlass       = $OnlyBreakGlass
}
if (-not [string]::IsNullOrWhiteSpace($BreakGlassUsername)) {
    $bootstrapParams['BreakGlassUsername'] = $BreakGlassUsername
}
if ($null -ne $BreakGlassPassword) {
    $bootstrapParams['BreakGlassPassword'] = $BreakGlassPassword
}
if ($WriteEnvFile) { $bootstrapParams['WriteEnvFile'] = $true }
if ($IncludeSdnTelemetry) { $bootstrapParams['IncludeSdnTelemetry'] = $true }

# --- Run secrets bootstrap ---
Write-Host "Running secrets bootstrap (networks already created)..."
& (Join-Path $scriptsDir "secrets-bootstrap.ps1") @bootstrapParams
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

# --- Optionally save Vault token for later runs ---
if ($SaveVaultToken -and $StartStack) {
    Write-Host "Saving Vault token to keystore for start-from-vault.ps1..."
    & (Join-Path $scriptsDir "save-vault-token-to-keystore.ps1")
    if ($LASTEXITCODE -ne 0) {
        Write-Warning "Could not save Vault token; next run use start-from-vault.ps1 only if you set VAULT_TOKEN manually or re-run launch-greenfield.ps1."
    }
}

Write-Host "Greenfield launch complete. Break-glass credentials were shown above (change after first login). Keycloak: http://127.0.0.1:8180/keycloak/admin"
