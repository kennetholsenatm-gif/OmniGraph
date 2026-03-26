# Pipeline run timeline: `omnigraph/run/v1`

This artifact records an **ordered list of execution steps** for a single CI/CD run so the web UI (and PR comments) can show an **AWX / Semaphore-style timeline** without centralizing execution in OmniGraph.

**Normative schema:** [`schemas/run.v1.schema.json`](../schemas/run.v1.schema.json).

## When to emit

- At the end of a workflow job, or incrementally append if using JSONL (future profile).
- Typical producers: GitHub Actions, Woodpecker, Jenkins, Gitea Actions—any step that already invokes `omnigraph orchestrate` or discrete CLI commands.

## Document shape

| Section | Purpose |
|---------|---------|
| `metadata.runId` | Unique id for this run (CI-generated). |
| `metadata.repository` / `ref` / `sha` | Tie the run to Git state. |
| `metadata.pullRequest` | Optional PR context for GitOps flows. |
| `spec.overallStatus` | Aggregate outcome after all steps complete. |
| `spec.steps[]` | Ordered steps with timing, status, and artifact pointers. |

### Step fields

- **`plugin`**: Logical tool (`opentofu`, `ansible`, `pulumi`, …). Maps to execution-matrix terminology in [execution-matrix.md](execution-matrix.md).
- **`phase`**: UI grouping (`plan`, `apply`, `post_apply`, …).
- **`status`**: Per-step lifecycle.
- **`artifacts`**: Named outputs (plan JSON, graph JSON, `omnigraph/security/v1` scan, etc.); **`ref`** is a URI or path relative to `spec.artifactsRoot`.
- **`logRef`**: Pointer to full logs (object storage, CI artifact, or build URL).

## Mapping from today’s orchestration

A minimal mapping from [`internal/orchestrate`](../internal/orchestrate/orchestrate.go) phases:

| Logical step | Suggested `plugin` | `phase` |
|----------------|-------------------|---------|
| Schema validate | `validate` | `preflight` |
| Coerce | `coerce` | `preflight` |
| tofu/terraform plan | `opentofu` or `terraform` | `plan` |
| ansible-playbook --check | `ansible` | `plan` |
| tofu apply | `opentofu` | `apply` |
| ansible-playbook | `ansible` | `apply` |
| graph emit | `script` or `custom` | `post_apply` |
| netbox sync | `custom` | `post_apply` |

Exact names are not enforced by the schema; **`plugin`** values are extensible via `custom` until enums are widened.

## UI consumption

- Render `steps` in order; color by `status`.
- Link `artifacts` to download viewers (graph, security scan).
- Deep-link `logRef` to CI provider.

## Versioning

- **`apiVersion`:** `omnigraph/run/v1` — bump only on breaking field removals or semantic changes.
- Optional future: `kind: PipelineRunChunk` or JSONL stream for live updates.

## Storage

- Store as **CI artifact** next to `graph.json` and plan files; no requirement for OmniGraph to host logs if `logRef` points to the CI system.
