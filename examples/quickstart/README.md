# Quickstart fixtures

Minimal Terraform **JSON state** and a valid **`.omnigraph.schema`** for trying the Go CLI without a full cloud stack.

## Build the binary

```bash
go build -o bin/omnigraph ./cmd/omnigraph
```

Windows (PowerShell):

```powershell
go build -o bin\omnigraph.exe .\cmd\omnigraph
```

Use `./bin/omnigraph` below, or `.\bin\omnigraph.exe` on Windows.

## Emit graph JSON (Topology tab / CI artifact)

Fold in OpenTofu/Terraform state (optional `--plan-json` for a plan file from `terraform show -json tfplan`):

```bash
./bin/omnigraph graph emit examples/quickstart/.omnigraph.schema \
  --tfstate examples/quickstart/minimal.tfstate.json > graph.json
```

## Ansible INI from state

```bash
./bin/omnigraph inventory from-state examples/quickstart/minimal.tfstate.json
```

## Optional: same-origin UI + API

Build the web app and run `serve` with `--web-dist` so Inventory and SSE work without CORS setup. See [docs/using-the-web.md](../../docs/using-the-web.md).

Parity sample used in tests: [testdata/sample.omnigraph.schema](../../testdata/sample.omnigraph.schema).
