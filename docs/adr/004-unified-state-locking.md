# ADR 004: Unified state locking across IaC engines

## Status

Proposed

## Context

OmniGraph orchestrates multiple declarative engines (OpenTofu/Terraform, Pulumi, and future Packer stacks). In PR-driven GitOps, two changes can race: the same remote state or overlapping stacks may be planned or applied concurrently from different branches or pipelines. Without a shared lock contract, operators risk corrupted state, partial applies, and conflicting automation.

Goals aligned with the **Git-native orchestrator** pillar:

- Keep **execution on existing CI runners**; locking must work from any runner with access to a chosen backend.
- Support **multiple engines** tagging the same logical “stack” or distinct stack keys in one repository.
- Remain **optional**: repositories that use a single engine and external locking (e.g. Terraform Cloud) may not adopt OmniGraph locks.

## Decision

Introduce a versioned lease contract **`omnigraph/lock/v1`** (JSON). A lock is a short-lived **lease** held by a single **holder** identity (CI job id, PR number + run id, or human operator id). Locks are keyed by a **stack key** chosen by the project (e.g. `s3://bucket/env/prod.tfstate`, `pulumi:org/proj/stack`, or a repo-relative `stackId` in `.omnigraph.yaml`).

### Lock document shape (normative)

| Field | Required | Description |
|-------|----------|-------------|
| `apiVersion` | yes | Constant `omnigraph/lock/v1` |
| `kind` | yes | Constant `StateLock` |
| `metadata.stackKey` | yes | Opaque UTF-8 string; globally unique per estate (recommended: URI-style prefix by engine) |
| `metadata.engine` | yes | `opentofu` \| `terraform` \| `pulumi` \| `packer` \| `other` |
| `spec.holder` | yes | Opaque string (e.g. `gh:repo:123:run-456`, `jenkins:job:789`) |
| `spec.leaseUntil` | yes | RFC3339 UTC; lock is invalid after this instant |
| `spec.acquiredAt` | yes | RFC3339 UTC |
| `spec.reason` | no | Human-readable (e.g. PR URL) |
| `spec.extensions` | no | Engine-specific metadata (OpenTofu workspace name, Pulumi stack fqdn) |

**Acquire semantics (intended implementation):**

1. Client reads current lock for `stackKey` (if any).
2. If no lock or `leaseUntil` is in the past, client may **compare-and-set** a new lease with a monotonically newer `acquiredAt` (or backend-native CAS).
3. If an active lock exists with a different `holder`, acquire fails; client must wait, escalate, or break-glass per policy.

**Release:** Delete lock document or set `leaseUntil` to now (backend-specific). Heartbeat may extend `leaseUntil` for long applies.

**Break-glass:** Documented out-of-band procedure (e.g. admin deletes lock object); audited.

### Backend options

| Backend | Mechanism | Pros | Cons |
|---------|-----------|------|------|
| **S3** (or compatible) | Object `PUT` with `If-None-Match` / conditional on `ETag` | Simple, cheap, ubiquitous | No native TTL; rely on short leases + sweeper |
| **DynamoDB** | Item `stackKey` + conditional `attribute_not_exists` or version | TTL support, CAS | AWS-centric |
| **PostgreSQL** | Row per `stackKey` + `FOR UPDATE` or optimistic version column | Strong consistency, SQL ops | Must operate DB |
| **Redis** | `SET key NX PX` | Fast TTL | Volatility; persistence policy matters |
| **HTTP lock service** | Dedicated microservice implementing the same CAS semantics | Fits air-gapped adapters | Another component to run |

The reference CLI surface (future work) is expected to be:

- `omnigraph lock acquire --stack-key … --holder … --ttl 30m`
- `omnigraph lock release …`
- `omnigraph lock status …`

Integration points: invoked **before** `tofu plan/apply` and Pulumi `preview/up` in CI, wired via [`internal/orchestrate`](../../internal/orchestrate/orchestrate.go) once implemented.

## Consequences

- **Positive:** Cross-engine coordination in multi-stack repos; clearer PR automation story; aligns with Digger-style “brain not brawn” (lock is metadata, not a build executor).
- **Negative:** Operators must choose and operate a lock backend; misconfigured TTLs can stall pipelines or allow overlap if clocks skew.
- **Mitigation:** Default short leases (e.g. 15–30m), heartbeat for long applies, document clock sync (NTP), metrics on lock contention.

## References

- Strategic roadmap: Pillar 1 (Git-native orchestrator).
- JSON Schema (optional validation): [`schemas/lock.v1.schema.json`](../../schemas/lock.v1.schema.json) (shipped alongside this ADR for tooling).
