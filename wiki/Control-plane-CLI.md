# Control plane CLI

The `omnigraph` binary ([`cmd/omnigraph`](../cmd/omnigraph)) orchestrates tools; it does not replace OpenTofu or Ansible.

## Common commands

| Command | Purpose |
|--------|---------|
| `omnigraph validate [path]` | Validate `.omnigraph.schema` (JSON or YAML) against the embedded JSON Schema |
| `omnigraph coerce [path]` | Print in-memory `terraform.tfvars.json` shape, Ansible-style group vars YAML, and env lines (`--format=tfvars\|groupvars\|env\|all`) |
| `omnigraph state parse \| hosts <file>` | Parse Terraform/OpenTofu **JSON** state and list extracted host/IP candidates |
| `omnigraph inventory from-state <file>` | Render an Ansible INI inventory from state |
| `omnigraph graph emit [path]` | Emit **omnigraph/graph/v1** JSON (`--plan-json`, `--tfstate` optional) |
| `omnigraph run -- <cmd> [args]` | Local `os/exec` runner (dev/CI) |
| `omnigraph netbox sync` | POST illustrative sync JSON to a webhook URL (`--url`, `--ip`, `--role`, `--action`) |

Secrets are not written to disk by design; see [ADR 003: Memory-only secrets](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/docs/adr/003-memory-only-secrets.md).

## Sample project file

See [`testdata/sample.omnigraph.schema`](../testdata/sample.omnigraph.schema) in the repository.

## Plan JSON

For projected resources, use OpenTofu/Terraform output from:

`terraform show -json tfplan` (or OpenTofu equivalent) and pass the file to `graph emit --plan-json`.

State fixtures for tests use JSON shaped like real state; example: `internal/state/testdata/minimal.state.json` (filename avoids `.gitignore` patterns for real state files).
