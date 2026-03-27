# Kustomize stub

Optional pattern: wrap Helm-generated manifests or maintain thin overlays per environment (`prod`, `staging`).

Example structure (create as needed):

```text
├── base
│   └── kustomization.yaml
└── overlays
    ├── prod
    │   └── kustomization.yaml
    └── staging
        └── kustomization.yaml
```

For this migration, **Helm is the source of truth** unless your GitOps policy requires Kustomize.
