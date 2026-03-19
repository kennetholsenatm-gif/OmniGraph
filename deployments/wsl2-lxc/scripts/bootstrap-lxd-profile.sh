#!/usr/bin/env bash
# Create or update LXD profile "docker-nesting" (Docker-in-LXC) without relying on stdin edit.
set -euo pipefail
if ! command -v lxc >/dev/null 2>&1; then
  echo "lxc not found. Install LXD/Incus." >&2
  exit 1
fi
if ! lxc profile show docker-nesting >/dev/null 2>&1; then
  lxc profile create docker-nesting
fi
lxc profile set docker-nesting description 'Nested Docker (DevSecOps pipeline)'
lxc profile set docker-nesting security.nesting true
lxc profile set docker-nesting security.syscalls.intercept.mknod true
lxc profile set docker-nesting security.syscalls.intercept.setxattr true
lxc profile set docker-nesting 'linux.kernel_modules' 'ip_tables,ip6_tables,iptable_nat,iptable_mangle,iptable_filter,nf_nat,nf_conntrack,bridge,br_netfilter,overlay'
RAW_LXC="$(printf '%s\n' 'lxc.cgroup2.devices.allow = c 10:200 rwm' 'lxc.apparmor.profile = unconfined')"
lxc profile set docker-nesting raw.lxc "$RAW_LXC"
echo "Profile docker-nesting configured."
echo "Attach to instance: lxc profile assign <name> docker-nesting"
