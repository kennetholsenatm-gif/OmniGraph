# Integrations: data, telemetry, and security

## Secret backends (zero-disk)

- **HashiCorp Vault** — AppRole, Kubernetes auth, or OIDC/JWT where applicable.
- **AWS Secrets Manager** — IAM/OIDC in CI.
- **SOPS** — Encrypted files in Git; decrypt only in runner memory.

OmniGraph fetches secrets at runtime, injects them as **environment variables** for the runner process, and **redacts** env-derived secret substrings in captured stdout/stderr from `ExecRunner` and `ContainerRunner`. Secrets are not written to `.env` or `terraform.tfvars` on disk ([ADR 003](adr/003-memory-only-secrets.md)).

## CMDB and telemetry

| Source | Use in OmniGraph |
|--------|-------------------|
| **NetBox** | IPAM, sites, devices, intended logical state; webhook or poll for updates |
| **Zabbix / Prometheus** | Live health and metrics for nodes |

Purpose: populate **gray / unchanged** nodes in the dependency graph so operators see where new infrastructure attaches relative to existing estate.

## NetBox sync webhook payloads

OmniGraph owns a **versioned** contract for a **custom receiver** (your webhook), not every field of the NetBox REST API.

### Version `omnigraph/netbox-sync/v1` (default CLI)

| Field | Required | Notes |
|-------|----------|--------|
| `apiVersion` | set automatically | constant `omnigraph/netbox-sync/v1` |
| `action` | yes | e.g. `create`, `upsert`, `delete` |
| `ip` or `cidr` | one required | validated with `net.ParseIP` / `net.ParseCIDR` |
| `role` | no | free-form label |
| `siteId`, `deviceId` | no | integer foreign keys when your receiver maps them |
| `environment` | no | e.g. `staging` |
| `idempotencyKey` | no | duplicated on the wire in JSON and as header `X-Omnigraph-Idempotency-Key` when using `omnigraph netbox sync --idempotency-key` |

Example:

```json
{
  "apiVersion": "omnigraph/netbox-sync/v1",
  "action": "upsert",
  "ip": "10.0.5.21",
  "cidr": "10.0.5.0/24",
  "role": "web-server",
  "siteId": 12,
  "deviceId": 340,
  "environment": "production",
  "idempotencyKey": "apply-2025-03-21T12:00:00Z"
}
```

CLI:

```bash
omnigraph netbox sync --url https://receiver.example/hooks/netbox \
  --payload-version v1 \
  --action upsert \
  --ip 10.0.5.21 \
  --cidr 10.0.5.0/24 \
  --role web-server \
  --site-id 12 \
  --device-id 340 \
  --environment production \
  --idempotency-key apply-2025-03-21
```

Use `--payload-version legacy` for the older illustrative shape (no `apiVersion`):

```json
{
  "action": "create",
  "ip": "10.0.5.21",
  "role": "web-server"
}
```

Recommended receiver behavior: accept `Content-Type: application/json`, validate `apiVersion`, enforce auth (e.g. signed secret or mTLS), and treat `X-Omnigraph-Idempotency-Key` as an idempotency token for safe retries.

## Triangulated inventory (pull path)

For **NetBox / NetDisco / Zabbix** (and similar) merged into one normalized model, see [inventory-sources.md](inventory-sources.md) and schema [`schemas/inventory-source.v1.schema.json`](../schemas/inventory-source.v1.schema.json) (`omnigraph/inventory-source/v1`). This complements the NetBox **push** webhook above.

## Pipeline run artifact

CI may publish **omnigraph/run/v1** JSON for a multi-step timeline UI: [run-v1.md](run-v1.md), schema [`schemas/run.v1.schema.json`](../schemas/run.v1.schema.json).

## Enterprise identity (Keycloak, FreeIPA, RBAC)

OmniGraph is designed to deploy in **enterprise** environments where a single static API token is insufficient.

| Source | Role in OmniGraph |
|--------|-------------------|
| **Keycloak** | OIDC / OAuth2 broker: validate access tokens for `serve`, future webhooks, and CI callbacks; map **realm roles** and **client roles** to OmniGraph permissions. |
| **FreeIPA** | LDAP directory (and optional Kerberos at the edge): resolve **group membership** (`memberOf`, IPA user groups) into `Subject.Groups` for RBAC mapping. |

**RBAC:** Permission constants live in `internal/identity` (e.g. `serve:inventory:read`, `ir:emit`, `serve:host-ops:write`). `ClaimMapper` expands directory groups and OIDC roles into effective permission sets; `StaticRBAC` supports bootstrap and CI service accounts. See [ADR 005](adr/005-native-ir-enterprise-identity.md).

**Operational requirements:** TLS for LDAPS/OIDC, minimal LDAP bind ACLs, no long-lived directory passwords in config files (prefer secret stores per [ADR 003](adr/003-memory-only-secrets.md)). `OMNIGRAPH_SERVE_TOKEN` remains supported as **bootstrap** when OIDC/LDAP is not configured; production should prefer IdP-backed subjects and audit logs keyed by principal id.

**Infrastructure IR:** Enterprise policy can require `omnigraph ir validate` in CI and gate `ir:emit` separately from raw orchestrate privileges—see [omnigraph-ir.md](omnigraph-ir.md).
