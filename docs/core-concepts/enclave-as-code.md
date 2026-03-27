# Enclave as Code (Concept)

Enclave as Code is a planning model for expressing isolation boundaries, trust zones,
and policy constraints as declarative infrastructure intent.

This page is conceptual and provider-neutral. Implementation-specific examples and
design experiments are documented under `docs/reference-architectures/`.

## Conceptual Elements

- Isolation boundary definitions
- Workload grouping and relationship constraints
- Policy checks at validation and orchestration time
- Auditability of intent and change history

## Scope Boundary

Core OmniGraph defines the contract and orchestration hooks. Concrete topology,
identity provider wiring, and host/network assumptions belong to reference
architectures, not core concepts.
