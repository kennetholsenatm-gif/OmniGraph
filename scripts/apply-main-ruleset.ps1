# Apply branch ruleset for main via GitHub API. Requires: gh CLI, repo admin, repo scope.
$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Resolve-Path (Join-Path $ScriptDir "..")
$JsonFile = Join-Path $ScriptDir "github-ruleset-main.json"

if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    Write-Error "Install GitHub CLI: https://cli.github.com/"
}

if (-not (Test-Path $JsonFile)) {
    Write-Error "Missing $JsonFile"
}

Set-Location $RepoRoot
$repoSlug = gh repo view --json nameWithOwner -q .nameWithOwner 2>$null
if ([string]::IsNullOrWhiteSpace($repoSlug)) {
    Write-Error "Run from a clone with 'gh auth login', or set env GH_REPO=owner/name."
}

Write-Host "Creating ruleset on $repoSlug ..."
Get-Content -Raw $JsonFile | gh api "repos/$repoSlug/rulesets" --method POST --input -
Write-Host "Done. Verify under Settings → Rules → Rulesets."
