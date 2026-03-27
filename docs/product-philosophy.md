# Product philosophy

OmniGraph exists so teams can **see** infrastructure intent and operational context as an **interactive graph and web workspace**, not only as files, logs, and ticket history.

## What we optimize for

- **Shared understanding:** topology, dependencies, and posture visible in one surface (`omnigraph/graph/v1` and related artifacts).
- **Schema-first intent:** `.omnigraph.schema` and versioned contracts as the spine; the UI and CLI both consume the same shapes.
- **Honest boundaries:** OpenTofu, Terraform, Ansible, and cloud APIs remain your tools. OmniGraph coordinates visibility and handoff—it does not replace your provider layer.

## What the CLI is (and is not)

The **`omnigraph`** binary is the **control plane and automation surface**: validate and policy-check documents, emit graph JSON for CI, run orchestrated pipelines, scan posture, and serve HTTP APIs. It is essential for **headless workflows** and **integration**.

It is **not** how we want strangers to categorize the project. OmniGraph is **not** positioned as “one more generic CI/CD CLI”—that market is crowded and misses the point. The differentiated value is **graph-forward exploration** and a **first-class web workspace**.

## Copy and docs

When you write README text, help strings, or onboarding: **lead with the graph and UI**; place CLI and pipeline detail second unless the audience is explicitly automation-only. The **root [README.md](../README.md)** should hook on the product with a **web-only quickstart**—no `omnigraph` command blocks or validation walkthroughs there; those belong in [cli-and-ci.md](cli-and-ci.md) and contributor docs.
