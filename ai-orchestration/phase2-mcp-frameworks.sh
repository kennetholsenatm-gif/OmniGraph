#!/usr/bin/env bash
# Phase 2: MCP Servers & Agent Frameworks (~/ai-orchestration)
# Run inside AlmaLinux 10 with repo mounted (e.g. -v $(pwd):/workspace/ai-orchestration)
# Usage: AI_ORCH=/workspace/ai-orchestration bash phase2-mcp-frameworks.sh
# Or: export AI_ORCH=~/ai-orchestration; bash phase2-mcp-frameworks.sh

set -e
AI_ORCH="${AI_ORCH:-$HOME/ai-orchestration}"
mkdir -p "$AI_ORCH"
cd "$AI_ORCH"

# --- 2.1 Clone and build MCP servers ---
clone_build_mcp() {
  # Sequential Thinking (npx, no clone required; optional clone for customization)
  if ! npm list -g @modelcontextprotocol/server-sequential-thinking &>/dev/null; then
    npm install -g @modelcontextprotocol/server-sequential-thinking 2>/dev/null || true
  fi
  echo "Sequential Thinking: npx -y @modelcontextprotocol/server-sequential-thinking"

  # modelcontextprotocol/servers (multiple servers)
  if [[ ! -d servers ]]; then
    git clone --depth 1 https://github.com/modelcontextprotocol/servers.git
  fi
  for dir in servers/src/sequential-thinking servers/src/sqlite servers/src/github 2>/dev/null; do
    if [[ -d "$dir" ]] && [[ -f "$dir/package.json" ]]; then
      (cd "$dir" && npm install && npm run build 2>/dev/null || true)
    fi
  done

  # obra/superpowers
  if [[ ! -d superpowers ]]; then
    git clone --depth 1 https://github.com/obra/superpowers.git
  fi
  if [[ -f superpowers/package.json ]]; then
    (cd superpowers && npm install && npm run build 2>/dev/null || true)
  fi
  echo "Superpowers: node $AI_ORCH/superpowers/dist/index.js (or per-repo README)"

  # BrowserMCP (browser-use/browsermcp or similar)
  if [[ ! -d browsermcp ]]; then
    git clone --depth 1 https://github.com/browser-use/browsermcp.git 2>/dev/null || true
  fi
  if [[ -d browsermcp ]] && [[ -f browsermcp/package.json ]]; then
    (cd browsermcp && npm install && npm run build 2>/dev/null || true)
  fi
}

# --- 2.2 Python venv and agent frameworks ---
setup_python_venv() {
  if [[ ! -d venv ]]; then
    python3 -m venv venv
  fi
  source venv/bin/activate
  pip install --upgrade pip
  pip install crewai autogen swarms llama-index langchain langgraph 2>/dev/null || true
  pip install activepieces composio-core 2>/dev/null || pip install composio-core 2>/dev/null || true
  python -c "
import sys
for p in ('crewai','autogen','llama_index','langchain','langgraph'):
    try: __import__(p.replace('-','_').split('_')[0]); print(p, 'ok')
    except Exception as e: print(p, e)
"
  deactivate
  echo "Python venv at $AI_ORCH/venv (source venv/bin/activate)"
}

clone_build_mcp
setup_python_venv

echo "Phase 2 complete. MCP run commands documented in MCP_SERVERS.md"
