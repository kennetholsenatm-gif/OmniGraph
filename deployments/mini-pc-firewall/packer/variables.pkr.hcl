# Override on CLI: packer build -var='cloud_image_checksum=sha256:...'

variable "cloud_image_url" {
  type        = string
  description = "AlmaLinux Generic Cloud qcow2 (x86_64)."
  default     = "https://repo.almalinux.org/almalinux/10/cloud/x86_64/images/AlmaLinux-10-GenericCloud-x86_64-latest.x86_64.qcow2"
}

variable "cloud_image_checksum" {
  type        = string
  description = "Checksum for cloud_image_url (sha256:... or file:URL to CHECKSUM file)."
  default     = "file:https://repo.almalinux.org/almalinux/10/cloud/x86_64/images/CHECKSUM"
}

variable "disk_size_gb" {
  type        = number
  description = "Virtual disk size (GB) after resize."
  default     = 32
}

variable "memory_mb" {
  type    = number
  default = 4096
}

variable "cpus" {
  type    = number
  default = 4
}

variable "ssh_username" {
  type    = string
  default = "packer"
}

variable "ssh_password" {
  type        = string
  sensitive   = true
  description = "Temporary SSH password during image build only."
  default     = "packer"
}
