# Execution matrix (runners)

OmniGraph uses a **pluggable interface** for actual execution. Ephemeral, sandboxed environments run the right toolchain per phase.

## Implemented

| Component | Location | Notes |
|-----------|----------|--------|
| `Runner` / `Step` / `Result` | [`internal/runner/runner.go`](../internal/runner/runner.go) | `Step` supports optional container fields (`ContainerImage`, `Mounts`, `ContainerWorkdir`) and `RedactExtra` for log masking. |
| `ExecRunner` | [`internal/runner/exec.go`](../internal/runner/exec.go) | Host `os/exec` for local dev and CI. |
| `ContainerRunner` | [`internal/runner/container.go`](../internal/runner/container.go) | `docker` or `podman` CLI; `docker run --rm -i -v … -e KEY=val` (no `--env-file` for secrets; see [ADR 003](adr/003-memory-only-secrets.md)). |
| Pipeline | `omnigraph orchestrate` | [`internal/orchestrate`](../internal/orchestrate); `--runner=exec` (default) or `--runner=container`. |
| Log redaction | [`internal/runner/mask.go`](../internal/runner/mask.go) | Substrings from `Step.Env` values (and `RedactExtra`) replaced in captured stdout/stderr before `Result` is returned. |

### Reference container images (examples)

Pin by digest in production; tags below are illustrative.

| Tool | Example image |
|------|----------------|
| OpenTofu | `ghcr.io/opentofu/opentofu:1.8` (default in orchestrate) |
| Ansible | `cytopia/ansible:latest` (default in orchestrate) |
| Pulumi | `pulumi/pulumi:latest` (not wired in `orchestrate` yet; use manually or extend `--iac-engine`) |

### LXC (host CLI)

`ContainerRunner` targets Docker/Podman. For **LXC**, run tools from the host with `ExecRunner` or invoke `lxc exec` explicitly:

```bash
lxc exec my-build-container -- bash -lc 'cd /workspace && tofu version'
```

Mount your workspace into the container’s filesystem the way your site standardizes (e.g. bind mount on the LXC host), then point `omnigraph orchestrate --workdir` at that tree on the host.

### Pulumi (container spike)

Until `orchestrate` grows a Pulumi argv mapping (`--iac-engine=pulumi` currently returns a clear error), you can still run Pulumi in a container with the same pattern as OpenTofu:

```bash
docker run --rm -i -v "$PWD:/workspace" -w /workspace \
  -e PULUMI_ACCESS_TOKEN \
  pulumi/pulumi:latest \
  preview --stack org/stack
```

Document stacks, secrets, and state backends separately from OpenTofu; merging both engines in one pipeline is a follow-on ([architecture](architecture.md) Layer 3 roadmap).

## Plugin families (roadmap)

| Family | Examples | Role |
|--------|----------|------|
| Provisioning | OpenTofu, Terraform, Pulumi | Plan and apply infrastructure |
| Configuration | Ansible, Puppet, Chef | Configure instances after provision |
| Container / compute | Podman, Docker, LXC | Sandboxed job execution |
| MicroVM (optional) | Firecracker | Stronger isolation for untrusted or sensitive workloads |

## Contract

- **Runner input:** Resolved workspace path (bind-mount for containers), environment variables (including injected secrets **in memory only**), timeout, and argv.
- **Isolation:** Prefer a fresh container per step; secrets are not written to `.env` or `terraform.tfvars` on disk ([ADR 003](adr/003-memory-only-secrets.md)).
- **Observability:** Exit codes and stdout/stderr captured for the graph and PR comments; **secret redaction** applies to those streams in the shipped runners ([ADR 003](adr/003-memory-only-secrets.md)).
