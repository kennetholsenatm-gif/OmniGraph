# ADR 007: Enclave as Code

## Status

Proposed

## Context

OmniGraph needs a declarative way to bridge its continuous state engine with QminiWasm-core's zero-trust execution environment. Current infrastructure-as-code handles metal and networking but doesn't address cognitive boundaries, trust enrollment, and memory-constrained ML constraints of edge nodes.

## Decision

We will implement **Enclave as Code** by extending OmniGraph's schema to natively understand QminiWasm's unique architecture: Zero-Trust Enclave Enrollment (ZTEE), Ternary-Packed Memory (TPEM), and Quantum Routing.

### Schema Extension

Introduce `wasm_enclave` as a first-class target type in `.omnigraph.schema`:

```yaml
targets:
  inference-agents:
    type: wasm_enclave
    depends_on: [targets.edge-gateway]
    deployment_strategy: agent_mesh
    
    runtime:
      engine: wasmedge
      memory_limit_mb: 128
      deterministic_execution: true
    
    trust_boundary:
      enrollment: strict
      attestation_provider: keycloak
      allowed_peers: ["targets.inference-agents.*"]
    
    cognitive_payload:
      weight_format: ternary_packed
      source_uri: "s3://models/agent-v2-packed.bin"
    
    routing:
      strategy: dynamic_fallback
      classical_heuristic: cascade_rl
      quantum_fallback:
        enabled: true
        provider: ibm_qiskit_runtime
        threshold: "mesh_latency > 50ms"
```

### Key Features

#### 1. Runtime Constraints
- **Memory limits**: TPEM optimization with configurable memory bounds
- **Deterministic execution**: Reproducible inference for debugging
- **Network/filesystem access**: Zero-trust isolation policies
- **Max instances**: Concurrency control for Wasm sandboxes

#### 2. Zero-Trust Enclave Enrollment (ZTEE)
- **Enrollment modes**: open, strict, mutual
- **Attestation providers**: Keycloak, Vault, custom
- **Certificate rotation**: Automated certificate lifecycle
- **Peer allowlisting**: Controlled enclave-to-enclave communication

#### 3. Cognitive Payload Management
- **Source URI**: S3, IPFS, file:// URIs for ML models
- **Weight formats**: float32, float16, ternary_packed, binary
- **Cryptographic verification**: Checksum and signature validation
- **Inference modes**: batch, streaming, edge

#### 4. Quantum-Assisted Routing
- **Dynamic fallback**: Classical → Quantum when latency exceeds threshold
- **QAOA algorithms**: Quantum Approximate Optimization Algorithm
- **Telemetry triggers**: Mesh latency, error rate conditions
- **Multi-provider**: IBM Qiskit, AWS Braket, Azure Quantum

### Architecture

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

### Implementation

#### File Structure
```
internal/enclave/
├── types.go          # Data structures
└── manager.go        # Enclave lifecycle manager

schemas/
└── enclave.v1.schema.json  # JSON Schema

testdata/enclaves/
└── edge-inference-agent.yaml  # Example configuration
```

#### Key Types
- `Enclave`: Top-level configuration
- `EnclaveSpec`: Desired state
- `RuntimeConfig`: Wasm runtime settings
- `TrustBoundary`: ZTEE configuration
- `CognitivePayload`: ML model configuration
- `RoutingConfig`: Quantum routing settings

### Consequences

**Positive:**
- Declarative management of Wasm enclaves
- Automated ZTEE lifecycle
- TPEM memory optimization
- Quantum routing for complex optimizations
- GitOps-friendly configuration

**Negative:**
- Additional complexity in orchestration
- WasmEdge dependency
- Quantum provider integration overhead

## Related

- [ADR 006: QminiWasm-Core Agent Mesh](006-qminiwasm-agent-mesh.md)
- [Architecture](../architecture.md)
- [Integrations](../integrations.md)