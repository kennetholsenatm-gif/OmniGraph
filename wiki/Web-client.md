# Web client

The UI under [`web/`](../web/) is a **shift-left** surface: validation runs in the browser to keep feedback fast.

## Schema validation

- The editor validates **`.omnigraph.schema`** as you type (debounced).
- Uses **JSON Schema draft 2020-12** (same contract as the control plane), via Ajv.
- Supports **YAML** or **JSON** in the textarea.

## Wasm linters (roadmap)

ADR [001 — Wasm linters](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/docs/adr/001-wasm-linters.md) describes porting **tflint**, **checkov**, and **ansible-lint** to Wasm. The repo includes a **minimal Wasm spike** (exported `add` function) gated by:

`VITE_ENABLE_WASM_SPIKE=true`

See [CONTRIBUTING.md — Web app](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/CONTRIBUTING.md) for local dev.

## Visualizer

The dependency graph (**D3 / SVG**) is not implemented yet; the canvas area is a placeholder until graph data from the control plane is wired end-to-end.
