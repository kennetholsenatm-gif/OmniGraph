<#
.SYNOPSIS
  Save the current Vault token to the OS keystore so start-from-vault.ps1 can read it without a static file.
.DESCRIPTION
  Reads VAULT_TOKEN or VAULT_DEV_ROOT_TOKEN_ID from the environment (e.g. after running secrets-bootstrap.ps1)
  and stores it in Windows Credential Manager (CredentialManager module) or PowerShell SecretStore.
  Run once after bootstrap so that start-from-vault.ps1 can start the stack on later runs without typing the token.
#>

$ErrorActionPreference = "Stop"
$token = $env:VAULT_TOKEN
if (-not $token) { $token = $env:VAULT_DEV_ROOT_TOKEN_ID }

if (-not $token) {
    Write-Host "Set VAULT_TOKEN or VAULT_DEV_ROOT_TOKEN_ID first (e.g. after .\secrets-bootstrap.ps1)."
    exit 1
}

$target = "devsecops-vault-token"
$saved = $false

# Windows: CredentialManager module
$cm = Get-Module -ListAvailable -Name CredentialManager
if ($cm) {
    try {
        Import-Module CredentialManager -ErrorAction Stop
        $cred = [PSCredential]::new($target, (ConvertTo-SecureString -String $token -AsPlainText -Force))
        New-StoredCredential -Target $target -Credential $cred -Persist LocalMachine -ErrorAction Stop | Out-Null
        Write-Host "Saved Vault token to Windows Credential Manager (target: $target)."
        $saved = $true
    } catch {
        if ($_.Exception.Message -notmatch "already exists") {
            Write-Host "CredentialManager: $_"
        } else {
            Write-Host "Credential Manager entry already exists; update manually or delete 'devsecops-vault-token' first."
            $saved = $true
        }
    }
}

# PowerShell SecretStore (cross-platform)
if (-not $saved) {
    try {
        $vaults = Get-SecretVault -ErrorAction SilentlyContinue
        if ($vaults -match "SecretStore") {
            $sec = ConvertTo-SecureString -String $token -AsPlainText -Force
            Set-Secret -Name "devsecops-vault-token" -Secret $sec -Vault "SecretStore" -ErrorAction Stop
            Write-Host "Saved Vault token to PowerShell SecretStore (name: devsecops-vault-token)."
            $saved = $true
        }
    } catch {
        Write-Host "SecretStore: $_"
    }
}

if (-not $saved) {
    Write-Host "Could not save to keystore. Install one of:"
    Write-Host "  Install-Module CredentialManager -Scope CurrentUser"
    Write-Host "  Install-Module Microsoft.PowerShell.SecretStore -Scope CurrentUser; Register-SecretVault -Name SecretStore -ModuleName Microsoft.PowerShell.SecretStore"
    Write-Host "Or on next run set: `$env:VAULT_TOKEN = '<token>'"
}
