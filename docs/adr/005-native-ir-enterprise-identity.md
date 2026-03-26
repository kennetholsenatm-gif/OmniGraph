# ADR 005: Native infrastructure IR, multi-format backends, and enterprise identity

## Status

Accepted (specification and interfaces; emitters and IdP wiring are phased)

## Context

OmniGraph today **orchestrates** mature engines (OpenTofu/Terraform, Ansible, containers) while normalizing graphs, inventory, and security posture. Product direction calls for a **unified control plane** that can still interoperate with **all major IaC and packaging formats** in enterprise estates.

Enterprises also require **central identity**: **FreeIPA** (LDAP/Kerberos-backed directory), **Keycloak** (OIDC/OAuth2 broker, realm roles, client scopes), and **RBAC** so that APIs, CI hooks, and future web surfaces enforce least privilege—not a single shared bearer token as the long-term model.

Goals:

- A **versioned internal representation (IR)** that is engine-neutral and can be **serialized** to multiple downstream formats (emit path).
- **Optional** continued use of existing runners for **apply** (Git-native, runner-agnostic) until native apply backends are explicitly adopted per stack.
- **Pluggable authentication and authorization**: OIDC (Keycloak), LDAP directory semantics (FreeIPA), and a clear **RBAC** mapping from directory groups / realm roles to OmniGraph permissions.
- No mandate to reimplement Terraform providers or Ansible modules in Go on day one; **parity is incremental** per backend.

## Decision

### 1. Infrastructure IR (`omnigraph/ir/v1`)

Introduce a **Kubernetes-style API document** for high-level intent:

- **`apiVersion`:** `omnigraph/ir/v1`
- **`kind`:** `InfrastructureIntent`
- **`spec.targets`:** logical hosts/endpoints (inventory-oriented labels, connectivity hints).
- **`spec.components`:** abstract building blocks (`componentType` + `config`)—not HCL or YAML for a specific engine inside the IR.
- **`spec.relations`:** directed edges between components (depends-on, exposes-to, member-of).
- **`spec.emitHints`:** optional ordered list of **backend format ids** to generate (see below).

The IR is the **canonical interchange** inside OmniGraph (validation, graph projection, policy). **Backends** translate IR → concrete artifacts (files or API calls). Parsing **from** foreign formats (HCL import, Ansible playbooks) is a **separate** “ingest” track with lossy round-trip called out per format.

Normative JSON Schema: [`schemas/ir.v1.schema.json`](../schemas/ir.v1.schema.json). Narrative and format matrix: [`omnigraph-ir.md`](../omnigraph-ir.md).

### 2. Backend formats (in scope)

Backends are identified by stable string ids. **All are in scope** for the product; **implementation is phased**.

| Backend id | Typical artifacts | Notes |
|------------|-------------------|--------|
| `opentofu-hcl` / `terraform-hcl` | `.tf` | HCL generation; apply via existing runner |
| `pulumi-typescript` / `pulumi-python` / `pulumi-go` | Pulumi program sources | Generation or partial snippets |
| `ansible-playbook` | playbooks, roles, inventory fragments | Aligns with current orchestration |
| `ansible-inventory-ini` | INI | Already aligned with `internal/inventory` |
| `kubernetes-yaml` | manifests, Kustomize-oriented bundles | |
| `helm-chart` | chart tree | |
| `packer-hcl` | Packer templates | |
| `docker-compose` | Compose v3+ | |
| `cloudformation-json` / `cloudformation-yaml` | CFN templates | |
| `puppet-manifest` / `puppet-hiera` | Puppet DSL / data | |

Each backend implements a shared Go **`ir.Backend`** interface: **emit-only** in the first milestones; **plan/apply** remains delegated to external CLIs unless a backend explicitly adds native execution later.

### 3. Enterprise identity and RBAC

**Authentication (AuthN)**

- **Keycloak (OIDC):** Preferred for browser flows, service accounts, and machine tokens. Validate JWTs (issuer, audience, expiry, optional step-up). Map **realm roles** and **client roles** into OmniGraph claims.
- **FreeIPA:** Treat as **LDAP** directory (and optionally **Kerberos** for on-host SSO patterns). Use LDAP bind + search for group membership; do not store directory passwords in OmniGraph. Kerberos integration is **optional** and deployment-specific (SPNEGO gateways, not assumed inside the static binary).

**Authorization (AuthZ)**

- Define a **fixed vocabulary of permissions** (strings), e.g. `serve:inventory:read`, `serve:security:scan`, `serve:host-ops:read`, `serve:host-ops:write`, `ir:emit`, `lock:acquire`, `ci:report`.
- Provide a **`identity.Authorizer`** that answers `Can(subject, permission, resourceScope) bool`.
- **Mapping rules:** configurable mapping from OIDC claims (`groups`, `realm_access.roles`, custom claims) and LDAP `memberOf` / group CN patterns to OmniGraph roles or permission sets.
- **Audit:** privileged actions continue to append to the serve audit ring (and future centralized audit) with **stable subject id** from IdP, not only “bearer token matched.”

**Migration path:** Existing **`OMNIGRAPH_SERVE_TOKEN`** remains supported as a **bootstrap / break-glass** credential when OIDC/LDAP is disabled; documentation must state that it is **not** sufficient for enterprise multi-tenant deployments.

### 4. Threat model (summary)

- IR documents may encode **sensitive intent**; treat emitted artifacts like any IaC repo (secrets via [ADR 003](003-memory-only-secrets.md), not embedded in IR).
- Directory and OIDC clients must use **TLS**, **scoped credentials**, and **minimal LDAP bind DN** privileges.
- RBAC defaults should **deny** unknown permissions; mapping tables are versioned config (e.g. future `omnigraph/rbac/v1` document).

## Consequences

- New packages: `internal/ir` (model + backend registry), `internal/identity` (AuthN/AuthZ interfaces and static RBAC helpers).
- Future work: wire **`serve`** to OIDC/LDAP, replace single-token gate with **`Authorizer`**; implement emitters per backend; optional HCL/Ansible **ingest** to IR.
- Documentation: [omnigraph-ir.md](../omnigraph-ir.md), [integrations.md](../integrations.md) (enterprise identity), [architecture.md](../architecture.md) (ADR table).

## Alternatives considered

- **Pulumi-only as IR:** Rejected as sole path—enterprises still require Ansible, Helm, and Tofu artifacts.
- **Keycloak-only, no LDAP:** Rejected for many on-prem estates where FreeIPA is the source of truth; LDAP remains first-class alongside OIDC.
- **Embed Varlock as the IR core:** Varlock targets **env/spec and secrets** ([varlock.dev](https://varlock.dev/)); OmniGraph IR targets **infrastructure intent**. Varlock may integrate later as a **secrets/config** adapter, not as the IR schema itself.
