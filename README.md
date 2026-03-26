# OmniGraph

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![CI](https://github.com/kennetholsenatm-gif/OmniGraph/actions/workflows/ci.yml/badge.svg)](https://github.com/kennetholsenatm-gif/OmniGraph/actions)

> **State-aware DevSecOps orchestration: bridge OpenTofu, Ansible & telemetry in one GitOps pipeline**

## What is OmniGraph?

OmniGraph is a **multi-paradigm orchestration platform** that unifies declarative provisioning (OpenTofu/Terraform), imperative configuration (Ansible), and real-time telemetry (NetBox, Zabbix) into a single GitOps workflow.

**Why it exists:** Traditional DevSecOps treats infrastructure deployments as isolated scripts. OmniGraph acts as a **continuous state engine**, intercepting outputs from provisioning tools and automatically mapping them into configuration management context—no manual inventory stitching, no state file hunting.

**Key capabilities:**
- **Schema-first design** with real-time validation in browser and CLI
- **State interception** from OpenTofu/Terraform JSON output
- **Dynamic inventory generation** for Ansible from infrastructure state
- **Unified graph visualization** of dependencies and telemetry
- **Enterprise RBAC** with Keycloak (OIDC) and FreeIPA (LDAP)
- **Policy-as-Code** with OPA/Rego integration
- **Bare-metal provisioning** via Redfish/IPMI

## How It Works

```
                         .omnigraph.schema
                        (Single Source of Truth)
                                │
                                ▼
                      omnigraph validate
                  (Schema + Policy Validation)
                                │
            ┌───────────────────┼───────────────────┐
            │                   │                   │
            ▼                   ▼                   ▼
    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
    │ OpenTofu     │    │ Ansible      │    │ Telemetry    │
    │ Provisioning │    │ Configuration│    │ Monitoring   │
    │              │    │              │    │              │
    │ • Plan       │    │ • Inventory  │    │ • NetBox     │
    │ • Apply      │    │ • Playbooks  │    │ • Zabbix     │
    │ • State      │    │ • Roles      │    │ • Prometheus  │
    └──────┬───────┘    └──────┬───────┘    └──────┬───────┘
           │                   │                   │
           └───────────────────┼───────────────────┘
                               │
                               ▼
                     omnigraph graph emit
                       (Visualization)
```

## Quick Start

**Prerequisites:** Go 1.22+

```bash
# 1. Clone the repository
git clone https://github.com/kennetholsenatm-gif/OmniGraph.git
cd OmniGraph

# 2. Build the CLI
go build -o omnigraph ./cmd/omnigraph

# 3. Validate a sample schema
./omnigraph validate testdata/sample.omnigraph.schema
```

**Expected output:**
```
ok
```

## Configuration Example

Create a `.omnigraph.schema` file to define your infrastructure:

```yaml
apiVersion: omnigraph/v1
kind: Schema
metadata:
  name: production-cluster
  version: "1.0.0"
spec:
  variables:
    environment:
      type: string
      default: production
    region:
      type: string
      default: us-east-1

  targets:
    web-servers:
      type: aws_instance
      count: 3
      ami: ami-12345678
      instance_type: t3.medium
      tags:
        Name: web-${count.index}
        Environment: ${variables.environment}

  outputs:
    server_ips:
      value: ${targets.web-servers.*.private_ip}
    load_balancer:
      value: ${targets.web-servers.0.public_ip}
```

**Coerce to different formats:**
```bash
# Generate Terraform variables
./omnigraph coerce .omnigraph.schema --format=tfvars

# Generate Ansible group vars
./omnigraph coerce .omnigraph.schema --format=groupvars

# Generate environment variables
./omnigraph coerce .omnigraph.schema --format=env
```

## Documentation

- [Architecture Overview](docs/architecture.md) - System design and components
- [Execution Matrix](docs/execution-matrix.md) - Runner plugins and execution modes
- [Integrations](docs/integrations.md) - NetBox, Vault, Keycloak, Zabbix
- [Bare-Metal Provisioning](docs/bare-metal-provisioning.md) - Redfish/IPMI support
- [Policy-as-Code](docs/IMPROVEMENTS.md) - OPA/Rego integration
- [Wiki](wiki/Home.md) - User guides and tutorials
- [Architecture Decision Records](docs/adr/) - Design rationale
- [Branch protection](docs/branch-protection.md) - Rulesets and required checks for `main`

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing, and contribution guidelines.

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE).