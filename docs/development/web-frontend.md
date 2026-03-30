# Web Frontend Development

The OmniGraph web application is implemented in React + TypeScript and built with
Vite. It is the **primary user-facing surface** for exploring graphs and workspace state.

**Package root:** [`packages/web`](../../packages/web).

## What ships in the MVP UI

Implemented in [`packages/web/src/mvp/OmniGraphMVP.tsx`](../../packages/web/src/mvp/OmniGraphMVP.tsx) with these sidebar tabs:

| Tab | Role |
|-----|------|
| **Topology** | `omnigraph/graph/v1` text editor + interactive canvas; Inspector + `attributes.debugLog` ([`GraphVisualizerTab`](../../packages/web/src/mvp/GraphVisualizerTab.tsx)). |
| **Schema Contract** | `.omnigraph.schema` editing and validation ([`SchemaTab`](../../packages/web/src/mvp/SchemaTab.tsx)). |
| **Web IDE** | HCL scratch + WASM diagnostics when available ([`WebIDETab`](../../packages/web/src/mvp/WebIDETab.tsx)). |
| **Inventory** | State/plan/INI paste, folder scan, `serve` summary, SSE `GET /api/v1/workspace/stream` ([`InventoryTab`](../../packages/web/src/mvp/InventoryTab.tsx), [`useWorkspaceSummaryStream`](../../packages/web/src/mvp/useWorkspaceSummaryStream.ts)). |
| **Pipeline** | OpenTofu/Ansible execution-matrix context ([`PipelineTab`](../../packages/web/src/mvp/PipelineTab.tsx)). |
| **Posture** | `omnigraph/security/v1` JSON ([`PostureTab`](../../packages/web/src/mvp/PostureTab.tsx)). |

Workspace state persists as **v1** in `localStorage` ([`workspaceStorage.ts`](../../packages/web/src/mvp/workspaceStorage.ts)). End-user-oriented walkthrough: [using-the-web.md](../using-the-web.md). UX narrative (Topology / Reconciliation / Posture): [Understanding the UI modes](../guides/ui-modes.md), [UX architecture](../core-concepts/ux-architecture.md).

## Start development server

```bash
cd packages/web
npm ci
npm run dev
```

## Validate frontend changes

```bash
cd packages/web
npm run lint
npm run build
```

## Optional Wasm flow

If you are changing WebAssembly-backed diagnostics, rebuild Wasm assets first, then
run the web app for integration testing.

```bash
make wasm-hcldiag
cd packages/web && npm run dev
```
