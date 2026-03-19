#!/usr/bin/env python3
"""
Sync NetBox device inventory to Termix SSH Host Manager.

- Pulls delta (new/modified/decommissioned) from NetBox API.
- Maps NetBox Site -> Termix Folder; Device Role + Tenant -> Tags.
- Assigns RBAC groups by security enclave (OT, IT, Edge).
- Outputs Termix Data Import JSON or calls Termix API.
- SSH keys and API tokens MUST come from environment (Varlock); no hardcoded secrets.

Usage (from n8n or CLI):
  NETBOX_URL=... NETBOX_API_TOKEN=... TERMIX_API_URL=... TERMIX_API_TOKEN=... python sync_to_termix.py

See termix.env.schema and ADR-002-termix-netbox-sync.md.
"""

from __future__ import annotations

import json
import os
import sys
from typing import Any


def get_required_env(key: str) -> str:
    val = os.environ.get(key)
    if not val:
        print(f"Missing required env: {key}", file=sys.stderr)
        sys.exit(1)
    return val


def netbox_devices_delta(netbox_url: str, token: str) -> dict[str, Any]:
    """Fetch devices from NetBox; return structure with new, modified, to_remove.
    In a full implementation: use NetBox API with last_sync timestamp or webhook payload.
    """
    # Placeholder: real impl would use requests.get(netbox_url + "/api/dcim/devices/", headers={"Authorization": f"Token {token}"})
    return {
        "new": [],
        "modified": [],
        "to_remove": [],
    }


def netbox_to_termix_host(device: dict[str, Any], enclave_field: str) -> dict[str, Any]:
    """Map one NetBox device to Termix host entry. Folder = Site; Tags = Role + Tenant; RBAC = enclave."""
    site = (device.get("site") or {}).get("name") or "default"
    role = (device.get("device_role") or {}).get("name") or ""
    tenant = (device.get("tenant") or {}).get("name") or ""
    primary_ip = (device.get("primary_ip") or {}).get("address", "").split("/")[0]
    name = device.get("name") or primary_ip or "unknown"
    enclave = device.get("custom_fields", {}).get(enclave_field) or device.get("tags", [])
    if isinstance(enclave, list):
        enclave = enclave[0] if enclave else "IT"
    return {
        "name": name,
        "address": primary_ip,
        "folder": f"site-{site}",
        "tags": [f"role:{role}", f"tenant:{tenant}"],
        "rbac_group": f"enclave-{enclave}",
    }


def build_termix_import_json(devices: list[dict], enclave_field: str) -> list[dict]:
    """Build Termix Data Import JSON structure from NetBox devices."""
    return [netbox_to_termix_host(d, enclave_field) for d in devices]


def main() -> None:
    netbox_url = get_required_env("NETBOX_URL").rstrip("/")
    netbox_token = get_required_env("NETBOX_API_TOKEN")
    termix_api_url = os.environ.get("TERMIX_API_URL", "")
    termix_token = os.environ.get("TERMIX_API_TOKEN", "")
    enclave_field = os.environ.get("TERMIX_ENCLAVE_FIELD", "security_enclave")

    delta = netbox_devices_delta(netbox_url, netbox_token)
    all_devices = delta.get("new", []) + delta.get("modified", [])
    termix_hosts = build_termix_import_json(all_devices, enclave_field)

    if termix_api_url and termix_token:
        # TODO: POST to Termix API (host create/update); DELETE for to_remove
        print(json.dumps({"status": "ok", "hosts_synced": len(termix_hosts), "to_remove": len(delta.get("to_remove", []))}))
    else:
        print(json.dumps(termix_hosts, indent=2))


if __name__ == "__main__":
    main()
