output "network_ids" {
  description = "Docker network IDs for compose external reference"
  value = {
    gitea_net        = docker_network.gitea_net.id
    n8n_net          = docker_network.n8n_net.id
    zammad_net       = docker_network.zammad_net.id
    bitwarden_net    = docker_network.bitwarden_net.id
    gateway_net     = docker_network.gateway_net.id
    msg_backbone_net = docker_network.msg_backbone_net.id
    iam_net          = docker_network.iam_net.id
    agent_mesh_net   = docker_network.agent_mesh_net.id
  }
}
