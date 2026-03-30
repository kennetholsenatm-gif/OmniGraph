# Getting started (workspace only)

This guide follows a **“quiet until it needs to speak”** rhythm: you explore the product in the **browser workspace** first. No terminal or automation steps here—those live in [development/local-dev.md](development/local-dev.md) and [ci-and-contributor-automation.md](ci-and-contributor-automation.md).

## 1. Open the workspace

Use the URL your environment provides (for example a local dev server started by a teammate, or your own setup per [local-dev.md](development/local-dev.md)). You should land in **Topology**.

## 2. Meet the sample graph

The app opens with a **sample graph**: nodes and edges already on the canvas. **Pan and zoom** to orient yourself. **Click a node** to open the **Inspector** on the side: you will see **id**, **kind**, **label**, and optional fields such as **state** or **attributes**. This is the core loop—**the graph is the primary surface**, not a report you scroll past.

## 3. Follow one relationship

Pick an **edge** if the sample includes one, or switch between two connected nodes. Notice how the Inspector **grounds** what you are looking at: one vertex at a time, with context kept on the canvas.

## 4. Peek at other modes—without leaving the story

When you are ready, use the sidebar:

- **Schema Contract** — where **schema-first intent** for the project lives; validation feedback appears when you edit.
- **Inventory** — where **state-shaped** and **inventory-shaped** inputs attach to the same narrative as the graph (including optional **local file** access when your deployment enables it).
- **Posture** — where **security-shaped** JSON can sit beside topology for a single mental model.

You do not need to paste real infrastructure data on day one. The sample graph is enough to learn **how OmniGraph wants you to think**: **shared understanding** through an **interactive topology**, not a dump of logs or files.

## 5. When you add real data

Bring **Terraform/OpenTofu JSON state**, **plan JSON**, or **Ansible INI** when you are ready—through the **Inventory** flows your team enables. Same-origin **live summaries** are described in [using-the-web.md](using-the-web.md).

## 6. Five-minute observation drill (Topology + live sync)

OmniGraph is a **visibility and coordination** surface for infrastructure intent and state—not an execution engine for outages. This drill rehearses **watching the graph react** when **you** change the world outside the UI.

1. Open **Topology** with the **sample** or a real **`omnigraph/graph/v1`** document, and enable **same-origin** ingest if your deployment supports it (background sync / SSE summaries are described in [using-the-web.md](using-the-web.md)).
2. In a **terminal** (or automation you control), trigger a **safe, disposable** change: for example run a small shell script such as [`examples/quickstart/break_network.sh`](../examples/quickstart/break_network.sh) after editing it for your lab, **or** apply a mock **Terraform/OpenTofu drift** in a throwaway workspace, **or** stop a local dependency your graph represents.
3. Stay in the workspace and watch **Topology**, **Inventory**, and any **live summary** stream: node state, attributes, and graph updates should move as refreshed inputs arrive—not because the web app “simulated” the failure for you.
4. Turn on **Triage mode** when you want the **node-scoped** panel: practice narrowing from the **selected** vertex while the canvas keeps context.

**Why this matters:** Edges can declare **`dependencyRole`** (`necessary` vs `sufficient`), which shapes **blast radius** and triage semantics. You practice that mapping against **real or lab telemetry**, not an in-browser fake outage. Full reference: [Graph dependencies and blast radius](guides/graph-dependencies-and-blast-radius.md). Fixture-oriented steps: [examples/quickstart/README.md](../examples/quickstart/README.md).

For **how OpenTofu/Terraform/Ansible artifacts become Inventory and Topology updates** (without duplicating this guide), see [Data handoff](core-concepts/data-handoff.md).

## See also

- [Product philosophy](product-philosophy.md) — graph-first positioning; honest boundaries with Terraform/Ansible/OpenTofu.
- [UI modes](guides/ui-modes.md) — Topology, Reconciliation, Posture.
- [NOC / SRE workflow](guides/workflows-noc-sre.md) · [SOC / SecOps workflow](guides/workflows-soc-secops.md)
- [UX architecture](core-concepts/ux-architecture.md) — progressive disclosure and backend truth.
