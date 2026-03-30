# Contributor commands (internal)

**Audience:** maintainers, CI, and local development only. OmniGraph’s product surface is the **browser workspace**; this file is the **single** place in the repo where copy-paste **shell** sequences live. All other Markdown should link here instead of embedding fenced `bash` / `powershell` blocks.

## Prerequisites

- **Go** `1.23+` (see [CI workflow](../../.github/workflows/ci.yml) for the exact pin; CONTRIBUTING may mention an older minimum—in CI, the workflow wins).
- **Node.js** `20+` and npm (web UI under `packages/web`).
- **Git**

Optional: **Make** (convenience targets), OpenTofu/Terraform/Ansible for lab tests, Docker/Podman.

## Clone repository

```bash
git clone https://github.com/<ORG_OR_USER>/<REPOSITORY>.git
cd <REPOSITORY>
```

## Go workspace sync

From the **repository root**:

```bash
go work sync
```

## Go — vet, test, build control plane

```bash
go vet ./...
go test ./...
go build -o bin/omnigraph ./cmd/omnigraph
./bin/omnigraph -h
```

**Windows PowerShell:**

```powershell
go build -o bin\omnigraph.exe .\cmd\omnigraph
.\bin\omnigraph.exe -h
```

**Using Make (optional):**

```bash
make vet
make test
make build
```

**Fuzzing (Wasm / parser hardening):** from the relevant module (for example `wasm/hcldiag`), run `go test -fuzz` targets as documented in [ADR 008](../core-concepts/adr/008-wasm-bridge-hardening.md).

## Web frontend — install, dev server, lint, build

```bash
cd packages/web
npm ci
npm run dev
```

CI parity for the web package:

```bash
cd packages/web
npm ci
npm run lint
npm run build
```

## Browser Wasm (HCL diagnostics)

```bash
make wasm-hcldiag
```

Optional Wasm spike in the web app:

```bash
cd packages/web
# Bash/Linux/macOS
VITE_ENABLE_WASM_SPIKE=true npm run dev
```

**PowerShell:**

```powershell
cd packages/web
$env:VITE_ENABLE_WASM_SPIKE = "true"
npm run dev
```

Rebuild diagnostics then run the app:

```bash
make wasm-hcldiag
cd packages/web && npm run dev
```

`wasm/README.md` describes artifacts; CI builds `hcldiag` in the Go job and passes it to the web job.

## End-to-end tests

From the repository root (after `go work sync`):

```bash
go test ./e2e/...
```

## PR / CI verification checklist

```bash
go work sync
go vet ./...
go test ./...
go test ./e2e/...
cd packages/web && npm run lint
```

## Local workspace server (API + optional static UI)

**API only:**

```bash
go build -o bin/omnigraph ./cmd/omnigraph
./bin/omnigraph
```

**Windows PowerShell:**

```powershell
go build -o bin\omnigraph.exe .\cmd\omnigraph
.\bin\omnigraph.exe
```

**With a built web app** (build the UI first with `cd packages/web && npm run build`):

```bash
./bin/omnigraph --web-dist packages/web/dist
```

**One-shot from repo root** (no permanent binary):

```bash
cd packages/web && npm run build
cd ../..
go run ./cmd/omnigraph --web-dist packages/web/dist
```

**Defaults:** listens on loopback (`127.0.0.1:38671`). Experimental routes need matching `--enable-*` flags plus `--auth-token` (or `OMNIGRAPH_SERVE_TOKEN`). Use `./bin/omnigraph -h` for the full flag list.

**Quickstart / fixtures:** same server invocation as above is used with `examples/quickstart/`—build UI, then binary with `--web-dist packages/web/dist`.

## `go test` at repo root (CI-style)

```bash
go test ./...
```

## Backend WASI plugins — build artifacts

**Class A — Ansible INI parser** (from repository root):

```bash
GOOS=wasip1 GOARCH=wasm go build -o ansible-ini-parser.wasm ./wasm/plugins/ansibleini
```

**Class B — NetBox / Zabbix integration guests:**

```bash
GOOS=wasip1 GOARCH=wasm go build -o netbox.wasm ./wasm/plugins/netbox
GOOS=wasip1 GOARCH=wasm go build -o zabbix.wasm ./wasm/plugins/zabbix
```

## WASM integration stdin/stdout (maintainer automation)

When testing integration guests outside the HTTP API, the `omnigraph` binary accepts a subcommand that reads **`omnigraph/integration-run/v1`** JSON from stdin. **`--wasm` must be relative** to the current working directory (absolute paths are rejected).

```bash
./bin/omnigraph integration-run --wasm=path/to/netbox.wasm < run.json
```

Operators should prefer **`POST /api/v1/integrations/run`** with **`--enable-integration-run-api`** on the workspace server when available; this block is for contributors validating plugins.

## Wiki manual sync (optional)

Clone the GitHub Wiki git remote (example URL—replace with your fork/org):

```bash
git clone https://github.com/kennetholsenatm-gif/OmniGraph.wiki.git
cd OmniGraph.wiki
```

Copy markdown from a checkout of `main`:

```bash
cp ../OmniGraph/wiki/*.md .
```

Commit and push:

```bash
git add .
git status
git commit -m "docs: sync wiki from main"
git push
```

See [wiki/SYNC.md](../../wiki/SYNC.md) for CI automation and token setup.
