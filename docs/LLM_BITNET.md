# BitNet LLM (1.58-bit inference)

The pipeline can run a local **BitNet** inference server (Microsoft BitNet 1.58-bit, OpenAI API compatible) behind the Single Pane of Glass gateway.

## Prerequisites

- **Network:** Create `llm_net` before starting the LLM stack. From repo root:
  - Windows: `.\scripts\create-networks.ps1`
  - Linux/macOS: `./scripts/create-networks.sh`
  (These scripts now include `llm_net` with subnet 100.64.6.0/24.)

- **Build:** The image is built from `docker-compose/bitnet/Dockerfile` (clones Microsoft BitNet, downloads the GGUF model, builds C++ backend). First build can take several minutes.

## Start the LLM stack

From the `docker-compose` directory:

```powershell
docker compose -f docker-compose.llm.yml up -d
```

Or use `launch-stack.ps1` from the same directory; it runs one merged `docker compose` for the core stack (IAM, messaging, tooling, ChatOps), then the LLM compose file (unless `-SkipLlm` or `DEVSECOPS_INCLUDE_LLM=0`).

## Access

- **Via gateway (Traefik):** When the Single Pane of Glass is running, use base URL **http://localhost/llm** (or `https://<gateway>/llm` with TLS). OpenAI-compatible endpoints:
  - `GET/POST http://localhost/llm/v1/models`
  - `POST http://localhost/llm/v1/chat/completions`
  - etc. The gateway strips the `/llm` prefix so the backend receives `/v1/...`.

- **Direct (host port):** **http://localhost:8090** (host port 8090 maps to container 8080). Use when the gateway is not running or for local testing. Port 8090 is used to avoid conflict with Zammad (8080).

## Compose and Traefik

- **Compose file:** `docker-compose/docker-compose.llm.yml` — service `bitnet-inference`, container name `devsecops-bitnet-inference`, hostname `bitnet-inference`, network `llm_net`, resource limits (e.g. 4 CPUs, 2G RAM).
- **Traefik:** The gateway (`single-pane-of-glass`) attaches to `llm_net` and routes `PathPrefix(/llm)` to `http://bitnet-inference:8080` with middleware `strip-llm` so the backend sees `/v1/...`.

## References

- [microsoft/BitNet](https://github.com/microsoft/BitNet) — inference framework and model.
- Model: `microsoft/BitNet-b1.58-2B-4T-gguf` (downloaded in the Docker build).
