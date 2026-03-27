# A2A Task Request Payload Schema

n8n publishes to Solace topic `a2a/v1/agent/request/mesh_orchestrator` with a structured payload. The SAM Orchestrator and specialized agents expect the following shape so they can decompose and delegate.

## Mesh Orchestrator Request (from n8n)

Published to: `a2a/v1/agent/request/mesh_orchestrator`

```json
{
  "intent": "string (macro objective from ITIL ticket)",
  "requirements": ["string"],
  "compliance_needs": ["CMMC_L2", "PQC_READY"],
  "task_breakdown": [
    { "agent": "cline_coder_agent", "description": "string", "context": {} },
    { "agent": "security_critique_agent", "description": "string", "context": {} },
    { "agent": "doc_agent", "description": "string", "context": {} }
  ],
  "itil_ticket_id": "string",
  "change_request_approval_status": "approved"
}
```

- **intent**: High-level goal from the parsed Zammad ticket.
- **requirements**: List of concrete requirements.
- **compliance_needs**: Compliance tags (CMMC 2.0 Level 2, PQC).
- **task_breakdown**: Array of sub-tasks; each `agent` is the Solace topic suffix (`a2a/v1/agent/request/<agent>`).
- **itil_ticket_id**: Zammad ticket ID for traceability and JIT approval linkage.
- **change_request_approval_status**: e.g. `approved` so the orchestrator only delegates when the ticket is approved.

## Specialized Agent Request (from SAM)

SAM publishes to `a2a/v1/agent/request/cline_coder_agent`, `a2a/v1/agent/request/security_critique_agent`, `a2a/v1/agent/request/doc_agent` with a per-agent payload, e.g.:

```json
{
  "task_id": "uuid",
  "parent_ticket_id": "string",
  "description": "string",
  "context": { "repo": "url", "branch": "string", "paths": [] },
  "capabilities_used": ["code_generation", "git_operations"]
}
```

Agents publish responses to `a2a/v1/agent/response/<agent_id>` and status to `a2a/v1/agent/status/<agent_id>` so the mesh orchestrator can aggregate and report back (e.g. to n8n or Zammad).
