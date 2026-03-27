#!/usr/bin/env bash
# Apply sane defaults to the default Incus/LXD managed bridge so containers get NAT + DHCP + DNS.
# Run on the host where incus/lxc is installed (often WSL Alma).
set -euo pipefail

BR="${INCUS_BRIDGE:-}"
if [[ -z "${BR}" ]]; then
  if command -v incus >/dev/null 2>&1; then
    BR="incusbr0"
  elif command -v lxc >/dev/null 2>&1; then
    BR="lxdbr0"
  else
    echo "ERROR: neither incus nor lxc found in PATH." >&2
    exit 1
  fi
fi

NS="${DNS_SERVERS:-1.1.1.1,8.8.8.8}"

CLI=""
if command -v incus >/dev/null 2>&1; then
  CLI="incus"
elif command -v lxc >/dev/null 2>&1; then
  CLI="lxc"
else
  echo "ERROR: neither incus nor lxc found in PATH." >&2
  exit 1
fi

echo "Configuring bridge ${BR} via ${CLI} (dns.nameservers=${NS})..."
sudo "${CLI}" network set "${BR}" \
  dns.mode=managed \
  ipv4.nat=true \
  ipv4.dhcp=true \
  "dns.nameservers=${NS}"

echo "Done. Verify: sudo ${CLI} network show ${BR}"
