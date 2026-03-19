# Register admin user kbolsen and SSH public key in Vault (Varlock).
# Prerequisites: SSH key in .dev/kbolsen_admin(.pub); Vault CLI and VAULT_ADDR/VAULT_TOKEN set.
# MFA is set to admin_mfa_email (default kenneth.olsen.atm@gmail.com) for OTP/recovery.
# Run from repo root (e.g. C:\GiTeaRepos\devsecops-pipeline).

$ErrorActionPreference = "Stop"
$adminMfaEmail = $env:ADMIN_MFA_EMAIL
if (-not $adminMfaEmail) { $adminMfaEmail = "kenneth.olsen.atm@gmail.com" }
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir
$devDir = Join-Path $pipelineRoot ".dev"
$keyPath = Join-Path $devDir "kbolsen_admin"
$keyPathPub = Join-Path $devDir "kbolsen_admin.pub"
$passwordFile = Join-Path $devDir "kbolsen_keycloak_password.txt"

if (-not (Test-Path $keyPathPub)) {
    Write-Error "Public key not found at $keyPathPub. Run: ssh-keygen -t ed25519 -C kbolsen -f $keyPath -N ''"
    exit 1
}

$pubKey = (Get-Content $keyPathPub -Raw).Trim()

# Use existing password or generate one (saved to .dev for one-time reference)
if (Test-Path $passwordFile) {
    $keycloakPassword = Get-Content $passwordFile -Raw
} else {
    $bytes = New-Object byte[] 24
    [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
    $keycloakPassword = [Convert]::ToBase64String($bytes) -replace '[+/=]', ''
    $keycloakPassword = $keycloakPassword.Substring(0, [Math]::Min(24, $keycloakPassword.Length))
    Set-Content -Path $passwordFile -Value $keycloakPassword -NoNewline
    Write-Host "Generated KEYCLOAK_ADMIN_PASSWORD; saved once to $passwordFile (gitignored)."
}

$vaultPath = $env:VAULT_SECRET_PATH
if (-not $vaultPath) { $vaultPath = "secret/devsecops" }

# KV v2: path is the mount path, not including /data/
$mount = ($vaultPath -split '/')[0]
$secretPath = ($vaultPath -split '/')[1..999] -join '/'

Write-Host "Username: kbolsen"
Write-Host "MFA email: $adminMfaEmail"
Write-Host "Vault path: $vaultPath"
Write-Host "Admin SSH public key (first 60 chars): $($pubKey.Substring(0, [Math]::Min(60, $pubKey.Length)))..."

$payload = @{
    KEYCLOAK_ADMIN           = "kbolsen"
    KEYCLOAK_ADMIN_PASSWORD  = $keycloakPassword
    admin_ssh_public_key     = $pubKey
    admin_mfa_email          = $adminMfaEmail
}

# Write payload to .dev/ for reference (gitignored)
$payloadPath = Join-Path $devDir "vault-payload-kbolsen.json"
$payload | ConvertTo-Json | Set-Content -Path $payloadPath -Encoding UTF8
Write-Host "Payload (with secret) written to $payloadPath for reference."

$vaultAddr = $env:VAULT_ADDR
if (-not $vaultAddr) { $vaultAddr = "http://localhost:8200" }
$vaultToken = $env:VAULT_TOKEN
if (-not $vaultToken) { $vaultToken = "devsecops-dev-root" }

function Write-VaultSecretViaApi {
    param([hashtable]$data, [string]$mountPath, [string]$secretPath)
    $headers = @{ "X-Vault-Token" = $vaultToken; "Content-Type" = "application/json" }
    # KV v2: POST /v1/secret/data/devsecops with body {"data": {...}}
    $uriV2 = "$vaultAddr/v1/$($mountPath.Trim('/'))/data/$($secretPath.Trim('/'))"
    $bodyV2 = @{ data = $data } | ConvertTo-Json -Depth 10
    try {
        Invoke-RestMethod -Uri $uriV2 -Method Post -Headers $headers -Body $bodyV2 | Out-Null
        return $true
    } catch {
        # KV v1: POST /v1/secret/devsecops with body {...}
        $uriV1 = "$vaultAddr/v1/$($mountPath.Trim('/'))/$($secretPath.Trim('/'))"
        try {
            $bodyV1 = $data | ConvertTo-Json -Depth 10
            Invoke-RestMethod -Uri $uriV1 -Method Post -Headers $headers -Body $bodyV1 | Out-Null
            return $true
        } catch {
            Write-Host "Vault API write failed: $_"
            return $false
        }
    }
}

$written = $false
if (Get-Command vault -ErrorAction SilentlyContinue) {
    $keyEscaped = $pubKey -replace '"', '\"'
    & vault kv put $vaultPath KEYCLOAK_ADMIN=kbolsen "KEYCLOAK_ADMIN_PASSWORD=$keycloakPassword" "admin_ssh_public_key=$keyEscaped" "admin_mfa_email=$adminMfaEmail"
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Vault updated successfully (CLI). Keycloak login: kbolsen / (password in $passwordFile or in Vault)."
        $written = $true
    }
}
if (-not $written) {
    $mount = ($vaultPath -split '/')[0]
    $secretPath = ($vaultPath -split '/')[1..999] -join '/'
    if (-not $secretPath) { $secretPath = "devsecops" }
    $data = @{
        KEYCLOAK_ADMIN          = "kbolsen"
        KEYCLOAK_ADMIN_PASSWORD = $keycloakPassword
        admin_ssh_public_key    = $pubKey
        admin_mfa_email         = $adminMfaEmail
    }
    if (Write-VaultSecretViaApi -data $data -mountPath $mount -secretPath $secretPath) {
        Write-Host "Vault updated successfully (API). Keycloak login: kbolsen / (password in $passwordFile or in Vault)."
    } else {
        Write-Host "Vault CLI not found and API write failed. Set VAULT_ADDR and VAULT_TOKEN, then re-run. Values in $payloadPath and $passwordFile."
    }
}
