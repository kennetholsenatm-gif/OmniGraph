#!/bin/bash
# Runs as root inside the Packer VM (AlmaLinux 10).
set -euxo pipefail
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

# Grow root FS if the qcow2 was expanded by Packer
if command -v growpart >/dev/null 2>&1; then
  ROOT_DEV="$(findmnt -n -o SOURCE / | sed 's/\[.*\]//')"
  DISK="${ROOT_DEV//[0-9]/}"
  PART="${ROOT_DEV//[^0-9]/}"
  if [[ -n "${DISK}" && -n "${PART}" ]]; then
    growpart "${DISK}" "${PART}" || true
  fi
fi
if command -v xfs_growfs >/dev/null 2>&1 && findmnt -n -o FSTYPE / | grep -q xfs; then
  xfs_growfs / || true
elif command -v resize2fs >/dev/null 2>&1 && [[ -n "${ROOT_DEV:-}" ]]; then
  resize2fs "${ROOT_DEV}" || true
fi

dnf install -y epel-release
dnf install -y --allowerasing \
  qemu-guest-agent cloud-init firewalld openssh-server \
  curl tar gzip git jq openssl \
  python3 python3-libselinux \
  bridge-utils iproute nftables iptables-nft \
  kmod

systemctl enable qemu-guest-agent || true
systemctl enable firewalld || true
systemctl enable sshd || systemctl enable ssh || true

# Router / L3 host (VyOS LXC / bridges)
cat >/etc/sysctl.d/90-mini-pc-router.conf <<'EOF'
net.ipv4.ip_forward = 1
net.ipv6.conf.all.forwarding = 1
EOF
sysctl --system || sysctl -p /etc/sysctl.d/90-mini-pc-router.conf || true

# Load common bridge / netfilter modules at boot
cat >/etc/modules-load.d/mini-pc-incus.conf <<'EOF'
bridge
br_netfilter
nf_nat
nf_conntrack
overlay
EOF

# Optional: leave Incus install to Ansible (repo URLs vary by EL release).
dnf clean all || true

echo "bootstrap complete"
