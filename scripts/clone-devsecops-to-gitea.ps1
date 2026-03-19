<#
.SYNOPSIS
  Commit and push devsecops-pipeline to Gitea (canonical repo at C:\GiTeaRepos\devsecops-pipeline).
.DESCRIPTION
  Run from the devsecops-pipeline repo (e.g. C:\GiTeaRepos\devsecops-pipeline). Inits git if needed, commits, creates repo on Gitea if GITEA_TOKEN set, adds origin if missing, and pushes. No copy from elsewhere; work in this repo.
.PARAMETER GiteaUrl
  Gitea base URL (default http://localhost:3000).
.PARAMETER GiteaUser
  Gitea username (default kbolsen or env GITEA_USER).
.PARAMETER RepoName
  Repository name on Gitea (default devsecops-pipeline).
.PARAMETER GiteaReposRoot
  Canonical host path for this repo (default C:\GiTeaRepos; used in messages only).
.PARAMETER NoPush
  Only commit; do not add remote or push.
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
$repoRoot = Split-Path -Parent $scriptDir

Push-Location $repoRoot

if (-not (Test-Path ".git")) {
    Write-Host "Initializing git in $repoRoot"
    git init
    git branch -M main
}

$status = git status --porcelain 2>$null
if ($status) {
    git add -A
    git status
    git commit -m "Sync: devsecops-pipeline"
} else {
    Write-Host "No changes to commit."
}

if ($NoPush) {
    Write-Host "NoPush: add remote and push manually: git remote add origin $GiteaUrl/$GiteaUser/$RepoName.git; git push -u origin main"
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
$existing = $null
$ErrorActionPreferenceSave = $ErrorActionPreference
$ErrorActionPreference = "SilentlyContinue"
try { $existing = (git remote get-url origin 2>$null) | Select-Object -First 1 }
finally { $ErrorActionPreference = $ErrorActionPreferenceSave }
if ($existing) {
    if ($existing -ne $remoteUrl) { git remote set-url origin $remoteUrl; Write-Host "Set origin to $remoteUrl" }
} else {
    git remote add origin $remoteUrl
    Write-Host "Added origin: $remoteUrl"
}

Write-Host "Pushing to origin main..."
git push -u origin main
if ($LASTEXITCODE -ne 0) {
    Write-Host "Push failed. Create repo in Gitea UI if needed, then: git push -u origin main"
    Pop-Location
    exit 1
}

Pop-Location
Write-Host "Done. Repo: $GiteaUrl/$GiteaUser/$RepoName (canonical: $GiteaReposRoot\$RepoName)"