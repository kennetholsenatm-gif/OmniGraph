# Solace Agent Mesh config (Phase 4)

Reuses **devsecops-solace** and **devsecops-sam** from [docker-compose.messaging.yml](../../docker-compose/docker-compose.messaging.yml). No new broker or SAM container.

## Contents

- **topic-routing.yaml** – A2A topic mapping (data-ingestion → LlamaIndex, code-execution → AutoGen, workflow-automation → n8n). Merge or reference per [Solace Agent Mesh docs](https://github.com/SolaceLabs/solace-agent-mesh).
- **agent-cards/** – Self-description YAML for discovery:
  - **crewai-agent.yaml** – CrewAI agent (data-ingestion)
  - **langgraph-agent.yaml** – LangGraph agent (code-execution)

## Using with devsecops-sam

1. **Clone full repo for examples:** `git clone https://github.com/SolaceLabs/solace-agent-mesh.git ~/ai-orchestration/solace-agent-mesh`
2. **Mount this config** into the SAM container (e.g. add volume to `solace-agent-mesh` in docker-compose.messaging.yml):  
   `- ./ai-orchestration/solace-config:/config:ro`
3. Or copy `topic-routing.yaml` and `agent-cards/*` into the cloned repo’s config path and point SAM at it via env.
4. Replace placeholder endpoints in agent cards with real URLs when agents are running.
