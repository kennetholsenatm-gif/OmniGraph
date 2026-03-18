<#
.SYNOPSIS
  Copy devsecops-pipeline folder to C:\GiTeaRepos and push to Gitea.
.DESCRIPTION
  Copies the current folder (excluding .git) to C:\GiTeaRepos\devsecops-pipeline, inits git, commits, creates repo on Gitea if GITEA_TOKEN set, and pushes.
.PARAMETER GiteaUrl
  Gitea base URL (default http://localhost:3000).
.PARAMETER GiteaUser
  Gitea username (default kbolsen or env GITEA_USER).
.PARAMETER RepoName
  Repository name on Gitea (default devsecops-pipeline).
.PARAMETER GiteaReposRoot
  Host path for the copy (default C:\GiTeaRepos or env GITEA_REPOS_ROOT).
.PARAMETER NoPush
  Only copy and commit; do not add remote or push.
#>
param(
    [string]$GiteaUrl = $env:GITEA_URL,
    [string]$GiteaUser = $env:GITEA_USER,
    [string]$RepoName = "devsecops-pipeline",
    [string]$GiteaReposRoot = $env:GITEA_REPOS_ROOT,
    [switch]$NoPush = $false
)
if (-not $GiteaUrl) { $GiteaUrl = "http://localhost:3000" }
if (-not $GiteaUser) { $GiteaUser = "kbolsen" }
if (-not $GiteaReposRoot) { $GiteaReposRoot = "C:\GiTeaRepos" }

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$sourceDir = Split-Path -Parent $scriptDir
$targetDir = Join-Path $GiteaReposRoot $RepoName

if (-not (Test-Path $GiteaReposRoot)) {
    New-Item -ItemType Directory -Path $GiteaReposRoot -Force | Out-Null
    Write-Host "Created $GiteaReposRoot"
}

Write-Host "Copying $sourceDir -> $targetDir (excluding .git)..."
if (Test-Path $targetDir) {
    Remove-Item -Recurse -Force $targetDir
}
New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
Get-ChildItem -Path $sourceDir -Force | Where-Object { $_.Name -ne ".git" } | ForEach-Object {
    Copy-Item -Path $_.FullName -Destination (Join-Path $targetDir $_.Name) -Recurse -Force
}
Write-Host "Copy done."

Push-Location $targetDir

git init
git branch -M main
git add -A
git commit -m "Initial commit: devsecops-pipeline"

if ($NoPush) {
    Write-Host "NoPush: add Gitea remote and push manually: git remote add origin $GiteaUrl/$GiteaUser/$RepoName.git; git push -u origin main"
    Pop-Location
    exit 0
}

$token = $env:GITEA_TOKEN
if ($token) {
    $createUri = "$GiteaUrl/api/v1/user/repos"
    $body = @{ name = $RepoName; description = "Autonomous Zero Trust DevSecOps Pipeline (Vault, Keycloak, n8n, Zammad, Gitea)"; private = $false } | ConvertTo-Json
    $headers = @{ "Authorization" = "token $token"; "Content-Type" = "application/json" }
    try {
        Invoke-RestMethod -Uri $createUri -Method Post -Headers $headers -Body $body -ErrorAction Stop | Out-Null
        Write-Host "Created repository $GiteaUser/$RepoName on Gitea."
    } catch {
        if ($_.Exception.Response.StatusCode -eq 409) { Write-Host "Repository $RepoName already exists on Gitea." }
        else { Write-Warning "Could not create repo via API: $_" }
    }
} else {
    Write-Host "GITEA_TOKEN not set. Create repo in Gitea: $GiteaUrl -> New Repository -> Name: $RepoName (no README), then run again or push manually."
}

$remoteUrl = "$GiteaUrl/$GiteaUser/$RepoName.git"
git remote add origin $remoteUrl
Write-Host "Pushing to $remoteUrl..."
git push -u origin main
if ($LASTEXITCODE -ne 0) {
    Write-Host "Push failed. Create the repo in Gitea UI if needed, then: git push -u origin main"
    Pop-Location
    exit 1
}

Pop-Location
Write-Host "Done. Repo: $GiteaUrl/$GiteaUser/$RepoName (copy at $targetDir)"
