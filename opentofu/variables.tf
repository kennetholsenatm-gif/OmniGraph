variable "tooling_gitea_cidr" {
  description = "Subnet for Gitea tooling segment"
  type        = string
  default     = "100.64.1.0/24"
}

variable "tooling_n8n_cidr" {
  description = "Subnet for n8n tooling segment"
  type        = string
  default     = "100.64.2.0/24"
}

variable "tooling_zammad_cidr" {
  description = "Subnet for Zammad tooling segment"
  type        = string
  default     = "100.64.3.0/24"
}

variable "tooling_bitwarden_cidr" {
  description = "Subnet for Bitwarden (Vaultwarden) tooling segment"
  type        = string
  default     = "100.64.4.0/24"
}

variable "tooling_gateway_cidr" {
  description = "Subnet for Single Pane of Glass gateway (Traefik, dashboard)"
  type        = string
  default     = "100.64.5.0/24"
}

variable "msg_backbone_cidr" {
  description = "Subnet for messaging backbone"
  type        = string
  default     = "100.64.10.0/24"
}

variable "iam_cidr" {
  description = "Subnet for IAM"
  type        = string
  default     = "100.64.20.0/24"
}

variable "tooling_portainer_cidr" {
  description = "Subnet for Portainer"
  type        = string
  default     = "100.64.6.0/24"
}

variable "llm_net_cidr" {
  description = "Subnet for LLM / inference tooling (e.g. BitNet gateway)"
  type        = string
  default     = "100.64.7.0/24"
}

variable "chatops_net_cidr" {
  description = "Subnet for ChatOps (e.g. Zulip)"
  type        = string
  default     = "100.64.8.0/24"
}

variable "freeipa_cidr" {
  description = "Subnet for FreeIPA"
  type        = string
  default     = "100.64.21.0/24"
}

variable "discovery_net_cidr" {
  description = "Subnet for full-stack discovery (NetBox, NetDISCO, Dependency-Track, etc.)"
  type        = string
  default     = "100.64.40.0/24"
}

variable "agent_mesh_cidr" {
  description = "Subnet for agent mesh"
  type        = string
  default     = "100.64.30.0/24"
}

variable "sdn_lab_net_cidr" {
  description = "Subnet for SDN lab (VyOS leg, sFlow sources, n8n SDN attachment)"
  type        = string
  default     = "100.64.50.0/24"
}

variable "telemetry_net_cidr" {
  description = "Subnet for Prometheus, Grafana, sFlow-RT; Traefik attaches for /grafana and /sflow-rt"
  type        = string
  default     = "100.64.51.0/24"
}

variable "docs_net_cidr" {
  description = "Subnet for Docsify (Nginx) static docs; Traefik attaches for /docs"
  type        = string
  default     = "100.64.52.0/24"
}

variable "sonarqube_net_cidr" {
  description = "Subnet for SonarQube; Traefik attaches for /sonarqube"
  type        = string
  default     = "100.64.53.0/24"
}

variable "siem_net_cidr" {
  description = "Subnet for Wazuh stack; Traefik attaches for /wazuh"
  type        = string
  default     = "100.64.54.0/24"
}
