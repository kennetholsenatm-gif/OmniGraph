param(
    [string]$Inventory = "ansible/inventory/opennebula-hybrid.example.yml"
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$outDir = Join-Path $repoRoot "docs\artifacts\ansible-cmdb"
New-Item -ItemType Directory -Force -Path $outDir | Out-Null

$inventoryPath = Join-Path $repoRoot $Inventory
if (-not (Test-Path $inventoryPath)) {
    Write-Error "Inventory not found: $inventoryPath"
}

if (-not (Get-Command ansible-cmdb -ErrorAction SilentlyContinue)) {
    Write-Error "ansible-cmdb not found in PATH."
}

ansible-cmdb -i $inventoryPath --format html_fancy --output-dir $outDir | Out-Null
Write-Host "Ansible-CMDB report generated in $outDir"
