# Implementation Plan: Termix / NetBox Sync

## Objective

Automate sync from NetBox (SoT) to Termix SSH Host Manager: folders, tags, RBAC, and credential linking via Varlock only.

## Deliverables (cline_coder_agent)

| # | Item | Status / Location |
|---|------|-------------------|
| 1 | **ADR** | docs/architecture/ADR-002-termix-netbox-sync.md |
| 2 | **Sync script** | scripts/sync_to_termix.py (NetBox delta -> Termix JSON/API; no hardcoded creds) |
| 3 | **Varlock** | termix.env.schema (@sensitive @required for SSH keys, API tokens) |
| 4 | **n8n** | Workflow: Gitea webhook or NetBox webhook -> sync_to_termix.py -> Termix |

## Security Critique

- Zero Trust: Termix SSH backend must use modern DH group key exchange only.
- CMMC AC: RBAC assignments must align with least-privilege SSH access.

## Doc_Agent (after PR merge)

- Tagging taxonomy: NetBox Site/Role/Tenant -> Termix Folder/Tags (NoteBookLM-style chunking).
- Event flow table: Gitea webhook -> n8n -> Termix host create/update/delete.

## Execution Order

1. Implement sync_to_termix.py with real NetBox API client and Termix API or import file output.
2. Load NETBOX_* and TERMIX_* (and TERMIX_SSH_*) from Varlock in n8n or wrapper script.
3. Add n8n workflow; trigger on webhook or schedule.
4. Doc_Agent: publish taxonomy and event flow to Wiki.
