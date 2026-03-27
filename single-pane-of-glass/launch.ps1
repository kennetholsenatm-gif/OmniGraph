# Start Single Pane of Glass. Run from single-pane-of-glass folder.
# Ensure gateway_net and backend networks (gitea_net, n8n_net, zammad_net) exist and tooling stack is up.
$ErrorActionPreference = "Stop"
$net = "gateway_net"
Write-Host "Ensuring network $net exists..."
docker network create --driver bridge --subnet 100.64.5.0/24 $net 2>$null
if ($LASTEXITCODE -ne 0) { Write-Host "  (network may already exist)" }
Write-Host "Starting gateway (Traefik, dashboard, webhook-listener)..."
docker compose up -d
if ($LASTEXITCODE -eq 0) {
    Write-Host "Done. Open http://localhost for the dashboard. Gitea at http://localhost/gitea/"
}
