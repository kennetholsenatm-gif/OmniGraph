# ADR 008: WebAssembly bridge hardening

Status: Accepted

## Context

OmniGraph ships **Go-compiled WebAssembly** for in-browser diagnostics (see [ADR 001](001-wasm-linters.md)). JavaScript invokes exported functions that accept **untrusted strings** (pasted HCL, large buffers, odd encodings). Historically, defects at this boundary could surface as:

- **Panics** inside the Wasm runtime on unexpected inputs, destabilizing the **entire browser tab** rather than a single panel.
- **Malformed JSON** returned across the bridge, causing **`JSON.parse`** to throw on the TypeScript side with the same user-visible severity.

The bridge is therefore treated as a **client-side execution surface**: not a network trust zone, but still a place where **robustness equals product quality**.

## Decision

1. **No panics on user-controlled paths** in Wasm-exported handlers. Parse, scan, and marshal steps **recover** into structured diagnostic output or stable error envelopes encoded as JSON arrays/objects the UI already understands.
2. **Bounded behavior** for oversized or pathological inputs: prefer early rejection with a clear diagnostic over unbounded work in the browser.
3. **Consistent JSON on the wire** from Go: if marshaling ever fails, return a **minimal valid JSON** error diagnostic list rather than invalid UTF-8 or truncated output.
4. **Fuzz testing** in Go: use **`go test -fuzz`** against the same code paths that run under `js/wasm` (parsers, scanners, and JSON marshaling helpers), with seed corpora under **`testdata/fuzz/`** in the relevant modules (for example `wasm/hcldiag`, `wasm/tfpattern`). Fuzz targets run in CI on a schedule or on change, consistent with repository policy.
5. **TypeScript resilience**: callers **wrap** bridge invocations, catch failures, and surface **inline errors** in the editor or status UI instead of letting exceptions unwind unrelated React state.

## Consequences

- **Security posture** documentation links Wasm alongside `serve` hardening for contributor awareness ([Security posture](../../security/posture.md)).
- New Wasm exports must include **fuzz seeds** and a **short doc note** in [`wasm/README.md`](../../../wasm/README.md).
- ADR 001 remains the **product** decision (“run linters in Wasm”); ADR 008 records **engineering discipline** at the boundary.

## Related

- [ADR 001 — Wasm linters](001-wasm-linters.md)
- [`wasm/README.md`](../../../wasm/README.md)
- [`packages/web/src/hclWasm.ts`](../../../packages/web/src/hclWasm.ts) (bridge loader and callers)
