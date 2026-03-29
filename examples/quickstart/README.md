# Quickstart fixtures

Minimal Terraform **JSON state** and a valid **Project** document for trying the Go CLI without a full cloud stack. Use [`.omnigraph.schema`](.omnigraph.schema) (YAML) or [`.omnigraph.schema.toml`](.omnigraph.schema.toml) (TOML)—both validate the same; TOML is often easier to edit by hand.

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

## Observation drill (external trigger)

To practice **live** topology behavior (aligned with [getting started — observation drill](../../docs/getting-started.md)):

1. Run the web workspace with ingest / SSE enabled per [using-the-web.md](../../docs/using-the-web.md).
2. Edit and run a **lab-only** script such as [`break_network.sh`](break_network.sh) (or your own automation) while the workspace is open.
3. Expect the UI to **reflect** refreshed graph or inventory data as the backend syncs—OmniGraph does not run the outage for you from a button in the UI.
