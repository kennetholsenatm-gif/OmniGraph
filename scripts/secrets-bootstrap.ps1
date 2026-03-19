<#
.SYNOPSIS
  Greenfield secrets bootstrap: generate strong random secrets, inject into Vault, start stack with process env (default: no docker-compose/.env on disk).
.DESCRIPTION
  Generates cryptographically random values for all pipeline secrets, sets them in the current process
  environment, starts the Docker Compose stacks (so containers get them via env), then writes the same
  secrets to Vault at secret/devsecops for Varlock and other consumers. By default **no docker-compose/.env file** is written
  (zero-disk); use -WriteEnvFile for a local docker-compose/.env only if needed. Includes break-glass admin:
  generates a random password (OpenNebula-style), creates that user in Keycloak or FreeIPA, injects
  secrets into Bitwarden (Vaultwarden) via the Bitwarden CLI (bw), and displays credentials once at
  the end. No default password; change after first login. Vault stores ciphertext at rest; use -WriteEnvFile only if you accept a local docker-compose/.env.
.PARAMETER StartStack
  If set (default), start the merged core stack after generating secrets (compose files in
  docker-compose/stack-manifest.json: IAM, messaging, tooling, ChatOps; optional SDN+telemetry with -IncludeSdnTelemetry).
  If not set, only inject into Vault (assumes Vault is already running).
