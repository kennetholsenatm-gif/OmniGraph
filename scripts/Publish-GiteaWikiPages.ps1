<#
.SYNOPSIS
  Publish markdown files from wiki/gitea-pages to a Gitea repository wiki via REST API.
.DESCRIPTION
  Uses POST /api/v1/repos/{owner}/{repo}/wiki/new (create) and PATCH .../wiki/page/{pageName} (update).
  Body: { "title", "content_base64" (UTF-8), "message" } per Gitea 1.22+ OpenAPI.
  See docs/GITEA_WIKI.md.
.PARAMETER Owner
  Repository owner (e.g. kbolsen).
.PARAMETER Repo
  Repository name (e.g. devsecops-pipeline).
.PARAMETER GiteaUrl
  Gitea base URL without trailing slash (e.g. http://localhost:3000).
.PARAMETER Token
  Personal access token (required unless -DryRun).
.PARAMETER PagesDir
  Directory containing *.md files; each file basename (without .md) becomes the wiki page title.
.PARAMETER DryRun
  If set, only lists actions; no HTTP writes.
.PARAMETER Message
  Git commit message for wiki repo.
.EXAMPLE
  .\scripts\Publish-GiteaWikiPages.ps1 -Owner kbolsen -Repo devsecops-pipeline -GiteaUrl http://localhost:3000 -Token YOUR_PAT
.EXAMPLE
  .\scripts\Publish-GiteaWikiPages.ps1 -Owner kbolsen -Repo devsecops-pipeline -GiteaUrl http://localhost:3000 -DryRun
#>
param(
    [Parameter(Mandatory = $true)][string]$Owner,
    [Parameter(Mandatory = $true)][string]$Repo,
    [Parameter(Mandatory = $true)][string]$GiteaUrl,
    [string]$Token = "",
    [string]$PagesDir = "",
    [switch]$DryRun,
    [string]$Message = "Publish wiki pages from devsecops-pipeline repo"
)

$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptDir

if (-not $Token -and -not $DryRun) {
    Write-Error "Pass -Token (personal access token) or use -DryRun."
    exit 1
}

$headers = if ($DryRun) {
    @{}
} else {
    @{ Authorization = "token $Token" }
}

if (-not $PagesDir) {
    $PagesDir = Join-Path $repoRoot "wiki\gitea-pages"
}
if (-not (Test-Path -LiteralPath $PagesDir)) {
    Write-Error "Pages directory not found: $PagesDir"
    exit 1
}

$base = $GiteaUrl.TrimEnd("/")
$apiBase = "$base/api/v1/repos/$([uri]::EscapeDataString($Owner))/$([uri]::EscapeDataString($Repo))"

function Get-ExistingWikiPages {
    $listUrl = "$apiBase/wiki/pages"
    try {
        return @(Invoke-RestMethod -Uri $listUrl -Method GET -Headers $headers)
    } catch {
        $code = $null
        if ($_.Exception.Response) {
            try { $code = [int]$_.Exception.Response.StatusCode } catch {}
        }
        if ($code -eq 404) {
            Write-Warning "Wiki may be empty or repo/wiki not found (404 on list). New pages will use POST only."
        } elseif ($code) {
            Write-Warning "Wiki list failed (HTTP $code): $($_.Exception.Message)"
        }
        return @()
    }
}

function Get-PageSlugFromList {
    param($Pages, [string]$Title)
    $match = $Pages | Where-Object { $_.title -eq $Title } | Select-Object -First 1
    if (-not $match) { return $null }
    foreach ($prop in @("slug", "name")) {
        if ($match.PSObject.Properties.Name -contains $prop -and $match.$prop) {
            return [string]$match.$prop
        }
    }
    return [string]$match.title
}

function Publish-OneWikiPage {
    param(
        [string]$Title,
        [string]$Markdown,
        [array]$ExistingPages
    )
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($Markdown)
    $b64 = [Convert]::ToBase64String($bytes)
    $bodyObj = @{
        title           = $Title
        content_base64  = $b64
        message         = $Message
    }
    $json = $bodyObj | ConvertTo-Json -Compress
    $slug = Get-PageSlugFromList -Pages $ExistingPages -Title $Title

    if ($DryRun) {
        Write-Host "[DryRun] Would sync page: $Title ($($Markdown.Length) chars, slug on server: $(if ($slug) { $slug } else { '(new)' }))"
        return
    }

    $newUrl = "$apiBase/wiki/new"
    try {
        Invoke-RestMethod -Uri $newUrl -Method POST -Headers $headers -ContentType "application/json; charset=utf-8" -Body $json | Out-Null
        Write-Host "Created wiki page: $Title"
        return
    } catch {
        $status = $null
        if ($_.Exception.Response) { $status = [int]$_.Exception.Response.StatusCode }
        if ($status -ne 400 -and $status -ne 409) {
            Write-Warning "POST $Title failed ($status): $($_.Exception.Message)"
        }
    }

    if (-not $slug) {
        $ExistingPages = Get-ExistingWikiPages
        $slug = Get-PageSlugFromList -Pages $ExistingPages -Title $Title
    }
    if (-not $slug) {
        Write-Error "Could not create or resolve slug for page: $Title"
        exit 1
    }

    $enc = [uri]::EscapeDataString($slug)
    $patchUrl = "$apiBase/wiki/page/$enc"
    try {
        Invoke-RestMethod -Uri $patchUrl -Method PATCH -Headers $headers -ContentType "application/json; charset=utf-8" -Body $json | Out-Null
        Write-Host "Updated wiki page: $Title (slug: $slug)"
    } catch {
        Write-Error "PATCH failed for $Title ($patchUrl): $($_.Exception.Message)"
        exit 1
    }
}

$mdFiles = Get-ChildItem -LiteralPath $PagesDir -Filter "*.md" -File | Sort-Object Name
if ($mdFiles.Count -eq 0) {
    Write-Warning "No .md files in $PagesDir"
    exit 0
}

$existing = @()
if (-not $DryRun) {
    $existing = @(Get-ExistingWikiPages)
}

foreach ($f in $mdFiles) {
    $title = [System.IO.Path]::GetFileNameWithoutExtension($f.Name)
    $content = Get-Content -LiteralPath $f.FullName -Raw -Encoding UTF8
    Publish-OneWikiPage -Title $title -Markdown $content -ExistingPages $existing
}

Write-Host "Done. $($mdFiles.Count) file(s) processed from $PagesDir"
