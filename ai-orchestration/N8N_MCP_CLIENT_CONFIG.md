# n8n MCP Client Configuration (Phase 3)

The **existing** n8n instance is `devsecops-n8n` (port 5678). The **n8n-mcp** service exposes n8n as an MCP server for AI clients (Claude, Cursor, etc.); it does not add MCP *client* tools *inside* n8n.

## 1. Start n8n-mcp (optional overlay)

From repo root:

```bash
cd docker-compose
docker compose -f docker-compose.iam.yml -f docker-compose.messaging.yml -f docker-compose.tooling.yml -f docker-compose.chatops.yml -f docker-compose.ai-orchestration.yml up -d n8n-mcp
```

Set `N8N_API_KEY` (from n8n Settings → API) so n8n-mcp can manage workflows. Default URL: `http://n8n:5678`. MCP endpoint is on port 3001 (host) by default.

## 2. Using MCP servers *inside* n8n (Sequential Thinking, Superpowers, BrowserMCP)

n8n's **MCP Client Tool** nodes expect SSE/HTTP. Our MCP servers use **stdio**. Options:

- **Option A – mcp-remote proxy:** Run a proxy that wraps stdio servers and exposes them over SSE. Point n8n MCP Client at the proxy URL (e.g. `http://mcp-remote:4000/sse`). You need to run the proxy in a container or on the host that n8n can reach.
- **Option B – n8n subprocess:** If your n8n version supports running MCP servers via command (stdio), add an MCP Client with command e.g. `npx -y @modelcontextprotocol/server-sequential-thinking`.

### MCP server run commands (from Phase 2)

| Tool        | Command |
|------------|---------|
| Sequential Thinking | `npx -y @modelcontextprotocol/server-sequential-thinking` |
| Superpowers         | `node /path/to/ai-orchestration/superpowers/dist/index.js` |
| BrowserMCP          | Per BrowserMCP repo README |

### n8n UI steps (when using a proxy)

1. Open n8n (http://localhost:5678 or via Traefik /n8n).
2. In a workflow, add an **MCP** or **MCP Client Tool** node (name may vary by n8n version).
3. Set the MCP server URL to your stdio-to-SSE proxy URL.
4. Save and test.

## 3. References

- Existing workflows: [n8n-workflows/](../n8n-workflows/)
- Deployment: [docs/DEPLOYMENT.md](../docs/DEPLOYMENT.md) step 6
