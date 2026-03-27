$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$outDir = Join-Path $repoRoot "docs\artifacts\inframap"
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

if (-not (Get-Command inframap -ErrorAction SilentlyContinue)) {
    Write-Error "inframap binary not found in PATH."
}

$tfPlan = Join-Path $repoRoot "opentofu\terraform.tfstate"
$tfDir = Join-Path $repoRoot "opentofu"

if (Test-Path $tfPlan) {
    inframap generate $tfPlan | Out-File -Encoding utf8 (Join-Path $outDir "inframap.dot")
} elseif (Test-Path $tfDir) {
    Set-Location $tfDir
    inframap generate . | Out-File -Encoding utf8 (Join-Path $outDir "inframap.dot")
    Set-Location $repoRoot
} else {
    Write-Error "No opentofu/terraform path found for inframap input."
}

Write-Host "Inframap output: $outDir\inframap.dot"
