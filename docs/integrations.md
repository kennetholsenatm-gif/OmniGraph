# Integrations: data, telemetry, and security

## Secret backends (zero-disk)

- **HashiCorp Vault** — AppRole, Kubernetes auth, or OIDC/JWT where applicable.
- **AWS Secrets Manager** — IAM/OIDC in CI.
- **SOPS** — Encrypted files in Git; decrypt only in runner memory.

OmniGraph fetches secrets at runtime, injects them as **environment variables** for the runner process, and **masks** known secret values in stdout/stderr. Secrets are not written to `.env` or `terraform.tfvars` on disk ([ADR 003](adr/003-memory-only-secrets.md)).

## CMDB and telemetry

| Source | Use in OmniGraph |
|--------|-------------------|
| **NetBox** | IPAM, sites, devices, intended logical state; webhook or poll for updates |
| **Zabbix / Prometheus** | Live health and metrics for nodes |

Purpose: populate **gray / unchanged** nodes in the dependency graph so operators see where new infrastructure attaches relative to existing estate.

## Example NetBox sync webhook payload

After apply and configuration, the control plane may emit a small JSON payload (shape illustrative):

```json
{
  "action": "create",
  "ip": "10.0.5.21",
  "role": "web-server"
}
```

Exact schema will align with NetBox API versioning and tenant conventions when implemented.
