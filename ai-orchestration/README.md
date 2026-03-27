# AI Orchestration (event-driven AI dev environment)

Scripts and config for the AI Orchestration Environment plan (see repo docs or plan file). Target: **AlmaLinux 10** (e.g. Docker container). Reuses **n8n** and **Solace/SAM** from devsecops-pipeline.

## Quick start (inside AlmaLinux 10 container)

1. **Phase 1 – Core CLI and code-server**  
   `bash phase1-core-setup.sh`  
   Then start code-server: `PASSWORD=<pwd> code-server --bind-addr 0.0.0.0:8080 --auth password`

2. **Phase 2 – MCP servers and Python frameworks**  
   `export AI_ORCH=~/ai-orchestration  # or /workspace/ai-orchestration if mounted`  
   `bash phase2-mcp-frameworks.sh`

3. **Phase 3 – n8n-mcp (optional)**  
   From repo `docker-compose/`:  
   `docker compose -f docker-compose.iam.yml -f docker-compose.messaging.yml -f docker-compose.tooling.yml -f docker-compose.chatops.yml -f docker-compose.ai-orchestration.yml up -d n8n-mcp`  
   See [N8N_MCP_CLIENT_CONFIG.md](N8N_MCP_CLIENT_CONFIG.md).

4. **Phase 4 – Solace**  
   Reuse existing devsecops-solace and devsecops-sam. Topic routing and agent cards: [solace-config/](solace-config/).

5. **Phase 5 – Validate**  
   `bash phase5-validate.sh`  
   See [SERVICE_SUMMARY.md](SERVICE_SUMMARY.md) and [.env.example](.env.example).

## Files

| File | Purpose |
|------|--------|
| phase1-core-setup.sh | code-server, gcloud, OpenCode, Cline (AlmaLinux 10) |
| phase2-mcp-frameworks.sh | Clone/build MCP servers; Python venv (CrewAI, LangChain, etc.) |
| phase5-validate.sh | OpenCode and Cline version / run check |
| MCP_SERVERS.md | MCP server run commands for n8n/proxy config |
| N8N_MCP_CLIENT_CONFIG.md | n8n-mcp and MCP client setup in n8n |
| solace-config/ | Topic routing and agent cards for SAM |
| SERVICE_SUMMARY.md | Ports, services, passwords |
| .env.example | API keys and URLs to fill |
