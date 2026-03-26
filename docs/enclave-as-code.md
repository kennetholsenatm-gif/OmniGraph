# Enclave as Code (EaC) Documentation

## Overview

Enclave as Code (EaC) is a declarative approach to managing Wasm enclaves within OmniGraph. It treats secure, isolated execution environments (Zero-Trust Execution Environments - ZTEEs) as programmable, declarative entities.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     OmniGraph Control Plane                  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Enclave Manager                         │  │
│  │  • ZTEE lifecycle management                        │  │
│  │  • Model deployment and versioning                  │  │
│  │  • Resource constraint enforcement                  │  │
│  │  • Health monitoring and metrics                    │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Graph Connector                         │  │
│  │  • Topology synchronization                         │  │
│  │  • Edge routing configuration                       │  │
│  │  • Event-driven communication                       │  │
│  │  • Telemetry integration                           │  │
│  └──────────────────────────────────────────────────────┘  │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    QminiWasm-Core Runtime                    │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   WasmEdge  │  │   TPEM      │  │   Quantum Router    │ │
│  │   Sandbox   │  │   Memory    │  │   (QAOA)            │ │
│  │             │  │   Manager   │  │                     │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Key Features

### 1. Runtime Constraints

- **Memory Limits**: TPEM optimization with configurable memory bounds
- **Deterministic Execution**: Reproducible inference for debugging
- **Network Access**: Zero-trust isolation policies
- **Filesystem Access**: Controlled access to host resources
- **Max Instances**: Concurrency control for Wasm sandboxes

### 2. Zero-Trust Enclave Enrollment (ZTEE)

- **Enrollment Modes**: open, strict, mutual
- **Attestation Providers**: Keycloak, Vault, custom
- **Certificate Rotation**: Automated certificate lifecycle
- **Peer Allowlisting**: Controlled enclave-to-enclave communication
- **Audit Logging**: Comprehensive security audit trail

### 3. Cognitive Payload Management

- **Source URI**: S3, IPFS, file:// URIs for ML models
- **Weight Formats**: float32, float16, ternary_packed, binary
- **Cryptographic Verification**: Checksum and signature validation
- **Inference Modes**: batch, streaming, edge
- **Preprocessing**: Normalize, resize, quantize

### 4. Quantum-Assisted Routing

- **Dynamic Fallback**: Classical → Quantum when latency exceeds threshold
- **QAOA Algorithms**: Quantum Approximate Optimization Algorithm
- **Telemetry Triggers**: Mesh latency, error rate conditions
- **Multi-provider**: IBM Qiskit, AWS Braket, Azure Quantum

## CLI Commands

### Apply an Enclave Manifest

```bash
omnigraph enclave apply -f edge-agent.yaml
```

### Check Enclave Status

```bash
omnigraph enclave status my-enclave
```

### List All Enclaves

```bash
omnigraph enclave list
```

### Delete an Enclave

```bash
omnigraph enclave delete my-enclave
```

### Enroll with ZTEE

```bash
omnigraph enclave enroll my-enclave --provider keycloak
```

### Sync Graph Topology

```bash
omnigraph enclave graph-sync -f topology.yaml
```

## Manifest Structure

```yaml
apiVersion: omnigraph/v1alpha1
kind: Enclave
metadata:
  name: my-enclave
  namespace: production
  labels:
    app: my-app
  annotations:
    description: "My Wasm enclave"
spec:
  deployment_strategy: standalone
  replicas: 1
  
  runtime:
    engine: wasmedge
    memory_limit_mb: 256
    cpu_limit_ms: 1000
    deterministic_execution: true
    network_access: false
    filesystem_access: none
    max_instances: 1
  
  trust_boundary:
    enrollment: strict
    attestation_provider: keycloak
    allowed_peers:
      - "peer-enclave-1"
    certificate_rotation: 24h
    audit_log: true
  
  cognitive_payload:
    source_uri: "file:///models/my-model.wasm"
    weight_format: float32
    checksum: "sha256:abc123..."
    inference_mode: batch
  
  routing:
    strategy: dynamic_fallback
    classical_heuristic: cascade_rl
    quantum_fallback:
      enabled: true
      provider: ibm_qiskit_runtime
      algorithm: qaoa
      threshold: "mesh_latency > 50ms"
      max_qubits: 16
  
  resources:
    requests:
      cpu: "500m"
      memory: "256Mi"
    limits:
      cpu: "2"
      memory: "1Gi"
  
  environment:
    LOG_LEVEL: "info"
  
  health_check:
    enabled: true
    endpoint: /health
    interval_seconds: 30
  
  scaling:
    min_replicas: 1
    max_replicas: 10
    target_cpu_percent: 80
```

## Integration with QminiWasm-core

### Responsibilities

**OmniGraph (Control Plane)**:
- Define enclave topology and relationships
- Manage declarative state (Enclave as Code)
- Route data between enclaves
- Distribute workloads to agents
- Monitor health and collect telemetry

**QminiWasm-core (Execution Plane)**:
- Instantiate secure WASM enclaves
- Execute TPEM artifacts
- Handle hardware acceleration (Ternary, SYCL)
- Manage intra-enclave security
- Process inference requests

### Communication

The integration uses gRPC for communication between OmniGraph and QminiWasm-core:

- **Proto Files**: `proto/omnigraph_enclave.proto`
- **Service**: `OmniGraphEnclaveService`
- **Methods**: CreateEnclave, GetEnclaveStatus, SyncGraph, etc.

## Example Workflows

### Deploy an Edge Inference Agent

```bash
# 1. Create the enclave manifest
cat > edge-agent.yaml << EOF
apiVersion: omnigraph/v1alpha1
kind: Enclave
metadata:
  name: edge-agent-1
spec:
  runtime:
    engine: wasmedge
    memory_limit_mb: 128
  trust_boundary:
    enrollment: strict
  cognitive_payload:
    source_uri: "s3://models/agent-v1.wasm"
    weight_format: ternary_packed
EOF

# 2. Apply the manifest
omnigraph enclave apply -f edge-agent.yaml

# 3. Check status
omnigraph enclave status edge-agent-1

# 4. Enroll with ZTEE
omnigraph enclave enroll edge-agent-1 --provider keycloak
```

### Synchronize Graph Topology

```bash
# 1. Create topology file
cat > topology.yaml << EOF
graphId: my-graph
nodes:
  - id: edge-agent-1
    type: enclave
    properties:
      location: edge
  - id: central-agent
    type: enclave
    properties:
      location: cloud
edges:
  - from: edge-agent-1
    to: central-agent
    relationship: routes_to
    metadata:
      protocol: quantum
EOF

# 2. Sync the graph
omnigraph enclave graph-sync -f topology.yaml
```

## Best Practices

1. **Version Control**: Store manifests in Git
2. **Resource Limits**: Always set memory and CPU limits
3. **Health Checks**: Enable health checks for all enclaves
4. **ZTEE Enrollment**: Use strict enrollment for production
5. **Audit Logging**: Enable audit logs for compliance
6. **Certificate Rotation**: Set appropriate rotation intervals
7. **Peer Allowlisting**: Restrict communication to known peers

## Troubleshooting

### Enclave Not Starting

- Check resource limits (memory, CPU)
- Verify Wasm module exists at source_uri
- Review health check configuration
- Check ZTEE enrollment status

### Performance Issues

- Monitor memory usage via metrics
- Check quantum routing triggers
- Review TPEM weight format
- Verify network access permissions

### Security Issues

- Verify ZTEE enrollment
- Check certificate expiry
- Review allowed peers
- Audit security logs