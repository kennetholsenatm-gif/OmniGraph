#!/usr/bin/env bash
# Optional host-side plumbing after: docker compose -f docker-compose.network.yml up -d
# Target: AlmaLinux 10+ (or RHEL-family) with openvswitch RPM + Docker.
# OVS runs in-container; attaching container veths to br-int often uses ovs-docker or manual veth pairs.
#
# Usage:
#   sudo ./scripts/sdn-plumb.sh [BRIDGE_NAME]
# Default BRIDGE_NAME: br-int
#
set -euo pipefail
BRIDGE="${1:-br-int}"
OVS_CONTAINER="${OVS_CONTAINER:-devsecops-openvswitch}"

echo "[sdn-plumb] Using OVS container: $OVS_CONTAINER bridge: $BRIDGE"
echo "[sdn-plumb] Ensure OVS is up: docker compose -f docker-compose/docker-compose.network.yml up -d"

if ! docker ps --format '{{.Names}}' | grep -qx "$OVS_CONTAINER"; then
  echo "[sdn-plumb] ERROR: container $OVS_CONTAINER not running." >&2
  exit 1
fi

# Create integration bridge inside OVS (idempotent)
docker exec "$OVS_CONTAINER" ovs-vsctl --may-exist add-br "$BRIDGE" || true

echo ""
echo "=== Manual veth / Docker port attach (example) ==="
cat <<'EOF'
# Pattern A — ovs-docker (if installed on host):
#   ovs-docker add-port br-int eth1 <container_id> --ipaddress=10.200.0.2/24
#
# Pattern B — host veth into container namespace (replace PID with container init pid):
#   ip link add veth-h0 type veth peer name veth-c0
#   ip link set veth-h0 up
#   PID=$(docker inspect -f '{{.State.Pid}}' devsecops-vyos)
#   ip link set veth-c0 netns $PID
#   nsenter -t $PID -n ip link set veth-c0 name eth2
#   nsenter -t $PID -n ip link set eth2 up
#   ovs-vsctl add-port br-int veth-h0
#
# Pattern C — VXLAN from host OVS to remote VTEP (lab overlay):
#   ovs-vsctl add-port br-int vxlan0 -- set interface vxlan0 type=vxlan options:remote_ip=<peer> options:key=<vni>
#
# Verify:
#   docker exec devsecops-openvswitch ovs-vsctl show
EOF

echo "[sdn-plumb] Done (bridge ensured inside container)."

echo ""
echo "=== OVS mirror (SPAN) to Suricata tap (example; adjust port names from ovs-vsctl show) ==="
cat <<'OVS_MIRROR'
# Run inside OVS container: docker exec -it devsecops-openvswitch bash
# Replace br-int, vxlan0, suri-tap with your bridge and port names.

ovs-vsctl --may-exist add-br br-int
ovs-vsctl --may-exist add-port br-int suri-tap -- set Interface suri-tap type=internal
ip link set suri-tap up || true

VXLAN_UUID=$(ovs-vsctl get Port vxlan0 _uuid)
TAP_UUID=$(ovs-vsctl get Port suri-tap _uuid)
ovs-vsctl -- --id=@m create Mirror name=vxlan-span \
  select-src-port="$VXLAN_UUID" \
  select-dst-port="$VXLAN_UUID" \
  output-port="$TAP_UUID" \
  -- set Bridge br-int mirrors=@m

# Feed Suricata: attach a veth peer to suri-tap into the Suricata namespace, or run Suricata with
# network_mode: host and -i <host-if-receiving-mirror> (see docker-compose.network.yml suricata).
OVS_MIRROR
