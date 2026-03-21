<#
.SYNOPSIS
  List Gitea wiki pages for a repo (validates API + owner/repo). Use after Publish-GiteaWikiPages.ps1.
.PARAMETER Owner
.PARAMETER Repo
.PARAMETER GiteaUrl
.PARAMETER Token
  Personal access token (required).
.EXAMPLE
  .\scripts\verify-gitea-wiki.ps1 -Owner kbolsen -Repo devsecops-pipeline -GiteaUrl http://localhost:3000 -Token YOUR_PAT
#>
param(
    [Parameter(Mandatory = $true)][string]$Owner,
    [Parameter(Mandatory = $true)][string]$Repo,
    [Parameter(Mandatory = $true)][string]$GiteaUrl,
    [string]$Token = ""
)

$ErrorActionPreference = "Stop"
if (-not $Token) {
    Write-Error "Pass -Token (personal access token)."
    exit 1
}

$base = $GiteaUrl.TrimEnd("/")
$url = "$base/api/v1/repos/$([uri]::EscapeDataString($Owner))/$([uri]::EscapeDataString($Repo))/wiki/pages"
$headers = @{ Authorization = "token $Token" }

try {
    $pages = @(Invoke-RestMethod -Uri $url -Method GET -Headers $headers)
} catch {
    Write-Error "GET wiki/pages failed: $($_.Exception.Message)"
    exit 1
}

Write-Host "OK: $($pages.Count) wiki page(s) at $Owner/$Repo"
foreach ($p in $pages) {
    $t = $p.title
    Write-Host "  - $t"
}
exit 0
