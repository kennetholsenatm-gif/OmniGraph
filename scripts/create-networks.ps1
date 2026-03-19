# Phase 1: Create external Docker networks for Single Pane of Glass (Traefik) and backends.
# Subnets match docs/NETWORK_DESIGN.md and opentofu/variables.tf. Idempotent: skips create if network exists.
# Run from repo root or any directory; requires Docker.

$ErrorActionPreference = "Stop"
# Subnets align with docs/NETWORK_DESIGN.md, opentofu/variables.tf, and ansible/playbooks/deploy-devsecops-mesh.yml
$nets = @(
    @{ name = "gitea_net";        subnet = "100.64.1.0/24" },
    @{ name = "n8n_net";          subnet = "100.64.2.0/24" },
    @{ name = "zammad_net";       subnet = "100.64.3.0/24" },
    @{ name = "bitwarden_net";    subnet = "100.64.4.0/24" },
    @{ name = "gateway_net";      subnet = "100.64.5.0/24" },
    @{ name = "portainer_net";    subnet = "100.64.6.0/24" },
    @{ name = "llm_net";          subnet = "100.64.7.0/24" },
    @{ name = "chatops_net";      subnet = "100.64.8.0/24" },
    @{ name = "msg_backbone_net"; subnet = "100.64.10.0/24" },
    @{ name = "iam_net";          subnet = "100.64.20.0/24" },
    @{ name = "freeipa_net";      subnet = "100.64.21.0/24" },
    @{ name = "agent_mesh_net";   subnet = "100.64.30.0/24" },
    @{ name = "discovery_net";    subnet = "100.64.40.0/24" },
    @{ name = "sdn_lab_net";      subnet = "100.64.50.0/24" },
    @{ name = "telemetry_net";    subnet = "100.64.51.0/24" },
    @{ name = "docs_net";         subnet = "100.64.52.0/24" },
    @{ name = "sonarqube_net";    subnet = "100.64.53.0/24" },
    @{ name = "siem_net";         subnet = "100.64.54.0/24" }
)
foreach ($n in $nets) {
    $listOut = docker network ls -q --filter "name=^$($n.name)$" 2>&1
    if ($listOut -and "$listOut".Trim().Length -gt 0) {
        Write-Verbose "Network $($n.name) already exists; skipping."
        continue
    }
    $out = docker network create --driver bridge --subnet $n.subnet $n.name 2>&1
    if ($LASTEXITCODE -ne 0) {
        if ($out -match "already exists") {
            Write-Verbose "Network $($n.name) already exists (create reported it); continuing."
            continue
        }
        throw "Failed to create network $($n.name): $out"
    }
}
Write-Host "Docker networks ready (18): gitea_net, n8n_net, zammad_net, bitwarden_net, gateway_net, portainer_net, llm_net, chatops_net, msg_backbone_net, iam_net, freeipa_net, agent_mesh_net, discovery_net, sdn_lab_net, telemetry_net, docs_net, sonarqube_net, siem_net"
