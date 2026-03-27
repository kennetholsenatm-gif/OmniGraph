# Solace queue & topic naming — discovery / BOM ingestion

Broker objects (queues, topic endpoints, ACLs) are typically managed in the **infra** repository (e.g. `discovery-networks.tf`, `devsecops-variables.tf`). This repo defines a **naming contract** so pipeline services (n8n, SAM, Dependency-Track webhooks) stay aligned without duplicating broker state.

## VPN / mesh

| Setting | Example value | Notes |
|---------|----------------|-------|
| Message VPN | `discovery_tools_mesh` | Matches `SOLACE_VPN` in `devsecops.env.schema` when using the discovery mesh. |

## Topic names (Pub/Sub)

| Topic | Producers | Consumers | Payload intent |
|-------|-----------|-----------|----------------|
| `devsecops/discovery/sbom/ingested/v1` | CI, Syft/Trivy upload jobs | n8n, SAM | SBOM document URI + project + commit |
| `devsecops/discovery/vuln/critical/v1` | Dependency-Track notifier | n8n → Zulip | Already used by ChatOps patterns; same VPN |
| `devsecops/discovery/netbox/change/v1` | NetBox webhooks (via n8n relay) | Termius sync, CMDB jobs | Object type, id, action |
| `devsecops/discovery/inventory/full/v1` | Scheduled reconciliation | Audit agents | Optional full-table reference |

Topics are **case-sensitive**; keep versions (`v1`) for forward-compatible evolution.

## Queue names (point-to-point)

| Queue | Bound topic(s) | Consumer | Notes |
|-------|----------------|----------|-------|
| `Q.DEVSECOPS.BOM.INGEST` | `devsecops/discovery/sbom/ingested/v1` | NiFi / custom worker | Durable buffer for BOM writes to Dep-Track or Postgres |
| `Q.DEVSECOPS.NETBOX.EVENTS` | `devsecops/discovery/netbox/change/v1` | n8n bridge or sidecar | Drives [sync_netbox_to_termius.py](../scripts/sync_netbox_to_termius.py) on schedule |
| `Q.DEVSECOPS.DISCOVERY.DLQ` | — | Ops | Dead-letter for failed BOM/event processing |

Adjust prefixes (`DEVSECOPS`) to match your org naming standard in the infra repo.

## Copying into OpenTofu (infra repo)

Use the same strings as **locals** or **variables** in infra:

```hcl
locals {
  discovery_topic_sbom_ingested = "devsecops/discovery/sbom/ingested/v1"
  discovery_queue_bom_ingest    = "Q.DEVSECOPS.BOM.INGEST"
}
```

See also [docs/snippets/solace-discovery-queues.tf](snippets/solace-discovery-queues.tf) for a paste-ready OpenTofu `locals` snippet (keep in **infra** repo, not merged into this repo’s `opentofu/` apply).

## References

- [ADR_FULLSTACK_DISCOVERY.md](ADR_FULLSTACK_DISCOVERY.md)
- [A2A_PAYLOAD_SCHEMA.md](A2A_PAYLOAD_SCHEMA.md) (if extending payloads)
- `devsecops.env.schema` — `SOLACE_*`
