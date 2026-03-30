# Quickstart fixtures

Minimal Terraform **JSON state** and a valid **Project** document for learning **Topology** and contributor tests without a full cloud stack. Use [`.omnigraph.schema`](.omnigraph.schema) (YAML) or [`.omnigraph.schema.toml`](.omnigraph.schema.toml) (TOML)—both validate the same; TOML is often easier to edit by hand.

## Graph JSON for the Topology tab

The graph emit pipeline used in CI lives in Go (`go test`); see [docs/ci-and-contributor-automation.md](../../docs/ci-and-contributor-automation.md). For a quick manual loop, paste or load **`omnigraph/graph/v1`** JSON into **Topology** after your automation produces it.

## Local workspace server (optional)

Build and run the server so Inventory and SSE work same-origin:

```bash
go build -o bin/omnigraph ./cmd/omnigraph
./bin/omnigraph --web-dist packages/web/dist
```

(After `cd packages/web && npm run build`.) See [docs/using-the-web.md](../../docs/using-the-web.md) and [docs/development/local-dev.md](../../docs/development/local-dev.md).

Parity sample used in tests: [testdata/sample.omnigraph.schema](../../testdata/sample.omnigraph.schema).

## Observation drill (external trigger)

To practice **live** topology behavior (aligned with [getting started — observation drill](../../docs/getting-started.md)):

1. Run the web workspace with ingest / SSE enabled per [using-the-web.md](../../docs/using-the-web.md).
2. Edit and run a **lab-only** script such as [`break_network.sh`](break_network.sh) (or your own automation) while the workspace is open.
3. Expect the UI to **reflect** refreshed graph or inventory data as the backend syncs—OmniGraph does not run the outage for you from a button in the UI.
