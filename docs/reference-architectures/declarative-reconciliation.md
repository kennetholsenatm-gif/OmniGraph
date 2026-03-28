# Declarative Reconciliation (Reference Architecture)

This document captures a reference pattern for **Kubernetes-style reconciliation** in OmniGraph-oriented environments: **manifest desired state** compared to **provider actual state** in a control loop.

**Not to be confused with** the **Reconciler Engine**, which translates **`omnigraph/ir/v1`** into **emitted artifacts** (Ansible inventory, future formats). That translation layer is documented in [Reconciler Engine](../core-concepts/reconciler-engine.md).

## Model

- Desired state is defined in versioned manifests
- Controllers compare desired vs actual state
- Drift triggers apply/reconcile actions according to policy

## Provider neutrality

Examples may mention specific platforms (for example Incus, LXD, KVM, cloud APIs),
but provider choice is implementation-specific and not required by core OmniGraph.

## Generic manifest example

```yaml
apiVersion: omnigraph.io/v1
kind: InfrastructureManifest
metadata:
  name: example-platform
spec:
  resources:
    - kind: ComputeInstance
      spec:
        provider: "<COMPUTE_PROVIDER>"
        state: running
```

## See also

- [Reconciler Engine](../core-concepts/reconciler-engine.md) — IR → artifacts (different meaning of “reconciliation”)
- [Platform architecture for contributors](../development/platform-architecture.md) — glossary of terms
