# Execution Matrix

OmniGraph orchestrates external tools through pluggable runners.

## Runner Types

- `exec`: runs tools directly on the host
- `container`: runs tools in containerized environments

## Typical Pipeline Shape

1. Validate schema and intent artifacts
2. Coerce/prepare tool-specific inputs
3. Run planning steps
4. Run apply and post-apply workflows
5. Emit graph/run/security artifacts

## Compatibility Guidance

Execution strategy should be selected per environment constraints (security policy,
tool availability, reproducibility requirements). No single runner is required.
