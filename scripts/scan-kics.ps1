$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$outDir = Join-Path $repoRoot "docs\artifacts\kics"
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

$targets = @(
    (Join-Path $repoRoot "ansible"),
    (Join-Path $repoRoot "deployments"),
    (Join-Path $repoRoot "docker-compose"),
    (Join-Path $repoRoot "opentofu")
)

$existing = $targets | Where-Object { Test-Path $_ }
$pathsArg = ($existing -join ",")

if (-not $pathsArg) {
    Write-Error "No scan targets found."
}

if (-not (Get-Command kics -ErrorAction SilentlyContinue)) {
    Write-Error "kics binary not found in PATH."
}

kics scan -p $pathsArg --report-formats json,sarif --output-path $outDir
Write-Host "KICS reports written to $outDir"
