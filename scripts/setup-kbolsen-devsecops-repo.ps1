#Requires -Version 5.1
<#
.SYNOPSIS
  Calls the WSL bash helper to create kbolsen/devsecops-pipeline on local Gitea.

.EXAMPLE
  $env:GITEA_ADMIN_PASSWORD = 'your-admin-password'
  .\scripts\setup-kbolsen-devsecops-repo.ps1
#>
param(
  [string]$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path,
  [string]$AdminPassword = $env:GITEA_ADMIN_PASSWORD
)

if (-not $AdminPassword) {
  Write-Error "Set GITEA_ADMIN_PASSWORD or pass -AdminPassword."
  exit 1
}

$unixPath = $RepoRoot -replace '\\', '/'
if ($unixPath -match '^([A-Za-z]):') {
  $unixPath = '/mnt/' + $Matches[1].ToLower() + $unixPath.Substring(2)
}
$env:GITEA_ADMIN_PASSWORD = $AdminPassword
# WSL inherits this env var; bash script reads GITEA_ADMIN_PASSWORD.
wsl -e bash -lc "cd '$unixPath' && ./scripts/setup-kbolsen-devsecops-repo.sh"
