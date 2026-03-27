<#
.SYNOPSIS
  Set or reset Gitea admin (or any user) password via the running container. No web UI needed.
.DESCRIPTION
  Runs gitea admin user change-password inside the devsecops-gitea container. Use when you don't have the web password.
  If the user does not exist, create them first with -CreateUser (then log in at http://localhost:3000).
.PARAMETER Username
  Gitea username (default kbolsen).
.PARAMETER Password
  New password (will prompt if not provided).
.PARAMETER ContainerName
  Gitea container name (default devsecops-gitea).
.PARAMETER CreateUser
  If set, create the user instead of changing password (use when no account exists yet).
.PARAMETER Email
  Email for -CreateUser (default kbolsen@localhost).
#>
param(
    [string]$Username = $env:GITEA_USER,
    [string]$Password = $env:GITEA_ADMIN_PASSWORD,
    [string]$ContainerName = "devsecops-gitea",
    [switch]$CreateUser = $false,
    [string]$Email = "kbolsen@localhost"
)
if (-not $Username) { $Username = "kbolsen" }
if (-not $Password) {
    $sec = Read-Host "Enter new password for $Username" -AsSecureString
    $Password = [Runtime.InteropServices.Marshal]::PtrToStringAuto([Runtime.InteropServices.Marshal]::SecureStringToBSTR($sec))
}
if ([string]::IsNullOrWhiteSpace($Password)) { Write-Error "Password required"; exit 1 }

$configPath = "/data/gitea/conf/app.ini"
if ($CreateUser) {
    Write-Host "Creating user $Username in Gitea (container $ContainerName)..."
    docker exec -u git $ContainerName gitea admin user create --username $Username --password $Password --email $Email --admin --config $configPath
} else {
    Write-Host "Setting password for $Username in Gitea (container $ContainerName)..."
    docker exec -u git $ContainerName gitea admin user change-password --username $Username --password $Password --config $configPath
}
if ($LASTEXITCODE -eq 0) {
    Write-Host "Done. Log in at http://localhost:3000 as $Username with the new password."
} else {
    Write-Host "Failed. Is the container running? (docker ps | findstr gitea). If the user does not exist, run with -CreateUser."
    exit 1
}
