# Incus / LXD bridge networking + DNS (lab WSL / Alma)

Containers need a **managed bridge** with **NAT**, **DHCP**, and working **DNS** so `dnf`, `curl`, and Ansible can reach the internet.

## What this repo does automatically

`ansible/roles/lxd_devsecops_stack/tasks/managed_bridge.yml` ensures a bridge exists (default **`incusbr0`** for `incus`, **`lxdbr0`** for `lxc`) and applies:

- `dns.mode=managed`
- `ipv4.nat=true`
- `ipv4.dhcp=true`
- `dns.nameservers=<your list>` (default `1.1.1.1,8.8.8.8`)

Variables (override in inventory or `-e`):

- `lxd_bridge_tune_dns` (default `true`)
- `lxd_bridge_dns_nameservers` (default Cloudflare + Google)
- `lxd_ensure_managed_bridge` (default `true`)

## One-shot on the Incus host (manual)

Pick the bridge name (`incus network list` — usually `incusbr0`):

```bash
sudo incus network set incusbr0 dns.mode=managed ipv4.nat=true ipv4.dhcp=true dns.nameservers=1.1.1.1,8.8.8.8
```

Or use the helper:

```bash
./scripts/setup-incus-network-dns.sh
```

## Verify from a container

```bash
incus exec <instance> -- getent hosts mirrors.almalinux.org
incus exec <instance> -- cat /etc/resolv.conf
```

You should see **either** upstream DNS servers (if advertised via DHCP) **or** the bridge IP as resolver, with dnsmasq forwarding upstream.

## If DNS still fails

1. **Host firewall** — Incus inserts `iptables`/`nft` rules; on **firewalld** see [Network bridge + firewalld](https://linuxcontainers.org/incus/docs/main/howto/network_bridge_firewalld/).
2. **systemd-resolved on the host** — If port `53` is owned by `systemd-resolved`, see [Bridge + systemd-resolved](https://linuxcontainers.org/incus/docs/main/howto/network_bridge_resolved/).
3. **WSL2** — Confirm the **Linux** environment running Incus can reach the internet (`ping 1.1.1.1`, `getent hosts mirrors.almalinux.org`). Windows DNS issues do not always propagate into WSL; fixing WSL `/etc/resolv.conf` or `systemd-resolved` may be required **outside** the container.
4. **Static resolv.conf in CT** — As a last resort, the playbooks can still use `lxd_dns_fallback_enabled` + `lxd_dns_fallback_servers` inside instances (see `lxd_devsecops_stack` defaults), or `lxd_sync_host_resolv_conf=true` to push the controller’s `/etc/resolv.conf` into each instance.

## Advanced: extra dnsmasq options

To append custom dnsmasq lines (e.g. extra `server=` entries), use:

```bash
sudo incus network set incusbr0 raw.dnsmasq='server=1.1.1.1
server=8.8.8.8'
```

(Prefer a single `network set` with a here-doc or file if you need many lines.)