.PARAMETER VaultAddr
  Vault address (default http://127.0.0.1:8200).
.PARAMETER KeycloakAdminUsername
  Keycloak admin username (default admin).
.PARAMETER BreakGlassUsername
  Break-glass admin username. If not provided, defaults to "admin".
.PARAMETER BreakGlassPassword
  Break-glass admin password (SecureString). If not provided, a random password is generated and shown once at the end.
.PARAMETER IdentityBackend
  Where to create the break-glass user: Keycloak (default) or FreeIPA.
.PARAMETER SkipBitwardenInject
  Skip injecting secrets into Bitwarden via bw (e.g. when bw is not installed).
.PARAMETER OnlyBreakGlass
  When set with -StartStack:$false, run only break-glass steps (LDAP/Keycloak user + Bitwarden inject). Stack and Vaultwarden must already be up.
.PARAMETER WriteEnvFile
  If set, write docker-compose/.env with generated secrets (opt-in; default is process env + Vault only).
.PARAMETER IncludeSdnTelemetry
  If set with -StartStack, add docker-compose.network.yml and docker-compose.telemetry.yml to the same compose up (AlmaLinux / Linux host recommended).
.EXAMPLE
  .\secrets-bootstrap.ps1
  Generate secrets, set env, start full stack, push to Vault, create break-glass user, inject into Bitwarden.
.EXAMPLE
  .\secrets-bootstrap.ps1 -StartStack:$false
  Only push currently generated or existing env secrets to Vault (Vault must be up). Break-glass steps are skipped.
.EXAMPLE
  .\secrets-bootstrap.ps1 -StartStack:$false -OnlyBreakGlass
  Run only break-glass user creation and Bitwarden inject (stack and Vaultwarden must be up).
#>
param(
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
$composeDir = Join-Path $pipelineRoot "docker-compose"
$script:BreakGlassPasswordPlain = $null   # Set when we generate; displayed once at end then cleared

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

# --- Break-glass admin: default username "admin", generate random password (OpenNebula-style; display once at end) ---
if ([string]::IsNullOrWhiteSpace($BreakGlassUsername)) {
    $BreakGlassUsername = "admin"
}
if (-not $BreakGlassPassword) {
    $script:BreakGlassPasswordPlain = New-StrongPassword -Length 24
    $BreakGlassPassword = ConvertTo-SecureString -String $script:BreakGlassPasswordPlain -AsPlainText -Force
} else {
    $tempCred = [PSCredential]::new("u", $BreakGlassPassword)
    $tempPlain = $tempCred.GetNetworkCredential().Password
    if ([string]::IsNullOrEmpty($tempPlain) -or $tempPlain.Length -lt 8) {
        Write-Error "Break-glass admin password must be at least 8 characters."
        exit 1
    }
    $tempPlain = $null
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

# Create break-glass admin user in Keycloak (master realm) or FreeIPA. No secrets to disk.
function New-BreakGlassUser {
    param(
        [string]$Username,
        [SecureString]$SecurePassword,
        [ValidateSet("Keycloak", "FreeIPA")]
        [string]$IdentityBackend = "Keycloak",
        [string]$KeycloakBaseUrl = "",
        [string]$KeycloakAdminUser = "",
        [string]$KeycloakAdminPassword = ""
    )
    $plainPassword = $null
    try {
        $plainPassword = [PSCredential]::new("u", $SecurePassword).GetNetworkCredential().Password
        if ($IdentityBackend -eq "Keycloak") {
            $baseUrl = if ($KeycloakBaseUrl) { $KeycloakBaseUrl.TrimEnd("/") } elseif ($env:KEYCLOAK_PUBLIC_URL) { $env:KEYCLOAK_PUBLIC_URL -replace "/$", "" } else { "http://127.0.0.1:8180/keycloak" }
            $maxAttempts = 48
            $attempt = 0
            while ($attempt -lt $maxAttempts) {
                try {
                    $null = Invoke-RestMethod -Uri "$baseUrl/realms/master" -Method Get -TimeoutSec 5 -ErrorAction Stop
                    Write-Host "Keycloak is ready."
                    break
                } catch {
                    $attempt++
                    Write-Host "Waiting for Keycloak... attempt $attempt/$maxAttempts"
                    if ($attempt -ge $maxAttempts) {
                        Write-Warning "Keycloak not reachable at $baseUrl after $maxAttempts attempts. Skipping break-glass user creation."
                        return
                    }
                    Start-Sleep -Seconds 5
                }
            }
            $tokenBodyStr = "grant_type=password&client_id=admin-cli&username=" + [uri]::EscapeDataString($KeycloakAdminUser) + "&password=" + [uri]::EscapeDataString($KeycloakAdminPassword)
            $tokenResponse = Invoke-RestMethod -Uri "$baseUrl/realms/master/protocol/openid-connect/token" -Method Post -Body $tokenBodyStr -ContentType "application/x-www-form-urlencoded" -ErrorAction Stop
            $accessToken = $tokenResponse.access_token
            $userBody = @{
                username   = $Username
                email      = "$Username@breakglass.local"
                enabled    = $true
                credentials = @(@{ type = "password"; value = $plainPassword; temporary = $false })
            } | ConvertTo-Json -Depth 5
            $headers = @{ "Authorization" = "Bearer $accessToken"; "Content-Type" = "application/json" }
            try {
                Invoke-RestMethod -Uri "$baseUrl/admin/realms/master/users" -Method Post -Headers $headers -Body $userBody -ErrorAction Stop | Out-Null
                Write-Host "Break-glass user '$Username' created in Keycloak (master realm)."
            } catch {
                if ($_.Exception.Response.StatusCode.value__ -eq 409) {
                    Write-Host "Break-glass user '$Username' already exists in Keycloak; skipping."
                } else {
                    Write-Warning "Keycloak user creation failed: $_"
                }
            }
        } elseif ($IdentityBackend -eq "FreeIPA") {
            $containerName = "devsecops-freeipa"
            $running = docker ps --filter "name=$containerName" --format "{{.Names}}" 2>$null
            if (-not $running) {
                Write-Warning "Container $containerName is not running. Skipping break-glass user creation in FreeIPA."
                return
            }
            $plainPassword | docker exec -i $containerName ipa user-add $Username --first=BreakGlass --last=Admin --password 2>&1 | Out-Null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "Break-glass user '$Username' created in FreeIPA."
            } else {
                Write-Warning "FreeIPA user creation failed (user may already exist). Continuing."
            }
        }
    } finally {
        $plainPassword = $null
    }
}

# Inject pipeline secrets into Bitwarden (Vaultwarden) via Bitwarden CLI. Requires 'bw' on PATH. No secrets to disk.
function Write-BitwardenSecrets {
    param(
        [hashtable]$Secrets,
        [string]$BwServerUrl,
        [string]$Username,
        [SecureString]$SecurePassword
    )
    $serverUrl = if ($BwServerUrl) { $BwServerUrl.TrimEnd("/") } else { ($env:BITWARDEN_DOMAIN -replace "/$", "") }
    if ([string]::IsNullOrWhiteSpace($serverUrl)) { $serverUrl = "http://localhost:8484" }
    $env:BW_SERVER = $serverUrl
    $maxAttempts = 36
    $bwTimeoutSec = 10
    $attempt = 0
    while ($attempt -lt $maxAttempts) {
        $attempt++
        Write-Host "Waiting for Vaultwarden (bw)... attempt $attempt/$maxAttempts (timeout ${bwTimeoutSec}s)"
        $job = Start-Job { & bw status 2>&1 }
        $null = Wait-Job $job -Timeout $bwTimeoutSec
        if ($job.State -eq 'Completed') {
            $null = Receive-Job $job
            Remove-Job $job -Force -ErrorAction SilentlyContinue
            Write-Host "Bitwarden CLI reached server."
            break
        }
        Stop-Job $job -ErrorAction SilentlyContinue
        Remove-Job $job -Force -ErrorAction SilentlyContinue
        if ($attempt -ge $maxAttempts) {
            Write-Warning "Vaultwarden not reachable at $serverUrl after $maxAttempts attempts. Ensure BITWARDEN_DOMAIN is correct, container is up, and 'bw' CLI is installed. Skipping Bitwarden inject."
            return
        }
        Start-Sleep -Seconds 5
    }
    $plainPassword = $null
    try {
        $plainPassword = [PSCredential]::new("u", $SecurePassword).GetNetworkCredential().Password
        & bw login $Username $plainPassword 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            Write-Warning "Bitwarden login failed. Break-glass account may not exist. Sign up once at $serverUrl with this username and password, then re-run this script to inject secrets."
            return
        }
        $unlockOut = & bw unlock $plainPassword --raw 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Warning "Bitwarden unlock failed. Skipping secret inject."
            return
        }
        $env:BW_SESSION = $unlockOut.Trim()
        $created = 0
        foreach ($key in $Secrets.Keys) {
            $value = $Secrets[$key]
            if ($null -eq $value -or $value -is [hashtable]) { continue }
            try {
                $templateJson = & bw get template item 2>$null
                $template = $templateJson | ConvertFrom-Json
                $template.type = 1
                $template.name = $key
                $template.login.username = $key
                $template.login.password = $value
                $itemJson = $template | ConvertTo-Json -Depth 10 -Compress
                $itemJson | & bw encode | & bw create item 2>&1 | Out-Null
                if ($LASTEXITCODE -eq 0) { $created++ }
            } catch {
                Write-Warning "Bitwarden item '$key' failed: $_"
            }
        }
        if ($created -gt 0) {
            Write-Host "Secrets injected into Bitwarden ($created items) for break-glass user '$Username'."
        }
    } finally {
        $plainPassword = $null
        $env:BW_SESSION = $null
    }
}

