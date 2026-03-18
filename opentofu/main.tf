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

resource "docker_network" "agent_mesh_net" {
  name = "agent_mesh_net"
  driver = "bridge"
  ipam_config {
    subnet = var.agent_mesh_cidr
  }
}
