# Backend WASI parser plugins

The Go control plane can execute **WebAssembly** modules built with **`GOOS=wasip1` `GOARCH=wasm`** inside a **WASI** sandbox ([`internal/runner`](../../internal/runner)). This is a **deny-by-default** path: guests receive **stdin/stdout only**—no host directory mounts, no network, and environment variables are cleared except what the host explicitly sets (currently minimal/empty).

## Output contract

Plugins must write a single JSON document to **stdout** that validates as **`omnigraph/graph/v1`** (see [`schemas/graph.v1.schema.json`](../../schemas/graph.v1.schema.json)): `apiVersion` = `omnigraph/graph/v1`, `kind` = `Graph`, `metadata.generatedAt` (RFC3339), and `spec.nodes` (each with `id`, `kind`, `label`).

**Input** is domain-specific bytes on **stdin** (for example raw Ansible INI text). The host does not interpret stdin; the guest owns parsing.

## Reference implementation

- Source: [`wasm/plugins/ansibleini/main.go`](../../wasm/plugins/ansibleini/main.go) — reads INI from stdin, emits graph nodes for each host.
- Build (from repository root):

```bash
GOOS=wasip1 GOARCH=wasm go build -o ansible-ini-parser.wasm ./wasm/plugins/ansibleini
```

- Run through the host API from Go tests: see [`internal/runner/wasiparser_test.go`](../../internal/runner/wasiparser_test.go).

## Contributing a plugin without a core PR

1. Implement a **`package main`** with `//go:build wasip1` that reads **stdin** and writes **graph/v1 JSON** to **stdout**.
2. Build with **`GOOS=wasip1 GOARCH=wasm`** (Go 1.21+), or compile from **Rust/C++** to WASI using your toolchain, as long as the module runs under **wazero**’s WASI snapshot preview1 imports.
3. Ship the **`.wasm`** artifact to operators; load it from your integration by calling **`runner.RunWASIParser`** (or a thin wrapper) with the bytes you want the guest to parse.

**Security:** Treat plugins as **untrusted** unless you built them yourself. The runtime caps stdout size (see `defaultMaxStdout` in [`wasiparser.go`](../../internal/runner/wasiparser.go)) and does not expose the host filesystem to the guest.
