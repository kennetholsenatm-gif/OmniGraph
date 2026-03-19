# MCP Server Run Commands (Phase 2)

Use these when configuring n8n MCP Client Tool nodes or an mcp-remote proxy.

| Server | Run command (stdio) |
|--------|----------------------|
| Sequential Thinking | `npx -y @modelcontextprotocol/server-sequential-thinking` |
| SQLite | From clone: `node $AI_ORCH/servers/src/sqlite/dist/index.js` or `npx -y @modelcontextprotocol/server-sqlite` |
| GitHub | From clone: `node $AI_ORCH/servers/src/github/dist/index.js` or `npx -y @modelcontextprotocol/server-github` (requires GITHUB_TOKEN) |
| Superpowers | `node $AI_ORCH/superpowers/dist/index.js` (after npm run build) |
| BrowserMCP | Per browsermcp README; typically `node dist/index.js` in repo root |

Replace `$AI_ORCH` with your `~/ai-orchestration` (or `/workspace/ai-orchestration`) path.
