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
| `OPENVS_CODE_SKIP` | `1` = skip OpenVSCode Server |
| `OPENVS_CODE_PORT` | Browser IDE port (**default `3010`** so Gitea can keep `3000`) |
| `OPENVS_CODE_REQUIRE_TOKEN` | `1` = URL token + `openvscode-server-token.service.example`; **`0` (default)** = no token |
| `OPENVS_CODE_BIND` | Listen address (default **`127.0.0.1`** without token; use **`0.0.0.0`** if Windows cannot reach the editor) |
| `HERMES_BITNET_CONFIG_SKIP` | `1` = skip Hermes→BitNet wiring |
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
| 08 | [08-hermes-bitnet-config.sh](08-hermes-bitnet-config.sh) — point Hermes at BitNet |

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
