# Web client

The UI under [`web/`](../web/) is a **shift-left** surface: validation runs in the browser to keep feedback fast.

## Schema validation

- The editor validates **`.omnigraph.schema`** as you type (debounced).
- Uses **JSON Schema draft 2020-12** (same contract as the control plane), via Ajv.
- Supports **YAML** or **JSON** in the textarea.

## Wasm linters (ADR 001)

ADR [001 — Wasm linters](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/docs/adr/001-wasm-linters.md) describes browser-side linters. Shipped today:

- **`hcldiag.wasm`** — Hashicorp HCL parse diagnostics, a small **structure** heuristic (`terraform {}` vs `resource`), and a **pattern scan** for obvious quoted secrets / `AKIA…` literals (see `wasm/tfpattern`, exported as `omnigraphTfPatternLint`).
- **Optional smoke Wasm** — a tiny `add(i32,i32)` module gated by `VITE_ENABLE_WASM_SPIKE=true` (proves the loader path).

Full **tflint**, **checkov**, and **ansible-lint** ports remain roadmap. See [`wasm/README.md`](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/wasm/README.md) and [CONTRIBUTING.md — Web app](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/CONTRIBUTING.md) for local dev and `make wasm-hcldiag`.

## Visualizer

Paste **`omnigraph graph emit`** JSON (`omnigraph/graph/v1`) into the graph panel. **React Flow** with **Dagre** lays out nodes and edges; **custom node types** include **telemetry** nodes (CMDB context, gray state, SVG accent). A separate D3 canvas is optional only if product needs layouts React Flow cannot express.
