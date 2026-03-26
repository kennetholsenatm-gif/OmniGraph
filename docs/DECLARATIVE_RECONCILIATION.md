# Declarative Reconciliation Architecture

This document describes the Kubernetes-style declarative reconciliation system added to OmniGraph, including Incus/IncusOS support.

## Overview

OmniGraph now supports a fully declarative approach to infrastructure management, similar to Kubernetes. Instead of imperative commands, you define the desired state of your infrastructure in YAML/JSON manifests, and OmniGraph continuously reconciles the actual state with the desired state.

## Key Features

### 1. Kubernetes-Style Resource Model

Resources are defined using a standard schema:
- `apiVersion`: API version (e.g., `omnigraph.io/v1`)
- `kind`: Resource type (e.g., `ComputeInstance`, `Network`)
- `metadata`: Resource identification (name, namespace, labels)
- `spec`: Desired state
- `status`: Actual state (managed by the system)

### 2. Reconciliation Loop

The reconciliation controller continuously:
1. Watches for changes in desired state (manifests)
2. Queries actual state from providers
3. Computes differences (diff)
4. Applies changes to reconcile actual with desired
5. Updates status and conditions

### 3. Provider Abstraction

Providers implement the `Provider` interface:
- `GetActualState()`: Query current state
- `Apply()`: Apply desired state
- `Delete()`: Remove resource
- `Exists()`: Check if resource exists
- `Watch()`: Watch for changes

### 4. Incus/IncusOS Support

Full support for Incus resources:
- **ComputeInstance**: Containers and VMs
- **Network**: Bridges, OVN, macvlan
- **StoragePool**: ZFS, Btrfs, LVM, Ceph
- **Profile**: Configuration profiles

## Resource Types

### ComputeInstance

```yaml
apiVersion: omnigraph.io/v1
kind: ComputeInstance
metadata:
  name: web-server
  labels:
    app: nginx
    role: web
spec:
  provider: incus
  type: container  # or "virtual-machine"
  source:
    server: https://images.linuxcontainers.org
    protocol: simplestreams
    alias: ubuntu/22.04
  config:
    limits.cpu: "4"
    limits.memory: 8GiB
    security.nesting: "true"
  devices:
    eth0:
      type: nic
      name: eth0
      nictype: bridged
      parent: lxdbr0
    root:
      type: disk
      path: /
      pool: default
      size: 20GiB
  profiles:
    - default
  state: running
```

### Network

```yaml
apiVersion: omnigraph.io/v1
kind: Network
metadata:
  name: web-network
spec:
  provider: incus
  type: bridge
  config:
    ipv4.address: 10.0.0.1/24
    ipv4.nat: "true"
    ipv6.address: none
  description: Network for web services
```

### StoragePool

```yaml
apiVersion: omnigraph.io/v1
kind: StoragePool
metadata:
  name: fast-storage
spec:
  provider: incus
  driver: zfs
  config:
    source: tank/incus
    zfs.pool_name: tank
  description: High-performance ZFS storage
```

### Profile

```yaml
apiVersion: omnigraph.io/v1
kind: Profile
metadata:
  name: web-server
spec:
  provider: incus
  config:
    limits.cpu: "2"
    limits.memory: 4GiB
  devices:
    eth0:
      type: nic
      name: eth0
      nictype: bridged
      parent: lxdbr0
  description: Web server profile
```

## Manifest Structure

A manifest groups multiple resources:

```yaml
apiVersion: omnigraph.io/v1
kind: InfrastructureManifest
metadata:
  name: production-infra
  namespace: platform
spec:
  resources:
    - apiVersion: omnigraph.io/v1
      kind: StoragePool
      # ... storage pool spec
    
    - apiVersion: omnigraph.io/v1
      kind: Network
      # ... network spec
    
    - apiVersion: omnigraph.io/v1
      kind: ComputeInstance
      # ... instance spec
  
  reconciliation:
    interval: 5m
    onDrift: auto  # auto, manual, alert
    retryPolicy:
      maxAttempts: 3
      backoff: exponential
```

## CLI Commands

### Apply

Apply a manifest to create or update resources:

