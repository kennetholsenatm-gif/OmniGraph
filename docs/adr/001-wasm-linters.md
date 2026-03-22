# ADR 001: WebAssembly for linters

## Status

Accepted

## Context

The OmniGraph IDE aims for real-time feedback (lint, security scan, playbook style) as the user edits. Running linters as backend microservices adds latency, consumes server compute, and weakens the “shift-left in the browser” story.

## Decision

Port or wrap Go/Python-based linters (for example Checkov, tflint, ansible-lint) so they execute as **WebAssembly** inside the browser, alongside schema validation.

## Consequences

- **Positive:** Sub-second local feedback; reduced backend cost; aligns with offline-capable editing.
- **Negative:** Porting or embedding large toolchains in Wasm is non-trivial; some tools may need subset features or WASM-specific builds.
- **Mitigation:** Ship JSON Schema validation first; add Wasm plugins incrementally (see `wasm/README.md`).

## Implementation status

- **Phase 1 (done):** `wasm/hcldiag` — Go `GOOS=js GOARCH=wasm` module using HashiCorp **HCL parse** diagnostics in the browser (`omnigraphHclValidate`), loaded from `web/public/wasm/hcldiag.wasm`. This is not a full **tflint** port; it validates HCL syntax/structure only.
- **Later phases:** `tflint.wasm`, `checkov.wasm`, `ansible-lint.wasm` remain roadmap items.
