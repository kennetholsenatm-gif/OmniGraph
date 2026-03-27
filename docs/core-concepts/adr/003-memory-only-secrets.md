# ADR 003: Memory-Only Secrets

Status: Accepted

Decision: retrieve secrets at runtime and inject only in memory for command
execution; avoid storing secrets in repo artifacts and generated files.

Rationale:

- Lower secret exposure risk
- Better alignment with CI security practices
- Provider-agnostic secret backend integration
