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
