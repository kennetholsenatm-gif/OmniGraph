# AI Orchestration – Service Summary (Phase 5)

## New in this plan (run in AlmaLinux 10 container or host)

| Service / component | Port | Notes |
|---------------------|------|------|
| **code-server** (VSCode Web UI) | 8080 | Start with `PASSWORD=<pwd> code-server --bind-addr 0.0.0.0:8080 --auth password`. Password from Phase 1 script or env `CODE_SERVER_PASSWORD`. |
| **n8n-mcp** | 3001 (host) → 3000 (container) | MCP server exposing n8n tools to AI clients. Optional overlay: [docker-compose.ai-orchestration.yml](../docker-compose/docker-compose.ai-orchestration.yml). |
| **~/ai-orchestration** | — | MCP server clones, Python venv (CrewAI, AutoGen, LlamaIndex, LangChain, etc.), Solace config and agent cards. |

## Reused from devsecops-pipeline (do not duplicate)

| Service | Container | Port | Notes |
|---------|-----------|------|------|
| **n8n** | devsecops-n8n | 5678 | [docker-compose.tooling.yml](../docker-compose/docker-compose.tooling.yml). Access via Traefik at /n8n or http://localhost:5678. |
| **Solace broker** | devsecops-solace | 8008 (WS) | [docker-compose.messaging.yml](../docker-compose/docker-compose.messaging.yml). |
| **Solace Agent Mesh** | devsecops-sam | — | Same compose; uses broker. Env: SOLACE_BROKER_URL, SOLACE_* from secrets-bootstrap. |

## Default passwords and auth

- **code-server:** Set in Phase 1 (generated or `CODE_SERVER_PASSWORD`). Document in your run environment.
- **n8n:** Created on first access to n8n UI (no default from this repo). Existing instance may already have admin; use Vault / secrets-bootstrap for pipeline secrets.
- **Solace / SAM:** `SOLACE_ADMIN_PASSWORD`, `SOLACE_PASSWORD` (and related) from [secrets-bootstrap.ps1](../scripts/secrets-bootstrap.ps1); stored in Vault at `secret/devsecops`.

## Start order (full stack)

1. Create networks: `.\scripts\create-networks.ps1`
2. Bootstrap and start: `.\scripts\secrets-bootstrap.ps1`
3. Optional AI overlay: from `docker-compose`:  
   `docker compose -f docker-compose.iam.yml -f docker-compose.messaging.yml -f docker-compose.tooling.yml -f docker-compose.chatops.yml -f docker-compose.ai-orchestration.yml up -d`

## CLI validation (Phase 5)

```bash
opencode run "echo 'Environment Ready'"
cline version
# or: cline "echo Environment Ready"
```
