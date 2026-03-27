#!/usr/bin/env bash
# Create all 17 external Docker networks (100.64.x.0/24) per docs/NETWORK_DESIGN.md.
# Idempotent: ignores "already exists". Run inside each LXC (or host) before compose.
set -euo pipefail
create() {
  docker network create --driver bridge --subnet "$1" "$2" 2>/dev/null || true
}
create "100.64.1.0/24"  gitea_net
create "100.64.2.0/24"  n8n_net
create "100.64.3.0/24"  zammad_net
create "100.64.4.0/24"  bitwarden_net
create "100.64.5.0/24"  gateway_net
create "100.64.6.0/24"  portainer_net
create "100.64.7.0/24"  llm_net
create "100.64.8.0/24"  chatops_net
create "100.64.10.0/24" msg_backbone_net
create "100.64.20.0/24" iam_net
create "100.64.30.0/24" agent_mesh_net
create "100.64.40.0/24" discovery_net
create "100.64.50.0/24" sdn_lab_net
create "100.64.51.0/24" telemetry_net
create "100.64.52.0/24" docs_net
create "100.64.53.0/24" sonarqube_net
create "100.64.54.0/24" siem_net
echo "Docker networks ready (17): matches scripts/create-networks.ps1 / NETWORK_DESIGN.md"
