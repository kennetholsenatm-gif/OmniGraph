# Getting started (workspace only)

This guide follows a **“quiet until it needs to speak”** rhythm: you explore the product in the **browser workspace** first. No terminal or automation steps here—those live in [development/local-dev.md](development/local-dev.md) and [cli-and-ci.md](cli-and-ci.md).

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

## See also

- [Product philosophy](product-philosophy.md) — graph-first positioning; honest boundaries with Terraform/Ansible/OpenTofu.
- [UI modes](guides/ui-modes.md) — Topology, Reconciliation, Posture.
- [UX architecture](core-concepts/ux-architecture.md) — progressive disclosure and backend truth.
