<#
.SYNOPSIS
  Create or update Keycloak OIDC clients for Gitea, n8n, and Zammad (Phase 2 IAM mesh).
.DESCRIPTION
  Uses Keycloak Admin API to ensure three clients exist in the realm with correct redirect URIs and scopes.
  Prefers service-account token (client credentials); no admin password. See docs/IAM_LDAP_AND_AUTOMATION.md.
  Idempotent: gets existing clients by clientId and updates redirectUris if present; otherwise creates.
.PARAMETER KeycloakUrl
  Keycloak base URL (default http://127.0.0.1:8180 or via gateway http://localhost/keycloak).
.PARAMETER Realm
  Realm name (default master).
.PARAMETER GatewayBaseUrl
  Public base URL for redirect URIs (default http://localhost). No trailing slash.
.PARAMETER AutomationClientId
  Service-account client ID (default env KEYCLOAK_AUTOMATION_CLIENT_ID or Vault).
.PARAMETER AutomationClientSecret
  Service-account client secret (default env KEYCLOAK_AUTOMATION_CLIENT_SECRET or Vault).
#>
param(
    [string]$KeycloakUrl = $env:KEYCLOAK_PUBLIC_URL,
    [string]$Realm = $env:KEYCLOAK_REALM,
    [string]$GatewayBaseUrl = $env:GATEWAY_BASE_URL,
    [string]$AutomationClientId = $env:KEYCLOAK_AUTOMATION_CLIENT_ID,
    [string]$AutomationClientSecret = $env:KEYCLOAK_AUTOMATION_CLIENT_SECRET,
    [string]$VaultAddr = $env:VAULT_ADDR,
    [string]$VaultToken = $env:VAULT_TOKEN
)

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir

if (-not $KeycloakUrl) { $KeycloakUrl = "http://127.0.0.1:8180" }
if (-not $Realm) { $Realm = "master" }
if (-not $GatewayBaseUrl) { $GatewayBaseUrl = "http://localhost" }
$GatewayBaseUrl = $GatewayBaseUrl.TrimEnd('/')

# Prefer automation client (token); no password or password files
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
if (-not $AutomationClientId -or -not $AutomationClientSecret) {
    Write-Host "Keycloak automation (service-account) credentials are required. No admin password." -ForegroundColor Yellow
    Write-Host "  1. Create a confidential client in Keycloak with Service accounts enabled; assign realm-management roles to its service account." -ForegroundColor Gray
    Write-Host "  2. Store in Vault at secret/data/devsecops: KEYCLOAK_AUTOMATION_CLIENT_ID, KEYCLOAK_AUTOMATION_CLIENT_SECRET" -ForegroundColor Gray
    Write-Host "  3. Or set env: KEYCLOAK_AUTOMATION_CLIENT_ID, KEYCLOAK_AUTOMATION_CLIENT_SECRET" -ForegroundColor Gray
    Write-Host "  See docs/IAM_LDAP_AND_AUTOMATION.md" -ForegroundColor Gray
    Write-Error "Set KEYCLOAK_AUTOMATION_CLIENT_ID and KEYCLOAK_AUTOMATION_CLIENT_SECRET (env or Vault)."
}

# Token via client credentials (service account)
$tokenBody = "grant_type=client_credentials&client_id=$([uri]::EscapeDataString($AutomationClientId))&client_secret=$([uri]::EscapeDataString($AutomationClientSecret))"
$tokenResp = Invoke-RestMethod -Uri "$KeycloakUrl/realms/$Realm/protocol/openid-connect/token" -Method Post -Body $tokenBody -ContentType "application/x-www-form-urlencoded"
$headers = @{ "Authorization" = "Bearer $($tokenResp.access_token)"; "Content-Type" = "application/json" }

# Get existing clients
$clientsUri = "$KeycloakUrl/admin/realms/$Realm/clients"
$existing = Invoke-RestMethod -Uri $clientsUri -Headers $headers -Method Get

function Ensure-Client {
    param(
        [string]$ClientId,
        [string[]]$RedirectUris,
        [string[]]$WebOrigins = @("+"),
        [bool]$Confidential = $true
    )
    $c = $existing | Where-Object { $_.clientId -eq $ClientId }
    if ($c) {
        $full = Invoke-RestMethod -Uri "$clientsUri/$($c.id)" -Headers $headers -Method Get
        $full.redirectUris = $RedirectUris
        $full.webOrigins = $WebOrigins
        $payload = $full | ConvertTo-Json -Depth 10 -Compress
        Invoke-RestMethod -Uri "$clientsUri/$($c.id)" -Headers $headers -Method Put -Body $payload
        Write-Host "Updated client: $ClientId"
    } else {
        $payload = @{
            clientId                  = $ClientId
            name                      = $ClientId
            enabled                   = $true
            redirectUris              = $RedirectUris
            webOrigins                = $WebOrigins
            publicClient              = -not $Confidential
            standardFlowEnabled       = $true
            directAccessGrantsEnabled = $false
        } | ConvertTo-Json -Depth 5 -Compress
        Invoke-RestMethod -Uri $clientsUri -Headers $headers -Method Post -Body $payload
        Write-Host "Created client: $ClientId"
    }
}

# Gitea: callback path is /gitea/user/oauth2/<source_name>/callback; source name "Keycloak" is typical
Ensure-Client -ClientId "gitea" -RedirectUris @(
    "$GatewayBaseUrl/gitea/user/oauth2/Keycloak/callback"
) -Confidential $true

# n8n
Ensure-Client -ClientId "n8n" -RedirectUris @(
    "$GatewayBaseUrl/n8n/rest/sso/oidc/callback"
) -Confidential $true

# Zammad: post_logout and backchannel optional
Ensure-Client -ClientId "zammad" -RedirectUris @(
    "$GatewayBaseUrl/zammad/auth/openid_connect/callback"
) -Confidential $false

Write-Host "Done. OIDC clients gitea, n8n, zammad are configured. Copy client secrets from Keycloak Admin UI (Clients -> each client -> Credentials) to Vault keys: GITEA_OIDC_CLIENT_SECRET, N8N_OIDC_CLIENT_SECRET. Zammad uses a public client (no secret) unless you enable client authentication."
