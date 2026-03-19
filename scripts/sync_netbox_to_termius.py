#!/usr/bin/env python3
"""
NetBox → Termius sync: fetch device/VM inventory and emit Termius-friendly JSON or POST to Teams API.

Secrets only from environment (Varlock/Vault-injected). No hardcoded passwords or keys.

Usage:
  export NETBOX_URL=http://netbox:8080 NETBOX_API_TOKEN=...
  python sync_netbox_to_termius.py --format termius-json -o termius-hosts.json
  python sync_netbox_to_termius.py --format teams-api --dry-run

See docs/NETBOX_TERMIUS_SYNC.md for taxonomy and pruning policy.
"""

from __future__ import annotations

import argparse
import json
import os
import sys
import urllib.error
import urllib.request
from abc import ABC, abstractmethod
from typing import Any, Dict, Iterator, List, Optional


def _env(name: str, default: Optional[str] = None) -> Optional[str]:
    v = os.environ.get(name)
    if v is not None and str(v).strip() != "":
        return v
    return default


class TermiusTransport(ABC):
    """Isolate Termius delivery behind a small adapter (import file vs Teams API)."""

    @abstractmethod
    def deliver(self, payload: Dict[str, Any], dry_run: bool) -> None:
        raise NotImplementedError


class TermiusImportJsonTransport(TermiusTransport):
    """Write a JSON document operators can import or transform (Termius UI / CLI)."""

    def __init__(self, output_path: str) -> None:
        self.output_path = output_path

    def deliver(self, payload: Dict[str, Any], dry_run: bool) -> None:
        text = json.dumps(payload, indent=2)
        if dry_run:
            print(text)
            return
        with open(self.output_path, "w", encoding="utf-8") as f:
            f.write(text)
        print(f"Wrote {self.output_path}", file=sys.stderr)


class TermiusTeamsApiTransport(TermiusTransport):
    """
    POST JSON to Termius Teams-compatible endpoint.
    Set TERMIUS_TEAMS_API_BASE (e.g. https://api.termius.com) and TERMIUS_API_TOKEN.
    Path defaults to /v1/host-groups/sync — override with TERMIUS_TEAMS_SYNC_PATH if your tenant differs.
    """

    def __init__(self, base_url: str, token: str, path: str) -> None:
        self.base_url = base_url.rstrip("/")
        self.token = token
        self.path = path if path.startswith("/") else f"/{path}"

    def deliver(self, payload: Dict[str, Any], dry_run: bool) -> None:
        url = f"{self.base_url}{self.path}"
        body = json.dumps(payload).encode("utf-8")
        if dry_run:
            print(f"DRY RUN POST {url}\n{body.decode()}", file=sys.stderr)
            return
        req = urllib.request.Request(
            url,
            data=body,
            method="POST",
            headers={
                "Content-Type": "application/json",
                "Authorization": f"Bearer {self.token}",
                "Accept": "application/json",
            },
        )
        try:
            with urllib.request.urlopen(req, timeout=120) as resp:
                out = resp.read().decode("utf-8", errors="replace")
                print(out or "OK", file=sys.stderr)
        except urllib.error.HTTPError as e:
            err = e.read().decode("utf-8", errors="replace")
            print(f"HTTP {e.code}: {err}", file=sys.stderr)
            sys.exit(1)


def netbox_get_json(url: str, token: str) -> Dict[str, Any]:
    req = urllib.request.Request(
        url,
        headers={
            "Authorization": f"Token {token}",
            "Accept": "application/json",
        },
    )
    with urllib.request.urlopen(req, timeout=60) as resp:
        return json.loads(resp.read().decode("utf-8"))


def iter_netbox_devices(base: str, token: str) -> Iterator[Dict[str, Any]]:
    """Paginate /api/dcim/devices/."""
    next_url = f"{base.rstrip('/')}/api/dcim/devices/?limit=100"
    while next_url:
        data = netbox_get_json(next_url, token)
        for item in data.get("results", []):
            yield item
        next_url = data.get("next") or ""


def iter_netbox_vms(base: str, token: str) -> Iterator[Dict[str, Any]]:
    next_url = f"{base.rstrip('/')}/api/virtualization/virtual-machines/?limit=100"
    while next_url:
        data = netbox_get_json(next_url, token)
        for item in data.get("results", []):
            yield item
        next_url = data.get("next") or ""


