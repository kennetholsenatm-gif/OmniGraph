# NetBox → Termius sync

## Overview

[scripts/sync_netbox_to_termius.py](../scripts/sync_netbox_to_termius.py) reads **devices** (and optionally **virtual machines**) from the NetBox API and either:

- writes a **Termius-friendly JSON** file (`--format termius-json`), or
- **POSTs** a JSON payload to a **Termius Teams**-style HTTP API (`--format teams-api`), using `TERMIUS_TEAMS_API_BASE` and `TERMIUS_API_TOKEN`.

**No SSH private keys or passwords** are embedded. Set optional `TERMIUS_SSH_USERNAME_DEFAULT` if NetBox does not model per-device SSH users.

## Environment (Varlock / Vault)

| Variable | Required | Purpose |
|----------|----------|---------|
| `NETBOX_URL` | Yes | e.g. `http://netbox:8080` on `discovery_net` |
| `NETBOX_API_TOKEN` | Yes | NetBox user API token |
| `TERMIUS_SSH_USERNAME_DEFAULT` | No | Default SSH username on exported hosts |
| `TERMIUS_TEAMS_API_BASE` | For `teams-api` | HTTPS API base per Termius Teams docs |
| `TERMIUS_API_TOKEN` | For `teams-api` | Bearer token (Vault only) |
| `TERMIUS_TEAMS_SYNC_PATH` | No | Default `/v1/host-groups/sync` — replace if your API differs |

Schema: `devsecops.env.schema` (`@env-spec: NETBOX`, `TERMIUS`).

## Taxonomy (folder / tags)

| NetBox concept | Export field | Notes |
|----------------|--------------|--------|
| Tenant + Site | `group` | `"<tenant> / <site>"` or site only if no tenant |
| Tags | `tags` | NetBox tag names |
| Device role | `tags` | Added as `role:<name>` |

Adjust mapping in the script if you use custom fields (e.g. `cf.ssh_user`).

## Pruning / decommission (conservative)

**Default:** the script **never deletes** Termius hosts. Absent devices are **not** auto-removed.

**Recommended:** after NetBox decommission workflow, operators **manually** remove hosts in Termius or use a future `--prune-mode` once Termius list/export APIs are confirmed for your tenant.

**Conservative interim:** tag stale entries in NetBox (`decommissioning`) and filter exports with a NetBox query (extend script) rather than deleting in Termius.

## n8n

Import [n8n-workflows/netbox-to-termius-sync.json](../n8n-workflows/netbox-to-termius-sync.json): scheduled run that executes the Python script on a host or sidecar where `NETBOX_*` and `TERMIUS_*` are injected. Ensure the n8n **Execute Command** path matches your deployment (or replace with an HTTP wrapper service).

Webhook alternative: NetBox → webhook → n8n → run sync (after NetBox change); throttle to avoid API storms.

## Related docs

- [ADR_FULLSTACK_DISCOVERY.md](ADR_FULLSTACK_DISCOVERY.md)
- [TELEPORT_JIT.md](TELEPORT_JIT.md) (separate from Termius; same IAM plane)
- [WIKI_EXPORT/DISCOVERY_AND_TERMIUS_WIKI.md](WIKI_EXPORT/DISCOVERY_AND_TERMIUS_WIKI.md)
