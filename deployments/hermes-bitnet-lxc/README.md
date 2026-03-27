# Hermes + BitNet + OpenVSCode Server + qminiwasm-core (Alma / Incus LXC)

## Canonical paths (this workspace)

| Role | Windows | WSL (typical) |
|------|---------|----------------|
| **Automation / bootstrap** | `C:\GiTeaRepos\devsecops-pipeline` | `/mnt/c/GiTeaRepos/devsecops-pipeline` |
| **qminiwasm-core (train / edit)** | `C:\GitHub\LLM_Pract\qminiwasm-core` | `/mnt/c/GitHub/LLM_Pract/qminiwasm-core` |

Bootstrap scripts auto-pick **qminiwasm-core** from the WSL path above when `.git` exists; otherwise they clone `QMINI_REPO` into `~/src/qminiwasm-core`. Set `QMINI_DIR` to force a path.

## Browser URLs on your machine

If **Gitea** is at **`http://localhost:3000`**, use this bundle as follows so ports do not clash:

| App | URL |
|-----|-----|
| Gitea | `http://localhost:3000` |
| OpenVSCode Server (default) | `http://localhost:3010/` — **no token in URL** (`--without-connection-token`, binds **127.0.0.1**). If the browser cannot reach WSL, set `OPENVS_CODE_BIND=0.0.0.0` (trusted network only). |
| OpenVSCode (optional token mode) | `OPENVS_CODE_REQUIRE_TOKEN=1` + `systemd/openvscode-server-token.service.example` → use `?tkn=` |
| Hermes in OpenVSCode | **ACP Client** ([formulahendry.acp-client](https://open-vsx.org/extension/formulahendry/acp-client) on Open VSX) — run [09-openvscode-hermes-acp.sh](09-openvscode-hermes-acp.sh): BitNet row uses `hermes acp`; OpenRouter coding agent uses `hermes-acp-coding-agent` (`python -m acp_adapter`). See **Hermes Agent + OpenVSCode (ACP)** below. |
| BitNet `llama-server` (Hermes OpenAI base) | `http://localhost:8080/v1` |

---

Idempotent bootstrap for an **AlmaLinux**-style environment (bare metal, VM, **Incus LXC**, or **WSL AlmaLinux**) to run:

- **Hermes Agent** ([NousResearch/hermes-agent](https://github.com/NousResearch/hermes-agent)) via official `install.sh`
- **BitNet.cpp** `llama-server` with the largest supported **1.58-bit** stack in BitNet’s `setup_env.py` by default: **`tiiuae/Falcon3-10B-Instruct-1.58bit`** → `ggml-model-i2_s.gguf`
- **OpenVSCode Server** ([gitpod-io/openvscode-server](https://github.com/gitpod-io/openvscode-server)) in the browser (default: **no URL token**, localhost bind; optional token mode for LAN)
- Optional **Coder code-server** ([03-code-server.sh](03-code-server.sh)) — set `CODE_SERVER_SKIP=1` to install only OpenVSCode
- **qminiwasm-core** from [kennetholsenatm-gif/qminiwasm-core](https://github.com/kennetholsenatm-gif/qminiwasm-core) (override with `QMINI_REPO` / `QMINI_REPO_URL`)

## Quick start (inside the guest)

```bash
cd deployments/hermes-bitnet-lxc
export HERMES_BITNET_SRC_ROOT="$HOME/src"    # optional
# Large download + conversion — hours and tens of GB:
#   BITNET_SKIP_MODEL_DOWNLOAD=1 ./bootstrap-all.sh   # build only, model later
./bootstrap-all.sh
```

Then install **user systemd** units (optional):

```bash
mkdir -p ~/.config/systemd/user
cp systemd/bitnet-llama-server.service.example ~/.config/systemd/user/bitnet-llama-server.service
cp systemd/openvscode-server.service.example ~/.config/systemd/user/openvscode-server.service
# LAN + URL token instead: use systemd/openvscode-server-token.service.example
# Edit BITNET_GGUF in bitnet unit if non-default
systemctl --user daemon-reload
systemctl --user enable --now bitnet-llama-server openvscode-server
loginctl enable-linger "$USER"
```

## Incus (host side)

See [incus/README.md](incus/README.md) and [incus/create-profile.sh](incus/create-profile.sh). Run [00-incus-preflight.sh](00-incus-preflight.sh) on the **host** to sanity-check `incus info`.

**Nested Incus under WSL** is fragile; if `incus launch` fails, run this bootstrap **directly on Alma WSL** instead.

## Ports and WSL forwarding

**Conflict check:** Gitea and other UIs often use **3000**. This bundle defaults OpenVSCode Server to **3010** so **3000 stays free** for Gitea. Set `OPENVS_CODE_PORT` if **3010** is taken (e.g. `export OPENVS_CODE_PORT=3100` before `07-openvscode-server.sh`).

| Service            | Port (default) | Notes                                      |
|--------------------|----------------|--------------------------------------------|
| BitNet llama-server| 8080           | OpenAI-compatible `/v1`                   |
| OpenVSCode Server  | **3010** (default) | Token in `~/.config/hermes-bitnet-lxc/openvscode.token` — **not 3000** (common Gitea port) |
| Coder code-server  | 8443           | If not skipped                             |

From **Windows → WSL2**: `localhost:8080` usually forwards to WSL automatically. If not, use `wsl hostname -I` and connect to that IP.

## RAM and disk

- **Falcon 10B `i2_s` conversion** inside the guest: plan **≥48–56 GiB RAM** (F32 export peak), matching a generous **WSL `.wslconfig`** on the Windows side if applicable.
- **Disk**: **≥80 GiB** root recommended for sources, venvs, HF cache, and GGUF.

## Hermes as a coding agent with BitNet (what works today)

A **coding agent** here means Hermes drives an **edit/run loop**: the model receives **OpenAI-style `tools`**, returns **`tool_calls`**, Hermes executes tools (files, terminal, etc.), and the loop continues. That requires an HTTP API that **accepts** `tools` / `tool_choice` and returns compatible **`tool_calls`**.

**Working setup in this bundle (recommended):** run [09-openvscode-hermes-acp.sh](09-openvscode-hermes-acp.sh), then in OpenVSCode **ACP → Agents** connect to **`Hermes (OpenRouter - tools)`**. Put **`OPENROUTER_API_KEY`** (or **`OPENAI_API_KEY`**) in **`~/.hermes/.env`**. The wrapper [hermes-acp-coding-agent.sh](hermes-acp-coding-agent.sh) (installed to **`~/.local/bin/hermes-acp-coding-agent`**) exports OpenRouter **`OPENAI_BASE_URL`**, sets **`HERMES_CHAT_COMPLETIONS_NO_TOOLS=0`**, and runs **`$HERMES_AGENT_DIR/venv/bin/python3 -m acp_adapter`** (not **`hermes acp`**) so **`hermes_cli/main.py`** never loads **`~/.hermes/.env`** with **`override=True`** and wipes the wrapper. [lib/patch_hermes_acp_entry_dotenv_no_override.py](lib/patch_hermes_acp_entry_dotenv_no_override.py) patches **`acp_adapter/entry.py`** so ACP’s own **`.env`** load also uses **`override=False`**, filling keys from **`~/.hermes/.env`** without clobbering **`OPENAI_BASE_URL`**. [lib/patch_hermes_acp_session_model_env.py](lib/patch_hermes_acp_session_model_env.py) lets **`HERMES_ACP_MODEL`** override the default model while **`config.yaml` still names your BitNet GGUF** for other flows.

**Local BitNet:** connect to **`Hermes (BitNet - chat)`** for **multi-turn chat only**. Stock BitNet **`llama-server`** **rejects** `tools` in **`examples/server/utils.hpp`** → HTTP 500. [08-hermes-bitnet-config.sh](08-hermes-bitnet-config.sh) + **`HERMES_CHAT_COMPLETIONS_NO_TOOLS=1`** strip `tools` so chat works — **not** a full coding-agent loop on BitNet.

**Future: coding agent on BitNet weights only** needs a **`llama-server` build whose OpenAI layer supports `tools`** with BitNet inference (see [microsoft/BitNet#10](https://github.com/microsoft/BitNet/issues/10), [microsoft/BitNet#432](https://github.com/microsoft/BitNet/issues/432)). Then remove **`HERMES_CHAT_COMPLETIONS_NO_TOOLS`** from **`.env`** / **systemd** and use a single agent entry if you prefer.

## Hermes Agent + OpenVSCode (ACP)

OpenVSCode Server uses the **Open VSX** registry, not the Visual Studio Marketplace. Hermes’s own [ACP setup](https://github.com/NousResearch/hermes-agent/blob/main/docs/acp-setup.md) references **Anysphere ACP Client** (`anysphere.acp-client`) and `acpClient.agents` with a **`registryDir`**; that extension is **not** what this bundle installs. Instead:

1. **Python:** `agent-client-protocol` in the Hermes venv (same dependency as `pip install -e ".[acp]"` in the upstream doc). The bootstrap uses **`uv pip install`** when `uv` is available (Hermes installer provides it); otherwise **`ensurepip` + pip**.
2. **Editor:** install **[ACP Client](https://open-vsx.org/extension/formulahendry/acp-client)** (`formulahendry.acp-client`) via the OpenVSCode CLI.
3. **Hermes patches:** [lib/patch_hermes_acp_session_model_env.py](lib/patch_hermes_acp_session_model_env.py) on **`acp_adapter/session.py`** ( **`HERMES_ACP_MODEL`** for OpenRouter while **`config.yaml` keeps BitNet** ); [lib/patch_hermes_acp_entry_dotenv_no_override.py](lib/patch_hermes_acp_entry_dotenv_no_override.py) on **`acp_adapter/entry.py`** (ACP loads **`~/.hermes/.env`** with **`override=False`** so wrapper **`OPENAI_BASE_URL`** wins); [lib/patch_hermes_run_agent_preserve_acp_openrouter_env.py](lib/patch_hermes_run_agent_preserve_acp_openrouter_env.py) on **`run_agent.py`** (after **`load_hermes_dotenv`**, restore **`OPENAI_BASE_URL`** / **`HERMES_CHAT_COMPLETIONS_NO_TOOLS`** when **`HERMES_ACP_PRESERVE_OPENROUTER_ENV=1`** — otherwise logs still show probes against **`127.0.0.1:8080`** and **`Unsupported param: tools`**).
4. **Wrapper:** [hermes-acp-coding-agent.sh](hermes-acp-coding-agent.sh) → **`~/.local/bin/hermes-acp-coding-agent`** (OpenRouter + **`HERMES_CHAT_COMPLETIONS_NO_TOOLS=0`**, **`python -m acp_adapter`**).
5. **Settings:** merge **`acp.agents`** with two entries — **`Hermes (OpenRouter - tools)`** and **`Hermes (BitNet - chat)`** — into **User** and **Machine** `settings.json`.

Run idempotently:

```bash
./09-openvscode-hermes-acp.sh
```

OpenVSCode often reads **Machine** settings; both paths are written so **Hermes** rows appear reliably. Reload (**Developer: Reload Window**), then **ACP → Agents** → connect to **`Hermes (OpenRouter - tools)`** for a **coding agent** (needs **`OPENROUTER_API_KEY`** in **`~/.hermes/.env`**) or **`Hermes (BitNet - chat)`** for **local chat** only.

**Skip:** `OPENVS_CODE_HERMES_ACP_SKIP=1` or `OPENVS_CODE_SKIP=1`. **Override extension id:** `OPENVS_CODE_ACP_EXTENSION=...`. **Non-default Hermes checkout:** `HERMES_AGENT_DIR`, `HERMES_BIN`, `HERMES_VENV_PY`. **Coding-agent wrapper path:** `HERMES_ACP_CODING_AGENT_BIN`. **Custom server data dir** (if you pass `--user-data-dir` to openvscode-server): set **`OPENVS_CODE_USER_DATA`** so both **User** and **Machine** `settings.json` paths match; or set **`OPENVS_CODE_USER_SETTINGS`** / **`OPENVS_CODE_MACHINE_SETTINGS`** explicitly.

User **systemd** units set **`PATH=%h/.local/bin:...`** so the integrated terminal finds **`hermes`**; reload the unit after updating from this repo.

## Security

- **OpenVSCode Server (default):** runs **without** a connection token on **127.0.0.1** only — fine for single-user WSL; do **not** switch to **`0.0.0.0`** without a token on untrusted networks. For LAN exposure use **`OPENVS_CODE_REQUIRE_TOKEN=1`** and the **token** unit example.
- **BitNet / Hermes**: `OPENAI_API_KEY=dummy` is fine for local `llama-server`; do not expose **8080** without a firewall.

## Environment variables (common)

| Variable | Purpose |
|----------|---------|
| `QMINI_REPO` / `QMINI_REPO_URL` | Git URL for qminiwasm-core |
| `QMINI_DIR` | Workspace path (default `~/src/qminiwasm-core` or `/mnt/c/...`) |
| `BITNET_SKIP_MODEL_DOWNLOAD` | `1` = build BitNet only |
| `BITNET_HF_REPO` | Hugging Face repo for `setup_env.py` |
| `BITNET_MODEL_DIR` | `-md` parent dir under BitNet repo |
| `CODE_SERVER_SKIP` | `1` = skip Coder code-server |
| `CODE_SERVER_INSTALL_METHOD` | **`standalone`** (default): tarball under **`~/.local`**, no **`sudo rpm`**. Use **`detect`** for rpm/deb (needs sudo on Alma/RHEL). |
| `CODE_SERVER_INSTALL_VERSION` | Optional pin, e.g. **`4.91.1`** (passed to official install.sh **`--version`**) |
| `OPENVS_CODE_SKIP` | `1` = skip OpenVSCode Server |
| `OPENVS_CODE_PORT` | Browser IDE port (**default `3010`** so Gitea can keep `3000`) |
| `OPENVS_CODE_REQUIRE_TOKEN` | `1` = URL token + `openvscode-server-token.service.example`; **`0` (default)** = no token |
| `OPENVS_CODE_BIND` | Listen address (default **`127.0.0.1`** without token; use **`0.0.0.0`** if Windows cannot reach the editor) |
| `OPENVS_CODE_HERMES_ACP_SKIP` | `1` = skip [09-openvscode-hermes-acp.sh](09-openvscode-hermes-acp.sh) (ACP extension + settings merge) |
| `OPENVS_CODE_ACP_EXTENSION` | Open VSX extension id (default **`formulahendry.acp-client`**) |
| `OPENVS_CODE_USER_DATA` | OpenVSCode user data root (default **`~/.openvscode-server`**) — used for **User** + **Machine** `settings.json` |
| `OPENVS_CODE_USER_SETTINGS` | Override User `settings.json` path (default **`$OPENVS_CODE_USER_DATA/data/User/settings.json`**) |
| `OPENVS_CODE_MACHINE_SETTINGS` | Override Machine `settings.json` path (default **`$OPENVS_CODE_USER_DATA/data/Machine/settings.json`**) |
| `HERMES_AGENT_DIR` | Hermes git checkout used for venv + docs (default **`~/.hermes/hermes-agent`**) |
| `HERMES_BITNET_CONFIG_SKIP` | `1` = skip Hermes→BitNet wiring |
| `HERMES_RUN_AGENT` | Path to Hermes **`run_agent.py`** for the no-tools patch (default **`~/.hermes/hermes-agent/run_agent.py`**) |
| `HERMES_CHAT_COMPLETIONS_NO_TOOLS` | Set to **`1`** (via **`08`** + **`.env`** / systemd) so Hermes omits **`tools`** for local BitNet **`llama-server`**; coding-agent wrapper forces **`0`** |
| `OPENROUTER_API_KEY` / `OPENAI_API_KEY` | Required for **`Hermes (OpenRouter - tools)`** (e.g. in **`~/.hermes/.env`**) |
| `HERMES_ACP_MODEL` | OpenRouter (or other) model id for the coding-agent wrapper (default **`openai/gpt-4o-mini`**) |
| `HERMES_ACP_OPENROUTER_BASE` | Override API base (default **`https://openrouter.ai/api/v1`**) |
| `HERMES_ACP_CODING_AGENT_BIN` | Install path for [hermes-acp-coding-agent.sh](hermes-acp-coding-agent.sh) (default **`~/.local/bin/hermes-acp-coding-agent`**) |
| `HERMES_ACP_PRESERVE_OPENROUTER_ENV` | Set to **`1`** by the OpenRouter wrapper so [lib/patch_hermes_run_agent_preserve_acp_openrouter_env.py](lib/patch_hermes_run_agent_preserve_acp_openrouter_env.py) keeps **`OPENAI_BASE_URL`** and **`HERMES_CHAT_COMPLETIONS_NO_TOOLS`** after **`run_agent`** reloads **`~/.hermes/.env`** |
| `HERMES_INSTALL_LOG` | Path for installer `tee` log |

## Scripts

| Step | Script |
|------|--------|
| 00 | [00-incus-preflight.sh](00-incus-preflight.sh) — host-side Incus check (non-fatal) |
| 01 | [01-dnf-prereqs.sh](01-dnf-prereqs.sh) — DNF packages |
| 02 | [02-hermes.sh](02-hermes.sh) — Hermes `install.sh` + log |
| 03 | [03-code-server.sh](03-code-server.sh) — optional Coder code-server |
| 04 | [04-bitnet-build.sh](04-bitnet-build.sh) — BitNet build + optional weights |
| 05 | [05-qminiwasm.sh](05-qminiwasm.sh) — qminiwasm venv |
| 06 | [06-playwright-chromium.sh](06-playwright-chromium.sh) — optional (`HERMES_BITNET_RUN_PLAYWRIGHT=1`) |
| 07 | [07-openvscode-server.sh](07-openvscode-server.sh) — OpenVSCode Server tarball |
| 08 | [08-hermes-bitnet-config.sh](08-hermes-bitnet-config.sh) — point Hermes at BitNet; patch **`run_agent.py`** for **`HERMES_CHAT_COMPLETIONS_NO_TOOLS`** + **`.env`** |
| 09 | [09-openvscode-hermes-acp.sh](09-openvscode-hermes-acp.sh) — Hermes **ACP** (extension, venv, session + entry + **`run_agent`** OpenRouter patches, wrapper, **User** + **Machine** `settings.json`) |

## Training (qminiwasm-core)

After `05-qminiwasm.sh`:

```bash
source ~/venvs/qminiwasm-core/bin/activate
# Repo checkout (canonical on your PC): C:\GitHub\LLM_Pract\qminiwasm-core → WSL:
cd /mnt/c/GitHub/LLM_Pract/qminiwasm-core
# If bootstrap cloned to ~/src instead, cd there. From repo root: source deployments/hermes-bitnet-lxc/lib/common.sh && cd "$QMINI_DIR"
# Follow qminiwasm-core README (training loop, optional pip install -e ".[serve]"); prefer checkpoints on Linux ext4 for IO.
```

## Troubleshooting

- **`http://localhost:8080/` says “llama.cpp” — is this really BitNet?**  
  **Yes, if you followed this bundle.** Microsoft [BitNet](https://github.com/microsoft/BitNet) builds a server binary still named **`llama-server`**, from a **vendored llama.cpp** tree with **BitNet** kernels and codegen. The web UI and `/v1/models` JSON often still say **llama.cpp** / `owned_by: llamacpp` — that is upstream branding, not proof you built vanilla llama.cpp. **Verify the binary path** is under your BitNet clone, e.g.  
  `readlink -f ~/src/BitNet/build/bin/llama-server`  
  If you instead built and run **`llama-server` from a plain [ggml-org/llama.cpp](https://github.com/ggml-org/llama.cpp) clone** (no BitNet repo), that would *not* be BitNet.cpp — rebuild from **`microsoft/BitNet`** and use `04-bitnet-build.sh` / `run-bitnet-server.sh`.
- **Paths / ports (WSL):** run `bash deployments/hermes-bitnet-lxc/lib/verify-wsl-paths.sh` from WSL — confirms `qminiwasm-core` and **QMINI_DIR**, and prints **3010** vs Gitea **3000**.
- **BitNet CMake / Clang errors**: `04-bitnet-build.sh` patches `setup_env.py` to **GCC** + `LLAMA_BUILD_SERVER=ON`; manual CMake already uses GCC.
- **OOM during conversion**: increase guest RAM, use WSL `.wslconfig`, or convert on another machine and copy `ggml-model-i2_s.gguf`.
- **Hermes still on OpenRouter**: re-run `08-hermes-bitnet-config.sh` after the GGUF exists, or fix `~/.hermes/config.yaml` manually.
- **code-server install / `rpm -U` failed**: On Alma/RHEL the upstream script’s **`detect`** path uses **`sudo rpm`**. This bundle defaults **`CODE_SERVER_INSTALL_METHOD=standalone`** in [03-code-server.sh](03-code-server.sh) so install does not need root. If you overrode with **`detect`**, run **`sudo rpm -U ~/.cache/code-server/code-server-<version>-amd64.rpm`** (exact filename under **`~/.cache/code-server/`**), or switch back to standalone.
- **Hermes missing in ACP Client (only the default `npx` agents)**: Re-run [09-openvscode-hermes-acp.sh](09-openvscode-hermes-acp.sh) so **`acp.agents`** is written under **`data/Machine/settings.json`** as well as **User**; then **Developer: Reload Window**. Confirm **`OPENVS_CODE_USER_DATA`** matches your **`openvscode-server`** **`--user-data-dir`** if non-default.
- **Hermes / ACP: `Unsupported param: tools` (HTTP 500 from `llama-server`)**: BitNet’s vendored **`llama-server`** OpenAI-compat layer **rejects** the **`tools`** (and **`tool_choice`**) fields outright (`3rdparty/llama.cpp/examples/server/utils.hpp`). Hermes normally sends **`tools`** for function calling. This bundle runs [lib/patch_hermes_run_agent_no_tools.py](lib/patch_hermes_run_agent_no_tools.py) from [08-hermes-bitnet-config.sh](08-hermes-bitnet-config.sh), sets **`HERMES_CHAT_COMPLETIONS_NO_TOOLS=1`** in **`~/.hermes/.env`**, and adds the same variable to [systemd/openvscode-server.service.example](systemd/openvscode-server.service.example) so **`hermes acp`** spawned from the browser inherits it. Re-run **`08`**, restart OpenVSCode (or **`systemctl --user restart openvscode-server`**), and try again. **Trade-off:** the backend does not receive tool schemas, so you get **plain chat** with that model, not the full **coding-agent** tool loop; see **Hermes as a coding agent with BitNet** above for the gap and what has to change upstream.
- **OpenRouter ACP agent still probes `http://127.0.0.1:8080/v1` or returns `Unsupported param: tools`:** **`run_agent.py`** reloads **`~/.hermes/.env`** at import with **`override=True`**, which resets **`OPENAI_BASE_URL`** to BitNet and **`HERMES_CHAT_COMPLETIONS_NO_TOOLS`** to **`1`**. Re-run [09-openvscode-hermes-acp.sh](09-openvscode-hermes-acp.sh) (applies [lib/patch_hermes_run_agent_preserve_acp_openrouter_env.py](lib/patch_hermes_run_agent_preserve_acp_openrouter_env.py)), ensure **`~/.local/bin/hermes-acp-coding-agent`** exports **`HERMES_ACP_PRESERVE_OPENROUTER_ENV=1`**, then reconnect the agent.
