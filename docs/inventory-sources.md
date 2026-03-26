# Triangulated inventory: `omnigraph/inventory-source/v1`

This document specifies how OmniGraph **ingests** NetBox, NetDisco, Zabbix (and similar) into a **normalized snapshot**, and how snapshots are **merged** into a single source of truth for dynamic inventory APIs and graph visualization.

**Normative schema:** [`schemas/inventory-source.v1.schema.json`](../schemas/inventory-source.v1.schema.json).

## Goals

- One **machine-readable** shape per pull from each system (no ad-hoc JSON per connector).
- Deterministic **merge rules** so CI and `omnigraph serve` can reproduce the same merged inventory.
- Correlation keys so **planned** IaC hosts can attach to **observed** CMDB and telemetry rows.

## Snapshot document

Each connector emits an `InventorySnapshot`:

- `metadata.source`: `netbox` | `netdisco` | `zabbix` | `prometheus` | `manual` | `merged`
- `metadata.sourceInstance`: disambiguates multiple NetBox or Zabbix servers.
- `spec.records[]`: [`inventoryRecord`](../schemas/inventory-source.v1.schema.json) entries.

### Record fields (usage)

| Field | Purpose |
|-------|---------|
| `id` | Stable within the source (string form of primary key). |
| `recordType` | Distinguishes hosts, interfaces, L2 attachments, etc. |
| `ansibleHost` | Preferred value for Ansible dynamic inventory when the row is a target. |
| `confidence` | Drives merge precedence (see below). |
| `liveness` | Populated from Zabbix/Prometheus for **reachability**; NetBox may use `unknown`. |
| `links` | Store cross-system IDs (`zabbixHostId`, `netboxDeviceId`) after correlation. |

## Merge rules (deterministic)

Merged output is itself an `InventorySnapshot` with `metadata.source: merged` and `metadata.generatedAt` set at merge time.

### Precedence by `confidence`

When two records **represent the same endpoint** (matched by correlation; see below), field-level precedence is:

1. **`authoritative`** wins over `high` over `medium` over `low` over `unknown`.
2. **NetBox**-origin records should normally be tagged **`authoritative`** for **intent** fields: `site`, `role`, `primaryIpv4` (if NetBox is IPAM of record), custom attributes that define policy.
3. **Zabbix** / **Prometheus** should supply **`liveness`** and operational labels; they override NetBox when the field is `liveness` only.
4. **NetDisco** (L2/L3 discovery) should populate **`port_attachment`** / **`interface`** rows and **`attributes`** (switch, port, VLAN); it does not override NetBox **role** unless policy marks discovery as authoritative for that attribute.

### Correlation (matching endpoints)

Implementations should try, in order:

1. **Explicit links:** `links.netboxDeviceId`, `links.zabbixHostId`, etc., when both sides exist.
2. **IP equality:** `primaryIpv4` / `primaryIpv6` normalized (no zone id for IPv6 comparison policy TBD).
3. **Name heuristics:** longest common hostname / FQDN token match with configurable ambiguity threshold (reject on collision).

Unmatched records remain in the merged snapshot with their original `source` traceable via a convention in `attributes` (e.g. `attributes._omnigraph_originSource`).

### Conflicts

When two **authoritative** records disagree on the same scalar field (e.g. two NetBox imports), the merge **must** emit a conflict entry (implementation detail: separate `omnigraph/inventory-merge/v1` report or `records[].attributes._omnigraph_conflict`); automated consumers should not silently pick a winner without policy.

## Relationship to existing OmniGraph pieces

- **NetBox push** today uses [`omnigraph/netbox-sync/v1`](integrations.md) for webhooks; **pull** NetBox REST into `InventorySnapshot` is the complementary path.
- **Telemetry** graph merge (see `internal/telemetry` and [integrations.md](integrations.md)) can consume merged **host** rows to align `kind: telemetry` nodes with `ansible_host`.
- **`GET /api/v1/inventory`** (`omnigraph serve --enable-inventory-api` with `--auth-token`): returns aggregated Terraform/OpenTofu state hosts for a workspace (`path` query, `format=json|ini|ansible-json`); same derivation as `POST /api/v1/workspace/summary` state inventory. **Future:** serve merged `InventorySnapshot` rows (NetBox / NetDisco / Zabbix) using the merge rules above.

## Security

- Snapshots may contain sensitive names or IPs; treat as **internal** data, redact in logs, and scope API access (see posture / serve threat model in [architecture](architecture.md)).
