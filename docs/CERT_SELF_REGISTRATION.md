# Certificate-Based Self-Registration with Human-in-the-Loop Approval

## Overview

Users **self-register** by presenting a **client certificate** (or submitting cert metadata). The system creates a **pending** registration and a **Zammad ticket** for approval. **No privileges are granted** until a human approves the ticket in the ticketing workflow. On approval, n8n triggers privilege grant (Keycloak user + realm role, optional LDAP/identity update).

## Principles

- **Certificate-based**: Registration is bound to a client cert (fingerprint, subject DN, or full cert). No password until after approval (optional password or cert-only auth).
- **Zero privilege until approved**: Pending registrations have no Keycloak roles, no access to protected resources. Only after a human approves the linked Zammad ticket are roles assigned.
- **Ticketing workflow**: All approval flows through the existing ITSM/ticketing system (Zammad). Auditable, assignable, and consistent with change management.

## Flow

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────┐
│ User (mTLS/cert)│────▶│ n8n webhook      │────▶│ Zammad      │
│ POST /cert-reg  │     │ create ticket    │     │ ticket      │
│ (cert or fp+CN) │     │ store pending    │     │ "Approve…"  │
└─────────────────┘     └──────────────────┘     └──────┬──────┘
                                                         │
                        ┌────────────────────────────────┘
                        │ Human approves ticket (workflow)
                        ▼
┌─────────────────┐     ┌──────────────────┐
│ Keycloak / IAM   │◀────│ n8n (Zammad      │
│ user + role      │     │ webhook on       │
│ granted          │     │ ticket approved) │
└─────────────────┘     └──────────────────┘
```

1. **Self-registration request**  
   User (or service) calls the registration webhook with certificate evidence:
   - **mTLS**: Request is made with a client certificate; server extracts subject and fingerprint.
   - **Or** POST body with `cert_fingerprint`, `cert_subject_dn`, `cert_subject_cn` (e.g. from a portal that verified the cert).

2. **n8n: create pending + ticket**  
   - Validates request (required: cert fingerprint or subject CN).
   - Creates a **Zammad ticket** with (adjust `group` to your Zammad group name or use `group_id` with a numeric id):
     - Title: e.g. `Approve certificate registration: CN=<cn>`
     - Body (or article): structured registration payload (see [Registration payload](#registration-payload)) so the approver and the approval workflow have full context.
     - Group/state so the workflow can route it (e.g. group "Access Request" or tag "cert-registration").
   - Optionally stores **pending registration** (Vault path `secret/pending_registrations/<ticket_id>` or Zammad ticket custom fields) with: `ticket_id`, `cert_fingerprint`, `cert_subject_dn`, `requested_privilege_level`, `email`, `created_at`.
   - Returns HTTP 202 with `ticket_id` and message: "Registration pending; no privileges until ticket is approved."

3. **Human in the loop**  
   Operator/admin works the ticket in Zammad (review, approve or reject, close with state "approved" or equivalent per your workflow).

4. **Zammad → n8n on approval**  
   Zammad is configured to send a **webhook** to n8n when a ticket is updated/closed (e.g. `POST http://n8n:5678/webhook/zammad-ticket` or a dedicated `webhook/cert-registration-approved`). n8n:
   - Receives the ticket event.
   - Filters for tickets that are **cert-registration** and **approved** (e.g. state closed + tag "approved" or group "Access Request" with closed state).
   - Loads registration context from ticket body or from Vault `secret/pending_registrations/<ticket_id>`.
   - **Grants privileges**: creates or updates the user in Keycloak, assigns the realm role for `requested_privilege_level` (using `privilege_levels.json`), and optionally updates identity store or LDAP. Passwords can be generated and stored in Vault; user can be instructed to set password on first login or use cert-only.
   - Optionally posts a reply to the Zammad ticket: "Privileges granted for CN=…".

5. **Rejection**  
   If the ticket is closed as rejected, no privilege grant; optional cleanup of pending registration.

## Registration Payload

