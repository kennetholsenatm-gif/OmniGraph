# ADR 004: Unified State Locking

Status: Accepted

Decision: define a consistent lock model for orchestration phases to prevent unsafe
concurrent apply operations.

Rationale:

- Predictable behavior in concurrent CI pipelines
- Reduced race conditions during state mutation
- Better auditability of execution ownership
