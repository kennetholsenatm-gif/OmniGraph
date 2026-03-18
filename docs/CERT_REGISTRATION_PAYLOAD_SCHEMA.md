# Certificate Registration Request and Stored Payload Schema

Used by the [certificate-based self-registration](CERT_SELF_REGISTRATION.md) flow. No privileges are granted until a human approves the linked Zammad ticket.

## Request (POST to registration webhook)

**Content-Type:** `application/json`

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cert_fingerprint` | string | Yes* | SHA-256 (or SHA-1) fingerprint of client certificate (hex or base64). *Required if cert_subject_cn not provided. |
| `cert_subject_dn` | string | No | Full subject DN (e.g. `CN=john,OU=dev,O=acme`). |
| `cert_subject_cn` | string | Yes* | Common Name from certificate; used as default uid. *Required if cert_fingerprint not provided. |
| `uid` | string | No | Desired username; defaults to sanitized cert_subject_cn. |
| `requested_privilege_level` | string | No | One of `admin`, `operator`, `developer`, `viewer`, `auditor`. Default `viewer`. |
| `email` | string | No | Email for the identity. |
| `cn` | string | No | Display name; defaults to cert_subject_cn or uid. |

**Example (body):**

```json
{
  "cert_fingerprint": "a1b2c3d4e5f6...",
  "cert_subject_cn": "john.doe",
  "cert_subject_dn": "CN=john.doe,OU=developers,O=Example",
  "requested_privilege_level": "developer",
  "email": "john.doe@example.com",
  "cn": "John Doe"
}
```

## Response (202 Accepted)

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | `"pending"` |
| `message` | string | Human-readable message: no privileges until ticket approved. |
| `ticket_id` | string | Zammad ticket ID. |
| `ticket_url` | string | Optional; link to the ticket in Zammad. |

**Example:**

```json
{
  "status": "pending",
  "message": "No privileges granted until a human approves the ticket.",
  "ticket_id": "42",
  "ticket_url": "http://zammad/ticket/42"
}
```

## Stored payload (pending registration)

Stored in Vault at `secret/pending_registrations/<ticket_id>` or in Zammad ticket body/article for the approval workflow to read.

| Field | Type | Description |
|-------|------|-------------|
| `ticket_id` | string | Zammad ticket ID. |
| `cert_fingerprint` | string | As in request. |
| `cert_subject_dn` | string | As in request. |
| `cert_subject_cn` | string | As in request. |
| `uid` | string | Resolved username. |
| `requested_privilege_level` | string | As in request. |
| `email` | string | As in request. |
| `cn` | string | As in request. |
| `created_at` | string | ISO 8601 timestamp. |
| `registration_state` | string | `pending` \| `approved` \| `rejected`. |

When the ticket is **approved**, the grant workflow uses this payload to create the Keycloak user and assign the realm role for `requested_privilege_level`; then it may set `registration_state` to `approved` or remove the pending record.
