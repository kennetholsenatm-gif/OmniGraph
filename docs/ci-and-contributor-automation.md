# CI and contributor automation

This page describes how **contributors and CI** verify OmniGraph: **`go test`**, the **local workspace server** (`cmd/omnigraph`), and **HTTP APIs** that feed the web workspace. For the interactive graph UI, start with [using-the-web.md](using-the-web.md) and [README.md](../README.md).

**Prerequisites:** Go 1.23+, Node.js 20+ (if you run the web app), Git.

**All copy-paste shell sequences** for verification, builds, and running the workspace server live in **[Contributor commands](development/contributor-commands.md)**.

## Verification with `go test`

From the repository root, run the full Go test suite as documented in **[Contributor commands](development/contributor-commands.md)**.

Schema validation, policy gates where configured, graph emit smoke paths, and the **`e2e/`** harness exercise the same Go packages the workspace depends on. Fixtures live under [`testdata/`](../testdata/) and [`examples/quickstart/`](../examples/quickstart/).

## Local workspace server

Build and run the server (APIs only, or with a built static UI) using the steps in **[Contributor commands](development/contributor-commands.md)**.

**Defaults:** listens on loopback (`127.0.0.1:38671`). Treat any non-loopback bind as requiring strong authentication and network controls.

### Core HTTP routes

- `GET /api/v1/health`
- `POST /api/v1/repo/scan` — body `{"path":"."}`
- `POST /api/v1/workspace/summary` — body `{"path":"."}`
- `GET /api/v1/workspace/stream` — SSE `workspace_summary` events (query `path`)

Experimental routes (`POST /api/v1/security/scan`, inventory, host-ops, ingest, sync WebSocket, drift) stay **off** unless you pass the matching **`--enable-*`** flags **and** `--auth-token` (or `OMNIGRAPH_SERVE_TOKEN`). See **[Contributor commands](development/contributor-commands.md)** for inspecting the full workspace server flag list.

## Artifacts and Topology

- **Project intent** — `.omnigraph.schema` (TOML recommended); validated in the Schema Contract tab and in tests.
- **`omnigraph/graph/v1` JSON** — topology for the Topology tab; produced by the graph emit path exercised in `go test` and ingest flows your team enables.

Minimal fixtures: see **[examples/quickstart/README.md](../examples/quickstart/README.md)**.

## End-to-end suite

Run **`e2e/`** as in **[Contributor commands](development/contributor-commands.md)**.

Details: [E2E testing](development/e2e-testing.md).

## Where to go next

- [Using the web workspace](using-the-web.md)
- [Overview](overview.md)
- [Security posture](security/posture.md)
- [Execution matrix](core-concepts/execution-matrix.md)
- [Local development](development/local-dev.md)
- [Contributor commands](development/contributor-commands.md)
- [CONTRIBUTING.md](../CONTRIBUTING.md)
