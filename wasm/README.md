# WebAssembly linters

Browser-side tooling aligned with [ADR 001](../docs/adr/001-wasm-linters.md).

## Shipped: HCL parse diagnostics (`hcldiag.wasm`)

A **Go `js/wasm`** build ([`wasm/hcldiag`](./hcldiag)) uses `github.com/hashicorp/hcl/v2` to return parse diagnostics for Terraform-style HCL pasted in the web UI. The UI loads `web/public/wasm/hcldiag.wasm` plus `wasm_exec.js` from the Go toolchain (vendored under `web/public/wasm/`; same BSD license as Go).

**Build locally** (requires Go 1.22+):

```bash
make wasm-hcldiag
```

**CI** builds the wasm artifact in the `go` job and passes it to the `web` job via GitHub Actions artifacts.

## Roadmap artifacts

| Artifact | Source ecosystem | Purpose |
|----------|------------------|---------|
| `tflint.wasm` | OpenTofu/Terraform | Full rule engine (large port) |
| `checkov.wasm` | Checkov | Security and compliance scanning |
| `ansible-lint.wasm` | ansible-lint | Playbook best practices |

## Approach

- Ship incremental Wasm tools (HCL first); keep JSON Schema validation as the always-on baseline in `schemas/`.
- Full tflint/checkov/ansible-lint ports are optional accelerators; track large efforts in issues before starting.
