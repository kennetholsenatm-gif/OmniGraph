# Execution matrix (runners)

OmniGraph uses a **pluggable interface** for actual execution. Ephemeral, sandboxed environments run the right toolchain per phase.

## Implemented

| Component | Location | Notes |
|-----------|----------|--------|
| `Runner` / `Step` / `Result` | [`internal/runner/runner.go`](../internal/runner/runner.go) | `Step` supports optional container fields (`ContainerImage`, `Mounts`, `ContainerWorkdir`). |
| `ExecRunner` | [`internal/runner/exec.go`](../internal/runner/exec.go) | Host `os/exec` for local dev and CI. |
| `ContainerRunner` | [`internal/runner/container.go`](../internal/runner/container.go) | `docker` or `podman` CLI; `docker run --rm -i -v … -e KEY=val` (no `--env-file` for secrets; see [ADR 003](adr/003-memory-only-secrets.md)). |
| Pipeline | `omnigraph orchestrate` | [`internal/orchestrate`](../internal/orchestrate); `--runner=exec` (default) or `--runner=container`. |

### Reference container images (examples)

Pin by digest in production; tags below are illustrative.

| Tool | Example image |
|------|----------------|
| OpenTofu | `ghcr.io/opentofu/opentofu:1.8` (default in orchestrate) |
| Ansible | `cytopia/ansible:latest` (default in orchestrate) |

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
- **Observability:** Exit codes and stdout/stderr captured for the graph and PR comments; secret masking remains roadmap.