**Request (POST to n8n webhook e.g. `/webhook/cert-registration-request`):**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cert_fingerprint` | string | One of fingerprint or subject | SHA-256 (or SHA-1) fingerprint of client cert (hex or base64). |
| `cert_subject_dn` | string | One of fingerprint or subject | Full subject DN of the certificate. |
| `cert_subject_cn` | string | Recommended | Common Name from cert; used as preferred username (uid) if no `uid` provided. |
| `uid` | string | Optional | Desired username; defaults to `cert_subject_cn` sanitized. |
| `requested_privilege_level` | string | Optional | One of `admin`, `operator`, `developer`, `viewer`, `auditor`; default `viewer`. |
| `email` | string | Optional | Email for the identity (Keycloak, notifications). |
| `cn` | string | Optional | Display name; defaults to CN or uid. |

**Response (202 Accepted):**

```json
{
  "status": "pending",
  "message": "No privileges granted until a human approves the ticket.",
  "ticket_id": "123",
  "ticket_url": "https://zammad.example/ticket/123"
}
```

**Stored with ticket (in body or Vault):**  
Same fields plus `ticket_id`, `created_at`, and optionally `registration_id` (UUID) for idempotency.

## Identity and Registration State (VARLOCK)

- **Identities** in `identities.yaml` / Keycloak represent **approved** users only. Pending registrations are **not** in that list until approved.
- **Registration state** (for tooling that persists it): `pending` | `approved` | `rejected`. Stored in Vault at `secret/pending_registrations/<ticket_id>` or in Zammad ticket custom fields. On approval, the identity is added to Keycloak (and optionally to `identities.yaml` or LDAP) and the registration record is marked `approved` or removed from pending.
- **Certificate binding**: After approval, the user's identity can be linked to the cert (e.g. Keycloak x509 authenticator, or store `cert_fingerprint` in user attributes for audit).

## Zammad Configuration

1. **Webhook / trigger** when a ticket is closed or state changes to "approved" (or your equivalent), POST to n8n:
   - URL: `http://n8n:5678/webhook/zammad-ticket` (shared with other flows) or `http://n8n:5678/webhook/cert-registration-approved`.
   - Payload: Zammad ticket object (id, state, group, title, etc.).

2. **Ticket group/tag**: Use a dedicated group (e.g. "Access Request") or tag **cert-registration** so n8n can filter and only run the "grant privileges" path for these tickets.

3. **Approval state**: Define in your workflow what "approved" means (e.g. state = closed + tag "approved", or a custom field "Approval" = "Approved").

## n8n Workflows

- **cert-registration-request** (`n8n-workflows/cert-registration-request.json`): Webhook at `POST /webhook/cert-registration-request` receives registration payload, validates cert_fingerprint or cert_subject_cn, creates Zammad ticket with full context in the first article body, returns 202 with `ticket_id`. No privileges granted.
- **cert-registration-approved** (see below): Triggered when Zammad notifies that a cert-registration ticket was approved. Loads registration payload from ticket body → creates Keycloak user and assigns realm role → posts reply to ticket.

### Approval flow (Zammad webhook → grant privileges)

1. **Zammad** is configured to send a webhook when a ticket is closed (or state set to “approved”). Target URL: `http://n8n:5678/webhook/cert-registration-approved` (or reuse `zammad-ticket` and add a Switch/IF in n8n to route by ticket title/group).
2. **n8n workflow** receives the ticket event. Filter so only tickets that are (a) cert-registration (e.g. title starts with “Approve certificate registration:”) and (b) approved (e.g. state = closed and your workflow marks it approved).
3. **Load payload**: GET the ticket’s first article from Zammad API (`GET /api/v1/tickets/<id>/articles`), parse the body to extract the stored JSON (uid, requested_privilege_level, email, cn, cert_fingerprint).
4. **Grant**: Call Keycloak Admin REST API (get token with admin credentials from Vault, then `POST /admin/realms/master/users`, then `PUT .../users/<id>/reset-password`, then `POST .../users/<id>/role-mappings/realm` with role from `privilege_levels.json`). Or call an internal “grant” service that runs `sync-identities-to-keycloak.ps1` logic for a single identity.
5. **Reply**: POST a new article to the Zammad ticket: “Privileges granted for uid=…; role=…”.

Keycloak credentials (KEYCLOAK_ADMIN, KEYCLOAK_ADMIN_PASSWORD) should be in n8n credentials or read from Vault at runtime. The approval workflow JSON is in `n8n-workflows/cert-registration-approved.json` (skeleton); complete with your Keycloak base URL and credential references.

## Security Notes

- **Validate certificate**: If using mTLS, the registration endpoint must require client cert and use the peer cert for fingerprint/subject. If accepting POST body only, consider a separate channel to verify the cert (e.g. signed assertion from a portal that performed cert verification).
- **Rate limiting**: Apply rate limits on the registration webhook to prevent abuse.
- **Audit**: Log all registration requests and approval outcomes; Zammad tickets provide the audit trail for human decisions.

## Artifacts

| Artifact | Purpose |
|----------|---------|
| This doc | Design and flow for cert self-registration + human approval. |
| `docs/CERT_REGISTRATION_PAYLOAD_SCHEMA.md` | Request/response and stored payload schema. |
| `n8n-workflows/cert-registration-request.json` | n8n workflow: webhook `cert-registration-request` → validate → create Zammad ticket → 202 with ticket_id. |
| `n8n-workflows/cert-registration-approved.json` | n8n workflow: webhook `cert-registration-approved` (from Zammad on ticket close) → filter cert-reg + approved → get ticket articles → parse payload → create Keycloak user → reply to ticket. Complete with Keycloak token and role-mapping steps as needed. |
| Identity schema | `devsecops.identities.schema`: optional `registration_state`, `cert_fingerprint`; pending registrations are not in `identities.yaml` until approved. |
