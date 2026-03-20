<#
.SYNOPSIS
  Verify ansible/roles/lxd_devsecops_stack/defaults/main.yml references every compose file from
  docker-compose/stack-manifest.json (optionalStacks + sdnTelemetry) so OpenNebula LXC parity does not drift.
.EXAMPLE
  .\scripts\verify-lxd-manifest-parity.ps1
#>
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$manifestPath = Join-Path $root "docker-compose\stack-manifest.json"
$lxdDefaultsPath = Join-Path $root "ansible\roles\lxd_devsecops_stack\defaults\main.yml"

if (-not (Test-Path $manifestPath)) {
    Write-Error "Missing $manifestPath"
    exit 1
}
if (-not (Test-Path $lxdDefaultsPath)) {
    Write-Error "Missing $lxdDefaultsPath"
    exit 1
}

$manifest = Get-Content -LiteralPath $manifestPath -Raw -Encoding utf8 | ConvertFrom-Json
$defaults = Get-Content -LiteralPath $lxdDefaultsPath -Raw -Encoding utf8

function Assert-FileReferenced {
    param([string]$Label, [string[]]$Files)
    foreach ($f in $Files) {
        $escaped = [regex]::Escape($f)
        if ($defaults -notmatch $escaped) {
            Write-Error "lxd_devsecops_stack defaults missing reference to $Label file '$f' (expected in compose_cli)."
            exit 1
        }
    }
}

# sdnTelemetry bundle
$sdnFiles = @($manifest.sdnTelemetry.files | ForEach-Object { [string]$_ })
Assert-FileReferenced -Label "sdnTelemetry" -Files $sdnFiles

foreach ($prop in $manifest.optionalStacks.PSObject.Properties) {
    $name = $prop.Name
    $stack = $prop.Value
    foreach ($f in $stack.composeFiles) {
        Assert-FileReferenced -Label "optionalStacks.$name" -Files @([string]$f)
    }
}

# Ensure enable-map keys exist in defaults for every optional stack key we use in Ansible
$expectedKeys = @('discovery', 'llm', 'ai_orchestration', 'identity', 'siem', 'sdn_telemetry', 'gitea_lite')
foreach ($k in $expectedKeys) {
    if ($defaults -notmatch "lxd_optional_stack_enable:") {
        Write-Error "lxd_optional_stack_enable: block missing in LXD defaults."
        exit 1
    }
    $pat = "(?m)^\s{2}${k}:"
    if ($defaults -notmatch $pat) {
        Write-Error "lxd_optional_stack_enable missing key '$k' in $lxdDefaultsPath"
        exit 1
    }
}

Write-Host "OK: LXD defaults reference optional/manifest compose files and lxd_optional_stack_enable keys are present."
