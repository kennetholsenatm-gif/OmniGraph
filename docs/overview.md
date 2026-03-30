# Overview

This page orients you in one pass: **who** typically uses OmniGraph, **what** it does (and does not do), **where** the important pieces live in the repository, and how major artifacts relate.

## Why OmniGraph

Infrastructure work rarely fails because people lack tools—it fails because **too much truth arrives at once**. Dashboards become cockpits covered in dials. Declarative intent lives in one pane, failure signals in another, and neither pane agrees with what actually ran. OmniGraph exists to **compress that chaos into a single declarative graph** you can explore: **OpenTofu, Terraform, and Ansible** (and the contracts around them) stay your execution layer; OmniGraph **coordinates visibility and handoff**—it does **not** replace providers, modules, or playbooks. The **browser-based local workspace** visualizes schema-first intent with an interface disciplined enough to respect human attention.

The workspace is deliberately **quiet until it needs to speak**. Instead of mirroring every metric your platform knows, OmniGraph asks: *what matters for the decision in front of you right now?* It separates **topology** (what exists and how it connects), **reconciliation** (what the world claims versus what you declared), and **posture** (how safe and compliant that shape is). You move between those contexts the way you move between mental models—without carrying the entire cockpit into each step. See [Understanding the UI modes](guides/ui-modes.md) for how this maps to sidebar tools.

Behind that calm surface sits a stricter rule: **the browser does not get to invent reality**. A **Go control plane** performs discovery, validation, and aggregation; the **TypeScript frontend** renders what the backend has already proven. Live updates for authoritative state are designed to flow through a **unidirectional stream of Server-Sent Events (SSE)**, so the canvas does not “guess” that a connection succeeded or a run finished. When the UI changes, it is because **authoritative state changed**—not because a spinner raced ahead of the truth. Read [UX architecture](core-concepts/ux-architecture.md) for the full model.

OmniGraph also refuses the false choice between **elegant declarative graphs** and **opaque run output**. When runner output is attached to the model, it can be **anchored to the graph node it belongs to** and surfaced in the **Inspector**—so you explore **where the intent delta matters**, not only raw text streams.

## Who this is for

- **Operators, reviewers, and platform engineers** who want **graph-level visibility** into intent, topology, pipeline context, and posture without living in raw logs.
- **Teams standardizing on a shared workspace** (React UI + optional local **workspace server**) for exploration before or alongside automation.
- **Automation owners** who wire **CI** (`go test`) and **OpenTofu/Ansible** directly—the same graph contracts the UI displays are exercised in-repo.
- **Contributors** extending the web app, schemas, or Go control plane.

OmniGraph is not a replacement for Terraform, OpenTofu, Ansible, or your cloud APIs. It sits **above** those tools: contracts, visibility, and emitted artifacts for the workspace. Read [product-philosophy.md](product-philosophy.md) for positioning.

## What OmniGraph does

- **Interactive web workspace** ([`packages/web`](../packages/web)): Topology (graph JSON + Inspector), schema validation, pipeline context, inventory with optional workspace summary and SSE stream, posture JSON, optional WASM HCL IDE—see [using-the-web.md](using-the-web.md).
- **Versioned graph artifacts** (`omnigraph/graph/v1`) merged with optional `omnigraph/telemetry/v1` and `omnigraph/security/v1` for what you **see** in the UI and in CI consumers.
- **HTTP API** (local workspace server) for repository/workspace discovery and serving the built UI with `--web-dist`.
- **Schema-first project documents** (`.omnigraph.schema` and related JSON Schema) validated in the UI and in **`go test`**.
- **Policy-as-code** (Rego in policy sets) during contributor validation gates—results inform CI and can align with workspace context.
- **Orchestration libraries** in Go ([`internal/orchestrate`](../internal/orchestrate)) for chained OpenTofu/Ansible flows when you integrate them yourself; the product surface is the **web workspace**, not a terminal multi-command tool. See [execution-matrix.md](core-concepts/execution-matrix.md).

Experimental or stub areas (for example some IaC engine paths) are noted in [execution-matrix.md](core-concepts/execution-matrix.md) and [CI and contributor automation](ci-and-contributor-automation.md).

## System context

The diagram below is logical: the **browser** is the primary human entry for exploration; **CI** and the **workspace server** are parallel automation paths. All consume or produce the same contracts and artifacts.

```mermaid
flowchart TB
  subgraph people [People_and_automation]
    Browser
    Operator
    CIJob[CI_job]
  end
  subgraph entry [OmniGraph]
    UI[Web_workspace_and_static]
    SRV[workspace_server_HTTP]
  end
  subgraph external [Your_IaC_runtime]
    Tools[OpenTofu_Terraform_Ansible_others]
  end
  Browser --> UI
  Operator --> SRV
  CIJob --> Tools
  UI --> SRV
  SRV --> Tools
```

## Artifact relationships

A common path: validate a project document, emit a graph for the **Topology** view or pipelines, and enrich it with telemetry and security scans produced separately.

```mermaid
flowchart LR
  Schema[dot_omnigraph_schema]
  GraphOut[omnigraph_graph_v1]
  Tel[telemetry_v1_optional]
  SecDoc[security_v1_optional]
  Schema --> GraphOut
  Tel -->|"graph_emit_merge"| GraphOut
  SecDoc -->|"graph_emit_merge"| GraphOut
```

- **IR YAML** (`omnigraph/ir/v1`) describes infrastructure intent for validation and emission workflows; see [omnigraph-ir.md](core-concepts/omnigraph-ir.md). Example: [`testdata/sample.ir.v1.yaml`](../testdata/sample.ir.v1.yaml).
- Example telemetry and security JSON under [`testdata/`](../testdata/) mirror the shapes merged during graph emit.

## Where things live in the repo

| Path | Role |
|------|------|
| [`packages/web`](../packages/web) | React workspace (graph, schema, pipeline, inventory, posture). |
| [`wasm/`](../wasm/) | WASM used by the UI. |
| [`cmd/`](../cmd/), [`internal/`](../internal/) | Go control plane: graph emit, HTTP **`serve`** (SSE, optional ingest/sync), orchestration libraries, policy, security. |
| [`schemas/`](../schemas/) | Versioned JSON Schema and contract sources. |
| [`docs/`](../docs/) | Canonical documentation (this tree). |
| [`testdata/`](../testdata/) | Fixtures for validation, policies, sample graph/telemetry/security. |

## Related reading

- [Getting started (workspace only)](getting-started.md)
- [UX architecture](core-concepts/ux-architecture.md)
- [Understanding the UI modes](guides/ui-modes.md)
- [Using the web workspace](using-the-web.md)
- [CI and contributor automation](ci-and-contributor-automation.md)
- [Architecture (layers)](core-concepts/architecture.md)
- [Execution matrix](core-concepts/execution-matrix.md)
- [Security posture](security/posture.md)
- [Documentation hub](README.md)
