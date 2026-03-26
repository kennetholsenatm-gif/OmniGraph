# Getting Started with OmniGraph

This guide walks you through installing, configuring, and running OmniGraph for the first time.

## Prerequisites

### Required Software

- **Go 1.21+** - For building the control plane
- **OpenTofu/Terraform** - Infrastructure provisioning
- **Ansible** - Configuration management
- **Git** - Version control

### Optional Software

- **Node.js 18+** - For web UI development
- **Docker/Podman** - For containerized execution
- **NetBox** - For CMDB integration
- **HashiCorp Vault** - For secret management

### Operating Systems

OmniGraph supports:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/kennetholsenatm-gif/OmniGraph.git
cd OmniGraph

# Build the binary
make build

# Or build for Windows
make build-windows

# The binary will be in bin/omnigraph
```

### From Binary Releases

Download the latest release from the [GitHub Releases](https://github.com/kennetholsenatm-gif/OmniGraph/releases) page.

## Quick Start

### 1. Create a Schema File

Create a `.omnigraph.schema` file in your project:

```yaml
apiVersion: omnigraph/v1
kind: Schema
metadata:
  name: my-project
  version: "1.0.0"
spec:
  variables:
    environment:
      type: string
      default: dev
      description: Deployment environment
    region:
      type: string
      default: us-west-2
      description: AWS region
  
  hosts:
    web:
      type: aws_instance
      count: 2
      ami: ami-12345678
      instance_type: t3.micro
      tags:
        Name: web-${count.index}
        Environment: ${variables.environment}
  
  outputs:
    web_ips:
      value: ${hosts.web.*.private_ip}
      description: Web server IP addresses
```

### 2. Validate the Schema

```bash
# Validate syntax and types
omnigraph validate .omnigraph.schema

# Coerce to different formats
omnigraph coerce .omnigraph.schema --format=tfvars
omnigraph coerce .omnigraph.schema --format=groupvars
omnigraph coerce .omnigraph.schema --format=env
```

### 3. Plan Infrastructure

```bash
# Initialize OpenTofu
tofu init

# Create a plan
tofu plan -out=tfplan

# Generate graph visualization
omnigraph graph emit .omnigraph.schema \
  --plan-json <(tofu show -json tfplan) \
  --output graph.json
```

### 4. Apply and Interrogate State

```bash
# Apply infrastructure
tofu apply

# Extract hosts from state
omnigraph state hosts terraform.tfstate

# Generate Ansible inventory
omnigraph inventory from-state terraform.tfstate > inventory.ini
```

### 5. Run Configuration Management

```bash
# Run Ansible with generated inventory
ansible-playbook -i inventory.ini site.yml
```

## Configuration Files

### .omnigraph.schema

The main configuration file that defines:
- Variables and their types
- Host definitions
- Outputs
- Security policies

See [Configuration](Configuration) for detailed schema reference.

### omnigraph.workspace.json

Optional workspace configuration for the web UI:

```json
{
  "gitRepositoryRoot": ".",
  "pipelineWorkdir": ".",
  "pipelineAnsibleRoot": "../ansible",
  "pipelinePlaybookRel": "site.yml",
  "schemaCliPath": "omnigraph"
}
```

## Common Use Cases

### Use Case 1: Schema Validation in CI

Add to your CI pipeline:

```yaml
# .github/workflows/ci.yml
- name: Validate OmniGraph Schema
  run: |
    omnigraph validate .omnigraph.schema
    omnigraph graph emit .omnigraph.schema --output /dev/null
```

### Use Case 2: State Interrogation

```bash
# Parse state and extract specific outputs
omnigraph state parse terraform.tfstate

# Get hosts in JSON format
omnigraph state hosts terraform.tfstate --format=json

# Generate inventory
omnigraph inventory from-state terraform.tfstate --output=inventory.ini
```

### Use Case 3: Graph Visualization

```bash
# Generate graph from schema and plan
omnigraph graph emit .omnigraph.schema \
  --plan-json plan.json \
  --tfstate terraform.tfstate \
  --output graph.json

# View in browser UI
cd web && npm run dev
# Open http://localhost:5173 and paste graph.json
```

### Use Case 4: Declarative Infrastructure (New)

```yaml
# manifest.yaml
apiVersion: omnigraph.io/v1
kind: InfrastructureManifest
metadata:
  name: my-infrastructure
spec:
  resources:
    - apiVersion: omnigraph.io/v1
      kind: ComputeInstance
      metadata:
        name: web-server
      spec:
        provider: incus
        type: container
        source:
          alias: ubuntu/22.04
        config:
          limits.cpu: "2"
          limits.memory: 4GiB
        state: running
```

```bash
# Apply manifest
omnigraph apply -f manifest.yaml

# Check status
omnigraph get instances
omnigraph describe instance web-server
```

## Web UI

The web UI provides:
- Real-time schema validation
- Graph visualization with React Flow
- HCL diagnostics via WebAssembly
- Repository-wide IaC scanning

### Running the Web UI

```bash
cd web
npm install
npm run dev

# Open http://localhost:5173
```

### Building for Production

```bash
cd web
npm run build

# Serve with omnigraph
omnigraph serve --web-dist web/dist
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OMNIGRAPH_SERVE_TOKEN` | Bearer token for API authentication | - |
| `VAULT_ADDR` | HashiCorp Vault address | - |
| `VAULT_TOKEN` | HashiCorp Vault token | - |
| `AWS_PROFILE` | AWS profile for credentials | default |
| `NETBOX_URL` | NetBox API URL | - |
| `NETBOX_TOKEN` | NetBox API token | - |

## Troubleshooting

### Schema Validation Fails

```bash
# Check schema syntax
omnigraph validate .omnigraph.schema --verbose

# Coerce to see intermediate representation
omnigraph coerce .omnigraph.schema --format=all
```

### State Parsing Issues

```bash
# Ensure state is in JSON format
tofu show -json terraform.tfstate > state.json

# Parse with verbose output
omnigraph state parse state.json --verbose
```

### Graph Generation Problems

```bash
# Check plan JSON format
tofu show -json tfplan | jq . > plan.json

# Generate graph with debug
omnigraph graph emit .omnigraph.schema --plan-json plan.json --verbose
```

## Next Steps

- Read [Architecture](Architecture) to understand system design
- See [CLI Reference](CLI-Reference) for all commands
- Check [Integrations](Integrations) for external tool setup
- Review [Lifecycle](Lifecycle) for deployment workflows

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/kennetholsenatm-gif/OmniGraph/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kennetholsenatm-gif/OmniGraph/discussions)
- **Wiki**: [This wiki](Home)