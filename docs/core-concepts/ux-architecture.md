# UX architecture: disclosure, truth, and context

OmniGraph’s web workspace is organized around three ideas that directly answer familiar infrastructure pain: **cognitive overload** (too much on screen at once), **state drift** (the UI says one thing and the world says another), and **abstraction leaks** (declarative models and imperative logs never meet). Together they define how the product feels to use—not only which buttons exist.

For how those ideas map to sidebar tools today, read [Understanding the UI modes](../guides/ui-modes.md). For layer diagrams and repository layout, see [Architecture](architecture.md).

## Progressive disclosure (curing cognitive overload)

Traditional infra UIs punish curiosity: every panel stays open, every list competes for attention, and “power” is mistaken for density. OmniGraph reverses that. **Progressive disclosure** means the workspace reveals depth **in rhythm with your focus**—especially through **node selection** on the graph. When you select a resource, the Inspector and adjacent panels **surface structurally relevant variables, metadata, and log fragments** tied to that node. Overarching inventories and secondary metrics **withdraw** so the right pane stays about the object you are reasoning about, not the entire estate.

Operationally, the product organizes work into **three modes**—Topology, Reconciliation, and Posture—so you are not asked to read topology while reconciling drift, or reconcile drift while judging exposure. Each mode is a **cognitive buffer**: it holds the problem shape steady while the UI loads just enough detail to act.

## Absolute truth (curing state drift)

Split-brain UIs teach bad habits: they flash success before the network answers, then scramble when reality disagrees. OmniGraph **does not use optimistic UI updates** for authoritative outcomes. The **Go control plane** owns discovery, execution handoff, and aggregation; the **React/TypeScript client** is a **reactive projection** of that truth.

The contract is: **material visual state for live control-plane-backed views is delivered through a unidirectional stream of Server-Sent Events (SSE)** (`GET /api/v1/workspace/stream` emits `workspace_summary` JSON on `omnigraph serve`). The client applies those events in order and does not fabricate summary success. If the stream stalls, the UI shows **disconnection or staleness**, not fiction. That separation is what keeps the workspace view from drifting away from the repo or the last indexed state the server actually read.

## Contextual debugging (curing abstraction leaks)

Declarative graphs explain **intent**; Ansible-style logs explain **what the runner did in the ugly middle**. Most tools leave you to glue those worlds together by hand. OmniGraph **maps imperative evidence onto declarative structure**. When a task fails, the relevant log slice is **indexed to the graph node** it affected—not dumped into a generic console where you grep for hostnames and hope.

You still read the real log lines, but you read them **beside the node that failed**, with structure that tells you **why that node mattered** to the declared system. Abstraction carries its own receipts.

## WASM boundary (Web IDE)

The **shipping Topology experience** builds `omnigraph/graph/v1` in **Go** ([`internal/graph`](../../internal/graph), CLI in [`internal/cli/graph.go`](../../internal/cli/graph.go)) from project documents, optional plan/state paths, and merges—not from Wasm inside the CLI.

**What the bundled Wasm does:** the **Web IDE** loads [`wasm/hcldiag`](../../wasm/hcldiag) in the **browser** only. [`packages/web/src/hclWasm.ts`](../../packages/web/src/hclWasm.ts) instantiates it and exposes `omnigraphHclValidate` on `globalThis` so the UI can lint **HCL source strings** you type in the scratchpad. It does **not** read `terraform.tfstate`, plan JSON, or `.omnigraph.schema` from disk; those paths are handled by **Go** (`graph emit --tfstate`, `serve` discovery, Inventory paste) or by you pasting text into the workspace.

**What Wasm does not do:** it is not the authority for graph reconciliation, inventory, or OpenTofu/Terraform state. Treat Wasm as an **editor assist** for HCL-shaped text, separate from the control plane.

Other trees (for example [`internal/enclave`](../../internal/enclave)) describe different Wasm-adjacent designs for contributors; they do not power the default Topology tab today.

**Browser integration (excerpt from `hclWasm.ts`):**

```typescript
const go = new Go()
const res = await WebAssembly.instantiateStreaming(fetch('/wasm/hcldiag.wasm'), go.importObject)
void go.run(res.instance)
// …then validateHclText(src) calls globalThis.omnigraphHclValidate(src) and JSON.parses diagnostics
```

## See also

- [Understanding the UI modes](../guides/ui-modes.md)
- [Using the web workspace](../using-the-web.md)
- [Product philosophy](../product-philosophy.md)
- [Architecture](architecture.md)
