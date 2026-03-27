# AlmaLinux 10 cloud image → golden QCOW2 for mini PC Incus host (VyOS / PacketFence edge).
# Requires: Linux + KVM + qemu-system-x86_64. Run: packer init . && packer build .

packer {
  required_version = ">= 1.9.0"
  required_plugins {
    qemu = {
      version = ">= 1.0.9"
      source  = "github.com/hashicorp/qemu"
    }
  }
}

source "qemu" "alma10_incus_host" {
  # Existing cloud qcow2
  disk_image           = true
  iso_url              = var.cloud_image_url
  iso_checksum         = var.cloud_image_checksum
  # qemu plugin accepts e.g. "32G"
  disk_size            = "${var.disk_size_gb}G"
  disk_interface       = "virtio"
  format               = "qcow2"
  net_device           = "virtio-net"
  accelerator          = "kvm"
  headless             = true
  communicator         = "ssh"
  ssh_username         = var.ssh_username
  ssh_password         = var.ssh_password
  ssh_timeout          = "45m"
  ssh_handshake_attempts = 50
  ssh_wait_timeout     = "45m"
  boot_wait            = "10s"

  # cloud-init nocloud seed (user + password for Packer SSH)
  cd_files = [
    "${path.root}/http/meta-data",
    "${path.root}/http/user-data"
  ]

  qemuargs = [
    ["-machine", "accel=kvm"],
    ["-cpu", "host"]
  ]

  vm_name          = "alma10-incus-host"
  output_directory = "${path.root}/output-alma10-incus-host"
}

build {
  sources = ["source.qemu.alma10_incus_host"]

  provisioner "shell" {
    execute_command = "echo 'packer' | sudo -S sh -c '{{ .Vars }} {{ .Path }}'"
    scripts = [
      "${path.root}/scripts/00-bootstrap.sh"
    ]
  }

  post-processor "manifest" {
    output     = "${path.root}/output-alma10-incus-host/packer-manifest.json"
    strip_path = true
  }
}
