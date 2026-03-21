# Incus profile: Hermes + BitNet + OpenVSCode + qminiwasm-core

Run these commands on the **Incus host** (not inside the guest), after `incus admin init` (or equivalent).

## Memory and disk

- **Falcon3-10B-Instruct 1.58-bit (`i2_s`)**: allow **≥56 GiB** RAM for the guest if you run BitNet `setup_env.py` F32 conversion **inside** the container; otherwise build the GGUF elsewhere and copy it in.
- **Root disk**: **≥80 GiB** for sources, venvs, and weights.

## Create profile

```bash
chmod +x create-profile.sh
./create-profile.sh
```

Edit variables at the top of `create-profile.sh` if your pool name or limits differ.

## Launch AlmaLinux 10 guest

```bash
incus launch images:almalinux/10/cloud hermes-bitnet \
  -p hermes-bitnet \
  -c limits.memory=56GiB \
  -d root,size=80GiB
```

(`-d root,size=` depends on your storage pool defaults; adjust to grow disk.)

## Publish ports to the host

Proxy **BitNet** (8080) and **OpenVSCode Server** (3000) from the guest loopback to the host:

```bash
incus config device add hermes-bitnet proxy8080 proxy \
  listen=tcp:0.0.0.0:8080 connect=tcp:127.0.0.1:8080
incus config device add hermes-bitnet proxy3000 proxy \
  listen=tcp:0.0.0.0:3000 connect=tcp:127.0.0.1:3000
```

Then from Windows (WSL networking): use `localhost:8080` / `localhost:3000` or the WSL IP.

## Bootstrap inside the guest

```bash
incus exec hermes-bitnet -- bash -lc "dnf install -y git curl sudo && git clone <your-mirror> /tmp/devsecops-pipeline && cd /tmp/devsecops-pipeline/deployments/hermes-bitnet-lxc && ./bootstrap-all.sh"
```

Prefer copying this folder into the guest or mounting a read-only Incus disk device instead of cloning in production.

## WSL note

Nested Incus under WSL often fails or behaves oddly. If `incus launch` fails, run [../bootstrap-all.sh](../bootstrap-all.sh) **directly on Alma WSL** and treat Incus as optional.

## Security

- OpenVSCode Server **must** use a **connection token** (see `07-openvscode-server.sh` and `systemd/openvscode-server.service.example`).
- Do not expose 3000/8080 on untrusted networks without TLS and auth.