def _primary_host(device: Dict[str, Any]) -> Optional[str]:
    pip = device.get("primary_ip4") or device.get("primary_ip")
    if isinstance(pip, dict):
        addr = pip.get("address") or pip.get("display") or ""
        if isinstance(addr, str) and addr:
            return addr.split("/")[0].strip()
    return None


def _folder_name(device: Dict[str, Any]) -> str:
    site = (device.get("site") or {}) if isinstance(device.get("site"), dict) else {}
    tenant = (device.get("tenant") or {}) if isinstance(device.get("tenant"), dict) else {}
    site_name = site.get("name") or site.get("slug") or "no-site"
    tenant_name = tenant.get("name") or tenant.get("slug") or ""
    if tenant_name:
        return f"{tenant_name} / {site_name}"
    return str(site_name)


def _tags(device: Dict[str, Any]) -> List[str]:
    tags = device.get("tags") or []
    out: List[str] = []
    if isinstance(tags, list):
        for t in tags:
            if isinstance(t, dict) and t.get("name"):
                out.append(str(t["name"]))
            elif isinstance(t, str):
                out.append(t)
    role = device.get("device_role") or device.get("role")
    if isinstance(role, dict) and role.get("name"):
        out.append(f"role:{role['name']}")
    elif isinstance(role, dict) and role.get("slug"):
        out.append(f"role:{role['slug']}")
    return out


def build_hosts(
    records: List[Dict[str, Any]],
    default_user: str,
    kind: str,
) -> List[Dict[str, Any]]:
    hosts: List[Dict[str, Any]] = []
    for d in records:
        name = d.get("name") or d.get("display") or "unknown"
        hostname = _primary_host(d)
        if not hostname:
            continue
        user = default_user or ""
        port = 22
        hosts.append(
            {
                "name": name,
                "hostname": hostname,
                "port": port,
                "username": user,
                "tags": _tags(d),
                "group": _folder_name(d),
                "source": "netbox",
                "netbox_kind": kind,
                "netbox_id": d.get("id"),
            }
        )
    return hosts


def make_transport(fmt: str, output: str) -> TermiusTransport:
    if fmt == "termius-json":
        return TermiusImportJsonTransport(output)
    if fmt == "teams-api":
        base = _env("TERMIUS_TEAMS_API_BASE", "")
        token = _env("TERMIUS_API_TOKEN", "")
        if not base or not token:
            print(
                "teams-api requires TERMIUS_TEAMS_API_BASE and TERMIUS_API_TOKEN in the environment.",
                file=sys.stderr,
            )
            sys.exit(2)
        path = _env("TERMIUS_TEAMS_SYNC_PATH", "/v1/host-groups/sync")
        return TermiusTeamsApiTransport(base, token, path)
    print(f"Unknown format: {fmt}", file=sys.stderr)
    sys.exit(2)


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument(
        "--format",
        choices=["termius-json", "teams-api"],
        default="termius-json",
        help="termius-json: write file; teams-api: POST to Termius Teams (env-based URL)",
    )
    p.add_argument("-o", "--output", default="termius-hosts.json", help="Output path for termius-json")
    p.add_argument("--dry-run", action="store_true", help="Print payload / skip write or HTTP")
    p.add_argument("--include-vms", action="store_true", help="Include virtualization virtual-machines")
    args = p.parse_args()

    base = _env("NETBOX_URL", "http://netbox:8080")
    token = _env("NETBOX_API_TOKEN", "")
    if not token:
        print("NETBOX_API_TOKEN is required.", file=sys.stderr)
        sys.exit(2)

    default_user = _env("TERMIUS_SSH_USERNAME_DEFAULT", "") or ""

    devices = list(iter_netbox_devices(base, token))
    records: List[Dict[str, Any]] = [("device", x) for x in devices]
    if args.include_vms:
        for vm in iter_netbox_vms(base, token):
            records.append(("virtual-machine", vm))

    hosts: List[Dict[str, Any]] = []
    for kind, obj in records:
        hosts.extend(build_hosts([obj], default_user, kind))

    payload = {
        "version": 1,
        "hosts": hosts,
        "meta": {
            "taxonomy": "group=tenant/site, tags=netbox tags + role:*",
            "prune_policy": "none by default; do not auto-delete Termius entries",
        },
    }

    transport = make_transport(args.format, args.output)
    transport.deliver(payload, dry_run=args.dry_run)


if __name__ == "__main__":
    main()
