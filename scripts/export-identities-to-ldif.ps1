<#
.SYNOPSIS
  Export identities from identities.yaml or identities.json to LDIF for LDAP/FreeIPA.
.DESCRIPTION
  Reads privilege_levels.json for ldap_ou per privilege_level, builds organizationalUnit and inetOrgPerson
  entries, and writes a single LDIF file. No passwords in LDIF; bind credentials managed separately.
.PARAMETER IdentityFile
  Path to identities list (default: identities.yaml or identities.json in pipeline root).
.PARAMETER OutputPath
  Output LDIF path (default: identities.ldif in pipeline root).
.PARAMETER BaseDn
  LDAP base DN (default: dc=devsecops,dc=local).
.PARAMETER AppendGroupsLdifPath
  If set and the file exists, its contents are appended after user entries (e.g. ldap-groups.template.ldif).
#>
param(
    [string]$IdentityFile = "",
    [string]$OutputPath = "",
    [string]$BaseDn = "dc=devsecops,dc=local",
    [string]$AppendGroupsLdifPath = ""
)

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$pipelineRoot = Split-Path -Parent $scriptDir

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

if (-not $OutputPath) { $OutputPath = Join-Path $pipelineRoot "identities.ldif" }

# Parse identity list (same as sync-identities-to-keycloak)
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
            if ($line -match '^\s+(privilege_level|ou|cn|mail|account_kind|member_of):\s*(.+)') {
                $h[$matches[1]] = $matches[2].Trim().Trim('"')
            }
        }
        if ($uid -and $uid -notmatch '^#') { $identities += [PSCustomObject]$h }
    }
}

$levelsPath = Join-Path $pipelineRoot "privilege_levels.json"
if (-not (Test-Path $levelsPath)) { Write-Error "privilege_levels.json not found"; exit 1 }
$levelMap = Get-Content $levelsPath -Raw | ConvertFrom-Json

function Esc-LdapValue {
    param([string]$v)
    if (-not $v) { return "" }
    $v = $v.Trim()
    if ($v -match '[,+\;"\\<>]') {
        $v = $v -replace '\\','\5c' -replace ',','\2c' -replace '\+','\2b' -replace '"','\22' -replace ';','\3b' -replace '<','\3c' -replace '>','\3e'
    }
    $v
}

$ldif = @()
$ldif += "# LDIF generated from $IdentityFile (VARLOCK identities); base DN: $BaseDn"
$ldif += "# No userPassword here; set via LDAP bind or vault-fed tooling."
$ldif += ""

# Collect OUs from identities
$ous = @{}
foreach ($u in $identities) {
    $ou = $u.ou
    if (-not $ou -and $u.privilege_level) {
        $ou = $levelMap.($u.privilege_level).ldap_ou
    }
    if (-not $ou) { $ou = "users" }
    $ous[$ou] = $true
}

# Top-level entry (optional; some servers want it)
$ldif += "dn: $BaseDn"
$ldif += "objectClass: top"
$ldif += "objectClass: domain"
if ($BaseDn -match '^dc=([^,]+)') { $ldif += "dc: $($matches[1])" }
$ldif += ""

foreach ($ouName in ($ous.Keys | Sort-Object)) {
    $ldif += "dn: ou=$ouName,$BaseDn"
    $ldif += "objectClass: top"
    $ldif += "objectClass: organizationalUnit"
    $ldif += "ou: $ouName"
    $ldif += ""
}

foreach ($u in $identities) {
    $uid = $u.uid
    $ou = $u.ou
    if (-not $ou -and $u.privilege_level) {
        $ou = $levelMap.($u.privilege_level).ldap_ou
    }
    if (-not $ou) { $ou = "users" }
    $cn = Esc-LdapValue($u.cn); if (-not $cn) { $cn = $uid }
    $mail = Esc-LdapValue($u.mail); if (-not $mail) { $mail = "$uid@local" }
    $sn = $cn -replace '^([^\s]+)\s.*','$1'
    if ($sn -eq $cn) { $sn = $uid }

    $dn = "uid=$uid,ou=$ou,$BaseDn"
    $ldif += "dn: $dn"
    $ldif += "objectClass: top"
    $ldif += "objectClass: person"
    $ldif += "objectClass: organizationalPerson"
    $ldif += "objectClass: inetOrgPerson"
    $ldif += "uid: $uid"
    $ldif += "cn: $cn"
    $ldif += "sn: $sn"
    $ldif += "mail: $mail"
    $ldif += "ou: $ou"
    $descParts = @()
    if ($u.account_kind) { $descParts += "account_kind=$($u.account_kind)" }
    if ($u.member_of) { $descParts += "member_of=$($u.member_of)" }
    if ($descParts.Count -gt 0) {
        $ldif += "description: $(Esc-LdapValue(($descParts -join '; ')))"
    }
    $ldif += ""
}

if ($AppendGroupsLdifPath -and (Test-Path $AppendGroupsLdifPath)) {
    $ldif += ""
    $ldif += "# --- Appended from $AppendGroupsLdifPath ---"
    $ldif += (Get-Content -LiteralPath $AppendGroupsLdifPath -Encoding UTF8)
    $ldif += ""
}

$ldif | Set-Content -Path $OutputPath -Encoding UTF8
Write-Host "Wrote $OutputPath ($($identities.Count) users, $($ous.Count) OUs)."
Write-Host "Import with: ldapadd -x -D 'cn=admin,$BaseDn' -W -f $OutputPath (or use FreeIPA/389-ds import)."
