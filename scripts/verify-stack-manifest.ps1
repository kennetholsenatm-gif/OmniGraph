<#
.SYNOPSIS
  Verify docker-compose/stack-manifest.json exists, files are on disk, and Ansible lists the same core compose files.
.EXAMPLE
  .\scripts\verify-stack-manifest.ps1
#>
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$composeDir = Join-Path $root "docker-compose"
$manifestPath = Join-Path $composeDir "stack-manifest.json"
$roleTasks = Join-Path $root "ansible\roles\devsecops_containers\tasks\main.yml"

if (-not (Test-Path $manifestPath)) {
    Write-Error "Missing $manifestPath"
    exit 1
}

$manifest = Get-Content -LiteralPath $manifestPath -Raw -Encoding utf8 | ConvertFrom-Json
$requiredCore = @()
foreach ($f in $manifest.coreStack.files) { $requiredCore += [string]$f }

foreach ($f in $requiredCore) {
    $p = Join-Path $composeDir $f
    if (-not (Test-Path $p)) {
        Write-Error "Manifest lists missing file: $f"
        exit 1
    }
}

foreach ($f in $manifest.sdnTelemetry.files) {
    $p = Join-Path $composeDir ([string]$f)
    if (-not (Test-Path $p)) {
        Write-Error "Manifest sdnTelemetry lists missing file: $f"
        exit 1
    }
}

if ($manifest.PSObject.Properties.Name -contains "devSingleLxc" -and $manifest.devSingleLxc.files) {
    foreach ($f in $manifest.devSingleLxc.files) {
        $p = Join-Path $composeDir ([string]$f)
        if (-not (Test-Path $p)) {
            Write-Error "Manifest devSingleLxc lists missing file: $f"
            exit 1
        }
    }
}

foreach ($prop in $manifest.optionalStacks.PSObject.Properties) {
    $stack = $prop.Value
    foreach ($f in $stack.composeFiles) {
        $p = Join-Path $composeDir ([string]$f)
        if (-not (Test-Path $p)) {
            Write-Error "optionalStacks.$($prop.Name) lists missing file: $f"
            exit 1
        }
    }
}

$ansibleText = Get-Content -LiteralPath $roleTasks -Raw -Encoding utf8
foreach ($f in $requiredCore) {
    if ($ansibleText -notmatch [regex]::Escape($f)) {
        Write-Error "Ansible role tasks do not reference core manifest file $f (expected in docker_compose_v2 files). Edit ansible/roles/devsecops_containers/tasks/main.yml."
        exit 1
    }
}

Write-Host "OK: stack-manifest.json paths exist and Ansible references all $($requiredCore.Count) core compose files."
