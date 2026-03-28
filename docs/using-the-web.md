# Using the web workspace

The React app under [`packages/web`](../packages/web) is the **primary human interface** for OmniGraph today: a single-page **workspace** with a sidebar of tools around the same persisted state (localStorage as **workspace v1**).

## Run locally

Requires **Node.js 20+**.

```bash
cd packages/web
npm ci
npm run dev
```

Open the dev URL Vite prints. No Go binary is required to explore the UI with bundled samples.

## Sidebar: what each tab does

These labels match [`packages/web/src/mvp/OmniGraphMVP.tsx`](../packages/web/src/mvp/OmniGraphMVP.tsx).

| Tab | Purpose |
|-----|---------|
| **Visualizer** | Edit or paste **`omnigraph/graph/v1`** JSON; interactive graph (React Flow). Optional filename hint for export discipline. |
| **Schema Contract** | Edit **`.omnigraph.schema`** YAML/JSON; validate against the bundled schema; configure path to the **`omnigraph`** binary for CLI-backed validation when needed. |
| **Web IDE** | HCL scratchpad with **WASM-backed diagnostics** when `hcldiag.wasm` is available (`HCL Wasm ready` in the footer). |
| **Inventory** | Paste **Terraform/OpenTofu JSON state**, **plan JSON**, **Ansible INI**; optionally **scan a repository folder** for `.omnigraph.schema` files; or call **`omnigraph serve`** **workspace summary** when the API is reachable (set **Git repository root** in the header first). |
| **GitOps Pipeline** | Form for **`omnigraph orchestrate`** fields (workdir, playbook, runner, images, graph output path, etc.); generates a copy-paste shell command—useful for seeing how CLI flags map to your repo layout. |
| **Posture** | Edit **`omnigraph/security/v1`** JSON that can be merged into graph views or downstream tooling. |

Default tab is **Visualizer** (`visualizer`). The header shows **workspace / {project label} / {active tab}**; **Sync name from schema** reads `metadata.name` from the current schema text.

## Git repository root and export

The **Git repository root** field drives inventory scanning, manifest export, and server-backed summary paths. **Export omnigraph.workspace.json** downloads a manifest built from current pipeline + schema CLI settings ([`gitWorkspace.ts`](../packages/web/src/mvp/gitWorkspace.ts)).

**Reset workspace** clears persisted state and reloads defaults.

## Using a built UI with the API

For same-origin API calls from the browser (no CORS setup):

1. `cd packages/web && npm run build`
2. From repo root: `omnigraph serve --web-dist packages/web/dist`

Then open the URL **`serve`** prints (loopback by default). **Inventory → Load from OmniGraph server** can fill summary data when `/api/v1/workspace/summary` is available. See **`omnigraph serve --help`** for authentication and experimental endpoints.

## Contributor reference

Lint, build, and Wasm rebuild steps: [web-frontend.md](development/web-frontend.md).

## See also

- [Product philosophy](product-philosophy.md)
- [CLI and CI](cli-and-ci.md) — emit graph JSON from the terminal for paste into Visualizer or CI artifacts
- [Overview](overview.md)
