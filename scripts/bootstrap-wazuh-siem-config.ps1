<#
.SYNOPSIS
  Populate docker-compose/siem/wazuh-config from upstream wazuh-docker (v4.9.2), generate indexer TLS certs, append Traefik basePath for /wazuh.
.DESCRIPTION
  Run once on a Linux Docker host (Wazuh requires vm.max_map_count etc.). Requires git and docker on PATH.
  Does not write secrets to disk beyond the cloned config tree under docker-compose/siem/wazuh-config.
.EXAMPLE
  .\scripts\bootstrap-wazuh-siem-config.ps1
#>
$ErrorActionPreference = "Stop"
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Split-Path -Parent $scriptDir
$dest = Join-Path $repoRoot "docker-compose\siem\wazuh-config"
$tmp = Join-Path $env:TEMP ("wazuh-docker-" + [guid]::NewGuid().ToString("n"))

Write-Host "Cloning wazuh/wazuh-docker (v4.9.2) to $tmp ..."
git clone --depth 1 --branch v4.9.2 "https://github.com/wazuh/wazuh-docker.git" $tmp
$single = Join-Path $tmp "single-node"
if (-not (Test-Path $single)) {
    Write-Error "Expected single-node directory missing in clone."
}

Push-Location $single
try {
    Write-Host "Generating indexer/manager/dashboard certificates (docker compose -f generate-indexer-certs.yml run --rm generator)..."
    docker compose -f generate-indexer-certs.yml run --rm generator
    if ($LASTEXITCODE -ne 0) { throw "Certificate generation failed (exit $LASTEXITCODE)." }
} finally {
    Pop-Location
}

New-Item -ItemType Directory -Force -Path $dest | Out-Null
Write-Host "Copying config to $dest ..."
Copy-Item -Path (Join-Path $single "config\*") -Destination $dest -Recurse -Force

$odb = Join-Path $dest "wazuh_dashboard\opensearch_dashboards.yml"
if (Test-Path $odb) {
    $append = @"

# Appended for Traefik PathPrefix /wazuh (devsecops-pipeline)
server.basePath: "/wazuh"
server.rewriteBasePath: true
"@
    Add-Content -LiteralPath $odb -Value $append -Encoding utf8
    Write-Host "Appended server.basePath /wazuh to opensearch_dashboards.yml"
} else {
    Write-Warning "opensearch_dashboards.yml not found; add server.basePath manually (see docs/WAZUH_SIEM.md)."
}

Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
Write-Host "Done. Start SIEM: docker compose -f docker-compose/docker-compose.siem.yml up -d (with env from Vault)."
