# OmniGraph infrastructure IR (`omnigraph/ir/v1`)

This document describes the **native, engine-neutral intent model** and how it connects to **multi-format backends** and **enterprise RBAC**. Normative validation: [`schemas/ir.v1.schema.json`](../schemas/ir.v1.schema.json). Architecture decision: [ADR 005](adr/005-native-ir-enterprise-identity.md).

## Why an IR

- **One graph-friendly model** for visualization, inventory correlation, and policy (separate from HCL/YAML syntax).
- **Many outputs**: the same intent can emit OpenTofu, Ansible, Helm, etc., for teams that standardize on different tools.
- **Enterprise gates**: attach RBAC and audit at the IR and API boundary, not inside each external CLI.

## Document shape

| Field | Purpose |
|-------|---------|
| `apiVersion` | Must be `omnigraph/ir/v1` |
| `kind` | Must be `InfrastructureIntent` |
| `metadata.name` | Logical name of the intent bundle |
| `metadata.labels` | Arbitrary string labels (env, region, cost center) |
| `spec.targets[]` | Hosts or endpoints (`id`, optional `ansibleHost`, `labels`) |
| `spec.components[]` | Abstract resources (`id`, `componentType`, `config`) |
| `spec.relations[]` | `from`, `to`, `relationType` (e.g. `depends_on`, `member_of`) |
| `spec.emitHints.backends[]` | Ordered backend ids to request during `omnigraph ir emit` (future) |

`componentType` is an **extensible string** (conventions will grow in docs and linters). Examples: `omnigraph.compute.vm`, `omnigraph.network.vpc`, `omnigraph.app.container`.

## Backend format ids

Stable identifiers used in code (`internal/ir/format.go`) and config:

`opentofu-hcl`, `terraform-hcl`, `pulumi-typescript`, `pulumi-python`, `pulumi-go`, `ansible-playbook`, `ansible-inventory-ini`, `kubernetes-yaml`, `helm-chart`, `packer-hcl`, `docker-compose`, `cloudformation-json`, `cloudformation-yaml`, `puppet-manifest`, `puppet-hiera`

Emitters are registered by id; unknown ids fail validation at emit time.

## Phased delivery

1. **Now:** Schema + Go model + `omnigraph ir validate` + `omnigraph ir formats`; backend interface and registry (stubs).
2. **Next:** First real emitter (e.g. `ansible-inventory-ini` from `spec.targets`) and one declarative emitter (e.g. OpenTofu skeleton).
3. **Later:** Ingest from existing repos (lossy HCL/Ansible import), native apply only where explicitly scoped.

## Identity and RBAC (enterprise)

See [ADR 005](adr/005-native-ir-enterprise-identity.md) and [integrations.md](integrations.md#enterprise-identity-keycloak-freeipa-rbac).

OmniGraph permissions are **orthogonal** to IaC format: the same `Authorizer` protects `serve`, lock APIs, IR emit, and future webhook receivers.
