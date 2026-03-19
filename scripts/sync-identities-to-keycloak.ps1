<#
.SYNOPSIS
  Sync identities from identities.yaml or identities.json to Keycloak: create users and assign realm roles by privilege_level.
.DESCRIPTION
  Reads privilege_levels.json for keycloak_role per privilege_level, creates realm roles if missing,
  creates/updates users, and assigns the corresponding realm role. User passwords from Vault (secret/devsecops or users/<uid>)
  or generated and stored in Vault.
  Prefers service-account token (KEYCLOAK_AUTOMATION_CLIENT_ID/SECRET). See docs/IAM_LDAP_AND_AUTOMATION.md.
.PARAMETER IdentityFile
  Path to identities list: identities.yaml or identities.json (default: identities.yaml then identities.json).
.PARAMETER KeycloakUrl
  Keycloak base URL (default http://127.0.0.1:8180).
.PARAMETER Realm
  Realm name (default master).
.PARAMETER AutomationClientId
  Service-account client ID (default env KEYCLOAK_AUTOMATION_CLIENT_ID or Vault).
.PARAMETER AutomationClientSecret
  Service-account client secret (default env KEYCLOAK_AUTOMATION_CLIENT_SECRET or Vault).
.PARAMETER KeycloakAdmin
  (Deprecated) Admin username; used only if automation client not set.
.PARAMETER KeycloakAdminPassword
  (Deprecated) Admin password; used only if automation client not set.
#>
param(
    [string]$IdentityFile = "",
    [string]$KeycloakUrl = "http://127.0.0.1:8180",
    [string]$Realm = "master",
    [string]$AutomationClientId = $env:KEYCLOAK_AUTOMATION_CLIENT_ID,
    [string]$AutomationClientSecret = $env:KEYCLOAK_AUTOMATION_CLIENT_SECRET,
    [string]$KeycloakAdmin = $env:KEYCLOAK_ADMIN,
    [string]$KeycloakAdminPassword = $env:KEYCLOAK_ADMIN_PASSWORD,
    [string]$VaultAddr = $env:VAULT_ADDR,
    [string]$VaultToken = $env:VAULT_TOKEN,
    [string]$PrivilegeLevelsFile = ""
)

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir

# Prefer automation client (token); fall back to admin password (deprecated)
if (-not $AutomationClientId -or -not $AutomationClientSecret) {
    if ($VaultAddr -and $VaultToken) {
        try {
            $r = Invoke-RestMethod -Uri "$VaultAddr/v1/secret/data/devsecops" -Headers @{ "X-Vault-Token" = $VaultToken } -ErrorAction Stop
            $d = $r.data.data
            if (-not $AutomationClientId) { $AutomationClientId = $d.KEYCLOAK_AUTOMATION_CLIENT_ID }
            if (-not $AutomationClientSecret) { $AutomationClientSecret = $d.KEYCLOAK_AUTOMATION_CLIENT_SECRET }
        } catch {}
    }
}
if (-not $KeycloakAdmin -or -not $KeycloakAdminPassword) {
    if ($VaultAddr -and $VaultToken) {
        try {
            $r = Invoke-RestMethod -Uri "$VaultAddr/v1/secret/data/devsecops" -Headers @{ "X-Vault-Token" = $VaultToken } -ErrorAction Stop
            $d = $r.data.data
            if (-not $KeycloakAdmin) { $KeycloakAdmin = $d.KEYCLOAK_ADMIN }
            if (-not $KeycloakAdminPassword) { $KeycloakAdminPassword = $d.KEYCLOAK_ADMIN_PASSWORD }
        } catch {}
    }
    if (-not $KeycloakAdmin) { $KeycloakAdmin = "admin" }
}
$useAutomationClient = ($AutomationClientId -and $AutomationClientSecret)
if (-not $useAutomationClient -and (-not $KeycloakAdminPassword)) {
    Write-Error "Set KEYCLOAK_AUTOMATION_CLIENT_ID and KEYCLOAK_AUTOMATION_CLIENT_SECRET (env or Vault), or (deprecated) KEYCLOAK_ADMIN_PASSWORD. See docs/IAM_LDAP_AND_AUTOMATION.md."
}

# Resolve identity file
if (-not $IdentityFile) {
    foreach ($f in @("identities.yaml", "identities.json")) {
        $p = Join-Path $pipelineRoot $f
        if (Test-Path $p) { $IdentityFile = $p; break }
    }
}
if (-not $IdentityFile -or -not (Test-Path $IdentityFile)) {
    Write-Error "Identity file not found. Copy identities.example.yaml to identities.yaml or set -IdentityFile."
    exit 1
}

# Parse identity list (JSON or simple YAML)
$identities = @()
$ext = [System.IO.Path]::GetExtension($IdentityFile).ToLower()
if ($ext -eq ".json") {
    $obj = Get-Content $IdentityFile -Raw | ConvertFrom-Json
    $identities = @($obj.identities)
} else {
    $text = Get-Content $IdentityFile -Raw
    $blocks = $text -split '\r?\n\s*-\s+uid:'
    foreach ($block in $blocks) {
        if ($block -notmatch '\S') { continue }
        if ($block -match 'uid:\s*([^\s\r\n"]+)') { $uid = $matches[1].Trim() }
        elseif ($block -match '^\s*(\S+)') { $uid = $matches[1].Trim() }
        else { continue }
        $h = @{ uid = $uid }
        foreach ($line in ($block -split '\r?\n')) {
            if ($line -match '^\s+(privilege_level|ou|cn|mail):\s*(.+)') {
                $h[$matches[1]] = $matches[2].Trim().Trim('"')
            }
        }
        if ($uid -and $uid -notmatch '^#') { $identities += [PSCustomObject]$h }
    }
}

if ($identities.Count -eq 0) {
    Write-Warning "No identities found in $IdentityFile"
    exit 0
}

# Privilege level -> keycloak_role
$levelsPath = $PrivilegeLevelsFile
if (-not $levelsPath) { $levelsPath = Join-Path $pipelineRoot "privilege_levels.json" }
if (-not (Test-Path $levelsPath)) { Write-Error "privilege_levels.json not found at $levelsPath"; exit 1 }
$levelMap = Get-Content $levelsPath -Raw | ConvertFrom-Json

# Keycloak token: service account (preferred) or admin password (deprecated)
if ($useAutomationClient) {
    $tokenBody = "grant_type=client_credentials&client_id=$([uri]::EscapeDataString($AutomationClientId))&client_secret=$([uri]::EscapeDataString($AutomationClientSecret))"
} else {
    Write-Warning "Using KEYCLOAK_ADMIN/KEYCLOAK_ADMIN_PASSWORD is deprecated. Prefer KEYCLOAK_AUTOMATION_CLIENT_ID and KEYCLOAK_AUTOMATION_CLIENT_SECRET. See docs/IAM_LDAP_AND_AUTOMATION.md."
    $tokenBody = "grant_type=password&client_id=admin-cli&username=$([uri]::EscapeDataString($KeycloakAdmin))&password=$([uri]::EscapeDataString($KeycloakAdminPassword))"
}
$tokenResp = Invoke-RestMethod -Uri "$KeycloakUrl/realms/$Realm/protocol/openid-connect/token" -Method Post -Body $tokenBody -ContentType "application/x-www-form-urlencoded"
$accessToken = $tokenResp.access_token
$headers = @{ "Authorization" = "Bearer $accessToken"; "Content-Type" = "application/json" }

# Ensure realm roles exist
$rolesUri = "$KeycloakUrl/admin/realms/$Realm/roles"
$existingRoles = Invoke-RestMethod -Uri $rolesUri -Headers $headers -Method Get
$roleNames = $levelMap.PSObject.Properties.Value.keycloak_role
foreach ($r in $roleNames) {
    if ($existingRoles.name -notcontains $r) {
        Invoke-RestMethod -Uri $rolesUri -Headers $headers -Method Post -Body (@{ name = $r } | ConvertTo-Json) | Out-Null
        Write-Host "Created realm role: $r"
    }
}

# Realm id for role-mappings containerId
$realmRep = Invoke-RestMethod -Uri "$KeycloakUrl/admin/realms/$Realm" -Headers $headers -Method Get
$realmId = $realmRep.id
# Role id by name
$roleIds = @{}
$allRoles = Invoke-RestMethod -Uri $rolesUri -Headers $headers -Method Get
foreach ($ro in $allRoles) { $roleIds[$ro.name] = $ro.id }

# Create users and assign roles
$usersUri = "$KeycloakUrl/admin/realms/$Realm/users"
foreach ($u in $identities) {
    $uid = $u.uid
    $level = $u.privilege_level
    if (-not $level) { $level = "viewer" }
    $kcRole = $levelMap.$level.keycloak_role
    if (-not $kcRole) { $kcRole = "view-only" }
    $cn = $u.cn; if (-not $cn) { $cn = $uid }
    $parts = $cn -split '\s+', 2
    $firstName = $parts[0]; $lastName = if ($parts.Count -gt 1) { $parts[1] } else { "" }
    $mail = $u.mail; if (-not $mail) { $mail = "$uid@local" }

    # Check if user exists
    $search = Invoke-RestMethod -Uri "$usersUri?username=$([uri]::EscapeDataString($uid))" -Headers $headers -Method Get
    $userId = $null
    if ($search.Count -gt 0) { $userId = $search[0].id }

    $password = $null
    if ($VaultAddr -and $VaultToken) {
        try {
            $vr = Invoke-RestMethod -Uri "$VaultAddr/v1/secret/data/devsecops" -Headers @{ "X-Vault-Token" = $VaultToken } -ErrorAction Stop
            $password = $vr.data.data.KEYCLOAK_ADMIN_PASSWORD
            try {
                $ur = Invoke-RestMethod -Uri "$VaultAddr/v1/secret/data/users/$uid" -Headers @{ "X-Vault-Token" = $VaultToken } -ErrorAction Stop
                $password = $ur.data.data.password
            } catch {}
        } catch {}
    }
    if (-not $password -and $uid -eq $KeycloakAdmin) {
        $password = $KeycloakAdminPassword
    }
    if (-not $password) {
        $bytes = New-Object byte[] 24
        [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
        $password = [Convert]::ToBase64String($bytes) -replace '[+/=]',''
        $password = $password.Substring(0, [Math]::Min(24, $password.Length))
        if ($VaultAddr -and $VaultToken) {
            try {
                $body = @{ data = @{ password = $password } } | ConvertTo-Json
                Invoke-RestMethod -Uri "$VaultAddr/v1/secret/data/users/$uid" -Method Post -Headers @{ "X-Vault-Token" = $VaultToken; "Content-Type" = "application/json" } -Body $body -ErrorAction Stop | Out-Null
                Write-Host "Stored password for $uid in Vault secret/users/$uid"
            } catch {
                Write-Warning "Could not write password to Vault for $uid : $_"
            }
        }
    }

    $userBody = @{
        username  = $uid
        email     = $mail
        firstName = $firstName
        lastName  = $lastName
        enabled   = $true
    } | ConvertTo-Json

    if (-not $userId) {
        try {
            Invoke-RestMethod -Uri $usersUri -Headers $headers -Method Post -Body $userBody | Out-Null
            $search = Invoke-RestMethod -Uri "$usersUri?username=$([uri]::EscapeDataString($uid))" -Headers $headers -Method Get
            $userId = $search[0].id
            Write-Host "Created user: $uid"
            $credBody = @{ type = "password"; value = $password; temporary = $false } | ConvertTo-Json
            try {
                Invoke-RestMethod -Uri "$KeycloakUrl/admin/realms/$Realm/users/$userId/reset-password" -Headers $headers -Method Put -Body $credBody | Out-Null
            } catch { Write-Warning "Set password for $uid : $_" }
        } catch {
            if ($_.Exception.Message -match "409|already exists") {
                $search = Invoke-RestMethod -Uri "$usersUri?username=$([uri]::EscapeDataString($uid))" -Headers $headers -Method Get
                $userId = $search[0].id
            } else { Write-Warning "Create user $uid : $_"; continue }
        }
    } else {
        Write-Host "User exists: $uid"
    }

    $roleId = $roleIds[$kcRole]
    if (-not $roleId) { continue }
    $rolePayload = @(@{ id = $roleId; name = $kcRole; composite = $false; clientRole = $false; containerId = $realmId })
    try {
        Invoke-RestMethod -Uri "$KeycloakUrl/admin/realms/$Realm/users/$userId/role-mappings/realm" -Headers $headers -Method Post -Body ($rolePayload | ConvertTo-Json) | Out-Null
        Write-Host "Assigned role $kcRole to $uid"
    } catch {
        Write-Warning "Assign role $kcRole to $uid : $_"
    }
}

Write-Host "Done. Users and realm roles synced from $IdentityFile"
