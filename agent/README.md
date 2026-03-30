# OmniGraph sync agent (background)

The **sync agent** is an optional **long-running process** (sidecar, systemd unit, or cluster workload) that keeps a **WebSocket** open to the OmniGraph control plane. It is **not** a user onboarding tool and **not** a substitute for the browser workspace—it **feeds** normalized state **from disk** into the server and can **apply** whole-file mutations **back** to allowed paths when the control plane requests them.

## What it does

- **Downstream (agent → server):** On a fixed poll interval (about **10 seconds**), walks each writable root with the same discovery rules as `repo.Discover`, normalizes **`.tfstate`** and **Ansible inventory** files (INI-style names such as `inventory`, `hosts`, or `*.ini`), merges them, and sends a `state_delta` only when node/edge identity **changes**—so identical scans do not re-flood the hub.
- **Upstream (server → agent):** Handles `apply_mutation` messages: writes **utf8** or **base64** payload to a path that must stay **under** one of the configured roots (relative paths resolved and validated). Decoded payloads are capped (currently **32 MiB**).

## Configuration (environment)

| Variable | Required | Meaning |
|----------|----------|---------|
| `OMNIGRAPH_SYNC_WS_URL` | Yes | WebSocket URL, e.g. `ws://127.0.0.1:38671/api/v1/sync/ws` |
| `OMNIGRAPH_SYNC_TOKEN` | Yes | Bearer token (must match `serve` auth) |
| `OMNIGRAPH_SYNC_WRITABLE_PATHS` | Yes | Comma- or path-list-separated **directory roots** on the agent host; all reads/writes are confined here |

The control plane must be started with **`--enable-sync-ws-api`** (and authentication) for the route to exist. Workspace server invocation and flags: **[docs/development/contributor-commands.md](../docs/development/contributor-commands.md)** and [docs/ci-and-contributor-automation.md](../docs/ci-and-contributor-automation.md).

## Security notes

- Treat the token like a **secret**; use TLS (`wss://`) outside loopback.
- Roots should be **minimal**—only the checkout or state directories the agent must mirror.
- Mutations are **whole-file replace** by design; review server policy before enabling write paths in production.

## Implementation

Go package: [`internal/syncdaemon`](../internal/syncdaemon/).