# --- Generate or reuse from env (no disk) ---
# All pipeline secrets: generated here or from env, then written to Vault. Optional .env only with -WriteEnvFile.
# OIDC/automation: KEYCLOAK_AUTOMATION_CLIENT_SECRET, GITEA_OIDC_CLIENT_SECRET, N8N_OIDC_CLIENT_SECRET
# are populated after Keycloak clients exist (Ansible keycloak_iam or script); add to Vault then.
$secretKeys = @(
    "KEYCLOAK_DB_PASSWORD", "KEYCLOAK_ADMIN_PASSWORD", "VAULT_DEV_ROOT_TOKEN_ID",
    "ZAMMAD_POSTGRES_PASSWORD", "GITEA_API_TOKEN", "SOLACE_PASSWORD", "SOLACE_ADMIN_PASSWORD",
    "ZAMMAD_API_TOKEN", "WEBHOOK_HMAC_SECRET", "N8N_API_TOKEN", "TELEPORT_API_TOKEN",
    "POSTGRES_PASSWORD", "RABBITMQ_DEFAULT_PASS",
    "BITWARDEN_ADMIN_TOKEN", "GATEWAY_REFRESH_SECRET",
    # ChatOps (Zulip): zero-disk; injected via env and Vault
    "ZULIP_POSTGRES_PASSWORD", "ZULIP_MEMCACHED_PASSWORD", "ZULIP_RABBITMQ_PASSWORD", "ZULIP_REDIS_PASSWORD",
    "ZULIP_SECRET_KEY", "ZULIP_EMAIL_PASSWORD", "ZULIP_OIDC_CLIENT_SECRET",
    # Discovery stack (docker-compose.discovery.yml); inject via env / Vault (optional -WriteEnvFile only)
    "NETBOX_DB_PASSWORD", "NETBOX_REDIS_PASSWORD", "NETBOX_SECRET_KEY", "NETBOX_SUPERUSER_PASSWORD",
    "NETBOX_API_TOKEN", "DEP_TRACK_API_KEY", "TERMIUS_API_TOKEN",
    # SDN + telemetry (docker-compose.network.yml / docker-compose.telemetry.yml)
    "VYOS_USER_PASSWORD", "VYOS_ENROLL_KEY", "GRAFANA_ADMIN_PASSWORD", "GRAFANA_OIDC_CLIENT_SECRET", "SFLOW_RT_ADMIN_TOKEN",
    # SonarQube (docker-compose.tooling.yml + messaging sonar-db-init)
    "SONAR_JDBC_PASSWORD", "SONARQUBE_OIDC_CLIENT_SECRET"
)
$secrets = @{}
foreach ($k in $secretKeys) {
    $v = Get-Item -Path "Env:$k" -ErrorAction SilentlyContinue
    if ($v -and $v.Value) {
        $secrets[$k] = $v.Value
    } else {
        if ($k -match "PASSWORD") { $secrets[$k] = New-StrongPassword -Length 24 }
        elseif ($k -match "TOKEN|SECRET|HMAC") { $secrets[$k] = New-Token -Length 32 }
        else { $secrets[$k] = New-Token -Length 32 }
    }
}
$secrets["KEYCLOAK_ADMIN"] = if ($env:KEYCLOAK_ADMIN) { $env:KEYCLOAK_ADMIN } else { $KeycloakAdminUsername }
# Zulip: admin email for SETTING_ZULIP_ADMINISTRATOR (break-glass); not a random secret
$secrets["ZULIP_ADMINISTRATOR"] = if ($env:ZULIP_ADMINISTRATOR) { $env:ZULIP_ADMINISTRATOR } else { "admin@breakglass.local" }

