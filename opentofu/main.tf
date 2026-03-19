# Docker bridge networks for Zero Trust DevSecOps pipeline (100.64.0.0/10).
# No containers defined here; use docker-compose for service definitions.

resource "docker_network" "gitea_net" {
  name = "gitea_net"
  driver = "bridge"
  ipam_config {
    subnet = var.tooling_gitea_cidr
  }
}

resource "docker_network" "n8n_net" {
  name = "n8n_net"
  driver = "bridge"
  ipam_config {
    subnet = var.tooling_n8n_cidr
  }
}

resource "docker_network" "zammad_net" {
  name = "zammad_net"
  driver = "bridge"
  ipam_config {
    subnet = var.tooling_zammad_cidr
  }
}

resource "docker_network" "bitwarden_net" {
  name = "bitwarden_net"
  driver = "bridge"
  ipam_config {
    subnet = var.tooling_bitwarden_cidr
  }
}

resource "docker_network" "gateway_net" {
  name = "gateway_net"
  driver = "bridge"
  ipam_config {
    subnet = var.tooling_gateway_cidr
  }
}

resource "docker_network" "msg_backbone_net" {
  name = "msg_backbone_net"
  driver = "bridge"
  ipam_config {
    subnet = var.msg_backbone_cidr
  }
}

resource "docker_network" "iam_net" {
  name = "iam_net"
  driver = "bridge"
  ipam_config {
    subnet = var.iam_cidr
  }
}

resource "docker_network" "portainer_net" {
  name = "portainer_net"
  driver = "bridge"
  ipam_config {
    subnet = var.tooling_portainer_cidr
  }
}

resource "docker_network" "llm_net" {
  name   = "llm_net"
  driver = "bridge"
  ipam_config {
    subnet = var.llm_net_cidr
  }
}

resource "docker_network" "chatops_net" {
  name   = "chatops_net"
  driver = "bridge"
  ipam_config {
    subnet = var.chatops_net_cidr
  }
}

resource "docker_network" "freeipa_net" {
  name = "freeipa_net"
  driver = "bridge"
  ipam_config {
    subnet = var.freeipa_cidr
  }
}

# Full-stack discovery: NetBox, NetDISCO, Syft/Trivy, Dependency-Track.
# Solace queues for BOM ingestion are configured on the Solace broker (not Docker network).
resource "docker_network" "discovery_net" {
  name = "discovery_net"
  driver = "bridge"
  ipam_config {
    subnet = var.discovery_net_cidr
  }
}

resource "docker_network" "agent_mesh_net" {
  name = "agent_mesh_net"
  driver = "bridge"
  ipam_config {
    subnet = var.agent_mesh_cidr
  }
}

resource "docker_network" "sdn_lab_net" {
  name   = "sdn_lab_net"
  driver = "bridge"
  ipam_config {
    subnet = var.sdn_lab_net_cidr
  }
}

resource "docker_network" "telemetry_net" {
  name   = "telemetry_net"
  driver = "bridge"
  ipam_config {
    subnet = var.telemetry_net_cidr
  }
}

resource "docker_network" "docs_net" {
  name   = "docs_net"
  driver = "bridge"
  ipam_config {
    subnet = var.docs_net_cidr
  }
}

resource "docker_network" "sonarqube_net" {
  name   = "sonarqube_net"
  driver = "bridge"
  ipam_config {
    subnet = var.sonarqube_net_cidr
  }
}

resource "docker_network" "siem_net" {
  name   = "siem_net"
  driver = "bridge"
  ipam_config {
    subnet = var.siem_net_cidr
  }
}
