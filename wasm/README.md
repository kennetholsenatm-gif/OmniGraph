# WebAssembly linters (roadmap)

This directory is reserved for **browser-side** linter and scanner Wasm builds aligned with [ADR 001](../docs/adr/001-wasm-linters.md).

## Planned artifacts (not built in the bootstrap milestone)

| Artifact | Source ecosystem | Purpose |
|----------|------------------|---------|
| `tflint.wasm` | OpenTofu/Terraform | Syntax and API validation |
| `checkov.wasm` | Checkov | Security and compliance scanning |
| `ansible-lint.wasm` | ansible-lint | Playbook best practices |

## Approach (future work)

- Evaluate official or community Wasm builds versus embedding a smaller validated subset.
- Load modules from the web app (`web/`) behind feature flags; keep JSON Schema validation as the always-on baseline in `schemas/`.

Contributions should track issues in the main repository before large toolchain ports.
