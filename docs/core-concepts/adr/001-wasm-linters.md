# ADR 001: WebAssembly Linters in UI

Status: Accepted

Decision: run selected lightweight lint/diagnostic logic in-browser via WebAssembly
to provide fast local feedback before CI execution.

Rationale:

- Early feedback for developers
- Reduced server dependency for basic diagnostics
- Reuse of deterministic parser/lint logic