```bash
# Apply a manifest
omnigraph apply -f manifest.yaml

# Dry-run (show what would be done)
omnigraph apply -f manifest.yaml --dry-run

# Apply and wait for reconciliation
omnigraph apply -f manifest.yaml --wait --timeout 10m
```

### Get

List or get resources:

```bash
# List all instances
omnigraph get instances

# Get specific instance
omnigraph get instance web-1

# List networks
omnigraph get networks

# Get in JSON format
omnigraph get instances -o json
```

### Describe

Show detailed resource information:

```bash
# Describe an instance
omnigraph describe instance web-1

# Describe a network
omnigraph describe network web-network
```

### Diff

Show differences between manifest and actual state:

```bash
# Show diff
omnigraph diff -f manifest.yaml

# Show detailed diff
omnigraph diff -f manifest.yaml --detailed
```

## Status and Conditions

Resources track status using Kubernetes-style conditions:

```yaml
status:
  observedGeneration: 1
  state: running
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2026-03-25T16:45:00Z"
      reason: Reconciled
      message: Instance is running and healthy
    
    - type: NetworkReady
      status: "True"
      lastTransitionTime: "2026-03-25T16:45:00Z"
      reason: NetworkAttached
      message: Network eth0 is attached
  
  resources:
    cpu:
      used: "1.5"
      limit: "4"
    memory:
      used: 2GiB
      limit: 8GiB
  
  reconciliation:
    lastAttempt: "2026-03-25T16:45:00Z"
    lastSuccess: "2026-03-25T16:45:00Z"
    consecutiveSuccesses: 5
    consecutiveFailures: 0
```

## Architecture

### Controller

The reconciliation controller (`internal/reconcile/controller.go`) manages:
- Resource registration
- Provider management
- Reconciliation loop
- Status tracking

### Provider Interface

Providers implement the `Provider` interface (`internal/reconcile/controller.go`):
- Resource CRUD operations
- State queries
- Change watching

### Resource Types

Resource types are defined in `internal/resources/types.go`:
- `Resource`: Base resource structure
- `ComputeInstance`: Container/VM resources
- `Network`: Network resources
- `StoragePool`: Storage resources
- `Profile`: Profile resources

### Schemas

JSON schemas validate resources:
- `schemas/compute-instance.v1.schema.json`
- `schemas/network.v1.schema.json`
- `schemas/storage-pool.v1.schema.json`
- `schemas/profile.v1.schema.json`

## Example: Production Infrastructure

See `testdata/incus-manifest.yaml` for a complete example with:
- ZFS storage pool
- Bridge network
- Web server profile
- Multiple web server instances
- Database server

## Benefits

1. **Declarative**: Define what you want, not how to get there
2. **Self-Healing**: Automatic drift correction
3. **GitOps Ready**: Git as source of truth
4. **Observable**: Clear status and conditions
5. **Extensible**: Easy to add new providers
6. **Kubernetes Familiar**: Similar mental model

## Comparison with Imperative Approach

### Imperative (Before)

```bash
# Manual steps
incus launch images:ubuntu/22.04 web-1
incus config set web-1 limits.cpu 4
incus config set web-1 limits.memory 8GiB
incus network attach lxdbr0 web-1 eth0
incus start web-1
```

### Declarative (After)

```yaml
# manifest.yaml
apiVersion: omnigraph.io/v1
kind: ComputeInstance
metadata:
  name: web-1
spec:
  provider: incus
  type: container
  source:
    alias: ubuntu/22.04
  config:
    limits.cpu: "4"
    limits.memory: 8GiB
  state: running
```

```bash
# Apply once
omnigraph apply -f manifest.yaml

# System handles everything else
```

## Integration with Existing Features

The declarative system integrates with existing OmniGraph features:
- **Orchestration**: Can use imperative orchestration within declarative resources
- **Security**: Security scanning applies to declarative resources
- **Telemetry**: Resource status includes telemetry data
- **Graph**: Resources appear in dependency graphs

## Future Enhancements

1. **More Providers**: AWS, Azure, GCP, VMware
2. **Advanced Reconciliation**: Complex dependency handling
3. **Policy Enforcement**: OPA/Rego integration
4. **Multi-Cluster**: Cross-cluster orchestration
5. **GitOps Integration**: Automatic sync from Git
6. **Web UI**: Visual resource management