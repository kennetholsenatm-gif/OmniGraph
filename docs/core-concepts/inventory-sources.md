# Inventory Sources

OmniGraph can aggregate inventory context from multiple sources to support planning, graph generation, and post-apply operations.

## Normalized contract

Machine validation uses **`omnigraph/inventory-source/v1`** ([`schemas/inventory-source.v1.schema.json`](../../schemas/inventory-source.v1.schema.json)). Snapshots declare `metadata.source` (for example `netbox`, `zabbix`, `merged`) and `spec.records[]` with stable ids, record types, and optional cross-source `links`.

## How snapshots are produced today

**NetBox** and **Zabbix** inventory-style pulls are implemented as **WASM integration micro-containers** under [`wasm/plugins/`](../../wasm/plugins/): each plugin calls its vendor API **only** via the host **`http_fetch`** import (allowlisted URLs). The workspace server and other Go packages **do not** call NetBox or Zabbix HTTP APIs directly.

Operators build `.wasm` artifacts, then run:

- `omnigraph integration-run --wasm=path/to/netbox.wasm < run.json`, or
- the authenticated **`POST /api/v1/integrations/run`** API when enabled.

The **`omnigraph/integration-run/v1`** stdin document carries credentials and **`allowedFetchPrefixes`** that must match host configuration for the invocation.

## Common inputs (other sources)

- IaC state outputs
- Static or generated inventory files
- CMDB/device APIs (prefer mapping into **`inventory-source/v1`** via WASM integrations)
- Runtime telemetry snapshots

## Contract reference

Use versioned schema contracts from `schemas/` for machine validation and exchange. Keep source-specific field mappings in environment documentation when they are not part of the shared schema.
