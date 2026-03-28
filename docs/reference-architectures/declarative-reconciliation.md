# Declarative Reconciliation (Reference Architecture)

This document captures a reference pattern for **Kubernetes-style reconciliation** in OmniGraph-oriented environments: **manifest desired state** compared to **provider actual state** in a control loop.

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

- [Emitter Engine](../core-concepts/emitter-engine.md) — compiles `omnigraph/ir/v1` into execution artifacts (inventory, etc.); distinct from manifest reconciliation.
- [Platform architecture for contributors](../development/platform-architecture.md) — glossary of terms
