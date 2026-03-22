# ADR 003: Memory-only secret injection

## Status

Accepted

## Context

Writing secrets to `.env`, `terraform.tfvars`, or other workspace files risks accidental commits and leaves sensitive data on disk if a job crashes.

## Decision

Secrets are **never** persisted to those files on disk. The control plane resolves secrets from Vault, cloud secret managers, or SOPS, holds them in process memory, and injects them into the runner environment for the duration of the step only. Logs are scrubbed for known secret values.

## Consequences

- **Positive:** Smaller leak surface; better alignment with compliance expectations.
- **Negative:** Runners must support env-only injection; debugging may require redacted reproduction steps.
- **Mitigation:** Document allowed patterns in CONTRIBUTING; optional local-only dev workflows with fake secrets (never real credentials in repo).
