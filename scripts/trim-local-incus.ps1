param(
    [string[]]$KeepInstances = @("devsecops-gitea"),
    [switch]$WhatIfOnly
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command incus -ErrorAction SilentlyContinue) -and -not (Get-Command lxc -ErrorAction SilentlyContinue)) {
    throw "Neither incus nor lxc found in PATH. Run inside your WSL/Alma10 Incus controller environment."
}

$cli = if (Get-Command incus -ErrorAction SilentlyContinue) { "incus" } else { "lxc" }
Write-Host "Using CLI: $cli"

$all = & $cli list -c n --format csv
$instances = @($all -split "`n" | Where-Object { $_ -and $_.Trim().Length -gt 0 })

foreach ($name in $instances) {
    if ($KeepInstances -contains $name) {
        Write-Host "KEEP $name"
        continue
    }
    Write-Host "STOP $name"
    if (-not $WhatIfOnly) {
        & $cli stop $name --force | Out-Null
    }
}

Write-Host "Trim complete. Kept instances: $($KeepInstances -join ', ')"
