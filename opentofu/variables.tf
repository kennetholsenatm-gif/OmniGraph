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

variable "agent_mesh_cidr" {
  description = "Subnet for agent mesh"
  type        = string
  default     = "100.64.30.0/24"
}
