#!/usr/bin/env bash
# Create or update profile "docker-nesting" for LXD/Incus.
set -euo pipefail

if command -v incus >/dev/null 2>&1; then
  CLI="incus"
elif command -v lxc >/dev/null 2>&1; then
  CLI="lxc"
else
  echo "Neither incus nor lxc CLI found. Install LXD or Incus." >&2
  exit 1
fi

if ! "$CLI" profile show docker-nesting >/dev/null 2>&1; then
  "$CLI" profile create docker-nesting
fi

"$CLI" profile set docker-nesting description "Nested Docker (DevSecOps pipeline)"
"$CLI" profile set docker-nesting security.nesting true
"$CLI" profile set docker-nesting security.syscalls.intercept.mknod true
"$CLI" profile set docker-nesting security.syscalls.intercept.setxattr true
"$CLI" profile set docker-nesting "linux.kernel_modules" "ip_tables,ip6_tables,iptable_nat,iptable_mangle,iptable_filter,nf_nat,nf_conntrack,bridge,br_netfilter,overlay"
RAW_LXC="$(printf '%s\n' 'lxc.cgroup2.devices.allow = c 10:200 rwm' 'lxc.apparmor.profile = unconfined')"
"$CLI" profile set docker-nesting raw.lxc "$RAW_LXC"

echo "Profile docker-nesting configured with $CLI."
echo "Attach to instance: $CLI profile assign <name> docker-nesting"
