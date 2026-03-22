# Execution matrix (runners)

OmniGraph uses a **pluggable interface** for actual execution. Ephemeral, sandboxed environments run the right toolchain per phase.

## Planned plugin families

| Family | Examples | Role |
|--------|----------|------|
| Provisioning | OpenTofu, Terraform, Pulumi | Plan and apply infrastructure |
| Configuration | Ansible, Puppet, Chef | Configure instances after provision |
| Container / compute | Podman, Docker, LXC | Sandboxed job execution |
| MicroVM (optional) | Firecracker | Stronger isolation for untrusted or sensitive workloads |

## Interface sketch (future)

- **Runner contract:** Input = resolved workspace path (or OCI image), environment variables (including injected secrets in memory only), timeout, and a structured **step descriptor** (tool name, argv, working directory).
- **Isolation:** Each step runs in a fresh sandbox; secrets are not written to disk (see [ADR 003](adr/003-memory-only-secrets.md)).
- **Observability:** Structured logs with secret masking; exit codes and stdout/stderr captured for the graph and PR comments.

No runner plugins are implemented in the repository skeleton; this document captures the intended shape for control-plane integration.
