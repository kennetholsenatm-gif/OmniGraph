# Backend WASI plugins

The Go control plane runs **WebAssembly** modules built with **`GOOS=wasip1` `GOARCH=wasm`** inside **wazero** ([`internal/runner`](../../internal/runner)). There are **two** supported plugin classes with different contracts and capabilities.

## Class A — Parser plugins (stdio only, no network)

**Host API:** [`runner.RunWASIParser`](../../internal/runner/wasiparser.go) / `RunWASIParserLimit`.

**Sandbox:** Deny-by-default: **stdin/stdout only**—no host directory mounts, **no network**, minimal environment.

### Contract

- **stdin:** opaque domain bytes (e.g. Ansible INI). The host does not interpret them.
- **stdout:** a single JSON document validating as **`omnigraph/graph/v1`** ([`schemas/graph.v1.schema.json`](../../schemas/graph.v1.schema.json)).

### Reference

- Source: [`wasm/plugins/ansibleini/main.go`](../../wasm/plugins/ansibleini/main.go)
- Build:

```bash
GOOS=wasip1 GOARCH=wasm go build -o ansible-ini-parser.wasm ./wasm/plugins/ansibleini
```

- Tests: [`internal/runner/wasiparser_test.go`](../../internal/runner/wasiparser_test.go)

## Class B — Integration micro-containers (stdio + allowlisted host HTTP)

**Host API:** [`runner.RunIntegrationPlugin`](../../internal/runner/integration_host.go).

**Sandbox:** Same WASI baseline as Class A, plus one host module **`omnigraph`** exporting **`http_fetch`**. The guest calls this import with a small JSON request; the host enforces **URL prefix allowlists**, **method allowlists**, **request/response size caps**, and **HTTP timeouts**. There is **no** general-purpose `net/http` API exposed to Go application code for vendor-specific clients—the only egress from the guest is through this import.

### Contract

- **stdin:** validates as **`omnigraph/integration-run/v1`** ([`schemas/integration-run.v1.schema.json`](../../schemas/integration-run.v1.schema.json)). The envelope’s `spec.allowedFetchPrefixes` must **exactly match** the host’s configured list for the invocation (defense in depth against stale or tampered guests).
- **stdout:** validates as **`omnigraph/integration-result/v1`** ([`schemas/integration-result.v1.schema.json`](../../schemas/integration-result.v1.schema.json)). When `spec.inventorySnapshot` is set, it must also validate as **`omnigraph/inventory-source/v1`**.

### Shipped examples

| Plugin   | Source | Upstream API (inside guest) |
|----------|--------|-----------------------------|
| NetBox   | [`wasm/plugins/netbox/main.go`](../../wasm/plugins/netbox/main.go) | REST (`/api/dcim/devices/`, etc.) |
| Zabbix   | [`wasm/plugins/zabbix/main.go`](../../wasm/plugins/zabbix/main.go) | JSON-RPC (`/api_jsonrpc.php`) |

Build (from repository root):

```bash
GOOS=wasip1 GOARCH=wasm go build -o netbox.wasm ./wasm/plugins/netbox
GOOS=wasip1 GOARCH=wasm go build -o zabbix.wasm ./wasm/plugins/zabbix
```

### Running

- **CLI:** `omnigraph integration-run --wasm=path/to/netbox.wasm < run.json` (stdin is the integration-run document).
- **HTTP (optional):** enable **`--enable-integration-run-api`** on the workspace server and call **`POST /api/v1/integrations/run`** with JSON `{ "wasmPath": "…", "run": { … } }` (requires the same auth model as other privileged APIs).

### Security notes

- Treat plugins as **untrusted** unless you built and pinned them yourself.
- **Host HTTP imports are still privileged:** isolation is **allowlisting + budgets**, not “WASM magically blocks all egress.”
- **Secrets:** pass tokens only through the stdin envelope assembled by the host operator; never commit them to repos.

## Contributing without a core PR

1. **Parser:** `//go:build wasip1` `main`, stdin → graph/v1 stdout; run under **`RunWASIParser`**.
2. **Integration:** same build tag, stdin integration-run/v1 → integration-result/v1 stdout; implement `//go:wasmimport omnigraph http_fetch` with the JSON wire format expected by the host (see `ogFetchRequest` / `ogFetchResponse` in [`integration_host.go`](../../internal/runner/integration_host.go)).
3. Build with **Go 1.21+** wasip1 or another WASI toolchain compatible with wazero’s **WASI snapshot preview1** plus the **`omnigraph`** host module.

**Browser Wasm** (HCL diagnostics) is documented under [ADR 008](../core-concepts/adr/008-wasm-bridge-hardening.md); it is not the same ABI as backend WASI plugins.
