param(
    [switch]$SkipSemaphore
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $repoRoot

Write-Host "Lean local control-plane setup in $repoRoot"

if (-not $SkipSemaphore) {
    Write-Host "Semaphore: provision via Incus (Ansible), not Docker — run:"
    Write-Host "  $repoRoot\scripts\start-semaphore.ps1"
}

Write-Host "Installing/updating pre-commit hooks (if available)..."
try {
    pre-commit install | Out-Null
} catch {
    Write-Warning "pre-commit not found in PATH. Install Python + pre-commit."
}

Write-Host "Installing Ansible collections (if ansible-galaxy available)..."
try {
    Set-Location "$repoRoot\ansible"
    ansible-galaxy collection install -r collections/requirements.yml
} catch {
    Write-Warning "ansible-galaxy not found in PATH. Install Ansible in your controller environment."
} finally {
    Set-Location $repoRoot
}

Write-Host "Done. Open Semaphore at http://127.0.0.1:3001"
