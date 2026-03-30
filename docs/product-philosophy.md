# Product philosophy

OmniGraph exists so teams can **see** infrastructure intent and operational context as an **interactive graph and web workspace**, not only as files, logs, and ticket history.

## What we optimize for

- **Shared understanding:** topology, dependencies, and posture visible in one surface (`omnigraph/graph/v1` and related artifacts).
- **Schema-first intent:** `.omnigraph.schema` and versioned contracts as the spine; the workspace and in-repo libraries consume the same shapes.
- **Honest boundaries:** OpenTofu, Terraform, Ansible, and cloud APIs remain your tools. OmniGraph coordinates visibility and handoff—it does not replace your provider layer.

## What the local control plane is (and is not)

The **local Go workspace server** exposes HTTP APIs and optional static UI for the React workspace: health, repository scan, workspace summary, SSE streams, and optional privileged routes when authenticated. Contributor checks and CI use **`go test`** against the same validation and graph-emit logic the product relies on.

OmniGraph is **not** positioned as a generic terminal-first automation product—that misses the differentiated value. The differentiated value is **graph-forward exploration** and a **first-class web workspace**.

## Copy and docs

When you write README text, help strings, or onboarding: **lead with the graph and UI**; place contributor automation and server flags second unless the audience is explicitly maintainers only. The **root [README.md](../README.md)** should hook on the product with a **web-only quickstart**—no shell recipe blocks for subcommands; those belong in [ci-and-contributor-automation.md](ci-and-contributor-automation.md) and contributor docs.
