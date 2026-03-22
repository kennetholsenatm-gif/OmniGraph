# OmniGraph Wiki

Welcome. OmniGraph is a state-aware DevSecOps orchestration layer: **OpenTofu/Terraform**, **Ansible**, and **telemetry** (NetBox, Zabbix, Prometheus) in one GitOps-oriented flow.

## How this wiki is maintained

The canonical copy of these pages lives in the main repository under [`wiki/`](https://github.com/kennetholsenatm-gif/OmniGraph/tree/main/wiki). To use **GitHub Wiki**:

1. In the GitHub repo, enable **Wiki** (Settings → Features).
2. Either copy Markdown from `wiki/` into new wiki pages, or clone the wiki git remote and add the same files.

Long-form specs, ADRs, and diagrams remain in [`docs/`](../docs/) in the source tree.

## Quick links

- [Control plane CLI](Control-plane-CLI) — `omnigraph` commands, validation, coercion, graph JSON
- [Web client](Web-client) — browser schema validation and Wasm roadmap
- [Lifecycle and handoff](Lifecycle-and-handoff) — PR phases, state intercept, NetBox sync
- [Repository docs](../docs/architecture.md) — architecture overview (main branch)
