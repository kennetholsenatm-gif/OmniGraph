# Teleport JIT Access Integration

## Overview

Just-in-Time (JIT) access issues short-lived certificates for agents (e.g. Cline Coder) based on approved ITIL tickets. The flow is: **Zammad webhook → n8n → Teleport API → ephemeral cert**.

## Options

### Option 1: Teleport Access Request API

1. **n8n** receives a Zammad webhook when a ticket is approved for JIT.
2. n8n calls Teleport's **Access Request** API to create a request (e.g. for role `cline-coder`).
3. A Teleport plugin or automation approves the request when the ticket ID is valid (e.g. mapped in Teleport plugin config or via `tctl request approve`).
4. The Coder Agent (or n8n) retrieves the approved cert via `tsh login` or Teleport's certificate issuance API.

**Teleport docs**: [Access Requests](https://goteleport.com/docs/access-controls/access-requests/), [Machine ID (plugin)](https://goteleport.com/docs/machine-id/).

### Option 2: Custom JIT Sidecar

A small service (e.g. Python/Go) that:

1. Exposes `POST /api/v1/jit/request` (or the URL you set in `TELEPORT_JIT_REQUEST_URL` in devsecops.env.schema).
2. Accepts a body with `ticket_id`, `role`, `requester`.
3. Validates the ticket via Zammad API (or trusts n8n to have validated it).
4. Uses Teleport Auth API or `tctl` to issue a short-lived cert for the requested role and returns it (or writes to a mounted volume for the Coder Agent).

**n8n**: Configure the "Request JIT Access" HTTP node to call this URL with `Authorization: token {{ $credentials.teleport_api_token }}` and body `{ "ticket_id": "{{ $json.itil_ticket_id }}", "role": "cline-coder" }`.

## Configuration

- **devsecops.env.schema**: Set `TELEPORT_JIT_REQUEST_URL` to your Teleport proxy + custom path (Option 2) or Teleport Auth API endpoint (Option 1).
- **Credentials**: Store `teleport_api_token` (or Teleport Machine ID token) in n8n credentials; reference via `{{ $credentials.teleport_api_token }}`. Do not put the token in workflow JSON.
- **Cline/Coder Agent**: After JIT approval, the agent runtime must receive the cert (e.g. via volume mount or env with cert path). Run Cline with Teleport cert path so it can access repos and filesystems under the requested role.
