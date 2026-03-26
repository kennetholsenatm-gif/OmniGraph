# OmniGraph Wiki

Welcome to the OmniGraph documentation wiki. OmniGraph is a state-aware, multi-paradigm DevSecOps orchestration platform that bridges declarative provisioning (OpenTofu/Terraform), imperative configuration (Ansible), and real-time environment telemetry (NetBox/Zabbix) through a unified GitOps workflow.

## What is OmniGraph?

OmniGraph acts as a **continuous state engine** that orchestrates infrastructure deployments rather than treating them as isolated scripts. It provides:

- **Schema-first configuration** with real-time validation
- **Multi-paradigm orchestration** across IaC tools
- **State interception** and output mapping
- **Dynamic inventory generation** from infrastructure state
- **Unified graph visualization** of dependencies
- **Enterprise-grade security** with RBAC and audit logging

## Quick Links

### Getting Started
- [Getting Started Guide](Getting-Started) - Installation, prerequisites, and first steps
- [CLI Reference](CLI-Reference) - Complete command reference
- [Configuration](Configuration) - Schema and configuration files

### User Guide
- [Web UI](Web-UI) - Browser-based interface for validation and visualization
- [Lifecycle & Handoff](Lifecycle) - End-to-end deployment workflow
- [Integrations](Integrations) - NetBox, Vault, Keycloak, and more

### Developer Guide
- [Architecture](Architecture) - System design and components
- [Infrastructure IR](Infrastructure-IR) - Intent Reference model
- [Execution Matrix](Execution-Matrix) - Runners and execution environments
- [Declarative Reconciliation](Declarative-Reconciliation) - Kubernetes-style resource management

### Reference
- [Architecture Decisions](Architecture-Decisions) - ADR summaries and rationale
- [Pipeline Runs](Pipeline-Runs) - Run artifact schema and details
- [Inventory Sources](Inventory-Sources) - Triangulated inventory management
- [Troubleshooting](Troubleshooting) - Common issues and solutions

## Architecture Overview

OmniGraph operates in four core layers:

1. **Presentation Layer** - Browser-based UI with real-time validation and React Flow visualization
2. **Control Plane** - Go binary that orchestrates tools (not replaces them)
3. **Execution Matrix** - Pluggable runners (host exec, containers, microVMs)
4. **Data & Telemetry** - Integrations with NetBox, Vault, Zabbix, and more

```
┌─────────────────────────────────────────────────────────┐
│                  Browser UI (React)                      │
│         Schema Validation • Graph Visualization          │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────────────┐
│                 Control Plane (Go)                       │
│    Schema Coercion • State Interception • Orchestration  │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────────────┐
│                 Execution Matrix                         │
│        ExecRunner • ContainerRunner • Firecracker        │
└───────────────────────┬─────────────────────────────────┘
                        │
┌───────────────────────┴─────────────────────────────────┐
│                 Data & Telemetry                         │
│          NetBox • Vault • Zabbix • Prometheus            │
└─────────────────────────────────────────────────────────┘
```

## Key Concepts

### Schema-First Design
Everything starts with `.omnigraph.schema` - a declarative configuration file that defines your infrastructure intent. The schema is validated in real-time using JSON Schema draft 2020-12.

### State Interception
After `tofu apply`, OmniGraph intercepts the `.tfstate` file and extracts outputs (like instance IPs) to map into Ansible context - no manual state management needed.

### Graph Visualization
The `omnigraph graph emit` command produces a JSON graph document (`omnigraph/graph/v1`) that can be visualized in the browser UI or attached to pull requests.

### Enterprise Identity
Supports Keycloak (OIDC) and FreeIPA (LDAP) for RBAC, with permissions defined in `internal/identity` and protected by an `Authorizer` interface.

## Common Workflows

### 1. Schema Validation
```bash
# Validate your .omnigraph.schema file
omnigraph validate .omnigraph.schema

# Coerce schema to different formats
omnigraph coerce .omnigraph.schema --format=all
```

### 2. Infrastructure Deployment
```bash
# Plan infrastructure
tofu plan -out=tfplan

# Generate graph for visualization
omnigraph graph emit .omnigraph.schema --plan-json <(tofu show -json tfplan)

# Apply and intercept state
tofu apply
omnigraph state hosts terraform.tfstate
```

### 3. Declarative Management (New)
```bash
# Apply a declarative manifest
omnigraph apply -f manifest.yaml

# Check resource status
omnigraph get instances
omnigraph describe instance web-1

# Preview changes
omnigraph diff -f manifest.yaml
```

## How This Wiki is Maintained

The canonical copy of these pages lives in the main repository under [`wiki/`](https://github.com/kennetholsenatm-gif/OmniGraph/tree/main/wiki). To use **GitHub Wiki**:

1. In the GitHub repo, enable **Wiki** (Settings → Features).
2. Either copy Markdown from `wiki/` into new wiki pages, or clone the wiki git remote and add the same files.

Long-form specs, ADRs, and diagrams remain in [`docs/`](https://github.com/kennetholsenatm-gif/OmniGraph/tree/main/docs) in the source tree.

## Contributing

See [CONTRIBUTING.md](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/CONTRIBUTING.md) for:
- Prerequisites and local development
- CI parity commands
- Testing guidelines
- Pull request process

## License

OmniGraph is licensed under the MIT License. See [LICENSE](https://github.com/kennetholsenatm-gif/OmniGraph/blob/main/LICENSE) for details.