# ADR-002: Termix Population & SSH Host Management from NetBox

## Status

Proposed. Sync mechanism and RBAC mapping to be implemented by sync_to_termix.py and n8n.

## Context

Infrastructure SoT (NetBox) must drive Termix Server Management so that:

- New/modified/decommissioned assets in NetBox are reflected in Termix (SSH host manager).
- Assets are grouped (folders) and tagged; credentials are linked via centralized secrets (Varlock), not hardcoded.

## Decision (Orchestrator Selection)

| Concern | Choice | Justification |
|--------|--------|----------------|
| **Ingestion pathway** | **n8n webhook/API translation layer** | NetBox webhook or polling → n8n → script (sync_to_termix.py) → Termix Data Import API or direct API. Keeps schema evolution in one place and allows decommission pruning. |
| **Asset mapping** | NetBox **Site** → Termix **Folder**; **Device Role** + **Tenant** → **Tags**; **Name/Primary IP** → host entry. | Deterministic mapping; RBAC groups derived from Tenant or custom field (e.g. security_enclave: OT | IT | Edge). |
| **Decommission** | NetBox device status "offline" or "decommissioned" → sync script removes or disables host in Termix. | Same pipeline; delta includes deletions. |

## Sync Mechanism

1. **Trigger:** Gitea webhook (config change) or NetBox webhook / n8n schedule.
2. **n8n:** Fetch NetBox API (devices, sites, roles, tenants); compute delta (new, modified, to-remove).
3. **sync_to_termix.py:** Receives delta; formats Termix import schema (JSON); assigns RBAC groups by enclave; calls Termix API or writes import file. SSH keys/creds from Varlock only (no defaults in code).
4. **Prune:** Devices no longer present or status decommissioned → removed from Termix.

## RBAC & Folder Hierarchy

- **Folder:** One per NetBox Site (e.g. `site-dc1`, `site-edge`).
- **Tags:** From Device Role + Tenant (e.g. `role:router`, `tenant:acme`).
- **RBAC groups:** From security enclave (OT, IT, Edge) — least-privilege; CMMC AC alignment.

## Zero Trust & Security Critique

- Termix SSH backend: modern DH group key exchange only (validated by Critique Agent).
- RBAC assignments verified against CMMC 2.0 Access Control (AC).

## References

- sync_to_termix.py (NetBox delta → Termix structure; Varlock for creds).
- termix.env.schema (Varlock; SSH keys, API tokens @sensitive @required).
- Doc_Agent: Tagging taxonomy (NetBox → Termix); event flow table (Gitea webhook → n8n → Termix).
