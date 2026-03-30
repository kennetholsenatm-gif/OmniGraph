# OmniGraph Wiki

This folder serves two audiences:

1. **Browsing the repo** — Use the relative links below from the `main` tree.
2. **GitHub Wiki tab** — Wikis use a separate repo (`OmniGraph.wiki.git`). To publish these pages, follow **[SYNC](SYNC.md)**.

**Visualization-first:** start with the web workspace and philosophy, then automation.

OmniGraph is a web-first workspace that makes infrastructure intent visible as a graph, then connects that model to OpenTofu/Terraform + Ansible workflows. The goal is to replace fragmented CI/CD and terminal context with one place to see topology, desired state, handoffs, and posture.

What to understand first:

- **What makes OmniGraph different** - graph-first model + shared web workspace instead of scattered logs and scripts.
- **How Ansible becomes more declarative** - desired state and graph context drive convergence decisions.
- **How CI/CD pain is reduced** - fewer brittle handoffs, less pipeline opacity, faster triage with unified context.

- [Using the web workspace](../docs/using-the-web.md)
- [Product philosophy](../docs/product-philosophy.md)
- **[README](../README.md)** — web-first landing, differentiation, declarative Ansible model, and quickstart
- **[Documentation index](../docs/README.md)** — full reading order
- **[Overview](../docs/overview.md)** — who / what / where
- [CI and contributor automation](../docs/ci-and-contributor-automation.md) — `go test`, workspace server, fixtures
- [Security posture](../docs/security/posture.md)

## Topic stubs

- [Getting Started](Getting-Started.md)
- [Core Concepts](Core-Concepts.md)
- [Development](Development.md)
- [Schemas](Schemas.md)
- [Reference Architectures](Reference-Architectures.md)
- [Contributing](Contributing.md)
- **[Publish to GitHub Wiki](SYNC.md)**

Absolute **`https://github.com/.../blob/main/docs/...`** links work if you copy this file into the `.wiki` repository where `../docs/` is absent.