Write-Host "Secrets prepared (from env or generated); setting in environment."

# Export to environment so compose can substitute
foreach ($k in $secrets.Keys) {
    Set-Item -Path "Env:$k" -Value $secrets[$k]
}

if ([string]::IsNullOrWhiteSpace($env:KEYCLOAK_PUBLIC_URL)) {
    Set-Item -Path "Env:KEYCLOAK_PUBLIC_URL" -Value "http://127.0.0.1:8180/keycloak"
}

if ($WriteEnvFile) {
    $envPath = Join-Path $composeDir ".env"
    $envLines = @()
    foreach ($k in $secrets.Keys) {
        $v = $secrets[$k]
        if ($null -eq $v) { $v = "" }
        if ($v -match '[\s#"\\]') {
            $esc = $v -replace '\\', '\\\\' -replace '"', '\"'
            $envLines += "${k}=`"${esc}`""
        } else {
            $envLines += "${k}=$v"
        }
    }
    if ([string]::IsNullOrWhiteSpace($env:KEYCLOAK_PUBLIC_URL)) {
        $envLines += "KEYCLOAK_PUBLIC_URL=http://127.0.0.1:8180/keycloak"
    }
    $envLines | Set-Content -Path $envPath -Encoding utf8
    Write-Host "WriteEnvFile: secrets also written to $envPath (optional; prefer same-shell env or Vault)."
} else {
    Write-Host "No .env file written (zero-disk default). Use -WriteEnvFile to emit docker-compose/.env, or export from Vault before compose."
}

Write-Host "Secrets set in environment (KEYCLOAK_ADMIN=$($secrets['KEYCLOAK_ADMIN']))."

if ($StartStack) {
    if (-not (Test-Path $composeDir)) {
        Write-Error "Compose directory not found: $composeDir"
        exit 1
    }
    Push-Location $composeDir
    try {
        # Ensure Keycloak public URL is set for break-glass and compose (proxy serves at /keycloak)
        if ([string]::IsNullOrWhiteSpace($env:KEYCLOAK_PUBLIC_URL)) {
            Set-Item -Path "Env:KEYCLOAK_PUBLIC_URL" -Value "http://127.0.0.1:8180/keycloak"
        }
        # Merged compose file list from stack-manifest.json (same order as docker-compose/launch-stack.ps1)
        . (Join-Path $composeDir "DevSecOpsStackManifest.ps1")
        $stackManifest = Get-DevSecOpsStackManifest -ComposeDirectory $composeDir
        $coreComposeFiles = Get-MergedCoreComposeFileList -Manifest $stackManifest -IncludeSdnTelemetry:$IncludeSdnTelemetry
        if ($IncludeSdnTelemetry) {
            Write-Host "IncludeSdnTelemetry: adding network + telemetry compose files (Linux SDN host recommended)."
        }
        Write-Host "Starting $($stackManifest.coreStack.label) (single compose run)..."
        Invoke-DevSecOpsDockerComposeUp -ComposeFiles $coreComposeFiles -EnvFileArg @() -RemoveOrphans
        $rc = $LASTEXITCODE
        if ($rc -ne 0) { exit $rc }
        Write-Host "Waiting 5 minutes for Docker stacks to come up..."
        Start-Sleep -Seconds 300
    } finally {
        Pop-Location
    }
}

# Wait for Vault and inject (use VAULT_TOKEN if set, e.g. for inject-only when Vault already running)
$token = if ($env:VAULT_TOKEN) { $env:VAULT_TOKEN } else { $secrets["VAULT_DEV_ROOT_TOKEN_ID"] }
$maxAttempts = 24
$attempt = 0
$vaultOk = $false
while ($attempt -lt $maxAttempts) {
    try {
        $r = Invoke-RestMethod -Uri "$VaultAddr/v1/sys/health" -Method Get -TimeoutSec 5 -ErrorAction Stop
        $vaultOk = $true
        Write-Host "Vault is ready."
        break
    } catch {
        $attempt++
        Write-Host "Waiting for Vault... attempt $attempt/$maxAttempts"
        if ($attempt -ge $maxAttempts) {
            Write-Warning "Vault not reachable at $VaultAddr after $maxAttempts attempts. Push secrets later with: `$env:VAULT_TOKEN='...'; .\secrets-bootstrap.ps1 -StartStack:`$false"
            exit 1
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
Write-Host "Keycloak service admin: $KeycloakAdminUsername (password in Vault at secret/devsecops or in this session env)."
Write-Host "Bitwarden ADMIN_TOKEN and GATEWAY_REFRESH_SECRET were generated and written to Vault; inject via start-from-vault.ps1 or Ansible."
# Optional: set master realm to allow HTTP (avoids "HTTPS required" when using proxy)
if ($StartStack) {
    Write-Host "Waiting 5 minutes for Keycloak to init, then setting master realm ssl_required=NONE..."
    Start-Sleep -Seconds 300
    docker exec devsecops-keycloak-db psql -U keycloak -d keycloak -c "UPDATE realm SET ssl_required = 'NONE' WHERE name = 'master';" 2>$null
    if ($LASTEXITCODE -eq 0) { Write-Host "Master realm set to allow HTTP." } else { Write-Host "If you see HTTPS required, run: docker exec devsecops-keycloak-db psql -U keycloak -d keycloak -c \"UPDATE realm SET ssl_required = 'NONE' WHERE name = 'master';\"" }
}

# Break-glass: create admin user in Keycloak/FreeIPA and inject secrets into Bitwarden via bw CLI
if ($StartStack -or $OnlyBreakGlass) {
    Write-Host "Creating break-glass user in $IdentityBackend..."
    New-BreakGlassUser -Username $BreakGlassUsername -SecurePassword $BreakGlassPassword -IdentityBackend $IdentityBackend -KeycloakBaseUrl $env:KEYCLOAK_PUBLIC_URL -KeycloakAdminUser $secrets["KEYCLOAK_ADMIN"] -KeycloakAdminPassword $secrets["KEYCLOAK_ADMIN_PASSWORD"]
    if (-not $SkipBitwardenInject) {
        Write-Host "Injecting secrets into Bitwarden (bw CLI)..."
        $bwUrl = if ($env:BITWARDEN_DOMAIN) { $env:BITWARDEN_DOMAIN -replace "/$", "" } else { "http://localhost:8484" }
        Write-BitwardenSecrets -Secrets $secrets -BwServerUrl $bwUrl -Username $BreakGlassUsername -SecurePassword $BreakGlassPassword
    }
}

# One-time display of break-glass credentials when we generated the password (OpenNebula-style)
if ($null -ne $script:BreakGlassPasswordPlain -and $script:BreakGlassPasswordPlain.Length -gt 0) {
    Write-Host ""
    Write-Host "========== BREAK-GLASS ADMIN (one-time; change after first login) ==========" -ForegroundColor Cyan
    Write-Host "  Username: $BreakGlassUsername"
    Write-Host "  Password: $script:BreakGlassPasswordPlain"
    Write-Host "  Keycloak: http://127.0.0.1:8180/keycloak/admin"
    Write-Host "  Save this password; it will not be shown again." -ForegroundColor Yellow
    Write-Host "============================================================================" -ForegroundColor Cyan
    $script:BreakGlassPasswordPlain = $null
}

Write-Host ""
Write-Host "To avoid storing anything on disk, run this script on each host startup and start the stack from the same shell, or export env from Vault before starting compose."
